package main

import (
	"os"

	"github.com/darxkies/k8s-tew/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func listNodes() error {
	// Load config and check the rights
	if error := Bootstrap(false); error != nil {
		return error
	}

	utils.SetProgressSteps(1)

	for name, node := range _config.Config.Nodes {
		log.WithFields(log.Fields{"index": node.Index, "name": name, "ip": node.IP, "labels": node.Labels}).Info("Node")
	}

	return nil
}

var nodeListCmd = &cobra.Command{
	Use:   "node-list",
	Short: "List nodes",
	Long:  "List nodes",
	Run: func(cmd *cobra.Command, args []string) {
		if error := listNodes(); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Failed to list nodes")

			os.Exit(-1)
		}
	},
}

func init() {
	RootCmd.AddCommand(nodeListCmd)
}
