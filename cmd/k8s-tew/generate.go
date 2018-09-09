package main

import (
	"github.com/darxkies/k8s-tew/download"
	"github.com/darxkies/k8s-tew/generate"
	"github.com/darxkies/k8s-tew/utils"

	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var forceDownload bool
var parallel bool

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate assets",
	Long:  "Generate assets",
	Run: func(cmd *cobra.Command, args []string) {
		// Load config and check the rights
		if error := bootstrap(false); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Generate failed")

			os.Exit(-1)
		}

		downloader := download.NewDownloader(_config, forceDownload, parallel)
		generator := generate.NewGenerator(_config)

		utils.SetProgressSteps(2 + downloader.Steps() + generator.Steps() + 1)

		utils.ShowProgress()

		_config.Generate()

		log.Info("Generated config entries")

		utils.IncreaseProgressStep()

		if error := _config.Save(); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Generate failed")

			os.Exit(-1)
		}

		utils.IncreaseProgressStep()

		// Download binaries
		if error := downloader.DownloadBinaries(); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Generate failed")

			os.Exit(-1)
		}

		// Download binaries
		if error := generator.GenerateFiles(); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Generate failed")

			os.Exit(-1)
		}

		utils.HideProgress()

		log.Info("Done")
	},
}

func init() {
	generateCmd.Flags().UintVarP(&commandRetries, "command-retries", "r", 300, "The count of command retries during the setup")
	generateCmd.Flags().BoolVar(&forceDownload, "force-download", false, "Force download")
	generateCmd.Flags().BoolVar(&parallel, "parallel", false, "Download in parallel")
	RootCmd.AddCommand(generateCmd)
}
