package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func NewSetCommand(deps Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set KEY=VALUE [KEY=VALUE...]",
		Short: "Add or update encrypted secrets",
		Long:  "Encrypts provided values and writes them to secrets.enc.yaml.",
		Example: `  envseal set DATABASE_URL=postgres://localhost:5432/db
  envseal set API_KEY=12345 DEBUG=true`,
		Args: validateSetArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSet(cmd, args, deps)
		},
	}
	return cmd
}

func validateSetArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cobra.MinimumNArgs(1)(cmd, args)
	}

	for _, a := range args {
		key, _, ok := strings.Cut(a, "=")
		if !ok {
			return fmt.Errorf("invalid argument %q: expected KEY=VALUE", a)
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return fmt.Errorf("invalid argument %q: key cannot be empty", a)
		}
	}
	return nil
}

func runSet(cmd *cobra.Command, args []string, deps Deps) error {
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

	type pair struct{ k, v string }
	pairs := make([]pair, 0, len(args))
	for _, a := range args {
		k, v, _ := strings.Cut(a, "=")
		pairs = append(pairs, pair{
			k: strings.TrimSpace(k),
			// don't trim value: spaces may be intentional
			v: v,
		})
	}

	for _, p := range pairs {
		if err := sf.SetSecret(p.k, p.v); err != nil {
			return fmt.Errorf("failed to set %s: %w", p.k, err)
		}
		cmd.Printf("✓ Set %s\n", p.k)
	}

	if err := sf.Save(); err != nil {
		return fmt.Errorf("failed to save %s: %w", secretFilePath, err)
	}

	cmd.Printf("Updated %s\n", secretFilePath)
	return nil
}
