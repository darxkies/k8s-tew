package config

import "github.com/darxkies/k8s-tew/pkg/utils"

type Node struct {
	IP           string `yaml:"ip"`
	Index        uint   `yaml:"index"`
	StorageIndex uint   `yaml:"storage-index",omitempty`
	Labels       Labels `yaml:"labels"`
}

type Nodes map[string]*Node

func (nodes Nodes) HasControllerNode() bool {
	for _, node := range nodes {
		if node.IsController() {
			return true
		}
	}

	return false
}

func (nodes Nodes) HasWorkerNode() bool {
	for _, node := range nodes {
		if node.IsWorker() {
			return true
		}
	}

	return false
}

func (nodes Nodes) HasStorageNode() bool {
	for _, node := range nodes {
		if node.IsStorage() {
			return true
		}
	}

	return false
}

func NewNode(ip string, index, storageIndex uint, labels []string) *Node {
	return &Node{IP: ip, Index: index, StorageIndex: storageIndex, Labels: labels}
}

func (node *Node) IsController() bool {
	for _, label := range node.Labels {
		if label == utils.NodeController {
			return true
		}
	}

	return false
}

func (node *Node) IsWorker() bool {
	for _, label := range node.Labels {
		if label == utils.NodeWorker {
			return true
		}
	}

	return false
}

func (node *Node) IsStorage() bool {
	for _, label := range node.Labels {
		if label == utils.NodeStorage {
			return true
		}
	}

	return false
}

func (node *Node) IsControllerOnly() bool {
	return node.IsController() && !node.IsWorker() && !node.IsStorage()
}

func (node *Node) IsStorageOnly() bool {
	return !node.IsController() && !node.IsWorker() && node.IsStorage()
}

func (node *Node) IsWorkerOnly() bool {
	return !node.IsController() && node.IsWorker() && !node.IsStorage()
}
