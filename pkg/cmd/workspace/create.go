package workspace

import (
	"fmt"

	"github.com/salberternst/workspace/pkg/builder"
	"github.com/salberternst/workspace/pkg/helm"
	"github.com/salberternst/workspace/pkg/k8s"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/watch"
)

type CreateWorkspaceOptions struct {
	Name                 string
	Namespace            string
	NoWait               bool
	NoWaitEvents         bool
	WaitTimeoutInSeconds uint
	workspaceChart       helm.Chart
	args                 builder.WorkspaceArgs
}

func (o *CreateWorkspaceOptions) Complete(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing argument: name")
	}

	var err error

	o.Name = args[0]

	if o.Namespace, err = cmd.Flags().GetString("namespace"); err != nil {
		return err
	}

	if o.workspaceChart, err = helm.NewChart("workspace"); err != nil {
		return err
	}

	return nil
}

func (o *CreateWorkspaceOptions) Validate() error {
	if err := helm.ReleaseExists(o.Namespace, o.Name); err == nil {
		return fmt.Errorf("Release %s in namespace %s already exists", o.Name, o.Namespace)
	}

	return nil
}

func (o *CreateWorkspaceOptions) Run(cmd *cobra.Command) error {
	fmt.Printf("Creating workspace %s in %s\n", o.Name, o.Namespace)
	if _, err := o.workspaceChart.Install(o.Namespace, o.Name, false, o.args.BuildValues(cmd)); err != nil {
		return err
	}

	if !o.NoWait {
		fmt.Printf("Waiting for workspace %s in namespace %s to become ready\n", o.Name, o.Namespace)
		if err := k8s.WaitForStatefulSetReplica(o.Name, o.Namespace, o.WaitTimeoutInSeconds); err != nil {
			return err
		}

		var watcher watch.Interface
		var err error

		if !o.NoWaitEvents {
			watcher, err = k8s.WatchPodEvents(o.Name, o.Namespace)
			if err != nil {
				return err
			}

			defer watcher.Stop()
		}

		if err := k8s.WaitForStatefulSetReplicaReady(o.Name, o.Namespace, o.WaitTimeoutInSeconds); err != nil {
			return err
		}

		fmt.Printf("Workspace %s in namespace %s running\n", o.Name, o.Namespace)
		fmt.Printf("Use: workspace dev %s --namespace %s\n", o.Name, o.Namespace)
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

	command.Flags().BoolVar(&options.NoWait, "no-wait", false, "Do not wait until the workspace become ready")
	command.Flags().BoolVar(&options.NoWaitEvents, "no-wait-events", false, "Do not print events while waiting for the workspace to become ready")
	command.Flags().UintVar(&options.WaitTimeoutInSeconds, "wait-timeout", 200, "Time to wait for workspace to get ready in seconds")

	options.args.AddFlags(command)

	return command
}
