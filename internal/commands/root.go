package commands

import (
	"fmt"

	"github.com/xfrr/envseal-cli/internal/config"

	"github.com/spf13/cobra"
)

var (
	secretFilePath   string
	identityFilePath string
)

func Execute() error {
	rootCmd := &cobra.Command{
		Use:     "envseal",
		Short:   "Secure Git-native secrets management for teams.",
		Long:    "EnvSeal is a CLI tool to manage encrypted secrets files in your git repositories. It allows teams to securely share secrets without relying on external services.",
		Version: "v0.1.0",
	}

	rootCmd.PersistentFlags().StringVarP(
		&secretFilePath,
		"file",
		"f",
		config.DefaultSecretFileName,
		"Encrypted file path (e.g. secrets.enc.yaml, secrets.prod.enc.yaml)"+
			" to use for storing the secrets.",
	)

	defaultIdentityFilePath, err := config.GetDefaultIdentityFilePath()
	if err != nil {
		return fmt.Errorf("failed to get default identity file path: %w", err)
	}

	rootCmd.PersistentFlags().StringVarP(
		&identityFilePath,
		"identity",
		"i",
		defaultIdentityFilePath,
		fmt.Sprintf("Path to the identity key file (defaults to %s).", defaultIdentityFilePath),
	)

	deps := DefaultDeps()

	rootCmd.AddCommand(NewInitCommand(deps))
	rootCmd.AddCommand(NewExecCommand(deps))
	rootCmd.AddCommand(NewSetCommand(deps))
	rootCmd.AddCommand(NewUnsetCommand(deps))
	rootCmd.AddCommand(NewUsersCommand(deps))
	rootCmd.AddCommand(NewRekeyCommand(deps))
	rootCmd.AddCommand(NewDoctorCommand(deps))
	rootCmd.AddCommand(NewPrintCommand(deps))
	rootCmd.AddCommand(NewWhoamiCommand(deps))
	return rootCmd.Execute()
}
