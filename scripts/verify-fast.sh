#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "$ROOT_DIR"

echo "[verify-fast] format"
UNFORMATTED="$(gofmt -l main.go internal/cleaner/*.go)"
if [[ -n "$UNFORMATTED" ]]; then
  echo "$UNFORMATTED"
  echo "Go files are not formatted. Run: make fmt"
  exit 1
fi

echo "[verify-fast] unit tests"
go test ./internal/cleaner/...
