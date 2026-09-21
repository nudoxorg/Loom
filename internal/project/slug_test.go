package project

import "testing"

func TestSlug(t *testing.T) {
	tests := map[string]string{
		"/Users/alice/code/loom":   "Users-alice-code-loom",
		`C:\Users\alice\code\loom`: "C-Users-alice-code-loom",
	}

	for path, want := range tests {
		if got := Slug(path); got != want {
			t.Errorf("Slug(%q) = %q, want %q", path, got, want)
		}
	}
}
