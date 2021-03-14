package container

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"context"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/darxkies/k8s-tew/pkg/config"
	"github.com/darxkies/k8s-tew/pkg/utils"
	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
)

type Pods struct {
	config *config.InternalConfig
}

func NewPods(config *config.InternalConfig) *Pods {
	return &Pods{config: config}
}

func (pods *Pods) saveConfig(filename string, value map[string]string) error {
	config, _error := pods.extractConfig(value)
	if _error != nil {
		return _error
	}

	if _error := ioutil.WriteFile(filename, []byte(config), 0644); _error != nil {
		return errors.Wrapf(_error, "Could not write to '%s'", filename)
	}

	return nil
}

func (pods *Pods) extractConfig(value map[string]string) (string, error) {
	info, ok := value["info"]

	if !ok {
		return "", errors.New("Could not extract CRI info block")
	}

	var response map[string]interface{}

	if _error := json.Unmarshal([]byte(info), &response); _error != nil {
		return "", errors.Wrapf(_error, "Could not deserialize CRI info block")
	}

	config, ok := response["config"]

	if !ok {
		return "", errors.New("Could not extract CRI config block")
	}

	result, _error := json.Marshal(config)
	if _error != nil {
		return "", errors.Wrap(_error, "Could not serialize config")
	}

	return string(result), nil
}

func (pods *Pods) podsDirectory() string {
	return path.Join(pods.config.GetFullLocalAssetDirectory(utils.DirectoryDynamicData), "pods")
}

func (pods *Pods) podDirectory(podID string) string {
	return path.Join(pods.podsDirectory(), podID)
}

func (pods *Pods) containersDirectory(podID string) string {
	return path.Join(pods.podDirectory(podID), "containers")
}

func (pods *Pods) containerDirectory(podID string, containerID string) string {
	return path.Join(pods.containersDirectory(podID), containerID)
}

func (pods *Pods) podFilename(podID string) string {
	return path.Join(pods.podDirectory(podID), "pod.json")
}

func (pods *Pods) containerFilename(podID, containerID string) string {
	return path.Join(pods.containerDirectory(podID, containerID), "container.json")
}

func (pods *Pods) SavePodConfig(podID string, info map[string]string) error {
	directory := pods.podDirectory(podID)

	if _error := utils.CreateDirectoryIfMissing(directory); _error != nil {
		return _error
	}

	configFilename := pods.podFilename(podID)

	if _error := pods.saveConfig(configFilename, info); _error != nil {
		return _error
	}

	return nil
}

func (pods *Pods) SaveContainerConfig(podID string, containerID string, info map[string]string) error {
	directory := pods.containerDirectory(podID, containerID)

	if _error := utils.CreateDirectoryIfMissing(directory); _error != nil {
		return _error
	}

	configFilename := pods.containerFilename(podID, containerID)

	if _error := pods.saveConfig(configFilename, info); _error != nil {
		return _error
	}

	return nil
}

func (pods *Pods) cleanUp() {
	os.RemoveAll(pods.podsDirectory())
}

func (pods *Pods) getDirectoryEntries(directory string) ([]string, error) {
	result := []string{}

	files, _error := ioutil.ReadDir(directory)
	if _error != nil {
		return nil, _error
	}

	for _, file := range files {
		result = append(result, file.Name())
	}

	return result, nil
}

func (pods *Pods) Restore() {
	for {
		runtimeClient, _error := pods.getCRIClient()
		if _error != nil {
			log.WithFields(log.Fields{"error": _error}).Debug("CRI dial failed")

			time.Sleep(time.Second)

			continue
		}

		podIDs, _error := pods.getDirectoryEntries(pods.podsDirectory())
		if _error != nil {
			log.WithFields(log.Fields{"error": _error}).Debug("Could not retrieve pod ids")

			break
		}

		var waitGroup sync.WaitGroup

		for _, _podID := range podIDs {
			waitGroup.Add(1)

			go func(podID string) {
				restore := func() {
					logMessage := log.WithFields(log.Fields{"pod-id": podID})

					logMessage.Debug("Restoring pod")

					podFilename := pods.podFilename(podID)

					podContent, _error := ioutil.ReadFile(podFilename)
					if _error != nil {
						logMessage.WithFields(log.Fields{"error": _error, "filename": podFilename}).Debug("Could not read pod content")

						return
					}

					var podConfig cri.PodSandboxConfig

					if _error = json.Unmarshal(podContent, &podConfig); _error != nil {
						logMessage.WithFields(log.Fields{"error": _error, "filename": podFilename}).Debug("Could not deserialize pod content")

						return
					}

					podResponse, _error := runtimeClient.RunPodSandbox(context.Background(), &cri.RunPodSandboxRequest{Config: &podConfig})
					if _error != nil {
						logMessage.WithFields(log.Fields{"error": _error}).Debug("Could not start sandbox")

						return
					}

					containerIDs, _error := pods.getDirectoryEntries(pods.containersDirectory(podID))
					if _error != nil {
						logMessage.WithFields(log.Fields{"error": _error}).Debug("Could not retrieve pod ids")

						return
					}

					for _, containerID := range containerIDs {
						containerMessage := logMessage.WithFields(log.Fields{"container-id": containerID})

						containerFilename := pods.containerFilename(podID, containerID)

						containerContent, _error := ioutil.ReadFile(containerFilename)
						if _error != nil {
							containerMessage.WithFields(log.Fields{"error": _error, "filename": containerFilename}).Debug("Could not read container content")

							continue
						}

						var containerConfig cri.ContainerConfig

						if _error = json.Unmarshal(containerContent, &containerConfig); _error != nil {
							containerMessage.WithFields(log.Fields{"error": _error, "filename": containerFilename}).Debug("Could not deserialize container content")

							continue
						}

						containerResponse, _error := runtimeClient.CreateContainer(context.Background(), &cri.CreateContainerRequest{PodSandboxId: podResponse.PodSandboxId, Config: &containerConfig, SandboxConfig: &podConfig})
						if _error != nil {
							containerMessage.WithFields(log.Fields{"error": _error}).Debug("Could not create container")

							continue
						}

						_, _error = runtimeClient.StartContainer(context.Background(), &cri.StartContainerRequest{ContainerId: containerResponse.ContainerId})
						if _error != nil {
							containerMessage.WithFields(log.Fields{"error": _error}).Debug("Could not start  container")

							continue
						}

						logMessage.Debug("Restored container")
					}

					logMessage.Debug("Restored pod")
				}

				restore()

				waitGroup.Done()
			}(_podID)
		}

		waitGroup.Wait()

		break
	}
}

func (pods *Pods) getCRIClient() (cri.RuntimeServiceClient, error) {
	address := pods.config.GetFullTargetAssetFilename(utils.ContainerdSock)

	dialer := func(address string, timeout time.Duration) (net.Conn, error) {
		return net.DialTimeout("unix", address, timeout)
	}

	connection, _error := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second), grpc.WithDialer(dialer))
	if _error != nil {
		return nil, errors.Wrapf(_error, "Could not connect to '%s'", address)
	}

	runtimeClient := cri.NewRuntimeServiceClient(connection)

	return runtimeClient, nil
}

func (pods *Pods) Kill() {
	pods.cleanUp()

	spew.Config.Indent = "\t"
	spew.Config.DisableMethods = true
	spew.Config.DisablePointerMethods = true
	spew.Config.DisablePointerAddresses = true

	runtimeClient, _error := pods.getCRIClient()
	if _error != nil {
		log.WithFields(log.Fields{"error": _error}).Debug("CRI dial failed")

		return
	}

	filter := &cri.PodSandboxFilter{}

	stateValue := &cri.PodSandboxStateValue{}
	stateValue.State = cri.PodSandboxState_SANDBOX_READY
	filter.State = stateValue

	request := &cri.ListPodSandboxRequest{Filter: filter}

	// Get list of pods
	response, _error := runtimeClient.ListPodSandbox(context.Background(), request)
	if _error != nil {
		log.WithFields(log.Fields{"error": _error}).Debug("CRI pods list failed")

		return
	}

	getRemovalPriority := func(namespace string, name string, labels map[string]string) int {
		if value, ok := labels[utils.ClusterWeight]; ok {
			result, _error := strconv.ParseInt(value, 10, 64)
			if _error == nil {
				return int(result)
			}

			log.WithFields(log.Fields{"error": _error, "namespace": namespace, "pod": name}).Error("Cluster weight error")
		}

		return 0
	}

	hasClusterCache := func(labels map[string]string) bool {
		if _, ok := labels[utils.ClusterCache]; ok {
			return true
		}

		return false
	}

	podsList := response.GetItems()

	// Sort containers by their removal priority
	sort.Slice(podsList, func(i, j int) bool {
		iRemovalPriority := getRemovalPriority(podsList[i].Metadata.Namespace, podsList[i].Metadata.Name, podsList[i].Labels)
		jRemovalPriority := getRemovalPriority(podsList[j].Metadata.Namespace, podsList[j].Metadata.Name, podsList[j].Labels)

		if iRemovalPriority == jRemovalPriority {
			return podsList[i].Metadata.Name < podsList[j].Metadata.Name
		}

		return iRemovalPriority < jRemovalPriority
	})

	// Unmount CSI mount points
	for _, mount := range utils.GetCSIGlobalMounts(pods.config.GetFullLocalAssetDirectory(utils.DirectoryKubeletPlugins)) {
		if _error = utils.Unmount(mount); _error != nil {
			log.WithFields(log.Fields{"error": _error, "mount": mount}).Debug("Global unmount failed")
		} else {
			log.WithFields(log.Fields{"mount": mount}).Debug("Unmounted")
		}
	}

	for _, pod := range podsList {
		_hasClusterCache := hasClusterCache(pod.Labels)

		logMessage := log.WithFields(log.Fields{"pod-id": pod.Id, "namespace": pod.Metadata.Namespace, "name": pod.Metadata.Name})

		if _hasClusterCache {
			// Get Pod status
			response, _error := runtimeClient.PodSandboxStatus(context.Background(), &cri.PodSandboxStatusRequest{PodSandboxId: pod.Id, Verbose: true})
			if _error != nil {
				logMessage.WithFields(log.Fields{"error": _error}).Debug("CRI pod status failed")

			} else if _error := pods.SavePodConfig(pod.Id, response.Info); _error != nil {
				logMessage.WithFields(log.Fields{"error": _error}).Debug("CRI pod dump failed")
			}
		}

		logMessage.Debug("Listing pod containers with timeout")

		// ContainerStatus returns status of the container.
		containers, _error := runtimeClient.ListContainers(context.Background(), &cri.ListContainersRequest{Filter: &cri.ContainerFilter{PodSandboxId: pod.Id}})
		if _error != nil {
			logMessage.WithFields(log.Fields{"error": _error}).Debug("CRI containers list failed")

			continue
		}

		mounts := map[string]bool{}

		// Collect mounts points
		for _, container := range containers.Containers {
			containerLogMessage := logMessage.WithFields(log.Fields{"container-id": container.Id})

			containerLogMessage.WithFields(log.Fields{"pod-id": pod.Id, "namespace": pod.Metadata.Namespace, "name": pod.Metadata.Name, "container-id": container.Id}).Debug("CRI container status query")

			_context, _ := context.WithTimeout(context.Background(), time.Second)

			containerStatus, _error := runtimeClient.ContainerStatus(_context, &cri.ContainerStatusRequest{ContainerId: container.Id, Verbose: true})
			if _error != nil {
				containerLogMessage.WithFields(log.Fields{"error": _error}).Debug("CRI container status failed")

				continue

			} else if _hasClusterCache {
				if _error := pods.SaveContainerConfig(pod.Id, container.Id, containerStatus.Info); _error != nil {
					logMessage.WithFields(log.Fields{"error": _error}).Debug("CRI pod dump failed")
				}
			}

			for _, mount := range containerStatus.Status.Mounts {
				if (strings.Contains(mount.HostPath, "volumes/kubernetes.io~") || strings.Contains(mount.HostPath, "volume-subpaths")) && !strings.Contains(mount.HostPath, "~configmap") {
					mounts[mount.HostPath] = true
				}
			}
		}

		// Stop Pod
		_, _error = runtimeClient.StopPodSandbox(context.Background(), &cri.StopPodSandboxRequest{PodSandboxId: pod.Id})
		if _error != nil {
			logMessage.WithFields(log.Fields{"error": _error}).Debug("CRI pod stop failed")

			continue
		}

		// Remove Pod
		_, _error = runtimeClient.RemovePodSandbox(context.Background(), &cri.RemovePodSandboxRequest{PodSandboxId: pod.Id})
		if _error != nil {
			logMessage.WithFields(log.Fields{"error": _error}).Debug("CRI pod remove failed")

			continue
		}

		// Remove left over mounts
		for mount := range mounts {
			if _error = utils.Unmount(mount); _error != nil {
				logMessage.WithFields(log.Fields{"error": _error, "mount": mount}).Error("Unmount failed")
			} else {
				logMessage.WithFields(log.Fields{"mount": mount}).Debug("Unmounted")
			}
		}

		logMessage.Debug("CRI pod removed")
	}
}
