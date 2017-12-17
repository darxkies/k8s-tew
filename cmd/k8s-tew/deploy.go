package main

import (
	"os"
	"path"

	"github.com/darxkies/k8s-tew/deployment"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var identityFile string

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy",
	Long:  "Deploy artifacts",
	Run: func(cmd *cobra.Command, args []string) {
		if error := Bootstrap(false); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("deploy failed")

			os.Exit(-1)
		}

		if error := deployment.Deploy(_config, identityFile); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("deploy failed")

			os.Exit(-1)
		}
	},
}

func init() {
	deployCmd.Flags().StringVarP(&identityFile, "identity-file", "i", path.Join(os.Getenv("HOME"), ".ssh/id_rsa"), "SSH identity file")
	RootCmd.AddCommand(deployCmd)
}
