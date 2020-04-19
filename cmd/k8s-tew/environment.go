package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/darxkies/k8s-tew/pkg/utils"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var environmentCmd = &cobra.Command{
	Use:   "environment",
	Short: "Set environment variables",
	Long: `To set the environment variables run
	
	. <(k8s-tew environment)
	`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load config and check the rights
		if error := bootstrap(false); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Failed initializing")

			os.Exit(-1)
		}

		checkShell := func(name string) bool {
			command := fmt.Sprintf("/proc/%d/exe -c 'echo $%s_VERSION'", os.Getppid(), strings.ToUpper(name))

			output, error := utils.RunCommandWithOutput(command)

			return error == nil && len(strings.TrimSpace(output)) > 0
		}

		shell := ""

		// Figure out if bash or zsh
		if checkShell("bash") {
			shell = "bash"
		} else if checkShell("zsh") {
			shell = "zsh"
		}

		outputBashCompletion := func(binaryName string) {
			if len(shell) == 0 {
				return
			}

			binary := _config.GetFullLocalAssetFilename(binaryName)

			command := fmt.Sprintf("%s completion %s", binary, shell)

			content, error := utils.RunCommandWithOutput(command)
			if error != nil {
				return
			}

			fmt.Println(content)
		}

		outputBashCompletion(utils.BinaryK8sTew)
		outputBashCompletion(utils.BinaryKubectl)
		outputBashCompletion(utils.BinaryHelm)
		outputBashCompletion(utils.BinaryVelero)
		outputBashCompletion(utils.BinaryCrictl)

		kubeConfig := _config.GetFullLocalAssetFilename(utils.KubeconfigAdmin)

		if !utils.FileExists(kubeConfig) {
			kubeConfig = ""
		}

		content, error := utils.ApplyTemplate("Environment", utils.GetTemplate(utils.TemplateEnvironment), struct {
			CurrentPath    string
			K8STEWPath     string
			K8SPath        string
			EtcdPath       string
			CRIPath        string
			CNIPath        string
			VeleroPath     string
			HostPath       string
			KubeConfig     string
			ContainerdSock string
		}{
			CurrentPath:    os.Getenv("PATH"),
			K8STEWPath:     _config.GetFullLocalAssetDirectory(utils.DirectoryBinaries),
			K8SPath:        _config.GetFullLocalAssetDirectory(utils.DirectoryK8sBinaries),
			EtcdPath:       _config.GetFullLocalAssetDirectory(utils.DirectoryEtcdBinaries),
			CRIPath:        _config.GetFullLocalAssetDirectory(utils.DirectoryCriBinaries),
			CNIPath:        _config.GetFullLocalAssetDirectory(utils.DirectoryCniBinaries),
			VeleroPath:     _config.GetFullLocalAssetDirectory(utils.DirectoryVeleroBinaries),
			HostPath:       _config.GetFullLocalAssetDirectory(utils.DirectoryHostBinaries),
			KubeConfig:     kubeConfig,
			ContainerdSock: _config.GetFullTargetAssetFilename(utils.ContainerdSock),
		}, false)

		if error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Failed generating environment")

			os.Exit(-1)
		}

		fmt.Println(content)
	},
}

func init() {
	RootCmd.AddCommand(environmentCmd)
}
