package container

import (
	"context"
	"io/ioutil"
	"net"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/darxkies/k8s-tew/pkg/config"
	"github.com/darxkies/k8s-tew/pkg/utils"
	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
)

func getCSIGlobalMounts(destinationPrefix string) []string {
	mounts := []string{}

	bytes, error := ioutil.ReadFile("/proc/mounts")
	if error != nil {
		log.WithFields(log.Fields{"error": error}).Debug("Read mounts failed")

		return mounts
	}

	content := string(bytes)

	lines := strings.Split(content, "\n")

	for _, line := range lines {
		tokens := strings.Split(line, " ")

		if len(tokens) < 3 {
			continue
		}

		source := tokens[0]
		destination := tokens[1]

		if !strings.HasPrefix(source, "/dev/rbd") {
			continue
		}

		if !strings.HasPrefix(destination, destinationPrefix) {
			continue
		}

		if !strings.Contains(destination, "globalmount") {
			continue
		}

		mounts = append(mounts, destination)
	}

	return mounts
}

func Exists(path string) bool {
	if _, error := os.Stat(path); !os.IsNotExist(error) {
		return true
	}

	return false
}

func Unmount(path string) error {
	if Exists(path) {
		return unix.Unmount(path, 0)
	}

	return nil
}

func KillContainers(_config *config.InternalConfig) {
	// TODO
	address := "/run/containerd/containerd.sock"

	dialer := func(address string, timeout time.Duration) (net.Conn, error) {
		return net.DialTimeout("unix", address, timeout)
	}

	connection, _error := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second), grpc.WithDialer(dialer))
	if _error != nil {
		log.WithFields(log.Fields{"error": _error, "address": address}).Debug("CRI dial failed")

		return
	}

	runtimeClient := cri.NewRuntimeServiceClient(connection)

	filter := &cri.PodSandboxFilter{}

	stateValue := &cri.PodSandboxStateValue{}
	stateValue.State = cri.PodSandboxState_SANDBOX_READY
	filter.State = stateValue

	request := &cri.ListPodSandboxRequest{Filter: filter}

	response, _error := runtimeClient.ListPodSandbox(context.Background(), request)
	if _error != nil {
		log.WithFields(log.Fields{"error": _error}).Debug("CRI pods list failed")

		return
	}

	getRemovalPriority := func(namespace string, name string) int {
		if namespace == "kube-system" {
			if strings.HasPrefix(name, "etcd-") {
				return 5
			}

			if strings.HasPrefix(name, "kube-apiserver") {
				return 4
			}

			return 3
		}

		if namespace == "networking" {
			return 2
		}

		if namespace == "storage" {
			return 1
		}

		return 0
	}

	items := response.GetItems()

	sort.Slice(items, func(i, j int) bool {
		iRemovalPriority := getRemovalPriority(items[i].Metadata.Namespace, items[i].Metadata.Name)
		jRemovalPriority := getRemovalPriority(items[j].Metadata.Namespace, items[j].Metadata.Name)

		if iRemovalPriority == jRemovalPriority {
			return items[i].Metadata.Name < items[j].Metadata.Name
		}

		return iRemovalPriority < jRemovalPriority
	})

	for _, mount := range getCSIGlobalMounts(_config.GetFullLocalAssetDirectory(utils.DirectoryKubeletPlugins)) {
		if _error = Unmount(mount); _error != nil {
			log.WithFields(log.Fields{"error": _error, "mount": mount}).Debug("Global unmount failed")
		} else {
			log.WithFields(log.Fields{"mount": mount}).Debug("Unmounted")
		}
	}

	spew.Config.Indent = "\t"
	spew.Config.DisableMethods = true
	spew.Config.DisablePointerMethods = true
	spew.Config.DisablePointerAddresses = true

	for _, entry := range items {
		containers, _error := runtimeClient.ListContainers(context.Background(), &cri.ListContainersRequest{Filter: &cri.ContainerFilter{PodSandboxId: entry.Id}})
		if _error != nil {
			log.WithFields(log.Fields{"error": _error, "id": entry.Id, "namespace": entry.Metadata.Namespace, "name": entry.Metadata.Name}).Debug("CRI containers list failed")

			continue
		}

		mounts := map[string]bool{}

		for _, container := range containers.Containers {
			containerStatus, _error := runtimeClient.ContainerStatus(context.Background(), &cri.ContainerStatusRequest{ContainerId: container.Id, Verbose: true})
			if _error != nil {
				log.WithFields(log.Fields{"error": _error, "id": entry.Id, "namespace": entry.Metadata.Namespace, "name": entry.Metadata.Name, "container-id": container.Id}).Debug("CRI container status failed")

				continue
			}

			for _, mount := range containerStatus.Status.Mounts {
				if (strings.Contains(mount.HostPath, "volumes/kubernetes.io~") || strings.Contains(mount.HostPath, "volume-subpaths")) && !strings.Contains(mount.HostPath, "~configmap") {
					mounts[mount.HostPath] = true
				}

			}
		}

		_, _error = runtimeClient.StopPodSandbox(context.Background(), &cri.StopPodSandboxRequest{PodSandboxId: entry.Id})
		if _error != nil {
			log.WithFields(log.Fields{"error": _error, "id": entry.Id, "namespace": entry.Metadata.Namespace, "name": entry.Metadata.Name}).Debug("CRI pod stop failed")

			continue
		}

		_, _error = runtimeClient.RemovePodSandbox(context.Background(), &cri.RemovePodSandboxRequest{PodSandboxId: entry.Id})
		if _error != nil {
			log.WithFields(log.Fields{"error": _error, "id": entry.Id, "namespace": entry.Metadata.Namespace, "name": entry.Metadata.Name}).Debug("CRI pod remove failed")

			continue
		}

		for mount := range mounts {
			if _error = Unmount(mount); _error != nil {
				log.WithFields(log.Fields{"error": _error, "id": entry.Id, "namespace": entry.Metadata.Namespace, "name": entry.Metadata.Name, "mount": mount}).Error("Unmount failed")
			} else {
				log.WithFields(log.Fields{"id": entry.Id, "namespace": entry.Metadata.Namespace, "name": entry.Metadata.Name, "mount": mount}).Debug("Unmounted")
			}
		}

		log.WithFields(log.Fields{"id": entry.Id, "namespace": entry.Metadata.Namespace, "name": entry.Metadata.Name}).Debug("CRI pod removed")
	}
}
