package commands

import (
	"fmt"
	"os"

	"github.com/flootic/envseal/internal/cli/audit"

	"github.com/spf13/cobra"
)

func NewAuditLogCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "audit-log",
		Short: "View the audit log",
		Long:  "Displays the contents of the audit log at ~/.envseal/audit.log.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuditLog(cmd)
		},
	}
}

func runAuditLog(cmd *cobra.Command) error {
	logPath, err := audit.LogFilePath()
	if err != nil {
		return fmt.Errorf("resolving audit log path: %w", err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			cmd.Println("No audit log entries found.")
			return nil
		}
		return fmt.Errorf("reading audit log: %w", err)
	}

	if len(data) == 0 {
		cmd.Println("No audit log entries found.")
		return nil
	}

	cmd.Print(string(data))
	return nil
}
