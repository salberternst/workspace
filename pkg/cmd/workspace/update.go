package workspace

import (
	"errors"
	"fmt"

	"github.com/salberternst/workspace/pkg/builder"
	"github.com/salberternst/workspace/pkg/helm"
	"github.com/salberternst/workspace/pkg/k8s"
	"github.com/spf13/cobra"
)

type UpdateWorkspaceOptions struct {
	Name                 string
	Namespace            string
	WaitUntilReady       bool
	WaitTimeoutInSeconds uint
	workspaceChart       helm.Chart
	args                 builder.WorkspaceArgs
}

func (o *UpdateWorkspaceOptions) Complete(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument: name")
	}

	var err error

	o.Name = args[0]

	o.Namespace, err = cmd.Flags().GetString("project")
	if err != nil {
		return err
	}

	o.workspaceChart, err = helm.NewChart("workspace")
	if err != nil {
		return err
	}

	return nil
}

func (o *UpdateWorkspaceOptions) Validate() error {
	return o.workspaceChart.Get(o.Namespace, o.Name)
}

func (o *UpdateWorkspaceOptions) Run(cmd *cobra.Command) error {
	if _, err := o.workspaceChart.Update(o.Namespace, o.Name, false, o.args.BuildValues(cmd)); err != nil {
		return err
	}

	if o.WaitUntilReady {
		fmt.Printf("Waiting for workspace %s in project %s to be ready\n", o.Name, o.Namespace)

		if err := k8s.WaitForDeployment(o.Name, o.Namespace, o.WaitTimeoutInSeconds); err != nil {
			return err
		}
	}

	return nil
}

func NewCmdUpdateWorkspace() *cobra.Command {
	options := UpdateWorkspaceOptions{
		args: builder.NewWorkspaceArgs(""),
	}

	var command = &cobra.Command{
		Use: "update name",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("a name is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := options.Complete(cmd, args); err != nil {
				return err
			}

			if err := options.Validate(); err != nil {
				return err
			}

			if err := options.Run(cmd); err != nil {
				return err
			}

			return nil
		},
	}

	command.Flags().BoolVar(&options.WaitUntilReady, "wait-until-ready", false, "Wait until the workspace is ready")
	command.Flags().UintVar(&options.WaitTimeoutInSeconds, "wait-timeout", 60, "Time to wait for workspace to get ready in seconds")

	options.args.AddFlags(command)

	return command
}
