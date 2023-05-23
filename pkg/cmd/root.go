package cmd

import (
	"fmt"
	"os"

	"github.com/salberternst/workspace/pkg/cmd/version"
	"github.com/salberternst/workspace/pkg/cmd/workspace"
	"github.com/salberternst/workspace/pkg/k8s"
	"github.com/spf13/cobra"
)

var (
	kubeConfigPath string
	namespace      string
)

func NewRootCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:           "workspace",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if _, err := k8s.InitClient(kubeConfigPath); err != nil {
				return err
			}

			// if the namespaces was not provided by the user we use the one from the context or default
			if !cmd.Flags().Changed("project") {
				namespace = k8s.GetClient().Namespace
			}

			return nil
		},
	}

	command.PersistentFlags().StringVar(&kubeConfigPath, "kube-config", "", "absolute path to the kubeconfig file")
	command.PersistentFlags().StringVar(&namespace, "project", "default", "Name of the project")

	workspace.AddWorkspaceCommands(command)
	command.AddCommand(version.NewCmdVersion())

	return command
}

func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(1)
	}
}
