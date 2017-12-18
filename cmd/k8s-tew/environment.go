package main

import (
	"fmt"
	"os"
	"path"

	"github.com/darxkies/k8s-tew/utils"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var environmentCmd = &cobra.Command{
	Use:   "environment",
	Short: "Environment",
	Long:  "Displays environment variables",
	Run: func(cmd *cobra.Command, args []string) {
		// Load config and check the rights
		if error := Bootstrap(false); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("environment failed")

			os.Exit(-1)
		}

		currentPath := os.Getenv("PATH")
		k8sPath := path.Join(_config.BaseDirectory, utils.GetFullK8SBinariesDirectory())
		etcdPath := path.Join(_config.BaseDirectory, utils.GetFullETCDBinariesDirectory())
		criPath := path.Join(_config.BaseDirectory, utils.GetFullCRIBinariesDirectory())
		_path := fmt.Sprintf("export PATH=%s:%s:%s:%s", k8sPath, etcdPath, criPath, currentPath)

		fmt.Println(_path)

		adminKubeConfig := _config.GetFullDeploymentFilename(utils.ADMIN_KUBECONFIG)
		_kubeConfig := fmt.Sprintf("export KUBECONFIG=%s", adminKubeConfig)

		fmt.Println(_kubeConfig)
	},
}

func init() {
	RootCmd.AddCommand(environmentCmd)
}
