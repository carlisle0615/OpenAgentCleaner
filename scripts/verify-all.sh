#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "$ROOT_DIR"

echo "[verify-all] format"
UNFORMATTED="$(gofmt -l $(find . -name '*.go' -not -path './vendor/*' -print))"
if [[ -n "$UNFORMATTED" ]]; then
  echo "$UNFORMATTED"
  echo "Go files are not formatted. Run: make fmt"
  exit 1
fi

echo "[verify-all] build"
make build

echo "[verify-all] tests"
make test

echo "[verify-all] regression scripts"
bash tests/regression/run_ci.sh

echo "[verify-all] shell scripts"
sh -n install.sh scripts/install-local.sh scripts/generate-homebrew-formula.sh scripts/uninstall.sh scripts/verify-fast.sh scripts/verify-all.sh scripts/install-git-hooks.sh scripts/protect-main.sh
