# Task Plan: Install `oac` and inspect Ollama cleanup candidates

## Goal

- Install the `oac` CLI on the local machine using the public project distribution path
- Run a non-destructive scan focused on local cleanup guidance and determine whether Ollama has removable items
- Report findings to the user before any deletion step

## Touched Files

- `docs/current-plan.md`
- `docs/current-plans/install-oac-ollama-check.md`

## Todo

- [x] Install `oac` and confirm the binary is runnable from the shell
- [x] Run a non-destructive `oac` scan and capture Ollama-related findings
- [x] Summarize removable vs. manual-review items for the user without deleting anything
- [x] Update this plan file with validation status and completion notes
- [x] Update `docs/current-plan.md` to mark the task complete when finished

## Validation

- Verify `oac --help` or equivalent command exits successfully
- Verify the scan command exits successfully and produces Ollama-related results or an explicit no-findings result

## Doc Sync

- Update `docs/current-plan.md` status for this task
- Update this file with outcome notes and any unverified items

## Outcome Notes

- `oac` was already installed at `/Users/bytedance/.local/bin/oac`
- Verified CLI health with `oac --version` and `oac --help`; installed version is `0.1.3`
- Ran `oac scan --assistants ollama --mode agent --json` successfully
- Scan found 7 Ollama-related candidates totaling `19006424703` bytes
- Safe cleanup candidates:
  - `/Users/bytedance/.ollama/logs` (`17786` bytes)
  - `/Users/bytedance/Library/WebKit/com.electron.ollama` (`618920` bytes)
- Confirm-only candidates:
  - `/Users/bytedance/.ollama/models` (`19005552627` bytes)
  - `/Users/bytedance/Library/Application Support/Ollama` (`234852` bytes)
  - `/usr/local/bin/ollama` (`50` bytes)
- Manual-review candidates:
  - `/Users/bytedance/.ollama/id_ed25519`
  - `/Users/bytedance/.ollama/id_ed25519.pub`
- Local manifest inspection shows two installed models:
  - `registry.ollama.ai/library/gpt-oss/20b` layer size `13780154624` bytes
  - `registry.ollama.ai/library/deepseek-r1/8b` layer size `5225373760` bytes

## Unverified

- Did not execute deletion commands
- Did not verify whether the user still actively needs the two installed models or Ollama application state
