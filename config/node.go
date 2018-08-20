package config

import "github.com/darxkies/k8s-tew/utils"

type Node struct {
	IP     string `yaml:"ip"`
	Index  uint   `yaml:"index"`
	Labels Labels `yaml:"labels"`
}

type Nodes map[string]*Node

func NewNode(ip string, index uint, labels []string) *Node {
	return &Node{IP: ip, Index: index, Labels: labels}
}

func (node *Node) IsController() bool {
	for _, label := range node.Labels {
		if label == utils.NODE_CONTROLLER {
			return true
		}
	}

	return false
}

func (node *Node) IsWorker() bool {
	for _, label := range node.Labels {
		if label == utils.NODE_WORKER {
			return true
		}
	}

	return false
}

func (node *Node) IsControllerOnly() bool {
	return node.IsController() && !node.IsWorker()
}
