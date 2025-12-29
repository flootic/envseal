package commands

import (
	"filippo.io/age"
	"github.com/xfrr/envseal-cli/internal/config"
	"github.com/xfrr/envseal-cli/internal/crypto"
)

type IdentityStore interface {
	Load(path string) (*age.X25519Identity, error)
	Generate() (privKey string, pubKey string, err error)
}

type SecretsStore interface {
	Load(path string) (*config.SecretFile, error)
}

// ManifestStore abstracts manifest persistence (disk, memory, etc.).
type ManifestStore interface {
	Load() (*config.Manifest, error)
	Save(*config.Manifest) error
}

// Deps groups external dependencies for commands.
type Deps struct {
	ManifestStore ManifestStore

	IdentityManager IdentityStore
	SecretsStore    SecretsStore
}

// ---- Default implementations ----

type fileManifestStore struct{}

func (fileManifestStore) Load() (*config.Manifest, error) { return config.LoadManifest() }
func (fileManifestStore) Save(m *config.Manifest) error   { return m.Save() }

type identityManager struct{}

func (identityManager) Load(path string) (*age.X25519Identity, error) {
	return crypto.GetIdentityFromKeyFile(path)
}
func (identityManager) Generate() (string, string, error) {
	return crypto.GenerateIdentity()
}

type secretsStore struct{}

func (secretsStore) Load(path string) (*config.SecretFile, error) {
	return config.LoadSecretFile(path)
}

// DefaultDeps returns production dependencies.
func DefaultDeps() Deps {
	return Deps{
		ManifestStore:   fileManifestStore{},
		IdentityManager: identityManager{},
		SecretsStore:    secretsStore{},
	}
}
