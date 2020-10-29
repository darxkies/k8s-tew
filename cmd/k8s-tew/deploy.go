package main

import (
	"os"
	"path"

	"github.com/darxkies/k8s-tew/pkg/deployment"
	"github.com/darxkies/k8s-tew/pkg/utils"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var identityFile string
var commandRetries uint
var skipSetup bool
var skipUpload bool
var skipRestart bool
var skipStorageSetup bool
var skipMonitoringSetup bool
var skipLoggingSetup bool
var skipBackupSetup bool
var skipShowcaseSetup bool
var skipIngressSetup bool
var forceUpload bool
var importImages bool
var wait uint

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy assets to a remote cluster",
	Long:  "Deploy assets to a remote cluster",
	Run: func(cmd *cobra.Command, args []string) {
		if error := bootstrap(false); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Failed initializing")

			os.Exit(-1)
		}

		_deployment := deployment.NewDeployment(_config, identityFile, importImages, forceUpload, parallel, commandRetries, skipSetup, skipUpload, skipRestart, skipStorageSetup, skipMonitoringSetup, skipLoggingSetup, skipBackupSetup, skipShowcaseSetup, skipIngressSetup, wait)

		utils.SetProgressSteps(_deployment.Steps() + 1)

		utils.ShowProgress()

		if error := _deployment.Deploy(); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Failed deploying")

			os.Exit(-2)
		}

		utils.HideProgress()

		log.Info("Done")
	},
}

func init() {
	deployCmd.Flags().StringVarP(&identityFile, "identity-file", "i", path.Join(os.Getenv("HOME"), ".ssh/id_rsa"), "SSH identity file")
	deployCmd.Flags().UintVarP(&commandRetries, "command-retries", "r", 1200, "The number of command retries during the setup")
	deployCmd.Flags().BoolVar(&skipSetup, "skip-setup", false, "Skip setup steps")
	deployCmd.Flags().BoolVar(&skipUpload, "skip-upload", false, "Skip upload steps")
	deployCmd.Flags().BoolVar(&skipRestart, "skip-restart", false, "Skip restart steps")
	deployCmd.Flags().BoolVar(&skipStorageSetup, "skip-storage-setup", false, "Skip storage setup and all other feature setup steps")
	deployCmd.Flags().BoolVar(&skipMonitoringSetup, "skip-monitoring-setup", false, "Skip monitoring setup")
	deployCmd.Flags().BoolVar(&skipLoggingSetup, "skip-logging-setup", false, "Skip logging setup")
	deployCmd.Flags().BoolVar(&skipBackupSetup, "skip-backup-setup", false, "Skip backup setup")
	deployCmd.Flags().BoolVar(&skipShowcaseSetup, "skip-showcase-setup", false, "Skip showcase setup")
	deployCmd.Flags().BoolVar(&skipIngressSetup, "skip-ingress-setup", false, "Skip ingress setup")
	deployCmd.Flags().BoolVar(&importImages, "import-images", false, "Install images")
	deployCmd.Flags().BoolVar(&parallel, "parallel", false, "Run steps in parallel")
	deployCmd.Flags().BoolVar(&forceUpload, "force-upload", false, "Files are uploaded without checking if they are already installed")
	deployCmd.Flags().UintVar(&wait, "wait", 0, "Wait for all cluster relevant pods to be ready and jobs to be completed. The parameter reflects the number of seconds in which the pods have to run stable.")
	RootCmd.AddCommand(deployCmd)
}
