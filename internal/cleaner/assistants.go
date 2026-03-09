package cleaner

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func discoverCandidates(assistants []string) ([]Candidate, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	all := make([]Candidate, 0, 32)
	for _, assistant := range assistants {
		switch assistant {
		case "openclaw":
			all = append(all, discoverOpenClaw(home)...)
		case "ironclaw":
			all = append(all, discoverIronClaw(home)...)
		case "ollama":
			all = append(all, discoverOllama(home)...)
		}
	}

	all = dedupeCandidates(all)
	for i := range all {
		all[i].Path = cleanPath(all[i].Path)
		all[i].ID = candidateID(all[i].Assistant, all[i].Kind, all[i].Path)
		all[i].SizeBytes = pathSize(all[i].Path)
		all[i].Deletable = all[i].Safety != SafetyManual
		all[i].RequiresConfirmation = all[i].Safety == SafetyConfirm
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Assistant != all[j].Assistant {
			return all[i].Assistant < all[j].Assistant
		}
		if all[i].Safety != all[j].Safety {
			return all[i].Safety < all[j].Safety
		}
		return all[i].Path < all[j].Path
	})
	return all, nil
}

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

func discoverOllama(home string) []Candidate {
	baseDir := filepath.Join(home, ".ollama")
	modelsDir := cleanPath(os.Getenv("OLLAMA_MODELS"))
	if modelsDir == "." || modelsDir == "" {
		modelsDir = filepath.Join(baseDir, "models")
	}

	out := []Candidate{}
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ollama",
		Path:      filepath.Join(baseDir, "logs"),
		Kind:      "logs",
		Safety:    SafetySafe,
		Reason:    "Ollama app/server logs are disposable and recreated automatically.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ollama",
		Path:      filepath.Join(home, "Library", "Saved Application State", "com.electron.ollama.savedState"),
		Kind:      "saved_state",
		Safety:    SafetySafe,
		Reason:    "macOS saved window state can be safely removed.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ollama",
		Path:      filepath.Join(home, "Library", "Caches", "com.electron.ollama"),
		Kind:      "cache",
		Safety:    SafetySafe,
		Reason:    "Electron cache is disposable.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ollama",
		Path:      filepath.Join(home, "Library", "Caches", "ollama"),
		Kind:      "cache",
		Safety:    SafetySafe,
		Reason:    "App cache is disposable.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ollama",
		Path:      filepath.Join(home, "Library", "WebKit", "com.electron.ollama"),
		Kind:      "webkit_cache",
		Safety:    SafetySafe,
		Reason:    "Embedded WebKit cache is disposable.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ollama",
		Path:      filepath.Join(home, "Library", "Application Support", "Ollama"),
		Kind:      "app_support",
		Safety:    SafetyConfirm,
		Reason:    "Application Support may contain local UI state and account metadata.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ollama",
		Path:      modelsDir,
		Kind:      "models",
		Safety:    SafetyConfirm,
		Reason:    "Model blobs and manifests are large and valuable; require explicit confirmation.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ollama",
		Path:      filepath.Join(baseDir, "server.json"),
		Kind:      "config",
		Safety:    SafetyConfirm,
		Reason:    "Ollama configuration file should only be removed intentionally.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ollama",
		Path:      filepath.Join(baseDir, "id_ed25519"),
		Kind:      "auth_key",
		Safety:    SafetyManual,
		Reason:    "Private key can affect cloud publishing/auth flows and is not auto-removed.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ollama",
		Path:      filepath.Join(baseDir, "id_ed25519.pub"),
		Kind:      "auth_key",
		Safety:    SafetyManual,
		Reason:    "Public key is paired with Ollama auth state; remove manually if needed.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ollama",
		Path:      "/Applications/Ollama.app",
		Kind:      "app_bundle",
		Safety:    SafetyConfirm,
		Reason:    "Installed app bundle. Include this only when you want full product removal, not just leftovers.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ollama",
		Path:      "/usr/local/bin/ollama",
		Kind:      "cli_symlink",
		Safety:    SafetyConfirm,
		Reason:    "CLI entrypoint is part of the installation, not just cached residue.",
	})
	return out
}

func appendCandidateIfExists(dst []Candidate, candidate Candidate) []Candidate {
	if pathExists(candidate.Path) {
		dst = append(dst, candidate)
	}
	return dst
}

func appendGlobCandidates(dst []Candidate, assistant, pattern, kind string, safety Safety, reason string) []Candidate {
	matches, _ := filepath.Glob(pattern)
	for _, match := range matches {
		if pathExists(match) {
			dst = append(dst, Candidate{
				Assistant: assistant,
				Path:      match,
				Kind:      kind,
				Safety:    safety,
				Reason:    reason,
			})
		}
	}
	return dst
}

func pathExists(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	_, err := os.Lstat(path)
	return err == nil
}

func pathSize(path string) int64 {
	info, err := os.Lstat(path)
	if err != nil {
		return 0
	}
	if !info.IsDir() {
		return info.Size()
	}

	var total int64
	filepath.WalkDir(path, func(walkPath string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		total += info.Size()
		return nil
	})
	return total
}

func dedupeCandidates(candidates []Candidate) []Candidate {
	seen := map[string]struct{}{}
	out := make([]Candidate, 0, len(candidates))
	for _, candidate := range candidates {
		key := candidate.Assistant + "\x00" + cleanPath(candidate.Path)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, candidate)
	}
	return out
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
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
