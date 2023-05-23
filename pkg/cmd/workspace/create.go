package workspace

import (
	"fmt"

	"github.com/salberternst/workspace/pkg/builder"
	"github.com/salberternst/workspace/pkg/helm"
	"github.com/salberternst/workspace/pkg/k8s"
	"github.com/spf13/cobra"
)

type CreateWorkspaceOptions struct {
	Name                  string
	Namespace             string
	DisableVolumeCreation bool
	WaitUntilReady        bool
	WaitTimeoutInSeconds  uint
	workspaceChart        helm.Chart
	args                  builder.WorkspaceArgs
}

func (o *CreateWorkspaceOptions) Complete(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing argument: name")
	}

	var err error

	o.Name = args[0]

	if o.Namespace, err = cmd.Flags().GetString("project"); err != nil {
		return err
	}

	if o.workspaceChart, err = helm.NewChart("workspace"); err != nil {
		return err
	}

	return nil
}

func (o *CreateWorkspaceOptions) Validate() error {
	if err := helm.ReleaseExists(o.Namespace, o.Name); err == nil {
		return fmt.Errorf("Release %s in project %s already exists", o.Name, o.Namespace)
	}

	return nil
}

func (o *CreateWorkspaceOptions) Run(cmd *cobra.Command) error {
	if !o.DisableVolumeCreation {
		// if err := o.createVolumes(cmd); err != nil {
		// 	return err
		// }
	}

	// if err := ssh.InstallHostKeys(o.Namespace); err != nil {
	// 	return err
	// }

	fmt.Printf("Creating workspace %s in %s\n", o.Name, o.Namespace)
	if _, err := o.workspaceChart.Install(o.Namespace, o.Name, false, o.args.BuildValues(cmd)); err != nil {
		return err
	}

	fmt.Printf("Successfully created workspace %s in project %s\n", o.Name, o.Namespace)

	if o.WaitUntilReady {
		fmt.Printf("Waiting for workspace %s in project %s to be ready\n", o.Name, o.Namespace)

		if err := k8s.WaitForDeployment(o.Name, o.Namespace, o.WaitTimeoutInSeconds); err != nil {
			return err
		}
	}

	return nil
}

func NewCmdCreateWorkspace() *cobra.Command {
	options := CreateWorkspaceOptions{
		args: builder.NewWorkspaceArgs(""),
	}

	var command = &cobra.Command{
		Use: "create name",
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

	command.Flags().BoolVar(&options.DisableVolumeCreation, "disable-volume-creation", false, "Disable the automatic creation of the conda-env and home volume")
	command.Flags().BoolVar(&options.WaitUntilReady, "wait-until-ready", false, "Wait until the workspace is ready")
	command.Flags().UintVar(&options.WaitTimeoutInSeconds, "wait-timeout", 60, "Time to wait for workspace to get ready in seconds")

	options.args.AddFlags(command)

	return command
}
