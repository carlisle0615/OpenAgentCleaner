# Handoff: Agent-first CLI redesign

## Goal

- Move cleanup from a human-default `safe` vs `include-confirm` model to an agent-first contract with explicit selectors and stable candidate IDs

## What Changed

- `scan` JSON now includes:
  - `candidates[].id`
  - `candidates[].deletable`
  - `candidates[].requires_confirmation`
- `clean` now targets candidates via:
  - `--id`
  - `--kind`
  - `--safety`
- Removed the old `--include-confirm` path from CLI help, tests, README, and agent docs
- Non-interactive cleanup now fails unless both:
  - an explicit selector is present
  - `--yes` or `--dry-run` is present

## Validation

- `make fmt`
- `make build`
- `make test`
- Manual spot check:
  - `./bin/oac --help`
  - `./bin/oac scan --assistants ollama --mode agent --json`

## Remaining Risk

- Existing scripts that use `oac clean --include-confirm` will break and need migration
- Human-readable output still emphasizes kinds and safety labels instead of candidate IDs, which is acceptable for human use but limits copy-paste ID workflows outside JSON mode

## Suggested Follow-up

- If the tool needs stronger machine orchestration later, add a first-class `plan` subcommand that emits only selected candidates and can be replayed by `apply --plan <file>`
