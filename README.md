# go-tidy-all

[![GitHub Release](https://img.shields.io/github/v/release/logica-oss/go-tidy-all)](https://github.com/logica-oss/go-tidy-all/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![autofix.ci](https://github.com/logica-oss/go-tidy-all/actions/workflows/autofix.yaml/badge.svg)](https://github.com/logica-oss/go-tidy-all/actions/workflows/autofix.yaml)
[![Verify (App)](https://github.com/logica-oss/go-tidy-all/actions/workflows/verify-app.yaml/badge.svg)](https://github.com/logica-oss/go-tidy-all/actions/workflows/verify-app.yaml)
[![Verify (Actions)](https://github.com/logica-oss/go-tidy-all/actions/workflows/verify-actions.yaml/badge.svg)](https://github.com/logica-oss/go-tidy-all/actions/workflows/verify-actions.yaml)
[![CodeQL Advanced](https://github.com/logica-oss/go-tidy-all/actions/workflows/codeql.yaml/badge.svg)](https://github.com/logica-oss/go-tidy-all/actions/workflows/codeql.yaml)
[![Trivy](https://github.com/logica-oss/go-tidy-all/actions/workflows/trivy.yaml/badge.svg)](https://github.com/logica-oss/go-tidy-all/actions/workflows/trivy.yaml)
[![codecov](https://codecov.io/gh/logica-oss/go-tidy-all/graph/badge.svg)](https://codecov.io/gh/logica-oss/go-tidy-all)

Run `go mod tidy` for every module in the repository and pin each `go` directive to the lowest version that `mingo` computes.

## What it does

- Walks the repository from the current directory and finds every `go.mod`.
- Skips directories ignored by `.gitignore` (plus `.git`).
- Runs `go mod tidy` in each module directory.
- Scans each module with `mingo` and rewrites the `go` directive to the minimal `1.<N>.0`.

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

```bash
go-tidy-all
```

Run it from the repository root. There are no flags: it tidies every module it finds and exits non-zero with no modules found.

## Development

```bash
task tidy
task lint
task build
task test
```
