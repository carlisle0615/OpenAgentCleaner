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

- [ ] Confirm the release version and inspect current repo/release workflow state
- [ ] Finish any remaining code/doc alignment needed before release
- [ ] Run `make fmt`, `make build`, and `make test`
- [ ] Commit the release-ready state and push it to GitHub
- [ ] Tag and publish the new release
- [ ] Generate and sync the updated Homebrew formula
- [ ] Update plan status and leave a release handoff with final publish state

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
