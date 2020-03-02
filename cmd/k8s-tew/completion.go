package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var completionShell string

var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generates bash completion scripts",
	Long: `To load bash completion run
	
	. <(k8s-tew completion)

	or

	. <(k8s-tew completion bash)

	and for zsh run

	. <(k8s-tew completion zsh)
	
	To configure your bash shell to load completions for each session add to your bashrc
	
	# ~/.bashrc or ~/.profile
	. <(k8s-tew completion)
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return errors.New("Please specify only the shell: bash or zsh")
		}

		if len(args) == 0 || args[0] == "bash" {
			return RootCmd.GenBashCompletion(os.Stdout)

		} else if args[0] == "zsh" {
			return RootCmd.GenZshCompletion(os.Stdout)

		}

		return fmt.Errorf("Unknown shell '%s'", args[0])
	},
	ValidArgs: []string{"bash", "zsh"},
}

func init() {
	RootCmd.AddCommand(completionCmd)
}
