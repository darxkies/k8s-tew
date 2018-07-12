package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/darxkies/k8s-tew/utils"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var nodeName string
var nodeIP string
var nodeIndex uint
var nodeLabels string
var nodeSelf bool

func addNode() error {
	// Load config and check the rights
	if error := Bootstrap(false); error != nil {
		return error
	}

	var error error

	labels := []string{}

	for _, label := range strings.Split(nodeLabels, ",") {
		labels = append(labels, strings.Trim(label, "\n "))
	}

	if nodeSelf {
		log.Println("adding self as node")

		if len(nodeIP) == 0 {
			nodeIP, error = utils.RunCommandWithOutput("ip route get 8.8.8.8 | cut -d ' ' -f 7")
			if error != nil {
				return error
			}

			nodeIP = strings.Trim(nodeIP, "\n")

			if len(nodeIP) == 0 {
				return errors.New("Could not find own ip")
			}
		}

		if len(nodeName) == 0 {
			nodeName, error = os.Hostname()
			if error != nil {
				return error
			}
		}

		if len(labels) == 0 {
			labels = []string{utils.NODE_BOOTSTRAPPER, utils.NODE_CONTROLLER, utils.NODE_WORKER}
		}

		network, error := utils.RunCommandWithOutput(fmt.Sprintf("ip address | grep %s | cut -d ' ' -f 6", nodeIP))
		if error != nil {
			return error
		}

		_config.Config.PublicNetwork = network
	}

	if _, error = _config.AddNode(nodeName, nodeIP, nodeIndex, labels); error != nil {
		return error
	}

	if error := _config.Save(); error != nil {
		return error
	}

	return nil
}

var nodeAddCmd = &cobra.Command{
	Use:   "node-add",
	Short: "Add or update a node",
	Long:  "Add a node. This can be also called when updating a node, only the name has to be unique.",
	Run: func(cmd *cobra.Command, args []string) {
		if error := addNode(); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("node-add failed")

			os.Exit(-1)
		}
	},
}

func init() {
	nodeAddCmd.Flags().StringVarP(&nodeName, "name", "n", "", "Unique name of the node")
	nodeAddCmd.Flags().StringVarP(&nodeIP, "ip", "i", "", "IP of the node")
	nodeAddCmd.Flags().UintVarP(&nodeIndex, "index", "x", 0, "The unique index of the node.")
	nodeAddCmd.Flags().StringVarP(&nodeLabels, "labels", "l", fmt.Sprintf("%s,%s,%s", utils.NODE_BOOTSTRAPPER, utils.NODE_CONTROLLER, utils.NODE_WORKER), "The labels of the node which define the attributes of the node.")
	nodeAddCmd.Flags().BoolVarP(&nodeSelf, "self", "s", false, "Add this machine by infering the name, the ip and assuming it is a controller and a worker")
	RootCmd.AddCommand(nodeAddCmd)
}
