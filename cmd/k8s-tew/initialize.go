package main

import (
	"os"

	"github.com/darxkies/k8s-tew/config"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var force bool

var initializeCmd = &cobra.Command{
	Use:   "initialize",
	Short: "Initialize the configuration",
	Long:  "Initialize the configuration",
	Run: func(cmd *cobra.Command, args []string) {
		_config = config.DefaultInternalConfig(baseDirectory)

		if !force {
			if error := _config.Load(); error == nil {
				log.WithFields(log.Fields{"error": "already initialized"}).Error("initialize failed")

				os.Exit(-1)
			}
		}

		if error := _config.Save(); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("initialize failed")

			os.Exit(-1)
		}

		log.Info("initialized")
	},
}

func init() {
	initializeCmd.Flags().BoolVarP(&force, "force", "f", false, "Force initialization if already initialized. This will remove any config changes.")
	RootCmd.AddCommand(initializeCmd)
}
