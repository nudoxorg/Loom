package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// ListAll enumerates every project Loom has touched, based on the
// ~/.loom/projects/<slug> directories created by dbPathForSlug. Root
// comes from each project's "root" sidecar file when present, falling
// back to the raw slug for project directories that predate that file.
func ListAll() ([]Project, error) {
	homeDir, err := HomeDir()
	if err != nil {
		return nil, err
	}

	projectsDir := filepath.Join(homeDir, "projects")
	entries, err := os.ReadDir(projectsDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}

	var projects []Project
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		slug := entry.Name()
		dir := filepath.Join(projectsDir, slug)

		root := slug
		if data, err := os.ReadFile(filepath.Join(dir, "root")); err == nil {
			root = string(data)
		}

		projects = append(projects, Project{
			Root:   root,
			Slug:   slug,
			DBPath: filepath.Join(dir, "loom.db"),
		})
	}

	return projects, nil
}
