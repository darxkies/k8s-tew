package main

import (
	"fmt"
	"os"

	"github.com/darxkies/k8s-tew/pkg/k8s"
	"github.com/darxkies/k8s-tew/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var openAll bool
var openKubernetesDashboard bool
var openCephManager bool
var openCephRadosGateway bool
var openMinio bool
var openGrafana bool
var openKibana bool
var openCerebro bool
var openWordPressNodePort bool
var openWordPressIngress bool

func openWebBrowser(flag bool, name, protocol, ip string, port uint16) error {
	if !openAll && !flag {
		return nil
	}

	url := utils.GetURL(protocol, ip, port)

	log.WithFields(log.Fields{"name": name, "url": url}).Info("Opening Web Browser")

	return utils.OpenWebBrowser(name, url)
}

var openWebBrowserCmd = &cobra.Command{
	Use:   "open-web-browser",
	Short: "Opens the web browser pointing to various websites",
	Long:  "Opens the web browser pointing to various websites",
	Run: func(cmd *cobra.Command, args []string) {
		if error := bootstrap(false); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Failed initializing")

			os.Exit(-1)
		}

		ip, error := _config.GetWorkerIP()
		if error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Failed to get worker ip address")

			os.Exit(-3)
		}

		if openAll || openWordPressIngress {
			name := "WordPress-Ingress"
			url := fmt.Sprintf("https://%s.%s", utils.IngressSubdomainWordpress, _config.Config.IngressDomain)

			log.WithFields(log.Fields{"name": name, "url": url}).Info("Opening Web Browser")

			if error := utils.OpenWebBrowser(name, url); error != nil {
				log.WithFields(log.Fields{"error": error}).Error("Open Web Browser failed")

				os.Exit(-3)
			}
		}

		if error := openWebBrowser(openWordPressNodePort, "WordPress-NodePort", "http", ip, utils.PortWordpress); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Open Web Browser failed")

			os.Exit(-3)
		}

		if error := openWebBrowser(openCephManager, "Ceph Manager", "https", ip, utils.PortCephManager); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Open Web Browser failed")

			os.Exit(-3)
		}

		if error := openWebBrowser(openCephRadosGateway, "Ceph Rados Gateway", "http", ip, utils.PortCephRadosGateway); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Open Web Browser failed")

			os.Exit(-3)
		}

		if error := openWebBrowser(openMinio, "Minio", "http", ip, utils.PortMinio); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Open Web Browser failed")

			os.Exit(-3)
		}

		if error := openWebBrowser(openGrafana, "Grafana", "http", ip, utils.PortGrafana); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Open Web Browser failed")

			os.Exit(-3)
		}

		if error := openWebBrowser(openKibana, "Kibana", "http", ip, utils.PortKibana); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Open Web Browser failed")

			os.Exit(-3)
		}

		if error := openWebBrowser(openCerebro, "Cerebro", "http", ip, utils.PortCerebro); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Open Web Browser failed")

			os.Exit(-3)
		}

		if error := openWebBrowser(openKubernetesDashboard, "Kubernetes Dashboard", "https", ip, _config.Config.KubernetesDashboardPort); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Open Web Browser failed")

			os.Exit(-3)
		}

		kubernetesClient := k8s.NewK8S(_config)

		username, password, error := kubernetesClient.GetCredentials(utils.FeatureMonitoring, utils.GrafanaCredentials)
		if error == nil {
			log.WithFields(log.Fields{"username": username, "password": password}).Info("Grafana Credentials")
		}

		username, password, error = kubernetesClient.GetCredentials(utils.FeatureBackup, utils.MinioCredentials)
		if error == nil {
			log.WithFields(log.Fields{"username": username, "password": password}).Info("Minio Credentials")
		}

	},
}

func init() {
	openWebBrowserCmd.Flags().BoolVar(&openAll, "all", false, "Open all websites")
	openWebBrowserCmd.Flags().BoolVar(&openKubernetesDashboard, "kubernetes-dashboard", false, "Open Kubernetes Dashboard website")
	openWebBrowserCmd.Flags().BoolVar(&openCephManager, "ceph-manager", false, "Open Ceph Manager website")
	openWebBrowserCmd.Flags().BoolVar(&openCephRadosGateway, "ceph-rados-gateway", false, "Open Ceph Rados Gateway website")
	openWebBrowserCmd.Flags().BoolVar(&openMinio, "minio", false, "Open Minio website")
	openWebBrowserCmd.Flags().BoolVar(&openGrafana, "grafana", false, "Open Grafana website")
	openWebBrowserCmd.Flags().BoolVar(&openKibana, "kibana", false, "Open Kibana website")
	openWebBrowserCmd.Flags().BoolVar(&openCerebro, "cerebro", false, "Open Cerebro website")
	openWebBrowserCmd.Flags().BoolVar(&openWordPressNodePort, "wordpress-nodeport", false, "Open WordPress NodePort website")
	openWebBrowserCmd.Flags().BoolVar(&openWordPressIngress, "wordpress-ingress", false, "Open WordPress Ingress website")
	RootCmd.AddCommand(openWebBrowserCmd)
}
