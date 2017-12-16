package main

import (
	"os"

	"github.com/darxkies/k8s-tew/config"
	"github.com/darxkies/k8s-tew/utils"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var initializeCmd = &cobra.Command{
	Use:   "initialize",
	Short: "Initialize the configuration",
	Long:  "Initialize the configuration",
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.IsRoot() {
			log.WithFields(log.Fields{"error": "this program needs root rights"}).Error("nitialize failed")

			os.Exit(-1)
		}

		_config = config.DefaultInternalConfig()

		if error := _config.Load(baseDirectory); error == nil {
			log.WithFields(log.Fields{"error": "already initialized"}).Error("initialize failed")

			os.Exit(-1)
		}

		if error := _config.Save(); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("initialize failed")

			os.Exit(-1)
		}
	},
}

func init() {
	RootCmd.AddCommand(initializeCmd)
}
