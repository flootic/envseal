// manifest.go
package config

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/xfrr/envseal-cli/pkg/filesystem"
	"gopkg.in/yaml.v3"
)

const ManifestFileName = "envseal.yaml"

var (
	ErrUserExists    = errors.New("a user with this public key already exists")
	ErrUserNotFound  = errors.New("user not found")
	ErrInvalidName   = errors.New("name cannot be empty")
	ErrInvalidPubKey = errors.New("public key cannot be empty")
)

// User represents a user entry in the manifest.
type User struct {
	Name      string `yaml:"name"`
	PublicKey string `yaml:"public_key"`
}

// Manifest maps the structure of the envseal.yaml file.
//
// Note: methods are made concurrency-safe with an internal mutex.
// If you don't need concurrency, you can remove the mutex without changing behavior.
type Manifest struct {
	mu sync.RWMutex `yaml:"-"`

	ProjectName   string `yaml:"project_name"`
	AccessControl []User `yaml:"access_control"`
}

// LoadManifest reads and parses the configuration file from disk.
func LoadManifest() (*Manifest, error) {
	data, err := os.ReadFile(ManifestFileName)
	if os.IsNotExist(err) {
		// Return an empty manifest by default if it doesn't exist.
		return &Manifest{
			AccessControl: make([]User, 0),
		}, nil
	}
	if err != nil {
		return nil, err
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	// Normalize after loading (trim fields, dedupe, stable ordering).
	m.normalizeInPlace()

	return &m, nil
}

// Save writes the manifest to disk with safe permissions (0600) using an atomic write.
func (m *Manifest) Save() error {
	m.mu.RLock()
	// Work on a copy to avoid holding the lock across marshaling I/O if desired.
	// (Marshalling is pure CPU, but keeping it simple and safe.)
	data, err := yaml.Marshal(m)
	m.mu.RUnlock()
	if err != nil {
		return err
	}

	// Ensure directory exists (useful if ManifestFileName becomes configurable later).
	if err := os.MkdirAll(filepath.Dir(ManifestFileName), 0o755); err != nil {
		return err
	}

	return filesystem.AtomicWriteFile(ManifestFileName, data, 0o600)
}

// AddUser adds a user avoiding duplicate public keys.
func (m *Manifest) AddUser(name, pubKey string) error {
	name = strings.TrimSpace(name)
	pubKey = strings.TrimSpace(pubKey)

	if name == "" {
		return ErrInvalidName
	}
	if pubKey == "" {
		return ErrInvalidPubKey
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, u := range m.AccessControl {
		if u.PublicKey == pubKey {
			return ErrUserExists
		}
	}

	m.AccessControl = append(m.AccessControl, User{
		Name:      name,
		PublicKey: pubKey,
	})

	// Keep stable ordering for deterministic files.
	sort.Slice(m.AccessControl, func(i, j int) bool {
		return m.AccessControl[i].Name < m.AccessControl[j].Name
	})

	return nil
}

// RemoveUser removes a user by name or public key.
// Returns true if a user was removed.
func (m *Manifest) RemoveUser(identifier string) bool {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return false
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	newUsers := make([]User, 0, len(m.AccessControl))
	found := false

	for _, u := range m.AccessControl {
		if u.Name == identifier || u.PublicKey == identifier {
			found = true
			continue
		}
		newUsers = append(newUsers, u)
	}

	if found {
		m.AccessControl = newUsers
	}

	return found
}

// RemoveUserStrict removes a user by name or public key and returns an error if not found.
func (m *Manifest) RemoveUserStrict(identifier string) error {
	if ok := m.RemoveUser(identifier); !ok {
		return ErrUserNotFound
	}
	return nil
}

// GetPublicKeys extracts only the public keys (string slice) for cryptographic operations.
// The returned slice is a copy.
func (m *Manifest) GetPublicKeys() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0, len(m.AccessControl))
	for _, u := range m.AccessControl {
		if u.PublicKey == "" {
			continue
		}
		keys = append(keys, u.PublicKey)
	}
	return keys
}

// FindUserByPublicKey returns the user and true if found.
func (m *Manifest) FindUserByPublicKey(pubKey string) (User, bool) {
	pubKey = strings.TrimSpace(pubKey)

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, u := range m.AccessControl {
		if u.PublicKey == pubKey {
			return u, true
		}
	}
	return User{}, false
}

// normalizeInPlace trims fields, removes duplicates by public key (keeping first), and sorts by Name.
func (m *Manifest) normalizeInPlace() {
	m.mu.Lock()
	defer m.mu.Unlock()

	seen := make(map[string]struct{}, len(m.AccessControl))
	out := make([]User, 0, len(m.AccessControl))

	for _, u := range m.AccessControl {
		u.Name = strings.TrimSpace(u.Name)
		u.PublicKey = strings.TrimSpace(u.PublicKey)

		if u.PublicKey == "" {
			// Ignore invalid entries rather than failing hard on load.
			continue
		}

		if _, ok := seen[u.PublicKey]; ok {
			continue
		}
		seen[u.PublicKey] = struct{}{}
		out = append(out, u)
	}

	sort.Slice(out, func(i, j int) bool {
		// Primary by name, secondary by public key for determinism.
		if out[i].Name == out[j].Name {
			return out[i].PublicKey < out[j].PublicKey
		}
		return out[i].Name < out[j].Name
	})

	m.AccessControl = out
}
