package tidyall

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestFindModules(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/root\n\ngo 1.24.0\n")
	writeFile(t, filepath.Join(root, "sub", "go.mod"), "module example.com/sub\n\ngo 1.24.0\n")
	writeFile(t, filepath.Join(root, "sub", "nested", "go.mod"), "module example.com/nested\n\ngo 1.24.0\n")
	writeFile(t, filepath.Join(root, "notamodule", "main.go"), "package main\n")

	got, err := findModules(root, nil)
	if err != nil {
		t.Fatal(err)
	}

	resolved := resolveRoot(t, root)
	want := []string{resolved, filepath.Join(resolved, "sub"), filepath.Join(resolved, "sub", "nested")}
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("findModules = %v, want %v", got, want)
	}
}

func TestFindModulesEmpty(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "main.go"), "package main\n")

	got, err := findModules(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("findModules = %v, want empty", got)
	}
}

func TestFindModulesSkipDirs(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/root\n\ngo 1.24.0\n")
	writeFile(t, filepath.Join(root, "skipme", "go.mod"), "module example.com/skip\n\ngo 1.24.0\n")
	writeFile(t, filepath.Join(root, "keep", "go.mod"), "module example.com/keep\n\ngo 1.24.0\n")

	got, err := findModules(root, []string{"skipme"})
	if err != nil {
		t.Fatal(err)
	}

	resolved := resolveRoot(t, root)
	want := []string{resolved, filepath.Join(resolved, "keep")}
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("findModules = %v, want %v", got, want)
	}
}

func TestFindModulesSkipsGit(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/root\n\ngo 1.24.0\n")
	writeFile(t, filepath.Join(root, ".git", "hooks", "go.mod"), "module example.com/hooks\n\ngo 1.24.0\n")

	got, err := findModules(root, nil)
	if err != nil {
		t.Fatal(err)
	}

	resolved := resolveRoot(t, root)
	if len(got) != 1 || got[0] != resolved {
		t.Errorf("findModules = %v, want [%s]", got, resolved)
	}
}

func TestFindModulesSkipsIgnoredGoMod(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".gitignore"), "ignored/\n")
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/root\n\ngo 1.24.0\n")
	writeFile(t, filepath.Join(root, "ignored", "go.mod"), "module example.com/ignored\n\ngo 1.24.0\n")

	got, err := findModules(root, nil)
	if err != nil {
		t.Fatal(err)
	}

	resolved := resolveRoot(t, root)
	if len(got) != 1 || got[0] != resolved {
		t.Errorf("findModules = %v, want [%s]", got, resolved)
	}
}

func TestFindModulesSkipsIgnoredFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".gitignore"), "sub/go.mod\n")
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/root\n\ngo 1.24.0\n")
	writeFile(t, filepath.Join(root, "sub", "go.mod"), "module example.com/sub\n\ngo 1.24.0\n")
	writeFile(t, filepath.Join(root, "keep", "go.mod"), "module example.com/keep\n\ngo 1.24.0\n")

	got, err := findModules(root, nil)
	if err != nil {
		t.Fatal(err)
	}

	resolved := resolveRoot(t, root)
	want := []string{resolved, filepath.Join(resolved, "keep")}
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("findModules = %v, want %v", got, want)
	}
}

func resolveRoot(t *testing.T, root string) string {
	t.Helper()

	resolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}

	return resolved
}

func TestDirFilterSkip(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	filter, err := newDirFilter(root, []string{"skipme"})
	if err != nil {
		t.Fatal(err)
	}

	if !filter.skip(filepath.Join(root, "skipme"), "skipme") {
		t.Error("skip(skipme) = false, want true")
	}
	if filter.skip(filepath.Join(root, "keep"), "keep") {
		t.Error("skip(keep) = true, want false")
	}
	if !filter.skip(filepath.Join(root, ".git"), ".git") {
		t.Error("skip(.git) = false, want true")
	}
	if filter.skip(root, "go-tidy-all") {
		t.Error("skip(root) = true, want false")
	}
}
