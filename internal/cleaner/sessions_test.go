package cleaner

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodexConversationSessionsAndDelete(t *testing.T) {
	home := setTestHome(t)
	root := filepath.Join(home, ".codex")
	desktopRollout := filepath.Join(root, "sessions", "2026", "03", "09", "desktop.jsonl")
	cliRollout := filepath.Join(root, "sessions", "2026", "03", "09", "cli.jsonl")
	indexPath := filepath.Join(root, "session_index.jsonl")
	dbPath := filepath.Join(root, "state_5.sqlite")

	writeTestFile(t, desktopRollout, strings.Join([]string{
		`{"type":"session_meta","payload":{"timestamp":"2026-03-09T10:00:00Z","originator":"Codex Desktop"}}`,
		`{"type":"response_item","payload":{"type":"message","role":"assistant","content":[{"type":"output_text","text":"Desktop answer"}]}}`,
	}, "\n")+"\n")
	writeTestFile(t, cliRollout, strings.Join([]string{
		`{"type":"session_meta","payload":{"timestamp":"2026-03-09T11:00:00Z","originator":"codex_cli_rs"}}`,
		`{"type":"response_item","payload":{"type":"message","role":"assistant","content":[{"type":"output_text","text":"CLI answer"}]}}`,
	}, "\n")+"\n")
	writeTestFile(t, indexPath, strings.Join([]string{
		`{"id":"desktop-thread","thread_name":"Desktop thread","updated_at":"2026-03-09T10:00:00Z"}`,
		`{"id":"cli-thread","thread_name":"CLI thread","updated_at":"2026-03-09T11:00:00Z"}`,
	}, "\n")+"\n")

	db := createSQLiteDB(t, dbPath)
	mustExec(t, db, `CREATE TABLE threads (
		id TEXT PRIMARY KEY,
		rollout_path TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		source TEXT NOT NULL,
		title TEXT NOT NULL,
		tokens_used INTEGER NOT NULL DEFAULT 0,
		first_user_message TEXT NOT NULL DEFAULT ''
	)`)
	mustExec(t, db, `CREATE TABLE logs (id INTEGER PRIMARY KEY AUTOINCREMENT, thread_id TEXT)`)
	mustExec(t, db, `INSERT INTO threads (id, rollout_path, created_at, updated_at, source, title, tokens_used, first_user_message)
		VALUES (?, ?, 1741514400000, 1741514400000, 'vscode', 'Desktop thread', 10, 'desktop prompt')`,
		"desktop-thread", desktopRollout)
	mustExec(t, db, `INSERT INTO threads (id, rollout_path, created_at, updated_at, source, title, tokens_used, first_user_message)
		VALUES (?, ?, 1741518000000, 1741518000000, 'cli', 'CLI thread', 20, 'cli prompt')`,
		"cli-thread", cliRollout)
	mustExec(t, db, `INSERT INTO logs (thread_id) VALUES ('cli-thread')`)
	db.Close()

	desktopSessions, err := discoverAssistantSessions("codex")
	if err != nil {
		t.Fatalf("discoverAssistantSessions(codex) err = %v", err)
	}
	if len(desktopSessions) != 1 || desktopSessions[0].Title != "Desktop thread" {
		t.Fatalf("desktop sessions = %#v", desktopSessions)
	}

	cliSessions, err := discoverAssistantSessions("codex-cli")
	if err != nil {
		t.Fatalf("discoverAssistantSessions(codex-cli) err = %v", err)
	}
	if len(cliSessions) != 1 || cliSessions[0].Title != "CLI thread" {
		t.Fatalf("cli sessions = %#v", cliSessions)
	}
	preview, err := previewConversationSession(cliSessions[0])
	if err != nil || !strings.Contains(preview, "CLI answer") {
		t.Fatalf("previewConversationSession(codex-cli) = %q, %v", preview, err)
	}

	if err := deleteConversationSessions(cliSessions); err != nil {
		t.Fatalf("deleteConversationSessions(codex-cli) err = %v", err)
	}
	if pathExists(cliRollout) {
		t.Fatal("cli rollout should be deleted")
	}

	db = createSQLiteDB(t, dbPath)
	defer db.Close()
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM threads WHERE id = 'cli-thread'`).Scan(&count); err != nil {
		t.Fatalf("count cli thread: %v", err)
	}
	if count != 0 {
		t.Fatalf("cli thread still exists in DB: %d", count)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM logs WHERE thread_id = 'cli-thread'`).Scan(&count); err != nil {
		t.Fatalf("count cli logs: %v", err)
	}
	if count != 0 {
		t.Fatalf("cli logs still exist in DB: %d", count)
	}
	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("ReadFile(index) err = %v", err)
	}
	if strings.Contains(string(indexData), "cli-thread") {
		t.Fatalf("session index still contains deleted thread: %s", indexData)
	}
}

func TestClassifyCodexAssistantFromSource(t *testing.T) {
	tests := []struct {
		source string
		want   string
		ok     bool
	}{
		{source: "vscode", want: "codex", ok: true},
		{source: "cli", want: "codex-cli", ok: true},
		{source: "exec", want: "codex-cli", ok: true},
		{source: "mcp", want: "codex-cli", ok: true},
		{source: "unknown", want: "", ok: false},
		{source: "", want: "", ok: false},
	}

	for _, tt := range tests {
		got, ok := classifyCodexAssistantFromSource(tt.source)
		if got != tt.want || ok != tt.ok {
			t.Fatalf("classifyCodexAssistantFromSource(%q) = %q, %t", tt.source, got, ok)
		}
	}
}

func TestCodexConversationSessionsFallbackToRolloutForUnknownSource(t *testing.T) {
	home := setTestHome(t)
	root := filepath.Join(home, ".codex")
	desktopRollout := filepath.Join(root, "sessions", "2026", "03", "09", "desktop-unknown.jsonl")
	writeTestFile(t, desktopRollout, strings.Join([]string{
		`{"type":"session_meta","payload":{"timestamp":"2026-03-09T10:00:00Z","originator":"Codex Desktop"}}`,
		`{"type":"response_item","payload":{"type":"message","role":"assistant","content":[{"type":"output_text","text":"Desktop answer"}]}}`,
	}, "\n")+"\n")

	dbPath := filepath.Join(root, "state_5.sqlite")
	db := createSQLiteDB(t, dbPath)
	mustExec(t, db, `CREATE TABLE threads (
		id TEXT PRIMARY KEY,
		rollout_path TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		source TEXT NOT NULL,
		title TEXT NOT NULL,
		tokens_used INTEGER NOT NULL DEFAULT 0,
		first_user_message TEXT NOT NULL DEFAULT ''
	)`)
	mustExec(t, db, `INSERT INTO threads (id, rollout_path, created_at, updated_at, source, title, tokens_used, first_user_message)
		VALUES (?, ?, 1741514400000, 1741514400000, 'unknown', 'Desktop thread', 10, 'desktop prompt')`,
		"desktop-thread", desktopRollout)
	db.Close()

	sessions, err := discoverAssistantSessions("codex")
	if err != nil {
		t.Fatalf("discoverAssistantSessions(codex unknown) err = %v", err)
	}
	if len(sessions) != 1 || sessions[0].Title != "Desktop thread" {
		t.Fatalf("codex unknown-source sessions = %#v", sessions)
	}
}

func TestClaudeCodeConversationSessionsAndDelete(t *testing.T) {
	home := setTestHome(t)
	projectDir := filepath.Join(home, ".claude", "projects", "project-a")
	projectPath := filepath.Join(projectDir, "session-a.jsonl")
	globalPath := filepath.Join(home, ".claude", "transcripts", "session-b.jsonl")
	indexPath := filepath.Join(projectDir, "sessions-index.json")

	content := strings.Join([]string{
		`{"type":"user","timestamp":"2026-03-09T10:00:00Z","sessionId":"session-a","message":{"role":"user","content":[{"type":"text","text":"Need a plan"}]}}`,
		`{"type":"assistant","timestamp":"2026-03-09T10:01:00Z","sessionId":"session-a","message":{"role":"assistant","content":[{"type":"text","text":"Here is the plan"}]}}`,
	}, "\n") + "\n"
	writeTestFile(t, projectPath, content)
	writeTestFile(t, globalPath, content)
	writeTestFile(t, indexPath, `{"version":1,"entries":[{"sessionId":"session-a","path":"session-a.jsonl"},{"sessionId":"keep","path":"keep.jsonl"}]}`+"\n")

	sessions, err := discoverAssistantSessions("claudecode")
	if err != nil {
		t.Fatalf("discoverAssistantSessions(claudecode) err = %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("claude sessions len = %d", len(sessions))
	}

	preview, err := previewConversationSession(sessions[0])
	if err != nil || !strings.Contains(preview, "Need a plan") {
		t.Fatalf("previewConversationSession(claudecode) = %q, %v", preview, err)
	}

	var projectSession ConversationSession
	for _, session := range sessions {
		if strings.Contains(session.Path, "project-a") {
			projectSession = session
			break
		}
	}
	if projectSession.ID == "" {
		t.Fatalf("project session not found: %#v", sessions)
	}
	if err := deleteConversationSessions([]ConversationSession{projectSession}); err != nil {
		t.Fatalf("deleteConversationSessions(claudecode) err = %v", err)
	}
	if pathExists(projectPath) {
		t.Fatal("project transcript should be deleted")
	}
	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("ReadFile(project index) err = %v", err)
	}
	if strings.Contains(string(indexData), "session-a") {
		t.Fatalf("project index still contains deleted session: %s", indexData)
	}
}

func TestClaudeCodeConversationDeleteRollsBackOnIndexFailure(t *testing.T) {
	home := setTestHome(t)
	projectDir := filepath.Join(home, ".claude", "projects", "project-a")
	projectPath := filepath.Join(projectDir, "session-a.jsonl")
	indexPath := filepath.Join(projectDir, "sessions-index.json")

	writeTestFile(t, projectPath, `{"type":"user","timestamp":"2026-03-09T10:00:00Z","sessionId":"session-a","message":{"role":"user","content":[{"type":"text","text":"Need a plan"}]}}`+"\n")
	writeTestFile(t, indexPath, `{"version":1,"entries":[{"sessionId":"session-a","path":"session-a.jsonl"}]}`+"\n")

	sessions, err := discoverAssistantSessions("claudecode")
	if err != nil {
		t.Fatalf("discoverAssistantSessions(claudecode) err = %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("claude sessions len = %d", len(sessions))
	}

	if err := os.Remove(indexPath); err != nil {
		t.Fatalf("Remove(indexPath) err = %v", err)
	}
	if err := os.MkdirAll(indexPath, 0o755); err != nil {
		t.Fatalf("MkdirAll(indexPath as dir) err = %v", err)
	}
	if err := deleteConversationSessions([]ConversationSession{sessions[0]}); err == nil {
		t.Fatal("deleteConversationSessions(claudecode) should fail when sessions-index.json is invalid")
	}
	if !pathExists(projectPath) {
		t.Fatal("project transcript should be restored when index update fails")
	}
}

func TestCursorConversationSessionsAndDelete(t *testing.T) {
	home := setTestHome(t)
	dbPath := filepath.Join(home, "Library", "Application Support", "Cursor", "User", "globalStorage", "state.vscdb")
	db := createSQLiteDB(t, dbPath)
	mustExec(t, db, `CREATE TABLE cursorDiskKV (key TEXT PRIMARY KEY, value BLOB)`)

	composer := map[string]any{
		"composerId":    "composer-1",
		"name":          "Fix cursor tests",
		"subtitle":      "pkg/session",
		"createdAt":     1741514400000,
		"lastUpdatedAt": 1741518000000,
		"isArchived":    false,
		"fullConversationHeadersOnly": []map[string]any{
			{"bubbleId": "bubble-user", "type": 1},
			{"bubbleId": "bubble-assistant", "type": 2},
		},
	}
	insertJSONRow(t, db, "composerData:composer-1", composer)
	insertJSONRow(t, db, "bubbleId:composer-1:bubble-user", map[string]any{"type": 1, "text": "Please fix the failing tests"})
	insertJSONRow(t, db, "bubbleId:composer-1:bubble-assistant", map[string]any{"type": 2, "text": "I will inspect the failing tests first"})
	insertJSONRow(t, db, "messageRequestContext:composer-1:bubble-user", map[string]any{"request": "ctx"})
	db.Close()

	sessions, err := discoverAssistantSessions("cursor")
	if err != nil {
		t.Fatalf("discoverAssistantSessions(cursor) err = %v", err)
	}
	if len(sessions) != 1 || sessions[0].Title != "Fix cursor tests" {
		t.Fatalf("cursor sessions = %#v", sessions)
	}
	preview, err := previewConversationSession(sessions[0])
	if err != nil || !strings.Contains(preview, "Please fix the failing tests") || !strings.Contains(preview, "I will inspect the failing tests first") {
		t.Fatalf("previewConversationSession(cursor) = %q, %v", preview, err)
	}

	if err := deleteConversationSessions(sessions); err != nil {
		t.Fatalf("deleteConversationSessions(cursor) err = %v", err)
	}
	db = createSQLiteDB(t, dbPath)
	defer db.Close()
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM cursorDiskKV`).Scan(&count); err != nil {
		t.Fatalf("count cursor rows: %v", err)
	}
	if count != 0 {
		t.Fatalf("cursor session rows still exist: %d", count)
	}
}

func TestCursorConversationSessionsIgnoreNullSQLiteValues(t *testing.T) {
	home := setTestHome(t)
	dbPath := filepath.Join(home, "Library", "Application Support", "Cursor", "User", "globalStorage", "state.vscdb")
	db := createSQLiteDB(t, dbPath)
	mustExec(t, db, `CREATE TABLE cursorDiskKV (key TEXT PRIMARY KEY, value BLOB)`)
	mustExec(t, db, `INSERT INTO cursorDiskKV (key, value) VALUES (?, NULL)`, "composerData:null-session")
	insertJSONRow(t, db, "composerData:composer-1", map[string]any{
		"composerId":    "composer-1",
		"name":          "Keep good rows",
		"createdAt":     1741514400000,
		"lastUpdatedAt": 1741518000000,
		"isArchived":    false,
		"fullConversationHeadersOnly": []map[string]any{
			{"bubbleId": "bubble-user", "type": 1},
			{"bubbleId": "bubble-null", "type": 2},
		},
	})
	insertJSONRow(t, db, "bubbleId:composer-1:bubble-user", map[string]any{"type": 1, "text": "hello"})
	mustExec(t, db, `INSERT INTO cursorDiskKV (key, value) VALUES (?, NULL)`, "bubbleId:composer-1:bubble-null")
	db.Close()

	sessions, err := discoverAssistantSessions("cursor")
	if err != nil {
		t.Fatalf("discoverAssistantSessions(cursor) with NULL values err = %v", err)
	}
	if len(sessions) != 1 || sessions[0].Title != "Keep good rows" {
		t.Fatalf("cursor sessions with NULL values = %#v", sessions)
	}
	preview, err := previewConversationSession(sessions[0])
	if err != nil {
		t.Fatalf("previewConversationSession(cursor) with NULL values err = %v", err)
	}
	if !strings.Contains(preview, "hello") {
		t.Fatalf("previewConversationSession(cursor) with NULL values = %q", preview)
	}
}

func TestAntigravityConversationSessionsPreviewOnly(t *testing.T) {
	home := setTestHome(t)
	taskDir := filepath.Join(home, ".gemini", "antigravity", "brain", "task-1")
	taskPath := filepath.Join(taskDir, "task.md")
	writeTestFile(t, taskPath, "# Refactor Analyze TUI\n\nKeep the session browser generic.\n")
	writeTestFile(t, filepath.Join(taskDir, "task.md.metadata.json"), `{"summary":"Refactor plan","updatedAt":"2026-03-09T12:00:00Z"}`)
	writeTestBytes(t, filepath.Join(home, ".gemini", "antigravity", "conversations", "task-1.pb"), []byte("pb"))

	sessions, err := discoverAssistantSessions("antigravity")
	if err != nil {
		t.Fatalf("discoverAssistantSessions(antigravity) err = %v", err)
	}
	if len(sessions) != 1 || sessions[0].Deletable {
		t.Fatalf("antigravity sessions = %#v", sessions)
	}
	preview, err := previewConversationSession(sessions[0])
	if err != nil || !strings.Contains(preview, "Refactor Analyze TUI") {
		t.Fatalf("previewConversationSession(antigravity) = %q, %v", preview, err)
	}
}

func createSQLiteDB(t *testing.T, path string) *sql.DB {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s) err = %v", filepath.Dir(path), err)
	}
	db, err := openSQLiteDB(path)
	if err != nil {
		t.Fatalf("openSQLiteDB(%s) err = %v", path, err)
	}
	return db
}

func mustExec(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatalf("Exec(%s) err = %v", query, err)
	}
}

func insertJSONRow(t *testing.T, db *sql.DB, key string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("Marshal(%s) err = %v", key, err)
	}
	mustExec(t, db, `INSERT INTO cursorDiskKV (key, value) VALUES (?, ?)`, key, string(data))
}
