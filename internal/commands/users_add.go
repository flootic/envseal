// commands/users_add.go
package commands

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"filippo.io/age"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	// conservative alias validation: letters, numbers, _ . - ; 2..64 chars
	aliasRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{1,63}$`)

	ErrInvalidArgs = errors.New("invalid arguments")
)

func newUsersAddCommand(deps Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <alias> <public_key>",
		Short: "Add a user to the manifest",
		Long: `Adds a user alias and public key to envseal.yaml.

Note: Adding a user does NOT grant access to already-encrypted secrets.
You must run 'envseal rekey' afterwards to update recipients.`,
		Example: `  envseal users add jane age1ql3z7hjy54pw3hyww5...
  envseal users add ci-server age1yt8...`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUsersAdd(cmd, args, deps)
		},
	}

	return cmd
}

func runUsersAdd(cmd *cobra.Command, args []string, deps Deps) error {
	alias := strings.TrimSpace(args[0])
	if alias == "" {
		return fmt.Errorf("alias cannot be empty")
	}
	if !aliasRe.MatchString(alias) {
		return fmt.Errorf("invalid alias %q (allowed: letters, numbers, '_', '.', '-', 2-64 chars)", alias)
	}

	pubKey, err := resolvePublicKey(cmd, args)
	if err != nil {
		return err
	}

	// Validate Age recipient format.
	if _, err := age.ParseX25519Recipient(pubKey); err != nil {
		return fmt.Errorf("invalid public key format: %w", err)
	}

	manifest, err := deps.ManifestStore.Load()
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	if err := manifest.AddUser(alias, pubKey); err != nil {
		return fmt.Errorf("failed to add user: %w", err)
	}

	if err := deps.ManifestStore.Save(manifest); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	printUsersAddSuccess(cmd, alias)
	return nil
}

func resolvePublicKey(_ *cobra.Command, args []string) (string, error) {
	return strings.TrimSpace(args[1]), nil
}

func printUsersAddSuccess(cmd *cobra.Command, alias string) {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	cmd.Printf("%s User %q added to manifest.\n", green("✓"), alias)
	cmd.Println()
	cmd.Println(yellow("⚠️  PENDING ACTION:"))
	cmd.Printf("The user is listed, but %s access yet.\n", bold("DOES NOT HAVE"))
	cmd.Printf("Run %s to update the encrypted file permissions.\n", bold("envseal rekey"))
}
