package discoveryrules

import "path/filepath"

func DiscoverCodexDesktop(home string) []Candidate {
	out := []Candidate{}
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "codex",
		Path:      filepath.Join(home, "Library", "Application Support", "Codex", "Session Storage"),
		Kind:      "desktop_session_storage",
		Safety:    SafetyConfirm,
		Reason:    "Codex desktop app keeps browser-style session state here in addition to shared rollout storage under ~/.codex.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "codex",
		Path:      filepath.Join(home, "Library", "Application Support", "Codex", "Local Storage"),
		Kind:      "desktop_local_storage",
		Safety:    SafetyConfirm,
		Reason:    "Codex desktop app stores recent thread and UI state in local browser storage.",
	})
	return out
}

func DiscoverCodexCLI(home string) []Candidate {
	root := filepath.Join(home, ".codex")
	out := []Candidate{}
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "codex-cli",
		Path:      filepath.Join(root, "sessions"),
		Kind:      "session_store",
		Safety:    SafetyConfirm,
		Reason:    "Codex CLI stores active conversation rollouts as JSONL files under date-based session folders.",
		Notes:     []string{"Local observation: ~/.codex/sessions/YYYY/MM/DD/rollout-*.jsonl"},
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "codex-cli",
		Path:      filepath.Join(root, "archived_sessions"),
		Kind:      "archived_sessions",
		Safety:    SafetyConfirm,
		Reason:    "Archived Codex conversations remain readable locally until you delete the archived rollout files.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "codex-cli",
		Path:      filepath.Join(root, "session_index.jsonl"),
		Kind:      "session_index",
		Safety:    SafetyConfirm,
		Reason:    "Session index maps thread IDs and titles to on-disk rollout files.",
	})
	out = appendGlobCandidates(out, "codex-cli", filepath.Join(root, "state_*.sqlite*"), "session_db", SafetyConfirm, "SQLite metadata database tracks Codex threads, rollouts, and archive status.")
	return out
}

func DiscoverClaudeCode(home string) []Candidate {
	root := filepath.Join(home, ".claude")
	out := []Candidate{}
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "claudecode",
		Path:      filepath.Join(root, "transcripts"),
		Kind:      "transcripts",
		Safety:    SafetyConfirm,
		Reason:    "Claude Code stores conversation transcripts as JSONL files in the transcripts folder.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "claudecode",
		Path:      filepath.Join(root, "projects"),
		Kind:      "project_sessions",
		Safety:    SafetyConfirm,
		Reason:    "Per-project conversation files and session indexes live under ~/.claude/projects.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "claudecode",
		Path:      filepath.Join(root, "history.jsonl"),
		Kind:      "prompt_history",
		Safety:    SafetyConfirm,
		Reason:    "Prompt history contains past inputs tied to Claude Code sessions.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "claudecode",
		Path:      filepath.Join(home, "Library", "Application Support", "Claude", "Session Storage"),
		Kind:      "desktop_session_storage",
		Safety:    SafetyConfirm,
		Reason:    "Claude desktop app stores browser session state here.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "claudecode",
		Path:      filepath.Join(home, "Library", "Application Support", "Claude", "Local Storage"),
		Kind:      "desktop_local_storage",
		Safety:    SafetyConfirm,
		Reason:    "Claude desktop app stores recent thread and UI state in local browser storage.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "claudecode",
		Path:      filepath.Join(home, "Library", "Application Support", "Claude", "IndexedDB"),
		Kind:      "desktop_indexeddb",
		Safety:    SafetyConfirm,
		Reason:    "Claude desktop app persists web session data and thread metadata in IndexedDB.",
	})
	return out
}

func DiscoverCursor(home string) []Candidate {
	appSupport := filepath.Join(home, "Library", "Application Support", "Cursor", "User")
	out := []Candidate{}
	out = appendGlobCandidates(out, "cursor", filepath.Join(appSupport, "globalStorage", "state.vscdb*"), "global_chat_state", SafetyConfirm, "Cursor keeps global composer/chat state in VS Code SQLite storage.")
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "cursor",
		Path:      filepath.Join(appSupport, "workspaceStorage"),
		Kind:      "workspace_chat_state",
		Safety:    SafetyConfirm,
		Reason:    "Workspace-scoped Cursor composer/chat state lives in workspaceStorage alongside other workspace metadata.",
		Notes:     []string{"Local observation: per-workspace state.vscdb files under ~/Library/Application Support/Cursor/User/workspaceStorage/"},
	})
	return out
}

func DiscoverAntigravity(home string) []Candidate {
	appSupport := filepath.Join(home, "Library", "Application Support", "Antigravity", "User")
	out := []Candidate{}
	out = appendGlobCandidates(out, "antigravity", filepath.Join(appSupport, "globalStorage", "state.vscdb*"), "global_chat_state", SafetyConfirm, "Antigravity keeps global chat state in VS Code SQLite storage.")
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "antigravity",
		Path:      filepath.Join(appSupport, "workspaceStorage"),
		Kind:      "workspace_chat_state",
		Safety:    SafetyConfirm,
		Reason:    "Workspace-scoped Antigravity chat state lives in workspaceStorage alongside other workspace metadata.",
		Notes:     []string{"Local observation: chat.ChatSessionStore keys live in the VS Code state database."},
	})
	return out
}
