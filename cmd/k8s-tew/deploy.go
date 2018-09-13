package main

import (
	"os"
	"path"

	"github.com/darxkies/k8s-tew/deployment"
	"github.com/darxkies/k8s-tew/utils"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var identityFile string
var commandRetries uint
var skipSetup bool
var skipStorageSetup bool
var skipMonitoringSetup bool
var skipLoggingSetup bool
var skipBackupSetup bool
var skipShowcaseSetup bool
var skipIngressSetup bool
var skipPackagingSetup bool
var pullImages bool
var forceUpload bool

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy assets to a remote cluster",
	Long:  "Deploy assets to a remote cluster",
	Run: func(cmd *cobra.Command, args []string) {
		if error := bootstrap(false); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Failed initializing")

			os.Exit(-1)
		}

		_deployment := deployment.NewDeployment(_config, identityFile, pullImages, forceUpload, parallel, commandRetries, skipSetup, skipStorageSetup, skipMonitoringSetup, skipLoggingSetup, skipBackupSetup, skipShowcaseSetup, skipIngressSetup, skipPackagingSetup)

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
	deployCmd.Flags().UintVarP(&commandRetries, "command-retries", "r", 300, "The count of command retries during the setup")
	deployCmd.Flags().BoolVar(&skipSetup, "skip-setup", false, "Skip setup steps")
	deployCmd.Flags().BoolVar(&skipStorageSetup, "skip-storage-setup", false, "Skip storage setup and all other setup steps that require storage")
	deployCmd.Flags().BoolVar(&skipMonitoringSetup, "skip-monitoring-setup", false, "Skip monitoring setup")
	deployCmd.Flags().BoolVar(&skipLoggingSetup, "skip-logging-setup", false, "Skip logging setup")
	deployCmd.Flags().BoolVar(&skipBackupSetup, "skip-backup-setup", false, "Skip backup setup")
	deployCmd.Flags().BoolVar(&skipShowcaseSetup, "skip-showcase-setup", false, "Skip showcase setup")
	deployCmd.Flags().BoolVar(&skipIngressSetup, "skip-ingress-setup", false, "Skip ingress setup")
	deployCmd.Flags().BoolVar(&skipPackagingSetup, "skip-packaging-setup", false, "Skip packaging setup")
	deployCmd.Flags().BoolVar(&pullImages, "pull-images", false, "Pull images")
	deployCmd.Flags().BoolVar(&parallel, "parallel", false, "Run steps in parallel")
	deployCmd.Flags().BoolVar(&forceUpload, "force-upload", false, "Files are uploaded without without any checks")
	RootCmd.AddCommand(deployCmd)
}
