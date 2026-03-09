# Release And Homebrew Refresh

Last updated: 2026-03-09

## What Landed

- Merged PR `#5` into `main` with commit `049f880`, carrying the breaking agent-first cleanup CLI redesign.
- Published GitHub Release [`v0.2.0`](https://github.com/carlisle0615/OpenAgentCleaner/releases/tag/v0.2.0).
- Updated the Homebrew tap repository [`carlisle0615/homebrew-openagentcleaner`](https://github.com/carlisle0615/homebrew-openagentcleaner) with `Formula/oac.rb` for `v0.2.0`.
- Release artifacts now include:
  - `checksums.txt`
  - `oac_0.2.0_darwin_arm64.tar.gz`
  - `oac_0.2.0_darwin_amd64.tar.gz`

## Validation

- Local: `make fmt`, `make build`, `make test`
- Remote: PR check `build-and-test` passed, then Release workflow run `22841356905` succeeded

## Follow-Up

- The repository still has a small post-release follow-up to merge: regenerated `Formula/oac.rb` plus the formula generator script now restore the MIT license field for future tap refreshes.
