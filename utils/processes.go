package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

type Process struct {
	ParentProcessID int
	ProcessID       int
	Name            string
	Children        map[int]*Process
}

type Processes struct {
	processes map[int]*Process
}

type Children []*Process

func NewProcesses() *Processes {
	return &Processes{}
}

func (processes *Processes) updateProcessList() error {
	processes.processes = map[int]*Process{}

	statFiles, error := filepath.Glob("/proc/*/stat")
	if error != nil {
		return error
	}

	for _, statFile := range statFiles {
		process := &Process{Children: map[int]*Process{}}

		tokens := strings.Split(statFile, "/")

		// Skip if the stat file name does not contain the PID
		if process.ProcessID, error = strconv.Atoi(tokens[2]); error != nil {
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

		process.Name = tokens[1]

		if process.ParentProcessID, error = strconv.Atoi(tokens[3]); error != nil {
			continue
		}

		processes.processes[process.ProcessID] = process
	}

	// Process Children
	for _, process := range processes.processes {
		parentProcess, ok := processes.processes[process.ParentProcessID]

		// Parent does not exist, so add it
		if !ok {
			parentProcess = &Process{Children: map[int]*Process{}, ParentProcessID: process.ParentProcessID, ProcessID: process.ParentProcessID}

			processes.processes[process.ParentProcessID] = parentProcess
		}

		parentProcess.Children[process.ProcessID] = process
	}

	return nil
}

func (processes *Processes) appendChildren(process *Process, children *Children) {
	keys := []int{}

	for pid := range process.Children {
		keys = append(keys, pid)
	}

	sort.Ints(keys)

	for _, pid := range keys {
		child := process.Children[pid]

		log.WithFields(log.Fields{"pid": pid, "name": child.Name}).Debug("Add child")

		*children = append(*children, child)

		processes.appendChildren(child, children)
	}
}

func (processes *Processes) GetAllChildrenByParent(pid int) *Children {
	children := &Children{}

	if pid < 1 {
		return children
	}

	log.WithFields(log.Fields{"pid": pid}).Debug("Getting process children")

	if error := processes.updateProcessList(); error != nil {
		log.Error("Process update failed")

		return children
	}

	process, ok := processes.processes[pid]
	if !ok {
		return children
	}

	processes.appendChildren(process, children)

	return children
}

func (children *Children) sendSignal(pid int, signal os.Signal) error {
	process, error := os.FindProcess(pid)
	if error != nil {
		return fmt.Errorf("Process '%d' not found (%s)", pid, error.Error())
	}

	return process.Signal(signal)
}

func (children *Children) Kill(killTimeout uint) {
	log.Info("Cleaning up children")

	for _, child := range *children {
		log.WithFields(log.Fields{"name": child.Name, "pid": child.ProcessID}).Debug("Stopping process")

		_ = children.sendSignal(child.ProcessID, syscall.SIGINT)
	}

	time.Sleep(time.Duration(killTimeout) * time.Second)

	for _, child := range *children {
		log.WithFields(log.Fields{"name": child.Name, "pid": child.ProcessID}).Debug("Killing process")

		_ = children.sendSignal(child.ProcessID, syscall.SIGKILL)
	}

	log.Info("Cleaned up children")
}

func KillProcessChildren(pid int, timeout uint) {
	processes := NewProcesses()

	children := processes.GetAllChildrenByParent(os.Getpid())

	children.Kill(timeout)
}
