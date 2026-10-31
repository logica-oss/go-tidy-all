package tidyall

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/k1LoW/errors"
)

func findModules(root string) ([]string, error) {
	patterns, err := loadGitignore(root)
	if err != nil {
		return nil, err
	}
	matcher := gitignore.NewMatcher(patterns)

	var dirs []string
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if shouldSkipDir(matcher, root, path, entry.Name()) {
				return filepath.SkipDir
			}

			return nil
		}
		if entry.Name() != "go.mod" {
			return nil
		}

		dirs = append(dirs, filepath.Dir(path))

		return nil
	})
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("walk modules: %w", err))
	}

	return dirs, nil
}

func loadGitignore(root string) ([]gitignore.Pattern, error) {
	patterns, err := gitignore.ReadPatterns(osfs.New(root), nil)
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("read gitignore patterns: %w", err))
	}

	return patterns, nil
}

func shouldSkipDir(matcher gitignore.Matcher, root, path, name string) bool {
	if name == ".git" {
		return true
	}

	if matcher == nil {
		return false
	}

	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." {
		return false
	}

	return matcher.Match(strings.Split(rel, string(filepath.Separator)), true)
}
