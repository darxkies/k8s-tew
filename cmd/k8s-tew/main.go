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
var hideProgress *bool
var baseDirectory string
var _config *config.InternalConfig

func init() {
	utils.SetupLogger()
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "k8s-tew",
	Short: utils.ProjectTitle,
	Long:  utils.ProjectTitle,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\n", version.Version)
		fmt.Printf("OS: %s\n", utils.GetOSNameAndRelease())
		fmt.Println()

		_ = cmd.Help()
	},
}

// getDefaultBaseDirectory returns the base directory based on the current working directory
func getDefaultBaseDirectory() string {
	directory, error := os.Getwd()
	if error != nil {
		log.WithFields(log.Fields{"error": error}).Fatal("Failed to retrieve cwd")
	}

	return path.Join(directory, utils.BaseDirectory)
}

// getBaseDirectory returns the base directory pointing to the assets
func getBaseDirectory() string {
	result := baseDirectory

	environmentBaseDirectory := os.Getenv(utils.K8sTewBaseDirectory)

	if len(environmentBaseDirectory) > 0 {
		result = environmentBaseDirectory
	}

	return result
}

// bootstrap loads the configuration and performs other checks such as the need for root rights
func bootstrap(needsRoot bool) error {
	utils.SetDebug(*debug)
	utils.SupressProgress(*hideProgress)

	if needsRoot && !utils.IsRoot() {
		return errors.New("this program needs root rights")
	}

	_config = config.NewInternalConfig(getBaseDirectory())

	return _config.Load()
}

func main() {
	debug = RootCmd.PersistentFlags().BoolP("debug", "d", false, "Show debug messages")
	hideProgress = RootCmd.PersistentFlags().Bool("hide-progress", false, "Hide progress")
	RootCmd.PersistentFlags().StringVar(&baseDirectory, "base-directory", getDefaultBaseDirectory(), "Base directory")

	if _error := RootCmd.Execute(); _error != nil {
		fmt.Println(_error)

		os.Exit(-1)
	}
}
