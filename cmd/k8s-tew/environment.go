package main

import (
	"fmt"
	"os"

	"github.com/darxkies/k8s-tew/pkg/utils"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var environmentCmd = &cobra.Command{
	Use:   "environment",
	Short: "Displays environment variables",
	Long:  "Displays environment variables",
	Run: func(cmd *cobra.Command, args []string) {
		// Load config and check the rights
		if error := bootstrap(false); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Failed initializing")

			os.Exit(-1)
		}

		content, error := utils.ApplyTemplate("Environment", utils.GetTemplate(utils.TemplateEnvironment), struct {
			CurrentPath    string
			K8STEWPath     string
			K8SPath        string
			EtcdPath       string
			CRIPath        string
			CNIPath        string
			VeleroPath     string
			HostPath       string
			KubeConfig     string
			ContainerdSock string
		}{
			CurrentPath:    os.Getenv("PATH"),
			K8STEWPath:     _config.GetFullLocalAssetDirectory(utils.DirectoryBinaries),
			K8SPath:        _config.GetFullLocalAssetDirectory(utils.DirectoryK8sBinaries),
			EtcdPath:       _config.GetFullLocalAssetDirectory(utils.DirectoryEtcdBinaries),
			CRIPath:        _config.GetFullLocalAssetDirectory(utils.DirectoryCriBinaries),
			CNIPath:        _config.GetFullLocalAssetDirectory(utils.DirectoryCniBinaries),
			VeleroPath:     _config.GetFullLocalAssetDirectory(utils.DirectoryVeleroBinaries),
			HostPath:       _config.GetFullLocalAssetDirectory(utils.DirectoryHostBinaries),
			KubeConfig:     _config.GetFullLocalAssetFilename(utils.KubeconfigAdmin),
			ContainerdSock: _config.GetFullTargetAssetFilename(utils.ContainerdSock),
		}, false)

		if error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Failed generating environment")

			os.Exit(-1)
		}

		fmt.Println(content)
	},
}

func init() {
	RootCmd.AddCommand(environmentCmd)
}
