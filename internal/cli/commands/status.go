package commands

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func NewStatusCommand(deps Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show vault status and access control",
		Long:  "Displays current environment file, access status, and checks for drift between manifest and encrypted file.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, deps)
		},
	}
	return cmd
}

func runStatus(cmd *cobra.Command, deps Deps) error {
	bold := color.New(color.Bold).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	cmd.Printf("📊 %s\n", bold("Envseal Status"))
	cmd.Println(strings.Repeat("-", 40))

	cmd.Printf("%-20s %s\n", "Active Vault:", cyan(secretFilePath))

	// Load Local Identity
	identity, errIdentity := deps.IdentityManager.Load(identityFilePath)
	if errIdentity != nil {
		cmd.Printf("%-20s %s\n", "Local Identity:", red("Missing (run 'envseal init')"))
	} else {
		pubKey := identity.Recipient().String()
		cmd.Printf("%-20s %s...%s\n", "Local Identity:", green("OK "), pubKey[len(pubKey)-8:])
	}

	// Load Manifest & File State
	manifest, errManifest := deps.ManifestStore.Load()
	if errManifest != nil {
		cmd.Printf("\n%s\n", red("❌ Could not load manifest"))
		return nil
	}

	sf, errSecrets := deps.SecretsStore.Load(secretFilePath)
	if errSecrets != nil {
		cmd.Printf("\n%s\n", red(fmt.Sprintf("❌ Could not load vault (%s)", secretFilePath)))
		return nil
	}

	// Access Check
	canDecrypt := false
	if identity != nil {
		if err := sf.Unlock(identity); err == nil {
			canDecrypt = true
			defer sf.Lock()
		}
	}

	if canDecrypt {
		cmd.Printf("%-20s %s\n", "Vault Access:", green("UNLOCKED"))
	} else {
		cmd.Printf("%-20s %s\n", "Vault Access:", red("LOCKED (Access Denied)"))
	}

	cmd.Println(strings.Repeat("-", 40))
	cmd.Printf("👥 %s\n", bold("Access Control List"))

	// Detect Drift
	manifestKeys := manifest.GetPublicKeys()
	fileKeys, err := sf.GetRecipients()
	if err != nil {
		fileKeys = []string{}
	}

	for _, user := range manifest.AccessControl {
		hasRealAccess := false
		for _, fk := range fileKeys {
			if fk == user.PublicKey {
				hasRealAccess = true
				break
			}
		}

		statusTag := green("[SYNCED]")
		if !hasRealAccess {
			statusTag = yellow("[PENDING REKEY]")
		}

		isMe := ""
		if identity != nil && user.PublicKey == identity.Recipient().String() {
			isMe = cyan(" (You)")
		}

		cmd.Printf("  • %-15s %s%s\n", user.Name, statusTag, isMe)
	}

	if len(manifestKeys) != len(fileKeys) {
		cmd.Printf("\n%s\n", yellow("⚠️  DRIFT DETECTED: The manifest and the encrypted file are out of sync."))
		cmd.Printf("    Run %s to apply changes.\n", bold("envseal rekey"))
	} else {
		cmd.Printf("\n%s\n", green("✓ Vault is fully synchronized."))
	}

	return nil
}
