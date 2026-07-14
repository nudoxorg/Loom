package project

import (
	"fmt"
	"os"
	"path/filepath"
)

type Project struct {
	Root   string
	Slug   string
	DBPath string
}

// EnsureHome ensures that ~/.loom ~/.loom/projects and ~/.loom/global exist.
// Called at the top of every CLI command.
func EnsureHome() error {
	loomHome, err := HomeDir()
	if err != nil {
		return err
	}

	dirs := []string{
		loomHome,
		filepath.Join(loomHome, "projects"),
		filepath.Join(loomHome, "global"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("creating %s: %w", dir, err)
		}
	}

	return nil
}

// HomeDir returns ~/.loom, resolve user home directory
func HomeDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}

	return filepath.Join(homeDir, ".loom"), nil
}

// GlobalDBPath returns ~/.loom/global/global.db
func GlobalDBPath() (string, error) {
	homeDir, err := HomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}

	return filepath.Join(homeDir, "global", "global.db"), nil
}
