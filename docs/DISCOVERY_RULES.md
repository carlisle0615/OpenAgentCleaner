# Discovery Rules

OpenAgentCleaner currently supports `openclaw`, `ironclaw`, `ollama`, `codex`, `codex-cli`, `claudecode`, `cursor`, and `antigravity` on `macOS`.

## OpenClaw

`safe`

- `~/.openclaw/logs`
- `/tmp/openclaw/*.log*`

`confirm`

- `~/.openclaw/agents`
- `~/.openclaw/openclaw.json`
- `~/.openclaw/.env`
- `~/.openclaw/extensions`
- `~/Library/LaunchAgents/ai.openclaw*.plist`
- `~/Library/LaunchAgents/bot.molt*.plist`
- `~/Library/LaunchAgents/com.openclaw*.plist`

`manual`

- `~/.openclaw/workspace`

## IronClaw

`safe`

- `~/.ironclaw/logs`

`confirm`

- `~/.ironclaw/.env`
- `~/.ironclaw/ironclaw.db`
- `~/.ironclaw/config.toml`
- `~/.ironclaw/session.json`
- `~/.ironclaw/mcp-servers.json`
- `~/.ironclaw/settings.json`
- `~/.ironclaw/bootstrap.json`
- `~/.ironclaw/channels`
- `~/.ironclaw/tools`
- `~/.ironclaw/history`
- `~/.ironclaw/*-pairing.json`
- `~/.ironclaw/*-allowFrom.json`
- `~/.ironclaw/*-approve-attempts.json`
- `~/Library/LaunchAgents/com.ironclaw.daemon.plist`

## Ollama

`safe`

- `~/.ollama/logs`
- `~/Library/Saved Application State/com.electron.ollama.savedState`
- `~/Library/Caches/com.electron.ollama`
- `~/Library/Caches/ollama`
- `~/Library/WebKit/com.electron.ollama`

`confirm`

- `~/Library/Application Support/Ollama`
- `~/.ollama/models` or `OLLAMA_MODELS`
- `~/.ollama/server.json`
- `/Applications/Ollama.app`
- `/usr/local/bin/ollama`

`manual`

- `~/.ollama/id_ed25519`
- `~/.ollama/id_ed25519.pub`

## Codex Desktop

`confirm`

- `~/Library/Application Support/Codex/Session Storage`
- `~/Library/Application Support/Codex/Local Storage`

Notes

- Desktop app conversations are browsed from the shared `~/.codex` rollout/database store, but desktop cleanup candidates stay scoped to the desktop app's own Electron storage roots.

## Codex CLI

`confirm`

- `~/.codex/sessions`
- `~/.codex/archived_sessions`
- `~/.codex/session_index.jsonl`
- `~/.codex/state_*.sqlite*`

## Claude Code

`confirm`

- `~/.claude/transcripts`
- `~/.claude/projects`
- `~/.claude/history.jsonl`
- `~/Library/Application Support/Claude/Session Storage`
- `~/Library/Application Support/Claude/Local Storage`
- `~/Library/Application Support/Claude/IndexedDB`

## Cursor

`confirm`

- `~/Library/Application Support/Cursor/User/globalStorage/state.vscdb*`
- `~/Library/Application Support/Cursor/User/workspaceStorage`

Notes

- Conversation preview and per-session deletion use `cursorDiskKV` keys such as `composerData:*`, `bubbleId:*`, and `messageRequestContext:*`.

## Antigravity

`confirm`

- `~/Library/Application Support/Antigravity/User/globalStorage/state.vscdb*`
- `~/Library/Application Support/Antigravity/User/workspaceStorage`

Notes

- Analyze mode can preview task-level session content from `~/.gemini/antigravity/brain/*/task.md`.
- Antigravity session deletion is intentionally disabled until protobuf-backed indexes can be updated coherently.
