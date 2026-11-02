package tidyall_test

import (
	"path/filepath"
	"slices"
	"testing"

	tidyall "github.com/logica-oss/go-tidy-all"
)

func TestFindModules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		files    map[string]string
		skipDirs []string
		want     []string
	}{
		{
			name: "nested modules",
			files: map[string]string{
				"go.mod":             "module example.com/root\n\ngo 1.24.0\n",
				"sub/go.mod":         "module example.com/sub\n\ngo 1.24.0\n",
				"sub/nested/go.mod":  "module example.com/nested\n\ngo 1.24.0\n",
				"notamodule/main.go": "package main\n",
			},
			want: []string{".", "sub", "sub/nested"},
		},
		{
			name:  "empty",
			files: map[string]string{"main.go": "package main\n"},
			want:  nil,
		},
		{
			name: "skip dirs",
			files: map[string]string{
				"go.mod":        "module example.com/root\n\ngo 1.24.0\n",
				"skipme/go.mod": "module example.com/skip\n\ngo 1.24.0\n",
				"keep/go.mod":   "module example.com/keep\n\ngo 1.24.0\n",
			},
			skipDirs: []string{"skipme"},
			want:     []string{".", "keep"},
		},
		{
			name: "skips git",
			files: map[string]string{
				"go.mod":            "module example.com/root\n\ngo 1.24.0\n",
				".git/hooks/go.mod": "module example.com/hooks\n\ngo 1.24.0\n",
			},
			want: []string{"."},
		},
		{
			name: "skips ignored dir",
			files: map[string]string{
				".gitignore":     "ignored/\n",
				"go.mod":         "module example.com/root\n\ngo 1.24.0\n",
				"ignored/go.mod": "module example.com/ignored\n\ngo 1.24.0\n",
			},
			want: []string{"."},
		},
		{
			name: "skips ignored file",
			files: map[string]string{
				".gitignore":  "sub/go.mod\n",
				"go.mod":      "module example.com/root\n\ngo 1.24.0\n",
				"sub/go.mod":  "module example.com/sub\n\ngo 1.24.0\n",
				"keep/go.mod": "module example.com/keep\n\ngo 1.24.0\n",
			},
			want: []string{".", "keep"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			writeFiles(t, root, tt.files)

			got, err := tidyall.FindModules(root, tt.skipDirs)
			if err != nil {
				t.Fatal(err)
			}

			resolved := resolveRoot(t, root)
			want := make([]string, 0, len(tt.want))
			for _, rel := range tt.want {
				if rel == "." {
					want = append(want, resolved)
				} else {
					want = append(want, filepath.Join(resolved, filepath.FromSlash(rel)))
				}
			}
			slices.Sort(got)
			slices.Sort(want)
			if !slices.Equal(got, want) {
				t.Errorf("tidyall.FindModules = %v, want %v", got, want)
			}
		})
	}
}

func TestDirFilterSkip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		path  string
		isDir bool
		want  bool
	}{
		{name: "skip dir", path: "skipme", isDir: true, want: true},
		{name: "keep dir", path: "keep", isDir: true, want: false},
		{name: "git dir", path: ".git", isDir: true, want: true},
		{name: "root", path: ".", isDir: true, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			filter, err := tidyall.NewDirFilter(root, []string{"skipme"})
			if err != nil {
				t.Fatal(err)
			}

			path := root
			if tt.path != "." {
				path = filepath.Join(root, tt.path)
			}
			if got := filter.Skip(path, tt.isDir); got != tt.want {
				t.Errorf("filter.Skip(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
