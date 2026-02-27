package crypto

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GenerateJoinCode() (string, error) {
	codeInt, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", fmt.Errorf("failed to generate random code: %w", err)
	}
	return fmt.Sprintf("%d", codeInt.Int64()+100000), nil
}
