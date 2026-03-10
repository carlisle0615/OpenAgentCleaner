#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
cd "$ROOT_DIR"

ASSISTANTS="${OAC_MANUAL_ASSISTANTS:-codex,codex-cli,claudecode,cursor,antigravity,openclaw}"

echo "[manual] analyze latency assistants: $ASSISTANTS"
echo "[manual] this uses real local session stores and does not modify data"

OAC_MANUAL_ANALYZE_LATENCY=1 \
OAC_MANUAL_ASSISTANTS="$ASSISTANTS" \
go test ./internal/cleaner -run TestManualAnalyzeLatency -count=1 -v
