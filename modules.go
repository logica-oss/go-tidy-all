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

func findModules(root string, skipDirs []string) ([]string, error) {
	filter, err := newDirFilter(root, skipDirs)
	if err != nil {
		return nil, err
	}

	var dirs []string
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if filter.skip(path, entry.Name()) {
				return filepath.SkipDir
			}

			return nil
		}
		if entry.Name() != "go.mod" {
			return nil
		}

		if filter.skipFile(path) {
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

type dirFilter struct {
	root     string
	matcher  gitignore.Matcher
	skipDirs map[string]struct{}
}

func newDirFilter(root string, skipDirs []string) (dirFilter, error) {
	patterns, err := gitignore.ReadPatterns(osfs.New(root), nil)
	if err != nil {
		return dirFilter{}, errors.WithStack(fmt.Errorf("read gitignore patterns: %w", err))
	}

	skipDirsMap := make(map[string]struct{}, len(skipDirs))
	for _, dir := range skipDirs {
		rel, err := filepath.Rel(root, filepath.Join(root, filepath.FromSlash(dir)))
		if err != nil {
			continue
		}

		skipDirsMap[rel] = struct{}{}
	}

	return dirFilter{
		root:     root,
		matcher:  gitignore.NewMatcher(patterns),
		skipDirs: skipDirsMap,
	}, nil
}

func (f dirFilter) skip(path, name string) bool {
	if name == ".git" {
		return true
	}

	rel, err := filepath.Rel(f.root, path)
	if err != nil || rel == "." {
		return false
	}

	if _, ok := f.skipDirs[rel]; ok {
		return true
	}

	if f.matcher == nil {
		return false
	}

	return f.matcher.Match(strings.Split(rel, string(filepath.Separator)), true)
}

func (f dirFilter) skipFile(path string) bool {
	if f.matcher == nil {
		return false
	}

	rel, err := filepath.Rel(f.root, path)
	if err != nil {
		return false
	}

	return f.matcher.Match(strings.Split(rel, string(filepath.Separator)), false)
}
