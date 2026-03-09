# Clarify Agent Usage In CLI Help

## Goal

Update `oac help` so other agents can see the intended non-interactive workflow, the correct JSON-based commands, and the fact that `analyze` is human-only and not suitable for agent execution.

## Touched Files

- `internal/cleaner/run.go`
- `internal/cleaner/run_test.go`
- `docs/current-plan.md`
- `docs/current-plans/cli-help-agent-guidance.md`

## Validation Steps

1. Run `go test ./...`.
2. Review `oac --help` output in tests to confirm the new agent guidance is present and the `analyze` boundary is explicit.

## Doc-Sync Steps

- Keep `docs/current-plan.md` aligned with task status.
- No repo-map update is needed unless help output becomes part of a larger command model change.
