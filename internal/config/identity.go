package config

import (
	"os"
	"path/filepath"
)

const (
	DefaultIdentityFileName = "identity"
	DefaultIdentityDir      = ".envseal"
)

// GetDefaultIdentityFilePath returns the default path to the identity file
// in the user's home directory under .envseal/identity.
func GetDefaultIdentityFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, DefaultIdentityDir, DefaultIdentityFileName), nil
}
