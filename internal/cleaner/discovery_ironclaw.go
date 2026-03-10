package cleaner

import (
	"os"
	"path/filepath"
)

func discoverIronClaw(home string) []Candidate {
	baseDir := cleanPath(os.Getenv("IRONCLAW_BASE_DIR"))
	if baseDir == "." || baseDir == "" {
		baseDir = filepath.Join(home, ".ironclaw")
	}

	out := []Candidate{}
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ironclaw",
		Path:      filepath.Join(baseDir, "logs"),
		Kind:      "daemon_logs",
		Safety:    SafetySafe,
		Reason:    "Service stdout/stderr logs are disposable and recreated on next start.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ironclaw",
		Path:      filepath.Join(baseDir, ".env"),
		Kind:      "env_file",
		Safety:    SafetyConfirm,
		Reason:    "Bootstrap env file may contain DATABASE_URL, API keys, and backend settings.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ironclaw",
		Path:      filepath.Join(baseDir, "ironclaw.db"),
		Kind:      "local_database",
		Safety:    SafetyConfirm,
		Reason:    "Embedded libSQL/SQLite database contains chat history and settings.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ironclaw",
		Path:      filepath.Join(baseDir, "config.toml"),
		Kind:      "config",
		Safety:    SafetyConfirm,
		Reason:    "Persistent config file used by IronClaw CLI features.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ironclaw",
		Path:      filepath.Join(baseDir, "session.json"),
		Kind:      "oauth_session",
		Safety:    SafetyConfirm,
		Reason:    "OAuth session token file is sensitive and should only be removed with explicit confirmation.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ironclaw",
		Path:      filepath.Join(baseDir, "mcp-servers.json"),
		Kind:      "mcp_config",
		Safety:    SafetyConfirm,
		Reason:    "MCP server definitions are persistent integration settings.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ironclaw",
		Path:      filepath.Join(baseDir, "settings.json"),
		Kind:      "legacy_settings",
		Safety:    SafetyConfirm,
		Reason:    "Legacy settings file may still contain persistent state awaiting migration.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ironclaw",
		Path:      filepath.Join(baseDir, "bootstrap.json"),
		Kind:      "legacy_bootstrap",
		Safety:    SafetyConfirm,
		Reason:    "Legacy bootstrap file may contain database connection details.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ironclaw",
		Path:      filepath.Join(baseDir, "channels"),
		Kind:      "channels",
		Safety:    SafetyConfirm,
		Reason:    "Installed WASM channels live here and are removed by a full cleanup.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ironclaw",
		Path:      filepath.Join(baseDir, "tools"),
		Kind:      "tools",
		Safety:    SafetyConfirm,
		Reason:    "Installed tools live here and should only be removed intentionally.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ironclaw",
		Path:      filepath.Join(baseDir, "history"),
		Kind:      "repl_history",
		Safety:    SafetyConfirm,
		Reason:    "REPL history may contain prompts or secrets.",
	})
	out = appendGlobCandidates(out, "ironclaw", filepath.Join(baseDir, "*-pairing.json"), "pairing_store", SafetyConfirm, "Pairing approvals and pending requests are persistent user state.")
	out = appendGlobCandidates(out, "ironclaw", filepath.Join(baseDir, "*-allowFrom.json"), "allow_from", SafetyConfirm, "Channel allow-lists are persistent access control state.")
	out = appendGlobCandidates(out, "ironclaw", filepath.Join(baseDir, "*-approve-attempts.json"), "pairing_attempts", SafetyConfirm, "Pairing rate-limit files are safe to remove only with full state cleanup.")
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ironclaw",
		Path:      filepath.Join(home, "Library", "LaunchAgents", "com.ironclaw.daemon.plist"),
		Kind:      "service_plist",
		Safety:    SafetyConfirm,
		Reason:    "launchd plist keeps the IronClaw daemon registered on macOS.",
	})
	return out
}
