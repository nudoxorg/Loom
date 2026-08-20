package project

import (
	"fmt"
	"path/filepath"
	"strings"
)

// NormalizePath resolves path (as passed by a caller — relative to
// whatever directory that process happened to be in, or absolute) to a
// canonical form relative to root. Without this, "foo.go", "./foo.go",
// and "/abs/repo/foo.go" would each be tracked as a different claim,
// and releasing one wouldn't touch the others.
//
// A path outside root (rare, but possible if a caller passes something
// odd) is kept as a cleaned absolute path rather than an error, since
// claims are advisory and shouldn't fail a claim over it.
func NormalizePath(root, path string) (string, error) {
	abs := path
	if !filepath.IsAbs(abs) {
		var err error
		abs, err = filepath.Abs(abs)
		if err != nil {
			return "", fmt.Errorf("resolving path %q: %w", path, err)
		}
	}
	abs = filepath.Clean(abs)

	rel, err := filepath.Rel(root, abs)
	if err != nil || strings.HasPrefix(rel, "..") {
		return abs, nil
	}

	return filepath.ToSlash(rel), nil
}
