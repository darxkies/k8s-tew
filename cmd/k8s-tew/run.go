package main

import (
	"os"

	"github.com/darxkies/k8s-tew/servers"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run",
	Long:  "Run servers",
	Run: func(cmd *cobra.Command, args []string) {
		if error := Bootstrap(true); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("run failed")

			os.Exit(-1)
		}

		if len(_config.Config.Nodes) == 0 {
			log.WithFields(log.Fields{"error": "no nodes defined"}).Error("run failed")

			os.Exit(-1)
		}

		if _config.Node == nil {
			log.WithFields(log.Fields{"error": "current host not found in the list of nodes"}).Error("run failed")

			os.Exit(-1)
		}

		serversContainer := servers.NewServers(_config)

		if error := serversContainer.Run(); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("run failed")

			os.Exit(-1)
		}
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
}
