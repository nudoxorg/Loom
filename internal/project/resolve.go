package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// Resolve walks up from a given startDir looking for a .git to determine project root.
// If none is found, startDir itself becomes an ad-hoc project root, so non-git
// directories still get their own Loom project rather than erroring out.
// Returns fully populated Project (Root, Slug, DBPath)
func Resolve(startDir string) (*Project, error) {
	root, err := findRoot(startDir)
	if err != nil {
		return nil, err
	}

	slug := Slug(root)

	dbPath, err := dbPathForSlug(root, slug)
	if err != nil {
		return nil, err
	}

	return &Project{Root: root, Slug: slug, DBPath: dbPath}, nil
}

// findRoot walks upward from startDir looking for a .git entry. If none is
// found by the time it reaches the filesystem root, startDir itself (made
// absolute) is returned as an ad-hoc project root.
func findRoot(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("resolving path: %w", err)
	}
	start := dir

	for {
		_, err := os.Stat(filepath.Join(dir, ".git"))
		if err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return start, nil
		}
		dir = parent
	}
}

// dbPathForSlug builds ~/.loom/projects/<slug>/loom.db.
// Ensures that project's subdirectory exists, and records the real
// absolute path in a "root" sidecar file so ListAll can show a human
// label instead of the (lossily) slugified directory name.
func dbPathForSlug(root, slug string) (string, error) {
	homeDir, err := HomeDir()
	if err != nil {
		return "", err
	}

	projectsDir := filepath.Join(homeDir, "projects", slug)
	if err := os.MkdirAll(projectsDir, 0o700); err != nil {
		return "", fmt.Errorf("creating project dir: %w", err)
	}

	rootFile := filepath.Join(projectsDir, "root")
	if err := os.WriteFile(rootFile, []byte(root), 0o600); err != nil {
		return "", fmt.Errorf("writing project root marker: %w", err)
	}

	return filepath.Join(projectsDir, "loom.db"), nil
}
