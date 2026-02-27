package commands

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func NewWhoamiCommand(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show your local identity",
		Long:  "Displays your public key (Age format). Share this key with a project administrator to get access.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWhoami(cmd, deps)
		},
	}
}

func runWhoami(cmd *cobra.Command, deps Deps) error {
	identity, err := deps.IdentityManager.Load(identityFilePath)
	if err != nil {
		return fmt.Errorf("identity error (run 'envseal init' first?): %w", err)
	}

	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	cmd.Println("👋 Your Identity:")
	cmd.Println(cyan(identity.Recipient().String()))
	cmd.Println()
	cmd.Println(bold("Next step:"), "Send this key to your project administrator.")

	return nil
}
