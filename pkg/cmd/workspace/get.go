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
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getWorkspaceContainer(containers []corev1.Container) *corev1.Container {
	for _, container := range containers {
		if container.Name == "workspace" {
			return &container
		}
	}
	return nil
}

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
	workspacePods, err := k8s.GetClient().CoreV1.CoreV1().Pods(o.Namespace).List(context.TODO(), v1.ListOptions{
		LabelSelector: "workspace-name=" + o.Name,
	})
	if err != nil {
		return err
	}

	workspacePodsCount := len(workspacePods.Items)
	if workspacePodsCount == 0 {
		return fmt.Errorf("No pods found for workspace %s in namespace %s", o.Name, o.Namespace)
	}

	if workspacePodsCount > 1 {
		fmt.Println("Warning: more then one pod for workspace found. Using first one.")
	}

	err = printWorkspace(workspacePods.Items[0])
	if err != nil {
		return err
	}

	return nil
}

func printWorkspace(workspacePod corev1.Pod) error {
	workspaceContainer := getWorkspaceContainer(workspacePod.Spec.Containers)
	if workspaceContainer == nil {
		return fmt.Errorf("No container found for workspace %s in namespace %s", workspacePod.Name, workspacePod.Namespace)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 2, Align: text.AlignRight},
	})
	t.AppendRows([]table.Row{
		{"Name", workspacePod.Name},
		{"Project", workspacePod.Namespace},
		{"Created At", workspacePod.CreationTimestamp.Local()},
	})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Limits"})
	t.AppendSeparator()
	// fix: use gpu type
	gpuLimit := workspaceContainer.Resources.Limits["nvidia.com/gpu"]
	t.AppendRows([]table.Row{
		{"CPU", workspaceContainer.Resources.Limits.Cpu().String()},
		{"Memory", workspaceContainer.Resources.Limits.Memory().String()},
		{"GPU", gpuLimit.String()},
	})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Requests"})
	t.AppendSeparator()
	t.AppendRows([]table.Row{
		{"CPU", workspaceContainer.Resources.Requests.Cpu().String()},
		{"Memory", workspaceContainer.Resources.Requests.Memory().String()},
	})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Volumes"})
	t.AppendSeparator()
	for _, volume := range workspacePod.Spec.Volumes {
		volumeMountIndex := slices.IndexFunc(workspaceContainer.VolumeMounts, func(c corev1.VolumeMount) bool { return c.Name == volume.Name })
		t.AppendRow(table.Row{
			volume.Name,
			workspaceContainer.VolumeMounts[volumeMountIndex].MountPath,
		})
	}

	t.Render()

	return nil
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
