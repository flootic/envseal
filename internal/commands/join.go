package commands

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/xfrr/envseal-cli/internal/crypto"
	"github.com/xfrr/envseal-cli/internal/p2p"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func NewJoinCommand(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "join",
		Short: "Pair with an admin on the local network",
		Long:  "Broadcasts your public key via mDNS so an admin can add you without copy-pasting.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runJoin(cmd, deps)
		},
	}
}

func runJoin(cmd *cobra.Command, deps Deps) error {
	identity, err := deps.IdentityManager.Load(identityFilePath)
	if err != nil {
		return fmt.Errorf("identity error (run 'envseal init' first?): %w", err)
	}

	code, err := crypto.GenerateJoinCode()
	if err != nil {
		return fmt.Errorf("failed to generate join code: %w", err)
	}

	pubKey := identity.Recipient().String()

	// Start broadcast and TCP listener
	session, err := p2p.BroadcastKey(code, pubKey)
	if err != nil {
		return fmt.Errorf("failed to start local broadcast: %w", err)
	}
	defer session.Server.Shutdown()

	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	cmd.Println("📡 Broadcasting identity on local network...")
	cmd.Printf("\n👉 Tell your admin this code: %s\n\n", cyan(code))
	cmd.Println("Waiting for admin... (Press Ctrl+C to stop)")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Wait for EITHER the user to press Ctrl+C, OR the admin to send the ACK
	select {
	case <-sigs:
		cmd.Println("\nBroadcast stopped by user.")
	case <-session.AckChan:
		green := color.New(color.FgGreen, color.Bold).SprintFunc()
		cmd.Printf("\n%s Your key was successfully added to the project!\n", green("✅ Success!"))
	}

	return nil
}
