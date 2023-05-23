package workspace

import (
	"context"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/salberternst/workspace/pkg/k8s"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ListWorkspaceOptions struct {
	Namespace string
}

func (o *ListWorkspaceOptions) Init() error {
	return nil
}

func (o *ListWorkspaceOptions) Complete(cmd *cobra.Command, args []string) error {
	var err error

	o.Namespace, err = cmd.Flags().GetString("project")
	if err != nil {
		return nil
	}

	return nil
}

func (o *ListWorkspaceOptions) Validate() error {
	return nil
}

func (o *ListWorkspaceOptions) Run() error {
	workspaces, err := k8s.GetClient().CoreV1.AppsV1().Deployments(o.Namespace).List(context.TODO(), v1.ListOptions{
		LabelSelector: "workspace-name",
	})

	if err != nil {
		return err
	}

	if len(workspaces.Items) < 1 {
		fmt.Printf("No workspaces found in project %s\n", o.Namespace)
		return nil
	}

	printWorkspaces(workspaces)

	return nil
}

func printWorkspaces(workspaces *appsv1.DeploymentList) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Replicas Ready", "Created At", "Description"})

	for _, workspace := range workspaces.Items {
		t.AppendRows([]table.Row{
			{
				workspace.Name,
				fmt.Sprintf("%d/%d", workspace.Status.ReadyReplicas, *workspace.Spec.Replicas),
				workspace.CreationTimestamp.Local(),
				workspace.ObjectMeta.Annotations["workspace-description"],
			},
		})
	}

	t.Render()
}

func NewCmdListWorkspaces() *cobra.Command {
	options := ListWorkspaceOptions{}
	var command = &cobra.Command{
		Use: "list",
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

			return nil
		},
	}

	return command
}
