package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/xfrr/envseal-cli/internal/config"
)

func checkProjectManifest(deps Deps) func() error {
	return func() error {
		if _, err := os.Stat(config.ManifestFileName); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("%s not found in current directory", config.ManifestFileName)
			}
			return err
		}

		if _, err := deps.ManifestStore.Load(); err != nil {
			return fmt.Errorf("invalid manifest: %w", err)
		}

		return nil
	}
}

func checkSecretsAccess(deps Deps) func() error {
	return func() error {
		id, err := deps.IdentityManager.Load(identityFilePath)
		if err != nil {
			return fmt.Errorf("cannot load identity: %w", err)
		}

		sf, err := deps.SecretsStore.Load(secretFilePath)
		if err != nil {
			return fmt.Errorf("cannot load secrets file: %w", err)
		}

		// Ensure we always attempt to re-lock if Unlock succeeded.
		locked := true
		defer func() {
			if !locked {
				sf.Lock()
			}
		}()

		if err := sf.Unlock(id); err != nil {
			return errors.New("access denied: your key cannot decrypt this file (ask admin to run 'envseal rekey')")
		}
		locked = false

		return nil
	}
}
