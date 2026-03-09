# Decisions

Last updated: 2026-03-09

## 2026-03-09 Repository Context Governance Baseline

- Added `AGENTS.md`, `docs/repo-map.md`, `docs/invariants.md`, `docs/current-plan.md`, `docs/current-plans/`, and `docs/handoffs/`
- The goal is not extra process for its own sake, but stable repository-local context, task indexing, and handoff artifacts for long-term maintenance
- Dynamic task state does not belong in long-lived docs; it is tracked in `docs/current-plan.md` and the linked task plan files

## 2026-03-09 Deletion Safety Boundaries Are Long-Lived Product Rules

- The `safe`, `confirm`, and `manual` classes are behavior boundaries, not just display labels
- `manual` is never auto-deleted, and `confirm` is only eligible when explicitly opted in
- Non-interactive mode must require `--yes` or `--dry-run` to reduce accidental deletion risk in scripts and agent usage

## When to Add an ADR

- Add an ADR only when the repository needs to preserve the reasoning behind a long-lived design choice
- Examples: multi-platform support, release pipeline introduction, CLI command model refactors, or safety model changes
