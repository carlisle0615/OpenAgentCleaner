#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
MANUAL_DIR="$ROOT_DIR/tests/regression/manual"

if [[ ! -d "$MANUAL_DIR" ]]; then
  exit 0
fi

shopt -s nullglob
scripts=("$MANUAL_DIR"/test_*.sh)
shopt -u nullglob

if [[ ${#scripts[@]} -eq 0 ]]; then
  exit 0
fi

for script in "${scripts[@]}"; do
  bash "$script"
done
