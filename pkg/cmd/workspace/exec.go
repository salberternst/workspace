package workspace

import (
	"errors"
	"fmt"
	"strings"

	"github.com/salberternst/workspace/pkg/k8s"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
)

type ExecOptions struct {
	Name         string
	Namespace    string
	Command      []string
	Tty          bool
	workspacePod *v1.Pod
}

func (o *ExecOptions) Complete(cmd *cobra.Command, args []string, argsLengthAtDash int) error {
	if len(args) == 0 {
		return errors.New("missing argument: name")
	}

	if argsLengthAtDash <= 0 {
		return errors.New("you must specify a command")
	}

	var err error

	o.Name = args[0]

	if o.Namespace, err = cmd.Flags().GetString("namespace"); err != nil {
		return err
	}

	o.Command = append([]string{"bash", "-c"}, strings.Join(args[argsLengthAtDash:], " "))

	if o.workspacePod, err = k8s.GetWorkspacePod(o.Namespace, o.Name); err != nil {
		return err
	}

	if o.workspacePod == nil {
		return fmt.Errorf("workspace %s in namespace %s not found\n", o.Name, o.Namespace)
	}

	return nil
}

func (o *ExecOptions) Run() error {
	return k8s.ExecuteInPod(o.workspacePod.Namespace, o.workspacePod.Name, "workspace", o.Command, o.Tty)
}

func NewCmdExecWorkspace() *cobra.Command {
	options := ExecOptions{}

	var command = &cobra.Command{
		Use: "exec name -- ...args",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := options.Complete(cmd, args, cmd.ArgsLenAtDash()); err != nil {
				return err
			}

			if err := options.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	command.Flags().BoolVar(&options.Tty, "tty", false, "Stdin is a TTY")

	return command
}
