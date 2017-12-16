package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func removeNode() error {
	// Load config and check the rights
	if error := Bootstrap(false); error != nil {
		return error
	}

	if error := _config.RemoveNode(nodeName); error != nil {
		return error
	}

	if error := _config.Save(); error != nil {
		return error
	}

	return nil
}

var nodeRemoveCmd = &cobra.Command{
	Use:   "node-remove",
	Short: "Remove a node",
	Long:  "Remove a node",
	Run: func(cmd *cobra.Command, args []string) {
		if error := removeNode(); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("node-remove failed")

			os.Exit(-1)
		}
	},
}

func init() {
	nodeRemoveCmd.Flags().StringVarP(&nodeName, "name", "n", "", "Unique name of the node")
	RootCmd.AddCommand(nodeRemoveCmd)
}
