# Task Plan: Redesign CLI for agent-first cleanup

## Goal

- Replace the current human-first cleanup contract with an agent-first CLI that supports stable candidate selection and explicit non-interactive deletion
- Update help text and README so the intended human and agent workflows are obvious from the command surface
- Preserve the core safety boundary that `manual` items are never auto-deleted

## Touched Files

- `internal/cleaner/run.go`
- `internal/cleaner/types.go`
- `internal/cleaner/assistants.go`
- `internal/cleaner/output.go`
- `internal/cleaner/run_test.go`
- `README.md`
- `docs/repo-map.md`
- `docs/invariants.md`
- `docs/current-plan.md`
- `docs/current-plans/agent-first-cli-redesign.md`
- `docs/handoffs/agent-first-cli-redesign.md`

## Todo

- [x] Inspect current scan and clean flow plus existing JSON/report structures
- [x] Redesign CLI flags and command behavior around explicit plan/apply semantics or equivalent stable selectors
- [x] Implement stable candidate identifiers and machine-friendly selection fields in JSON output
- [x] Update help text and README examples to document the new workflow clearly
- [x] Add or update tests for selection, validation, and deletion safety boundaries
- [x] Run formatting, build, and tests
- [x] Update context docs and leave a handoff summary

## Validation

- `make fmt`
- `make build`
- `make test`

## Doc Sync

- Update `README.md` for the new command contract and examples
- Update `docs/repo-map.md` if command flow or module boundaries change materially
- Update `docs/invariants.md` for the new non-interactive safety contract
- Update `docs/current-plan.md` and this plan file with completion state
- Add `docs/handoffs/agent-first-cli-redesign.md` with results and residual risks

## Outcome Notes

- Replaced the old `--include-confirm` cleanup model with explicit selectors: `--id`, `--kind`, and `--safety`
- Added stable `candidates[].id` plus `deletable` and `requires_confirmation` fields to scan and clean JSON output
- Non-interactive cleanup now requires both an explicit selector and `--yes` or `--dry-run`
- Human flows still work, but the default guidance now pushes users toward `clean --safety safe` for recommended cleanup and targeted selectors for review items
- Verified help output manually from the built binary

## Validation Results

- `make fmt`
- `make build`
- `make test`

## Residual Risk

- This is an intentional breaking change for anyone depending on `oac clean --include-confirm`
- Human-readable scan output still does not print candidate IDs; exact ID-based targeting remains primarily a JSON/agent workflow
