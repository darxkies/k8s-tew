package container

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/darxkies/k8s-tew/pkg/config"
	"github.com/darxkies/k8s-tew/pkg/utils"
	log "github.com/sirupsen/logrus"
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
		for _, parameter := range tokens[1:] {
			if strings.HasPrefix(parameter, workdirPrefix) {
				workdirTokens := strings.Split(parameter, "/")
				containerID = workdirTokens[len(workdirTokens)-1]

				break
			}
		}

		if len(containerID) == 0 {
			log.WithFields(log.Fields{"error": error, "cmdline": cmdline}).Debug("ContainerD-Shim container ID retrieval failed")

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
	filename := fmt.Sprintf("/run/k8s-tew/containerd/io.containerd.runtime.v1.linux/k8s.io/%s/config.json", container.containerID)
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
		return syscall.Unmount(path, syscall.MNT_DETACH)
	}

	return nil
}

func KillContainers(_config *config.InternalConfig) {
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

	for _, mount := range *mounts {
		log.WithFields(log.Fields{"path": mount.Destination}).Debug("Unmounting path")

		if error := Unmount(mount.Destination); error != nil {
			log.WithFields(log.Fields{"error": error, "path": mount.Destination}).Debug("Unmount failed")
		}
	}
}
