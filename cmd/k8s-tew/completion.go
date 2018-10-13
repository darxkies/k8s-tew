package main

import (
	"os"

	"github.com/spf13/cobra"
)

var completionShell string

var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generates bash completion scripts",
	Long: `To load completion run
	
	. <(bitbucket completion)
	
	To configure your bash shell to load completions for each session add to your bashrc
	
	# ~/.bashrc or ~/.profile
	. <(bitbucket completion)
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if completionShell == "bash" {
			_ = RootCmd.GenBashCompletion(os.Stdout)
		} else if completionShell == "zsh" {
			_ = RootCmd.GenZshCompletion(os.Stdout)
		}
	},
}

func init() {
	completionCmd.Flags().StringVar(&completionShell, "shell", "bash", "Completion shell")

	RootCmd.AddCommand(completionCmd)
}
