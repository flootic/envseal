package commands

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newUsersRemoveCommand(deps Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <alias_or_pubkey>",
		Short: "Remove a user from the manifest",
		Long: `Removes a user from envseal.yaml.

CRITICAL: This does not revoke access to existing encrypted data until you rotate encryption keys.`,
		Example: `  envseal users remove jane
  envseal users remove age1ql3z...`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUsersRemove(cmd, args, deps)
		},
	}

	return cmd
}

func runUsersRemove(cmd *cobra.Command, args []string, deps Deps) error {
	identifier := strings.TrimSpace(args[0])
	if identifier == "" {
		return fmt.Errorf("identifier cannot be empty")
	}

	manifest, err := deps.ManifestStore.Load()
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Prefer strict remove if available, to keep behavior explicit and testable.
	// Fallback to bool-based RemoveUser if you haven't added RemoveUserStrict.
	if remover, ok := any(manifest).(interface{ RemoveUserStrict(string) error }); ok {
		if err := remover.RemoveUserStrict(identifier); err != nil {
			return fmt.Errorf("failed to remove user: %w", err)
		}
	} else {
		found := manifest.RemoveUser(identifier)
		if !found {
			return fmt.Errorf("user %q not found in manifest", identifier)
		}
	}

	if err := deps.ManifestStore.Save(manifest); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	printUsersRemoveSuccess(cmd, identifier)
	printUsersRemoveSecurityWarning(cmd)

	return nil
}

func printUsersRemoveSuccess(cmd *cobra.Command, identifier string) {
	green := color.New(color.FgGreen).SprintFunc()
	cmd.Printf("%s User %q removed from manifest.\n", green("✓"), identifier)
}

func printUsersRemoveSecurityWarning(cmd *cobra.Command) {
	red := color.New(color.FgRed).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	cmd.Println()
	cmd.Println(red("🚨 SECURITY WARNING:"))
	cmd.Println("The user has been removed from the list, but they may still decrypt")
	cmd.Println("the current file if they already have a copy of the old encryption key.")
	cmd.Println()
	cmd.Printf("You MUST run: %s\n", bold("envseal rekey --rotate"))
	cmd.Println("to re-encrypt all secrets with a new key and permanently revoke their access.")
}
