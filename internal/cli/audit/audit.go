package audit

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/flootic/envseal/internal/cli/config"
)

const LogFileName = "audit.log"

// Log appends an entry to ~/.envseal/audit.log.
// Format: <timestamp> <username> <action> <message>
func Log(action string, message string) error {
	logPath, err := LogFilePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(logPath), 0700); err != nil {
		return fmt.Errorf("creating audit log directory: %w", err)
	}

	username := currentUsername()
	timestamp := time.Now().UTC().Format(time.RFC3339)

	entry := fmt.Sprintf("%s %s %s %s\n", timestamp, username, action, message)

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("opening audit log: %w", err)
	}
	defer f.Close()

	_, err = f.WriteString(entry)
	return err
}

// LogFilePath returns the path to the audit log file.
func LogFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, config.Directory, LogFileName), nil
}

// LogCommand logs a CLI command invocation.
func LogCommand(args []string) error {
	if len(args) == 0 {
		return nil
	}
	action := args[0]
	message := strings.Join(args, " ")
	return Log(action, message)
}

// SanitizeArgs redacts values from KEY=VALUE arguments that could contain secrets.
func SanitizeArgs(args []string) []string {
	out := make([]string, len(args))
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			out[i] = arg
			continue
		}
		if key, _, ok := strings.Cut(arg, "="); ok {
			out[i] = key + "=***"
		} else {
			out[i] = arg
		}
	}
	return out
}

func currentUsername() string {
	u, err := user.Current()
	if err != nil {
		return "unknown"
	}
	return u.Username
}
