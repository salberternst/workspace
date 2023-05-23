package workspace

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/salberternst/workspace/pkg/helm"
	"github.com/salberternst/workspace/pkg/k8s"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GetWorkspaceOptions struct {
	Name           string
	Namespace      string
	workspaceChart helm.Chart
}

func (o *GetWorkspaceOptions) Init() error {
	var err error

	if o.workspaceChart, err = helm.NewChart("workspace"); err != nil {
		return err
	}

	return err
}

func (o *GetWorkspaceOptions) Complete(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing argument: name")
	}

	var err error

	o.Name = args[0]

	if o.Namespace, err = cmd.Flags().GetString("project"); err != nil {
		return err
	}

	return nil
}

func (o *GetWorkspaceOptions) Validate() error {
	return o.workspaceChart.Get(o.Namespace, o.Name)
}

func (o *GetWorkspaceOptions) Run() error {
	workspace, err := k8s.GetClient().CoreV1.AppsV1().Deployments(o.Namespace).Get(context.TODO(), o.Name, v1.GetOptions{})
	if err != nil {
		return err
	}

	printWorkspace(workspace)

	return nil
}

func printWorkspace(workspace *appsv1.Deployment) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 2, Align: text.AlignRight},
	})
	t.AppendRows([]table.Row{
		{"Name", workspace.Name},
		{"Project", workspace.Namespace},
		{"Created At", workspace.CreationTimestamp.Local()},
	})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Limits"})
	t.AppendSeparator()
	gpuLimit := workspace.Spec.Template.Spec.Containers[0].Resources.Limits["nvidia.com/gpu"]
	t.AppendRows([]table.Row{
		{"CPU", workspace.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().String()},
		{"Memory", workspace.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().String()},
		{"GPU", gpuLimit.String()},
	})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Requests"})
	t.AppendSeparator()
	t.AppendRows([]table.Row{
		{"CPU", workspace.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().String()},
		{"Memory", workspace.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().String()},
	})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Volumes"})
	t.AppendSeparator()
	for _, volume := range workspace.Spec.Template.Spec.Volumes {
		volumeMountIndex := slices.IndexFunc(workspace.Spec.Template.Spec.Containers[0].VolumeMounts, func(c corev1.VolumeMount) bool { return c.Name == volume.Name })
		t.AppendRow(table.Row{
			volume.Name,
			workspace.Spec.Template.Spec.Containers[0].VolumeMounts[volumeMountIndex].MountPath,
		})
	}

	t.Render()
}

func NewCmdGetWorkspace() *cobra.Command {
	options := GetWorkspaceOptions{}

	var command = &cobra.Command{
		Use: "get name",
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

			return nil
		},
	}
	return command
}
