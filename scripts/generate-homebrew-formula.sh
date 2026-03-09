#!/usr/bin/env sh
set -eu

REPO="${REPO:-carlisle0615/OpenAgentCleaner}"
VERSION="${1:-}"
CHECKSUMS_FILE="${2:-}"
FORMULA_PATH="${FORMULA_PATH:-}"

if [ -z "$VERSION" ] || [ -z "$CHECKSUMS_FILE" ]; then
  printf 'Usage: %s <version> <checksums.txt>\n' "$0" >&2
  exit 1
fi

version="${VERSION#v}"
arm_asset="oac_${version}_darwin_arm64.tar.gz"
amd_asset="oac_${version}_darwin_amd64.tar.gz"
arm_sha="$(grep "  $arm_asset\$" "$CHECKSUMS_FILE" | awk '{print $1}')"
amd_sha="$(grep "  $amd_asset\$" "$CHECKSUMS_FILE" | awk '{print $1}')"

[ -n "$arm_sha" ] || {
  printf 'Missing checksum for %s\n' "$arm_asset" >&2
  exit 1
}
[ -n "$amd_sha" ] || {
  printf 'Missing checksum for %s\n' "$amd_asset" >&2
  exit 1
}

output="${FORMULA_PATH:-/dev/stdout}"

cat >"$output" <<EOF
class Oac < Formula
  desc "Guided macOS cleaner for leftover AI assistant files"
  homepage "https://github.com/$REPO"
  version "$version"

  on_arm do
    url "https://github.com/$REPO/releases/download/v#{version}/$arm_asset"
    sha256 "$arm_sha"
  end

  on_intel do
    url "https://github.com/$REPO/releases/download/v#{version}/$amd_asset"
    sha256 "$amd_sha"
  end

  def install
    bin.install "oac"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/oac version")
  end
end
EOF

printf 'Generated Homebrew formula for %s using %s\n' "$VERSION" "$CHECKSUMS_FILE" >&2
