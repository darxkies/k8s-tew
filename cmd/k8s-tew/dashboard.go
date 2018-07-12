package main

import (
	"fmt"
	"os"

	"github.com/darxkies/k8s-tew/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Retrieves and shows the dashboard token",
	Long:  "Retrieves and shows the dashboard token",
	Run: func(cmd *cobra.Command, args []string) {
		if error := Bootstrap(false); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("dashboard failed")

			os.Exit(-1)
		}

		kubectlCommand := fmt.Sprintf("%s --kubeconfig %s", _config.GetFullLocalAssetFilename(utils.KUBECTL_BINARY), _config.GetFullLocalAssetFilename(utils.ADMIN_KUBECONFIG))
		dashboardKeyCommand := fmt.Sprintf("%s -n kube-system describe secret $(%s -n kube-system get secret | grep admin-user | awk '{print $1}') | grep token: | awk '{print $2}'", kubectlCommand, kubectlCommand)

		output, error := utils.RunCommandWithOutput(dashboardKeyCommand)
		if error != nil {
			log.WithFields(log.Fields{"error": error}).Error("dashboard failed")

			os.Exit(-2)
		}

		fmt.Printf(output)
	},
}

func init() {
	RootCmd.AddCommand(dashboardCmd)
}
