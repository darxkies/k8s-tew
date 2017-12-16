package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func listNodes() error {
	// Load config and check the rights
	if error := Bootstrap(false); error != nil {
		return error
	}

	for name, node := range _config.Config.Nodes {
		log.WithFields(log.Fields{"index": node.Index, "name": name, "ip": node.IP, "labels": node.Labels}).Info("node")
	}

	return nil
}

var nodeListCmd = &cobra.Command{
	Use:   "node-list",
	Short: "List nodes",
	Long:  "List nodes",
	Run: func(cmd *cobra.Command, args []string) {
		if error := listNodes(); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("node-list failed")

			os.Exit(-1)
		}
	},
}

func init() {
	RootCmd.AddCommand(nodeListCmd)
}
