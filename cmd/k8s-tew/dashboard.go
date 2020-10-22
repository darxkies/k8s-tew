package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/darxkies/k8s-tew/pkg/k8s"
	"github.com/darxkies/k8s-tew/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var quiet bool
var openWebBrowser bool

type getData func(kubernetesClient *k8s.K8S, ip string) (string, string, string, error)

func addCommand(subCommandName, description string, getData getData) *cobra.Command {
	cmd := &cobra.Command{
		Use:   subCommandName,
		Short: description,
		Long:  description,
		Run: func(cmd *cobra.Command, args []string) {
			if error := bootstrap(false); error != nil {
				log.WithFields(log.Fields{"error": error}).Error("Failed initializing")

				os.Exit(-1)
			}

			kubernetesClient := k8s.NewK8S(_config)

			ip, error := _config.GetWorkerIP()
			if error != nil {
				log.WithFields(log.Fields{"error": error}).Error("Failed to get worker ip address")

				os.Exit(-3)
			}

			url, username, password, error := getData(kubernetesClient, ip)

			if quiet {
				fmt.Printf("%s", password)

			} else {
				fields := log.Fields{"url": url}

				if len(username) > 0 {
					fields["username"] = username
				}

				if len(password) > 0 {
					fields["password"] = password
				}

				log.WithFields(fields).Info(subCommandName)
			}

			if error := utils.OpenWebBrowser(subCommandName, url); error != nil {
				log.WithFields(log.Fields{"error": error}).Error("Open Web Browser failed")

				os.Exit(-3)
			}
		},
	}

	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Display only password/secret")
	cmd.Flags().BoolVarP(&openWebBrowser, "open-web-browser", "o", false, "Open web browser")

	return cmd
}

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Display credentials of various dashboards",
	Long:  "Display credentials of various dashboards",
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.New("Missing sub-command")
	},
}

func init() {
	dashboardCmd.AddCommand(addCommand("kubernetes", "Display Kuberenetes Dashboard website related information", func(kubernetesClient *k8s.K8S, ip string) (string, string, string, error) {
		secret, error := kubernetesClient.GetSecretToken(utils.AdminUserNamespace, utils.AdminUserName)
		if error != nil {
			return "", "", "", error
		}

		return utils.GetURL("https", ip, _config.Config.KubernetesDashboardPort), "", secret, nil
	}))

	dashboardCmd.AddCommand(addCommand("ceph-manager", "Display Ceph Manager website related information", func(kubernetesClient *k8s.K8S, ip string) (string, string, string, error) {
		username, password, error := kubernetesClient.GetCredentials(utils.FeatureStorage, utils.CephManagerCredentials)
		if error != nil {
			return "", "", "", error
		}

		return utils.GetURL("http", ip, utils.PortCephManager), username, password, nil
	}))

	dashboardCmd.AddCommand(addCommand("ceph-rados-gateway", "Display Ceph Rados Gateway website related information", func(kubernetesClient *k8s.K8S, ip string) (string, string, string, error) {
		username, password, error := kubernetesClient.GetCredentials(utils.FeatureStorage, utils.CephRadosGatewayCredentials)
		if error != nil {
			return "", "", "", error
		}

		return utils.GetURL("http", ip, utils.PortCephRadosGateway), username, password, nil
	}))

	dashboardCmd.AddCommand(addCommand("minio", "Display Minio website related information", func(kubernetesClient *k8s.K8S, ip string) (string, string, string, error) {
		username, password, error := kubernetesClient.GetCredentials(utils.FeatureBackup, utils.MinioCredentials)
		if error != nil {
			return "", "", "", error
		}

		return utils.GetURL("https", ip, utils.PortMinio), username, password, nil
	}))

	dashboardCmd.AddCommand(addCommand("grafana", "Display Grafana website related information", func(kubernetesClient *k8s.K8S, ip string) (string, string, string, error) {
		username, password, error := kubernetesClient.GetCredentials(utils.FeatureMonitoring, utils.GrafanaCredentials)
		if error != nil {
			return "", "", "", error
		}

		return utils.GetURL("http", ip, utils.PortGrafana), username, password, nil
	}))

	dashboardCmd.AddCommand(addCommand("kibana", "Display Kibana website related information", func(kubernetesClient *k8s.K8S, ip string) (string, string, string, error) {
		username, password, error := kubernetesClient.GetCredentials(utils.FeatureLogging, utils.ElasticsearchCredentials)
		if error != nil {
			return "", "", "", error
		}

		return utils.GetURL("https", ip, utils.PortKibana), username, password, nil
	}))

	dashboardCmd.AddCommand(addCommand("cerebro", "Display Cerebro website related information", func(kubernetesClient *k8s.K8S, ip string) (string, string, string, error) {
		username, password, error := kubernetesClient.GetCredentials(utils.FeatureLogging, utils.CerebroCredentials)
		if error != nil {
			return "", "", "", error
		}

		return utils.GetURL("https", ip, utils.PortCerebro), username, password, nil
	}))

	dashboardCmd.AddCommand(addCommand("wordpress-nodeport", "Display WordPress Nodeport website related information", func(kubernetesClient *k8s.K8S, ip string) (string, string, string, error) {
		return utils.GetURL("http", ip, utils.PortWordpress), "", "", nil
	}))

	dashboardCmd.AddCommand(addCommand("wordpress-ingress", "Display WordPress Ingress website related information", func(kubernetesClient *k8s.K8S, ip string) (string, string, string, error) {
		return fmt.Sprintf("https://%s.%s", utils.IngressSubdomainWordpress, _config.Config.IngressDomain), "", "", nil
	}))

	RootCmd.AddCommand(dashboardCmd)
}
