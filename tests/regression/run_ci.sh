#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CI_DIR="$ROOT_DIR/tests/regression/ci"

if [[ ! -d "$CI_DIR" ]]; then
  exit 0
fi

shopt -s nullglob
scripts=("$CI_DIR"/test_*.sh)
shopt -u nullglob

if [[ ${#scripts[@]} -eq 0 ]]; then
  exit 0
fi

for script in "${scripts[@]}"; do
  bash "$script"
done
