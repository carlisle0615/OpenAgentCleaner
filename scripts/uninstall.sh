#!/usr/bin/env sh
set -eu

BINARY="${BINARY:-oac}"
PREFIX="${PREFIX:-$HOME/.local}"
BINDIR="${BINDIR:-$PREFIX/bin}"
TARGET="$BINDIR/$BINARY"

if [ -e "$TARGET" ]; then
  rm -f "$TARGET"
  printf 'Removed %s\n' "$TARGET"
  exit 0
fi

printf 'Nothing to remove at %s\n' "$TARGET"
