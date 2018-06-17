package main

import (
	"fmt"
	"os"

	"github.com/darxkies/k8s-tew/utils"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var environmentCmd = &cobra.Command{
	Use:   "environment",
	Short: "Displays environment variables",
	Long:  "Displays environment variables",
	Run: func(cmd *cobra.Command, args []string) {
		// Load config and check the rights
		if error := Bootstrap(false); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("environment failed")

			os.Exit(-1)
		}

		content, error := utils.ApplyTemplate(utils.ENVIRONMENT_TEMPLATE, struct {
			CurrentPath    string
			K8STEWPath     string
			K8SPath        string
			EtcdPath       string
			CRIPath        string
			CNIPath        string
			KubeConfig     string
			ContainerdSock string
		}{
			CurrentPath:    os.Getenv("PATH"),
			K8STEWPath:     _config.GetFullLocalAssetDirectory(utils.BINARIES_DIRECTORY),
			K8SPath:        _config.GetFullLocalAssetDirectory(utils.K8S_BINARIES_DIRECTORY),
			EtcdPath:       _config.GetFullLocalAssetDirectory(utils.ETCD_BINARIES_DIRECTORY),
			CRIPath:        _config.GetFullLocalAssetDirectory(utils.CRI_BINARIES_DIRECTORY),
			CNIPath:        _config.GetFullLocalAssetDirectory(utils.CNI_BINARIES_DIRECTORY),
			KubeConfig:     _config.GetFullLocalAssetFilename(utils.ADMIN_KUBECONFIG),
			ContainerdSock: _config.GetFullTargetAssetFilename(utils.CONTAINERD_SOCK),
		})

		if error != nil {
			log.WithFields(log.Fields{"error": error}).Error("environment failed")

			os.Exit(-1)
		}

		fmt.Println(content)
	},
}

func init() {
	RootCmd.AddCommand(environmentCmd)
}
