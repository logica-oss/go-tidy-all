# go-tidy-all

[![GitHub Release](https://img.shields.io/github/v/release/logica-oss/go-tidy-all)](https://github.com/logica-oss/go-tidy-all/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![autofix.ci](https://github.com/logica-oss/go-tidy-all/actions/workflows/autofix.yaml/badge.svg)](https://github.com/logica-oss/go-tidy-all/actions/workflows/autofix.yaml)
[![Verify (App)](https://github.com/logica-oss/go-tidy-all/actions/workflows/verify-app.yaml/badge.svg)](https://github.com/logica-oss/go-tidy-all/actions/workflows/verify-app.yaml)
[![Verify (Actions)](https://github.com/logica-oss/go-tidy-all/actions/workflows/verify-actions.yaml/badge.svg)](https://github.com/logica-oss/go-tidy-all/actions/workflows/verify-actions.yaml)
[![CodeQL Advanced](https://github.com/logica-oss/go-tidy-all/actions/workflows/codeql.yaml/badge.svg)](https://github.com/logica-oss/go-tidy-all/actions/workflows/codeql.yaml)
[![Trivy](https://github.com/logica-oss/go-tidy-all/actions/workflows/trivy.yaml/badge.svg)](https://github.com/logica-oss/go-tidy-all/actions/workflows/trivy.yaml)
[![codecov](https://codecov.io/gh/logica-oss/go-tidy-all/graph/badge.svg)](https://codecov.io/gh/logica-oss/go-tidy-all)

Tidy every Go module in your repo with one command.

## What it does

- Finds every `go.mod` under the current directory.
- Skips what `.gitignore` (and `.git`) already ignores.
- Runs `go mod tidy` in each module.
- Sets each `go` directive to the lowest version your code actually needs (via `mingo`).

## Install

```bash
go install github.com/logica-oss/go-tidy-all/cmd/go-tidy-all@latest
```

Or grab a binary from [Releases](https://github.com/logica-oss/go-tidy-all/releases) (`linux`/`darwin` × `amd64`/`arm64`):

```bash
VERSION=vX.Y.Z # pick from Releases
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/')

curl -fsSL "https://github.com/logica-oss/go-tidy-all/releases/download/${VERSION}/go-tidy-all_${VERSION#v}_${OS}_${ARCH}.tar.gz" | tar -xz go-tidy-all
sudo install -m 755 go-tidy-all /usr/local/bin/
```

## Usage

```plaintext
Usage: go-tidy-all [options]

Run go mod tidy for every module and pin each go directive.

Options:
  -n, --dry-run                 print planned changes without writing
      --min-go-version string   pin the go directive to at least this version (e.g. 1.24.0)
      --root string             repository root (default: current directory)
      --skip-dirs string        comma-separated root-relative directories to skip
  -v, --version                 print version and exit
```

It exits non-zero when no modules are found.

```bash
# Tidy a specific repo, skip fixtures, keep go 1.24+:
go-tidy-all --root /path/to/repo --skip-dirs e2e/testdata --min-go-version 1.24.0

# Preview only:
go-tidy-all -n
```
