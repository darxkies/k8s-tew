package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/darxkies/k8s-tew/pkg/utils"

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
	if error := bootstrap(false); error != nil {
		return error
	}

	utils.SetProgressSteps(1)

	var error error

	labels := []string{}

	for _, label := range strings.Split(nodeLabels, ",") {
		labels = append(labels, strings.Trim(label, "\n "))
	}

	if nodeSelf {
		log.Println("Adding self as node")

		// Get ip of the node
		nodeIP, error = utils.RunCommandWithOutput("ip route get 8.8.8.8 | cut -d ' ' -f 7")
		if error != nil {
			return error
		}

		// Parse the ip
		nodeIP = strings.Trim(nodeIP, "\n")

		// Throw error if the ip could not be retrieved
		if len(nodeIP) == 0 {
			return errors.New("Could not find own ip")
		}

		// Set name of the node
		nodeName, error = os.Hostname()
		if error != nil {
			return error
		}

		// Set labels
		labels = []string{utils.NodeBootstrapper, utils.NodeController, utils.NodeWorker}

		// Get public network settings
		network, error := utils.RunCommandWithOutput(fmt.Sprintf("ip address | grep %s | cut -d ' ' -f 6", nodeIP))
		if error != nil {
			return error
		}

		// Set public network settings
		_config.Config.PublicNetwork = network

		// Set deployment directory by assigning the base directory
		_config.Config.DeploymentDirectory = _config.BaseDirectory
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
			log.WithFields(log.Fields{"error": error}).Error("Failed to add node")

			os.Exit(-1)
		}
	},
}

func init() {
	nodeAddCmd.Flags().StringVarP(&nodeName, "name", "n", "single-node", "The hostname of the node")
	nodeAddCmd.Flags().StringVarP(&nodeIP, "ip", "i", "192.168.100.50", "IP of the node")
	nodeAddCmd.Flags().UintVarP(&nodeIndex, "index", "x", 0, "The unique index of the node which should never be reused")
	nodeAddCmd.Flags().StringVarP(&nodeLabels, "labels", "l", fmt.Sprintf("%s,%s", utils.NodeController, utils.NodeWorker), "The labels of the node which define the attributes of the node")
	nodeAddCmd.Flags().BoolVarP(&nodeSelf, "self", "s", false, "Add this machine by infering the host's name & ip and by setting the labels controller,worker,bootstrapper - The public-network and the deployment-directory are also updated")
	RootCmd.AddCommand(nodeAddCmd)
}
