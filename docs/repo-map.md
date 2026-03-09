# Repo Map

Last updated: 2026-03-09

## Purpose

- Give a new session or model a low-cost entry point before it dives into code
- Record only high-value, relatively stable structural information; task-level state belongs in `docs/current-plan.md`

## Source of Truth

1. Code and tests
2. `README.md`
3. `go.mod`, `Makefile`
4. `.github/workflows/ci.yml`

## Repository Overview

- `main.go`
  - Process entry point; only forwards into `internal/cleaner.Run`
- `internal/cleaner/`
  - `run.go`: command dispatch, flag parsing, interactive home menu, main `scan` / `clean` flow
  - `assistants.go`: assistant-specific discovery rules, path classification, deletion eligibility boundaries
  - `types.go`: core data structures such as `Candidate`, `Report`, `Summary`, plus mode normalization helpers
  - `output.go`: human-readable report output
- `scripts/`
  - `install.sh`, `uninstall.sh`: install and uninstall helpers
- `tests/regression/`
  - `run_ci.sh`: CI-safe regression entry point, runs `ci/test_*.sh` in order
  - `run_manual.sh`: manual regression entry point, runs `manual/test_*.sh` in order
  - `ci/`, `manual/`: regression script directories
- `.github/workflows/ci.yml`
  - Current CI workflow for format check, build, and tests
- `docs/`
  - `current-plan.md`: active task index
  - `current-plans/`: one plan file per task, supports parallel work
  - `handoffs/`: handoff summaries, only for information a future maintainer actually needs

## Key Entry Points

- Process entry: `main.go`
- Command dispatch: `internal/cleaner/run.go`
- Discovery rules: `internal/cleaner/assistants.go`
- Report structures: `internal/cleaner/types.go`
- User-facing output: `internal/cleaner/output.go`
- Build and test entry points: `Makefile`
- Regression entry point: `tests/regression/run_ci.sh`
- CI definition: `.github/workflows/ci.yml`

## Main Flow

1. The process enters through `main.go` into `cleaner.Run`
2. `run.go` decides between interactive home menu, `scan`, `clean`, or help output
3. `scanReport` / `cleanReport` assemble runtime options and call `discoverCandidates`
4. `assistants.go` gathers candidate paths per assistant and labels them as `safe`, `confirm`, or `manual`
5. `cleanReport` decides deletion eligibility and confirmation flow based on mode, `--include-confirm`, `--yes`, and `--dry-run`
6. The final report is emitted through `output.go` or JSON serialization

## Current Module Boundaries

- `main.go` should not contain business logic
- `run.go` owns CLI interaction and flow orchestration, not assistant-specific path knowledge
- `assistants.go` is the main source of truth for deletion boundaries and platform path knowledge
- `types.go` owns shared structures, mode handling, and small utilities
- `output.go` is presentation only and must not change deletion decisions

## Common Cross-File Changes

- Add support for a new assistant:
  - Update `internal/cleaner/assistants.go`
  - Update `internal/cleaner/types.go` or `internal/cleaner/output.go` if needed
  - Update `README.md` and the safety classification documentation
- Change CLI flags or command behavior:
  - Start in `internal/cleaner/run.go`
  - Then check `README.md`, CI, and related tests
- Change output formatting or JSON structure:
  - Update `internal/cleaner/output.go` or `internal/cleaner/types.go`
  - Evaluate downstream agent-mode consumers for breakage
- Add new regression or diagnostic scripts:
  - Update `tests/regression/*` or `scripts/*`
  - Sync `AGENTS.md` or `docs/invariants.md` if the workflow contract changes

## Fragile Areas / Cautions

- The rules in `assistants.go` are deletion boundaries; bad classification can become real data loss
- `manual` is the last safety boundary and should not be downgraded just because a path looks cache-like
- The non-interactive behavior in `cleanReport` depends on `--yes` / `--dry-run`; this is a core agent-mode guardrail
- CI runs on `macos-latest`, but `go test ./...` should remain independent from local machine state whenever possible
- `gofmt -w ... && git diff --exit-code` modifies files before validating cleanliness, so any format target changes must stay in sync across CI and `Makefile`

## Suggested Reading Order

1. `AGENTS.md`
2. `docs/repo-map.md`
3. `docs/invariants.md`
4. `docs/current-plan.md`
5. The relevant `docs/current-plans/*.md`
6. `README.md`
7. `internal/cleaner/run.go` and `internal/cleaner/assistants.go`
