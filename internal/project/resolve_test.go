package project

import (
	"os"
	"path/filepath"
	"testing"
)

func withResolveTempHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	if err := EnsureHome(); err != nil {
		t.Fatalf("EnsureHome() error = %v", err)
	}
}

func TestResolveNoGitAnywhereUsesStartDirAsAdHocRoot(t *testing.T) {
	withResolveTempHome(t)

	scratch := t.TempDir()

	proj, err := Resolve(scratch)
	if err != nil {
		t.Fatalf("Resolve() error = %v, want ad-hoc root instead of error", err)
	}

	want, err := filepath.EvalSymlinks(scratch)
	if err != nil {
		t.Fatalf("EvalSymlinks(%s) error = %v", scratch, err)
	}
	got, err := filepath.EvalSymlinks(proj.Root)
	if err != nil {
		t.Fatalf("EvalSymlinks(%s) error = %v", proj.Root, err)
	}
	if got != want {
		t.Fatalf("proj.Root = %q, want %q", got, want)
	}
}

func TestResolveFindsGitInAncestor(t *testing.T) {
	withResolveTempHome(t)

	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o700); err != nil {
		t.Fatalf("Mkdir(.git) error = %v", err)
	}

	nested := filepath.Join(root, "src", "pkg")
	if err := os.MkdirAll(nested, 0o700); err != nil {
		t.Fatalf("MkdirAll(nested) error = %v", err)
	}

	proj, err := Resolve(nested)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	wantRoot, _ := filepath.EvalSymlinks(root)
	gotRoot, _ := filepath.EvalSymlinks(proj.Root)
	if gotRoot != wantRoot {
		t.Fatalf("proj.Root = %q, want %q", gotRoot, wantRoot)
	}
}

func TestResolveNestedGitRepoUsesNearestAncestor(t *testing.T) {
	withResolveTempHome(t)

	outer := t.TempDir()
	if err := os.Mkdir(filepath.Join(outer, ".git"), 0o700); err != nil {
		t.Fatalf("Mkdir(outer .git) error = %v", err)
	}

	inner := filepath.Join(outer, "vendor", "nested-repo")
	if err := os.MkdirAll(inner, 0o700); err != nil {
		t.Fatalf("MkdirAll(inner) error = %v", err)
	}
	if err := os.Mkdir(filepath.Join(inner, ".git"), 0o700); err != nil {
		t.Fatalf("Mkdir(inner .git) error = %v", err)
	}

	innerSub := filepath.Join(inner, "src")
	if err := os.MkdirAll(innerSub, 0o700); err != nil {
		t.Fatalf("MkdirAll(innerSub) error = %v", err)
	}

	proj, err := Resolve(innerSub)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	wantRoot, _ := filepath.EvalSymlinks(inner)
	gotRoot, _ := filepath.EvalSymlinks(proj.Root)
	if gotRoot != wantRoot {
		t.Fatalf("proj.Root = %q, want nearest ancestor %q (not outer repo)", gotRoot, wantRoot)
	}
}
