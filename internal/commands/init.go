package commands

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/xfrr/envseal-cli/internal/config"
	"github.com/xfrr/envseal-cli/pkg/filesystem"
)

func NewInitCommand(deps Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize EnvSeal in the current directory",
		Long: `Initializes the current directory for EnvSeal:
- Ensures your local identity exists
- Creates envseal.yaml (manifest)
- Creates secrets.enc.yaml (encrypted secrets file)`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd, deps)
		},
	}

	cmd.Flags().String("name", "", "Project name for envseal.yaml (default: current directory name)")
	return cmd
}

func runInit(cmd *cobra.Command, deps Deps) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	cmd.Println("🚀 Initializing EnvSeal...")

	pubKey, createdIdentity, err := ensureIdentity(cmd, deps)
	if err != nil {
		return err
	}
	if createdIdentity {
		cmd.Println(green("✓ Identity created"))
	} else {
		cmd.Println(green("✓ Identity loaded"))
	}

	projectName, err := resolveProjectName(cmd)
	if err != nil {
		return err
	}

	manifestCreated, err := ensureManifest(cmd, deps, projectName, pubKey)
	if err != nil {
		return err
	}
	if !manifestCreated {
		cmd.Println(yellow("⚠️  envseal.yaml already exists. Skipping."))
	} else {
		cmd.Println(green("✓ envseal.yaml created"))
	}

	secretsCreated, err := ensureSecretsFile(cmd, deps, pubKey)
	if err != nil {
		return err
	}
	if !secretsCreated {
		cmd.Println(yellow("⚠️  secrets.enc.yaml already exists. Skipping."))
	} else {
		cmd.Println(green("✓ secrets.enc.yaml created"))
	}

	cmd.Println()
	cmd.Println(green("✅ Initialized successfully."))
	cmd.Printf("Public key: %s\n", bold(pubKey))
	cmd.Printf("Try: %s\n", bold("envseal print"))

	return nil
}

func ensureIdentity(cmd *cobra.Command, deps Deps) (pubKey string, created bool, err error) {
	if _, err := os.Stat(identityFilePath); err == nil {
		id, err := deps.IdentityManager.Load(identityFilePath)
		if err != nil {
			return "", false, fmt.Errorf("failed to read identity: %w", err)
		}
		return id.Recipient().String(), false, nil
	} else if !os.IsNotExist(err) {
		return "", false, fmt.Errorf("failed to stat identity: %w", err)
	}

	// Identity missing -> create.
	if err := os.MkdirAll(filepath.Dir(identityFilePath), 0o700); err != nil {
		return "", false, fmt.Errorf("failed to create identity directory: %w", err)
	}

	// Use deps so it is testable.
	priv, pub, err := deps.IdentityManager.Generate()
	if err != nil {
		return "", false, fmt.Errorf("failed to generate identity: %w", err)
	}

	if err := filesystem.AtomicWriteFile(identityFilePath, []byte(priv), 0o600); err != nil {
		return "", false, fmt.Errorf("failed to save identity: %w", err)
	}

	return pub, true, nil
}

func resolveProjectName(cmd *cobra.Command) (string, error) {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return "", err
	}
	name = strings.TrimSpace(name)
	if name != "" {
		return name, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	base := filepath.Base(cwd)
	base = strings.TrimSpace(base)
	if base == "" || base == "." || base == string(filepath.Separator) {
		return "", errors.New("could not infer project name; provide --name")
	}

	return base, nil
}

func ensureManifest(cmd *cobra.Command, deps Deps, projectName, pubKey string) (created bool, err error) {
	if _, err := os.Stat(config.ManifestFileName); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("failed to stat %s: %w", config.ManifestFileName, err)
	}

	userName := "admin"
	if u, err := user.Current(); err == nil && strings.TrimSpace(u.Username) != "" {
		userName = u.Username
	}

	m := &config.Manifest{
		ProjectName: projectName,
		AccessControl: []config.User{
			{Name: userName, PublicKey: pubKey},
		},
	}

	if err := deps.ManifestStore.Save(m); err != nil {
		return false, fmt.Errorf("failed to save manifest: %w", err)
	}

	return true, nil
}

func ensureSecretsFile(cmd *cobra.Command, deps Deps, pubKey string) (created bool, err error) {
	if _, err := os.Stat(secretFilePath); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("failed to stat %s: %w", secretFilePath, err)
	}

	sf := config.NewSecretFile(secretFilePath)
	if err := sf.Init([]string{pubKey}); err != nil {
		return false, fmt.Errorf("failed to initialize secrets: %w", err)
	}
	if err := sf.SetSecret("HELLO", "World (Encrypted with EnvSeal)"); err != nil {
		return false, fmt.Errorf("failed to set example secret: %w", err)
	}
	if err := sf.Save(); err != nil {
		return false, fmt.Errorf("failed to save %s: %w", secretFilePath, err)
	}

	return true, nil
}
