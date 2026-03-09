# Release And Homebrew Refresh

## Goal

Push the current staged repository changes through the protected-branch workflow, cut a new GitHub Release from a green `main`, and update the Homebrew tap formula to the newly published version.

## Touched Files

- `docs/current-plan.md`
- `docs/current-plans/release-brew-refresh.md`
- Release-related code or tests only if required to unblock CI
- Homebrew formula files in this repo or `carlisle0615/homebrew-openagentcleaner`

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
