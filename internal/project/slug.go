package project

import (
	"strings"
)

// Slug converts absolute path into a filesystem-safe, deterministic identifier.
// e.g. /Users/bob/Projects/hello-world -> Users-bob-Projects-hello-world
func Slug(absPath string) string {
	normalized := strings.ReplaceAll(absPath, `\`, "/")
	normalized = strings.TrimLeft(normalized, "/")
	normalized = strings.ReplaceAll(normalized, ":", "")
	return strings.ReplaceAll(normalized, "/", "-")
}
