# Agent CLI

This document is for non-interactive use of `oac` from scripts, CI jobs, or other local agents.

If you are using OpenAgentCleaner manually in a terminal, stay on the human path in [README.md](../README.md).

## What Agent Mode Is For

Use agent mode when you need:

- machine-readable JSON output
- deterministic non-interactive behavior
- explicit deletion guardrails for automation

Agent mode is centered on `scan` and `clean`. The `analyze` command is human-only and currently rejects `--mode agent` and `--json`.

## Core Rules

- Use `--mode agent --json` for structured output.
- Real deletion requires `--yes`.
- Safe previews can use `--dry-run` instead of `--yes`.
- `manual` items are never auto-deleted.
- `confirm` items are excluded unless `--include-confirm` is set.

## Read-Only Scans

Scan everything with JSON output:

```bash
oac scan --mode agent --json
```

Scan only specific assistants:

```bash
oac scan --assistants openclaw,ollama --mode agent --json
```

This prints a JSON report with:

- `operation`
- `mode`
- `dry_run`
- `assistants`
- `candidates`
- `summary`

## Non-Interactive Cleanup

Preview a cleanup without deleting anything:

```bash
oac clean --mode agent --dry-run --json
```

Delete only `safe` items:

```bash
oac clean --mode agent --yes --json
```

Delete both `safe` and `confirm` items:

```bash
oac clean --mode agent --include-confirm --yes --json
```

Preview both `safe` and `confirm` items:

```bash
oac clean --mode agent --include-confirm --dry-run --json
```

## Safety Model In Agent Runs

The deletion model is conservative:

- `safe`
  Disposable logs, caches, and runtime leftovers.
- `confirm`
  Persistent local state such as sessions, settings, or models. These are skipped unless `--include-confirm` is provided.
- `manual`
  Visibility-only items that require human judgment. These are reported but not deleted.

If you run `clean` non-interactively without `--yes` or `--dry-run`, the command fails instead of guessing intent.

## Recommended Automation Pattern

1. Run `scan --mode agent --json` and inspect `summary`.
2. Run `clean --dry-run --mode agent --json` before any real deletion.
3. Add `--include-confirm` only when you intentionally want to remove saved state.
4. Use `--yes` only in the final destructive step.

Example:

```bash
oac scan --assistants openclaw,ironclaw,ollama --mode agent --json
oac clean --assistants openclaw,ironclaw,ollama --mode agent --include-confirm --dry-run --json
oac clean --assistants openclaw,ironclaw,ollama --mode agent --include-confirm --yes --json
```

## Human-Only Boundaries

These workflows are intentionally not part of agent mode:

- `oac`
  The guided interactive home screen.
- `oac analyze`
  The Bubble Tea TUI for browsing leftovers and OpenClaw conversations.
- Item-by-item review in the TUI before deletion.

For those flows, use the human guide in [README.md](../README.md).
