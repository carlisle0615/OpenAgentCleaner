# Release And Homebrew Refresh

Last updated: 2026-03-09

## What Landed

- Merged PR `#2` into `main` with commit `80fa6d3`, carrying the OpenClaw session preview TUI work and the Linux-safe Bubble Tea smoke-test guard.
- Published GitHub Release [`v0.1.3`](https://github.com/carlisle0615/OpenAgentCleaner/releases/tag/v0.1.3).
- Updated the Homebrew tap repository [`carlisle0615/homebrew-openagentcleaner`](https://github.com/carlisle0615/homebrew-openagentcleaner) with `Formula/oac.rb` for `v0.1.3`.

## Validation

- Local: `make fmt`, `make build`, `make test`, `make verify-all`
- Remote: PR check `build-and-test` passed, then Release workflow run `22840576358` succeeded

## Follow-Up

- If you want Homebrew installs to feel complete, the next meaningful step is to add the tap install command to the main README once you are comfortable treating the tap as stable.
