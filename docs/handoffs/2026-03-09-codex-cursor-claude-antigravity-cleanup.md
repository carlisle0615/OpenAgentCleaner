# Handoff: Codex, Codex CLI, Claude Code, Cursor, and Antigravity session cleanup support

## Goal

- Extend OpenAgentCleaner with conservative discovery support for local conversation/session state used by Codex Desktop, Codex CLI, Claude Code, Cursor, and Antigravity
- Refactor discovery and analyze session handling so the codebase does not regress into another giant assistant file
- Keep session deletion index-safe: if a provider cannot update all linked indexes or state coherently, it must stay preview-only

## Results

- Split discovery logic from the former single `internal/cleaner/assistants.go` into:
  - `internal/cleaner/discovery.go`
  - `internal/cleaner/discovery_openclaw.go`
  - `internal/cleaner/discovery_ironclaw.go`
  - `internal/cleaner/discovery_ollama.go`
  - `internal/cleaner/discovery_session_tools.go`
- Added a shared conversation provider layer:
  - `internal/cleaner/sessions.go`
  - `internal/cleaner/sessions_openclaw.go`
  - `internal/cleaner/sessions_codex.go`
  - `internal/cleaner/sessions_claudecode.go`
  - `internal/cleaner/sessions_cursor.go`
  - `internal/cleaner/sessions_antigravity.go`
  - `internal/cleaner/sessions_sqlite.go`
  - `internal/cleaner/session_delete_tx.go`
- Split `codex` and `codex-cli` as separate assistants:
  - `codex` now means Codex Desktop app storage roots
  - `codex-cli` means `~/.codex` rollout, index, and SQLite state
  - session browsing classifies shared `~/.codex` threads into desktop vs CLI using rollout metadata plus DB source fields
- Analyze is no longer OpenClaw-only for sessions:
  - per-session title/content preview now works for `openclaw`, `codex`, `codex-cli`, `claudecode`, `cursor`, and task-level `antigravity`
  - providers without safe delete support remain preview-only in the TUI
- Added index-safe deletion behavior:
  - OpenClaw deletes transcripts together with `sessions.json`
  - Codex deletes SQLite thread rows plus `session_index.jsonl`, then finalizes rollout removal
  - Claude Code deletes transcripts together with `sessions-index.json`
  - Cursor deletes all related `cursorDiskKV` keys in one transaction
  - Antigravity session deletion is disabled because protobuf-backed indexes are still not safely writable
- Added rollback protection for file-backed providers:
  - staged transcript/rollout deletion
  - backup/restore of JSON indexes
  - backup/restore of Codex SQLite files when later steps fail
- Follow-up hardening:
  - Cursor session discovery and preview now ignore `NULL` `cursorDiskKV.value` rows instead of failing the entire analyze flow
- Follow-up performance work:
  - analyze startup now computes per-assistant summaries concurrently
  - assistant menu rendering now uses cached counts instead of re-running full session/leftover discovery each frame
- Follow-up UX work:
  - `scan`, `clean`, and `analyze` now support `-v` / `--verbose`
  - verbose progress is emitted on `stderr`, including assistant-level and store-level scan messages
- Follow-up Cursor performance fix:
  - Cursor session discovery and cleanup now use indexed key-range queries instead of `LIKE 'prefix%'` scans on `cursorDiskKV`
  - this avoids full-table scans on large `state.vscdb` files during analyze startup
- Follow-up TUI refinement:
  - session list panes are wider for longer conversation labels
  - selected sessions now render inline preview text in the detail pane while keeping `Enter` for the larger preview screen
  - mixed Chinese/English content now uses display-width-aware truncation and wrapping to avoid pane misalignment
  - detail panes now reserve explicit width so preview text stays in the right pane
- Follow-up manual diagnostics:
  - added `tests/regression/manual/test_analyze_latency.sh` and a gated `internal/cleaner/analyze_manual_test.go`
  - current local baseline shows session discovery is the dominant latency:
    - `codex` init ~`5.7s`, open conversations ~`4.2s`
    - `codex-cli` init ~`8.2s`
    - `cursor` init ~`4.1s`
    - preview open itself is comparatively cheap once the session list is loaded
- Follow-up session-list rendering fix:
  - large session and candidate lists now render only the visible window instead of every row
  - list padding now uses display width instead of `fmt` width specifiers, reducing mixed-width terminal misalignment
- Follow-up analyze discovery cache:
  - added `internal/cleaner/analyze_cache.go` so a single `oac analyze` run reuses discovered sessions and leftovers between summary, menu, and list screens
  - session and leftover caches are invalidated after delete actions to preserve consistency
  - added `internal/cleaner/analyze_cache_test.go` to cover cache reuse and invalidation
  - reran `tests/regression/manual/test_analyze_latency.sh` on 2026-03-10 and confirmed that entering conversation lists no longer replays multi-second rescans:
    - `codex` init ~`5.952s`, open conversations ~`0s`, open preview ~`48ms`
    - `claudecode` init ~`255ms`, open conversations ~`0s`, open preview ~`6ms`
    - `antigravity` init ~`186ms`, open conversations ~`0s`, open preview ~`0s`
    - `cursor` init ~`2.102s`, open preview ~`2ms`
    - `codex-cli` init ~`4.289s`, open preview ~`17ms`
  - the remaining latency hotspot is initial provider discovery for single-assistant runs, not session-list entry after startup
- Follow-up Codex discovery optimization:
  - Codex session discovery now trusts SQLite `threads.source` for known values and only scans rollout JSONL files for `source='unknown'`
  - `codex` discovery is filtered to `vscode` plus fallback-unknown rows; `codex-cli` discovery is filtered to non-`vscode` rows
  - added regression coverage for both direct source classification and unknown-source rollout fallback
  - reran manual latency on 2026-03-10:
    - `codex` init ~`39ms`, open conversations ~`0s`, open preview ~`45ms`
    - `codex-cli` init ~`44ms`, open preview ~`16ms`
  - after this change, the dominant remaining single-assistant startup cost is `cursor`, not Codex
- Follow-up structure cleanup:
  - moved non-OpenClaw session providers and their storage helpers into `internal/cleaner/sessionstore/`
  - kept OpenClaw-specific parsing/deletion in `internal/cleaner/` because it still shares helper and test surface with the root package
  - updated `make fmt` and `scripts/verify-all.sh` to recurse through all Go files so new subdirectories stay inside the default validation contract
- Follow-up discovery structure cleanup:
  - moved assistant-specific leftover discovery rules into `internal/cleaner/discoveryrules/`
  - kept `internal/cleaner/discovery.go` as the root aggregation surface so callers and tests can continue using `discoverCandidates` without package churn
  - preserved compatibility wrappers for OpenClaw-specific helpers still used by session parsing and older tests
- Follow-up analyze UI framework switch:
  - switched interactive `analyze` UI to `tview` (`internal/cleaner/analyze_tui_tview.go`) to replace hand-drawn overlay dialogs with modal components
  - kept legacy Bubble Tea path as non-TTY fallback (`runAnalyzeTUILegacy`) for deterministic tests and non-terminal runs
  - dialog-heavy flows (`f` date filter, `x` bulk delete, `d` confirm delete) now run through tview modal/forms
- Updated assistant parsing, output labels, README, discovery rules, repo map, and tests

## Validation

- `go test ./internal/cleaner/...`
- `make fmt`
- `make test`

## Remaining Risks

- Codex Desktop conversation browsing depends on the shared `~/.codex` rollout/database store; desktop cleanup candidates themselves remain limited to Electron storage roots because that is the clearer deletion boundary
- Antigravity still has protobuf-backed state in local app storage; until that format is understood well enough to rewrite safely, session deletion must remain disabled
- Cursor support is based on the observed `cursorDiskKV` schema in local installs; if Cursor changes those keys, provider parsing/deletion rules will need to be revisited

## Next Work

- If Antigravity exposes a stable writable session index format, add a deletion provider only after a full rollback-safe implementation exists
- If Codex Desktop later documents app-specific thread indexes separate from `~/.codex`, narrow desktop session classification to that store
- If Cursor expands conversation state beyond `composerData:*`, `bubbleId:*`, and `messageRequestContext:*`, update the provider and add regression coverage before enabling new deletion paths
