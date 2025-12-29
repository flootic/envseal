package commands

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func NewRekeyCommand(deps Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rekey",
		Short: "Refresh access permissions",
		Long: `Synchronizes the encrypted file with the users defined in envseal.yaml.

Modes:
  1) Standard (default): updates recipients header only.
  2) Rotate (--rotate): generates a new DEK and re-encrypts all secrets (required for revocation).`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRekey(cmd, deps)
		},
	}

	cmd.Flags().Bool("rotate", false, "Generate a new master key and re-encrypt all data (revocation)")
	return cmd
}

func runRekey(cmd *cobra.Command, deps Deps) error {
	rotate, err := cmd.Flags().GetBool("rotate")
	if err != nil {
		return err
	}

	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	cmd.Println("🔐 Starting rekey process...")

	identity, err := deps.IdentityManager.Load(identityFilePath)
	if err != nil {
		return fmt.Errorf("identity error (run 'envseal init' first?): %w", err)
	}

	manifest, err := deps.ManifestStore.Load()
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	sf, err := deps.SecretsStore.Load(secretFilePath)
	if err != nil {
		return fmt.Errorf("failed to load %s: %w", secretFilePath, err)
	}

	if err := sf.Unlock(identity); err != nil {
		return fmt.Errorf("failed to unlock %s: %w", secretFilePath, err)
	}
	defer sf.Lock()

	recipients := manifest.GetPublicKeys()
	if len(recipients) == 0 {
		return fmt.Errorf("manifest has no recipients; add at least one user before rekey")
	}
	cmd.Printf("Target recipients: %d\n", len(recipients))

	if rotate {
		cmd.Println(yellow("⚠️  Rotation mode: re-encrypting all secrets..."))

		all, err := sf.GetAllSecrets()
		if err != nil {
			return fmt.Errorf("failed to read current secrets: %w", err)
		}

		// Rotation: generate new DEK + recipients, then re-encrypt every secret.
		if err := sf.Init(recipients); err != nil {
			return fmt.Errorf("failed to initialize new key: %w", err)
		}
		for k, v := range all {
			if err := sf.SetSecret(k, v); err != nil {
				return fmt.Errorf("failed to re-encrypt %s: %w", k, err)
			}
		}

		cmd.Println(green("✓ Keys rotated and data re-encrypted."))
	} else {
		cmd.Println(cyan("ℹ️  Standard mode: updating recipients header only..."))

		if err := sf.RotateRecipients(recipients); err != nil {
			return fmt.Errorf("failed to update recipients: %w", err)
		}

		cmd.Println(green("✓ Access headers updated."))
	}

	if err := sf.Save(); err != nil {
		return fmt.Errorf("failed to save %s: %w", secretFilePath, err)
	}

	cmd.Printf("\n%s File updated successfully.\n", bold("SUCCESS:"))
	cmd.Println("Remember to commit the changes to Git:")
	cmd.Println(cyan("  git add envseal.yaml secrets.enc.yaml"))
	cmd.Println(cyan(`  git commit -m "Update access permissions"`))

	return nil
}
