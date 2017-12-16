package main

import (
	"github.com/darxkies/k8s-tew/download"
	"github.com/darxkies/k8s-tew/generate"
	"github.com/darxkies/k8s-tew/utils"

	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate",
	Long:  "Generate artifacts",
	Run: func(cmd *cobra.Command, args []string) {
		// Load config and check the rights
		if error := Bootstrap(false); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("generate failed")

			os.Exit(-1)
		}

		downloader := download.NewDownloader(_config, etcdVersion, flanneldVersion, k8sVersion, cniVersion, criVersion)

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

var rsaSize int
var caValidityPeriod int
var clientValidityPeriod int
var etcdVersion string
var flanneldVersion string
var k8sVersion string
var cniVersion string
var criVersion string

func init() {
	generateCmd.Flags().IntVar(&rsaSize, "rsa-size", utils.RSA_SIZE, "RSA Size")
	generateCmd.Flags().IntVar(&caValidityPeriod, "ca-validity-period", utils.CA_VALIDITY_PERIOD, "CA Validity Period")
	generateCmd.Flags().IntVar(&clientValidityPeriod, "client-validity-period", utils.CLIENT_VALIDITY_PERIOD, "Client Validity Period")
	generateCmd.Flags().StringVar(&etcdVersion, "etcd-version", utils.ETCD_VERSION, "Etcd version")
	generateCmd.Flags().StringVar(&flanneldVersion, "flanneld-version", utils.FLANNELD_VERSION, "Flanneld version")
	generateCmd.Flags().StringVar(&k8sVersion, "k8s-version", utils.K8S_VERSION, "Kubernetes version")
	generateCmd.Flags().StringVar(&cniVersion, "cni-version", utils.CNI_VERSION, "CNI version")
	generateCmd.Flags().StringVar(&criVersion, "cri-version", utils.CRI_VERSION, "CRI version")
	RootCmd.AddCommand(generateCmd)
}
