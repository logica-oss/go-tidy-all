package tidyall_test

import (
	"context"
	"log/slog"
	"path/filepath"
	"testing"

	tidyall "github.com/logica-oss/go-tidy-all"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func TestRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		files       map[string]string
		skipDirs    []string
		wantErr     bool
		wantVersion string
	}{
		{
			name:    "no modules",
			files:   map[string]string{"main.go": "package main\n"},
			wantErr: true,
		},
		{
			name: "dry run changes nothing",
			files: map[string]string{
				"go.mod":  "module example.com/m\n\ngo 1.24.0\n",
				"main.go": "package main\n\nfunc main() {}\n",
			},
			wantVersion: "1.24.0",
		},
		{
			name: "skips dirs",
			files: map[string]string{
				"go.mod":         "module example.com/root\n\ngo 1.24.0\n",
				"main.go":        "package main\n\nfunc main() {}\n",
				"skipme/go.mod":  "module example.com/skip\n\ngo 1.24.0\n",
				"skipme/main.go": "package main\n\nfunc main() {}\n",
			},
			skipDirs:    []string{"skipme"},
			wantVersion: "1.24.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			writeFiles(t, root, tt.files)

			cfg := tidyall.Config{Root: root, DryRun: true, SkipDirs: tt.skipDirs}
			err := tidyall.Run(context.Background(), newTestLogger(), cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatal("Run = nil, want error")
				}

				return
			}
			if err != nil {
				t.Fatal(err)
			}

			if tt.wantVersion != "" {
				got, err := tidyall.ReadGoVersion(filepath.Join(root, "go.mod"))
				if err != nil {
					t.Fatal(err)
				}
				if got != tt.wantVersion {
					t.Errorf("go directive = %q, want unchanged %q", got, tt.wantVersion)
				}
			}
		})
	}
}

func TestTidyDryRun(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/m\n\ngo 1.24.0\n")

	if err := tidyall.Tidy(context.Background(), newTestLogger(), root, true); err != nil {
		t.Fatal(err)
	}
}

func TestPinGoDirective(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		dryRun     bool
		minVersion string
		want       string
	}{
		{name: "dry run changes nothing", dryRun: true, minVersion: "", want: "1.21.0"},
		{name: "min version wins", dryRun: false, minVersion: "1.99.0", want: "1.99.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			writeFiles(t, dir, map[string]string{
				"go.mod":  "module example.com/m\n\ngo 1.21.0\n",
				"main.go": "package main\n\nfunc main() {}\n",
			})

			if err := tidyall.PinGoDirective(newTestLogger(), dir, tt.dryRun, tt.minVersion); err != nil {
				t.Fatal(err)
			}

			got, err := tidyall.ReadGoVersion(filepath.Join(dir, "go.mod"))
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("go directive = %q, want %q", got, tt.want)
			}
		})
	}
}
