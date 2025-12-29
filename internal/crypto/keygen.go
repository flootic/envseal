package crypto

import (
	"os"
	"strings"

	"filippo.io/age"
)

// GenerateIdentity generates a new X25519 identity and its corresponding recipient.
func GenerateIdentity() (string, string, error) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return "", "", err
	}

	return identity.String(), identity.Recipient().String(), nil
}

// GetIdentityFromKeyFile reads an identity from a given file path.
func GetIdentityFromKeyFile(path string) (*age.X25519Identity, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	keyStr := strings.TrimSpace(string(content))

	identity, err := age.ParseX25519Identity(keyStr)
	if err != nil {
		return nil, err
	}

	return identity, nil
}
