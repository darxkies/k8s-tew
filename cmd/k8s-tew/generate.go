package main

import (
	"github.com/darxkies/k8s-tew/download"
	"github.com/darxkies/k8s-tew/generate"
	"github.com/darxkies/k8s-tew/utils"

	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate assets",
	Long:  "Generate assets",
	Run: func(cmd *cobra.Command, args []string) {
		// Load config and check the rights
		if error := Bootstrap(false); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("generate failed")

			os.Exit(-1)
		}

		downloader := download.NewDownloader(_config)
		generator := generate.NewGenerator(_config)

		utils.SetProgressSteps(2 + downloader.Steps() + generator.Steps() + 1)

		utils.ShowProgress()

		_config.Generate()

		log.Info("generated config entries")

		utils.IncreaseProgressStep()

		if error := _config.Save(); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("generate failed")

			os.Exit(-1)
		}

		utils.IncreaseProgressStep()

		// Download binaries
		if error := downloader.DownloadBinaries(); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("generate failed")

			os.Exit(-1)
		}

		// Download binaries
		if error := generator.GenerateFiles(); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("generate failed")

			os.Exit(-1)
		}

		utils.HideProgress()

		log.Info("done")
	},
}

func init() {
	RootCmd.AddCommand(generateCmd)
}
