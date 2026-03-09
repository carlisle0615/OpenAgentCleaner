#!/usr/bin/env sh
set -eu

REPO="${REPO:-carlisle0615/OpenAgentCleaner}"
BINARY="${BINARY:-oac}"
PREFIX="${PREFIX:-$HOME/.local}"
BINDIR="${BINDIR:-$PREFIX/bin}"
VERSION="${VERSION:-latest}"

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    printf 'Missing required command: %s\n' "$1" >&2
    exit 1
  }
}

detect_os() {
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "$os" in
    darwin) printf '%s\n' "$os" ;;
    *)
      printf 'OpenAgentCleaner currently supports macOS only.\n' >&2
      exit 1
      ;;
  esac
}

detect_arch() {
  arch="$(uname -m)"
  case "$arch" in
    arm64|aarch64) printf 'arm64\n' ;;
    x86_64) printf 'amd64\n' ;;
    *)
      printf 'Unsupported CPU architecture: %s\n' "$arch" >&2
      exit 1
      ;;
  esac
}

resolve_version() {
  if [ "$VERSION" != "latest" ]; then
    case "$VERSION" in
      v*) printf '%s\n' "$VERSION" ;;
      *) printf 'v%s\n' "$VERSION" ;;
    esac
    return
  fi

  latest_url="https://api.github.com/repos/$REPO/releases/latest"
  tag="$(
    curl -fsSL "$latest_url" |
      sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' |
      head -n 1
  )"
  if [ -z "$tag" ]; then
    printf 'Failed to resolve the latest release tag from %s\n' "$latest_url" >&2
    exit 1
  fi
  printf '%s\n' "$tag"
}

verify_checksum() {
  file="$1"
  checksums="$2"

  if command -v shasum >/dev/null 2>&1; then
    expected="$(grep "  $(basename "$file")\$" "$checksums" | awk '{print $1}')"
    actual="$(shasum -a 256 "$file" | awk '{print $1}')"
    [ -n "$expected" ] || {
      printf 'Checksum entry not found for %s\n' "$file" >&2
      exit 1
    }
    [ "$expected" = "$actual" ] || {
      printf 'Checksum verification failed for %s\n' "$file" >&2
      exit 1
    }
  fi
}

require_cmd curl
require_cmd tar

os="$(detect_os)"
arch="$(detect_arch)"
tag="$(resolve_version)"
version="${tag#v}"
archive="${BINARY}_${version}_${os}_${arch}.tar.gz"
checksums="checksums.txt"
base_url="https://github.com/$REPO/releases/download/$tag"
archive_url="$base_url/$archive"
checksums_url="$base_url/$checksums"

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT INT TERM

printf 'Installing %s %s for %s/%s\n' "$BINARY" "$tag" "$os" "$arch"
curl -fsSL -o "$tmpdir/$archive" "$archive_url"
curl -fsSL -o "$tmpdir/$checksums" "$checksums_url"
verify_checksum "$tmpdir/$archive" "$tmpdir/$checksums"

mkdir -p "$BINDIR"
tar -xzf "$tmpdir/$archive" -C "$tmpdir"
install_path="$BINDIR/$BINARY"
mv "$tmpdir/$BINARY" "$install_path"
chmod 755 "$install_path"

printf 'Installed %s\n' "$install_path"
case ":${PATH:-}:" in
  *":$BINDIR:"*) ;;
  *)
    printf 'Add %s to PATH to run `%s` directly.\n' "$BINDIR" "$BINARY"
    ;;
esac
