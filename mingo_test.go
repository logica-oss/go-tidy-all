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

func TestReadGoVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		files   map[string]string
		want    string
		wantErr bool
	}{
		{
			name:  "directive",
			files: map[string]string{"go.mod": "module example.com/m\n\ngo 1.24.0\n"},
			want:  "1.24.0",
		},
		{
			name:  "no directive",
			files: map[string]string{"go.mod": "module example.com/m\n"},
			want:  "",
		},
		{
			name:    "missing file",
			files:   map[string]string{},
			wantErr: true,
		},
		{
			name:    "invalid file",
			files:   map[string]string{"go.mod": "not a go.mod {{{"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			writeFiles(t, dir, tt.files)

			got, err := tidyall.ReadGoVersion(filepath.Join(dir, "go.mod"))
			if tt.wantErr {
				if err == nil {
					t.Error("tidyall.ReadGoVersion = nil, want error")
				}

				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("tidyall.ReadGoVersion = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWriteGoVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		version string
		want    string
	}{
		{
			name:    "update directive",
			content: "module example.com/m\n\ngo 1.23.0\n",
			version: "1.24.0",
			want:    "1.24.0",
		},
		{
			name:    "add directive",
			content: "module example.com/m\n",
			version: "1.24.0",
			want:    "1.24.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			path := filepath.Join(dir, "go.mod")
			writeFile(t, path, tt.content)

			if err := tidyall.WriteGoVersion(path, tt.version); err != nil {
				t.Fatal(err)
			}

			got, err := tidyall.ReadGoVersion(path)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("tidyall.ReadGoVersion after write = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseModFileInvalid(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "go.mod")
	writeFile(t, path, "not a go.mod {{{")

	if _, err := tidyall.ParseModFile(path); err == nil {
		t.Error("tidyall.ParseModFile = nil, want error")
	}
}
