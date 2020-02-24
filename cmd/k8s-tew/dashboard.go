package main

import (
	"fmt"
	"os"
	"time"

	"github.com/darxkies/k8s-tew/pkg/k8s"
	"github.com/darxkies/k8s-tew/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var openBrowser bool

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Retrieves and shows the dashboard token",
	Long:  "Retrieves and shows the dashboard token",
	Run: func(cmd *cobra.Command, args []string) {
		if error := bootstrap(false); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Failed initializing")

			os.Exit(-1)
		}

		kubernetesClient := k8s.NewK8S(_config)
		secret, error := kubernetesClient.GetSecretToken(utils.AdminUserNamespace, utils.AdminUserName)
		if error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Failed retrieving token")

			os.Exit(-2)
		}

		fmt.Printf("%s", secret)

		if openBrowser {
			fmt.Printf("\nOpening web browser...\n")

			time.Sleep(3 * time.Second)

			ip, error := _config.GetWorkerIP()
			if error != nil {
				log.WithFields(log.Fields{"error": error}).Error("Failed to get worker ip address")

				os.Exit(-3)
			}

			url := utils.GetURL("https", ip, _config.Config.KubernetesDashboardPort)

			if error := utils.OpenWebBrowser("Kubernetes Dashboard", url); error != nil {
				log.WithFields(log.Fields{"error": error}).Error("Failed to open the web browser")

				os.Exit(-4)
			}

		}
	},
}

func init() {
	dashboardCmd.Flags().BoolVarP(&openBrowser, "open-browser", "o", false, "Open the web browser with a delay of 3 seconds")
	RootCmd.AddCommand(dashboardCmd)
}
