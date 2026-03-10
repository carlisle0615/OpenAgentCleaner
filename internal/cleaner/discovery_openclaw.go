package cleaner

import (
	"os"
	"path/filepath"
)

func discoverOpenClaw(home string) []Candidate {
	var out []Candidate
	for _, root := range openClawStateRoots(home) {
		configPath := filepath.Join(root, "openclaw.json")
		if explicitConfig := cleanPath(os.Getenv("OPENCLAW_CONFIG_PATH")); explicitConfig != "." && explicitConfig != "" {
			configPath = explicitConfig
		}

		out = appendCandidateIfExists(out, Candidate{
			Assistant: "openclaw",
			Path:      filepath.Join(root, "logs"),
			Kind:      "gateway_logs",
			Safety:    SafetySafe,
			Reason:    "OpenClaw runtime logs and trace files can be regenerated.",
		})
		out = appendGlobCandidates(out, "openclaw", filepath.Join("/tmp", "openclaw", "*.log*"), "launchd_logs", SafetySafe, "launchd stdout/stderr logs are disposable.")
		out = appendCandidateIfExists(out, Candidate{
			Assistant: "openclaw",
			Path:      filepath.Join(root, "agents"),
			Kind:      "session_store",
			Safety:    SafetyConfirm,
			Reason:    "Contains session store and transcript JSONL files under agents/<agent>/sessions.",
			Notes:     []string{"OpenClaw docs store sessions at ~/.openclaw/agents/<agentId>/sessions/."},
		})
		out = appendCandidateIfExists(out, Candidate{
			Assistant: "openclaw",
			Path:      configPath,
			Kind:      "config",
			Safety:    SafetyConfirm,
			Reason:    "Main config may contain gateway secrets, provider settings, and channel credentials.",
		})
		out = appendCandidateIfExists(out, Candidate{
			Assistant: "openclaw",
			Path:      filepath.Join(root, ".env"),
			Kind:      "env_file",
			Safety:    SafetyConfirm,
			Reason:    "Service env file may contain API keys used by launchd or CLI flows.",
		})
		out = appendCandidateIfExists(out, Candidate{
			Assistant: "openclaw",
			Path:      filepath.Join(root, "extensions"),
			Kind:      "extensions",
			Safety:    SafetyConfirm,
			Reason:    "Installed extensions/plugins belong to OpenClaw state, but deletion is irreversible.",
		})
		out = appendCandidateIfExists(out, Candidate{
			Assistant: "openclaw",
			Path:      filepath.Join(root, "workspace"),
			Kind:      "workspace",
			Safety:    SafetyManual,
			Reason:    "Workspace contains user-authored memory, skills, AGENTS.md, and other files that should not be auto-removed.",
			Notes:     []string{"Delete manually only if you want to wipe agent-authored content too."},
		})
	}

	out = appendGlobCandidates(out, "openclaw", filepath.Join(home, "Library", "LaunchAgents", "ai.openclaw*.plist"), "service_plist", SafetyConfirm, "OpenClaw launchd service plist may keep the gateway alive.")
	out = appendGlobCandidates(out, "openclaw", filepath.Join(home, "Library", "LaunchAgents", "bot.molt*.plist"), "service_plist", SafetyConfirm, "Newer macOS app installs use bot.molt launchd labels.")
	out = appendGlobCandidates(out, "openclaw", filepath.Join(home, "Library", "LaunchAgents", "com.openclaw*.plist"), "legacy_service_plist", SafetyConfirm, "Legacy launchd plists are safe to remove after uninstall.")
	return out
}

func isOpenClawStateRoot(path string) bool {
	if !pathExists(path) {
		return false
	}
	markers := []string{
		filepath.Join(path, "openclaw.json"),
		filepath.Join(path, "agents"),
		filepath.Join(path, "workspace"),
	}
	for _, marker := range markers {
		if pathExists(marker) {
			return true
		}
	}
	return false
}

func openClawStateRoots(home string) []string {
	stateRoots := []string{}
	if explicit := cleanPath(os.Getenv("OPENCLAW_STATE_DIR")); explicit != "." && explicit != "" {
		stateRoots = append(stateRoots, explicit)
	} else {
		stateRoots = append(stateRoots, filepath.Join(home, ".openclaw"))
	}

	if matches, _ := filepath.Glob(filepath.Join(home, ".openclaw-*")); len(matches) > 0 {
		for _, match := range matches {
			if isOpenClawStateRoot(match) {
				stateRoots = append(stateRoots, match)
			}
		}
	}

	return uniqueStrings(stateRoots)
}
