package commands

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	doctorStartMsg     = "🩺 Running EnvSeal Doctor..."
	doctorFailedMsg    = "❌ Doctor found issues that need attention."
	doctorSuccessMsg   = "✅ Everything looks good!"
	doctorChecksFailed = "doctor checks failed"

	identityPerms = 0o600
)

var errDoctorChecksFailed = errors.New(doctorChecksFailed)

type doctorCheck struct {
	name string
	fn   func() error
}

func NewDoctorCommand(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Troubleshoot issues",
		Long:  "Diagnose common configuration and permission issues in the current environment.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctor(cmd, deps)
		},
	}
}

func runDoctor(cmd *cobra.Command, deps Deps) error {
	cmd.Println(doctorStartMsg)

	checks := buildDoctorChecks(deps)
	hasErrors := runDoctorChecks(cmd, checks)

	cmd.Println()
	if hasErrors {
		cmd.Println(color.RedString(doctorFailedMsg))
		return errDoctorChecksFailed
	}

	cmd.Println(color.GreenString(doctorSuccessMsg))
	return nil
}

func buildDoctorChecks(deps Deps) []doctorCheck {
	checks := []doctorCheck{
		{
			name: "Local Identity",
			fn:   checkIdentityExists,
		},
	}

	if runtime.GOOS != "windows" {
		checks = append(checks, doctorCheck{
			name: "Identity Permissions",
			fn:   checkIdentityPermissions,
		})
	}

	checks = append(checks,
		doctorCheck{
			name: "Project Manifest",
			fn:   checkProjectManifest(deps),
		},
		doctorCheck{
			name: fmt.Sprintf("Access to %s", secretFilePath),
			fn:   checkSecretsAccess(deps),
		},
	)

	return checks
}

func runDoctorChecks(cmd *cobra.Command, checks []doctorCheck) (hasErrors bool) {
	for _, c := range checks {
		cmd.Printf("Checking %-22s ... ", c.name)

		if err := c.fn(); err != nil {
			cmd.Println(color.RedString("FAILED"))
			cmd.Printf("  ↳ %v\n", err)
			hasErrors = true
			continue
		}

		cmd.Println(color.GreenString("OK"))
	}
	return hasErrors
}

func checkIdentityExists() error {
	_, err := os.Stat(identityFilePath)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		return fmt.Errorf("file not found at %s (run 'envseal init')", identityFilePath)
	}
	return err
}

func checkIdentityPermissions() error {
	info, err := os.Stat(identityFilePath)
	if err != nil {
		return err
	}

	perm := info.Mode().Perm()
	if perm != identityPerms {
		return fmt.Errorf(
			"permissions are %o (should be 600). Run: chmod 600 %s",
			perm,
			identityFilePath,
		)
	}
	return nil
}
