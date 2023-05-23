package workspace

import (
	"errors"
	"fmt"

	"github.com/salberternst/workspace/pkg/helm"
	"github.com/salberternst/workspace/pkg/k8s"
	"github.com/spf13/cobra"
)

type DeleteWorkspaceOptions struct {
	Name           string
	Namespace      string
	DeleteVolumes  bool
	workspaceChart helm.Chart
	volumeChart    helm.Chart
}

func (o *DeleteWorkspaceOptions) Init() error {
	var err error

	o.workspaceChart, err = helm.NewChart("workspace")
	if err != nil {
		return err
	}

	if o.DeleteVolumes {
		o.volumeChart, err = helm.NewChart("volume")
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *DeleteWorkspaceOptions) Complete(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument: name")
	}

	var err error

	o.Name = args[0]

	if o.Namespace, err = cmd.Flags().GetString("project"); err != nil {
		return err
	}

	return nil
}

func (o *DeleteWorkspaceOptions) Validate() error {
	return o.workspaceChart.Get(o.Namespace, o.Name)
}

func (o *DeleteWorkspaceOptions) Run() error {
	if o.DeleteVolumes {
		volumes, err := k8s.GetWorkspaceVolumes(o.Namespace, o.Name)
		if err != nil {
			return err
		}

		for _, volume := range volumes {
			_, err := o.volumeChart.Delete(o.Namespace, volume.Name, false)
			if err != nil {
				return err
			}

			fmt.Printf("Successfully deleted volume %s in project %s\n", volume.Name, o.Namespace)
		}
	}

	if _, err := o.workspaceChart.Delete(o.Namespace, o.Name, false); err != nil {
		return err
	}

	return nil
}

func NewCmdDeleteWorkspace() *cobra.Command {
	options := DeleteWorkspaceOptions{}

	var command = &cobra.Command{
		Use: "delete name",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("a name is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := options.Init(); err != nil {
				return err
			}

			if err := options.Complete(cmd, args); err != nil {
				return err
			}

			if err := options.Validate(); err != nil {
				return err
			}

			if err := options.Run(); err != nil {
				return err
			}

			fmt.Printf("Successfully deleted workspace %s in project %s\n", args[0], options.Namespace)

			return nil
		},
	}

	command.Flags().BoolVar(&options.DeleteVolumes, "delete-volumes", false, "Delete automatic created volumes")

	return command
}
