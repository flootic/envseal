package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"filippo.io/age"
	"gopkg.in/yaml.v3"

	"github.com/xfrr/envseal-cli/internal/crypto"
	"github.com/xfrr/envseal-cli/pkg/filesystem"
)

const (
	DefaultSecretFileName = "secrets.enc.yaml"
	MetadataKey           = "_envseal"
	SecretsKey            = "secrets"

	encPrefix = "ENC[age,chacha20,"
	encSuffix = "]"
)

var (
	ErrLocked          = errors.New("file locked")
	ErrKeyNotFound     = errors.New("key not found")
	ErrAccessDenied    = errors.New("access denied: your private key is not in the recipients list")
	ErrMissingMetadata = errors.New("corrupt or uninitialized file: missing _envseal block")
)

// Recipient represents a single entry in the access control list.
type Recipient struct {
	Arg string `yaml:"arg"` // Public key (recipient identifier)
	Enc string `yaml:"enc"` // Encrypted DEK for this recipient
}

// Metadata defines the structure of the metadata block in the secret file.
type Metadata struct {
	Recipients []Recipient `yaml:"recipients"`
}

// SecretFile represents the file loaded in memory.
type SecretFile struct {
	mu sync.RWMutex

	// RawData contains the entire YAML content (Metadata + Secrets + any extra keys).
	RawData map[string]any

	// decryptedDEK is the 32-byte master key, present only after Unlock()/Init().
	// Unexported to reduce accidental exposure (logging, json/yaml dumps, etc.).
	decryptedDEK []byte

	// File path on disk
	path string
}

// NewSecretFile creates an empty structure ready to initialize.
func NewSecretFile(path string) *SecretFile {
	sf := &SecretFile{
		path:    path,
		RawData: make(map[string]any),
	}
	_, _ = sf.ensureSecretsMap(true)
	return sf
}

// LoadSecretFile reads the file from disk without decrypting the DEK yet.
func LoadSecretFile(path string) (*SecretFile, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return NewSecretFile(path), nil
	}
	if err != nil {
		return nil, err
	}

	raw := make(map[string]any)
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	sf := &SecretFile{
		path:    path,
		RawData: raw,
	}
	_, _ = sf.ensureSecretsMap(true)
	return sf, nil
}

// IsUnlocked indicates whether the file currently holds a DEK in memory.
func (sf *SecretFile) IsUnlocked() bool {
	sf.mu.RLock()
	defer sf.mu.RUnlock()
	return sf.decryptedDEK != nil
}

// Lock wipes the DEK from memory and locks the file.
func (sf *SecretFile) Lock() {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	zeroBytes(sf.decryptedDEK)
	sf.decryptedDEK = nil
}

// Unlock attempts to obtain the DEK by decrypting one of the recipient entries.
func (sf *SecretFile) Unlock(identity *age.X25519Identity) error {
	if identity == nil {
		return errors.New("identity is nil")
	}

	sf.mu.Lock()
	defer sf.mu.Unlock()

	metaInterface, ok := sf.RawData[MetadataKey]
	if !ok {
		return ErrMissingMetadata
	}

	metaBytes, err := yaml.Marshal(metaInterface)
	if err != nil {
		return fmt.Errorf("error encoding metadata: %w", err)
	}

	var meta Metadata
	if err := yaml.Unmarshal(metaBytes, &meta); err != nil {
		return fmt.Errorf("error parsing metadata: %w", err)
	}

	for _, recipient := range meta.Recipients {
		dek, err := crypto.DecryptDEK(recipient.Enc, identity)
		if err == nil {
			// Replace any previous key securely.
			zeroBytes(sf.decryptedDEK)
			sf.decryptedDEK = cloneBytes(dek)
			return nil
		}
	}

	return ErrAccessDenied
}

// Init initializes a new file by generating a new DEK and setting recipients.
func (sf *SecretFile) Init(initialRecipients []string) error {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	if len(initialRecipients) == 0 {
		return errors.New("initial recipients list cannot be empty")
	}

	dek, err := crypto.GenerateDEK()
	if err != nil {
		return err
	}

	// Replace any previous key securely.
	zeroBytes(sf.decryptedDEK)
	sf.decryptedDEK = cloneBytes(dek)

	_, _ = sf.ensureSecretsMap(true)
	return sf.rotateRecipientsLocked(initialRecipients)
}

// RotateRecipients updates who has access (rekey).
func (sf *SecretFile) RotateRecipients(publicKeys []string) error {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	if sf.decryptedDEK == nil {
		return errors.New("cannot rotate recipients without unlocking the file first")
	}
	return sf.rotateRecipientsLocked(publicKeys)
}

func (sf *SecretFile) rotateRecipientsLocked(publicKeys []string) error {
	publicKeys = normalizeAndDedupe(publicKeys)
	if len(publicKeys) == 0 {
		return errors.New("recipients list cannot be empty")
	}

	newRecipients := make([]Recipient, 0, len(publicKeys))
	for _, pubKey := range publicKeys {
		encDEK, err := crypto.EncryptDEK(sf.decryptedDEK, []string{pubKey})
		if err != nil {
			return fmt.Errorf("failed to encrypt for recipient %q: %w", pubKey, err)
		}
		newRecipients = append(newRecipients, Recipient{Arg: pubKey, Enc: encDEK})
	}

	sf.RawData[MetadataKey] = Metadata{Recipients: newRecipients}
	return nil
}

// SetSecret encrypts a value and stores it under the canonical `secrets:` map.
func (sf *SecretFile) SetSecret(key, value string) error {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	if sf.decryptedDEK == nil {
		return ErrLocked
	}

	key = strings.TrimSpace(key)
	if key == "" {
		return errors.New("key cannot be empty")
	}
	if key == MetadataKey || key == SecretsKey {
		return fmt.Errorf("cannot use reserved name %q", key)
	}

	encryptedVal, err := crypto.EncryptValue(value, sf.decryptedDEK)
	if err != nil {
		return err
	}

	secrets, err := sf.ensureSecretsMap(true)
	if err != nil {
		return err
	}

	secrets[key] = wrapEncrypted(encryptedVal)

	// Backwards-compat: if legacy top-level exists, keep canonical and remove legacy.
	delete(sf.RawData, key)

	return nil
}

// UnsetSecret removes a key from `secrets:` (and from legacy top-level if present).
func (sf *SecretFile) UnsetSecret(key string) error {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	if sf.decryptedDEK == nil {
		return ErrLocked
	}

	key = strings.TrimSpace(key)
	if key == "" {
		return errors.New("key cannot be empty")
	}
	if key == MetadataKey || key == SecretsKey {
		return fmt.Errorf("cannot unset reserved key %q", key)
	}

	secrets, err := sf.ensureSecretsMap(true)
	if err != nil {
		return err
	}

	_, inSecrets := secrets[key]
	_, inLegacy := sf.RawData[key]
	if !inSecrets && !inLegacy {
		return fmt.Errorf("key %q does not exist", key)
	}

	delete(secrets, key)
	delete(sf.RawData, key) // legacy
	return nil
}

// GetSecret retrieves and decrypts a value.
// It first checks `secrets:`, then falls back to legacy top-level keys.
func (sf *SecretFile) GetSecret(key string) (string, error) {
	sf.mu.RLock()
	defer sf.mu.RUnlock()

	if sf.decryptedDEK == nil {
		return "", ErrLocked
	}

	key = strings.TrimSpace(key)
	if key == "" {
		return "", errors.New("key cannot be empty")
	}
	if key == MetadataKey || key == SecretsKey {
		return "", fmt.Errorf("reserved key %q cannot be retrieved as a secret", key)
	}

	// Canonical location: secrets map
	if secrets, err := sf.ensureSecretsMap(false); err == nil && secrets != nil {
		if v, ok := secrets[key]; ok {
			return sf.decryptAnyLocked(v)
		}
	}

	// Backwards-compat: legacy top-level secret
	if v, ok := sf.RawData[key]; ok {
		return sf.decryptAnyLocked(v)
	}

	return "", ErrKeyNotFound
}

func (sf *SecretFile) decryptAnyLocked(v any) (string, error) {
	s, ok := v.(string)
	if !ok {
		return "", errors.New("value is not a string")
	}
	return decryptIfNeeded(s, sf.decryptedDEK)
}

// Save writes the entire RawData map to disk (0600) using an atomic write.
func (sf *SecretFile) Save() error {
	sf.mu.RLock()
	data, err := yaml.Marshal(sf.RawData)
	sf.mu.RUnlock()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(sf.path), 0o755); err != nil {
		return err
	}

	return filesystem.AtomicWriteFile(sf.path, data, 0o600)
}

// GetAllSecrets returns a map with all decrypted secrets.
// It reads from `secrets:` and also includes legacy top-level entries (excluding reserved keys).
func (sf *SecretFile) GetAllSecrets() (map[string]string, error) {
	sf.mu.RLock()
	defer sf.mu.RUnlock()

	if sf.decryptedDEK == nil {
		return nil, ErrLocked
	}

	out := make(map[string]string)

	// Canonical secrets map
	if secrets, err := sf.ensureSecretsMap(false); err == nil && secrets != nil {
		for k, v := range secrets {
			val, err := sf.decryptAnyLocked(v)
			if err != nil {
				if s, ok := v.(string); ok {
					out[k] = s
				} else {
					out[k] = fmt.Sprintf("%v", v)
				}
				continue
			}
			out[k] = val
		}
	}

	// Legacy top-level secrets (exclude reserved keys and nested maps)
	for k, v := range sf.RawData {
		if k == MetadataKey || k == SecretsKey {
			continue
		}
		if _, already := out[k]; already {
			continue
		}
		s, ok := v.(string)
		if !ok {
			continue
		}
		val, err := decryptIfNeeded(s, sf.decryptedDEK)
		if err != nil {
			out[k] = s
			continue
		}
		out[k] = val
	}

	return out, nil
}

// ensureSecretsMap returns the `secrets:` map, optionally creating it.
func (sf *SecretFile) ensureSecretsMap(create bool) (map[string]any, error) {
	raw, ok := sf.RawData[SecretsKey]
	if !ok || raw == nil {
		if !create {
			return nil, nil
		}
		m := make(map[string]any)
		sf.RawData[SecretsKey] = m
		return m, nil
	}

	m, ok := raw.(map[string]any)
	if ok {
		return m, nil
	}

	// YAML decoding can produce map[interface{}]interface{} in nested structures.
	if legacy, ok := raw.(map[any]any); ok {
		converted := make(map[string]any, len(legacy))
		for k, v := range legacy {
			ks, ok := k.(string)
			if !ok {
				return nil, errors.New("invalid secrets format: non-string key")
			}
			converted[ks] = v
		}
		sf.RawData[SecretsKey] = converted
		return converted, nil
	}

	return nil, errors.New("invalid secrets format")
}

func wrapEncrypted(cipherText string) string {
	return encPrefix + cipherText + encSuffix
}

func decryptIfNeeded(value string, dek []byte) (string, error) {
	if !strings.HasPrefix(value, encPrefix) || !strings.HasSuffix(value, encSuffix) {
		return value, nil
	}
	cipherText := value[len(encPrefix) : len(value)-len(encSuffix)]
	return crypto.DecryptValue(cipherText, dek)
}

func normalizeAndDedupe(keys []string) []string {
	set := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		set[k] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func zeroBytes(b []byte) {
	if b == nil {
		return
	}
	for i := range b {
		b[i] = 0
	}
}

func cloneBytes(b []byte) []byte {
	if b == nil {
		return nil
	}
	out := make([]byte, len(b))
	copy(out, b)
	return out
}
