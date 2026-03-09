# Handoff: Collaboration Doc Baseline Migration

Date: 2026-03-09

## Goal

- Establish a durable collaboration-doc baseline for `OpenAgentCleaner`
- Turn planning, handoff, long-lived constraints, and repository structure into repository-local assets instead of implicit knowledge

## Completed

- Added `AGENTS.md` defining source-of-truth priority, planning requirements, definition of done, test organization, and CI/CD contracts
- Added `docs/repo-map.md` documenting the responsibilities of `main.go`, `internal/cleaner/*`, `scripts/`, tests, and CI
- Added `docs/invariants.md` to lock in the deletion safety model, non-interactive protections, and doc update rules
- Added `docs/decisions.md` and `docs/adr/0001-repo-context-contract.md`
- Added `docs/current-plan.md`, `docs/current-plans/README.md`, and this task plan
- Added `tests/regression/run_ci.sh` and `tests/regression/run_manual.sh` so the documented regression entry points exist in the repo

## Validation

- Passed: `make test`
- Passed: `make build`
- Passed: `bash tests/regression/run_ci.sh`

## Risks and Notes

- The docs are now adapted to this Go CLI repository; the `tests/regression/` entry points exist, but there are still no concrete regression cases in those directories yet
- The repository still has no formal release pipeline, so `AGENTS.md` documents CD as "not configured yet"
- Deletion safety boundaries are now a long-lived rule; if the meaning of `safe` / `confirm` / `manual` changes later, update `README.md`, tests, and `docs/invariants.md` together

## Suggested Next Steps

- Add baseline `*_test.go` coverage for `run.go` and `assistants.go`
- If parallel work starts later, keep updating `docs/current-plan.md` as the index instead of pushing task todos back into long-lived docs
