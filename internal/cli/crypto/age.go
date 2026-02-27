package crypto

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"filippo.io/age"
	"golang.org/x/crypto/chacha20poly1305"
)

const (
	dekSize = chacha20poly1305.KeySize // 32 bytes

	errDecryptDEKDenied = "failed to decrypt DEK (denied or invalid identity)"
	errCiphertextShort  = "ciphertext too short"
	errDecryptValue     = "failed to decrypt value (possible corruption or incorrect DEK)"
)

// GenerateDEK creates a random 32-byte key (ChaCha20 key).
func GenerateDEK() ([]byte, error) {
	key := make([]byte, dekSize)
	if err := fillRandom(key); err != nil {
		return nil, err
	}
	return key, nil
}

// EncryptDEK encrypts the master key (DEK) for a list of recipients (Users).
// Returns an ASCII-armored age string.
func EncryptDEK(dek []byte, recipientPubKeys []string) (string, error) {
	if err := validateDEK(dek); err != nil {
		return "", err
	}

	recipients, err := parseRecipients(recipientPubKeys)
	if err != nil {
		return "", err
	}

	var out bytes.Buffer
	if err := encryptToAge(&out, dek, recipients); err != nil {
		return "", err
	}

	return out.String(), nil
}

// DecryptDEK decrypts the "envelope" to obtain the raw DEK using the user's identity.
func DecryptDEK(encryptedDEK string, identity age.Identity) ([]byte, error) {
	if encryptedDEK == "" {
		return nil, errors.New(errDecryptDEKDenied)
	}
	if identity == nil {
		return nil, errors.New(errDecryptDEKDenied)
	}

	r, err := age.Decrypt(bytes.NewBufferString(encryptedDEK), identity)
	if err != nil {
		// Avoid leaking details; keep behavior consistent.
		return nil, errors.New(errDecryptDEKDenied)
	}

	var dek bytes.Buffer
	if _, err := io.Copy(&dek, r); err != nil {
		return nil, err
	}

	b := dek.Bytes()
	if err := validateDEK(b); err != nil {
		return nil, errors.New(errDecryptDEKDenied)
	}

	// Return a detached copy to avoid holding onto the buffer's underlying array.
	out := make([]byte, len(b))
	copy(out, b)
	return out, nil
}

// EncryptValue encrypts a string using ChaCha20-Poly1305 with the DEK.
// Output format: Base64(Nonce + Ciphertext)
func EncryptValue(plaintext string, dek []byte) (string, error) {
	if err := validateDEK(dek); err != nil {
		return "", err
	}

	aead, err := chacha20poly1305.New(dek)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aead.NonceSize())
	if err := fillRandom(nonce); err != nil {
		return "", err
	}

	// Build output as Nonce || Ciphertext in a single slice (no aliasing tricks).
	out := make([]byte, 0, len(nonce)+len(plaintext)+aead.Overhead())
	out = append(out, nonce...)
	out = aead.Seal(out, nonce, []byte(plaintext), nil)

	return base64.StdEncoding.EncodeToString(out), nil
}

// DecryptValue decrypts a Base64 string using the DEK.
func DecryptValue(encryptedBase64 string, dek []byte) (string, error) {
	if err := validateDEK(dek); err != nil {
		return "", err
	}

	data, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return "", err
	}

	aead, err := chacha20poly1305.New(dek)
	if err != nil {
		return "", err
	}

	nonceSize := aead.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New(errCiphertextShort)
	}

	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New(errDecryptValue)
	}

	return string(plaintext), nil
}

func parseRecipients(pubKeys []string) ([]age.Recipient, error) {
	if len(pubKeys) == 0 {
		return nil, errors.New("no recipients provided")
	}

	recipients := make([]age.Recipient, 0, len(pubKeys))
	for _, pubKey := range pubKeys {
		r, err := age.ParseX25519Recipient(pubKey)
		if err != nil {
			return nil, fmt.Errorf("invalid public key: %w", err)
		}
		recipients = append(recipients, r)
	}
	return recipients, nil
}

func encryptToAge(w io.Writer, dek []byte, recipients []age.Recipient) error {
	aw, err := age.Encrypt(w, recipients...)
	if err != nil {
		return err
	}
	if _, err := aw.Write(dek); err != nil {
		_ = aw.Close()
		return err
	}
	return aw.Close()
}

func validateDEK(dek []byte) error {
	if len(dek) != dekSize {
		return fmt.Errorf("invalid DEK size: got %d, want %d", len(dek), dekSize)
	}
	return nil
}

func fillRandom(b []byte) error {
	_, err := io.ReadFull(rand.Reader, b)
	return err
}
