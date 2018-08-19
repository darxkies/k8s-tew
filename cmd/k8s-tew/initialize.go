package main

import (
	"os"

	"github.com/darxkies/k8s-tew/config"
	"github.com/darxkies/k8s-tew/utils"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var force bool

var initializeCmd = &cobra.Command{
	Use:   "initialize",
	Short: "Initialize the configuration",
	Long:  "Initialize the configuration",
	Run: func(cmd *cobra.Command, args []string) {
		_config = config.NewInternalConfig(baseDirectory)

		if force {
			utils.SetProgressSteps(3)
		} else {
			utils.SetProgressSteps(2)
		}

		utils.ShowProgress()

		if !force {
			if error := _config.Load(); error == nil {
				log.WithFields(log.Fields{"error": "already initialized"}).Error("Initialize failed")

				os.Exit(-1)
			}
		} else {
			log.Info("Forcing initialization")

			utils.IncreaseProgressStep()
		}

		if error := _config.Save(); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Initialize failed")

			os.Exit(-1)
		}

		utils.IncreaseProgressStep()

		log.Info("Done")
	},
}

func init() {
	initializeCmd.Flags().BoolVarP(&force, "force", "f", false, "Force initialization if already initialized. This will basically remove the previous configuration.")
	RootCmd.AddCommand(initializeCmd)
}
