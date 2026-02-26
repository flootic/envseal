package config

import (
	"os"
	"path/filepath"
)

const (
	IdentityFileName = "identity"
	Directory        = ".envseal"
)

// GetDefaultIdentityFilePath returns the default path to the identity file
// in the user's home directory under .envseal/identity.
func GetDefaultIdentityFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, Directory, IdentityFileName), nil
}
