# Collaboration Doc Baseline Migration

Last updated: 2026-03-09

## Goal

- Bring the proven long-lived collaboration framework from Hone-Financial into `OpenAgentCleaner`
- Adapt it to this repository's Go CLI structure, CI, and safety model instead of copying Rust-oriented rules literally

## Touched Files

- `AGENTS.md`
- `docs/repo-map.md`
- `docs/invariants.md`
- `docs/decisions.md`
- `docs/current-plan.md`
- `docs/current-plans/README.md`
- `docs/current-plans/context-doc-bootstrap.md`
- `docs/handoffs/2026-03-09-context-doc-bootstrap.md`
- `docs/adr/0001-repo-context-contract.md`
- `tests/regression/run_ci.sh`
- `tests/regression/run_manual.sh`
- `tests/regression/ci/.gitkeep`
- `tests/regression/manual/.gitkeep`

## Todo

- [x] Inspect the target repository structure, CI, build, and test flow
- [x] Migrate and adapt the long-lived collaboration docs
- [x] Create the dynamic task index and this task plan
- [x] Add a handoff for future maintainers
- [x] Run validation commands and record the results

## Validation

- [x] `make test`
- [x] `make build`
- [x] `bash tests/regression/run_ci.sh`

## Doc Sync

- [x] Add `AGENTS.md`
- [x] Add `docs/repo-map.md`
- [x] Add `docs/invariants.md`
- [x] Add `docs/decisions.md`
- [x] Add `docs/current-plan.md`
- [x] Add `docs/current-plans/README.md`
- [x] Add or update `docs/handoffs/2026-03-09-context-doc-bootstrap.md`
- [x] Add `docs/adr/0001-repo-context-contract.md`
- [x] Add `tests/regression/run_ci.sh`
- [x] Add `tests/regression/run_manual.sh`
