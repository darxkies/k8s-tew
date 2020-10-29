package main

import (
	"github.com/darxkies/k8s-tew/pkg/download"
	"github.com/darxkies/k8s-tew/pkg/generate"
	"github.com/darxkies/k8s-tew/pkg/utils"

	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var forceDownload bool
var parallel bool
var pullImages bool

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

		if !_config.Config.Nodes.HasControllerNode() {
			log.WithFields(log.Fields{"error": "No controller node found"}).Error("Generate failed")

			os.Exit(-2)
		}

		if !_config.Config.Nodes.HasWorkerNode() {
			log.WithFields(log.Fields{"error": "No worker node found"}).Error("Generate failed")

			os.Exit(-3)
		}

		if !_config.Config.Nodes.HasStorageNode() {
			log.WithFields(log.Fields{"error": "No storage node found"}).Error("Generate failed")

			os.Exit(-4)
		}

		downloader := download.NewDownloader(_config, forceDownload, parallel, pullImages)
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
	generateCmd.Flags().BoolVar(&forceDownload, "force-download", false, "Force downloading all binary dependencies from the internet")
	generateCmd.Flags().BoolVar(&parallel, "parallel", false, "Download binary dependencies in parallel")
	generateCmd.Flags().BoolVar(&pullImages, "pull-images", false, "Pull and convert images to OCI to be deployed later on")
	RootCmd.AddCommand(generateCmd)
}
