package main

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/darxkies/k8s-tew/config"
	"github.com/darxkies/k8s-tew/utils"
	"github.com/darxkies/k8s-tew/version"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	_ "github.com/spf13/viper"
)

var debug *bool
var baseDirectory string
var _config *config.InternalConfig

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "k8s-tew",
	Short: utils.PROJECT_TITLE,
	Long:  utils.PROJECT_TITLE,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\n", version.Version)
		fmt.Println()

		cmd.Help()
	},
}

func GetBaseDirectory() string {
	directory, error := os.Getwd()
	if error != nil {
		log.WithFields(log.Fields{"error": error}).Fatal("cwd failed")
	}

	return path.Join(directory, utils.BASE_DIRECTORY)
}

func Bootstrap(needsRoot bool) error {
	if needsRoot && !utils.IsRoot() {
		return errors.New("this program needs root rights")
	}

	_config = config.DefaultInternalConfig(baseDirectory)

	// TODO use environment to load the base directory

	return _config.Load()
}

func main() {
	debug = RootCmd.PersistentFlags().BoolP("debug", "d", false, "Show debug messages")
	RootCmd.PersistentFlags().StringVar(&baseDirectory, "base-directory", GetBaseDirectory(), "Base directory")

	if _error := RootCmd.Execute(); _error != nil {
		fmt.Println(_error)

		os.Exit(-1)
	}
}
