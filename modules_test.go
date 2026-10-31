package tidyall_test

import (
	"path/filepath"
	"slices"
	"testing"

	tidyall "github.com/logica-oss/go-tidy-all"
)

func TestFindModules(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/root\n\ngo 1.24.0\n")
	writeFile(t, filepath.Join(root, "sub", "go.mod"), "module example.com/sub\n\ngo 1.24.0\n")
	writeFile(t, filepath.Join(root, "sub", "nested", "go.mod"), "module example.com/nested\n\ngo 1.24.0\n")
	writeFile(t, filepath.Join(root, "notamodule", "main.go"), "package main\n")

	got, err := tidyall.FindModules(root, nil)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{root, filepath.Join(root, "sub"), filepath.Join(root, "sub", "nested")}
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

	got, err := tidyall.FindModules(root, nil)
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

	got, err := tidyall.FindModules(root, []string{"skipme"})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{root, filepath.Join(root, "keep")}
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

	got, err := tidyall.FindModules(root, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 1 || got[0] != root {
		t.Errorf("findModules = %v, want [%s]", got, root)
	}
}

func TestFindModulesSkipsIgnoredGoMod(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".gitignore"), "ignored/\n")
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/root\n\ngo 1.24.0\n")
	writeFile(t, filepath.Join(root, "ignored", "go.mod"), "module example.com/ignored\n\ngo 1.24.0\n")

	got, err := tidyall.FindModules(root, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 1 || got[0] != root {
		t.Errorf("findModules = %v, want [%s]", got, root)
	}
}

func TestFindModulesSkipsIgnoredFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".gitignore"), "sub/go.mod\n")
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/root\n\ngo 1.24.0\n")
	writeFile(t, filepath.Join(root, "sub", "go.mod"), "module example.com/sub\n\ngo 1.24.0\n")
	writeFile(t, filepath.Join(root, "keep", "go.mod"), "module example.com/keep\n\ngo 1.24.0\n")

	got, err := tidyall.FindModules(root, nil)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{root, filepath.Join(root, "keep")}
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("findModules = %v, want %v", got, want)
	}
}

func TestDirFilterSkip(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	filter, err := tidyall.NewDirFilter(root, []string{"skipme"})
	if err != nil {
		t.Fatal(err)
	}

	if !filter.Skip(filepath.Join(root, "skipme"), "skipme") {
		t.Error("skip(skipme) = false, want true")
	}
	if filter.Skip(filepath.Join(root, "keep"), "keep") {
		t.Error("skip(keep) = true, want false")
	}
	if !filter.Skip(filepath.Join(root, ".git"), ".git") {
		t.Error("skip(.git) = false, want true")
	}
	if filter.Skip(root, "go-tidy-all") {
		t.Error("skip(root) = true, want false")
	}
}
