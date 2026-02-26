// commands/users_add.go
package commands

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"filippo.io/age"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/xfrr/envseal-cli/internal/p2p"
)

var (
	// conservative alias validation: letters, numbers, _ . - ; 2..64 chars
	aliasRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{1,63}$`)

	ErrInvalidArgs = errors.New("invalid arguments")
)

func newUsersAddCommand(deps Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <alias> <public_key_or_code>",
		Short: "Add a user to the manifest",
		Long: `Adds a user alias and public key to envseal.yaml.

The second argument is an Age public key (age1...).
With --p2p, it is instead a 6-digit code from 'envseal join' which triggers
a local network scan via mDNS to discover the public key.

Note: Adding a user does NOT grant access to already-encrypted secrets.
You must run 'envseal rekey' afterwards to update recipients.`,
		Example: `  envseal users add jane age1ql3z7hjy54pw3hyww5...
  envseal users add jane --p2p 482910
  envseal users add ci-server age1yt8...`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUsersAdd(cmd, args, deps)
		},
	}

	cmd.Flags().Bool("p2p", false, "Treat the second argument as a join code and scan the local network via mDNS")

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

	// 1. Resolve the public key and retrieve the optional ACK trigger function
	pubKey, ackFunc, err := resolvePublicKey(cmd, args)
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

	// 2. Save to disk FIRST to guarantee data consistency
	if err := deps.ManifestStore.Save(manifest); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	// 3. Send the success signal (TCP ACK) back to the new user if applicable
	if ackFunc != nil {
		ackFunc()
	}

	printUsersAddSuccess(cmd, alias)
	return nil
}

// resolvePublicKey returns the Age public key and an optional ACK function
// that should be called when the key is successfully stored.
func resolvePublicKey(cmd *cobra.Command, args []string) (string, func(), error) {
	value := strings.TrimSpace(args[1])

	useP2P, _ := cmd.Flags().GetBool("p2p")
	if !useP2P {
		// Manual entry: no ACK needed
		return value, nil, nil
	}

	// The --p2p flag was set — scan the local network for the join code.
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	cmd.Printf("🔍 Scanning local network for code %s…\n", cyan(value))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Using the updated p2p module that returns a *DiscoverResult
	res, err := p2p.DiscoverKey(ctx, value)
	if err != nil {
		return "", nil, fmt.Errorf("could not find code %s on local network: %w", value, err)
	}

	cmd.Printf("✅ Found public key via local network.\n")

	// Return the discovered key and the callback to trigger the TCP ping
	return res.PubKey, res.SendAck, nil
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
