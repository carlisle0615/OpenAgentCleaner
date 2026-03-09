# Releasing

This project ships macOS release archives through GitHub Releases.

## Tag a Release

```bash
git tag v0.1.0
git push origin v0.1.0
```

The release workflow will build:

- `oac_<version>_darwin_arm64.tar.gz`
- `oac_<version>_darwin_amd64.tar.gz`
- `checksums.txt`

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
./scripts/generate-homebrew-formula.sh v0.1.0 dist/checksums.txt > Formula/oac.rb
```

Then commit `Formula/oac.rb` into the tap repository.

## Notes

- The current installer supports macOS only.
- The current Homebrew packaging plan targets a separate tap repo, not this source repository.
- `oac version` reads the release version injected at build time.
