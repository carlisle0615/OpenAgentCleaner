# Task Plan: Add Codex, Antigravity, Cursor, and Claude Code cleanup support

## Goal

- Research how Codex, Antigravity, Cursor, and Claude Code store local caches or persistent state on macOS
- Map those paths into OpenAgentCleaner with conservative safety classifications
- Expose the new assistants through scan/clean flows without weakening existing deletion guardrails
- Split Codex desktop vs `codex-cli` as separate assistants and data sources
- Generalize Analyze conversation browsing so non-OpenClaw assistants can show title and content previews
- Keep per-session deletion index-safe: do not allow session deletion unless all related indexes/state can be updated coherently

## Touched Files

- `docs/current-plan.md`
- `docs/current-plans/codex-cursor-claude-antigravity-cleanup.md`
- `internal/cleaner/analyze.go`
- `internal/cleaner/analyze_tui.go`
- `internal/cleaner/discovery.go`
- `internal/cleaner/discovery_session_tools.go`
- `internal/cleaner/openclaw_sessions.go`
- `internal/cleaner/session_delete_tx.go`
- `internal/cleaner/sessions*.go`
- `internal/cleaner/types.go`
- `internal/cleaner/run.go`
- `internal/cleaner/output.go`
- `internal/cleaner/*_test.go`
- `README.md`
- `docs/DISCOVERY_RULES.md`
- `docs/repo-map.md`
- `docs/handoffs/2026-03-09-codex-cursor-claude-antigravity-cleanup.md`

## Todo

- [x] Research the current macOS storage layout for Codex, Antigravity, Cursor, and Claude Code, with source links and clear safe/confirm/manual boundaries
- [x] Add discovery rules and assistant selection support in the CLI for the new tools
- [x] Add or update tests for assistant parsing and candidate discovery
- [x] Split `codex` desktop and `codex-cli` in assistant discovery, help text, output, and docs
- [x] Replace OpenClaw-only Analyze session handling with a shared conversation provider model
- [x] Add title/content preview support for assistants whose local session data is parseable (`openclaw`, `codex`, `codex-cli`, `claudecode`, `cursor`, and task-level `antigravity`)
- [x] Gate session deletion on index consistency and implement provider-specific index cleanup where feasible
- [x] Validate with `make fmt`, `make test`, and focused Go tests if needed
- [x] Sync docs in `README.md`, `docs/DISCOVERY_RULES.md`, `docs/repo-map.md`, `docs/current-plan.md`, and the existing handoff

## Validation

- Run `make fmt`
- Run `make test`
- Run targeted Go tests if implementation details need quicker confirmation during iteration

## Doc Sync

- Update `docs/current-plan.md` to track this task and final status
- Update this plan file with results and any unverified items
- Update `README.md` and `docs/DISCOVERY_RULES.md` for supported assistants and safety boundaries
- Update `docs/repo-map.md` if the module boundary summary changes
- Add `docs/handoffs/2026-03-09-codex-cursor-claude-antigravity-cleanup.md` on completion

## Outcome Notes

- Task started on 2026-03-09 to expand cleanup coverage beyond OpenClaw, IronClaw, and Ollama
- Scope refined on 2026-03-09: first identify where app/CLI user conversation sessions are stored before deciding cleanup integration
- Current local findings:
  - Codex CLI stores thread rollouts under `~/.codex/sessions/YYYY/MM/DD/*.jsonl`, archived rollouts under `~/.codex/archived_sessions/*.jsonl`, and indexes them via `~/.codex/session_index.jsonl` plus `~/.codex/state_5.sqlite`
  - Claude Code stores transcript JSONL files under `~/.claude/transcripts/*.jsonl` and per-project session files under `~/.claude/projects/<project>/*.jsonl` with `sessions-index.json`
  - Cursor does not expose plain-text transcript files on this machine; current evidence points to VS Code-style state in `~/Library/Application Support/Cursor/User/globalStorage/state.vscdb` and `~/Library/Application Support/Cursor/User/workspaceStorage/*/state.vscdb`
  - Antigravity shows the same VS Code-style pattern as Cursor, with chat/session state in `~/Library/Application Support/Antigravity/User/globalStorage/state.vscdb` and workspace-scoped `state.vscdb` files
- Implementation completed on 2026-03-09:
  - Discovery logic was split into multiple `internal/cleaner/discovery*.go` files to avoid growing a giant assistant rules file
  - Added assistant support for `codex`, `codex-cli`, `claudecode`, `cursor`, and `antigravity`
  - Analyze now uses a shared conversation provider model instead of an OpenClaw-only session flow
  - `codex`, `codex-cli`, `claudecode`, `cursor`, and task-level `antigravity` sessions now expose title/content previews in the TUI
  - Session deletion is now guarded by provider-specific consistency rules:
    - OpenClaw updates `sessions.json` coherently with transcript deletion
    - Codex updates SQLite thread state plus `session_index.jsonl` before final rollout removal, with backup/rollback protection
    - Claude Code updates `sessions-index.json` together with transcript deletion, with rollback protection
    - Cursor deletes all related `cursorDiskKV` keys in one SQLite transaction
    - Antigravity remains preview-only because protobuf-backed indexes are still not safely writable
  - Updated README, discovery rules, repo map, and tests
  - Follow-up hardening on 2026-03-09:
    - Cursor session discovery/preview now tolerates `NULL` `cursorDiskKV.value` rows instead of failing `oac analyze`
    - Added regression coverage for nullable Cursor SQLite values
  - Follow-up performance work on 2026-03-09:
    - `assistantAnalyzeSummary` now scans assistants concurrently during `oac analyze` startup
    - Assistant menu detail now reuses cached summary counts instead of re-scanning sessions and leftovers on every render
  - Follow-up UX work on 2026-03-09:
    - `scan`, `clean`, and `analyze` now accept `-v` / `--verbose`
    - verbose mode prints scan progress to `stderr` so human users can see which assistants and local stores are being scanned without polluting normal output
  - Follow-up performance fix on 2026-03-09:
    - Cursor session discovery no longer uses `LIKE 'composerData:%'` against `cursorDiskKV`
    - Cursor provider now uses key-range queries that hit the SQLite primary-key index instead of full-table scans on large `state.vscdb` files
  - Follow-up TUI refinement on 2026-03-09:
    - session list pane is wider for long Codex / Antigravity / Cursor conversation titles
    - selected sessions now show inline preview text in the right pane instead of metadata-only detail
    - `Enter` still opens the larger dedicated preview screen
    - mixed Chinese/English labels and preview text now use display-width-aware truncation/wrapping instead of raw byte/rune counts
    - the right detail pane now has an explicit width so preview content renders in-pane instead of collapsing the layout

## Validation Notes

- `go test ./internal/cleaner/...`
- `make fmt`
- `make test`

## Unverified

- For Codex desktop app, cleanup candidates stay scoped to Electron storage roots even though conversation browsing is classified from the shared `~/.codex` rollout/database store
- For Antigravity, the global VS Code state store looks protobuf-backed, but local task content is available under `~/.gemini/antigravity/brain/<id>/task.md`; deletion remains disabled until index-safe cleanup is defined
