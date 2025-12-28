package commands

import (
	"github.com/spf13/cobra"
)

func Execute() error {
	rootCmd := &cobra.Command{
		Use:     "envseal",
		Short:   "Secure Git-native secrets management for teams.",
		Long:    "EnvSeal is a CLI tool to manage encrypted secrets files in your git repositories. It allows teams to securely share secrets without relying on external services.",
		Version: "v0.1.0",
	}

	return rootCmd.Execute()
}
