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
- Read candidate `id` values from `scan` output and prefer `clean --id ...` for exact deletion.
- Real deletion requires `--yes`.
- Non-interactive cleanup also requires an explicit selector: `--id`, `--kind`, or `--safety`.
- Safe previews can use `--dry-run` instead of `--yes`.
- `manual` items are never auto-deleted.
- `confirm` items are deletable only when you target them explicitly.

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

- `candidates[].id`
- `candidates[].deletable`
- `candidates[].requires_confirmation`
- `operation`
- `mode`
- `dry_run`
- `assistants`
- `candidates`
- `summary`

## Non-Interactive Cleanup

Preview a cleanup without deleting anything:

```bash
oac clean --safety safe --mode agent --dry-run --json
```

Delete only `safe` items:

```bash
oac clean --safety safe --mode agent --yes --json
```

Preview specific review items:

```bash
oac clean --kind models --assistants ollama --mode agent --dry-run --json
```

Delete exactly one scanned candidate:

```bash
oac clean --id <candidate-id> --mode agent --yes --json
```

## Safety Model In Agent Runs

The deletion model is conservative:

- `safe`
  Disposable logs, caches, and runtime leftovers.
- `confirm`
  Persistent local state such as sessions, settings, or models. These are skipped unless you target them explicitly.
- `manual`
  Visibility-only items that require human judgment. These are reported but not deleted.

If you run `clean` non-interactively without an explicit selector plus `--yes` or `--dry-run`, the command fails instead of guessing intent.

## Recommended Automation Pattern

1. Run `scan --mode agent --json` and inspect `summary`.
2. Build a preview with `clean --id ... --dry-run` or `clean --kind ... --dry-run`.
3. Use `--yes` only in the final destructive step.
4. Prefer `--id` over broader selectors when you need deterministic cleanup.

Example:

```bash
oac scan --assistants openclaw,ironclaw,ollama --mode agent --json
oac clean --assistants ollama --kind models --mode agent --dry-run --json
oac clean --id <candidate-id> --mode agent --yes --json
```

## Human-Only Boundaries

These workflows are intentionally not part of agent mode:

- `oac`
  The guided interactive home screen.
- `oac analyze`
  The Bubble Tea TUI for browsing leftovers and OpenClaw conversations.
- Item-by-item review in the TUI before deletion.

For those flows, use the human guide in [README.md](../README.md).
