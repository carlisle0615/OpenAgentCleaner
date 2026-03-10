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
- `internal/cleaner/analyze_cache.go`
- `internal/cleaner/analyze_tui.go`
- `internal/cleaner/analyze_tui_tview.go`
- `internal/cleaner/discovery.go`
- `internal/cleaner/discoveryrules/*.go`
- `internal/cleaner/openclaw_sessions.go`
- `internal/cleaner/session_delete_tx.go`
- `internal/cleaner/sessions.go`
- `internal/cleaner/sessionstore/*.go`
- `internal/cleaner/types.go`
- `internal/cleaner/run.go`
- `internal/cleaner/output.go`
- `internal/cleaner/*_test.go`
- `tests/regression/manual/test_analyze_latency.sh`
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
  - Follow-up manual diagnostics on 2026-03-10:
    - there was no existing regression script for `analyze` panel open latency
    - added `tests/regression/manual/test_analyze_latency.sh` plus a gated manual timing test in `internal/cleaner/analyze_manual_test.go`
    - local timing baseline on this machine:
      - `codex` init ~`5.734s`, open conversations ~`4.232s`, open preview ~`36ms`
      - `codex-cli` init ~`8.238s`, open preview ~`17ms`
      - `cursor` init ~`4.12s`, open preview ~`2ms`
      - `claudecode` init ~`91ms`, open conversations ~`82ms`
      - `antigravity` init ~`70ms`, open conversations ~`23ms`
  - Follow-up session-list performance fix on 2026-03-10:
    - large session lists no longer render every row on each cursor move
    - session and candidate panes now render a visible window plus elision markers
    - list column padding now uses display width instead of `fmt %-Ns`, reducing ANSI/full-width misalignment
  - Follow-up analyze caching on 2026-03-10:
    - added process-local analyze discovery caches in `internal/cleaner/analyze_cache.go`
    - `assistantAnalyzeSummary`, `reloadSessions`, `reloadCandidates`, and bulk-delete source selection now reuse the same cached discovery results inside one `oac analyze` run
    - session and leftover caches are explicitly invalidated after delete actions so UI state stays coherent
    - added `internal/cleaner/analyze_cache_test.go` to verify cache reuse and invalidation behavior
    - updated manual latency on this machine after the cache wiring:
      - `antigravity` init ~`186ms`, open conversations ~`0s`, open preview ~`0s`
      - `claudecode` init ~`255ms`, open conversations ~`0s`, open preview ~`6ms`
      - `codex` init ~`5.952s`, open conversations ~`0s`, open preview ~`48ms`
      - `codex-cli` init ~`4.289s`, open preview ~`17ms`
      - `cursor` init ~`2.102s`, open preview ~`2ms`
    - remaining bottleneck is first-time provider discovery for single-assistant analyze runs, especially `codex`, `codex-cli`, and `cursor`
  - Follow-up Codex provider optimization on 2026-03-10:
    - `internal/cleaner/sessions_codex.go` now classifies known Codex thread sources directly from SQLite (`vscode` => desktop, `cli`/`exec`/`mcp` => CLI)
    - rollout JSONL metadata scans now only happen for `source='unknown'` rows instead of every thread
    - Codex SQL discovery is filtered by target assistant before row iteration
    - added tests for source-based classification plus fallback-to-rollout behavior for unknown sources
    - updated local timing after this change:
      - `codex` init ~`39ms`, open conversations ~`0s`, open preview ~`45ms`
      - `codex-cli` init ~`44ms`, open preview ~`16ms`
    - remaining first-load hotspot is now mainly `cursor`
  - Follow-up structure cleanup on 2026-03-10:
    - extracted non-OpenClaw session providers out of the flat `internal/cleaner` root into `internal/cleaner/sessionstore/`
    - root `internal/cleaner/sessions.go` now acts as a thin wrapper surface for analyze and deletion flows
    - OpenClaw session parsing/deletion stays in the root package because it still shares helper/test surface with the main cleaner package
    - format/verification scripts now recurse through all Go files so the new subdirectory is covered by default CI-safe validation
  - Follow-up discovery structure cleanup on 2026-03-10:
    - moved assistant-specific leftover discovery rules into `internal/cleaner/discoveryrules/`
    - root `internal/cleaner/discovery.go` now keeps aggregation, root-level filesystem helpers, and compatibility wrappers used by older tests/OpenClaw session parsing
    - this reduces the flat root package further without changing the external `discoverCandidates` flow
  - Follow-up analyze UI framework switch on 2026-03-10:
    - `runAnalyzeTUI` now uses a tview implementation for interactive TTY sessions (`internal/cleaner/analyze_tui_tview.go`)
    - legacy Bubble Tea implementation is retained as a non-TTY fallback (`runAnalyzeTUILegacy`) to keep existing tests and scripted execution deterministic
    - date filter / bulk-delete / delete confirm dialogs are now tview modal pages instead of hand-drawn overlay strings
    - added `tview` and `tcell` dependencies in `go.mod`

## Validation Notes

- `go test ./internal/cleaner/...`
- `make fmt`
- `make test`

## Unverified

- For Codex desktop app, cleanup candidates stay scoped to Electron storage roots even though conversation browsing is classified from the shared `~/.codex` rollout/database store
- For Antigravity, the global VS Code state store looks protobuf-backed, but local task content is available under `~/.gemini/antigravity/brain/<id>/task.md`; deletion remains disabled until index-safe cleanup is defined
