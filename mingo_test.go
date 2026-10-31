package tidyall_test

import (
	"path/filepath"
	"testing"

	tidyall "github.com/logica-oss/go-tidy-all"
)

func TestCompareGoVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{name: "equal", a: "1.24.0", b: "1.24.0", want: 0},
		{name: "equal with go prefix", a: "go1.24.0", b: "1.24.0", want: 0},
		{name: "minor greater", a: "1.25.0", b: "1.24.0", want: 1},
		{name: "minor smaller", a: "1.23.0", b: "1.24.0", want: -1},
		{name: "patch greater", a: "1.24.1", b: "1.24.0", want: 1},
		{name: "patch smaller", a: "1.24.0", b: "1.24.1", want: -1},
		{name: "major greater", a: "2.0.0", b: "1.99.99", want: 1},
		{name: "missing patch treated as zero", a: "1.24", b: "1.24.0", want: 0},
		{name: "missing minor treated as zero", a: "1", b: "1.0.0", want: 0},
		{name: "non-numeric treated as zero", a: "1.x.0", b: "1.0.0", want: 0},
		{name: "empty treated as zero", a: "", b: "0.0.0", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tidyall.CompareGoVersion(tt.a, tt.b); got != tt.want {
				t.Errorf("tidyall.CompareGoVersion(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestGoVersionPart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		parts []string
		index int
		want  int
	}{
		{name: "first", parts: []string{"1", "24", "0"}, index: 0, want: 1},
		{name: "second", parts: []string{"1", "24", "0"}, index: 1, want: 24},
		{name: "out of range", parts: []string{"1"}, index: 5, want: 0},
		{name: "non-numeric", parts: []string{"x"}, index: 0, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tidyall.GoVersionPart(tt.parts, tt.index); got != tt.want {
				t.Errorf("tidyall.GoVersionPart(%v, %d) = %d, want %d", tt.parts, tt.index, got, tt.want)
			}
		})
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := mkParent(path); err != nil {
		t.Fatal(err)
	}
	if err := writeTestFile(path, content); err != nil {
		t.Fatal(err)
	}
}

func TestReadGoVersion(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "go.mod")
	writeFile(t, path, "module example.com/m\n\ngo 1.24.0\n")

	got, err := tidyall.ReadGoVersion(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != "1.24.0" {
		t.Errorf("readGoVersion = %q, want %q", got, "1.24.0")
	}
}

func TestReadGoVersionNoDirective(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "go.mod")
	writeFile(t, path, "module example.com/m\n")

	got, err := tidyall.ReadGoVersion(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("readGoVersion = %q, want empty", got)
	}
}

func TestReadGoVersionMissing(t *testing.T) {
	t.Parallel()

	if _, err := tidyall.ReadGoVersion(filepath.Join(t.TempDir(), "go.mod")); err == nil {
		t.Error("readGoVersion = nil, want error")
	}
}

func TestWriteGoVersion(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "go.mod")
	writeFile(t, path, "module example.com/m\n\ngo 1.23.0\n")

	if err := tidyall.WriteGoVersion(path, "1.24.0"); err != nil {
		t.Fatal(err)
	}

	got, err := tidyall.ReadGoVersion(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != "1.24.0" {
		t.Errorf("readGoVersion after write = %q, want %q", got, "1.24.0")
	}
}

func TestWriteGoVersionAddsDirective(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "go.mod")
	writeFile(t, path, "module example.com/m\n")

	if err := tidyall.WriteGoVersion(path, "1.24.0"); err != nil {
		t.Fatal(err)
	}

	got, err := tidyall.ReadGoVersion(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != "1.24.0" {
		t.Errorf("readGoVersion after write = %q, want %q", got, "1.24.0")
	}
}

func TestParseModFileInvalid(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "go.mod")
	writeFile(t, path, "not a go.mod {{{")

	if _, err := tidyall.ParseModFile(path); err == nil {
		t.Error("parseModFile = nil, want error")
	}
}
