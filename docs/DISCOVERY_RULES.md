# Discovery Rules

OpenAgentCleaner currently supports `openclaw`, `ironclaw`, and `ollama` on `macOS`.

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
