# Invariants

Last updated: 2026-03-09

## Source-of-Truth and Documentation Priority

- Code and tests override explanatory docs
- `README.md`, `go.mod`, `Makefile`, and `.github/workflows/ci.yml` are the main implementation-facing guides
- `docs/repo-map.md`, `docs/decisions.md`, and `docs/adr/*.md` hold long-lived context
- `docs/current-plan.md` is a dynamic index, and `docs/current-plans/*.md` hold task state; neither is a place for long-lived rules

## Definition of Done

- Validation must be completed before a task is done
- Affected context docs must be updated before a task is done
- Cross-module long-term behavior or architectural decisions must be recorded in `docs/decisions.md`, with an ADR when appropriate
- Medium or larger tasks must leave a handoff when continuation, pause, or residual risk needs to be recorded explicitly

## Safety Boundary Constraints

- `scan -> classify -> confirm -> delete` is the core safety flow and must not skip classification or confirmation boundaries
- `manual` items are for visibility and human review only; they must never be auto-deleted by default
- `confirm` items may only become deletion candidates when an explicit selector such as `--id`, `--kind`, or `--safety confirm` targets them
- Non-interactive deletion must require both an explicit selector and `--yes` or `--dry-run`
- When rule confidence is low, prefer downgrading to `confirm` or `manual` instead of upgrading to `safe`

## Planning and Handoff Constraints

- Parallel tasks must not share a detailed plan file; each task gets its own `docs/current-plans/*.md`
- `docs/current-plan.md` only tracks the active task index, statuses, and links
- Ongoing work on the same topic should reuse the same handoff when possible instead of creating fragmented duplicates
- A handoff is not a command log; keep only the goal, results, validation, risks, and remaining work
- Small execution-only tasks usually do not require a handoff unless requested, needed for async continuation, or changing workflow, structure, or risk

## Tests and Scripts Constraints

- Go unit tests should live next to the implementation in `*_test.go`
- Add `tests/integration/` only when cross-module or end-to-end CLI behavior needs explicit coverage
- CI-safe regression scripts belong only in `tests/regression/ci/`
- Regressions that depend on real user directories, machine state, or extra tools belong only in `tests/regression/manual/`
- One-off debugging scripts belong only in `scripts/tmp/` and must never enter CI

## Environment and Platform Constraints

- The product is currently macOS-first; local path layout and LaunchAgents assumptions must not be generalized casually to other platforms
- Never hardcode secrets in docs, scripts, or tests
- Do not add flows that depend on external accounts, credentials, or private user data to the default CI gate
- Diagnostic scripts should avoid mutating user data

## Change Constraints

- If directory responsibilities, entry points, or major data flow change, update `docs/repo-map.md`
- If workflow, test contracts, or collaboration rules change, update `AGENTS.md` or this file
- If deletion classification, safety boundaries, or non-interactive behavior change, update `README.md` and the related tests as well
- If release or distribution strategy changes, update the corresponding docs and workflows
