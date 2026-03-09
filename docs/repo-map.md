# Repo Map

Last updated: 2026-03-09

## Purpose

- Give a new session or model a low-cost entry point before it dives into code
- Record only high-value, relatively stable structural information; task-level state belongs in `docs/current-plan.md`

## Source of Truth

1. Code and tests
2. `README.md`
3. `go.mod`, `Makefile`
4. `.github/workflows/ci.yml`, `.github/workflows/release.yml`

## Repository Overview

- `main.go`
  - Process entry point; only forwards into `internal/cleaner.Run`
- `internal/cleaner/`
  - `run.go`: command dispatch, flag parsing, interactive home menu, version/help output, main `scan` / `clean` flow
  - `analyze.go`: analyze command flag parsing and TUI bootstrap
  - `analyze_tui.go`: Bubble Tea TUI for assistant browsing, OpenClaw conversation review, and item-level deletion
  - `assistants.go`: assistant-specific discovery rules, path classification, deletion eligibility boundaries
  - `openclaw_sessions.go`: OpenClaw session indexing, transcript metadata extraction, and per-session deletion/update logic
  - `types.go`: core data structures such as `Candidate`, `Report`, `Summary`, plus mode normalization helpers
  - `output.go`: guided human-readable report output
  - `ui.go`: terminal presentation helpers for badges, sections, and color support
- `scripts/`
  - `install-local.sh`: local source-tree installer used by `make install`
  - `generate-homebrew-formula.sh`: generates `Formula/oac.rb` from release checksums
  - `uninstall.sh`: uninstall helper
- `install.sh`
  - release installer that downloads the latest GitHub Release archive
- `.goreleaser.yaml`
  - release packaging definition for macOS archives and checksums
- `.github/workflows/ci.yml`
  - CI workflow for format check, build, and tests
- `.github/workflows/release.yml`
  - tag-triggered GitHub Release workflow using GoReleaser
- `docs/`
  - `RELEASING.md`: release, installer, and Homebrew tap notes
  - `repo-map.md`: this file

## Key Entry Points

- Process entry: `main.go`
- Command dispatch: `internal/cleaner/run.go`
- Discovery rules: `internal/cleaner/assistants.go`
- Report structures: `internal/cleaner/types.go`
- User-facing output: `internal/cleaner/output.go`, `internal/cleaner/ui.go`
- Build and test entry points: `Makefile`
- Release packaging: `.goreleaser.yaml`
- CI and release definitions: `.github/workflows/ci.yml`, `.github/workflows/release.yml`

## Main Flow

1. The process enters through `main.go` into `cleaner.Run`
2. `run.go` decides between interactive home menu, `scan`, `clean`, or help output
3. `analyze.go` dispatches into the Bubble Tea analyze TUI
4. `analyze_tui.go` renders the full-screen browser for assistants, leftovers, and OpenClaw conversations
5. `scanReport` / `cleanReport` assemble runtime options and call `discoverCandidates`
6. `assistants.go` gathers candidate paths per assistant and labels them as `safe`, `confirm`, or `manual`
7. `openclaw_sessions.go` reads `sessions.json` plus `*.jsonl` transcript files to support selective conversation cleanup
8. `cleanReport` decides deletion eligibility and confirmation flow based on explicit selectors (`--id`, `--kind`, `--safety`), plus `--yes` and `--dry-run`
9. The final report is emitted through `output.go` or JSON serialization

## Current Module Boundaries

- `main.go` should not contain business logic
- `run.go` owns CLI interaction and flow orchestration, not assistant-specific path knowledge
- `assistants.go` is the main source of truth for deletion boundaries and platform path knowledge
- `analyze.go` owns analyze command setup and guardrails
- `analyze_tui.go` owns interactive navigation and item-level deletion UX
- `openclaw_sessions.go` owns OpenClaw session parsing and must keep metadata updates consistent with transcript deletions
- `types.go` owns shared structures, mode handling, and small utilities
- `output.go` and `ui.go` are presentation only and must not change deletion decisions

## Common Cross-File Changes

- Add support for a new assistant:
  - Update `internal/cleaner/assistants.go`
  - Update `internal/cleaner/types.go` or `internal/cleaner/output.go` if needed
  - Update `README.md` and the safety classification documentation
- Change CLI flags or command behavior:
  - Start in `internal/cleaner/run.go` or `internal/cleaner/analyze.go`
  - Then check `README.md`, `docs/RELEASING.md`, CI, and related tests
- Change OpenClaw conversation browsing or deletion behavior:
  - Update `internal/cleaner/openclaw_sessions.go`
  - Review `internal/cleaner/analyze_tui.go` and `README.md`
- Change output formatting or JSON structure:
  - Update `internal/cleaner/output.go`, `internal/cleaner/ui.go`, or `internal/cleaner/types.go`
  - Evaluate downstream agent-mode consumers for breakage
- Change installer or release packaging:
  - Update `install.sh`, `.goreleaser.yaml`, `.github/workflows/release.yml`, and `docs/RELEASING.md`
  - Update `scripts/generate-homebrew-formula.sh` if archive names or tap conventions change

## Fragile Areas / Cautions

- The rules in `assistants.go` are deletion boundaries; bad classification can become real data loss
- `manual` is the last safety boundary and should not be downgraded just because a path looks cache-like
- The non-interactive behavior in `cleanReport` depends on explicit selectors plus `--yes` / `--dry-run`; this is a core agent-mode guardrail
- `openclaw_sessions.go` edits `sessions.json` after transcript deletions; mismatches there can create orphaned metadata or confusing UX
- CI runs on `macos-latest`, while the release workflow runs GoReleaser plus `go test ./...` on `ubuntu-latest`; tests must stay cross-platform unless they are explicitly guarded
- The release installer and Homebrew formula generator assume archive names shaped like `oac_<version>_darwin_<arch>.tar.gz`
- `gofmt -w ... && git diff --exit-code` modifies files before validating cleanliness, so any format target changes must stay in sync across CI and `Makefile`

## Suggested Reading Order

1. `AGENTS.md`
2. `docs/repo-map.md`
3. `README.md`
4. `docs/RELEASING.md`
5. `internal/cleaner/run.go`
6. `internal/cleaner/analyze.go`
7. `internal/cleaner/analyze_tui.go`
8. `internal/cleaner/openclaw_sessions.go`
9. `internal/cleaner/assistants.go`
