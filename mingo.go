package tidyall

import (
	"cmp"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bobg/mingo"
	"github.com/k1LoW/errors"
	"golang.org/x/mod/modfile"
)

func pinGoDirective(log *slog.Logger, dir string, dryRun bool, minGoVersion string) error {
	log.Info("mingo", "dir", dir)

	scanner := mingo.Scanner{
		Deps:     true,
		Indirect: true,
		Tests:    true,
	}
	result, err := scanner.ScanDir(dir)
	if err != nil {
		return errors.WithStack(fmt.Errorf("mingo: %w", err))
	}

	minimum := "1." + strconv.Itoa(result.Version()) + ".0"
	if minGoVersion != "" && compareGoVersion(minimum, minGoVersion) < 0 {
		minimum = minGoVersion
	}

	goMod := filepath.Join(dir, "go.mod")
	declared, err := readGoVersion(goMod)
	if err != nil {
		return err
	}

	if declared == minimum {
		log.Info("go directive is already minimal", "version", declared)

		return nil
	}

	if dryRun {
		log.Info("dry-run: would pin go directive", "from", declared, "to", minimum)

		return nil
	}

	log.Info("pin go directive", "from", declared, "to", minimum)

	return writeGoVersion(goMod, minimum)
}

const maxParts = 3

func compareGoVersion(a, b string) int {
	aParts := strings.SplitN(strings.TrimPrefix(a, "go"), ".", maxParts)
	bParts := strings.SplitN(strings.TrimPrefix(b, "go"), ".", maxParts)

	for i := range maxParts {
		aNum := goVersionPart(aParts, i)
		bNum := goVersionPart(bParts, i)
		if aNum != bNum {
			return cmp.Compare(aNum, bNum)
		}
	}

	return 0
}

func goVersionPart(parts []string, i int) int {
	if i >= len(parts) {
		return 0
	}

	num, err := strconv.Atoi(parts[i])
	if err != nil {
		return 0
	}

	return num
}

func readGoVersion(path string) (string, error) {
	file, err := parseModFile(path)
	if err != nil {
		return "", err
	}
	if file.Go == nil {
		return "", nil
	}

	return file.Go.Version, nil
}

func writeGoVersion(goMod, version string) error {
	file, err := parseModFile(goMod)
	if err != nil {
		return err
	}

	if file.Go != nil {
		file.DropGoStmt()
	}
	if addErr := file.AddGoStmt(version); addErr != nil {
		return errors.WithStack(fmt.Errorf("set go %s: %w", version, addErr))
	}

	file.Cleanup()

	data := modfile.Format(file.Syntax)
	info, err := os.Stat(goMod)
	if err != nil {
		return errors.WithStack(fmt.Errorf("stat go.mod: %w", err))
	}

	tmp, err := os.CreateTemp(filepath.Dir(goMod), "go.mod.*")
	if err != nil {
		return errors.WithStack(fmt.Errorf("create temp go.mod: %w", err))
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)

		return errors.WithStack(fmt.Errorf("write temp go.mod: %w", err))
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)

		return errors.WithStack(fmt.Errorf("close temp go.mod: %w", err))
	}
	if err := os.Chmod(tmpName, info.Mode().Perm()); err != nil {
		_ = os.Remove(tmpName)

		return errors.WithStack(fmt.Errorf("chmod temp go.mod: %w", err))
	}
	if err := os.Rename(tmpName, goMod); err != nil {
		_ = os.Remove(tmpName)

		return errors.WithStack(fmt.Errorf("replace go.mod: %w", err))
	}

	return nil
}

func parseModFile(path string) (*modfile.File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("read go.mod: %w", err))
	}

	file, err := modfile.Parse(path, data, nil)
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("parse go.mod: %w", err))
	}

	return file, nil
}
