# OpenAgentCleaner

`OpenAgentCleaner` is a macOS-first CLI for cleaning leftover files from local AI assistants. It follows a `scan -> classify -> confirm -> delete` workflow inspired by tools like `Mole`, but focuses on assistant-specific state instead of generic developer artifacts.

The default command name is `oac`.

## Status

Current assistant support:

- `openclaw`
- `ironclaw`
- `ollama`

Current platform support:

- `macOS`

## Principles

- Human-friendly mode: interactive selection and confirmation before deletion.
- Agent-friendly mode: structured JSON output and non-interactive execution with `--yes`.
- Guided UX: a home screen, cleanup previews, and plain-language safety cues for non-technical users.
- Explicit safety classes:
  - `safe`: logs, caches, and disposable runtime leftovers.
  - `confirm`: persistent state that should only be removed intentionally.
  - `manual`: paths that are listed for review but never deleted automatically.

## Installation

Install the latest release:

```bash
curl -fsSL https://raw.githubusercontent.com/carlisle0615/OpenAgentCleaner/main/install.sh | bash
```

Install to `~/.local/bin/oac` from the local source tree:

```bash
make install
```

Install to a custom prefix:

```bash
PREFIX=/usr/local make install
```

Uninstall:

```bash
make uninstall
```

If `~/.local/bin` is not on your `PATH`, add it before running `oac`.

Homebrew is designed to use a separate tap repository. The planned install command is:

```bash
brew install carlisle0615/openagentcleaner/oac
```

The release and tap workflow is documented in [docs/RELEASING.md](docs/RELEASING.md).
The Homebrew tap layout is documented in [docs/HOMEBREW_TAP.md](docs/HOMEBREW_TAP.md).

## Usage

Launch the interactive home menu:

```bash
oac
```

Show the installed version:

```bash
oac version
```

Scan only:

```bash
oac scan
oac scan --assistants openclaw,ollama
oac scan --mode agent --json
```

Clean `safe` items only:

```bash
oac clean
oac clean --dry-run
oac clean --mode agent --yes --json
```

Include `confirm` items in the cleanup set:

```bash
oac clean --include-confirm
oac clean --include-confirm --dry-run
oac clean --include-confirm --mode agent --yes --json
```

## Safety Model

The tool classifies discovered paths into three buckets:

- `safe`: automatically eligible for cleanup.
- `confirm`: only eligible when `--include-confirm` is provided.
- `manual`: always excluded from deletion and shown for operator review only.

This is intentionally conservative. For example, OpenClaw workspaces and Ollama SSH keys are not auto-deleted even during a full cleanup flow.

## Discovery Rules

### OpenClaw

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

### IronClaw

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

### Ollama

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

## Scope

- Local macOS filesystem cleanup only.
- No automatic cleanup of external databases, Keychain items, or cloud resources.
- No package manager uninstall integration yet.
- No automatic deletion of user-authored workspace content.

## Development

Format code:

```bash
make fmt
```

Build:

```bash
make build
```

Run tests:

```bash
make test
```

## Contributing

Issues and pull requests are welcome. Read [CONTRIBUTING.md](CONTRIBUTING.md) before sending changes.

## Security

If you find a security issue or a deletion-safety bug, follow [SECURITY.md](SECURITY.md).

## License

This project is released under the MIT License. See [LICENSE](LICENSE).
