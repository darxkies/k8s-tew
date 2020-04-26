package main

import (
	"os"

	"github.com/darxkies/k8s-tew/pkg/container"
	"github.com/darxkies/k8s-tew/pkg/servers"
	"github.com/darxkies/k8s-tew/pkg/utils"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var killContainers bool

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run",
	Long:  "Run servers",
	Run: func(cmd *cobra.Command, args []string) {
		if error := bootstrap(true); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Failed initialization")

			os.Exit(-1)
		}

		if len(_config.Config.Nodes) == 0 {
			log.WithFields(log.Fields{"error": "no nodes defined"}).Error("Failed to run")

			os.Exit(-1)
		}

		if _config.Node == nil {
			log.WithFields(log.Fields{"error": "current host not found in the list of nodes"}).Error("Failed to run")

			os.Exit(-1)
		}

		serversContainer := servers.NewServers(_config)

		utils.SetProgressSteps(serversContainer.Steps())

		utils.ShowProgress()

		if error := serversContainer.Run(commandRetries, func() {
			if killContainers {
				container.KillContainers(_config)
			}

		}); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Failed to run")

			os.Exit(-1)
		}

	},
}

func init() {
	runCmd.Flags().UintVarP(&commandRetries, "command-retries", "r", 300, "The count of command retries")
	runCmd.Flags().BoolVarP(&killContainers, "kill-containers", "k", true, "Kill containers when shutting down")
	RootCmd.AddCommand(runCmd)
}
