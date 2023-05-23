package workspace

import (
	"github.com/spf13/cobra"
)

func AddWorkspaceCommands(command *cobra.Command) *cobra.Command {
	command.AddCommand(NewCmdUpdateWorkspace())
	command.AddCommand(NewCmdGetWorkspace())
	command.AddCommand(NewCmdExecWorkspace())
	command.AddCommand(NewCmdCreateWorkspace())
	command.AddCommand(NewCmdDeleteWorkspace())
	command.AddCommand(NewCmdListWorkspaces())
	command.AddCommand(NewCmdDev())
	return command
}
