#!/usr/bin/env sh
set -eu

BINARY="${BINARY:-oac}"
PREFIX="${PREFIX:-$HOME/.local}"
BINDIR="${BINDIR:-$PREFIX/bin}"
TARGET="$BINDIR/$BINARY"

mkdir -p "$BINDIR"
go build -o "$TARGET" .
chmod 755 "$TARGET"

printf 'Installed %s\n' "$TARGET"
case ":${PATH:-}:" in
  *":$BINDIR:"*) ;;
  *)
    printf 'Add %s to PATH to run `%s` directly.\n' "$BINDIR" "$BINARY"
    ;;
esac
