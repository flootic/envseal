package commands

import (
	"github.com/spf13/cobra"
)

// NewUsersCommand creates the parent command for user management.
func NewUsersCommand(deps Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "users",
		Short: "Manage access control (manifest)",
		Long:  `Add or remove users from the envseal.yaml manifest.`,
	}

	// Register subcommands
	cmd.AddCommand(newUsersAddCommand(deps))
	cmd.AddCommand(newUsersRemoveCommand(deps))
	return cmd
}
