package cleaner

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscoverSessionAssistants(t *testing.T) {
	home := setTestHome(t)

	for _, path := range []string{
		".codex/sessions/2026/03/09/rollout-1.jsonl",
		".codex/archived_sessions/rollout-archived.jsonl",
		".codex/session_index.jsonl",
		".codex/state_5.sqlite",
		".codex/state_5.sqlite-wal",
		"Library/Application Support/Codex/Session Storage/000003.log",
		"Library/Application Support/Codex/Local Storage/leveldb/000005.ldb",
		".claude/transcripts/ses_one.jsonl",
		".claude/projects/project-a/session-a.jsonl",
		".claude/projects/project-a/sessions-index.json",
		".claude/history.jsonl",
		"Library/Application Support/Claude/Session Storage/000004.log",
		"Library/Application Support/Claude/Local Storage/leveldb/000005.ldb",
		"Library/Application Support/Claude/IndexedDB/https_claude.ai_0.indexeddb.leveldb/000003.log",
		"Library/Application Support/Cursor/User/globalStorage/state.vscdb",
		"Library/Application Support/Cursor/User/globalStorage/state.vscdb-wal",
		"Library/Application Support/Cursor/User/workspaceStorage/abc/state.vscdb",
		"Library/Application Support/Antigravity/User/globalStorage/state.vscdb",
		"Library/Application Support/Antigravity/User/workspaceStorage/def/state.vscdb",
	} {
		writeTestFile(t, filepath.Join(home, path), "x")
	}

	all, err := discoverCandidates([]string{"codex", "codex-cli", "claudecode", "cursor", "antigravity"})
	if err != nil {
		t.Fatalf("discoverCandidates() err = %v", err)
	}
	if len(all) == 0 {
		t.Fatal("discoverCandidates() returned no candidates")
	}

	joined := make([]string, 0, len(all))
	for _, candidate := range all {
		joined = append(joined, candidate.Assistant+":"+candidate.Kind+":"+string(candidate.Safety))
	}
	got := strings.Join(joined, "\n")
	for _, want := range []string{
		"codex-cli:session_store:confirm",
		"codex-cli:archived_sessions:confirm",
		"codex-cli:session_index:confirm",
		"codex-cli:session_db:confirm",
		"codex:desktop_session_storage:confirm",
		"claudecode:transcripts:confirm",
		"claudecode:project_sessions:confirm",
		"claudecode:prompt_history:confirm",
		"claudecode:desktop_indexeddb:confirm",
		"cursor:global_chat_state:confirm",
		"cursor:workspace_chat_state:confirm",
		"antigravity:global_chat_state:confirm",
		"antigravity:workspace_chat_state:confirm",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("discoverCandidates() missing %q\n%s", want, got)
		}
	}
}
