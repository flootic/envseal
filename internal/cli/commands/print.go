package commands

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

func NewPrintCommand(deps Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "print",
		Short: "Show decrypted variables",
		Long:  "Decrypts the secrets file using your local identity and prints KEY=VALUE lines to stdout.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPrint(cmd, deps)
		},
	}
	return cmd
}

func runPrint(cmd *cobra.Command, deps Deps) error {
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

	plain, err := sf.GetAllSecrets()
	if err != nil {
		return fmt.Errorf("failed to decrypt secrets: %w", err)
	}

	keys := make([]string, 0, len(plain))
	for k := range plain {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		cmd.Printf("%s=%s\n", k, plain[k])
	}

	return nil
}
