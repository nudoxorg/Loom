package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// Resolve walks up from a given startDir looking for a .git to determine project root.
// If none is found throw error for not being initialized as a git repository.
// Returns fully populated Project (Root, Slug, DBPath)
func Resolve(startDir string) (*Project, error) {
	root, err := findRoot(startDir)
	if err != nil {
		return nil, err
	}

	slug := Slug(root)

	dbPath, err := dbPathForSlug(slug)
	if err != nil {
		return nil, err
	}

	return &Project{Root: root, Slug: slug, DBPath: dbPath}, nil
}

// findRoot does the actual upward dir walk looking for git
func findRoot(startDir string) (string, error) {
	dir := startDir
	for {
		_, err := os.Stat(filepath.Join(dir, ".git"))
		if err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not a git repository (or any parent of %s)", startDir)
		}
		dir = parent
	}
}

// dbPathForSlug builds ~/.loom/projects/<slug>/loom.db.
// Ensures that project's subdirectory exists.
func dbPathForSlug(slug string) (string, error) {
	homeDir, err := HomeDir()
	if err != nil {
		return "", err
	}

	projectsDir := filepath.Join(homeDir, "projects", slug)
	if err := os.MkdirAll(projectsDir, 0o700); err != nil {
		return "", fmt.Errorf("creating project dir: %w", err)
	}

	return filepath.Join(projectsDir, "loom.db"), nil
}
