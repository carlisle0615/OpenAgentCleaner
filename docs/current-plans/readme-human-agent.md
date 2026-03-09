# Split Human And Agent CLI Docs

## Goal

Restructure the top-level documentation so README keeps a concise human-first usage section, while detailed non-interactive and JSON-driven agent workflows move into a separate document linked from README.

## Touched Files

- `README.md`
- `docs/AGENT_CLI.md`
- `docs/current-plan.md`
- `docs/current-plans/readme-human-agent.md`

## Validation Steps

1. Review `README.md` to confirm the human path is self-contained and readable.
2. Review `docs/AGENT_CLI.md` to confirm agent mode flags, guardrails, and examples are explicit.
3. Run a markdown sanity check by reading both files in the terminal after editing.

## Doc-Sync Steps

- Keep `docs/current-plan.md` aligned with task status.
- Update `docs/repo-map.md` only if the documentation structure becomes part of the stable repository map.
- Add a handoff only if this doc split leaves follow-up work.
