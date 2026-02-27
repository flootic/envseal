package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func NewUnsetCommand(deps Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unset KEY [KEY...]",
		Short: "Remove secrets from the file",
		Long:  "Removes one or more variables from secrets.enc.yaml.",
		Example: `  envseal unset STRIPE_KEY
  envseal unset DB_HOST DB_PORT`,
		Args: validateUnsetArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUnset(cmd, args, deps)
		},
	}
	return cmd
}

func validateUnsetArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cobra.MinimumNArgs(1)(cmd, args)
	}
	for _, k := range args {
		if strings.TrimSpace(k) == "" {
			return fmt.Errorf("key cannot be empty")
		}
	}
	return nil
}

func runUnset(cmd *cobra.Command, args []string, deps Deps) error {
	identity, err := deps.IdentityManager.Load(identityFilePath)
	if err != nil {
		return fmt.Errorf("identity error (run 'envseal init' first?): %w", err)
	}

	sf, err := deps.SecretsStore.Load(secretFilePath)
	if err != nil {
		return fmt.Errorf("failed to load %s: %w", secretFilePath, err)
	}

	if err := sf.Unlock(identity); err != nil {
		return fmt.Errorf("failed to unlock %s: %w", secretFilePath, err)
	}
	defer sf.Lock()

	// Normalize keys and dedupe to avoid repeated work/noise.
	keys := normalizeKeys(args)

	var removed int
	var warned int

	for _, key := range keys {
		if err := sf.UnsetSecret(key); err != nil {
			// Non-fatal: keep processing, but make it visible.
			cmd.Printf("⚠️  %v\n", err)
			warned++
			continue
		}
		cmd.Printf("✓ Unset %s\n", key)
		removed++
	}

	if removed == 0 {
		cmd.Println("No changes made.")
		return nil
	}

	if err := sf.Save(); err != nil {
		return fmt.Errorf("failed to save %s: %w", secretFilePath, err)
	}

	cmd.Printf("Updated %s (%d removed", secretFilePath, removed)
	if warned > 0 {
		cmd.Printf(", %d warning(s)", warned)
	}
	cmd.Println(").")

	return nil
}

func normalizeKeys(args []string) []string {
	set := make(map[string]struct{}, len(args))
	out := make([]string, 0, len(args))
	for _, a := range args {
		k := strings.TrimSpace(a)
		if k == "" {
			continue
		}
		if _, ok := set[k]; ok {
			continue
		}
		set[k] = struct{}{}
		out = append(out, k)
	}
	return out
}
