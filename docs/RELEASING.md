# Releasing

This project ships macOS release archives through GitHub Releases and publishes a matching Homebrew formula through a separate tap repository.

## Tag a Release

1. Merge the release-ready branch into `main`.
2. Create and push a new semantic version tag.

```bash
git checkout main
git pull --ff-only
git tag v0.1.2
git push origin v0.1.2
```

The release workflow will build:

- `oac_<version>_darwin_arm64.tar.gz`
- `oac_<version>_darwin_amd64.tar.gz`
- `checksums.txt`

The workflow definition lives in `.github/workflows/release.yml` and runs on `ubuntu-latest`. Because `.goreleaser.yaml` executes `go test ./...` in a `before.hooks` step, all release-gating tests must be portable to Linux even if the shipped binaries are macOS-only.

## One-Line Install

Users can install the latest release with:

```bash
curl -fsSL https://raw.githubusercontent.com/carlisle0615/OpenAgentCleaner/main/install.sh | bash
```

Install a specific version:

```bash
curl -fsSL https://raw.githubusercontent.com/carlisle0615/OpenAgentCleaner/main/install.sh | VERSION=v0.1.0 bash
```

## Homebrew Tap

Homebrew direct installation is designed around a separate tap repository. Homebrew recommends direct installs such as:

```bash
brew install carlisle0615/openagentcleaner/oac
```

That command expects a tap repository named `homebrew-openagentcleaner` under the same GitHub account, with `Formula/oac.rb` inside it.

Generate the formula after a release:

```bash
gh release download v0.1.2 -p checksums.txt -D dist/release-artifacts
./scripts/generate-homebrew-formula.sh v0.1.2 dist/release-artifacts/checksums.txt > Formula/oac.rb
```

Then commit `Formula/oac.rb` into the tap repository `carlisle0615/homebrew-openagentcleaner`.

## Notes

- The current installer supports macOS only.
- The Homebrew formula should be updated only after the GitHub Release assets are visible and checksums are final.
- `oac version` reads the release version injected at build time.
