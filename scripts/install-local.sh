#!/usr/bin/env sh
set -eu

BINARY="${BINARY:-oac}"
PREFIX="${PREFIX:-$HOME/.local}"
BINDIR="${BINDIR:-$PREFIX/bin}"
TARGET="$BINDIR/$BINARY"
VERSION="${VERSION:-dev}"
LDFLAGS="${LDFLAGS:-"-s -w -X github.com/carlisle0615/OpenAgentCleaner/internal/cleaner.Version=$VERSION"}"

mkdir -p "$BINDIR"
go build -ldflags "$LDFLAGS" -o "$TARGET" .
chmod 755 "$TARGET"

printf 'Installed %s\n' "$TARGET"
case ":${PATH:-}:" in
  *":$BINDIR:"*) ;;
  *)
    printf 'Add %s to PATH to run `%s` directly.\n' "$BINDIR" "$BINARY"
    ;;
esac
