# ADR 0001: Repository Context Contract

Date: 2026-03-09

## Status

Accepted

## Context

`OpenAgentCleaner` is still small, but it already includes deletion safety boundaries, CLI behavior, CI expectations, and multiple categories of filesystem discovery rules. As assistant support grows, relying only on the README and temporary conversation context will make it easy to lose critical reasoning:

- adding rules that accidentally weaken `safe` / `confirm` / `manual` boundaries
- switching tasks and forgetting validation or README updates
- losing the "why" behind structural decisions across multiple sessions

## Decision

Create a set of repository-local collaboration docs:

- `AGENTS.md`: collaboration rules, test/release contracts, definition of done
- `docs/repo-map.md`: repository map, entry points, module boundaries
- `docs/invariants.md`: long-lived constraints that should not be weakened casually
- `docs/decisions.md` / `docs/adr/*.md`: durable design rationale
- `docs/current-plan.md` and `docs/current-plans/*.md`: dynamic task index and detailed task plans
- `docs/handoffs/*.md`: task handoff summaries

## Consequences

### Positive

- New executors can get into context faster
- Long-lived constraints and dynamic task state stay separated, reducing documentation drift
- The deletion safety model and validation expectations become explicit repository policy

### Cost

- Any change to structure, workflow, or safety boundaries now requires doc maintenance
- Maintainers need to follow plan and handoff discipline instead of changing code without leaving context
