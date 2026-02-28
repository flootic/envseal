package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const hookMarkerStart = "# --- envseal pre-commit hook start ---"
const hookMarkerEnd = "# --- envseal pre-commit hook end ---"

// preCommitPayload is the pure shell logic without the shebang
const preCommitPayload = `
if command -v envseal >/dev/null 2>&1; then
    MANIFEST_STAGED=$(git diff --cached --name-only | grep "^envseal\.yaml$")
    if [ -n "$MANIFEST_STAGED" ]; then
        ALL_VAULTS=$(git ls-files | grep "\.enc\.yaml$")
        for vault in $ALL_VAULTS; do
            IS_STAGED=$(git diff --cached --name-only | grep "^$vault$")
            if [ -z "$IS_STAGED" ]; then
                echo "❌ envseal: Manifest (envseal.yaml) modified but vault ($vault) is not."
                echo "   Did you forget to run 'envseal rekey --file $vault'?"
                exit 1
            fi
        done
    fi
fi
`

func NewHookCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hook",
		Short: "Manage Git integration hooks",
	}
	cmd.AddCommand(newHookInstallCommand())
	return cmd
}

func newHookInstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install the pre-commit Git hook",
		Long:  "Installs a Git pre-commit hook to prevent committing desynchronized files. Safely appends to existing hooks.",
		Args:  cobra.NoArgs,
		RunE:  runHookInstall,
	}
}

func runHookInstall(cmd *cobra.Command, args []string) error {
	gitDir := ".git"
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("not a git repository (or any of the parent directories)")
	}

	hooksDir := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	hookPath := filepath.Join(hooksDir, "pre-commit")

	// Create the well-defined block of code
	hookBlock := fmt.Sprintf("%s\n%s\n%s\n", hookMarkerStart, strings.TrimSpace(preCommitPayload), hookMarkerEnd)

	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	// 1. If the file doesn't exist, create it from scratch with a shebang
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		fullScript := fmt.Sprintf("#!/bin/sh\n\n%s", hookBlock)
		if err := os.WriteFile(hookPath, []byte(fullScript), 0755); err != nil {
			return fmt.Errorf("failed to write pre-commit hook: %w", err)
		}
		cmd.Printf("%s Git pre-commit hook created and installed at %s\n", green("✓"), hookPath)
		return nil
	}

	// 2. File exists, let's read it to check for idempotency
	content, err := os.ReadFile(hookPath)
	if err != nil {
		return fmt.Errorf("failed to read existing hook: %w", err)
	}

	contentStr := string(content)

	// Check if our block is already installed
	if strings.Contains(contentStr, hookMarkerStart) {
		cmd.Printf("%s envseal pre-commit hook is already installed in %s\n", green("✓"), hookPath)
		return nil
	}

	// 3. Append to existing file securely
	file, err := os.OpenFile(hookPath, os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		return fmt.Errorf("failed to open existing hook for appending: %w", err)
	}
	defer file.Close()

	// Ensure we start on a new line just in case the existing file didn't end with one
	prefix := "\n"
	if len(contentStr) > 0 && !strings.HasSuffix(contentStr, "\n") {
		prefix = "\n\n"
	}

	if _, err := file.WriteString(prefix + hookBlock); err != nil {
		return fmt.Errorf("failed to append to pre-commit hook: %w", err)
	}

	// Ensure it has execution permissions (sometimes existing hooks lose them)
	if err := os.Chmod(hookPath, 0755); err != nil {
		return fmt.Errorf("failed to make hook executable: %w", err)
	}

	cmd.Printf("%s envseal pre-commit hook safely appended to %s\n", green("✓"), hookPath)

	// Crucial warning for bash script edge cases
	cmd.Println(yellow("⚠️  Note: If your existing hook script ends with an 'exit' command,"))
	cmd.Println(yellow("   you may need to manually open the file and move the envseal block above it."))

	return nil
}
