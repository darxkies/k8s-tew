package main

import (
	"os"

	"github.com/darxkies/k8s-tew/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var removeNodeName string

func removeNode() error {
	// Load config and check the rights
	if error := bootstrap(false); error != nil {
		return error
	}

	utils.SetProgressSteps(1)

	if error := _config.RemoveNode(removeNodeName); error != nil {
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
			log.WithFields(log.Fields{"error": error}).Error("Failed to remove node")

			os.Exit(-1)
		}
	},
}

func init() {
	nodeRemoveCmd.Flags().StringVarP(&removeNodeName, "name", "n", "", "Unique name of the node")
	RootCmd.AddCommand(nodeRemoveCmd)
}
