# Task Plan: Install `oac` and inspect local cleanup candidates

## Goal

- Verify the public install path for the `oac` CLI on the local machine
- Run non-destructive local cleanup scans and inspect what the tool recommends on this machine
- Report findings to the user before any deletion step

## Touched Files

- `docs/current-plan.md`
- `docs/current-plans/install-oac-ollama-check.md`

## Todo

- [x] Verify install path behavior and confirm the binary is runnable from the shell
- [x] Run non-destructive `oac` scan commands for general local cleanup findings
- [x] Summarize removable vs. manual-review items for the user without deleting anything
- [x] Update this plan file with validation status and completion notes
- [x] Update `docs/current-plan.md` to mark the task complete when finished

## Validation

- Verify `oac --help` or equivalent command exits successfully
- Verify the install command path or existing binary state can be confirmed from the shell
- Verify one or more scan commands exit successfully and produce cleanup results or an explicit no-findings result

## Doc Sync

- Update `docs/current-plan.md` status for this task
- Update this file with outcome notes and any unverified items

## Outcome Notes

- Follow-up task completed on 2026-03-09 for broader local cleanup verification
- Existing binary was present at `/Users/bytedance/.local/bin/oac`; install script upgraded it from `0.1.3` to `0.2.0`
- Verified CLI health with `oac --help`
- Ran `oac scan --mode agent --json` successfully
- Ran `oac clean --safety safe --dry-run --mode agent --json` successfully
- Scan found 13 candidates totaling `81897182` bytes across Ollama and OpenClaw residues
- `safe` candidates: 5 items, `6973421` bytes
  - `/Users/bytedance/.ollama/logs`
  - `/Users/bytedance/Library/WebKit/com.electron.ollama`
  - `/Users/bytedance/.openclaw/logs`
  - `/tmp/openclaw/openclaw-2026-03-06.log`
  - `/tmp/openclaw/openclaw-2026-03-07.log`
- `confirm` candidates: 5 items, `240166` bytes
  - `/Users/bytedance/Library/Application Support/Ollama`
  - `/usr/local/bin/ollama`
  - `/Users/bytedance/.openclaw/agents`
  - `/Users/bytedance/.openclaw/openclaw.json`
  - `/Users/bytedance/Library/LaunchAgents/ai.openclaw.gateway.plist`
- `manual` candidates: 3 items, `74683595` bytes
  - `/Users/bytedance/.ollama/id_ed25519`
  - `/Users/bytedance/.ollama/id_ed25519.pub`
  - `/Users/bytedance/.openclaw/workspace`

## Unverified

- Did not execute deletion commands
- Did not verify whether the user still needs Ollama installation state, OpenClaw sessions/config, or the OpenClaw workspace contents
