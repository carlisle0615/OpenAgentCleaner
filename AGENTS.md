# AGENTS

## LLM Collaboration Context

### Stable Docs (Long-Lived Rules)

- `AGENTS.md`: repository-wide collaboration rules, test/release contracts, definition of done
- `docs/repo-map.md`: codebase map, entry points, module boundaries, common cross-file changes
- `docs/invariants.md`: constraints that should not be changed lightly, source-of-truth priority, validation boundaries
- `docs/decisions.md` / `docs/adr/*.md`: long-lived decision records

### Dynamic Docs (Task Handoffs)

- `docs/current-plan.md`: dynamic task index listing active tasks, status, and linked plan files
- `docs/current-plans/*.md`: one detailed plan per task; parallel tasks must not share a single plan
- `docs/handoffs/*.md`: end-of-task handoff summaries

### Source of Truth Priority

1. Code and tests
2. `README.md`, `go.mod`, `Makefile`, `.github/workflows/*.yml`
3. `docs/repo-map.md`, `docs/invariants.md`, `docs/decisions.md`
4. Other explanatory docs

### Documentation Maintenance Rules

- Update `docs/repo-map.md` when module boundaries, entry points, or major data flow change
- Update `AGENTS.md` or `docs/invariants.md` when long-lived constraints, test workflow, or directory conventions change
- Update `docs/current-plan.md` when starting or switching tasks
- If there are already parallel tasks, create or reuse a matching file under `docs/current-plans/` before updating `docs/current-plan.md`
- Add or update a `docs/handoffs/*.md` file when finishing a medium or larger task
- Update `docs/decisions.md` and add an ADR when needed for long-term architectural or cross-module behavior changes

### Dynamic Doc Governance

- `docs/current-plan.md` is an index only; it does not hold detailed per-task todos
- Each parallel task maps to exactly one `docs/current-plans/<topic>.md`
- Continue updating the same plan file for follow-up work on the same topic instead of creating duplicates
- `docs/handoffs/*.md` are only for completed work, paused work, or explicit asynchronous handoff
- Small execution-only tasks usually do not need a handoff unless the user asks for one or the task needs async continuation
- When the same topic progresses multiple times in a day, prefer updating the existing handoff instead of creating fragmented duplicates
- Handoffs must be written for the next maintainer, not as command logs or low-risk step-by-step transcripts

## Agent Workflow

- Before implementation, the agent must create a task plan or todo with at least: goal, touched files, validation steps, and doc-sync steps
- Do not skip planning and jump straight into coding; even small tasks need a short plan
- Every todo must explicitly include both validation and context-doc updates
- If a task changes behavior, structure, or workflow, the doc step must name the specific files, such as `docs/current-plan.md`, `docs/current-plans/*.md`, `docs/decisions.md`, or `docs/handoffs/*.md`
- Before closing a task, the agent must check that every todo item is complete, especially validation and doc-sync
- If scope changes mid-task, update the todo first and continue afterward

## Definition of Done

- A task is not done when code is written; it is done when validation and context docs are also complete
- If the change affects behavior, structure, or operations, the corresponding docs must be updated
- If the work results in no code changes, record the reason, impact scope, and any unverified items in the plan or handoff

## Test Organization Strategy

### 1. Go Unit Tests (Default)

- Location: `*_test.go` files next to the implementation
- Use: module logic, edge cases, argument parsing, output formatting
- Rule: prefer the standard `testing` package and keep tests close to the code they verify

### 2. Go Integration Tests

- Location: `tests/integration/` or package-specific locations when appropriate
- Use: cross-module command flow verification such as `scan -> clean`
- Rule: add these only when a full CLI or filesystem flow needs to be covered

### 3. Regression Scripts (CI Gate)

- Location: `tests/regression/ci/`
- Naming: `test_<topic>.sh`
- Rule: must be non-interactive, repeatable, deterministic, and free of external account dependencies; failures must exit non-zero

### 4. Regression Scripts (Manual)

- Location: `tests/regression/manual/`
- Naming: `test_<topic>.sh`
- Use: checks that depend on real machine state, actual macOS paths, or external tools
- Rule: keep them long-lived, but do not include them in the default CI gate

### 5. Temporary Scripts (One-Off Debugging)

- Directory: `scripts/tmp/`
- Naming: `tmp_<topic>_<yyyymmdd>.sh`
- Rule: never run in CI; never hardcode secrets; remove within 14 days after the issue is resolved

### 6. Manual Diagnostic Scripts (Long-Lived, Non-Gating)

- Directory: `scripts/`
- Naming: `diagnose_<topic>.sh`
- Rule: may remain long-term; failures must exit non-zero; avoid modifying business data

## CI Contract

- The default PR / push gate includes only:
  - Go format check (currently `gofmt -w main.go internal/cleaner/*.go && git diff --exit-code`)
  - Build check (`make build`)
  - Go tests (`make test`)
- Any check that depends on a real user home directory, external accounts, or extra local permissions must live in `tests/regression/manual/` and must not block mainline merges

## CD Contract

- Tag pushes matching `v*` trigger `.github/workflows/release.yml`, which runs GoReleaser on `ubuntu-latest`
- Release packaging publishes macOS archives plus `checksums.txt` for installer and Homebrew consumers
- Homebrew distribution is published through the separate tap repository `carlisle0615/homebrew-openagentcleaner`
- Any change to archive names, release assets, installer behavior, or tap publishing must update this file, `.goreleaser.yaml`, `docs/RELEASING.md`, and `scripts/generate-homebrew-formula.sh` together

## Runbook

- Format: `make fmt`
- Build: `make build`
- Test: `make test`
- CI-safe regression: `bash tests/regression/run_ci.sh`
- Manual regression: `bash tests/regression/run_manual.sh`
- Single manual regression: `bash tests/regression/manual/test_<topic>.sh`
