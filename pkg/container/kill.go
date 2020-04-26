package container

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/darxkies/k8s-tew/pkg/config"
	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
)

type Container struct {
	containerID string
	processIDs  []int
	bindMounts  []string
	hasPV       bool
}

type Containers []*Container

type Mount struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Type        string `json:"type"`
}

type Mounts []*Mount

type ContainerConfig struct {
	Mounts Mounts `json:"mounts"`
}

func (mounts Mounts) Dump() {
	for _, mount := range mounts {
		log.WithFields(log.Fields{"source": mount.Source, "destination": mount.Destination, "type": mount.Type}).Debug("Mount")
	}

	for _, path := range mounts.getPV() {
		log.WithFields(log.Fields{"path": path}).Debug("PV Mount")
	}
}

func (mounts Mounts) getPV() []string {
	result := []string{}

	for _, mount := range mounts {
		if strings.HasPrefix(mount.Source, "/dev/rbd") {
			result = append(result, mount.Destination)
		}
	}

	return result
}

func (containers Containers) Dump() {
	for _, container := range containers {
		log.WithFields(log.Fields{"has-pv": container.hasPV, "container-id": container.containerID}).Debug("Countainer PV")

		for _, pid := range container.processIDs {
			log.WithFields(log.Fields{"pid": pid, "container-id": container.containerID}).Debug("Countainer PID")
		}

		for _, path := range container.bindMounts {
			log.WithFields(log.Fields{"path": path, "container-id": container.containerID}).Debug("Countainer Bind-Mount")
		}
	}
}

func getContainerdShim(containerdShimBinary, workdirPrefix string, pvPaths []string, processParentMap map[int]int) *Containers {
	containers := Containers{}

	cmdlines, error := filepath.Glob("/proc/*/cmdline")
	if error != nil {
		log.WithFields(log.Fields{"error": error}).Debug("ContainerD-Shim search failed")

		return &containers
	}

	for _, cmdline := range cmdlines {
		tokens := strings.Split(cmdline, "/")

		processID, error := strconv.Atoi(tokens[2])
		if error != nil {
			continue
		}

		cmdlineContent, error := ioutil.ReadFile(cmdline)
		if error != nil {
			log.WithFields(log.Fields{"error": error, "cmdline": cmdline}).Debug("ContainerD-Shim cmdline read failed")

			continue
		}

		content := string(cmdlineContent)

		tokens = strings.Split(content, "\000")

		if tokens[0] != containerdShimBinary {
			continue
		}

		containerID := ""
		idFound := false

		for _, parameter := range tokens[1:] {
			if parameter == "-id" {
				idFound = true

				continue
			}

			if idFound {
				containerID = parameter

				break
			}
		}

		if len(containerID) == 0 {
			log.WithFields(log.Fields{"cmdline": cmdline, "_tokens": tokens}).Debug("ContainerD-Shim container ID retrieval failed")

			continue
		}

		container := &Container{containerID: containerID, processIDs: []int{processID}}
		container.collectChildrenPIDs(processParentMap)
		container.collectBindMounts(pvPaths)

		containers = append(containers, container)
	}

	sort.Slice(containers, func(i, j int) bool {
		return (containers[i].hasPV == true && containers[j].hasPV == false) || (containers[i].hasPV == false && containers[j].hasPV == false && len(containers[i].bindMounts) > len(containers[j].bindMounts))
	})

	return &containers
}

func (container *Container) inProcessIDs(pid int) bool {
	for _, processID := range container.processIDs {
		if pid == processID {
			return true
		}
	}

	return false
}

func (container *Container) collectBindMounts(pvPaths []string) {
	filename := fmt.Sprintf("/run/k8s-tew/containerd/io.containerd.runtime.v2.task/k8s.io/%s/config.json", container.containerID)
	bytes, error := ioutil.ReadFile(filename)
	if error != nil {
		log.WithFields(log.Fields{"error": error, "pid": container.processIDs[0], "container-id": container.containerID}).Debug("Collect bind mounts failed")

		return
	}

	config := ContainerConfig{}
	if error := json.Unmarshal(bytes, &config); error != nil {
		log.WithFields(log.Fields{"error": error, "pid": container.processIDs[0], "container-id": container.containerID}).Debug("Collect bind unmarshal failed")

		return
	}

	container.bindMounts = []string{}

	for _, mount := range config.Mounts {
		if mount.Type != "bind" {
			continue
		}

		container.bindMounts = append(container.bindMounts, mount.Source)

		for _, pvPath := range pvPaths {
			if mount.Source == pvPath {
				container.hasPV = true
			}
		}

	}
}

func (container *Container) collectChildrenPIDs(processParentMap map[int]int) {
	for {
		changed := false

		for processID, parentProcessID := range processParentMap {
			if container.inProcessIDs(parentProcessID) && !container.inProcessIDs(processID) {
				changed = true

				container.processIDs = append(container.processIDs, processID)
			}
		}

		if !changed {
			break
		}
	}

	return
}

func getProcessParentMap() map[int]int {
	result := map[int]int{}

	statFiles, error := filepath.Glob("/proc/*/stat")
	if error != nil {
		log.WithFields(log.Fields{"error": error}).Debug("Process mapping failed")

		return result
	}

	for _, statFile := range statFiles {
		tokens := strings.Split(statFile, "/")

		// Skip if the stat file name does not contain the PID
		processID, error := strconv.Atoi(tokens[2])
		if error != nil {
			continue
		}

		// Read stat file
		statContent, error := ioutil.ReadFile(statFile)
		if error != nil {
			continue
		}

		content := string(statContent)

		tokens = strings.Split(content, " ")

		if len(tokens) < 4 {
			continue
		}

		parentProcessID, error := strconv.Atoi(tokens[3])
		if error != nil {
			log.WithFields(log.Fields{"error": error}).Debug("Process mapping failed")

			continue
		}

		result[processID] = parentProcessID

	}

	return result
}

func getMounts(mountPrefixes []string) *Mounts {
	mounts := Mounts{}

	bytes, error := ioutil.ReadFile("/proc/mounts")
	if error != nil {
		log.WithFields(log.Fields{"error": error}).Debug("Read mounts failed")

		return &mounts
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
		_type := tokens[2]

		hasPrefix := false
		for _, prefix := range mountPrefixes {
			if strings.HasPrefix(destination, prefix) {
				hasPrefix = true

				break
			}
		}

		if !hasPrefix {
			continue
		}

		mount := &Mount{Source: source, Destination: destination, Type: _type}

		mounts = append(mounts, mount)
	}

	return &mounts
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
	/*
		// TODO
		containerdSocketFilename := "/run/containerd/containerd.sock"

		client, error := containerd.New(containerdSocketFilename, containerd.WithDefaultNamespace(utils.ContainerdKubernetesNamespace))
		defer client.Close()

		if error != nil {
			log.WithFields(log.Fields{"error": error, "containerd-socket-filename": containerdSocketFilename}).Debug("Containerd connection failed")

			return
		}

		taskService := client.TaskService()

		context := context.Background()

		response, error := taskService.List(context, &tasks.ListTasksRequest{})
		if error != nil {
			log.WithFields(log.Fields{"error": error, "containerd-socket-filename": containerdSocketFilename}).Debug("Containerd tasks list failed")

			return
		}

		for _, task := range response.Tasks {
			fmt.Println(task)

			killRequest := &tasks.KillRequest{All: true, ContainerID: task.ID, Signal: uint32(syscall.SIGKILL)}

			_, error := taskService.Kill(context, killRequest)
			if error != nil {
				log.WithFields(log.Fields{"error": error, "task-id": task.ID}).Debug("Containerd task kill failed")
			}
		}
	*/

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

	// TODO dashboard does not get killed properly
	getRemovalPriority := func(namespace string, name string) int {
		if namespace == "kube-system" {
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

		unmounts := []string{}

		for _, container := range containers.Containers {
			containerStatus, _error := runtimeClient.ContainerStatus(context.Background(), &cri.ContainerStatusRequest{ContainerId: container.Id, Verbose: true})
			if _error != nil {
				log.WithFields(log.Fields{"error": _error, "id": entry.Id, "namespace": entry.Metadata.Namespace, "name": entry.Metadata.Name, "container-id": container.Id}).Debug("CRI container status failed")

				continue
			}

			for _, unmount := range containerStatus.Status.Mounts {
				if strings.Contains(unmount.HostPath, "volumes/kubernetes.io~") {
					unmounts = append(unmounts, unmount.HostPath)
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

		for _, unmount := range unmounts {
			_ = Unmount(unmount)
		}

		log.WithFields(log.Fields{"id": entry.Id, "namespace": entry.Metadata.Namespace, "name": entry.Metadata.Name}).Debug("CRI pod removed")
	}

	/*
		containerdShimBinary := _config.GetFullLocalAssetFilename(utils.BinaryContainerdShimRuncV2)

		workdirPrefix := _config.GetFullLocalAssetDirectory(utils.DirectoryDynamicData)
		mountPrefixes := []string{
			workdirPrefix,
			_config.GetFullLocalAssetDirectory(utils.DirectoryRun),
			_config.GetFullLocalAssetDirectory(utils.DirectoryVarRun),
			_config.GetFullLocalAssetDirectory(utils.DirectoryKubeletData),
		}

		processParentMap := getProcessParentMap()
		mounts := getMounts(mountPrefixes)
		pvPaths := mounts.getPV()
		containers := getContainerdShim(containerdShimBinary, workdirPrefix, pvPaths, processParentMap)

		containers.Dump()
		mounts.Dump()

		for _, mount := range *mounts {
			log.WithFields(log.Fields{"path": mount.Destination}).Debug("Unmounting path")

			if error := Unmount(mount.Destination); error != nil {
				log.WithFields(log.Fields{"error": error, "path": mount.Destination}).Debug("Unmount failed")
			}
		}

		for _, container := range *containers {
			log.WithFields(log.Fields{"container-id": container.containerID}).Debug("Killing container")

			for _, pid := range container.processIDs {
				log.WithFields(log.Fields{"container-id": container.containerID, "pid": pid}).Debug("Killing process")

				syscall.Kill(pid, syscall.SIGKILL)
			}

			for {
				found := false

				for _, pid := range container.processIDs {
					if Exists(fmt.Sprintf("/proc/%d", pid)) {
						found = true

						break
					}
				}

				if !found {
					break
				}
			}

			for _, path := range container.bindMounts {
				for _, pvPath := range pvPaths {
					if pvPath != path {
						continue
					}

					log.WithFields(log.Fields{"container-id": container.containerID, "path": path}).Debug("Unmounting path")

					if error := Unmount(path); error != nil {
						log.WithFields(log.Fields{"error": error, "path": path}).Debug("Unmount failed")
					}
				}
			}
		}
	*/
}
