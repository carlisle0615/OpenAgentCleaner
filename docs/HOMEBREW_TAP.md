# Homebrew Tap Setup

`OpenAgentCleaner` is designed to publish its Homebrew formula through a separate tap repository.

## Recommended Repository

Use this repository under the same GitHub account:

```text
carlisle0615/homebrew-openagentcleaner
```

Homebrew will then support direct installation with:

```bash
brew install carlisle0615/openagentcleaner/oac
```

## Expected Tap Layout

```text
homebrew-openagentcleaner/
└── Formula/
    └── oac.rb
```

## Generate the Formula

After a successful release, download `checksums.txt` from the GitHub Release page and run:

```bash
gh release download v0.2.0 -p checksums.txt -D dist/release-artifacts
./scripts/generate-homebrew-formula.sh v0.2.0 dist/release-artifacts/checksums.txt > Formula/oac.rb
```

The generated formula will point to:

- `oac_<version>_darwin_arm64.tar.gz`
- `oac_<version>_darwin_amd64.tar.gz`

## Publish the Tap

Inside the tap repository:

```bash
mkdir -p Formula
cp /path/to/generated/oac.rb Formula/oac.rb
git add Formula/oac.rb
git commit -m "feat: add oac v0.2.0 formula"
git push origin main
```

## Notes

- This project currently publishes a formula, not a cask.
- The formula expects the release archives to contain an `oac` binary at the archive root.
- If archive names or repository ownership change, regenerate the formula instead of editing it by hand.
