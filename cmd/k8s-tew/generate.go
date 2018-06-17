package main

import (
	"path"
	"strings"

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
		deploymentDirectory = strings.Trim(deploymentDirectory, " ")

		if len(deploymentDirectory) == 0 {
			deploymentDirectory = baseDirectory

			if len(deploymentDirectory) > 0 && deploymentDirectory[0] != '/' {
				directory, error := os.Getwd()
				if error != nil {
					log.WithFields(log.Fields{"error": error}).Error("initialize failed")

					os.Exit(-1)
				}

				deploymentDirectory = path.Join(directory, baseDirectory)
			}
		}

		if len(deploymentDirectory) == 0 || (len(deploymentDirectory) > 0 && deploymentDirectory[0] != '/') {
			log.WithFields(log.Fields{"error": "deployment directory is invalid"}).Error("initialize failed")

			os.Exit(-1)
		}

		// Load config and check the rights
		if error := Bootstrap(false); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("generate failed")

			os.Exit(-1)
		}

		_config.Generate(deploymentDirectory)

		log.Info("generated config entries")

		if error := _config.Save(); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("generate failed")

			os.Exit(-1)
		}

		downloader := download.NewDownloader(_config, versions)

		// Download binaries
		if error := downloader.DownloadBinaries(); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("generate failed")

			os.Exit(-1)
		}

		generator := generate.NewGenerator(_config, rsaSize, caValidityPeriod, clientValidityPeriod)

		// Download binaries
		if error := generator.GenerateFiles(); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("generate failed")

			os.Exit(-1)
		}
	},
}

var deploymentDirectory string
var rsaSize int
var caValidityPeriod int
var clientValidityPeriod int

var versions download.Versions

func init() {
	versions = download.Versions{}

	generateCmd.Flags().StringVar(&deploymentDirectory, "deployment-directory", "", "Defines the directory where the files will be installed to when the deployment command is executed")
	generateCmd.Flags().IntVar(&rsaSize, "rsa-size", utils.RSA_SIZE, "RSA Size")
	generateCmd.Flags().IntVar(&caValidityPeriod, "ca-validity-period", utils.CA_VALIDITY_PERIOD, "CA Validity Period")
	generateCmd.Flags().IntVar(&clientValidityPeriod, "client-validity-period", utils.CLIENT_VALIDITY_PERIOD, "Client Validity Period")
	generateCmd.Flags().StringVar(&versions.Etcd, "etcd-version", utils.ETCD_VERSION, "Etcd version")
	generateCmd.Flags().StringVar(&versions.Flanneld, "flanneld-version", utils.FLANNELD_VERSION, "Flanneld version")
	generateCmd.Flags().StringVar(&versions.K8S, "k8s-version", utils.K8S_VERSION, "Kubernetes version")
	generateCmd.Flags().StringVar(&versions.Helm, "helm-version", utils.HELM_VERSION, "helm version")
	generateCmd.Flags().StringVar(&versions.CNI, "cni-version", utils.CNI_VERSION, "CNI version")
	generateCmd.Flags().StringVar(&versions.Containerd, "containerd-version", utils.CONTAINERD_VERSION, "containerd version")
	generateCmd.Flags().StringVar(&versions.Runc, "runc-version", utils.RUNC_VERSION, "runc version")
	generateCmd.Flags().StringVar(&versions.CriCtl, "crictl-version", utils.CRICTL_VERSION, "crictl version")
	generateCmd.Flags().StringVar(&versions.Gobetween, "gobetween-version", utils.GOBETWEEN_VERSION, "gobetween version")
	RootCmd.AddCommand(generateCmd)
}
