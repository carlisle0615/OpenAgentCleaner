# Release And Homebrew Refresh

## Goal

Push the current staged repository changes through the protected-branch workflow, cut a new GitHub Release from a green `main`, and update the Homebrew tap formula to the newly published version.

## 2026-03-09 Follow-up Scope

- Publish the breaking agent-first CLI redesign as the next release
- Refresh release documentation, release assets, and the generated Homebrew formula for the new version
- Record which publication steps were completed locally vs. on GitHub

## Touched Files

- `docs/current-plan.md`
- `docs/current-plans/release-brew-refresh.md`
- `README.md`
- `docs/AGENT_CLI.md`
- `docs/RELEASING.md`
- `.goreleaser.yaml`
- `scripts/generate-homebrew-formula.sh`
- `Formula/oac.rb`
- Release-related code or tests only if required to unblock CI
- Homebrew formula files in this repo or `carlisle0615/homebrew-openagentcleaner`

## Todo

- [x] Confirm the release version and inspect current repo/release workflow state
- [x] Finish any remaining code/doc alignment needed before release
- [x] Run `make fmt`, `make build`, and `make test`
- [x] Commit the release-ready state and push it to GitHub
- [x] Tag and publish the new release
- [x] Generate and sync the updated Homebrew formula
- [x] Update plan status and leave a release handoff with final publish state

## Outcome Notes

- Published the breaking agent-first CLI redesign as `v0.2.0`
- Because `main` is protected, the release commit shipped through PR `#5`, then the tag was pushed from the merged `main` commit
- GitHub Release workflow run `22841356905` completed successfully and published:
  - `checksums.txt`
  - `oac_0.2.0_darwin_arm64.tar.gz`
  - `oac_0.2.0_darwin_amd64.tar.gz`
- Updated the tap repository `carlisle0615/homebrew-openagentcleaner` on `main` with `Formula/oac.rb` for `v0.2.0`
- Regenerated this repository's `Formula/oac.rb` from release checksums and updated `scripts/generate-homebrew-formula.sh` to preserve the MIT license field

## Validation Results

- Local pre-release:
  - `make fmt`
  - `make build`
  - `make test`
- Branch CI:
  - PR `#5` check `build-and-test` passed
- Release:
  - Workflow run `22841356905` succeeded for tag `v0.2.0`

## Remaining Work

- Commit the post-release formula/script/doc updates in this repository and merge them back to `main`

## Validation Steps

1. Run `make fmt`.
2. Run `make build`.
3. Run `make test`.
4. Confirm the release workflow succeeds for the new tag.
5. Confirm the Homebrew formula points to the new tag and checksum.

## Doc-Sync Steps

- Keep `docs/current-plan.md` aligned with task status.
- Update `docs/repo-map.md` or `AGENTS.md` only if release or tap workflow changes materially.
- Add or update a handoff note under `docs/handoffs/` if the release/tap work leaves follow-up items.
