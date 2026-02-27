package commands

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

func NewExecCommand(deps Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec -- <command> [args...]",
		Short: "Run a command with injected secrets",
		Long: `Decrypts secrets in memory and starts a child process with them injected.

Examples:
  envseal exec -- npm start
  envseal exec -- python app.py`,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExec(cmd, args, deps)
		},
	}
	return cmd
}

func runExec(cmd *cobra.Command, args []string, deps Deps) error {
	args = stripDoubleDash(args)
	if len(args) == 0 {
		return fmt.Errorf("you must specify a command after '--' (e.g. envseal exec -- npm start)")
	}

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

	vars, err := sf.GetAllSecrets()
	if err != nil {
		return fmt.Errorf("failed to decrypt secrets: %w", err)
	}

	commandName := args[0]
	commandArgs := args[1:]

	binaryPath, err := exec.LookPath(commandName)
	if err != nil {
		return fmt.Errorf("command %q not found in PATH", commandName)
	}

	child := exec.Command(binaryPath, commandArgs...)
	child.Env = mergeEnv(os.Environ(), vars)
	child.Stdin = os.Stdin
	child.Stdout = os.Stdout
	child.Stderr = os.Stderr

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigs)
	defer close(sigs)

	if err := child.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	go forwardSignals(sigs, child)

	if err := child.Wait(); err != nil {
		return exitWithChildCode(err)
	}

	return nil
}

func stripDoubleDash(args []string) []string {
	if len(args) > 0 && args[0] == "--" {
		return args[1:]
	}
	return args
}

func mergeEnv(base []string, vars map[string]string) []string {
	out := make([]string, 0, len(base)+len(vars))
	out = append(out, base...)

	for k, v := range vars {
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}
	return out
}

func forwardSignals(sigs <-chan os.Signal, child *exec.Cmd) {
	for sig := range sigs {
		if child.Process != nil {
			_ = child.Process.Signal(sig)
		}
	}
}

func exitWithChildCode(waitErr error) error {
	if exitErr, ok := waitErr.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			os.Exit(status.ExitStatus())
		}
		os.Exit(1)
	}
	return fmt.Errorf("child process error: %w", waitErr)
}
