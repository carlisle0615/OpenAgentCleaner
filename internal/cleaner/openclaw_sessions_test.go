package cleaner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestOpenClawSessionsLifecycle(t *testing.T) {
	home := setTestHome(t)
	root := filepath.Join(home, ".openclaw")
	sessionsDir := filepath.Join(root, "agents", "main", "sessions")
	metadataPath := filepath.Join(sessionsDir, "sessions.json")
	transcriptOne := filepath.Join(sessionsDir, "session-1.jsonl")
	transcriptTwo := filepath.Join(sessionsDir, "session-2.jsonl")

	writeTestFile(t, metadataPath, `{
  "one": {
    "sessionId": "session-1",
    "displayName": "Debug crash",
    "updatedAt": 1700000000000,
    "inputTokens": 12,
    "outputTokens": 34,
    "totalTokens": 46,
    "origin": {
      "label": "Chat",
      "provider": "OpenAI",
      "surface": "Desktop",
      "chatType": "long"
    }
  }
}
`)
	writeTestFile(t, transcriptOne, "{\"timestamp\":\"2026-02-07T03:16:10.650Z\"}\n{\"timestamp\":\"2026-02-07T03:20:10.650Z\"}\n")
	writeTestFile(t, transcriptTwo, "{\"timestamp\":\"2026-01-01T10:00:00Z\"}\n")
	modTime := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	if err := os.Chtimes(transcriptTwo, modTime, modTime); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	sessions, err := discoverOpenClawSessions()
	if err != nil {
		t.Fatalf("discoverOpenClawSessions() err = %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("discoverOpenClawSessions() len = %d, want 2", len(sessions))
	}

	var withMeta, withoutMeta OpenClawSession
	for _, session := range sessions {
		switch session.SessionID {
		case "session-1":
			withMeta = session
		case "session-2":
			withoutMeta = session
		}
	}

	if withMeta.DisplayName != "Debug crash" || withMeta.TotalTokens != 46 || withMeta.MessageCount != 2 {
		t.Fatalf("metadata-backed session = %#v", withMeta)
	}
	if withMeta.Source != "OpenAI · Desktop · long · Chat" {
		t.Fatalf("withMeta.Source = %q", withMeta.Source)
	}
	if withMeta.SortTime().IsZero() {
		t.Fatal("withMeta.SortTime() should not be zero")
	}
	if withoutMeta.DisplayName != "session-2" {
		t.Fatalf("withoutMeta.DisplayName = %q", withoutMeta.DisplayName)
	}
	if !withoutMeta.UpdatedAt.Equal(modTime) {
		t.Fatalf("withoutMeta.UpdatedAt = %v, want %v", withoutMeta.UpdatedAt, modTime)
	}

	filtered := filterSessionsBefore(sessions, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	if len(filtered) != 2 {
		t.Fatalf("filterSessionsBefore() len = %d, want 2", len(filtered))
	}

	if err := deleteOpenClawSessions([]OpenClawSession{withMeta}); err != nil {
		t.Fatalf("deleteOpenClawSessions() err = %v", err)
	}
	if pathExists(transcriptOne) {
		t.Fatal("transcriptOne still exists")
	}

	meta, err := readOpenClawSessionsMetadata(metadataPath)
	if err != nil {
		t.Fatalf("readOpenClawSessionsMetadata() err = %v", err)
	}
	if len(meta) != 0 {
		t.Fatalf("metadata still contains deleted session: %#v", meta)
	}
}

func TestOpenClawSessionHelpers(t *testing.T) {
	root := t.TempDir()
	emptyPath := filepath.Join(root, "empty.json")
	invalidPath := filepath.Join(root, "invalid.json")
	transcriptPath := filepath.Join(root, "session.jsonl")
	jsonPath := filepath.Join(root, "data.json")

	writeTestFile(t, emptyPath, " \n")
	writeTestFile(t, invalidPath, "{")
	writeTestFile(t, transcriptPath, "{\"timestamp\":\"2026-02-07T03:16:10.650Z\"}\n{}\n")

	meta, err := readOpenClawSessionsMetadata(emptyPath)
	if err != nil || len(meta) != 0 {
		t.Fatalf("readOpenClawSessionsMetadata(empty) = %#v, %v", meta, err)
	}
	if _, err := readOpenClawSessionsMetadata(filepath.Join(root, "missing.json")); err == nil {
		t.Fatal("readOpenClawSessionsMetadata(missing) should fail")
	}
	if _, err := readOpenClawSessionsMetadata(invalidPath); err == nil {
		t.Fatal("readOpenClawSessionsMetadata(invalid) should fail")
	}

	started, lines, err := scanOpenClawTranscript(transcriptPath)
	if err != nil || lines != 2 || started.IsZero() {
		t.Fatalf("scanOpenClawTranscript() = %v, %d, %v", started, lines, err)
	}
	if _, _, err := scanOpenClawTranscript(filepath.Join(root, "missing")); err == nil {
		t.Fatal("scanOpenClawTranscript(missing) should fail")
	}
	writeTestFile(t, filepath.Join(root, "no-timestamp.jsonl"), "{}\n{}\n")
	started, lines, err = scanOpenClawTranscript(filepath.Join(root, "no-timestamp.jsonl"))
	if err != nil || lines != 2 || !started.IsZero() {
		t.Fatalf("scanOpenClawTranscript(no timestamp) = %v, %d, %v", started, lines, err)
	}

	if err := writeJSONAtomic(jsonPath, map[string]string{"a": "b"}); err != nil {
		t.Fatalf("writeJSONAtomic() err = %v", err)
	}
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("ReadFile() err = %v", err)
	}
	if !strings.HasSuffix(string(data), "\n") || !strings.Contains(string(data), `"a": "b"`) {
		t.Fatalf("writeJSONAtomic() content = %q", string(data))
	}
	if err := writeJSONAtomic(filepath.Join(root, "missing-dir", "data.json"), map[string]string{"a": "b"}); err == nil {
		t.Fatal("writeJSONAtomic(invalid path) should fail")
	}
	if err := writeJSONAtomic(filepath.Join(root, "bad.json"), make(chan int)); err == nil {
		t.Fatal("writeJSONAtomic(bad value) should fail")
	}

	if err := deleteOpenClawSessions([]OpenClawSession{{SessionID: "orphan", TranscriptPath: filepath.Join(root, "missing.jsonl")}}); err != nil {
		t.Fatalf("deleteOpenClawSessions(orphan) err = %v", err)
	}

	session := OpenClawSession{
		AgentID:     "worker",
		SessionID:   "abc",
		DisplayName: "A very long display label that should be trimmed by ShortLabel for narrow panes",
		Source:      "source",
		StartedAt:   time.Unix(10, 0),
	}
	if session.SortTime() != session.StartedAt {
		t.Fatal("SortTime() should fall back to StartedAt")
	}
	if !strings.Contains(session.ShortLabel(), "worker") {
		t.Fatalf("ShortLabel() = %q", session.ShortLabel())
	}
	if session.DisplayLabel() != session.DisplayName {
		t.Fatalf("DisplayLabel() = %q", session.DisplayLabel())
	}

	session.DisplayName = ""
	if session.DisplayLabel() != "source" {
		t.Fatalf("DisplayLabel() source fallback = %q", session.DisplayLabel())
	}
	session.Source = ""
	if session.DisplayLabel() != "abc" {
		t.Fatalf("DisplayLabel() id fallback = %q", session.DisplayLabel())
	}

	metaLabel := openClawSessionMeta{
		DisplayName: "Named",
		Origin: &openClawSessionFrom{
			Label:    "Chat",
			Provider: "OpenAI",
			Surface:  "Desktop",
			ChatType: "long",
		},
	}
	if got := bestOpenClawSessionLabel(metaLabel, "id"); got != "Named" {
		t.Fatalf("bestOpenClawSessionLabel(display) = %q", got)
	}
	metaLabel.DisplayName = ""
	if got := bestOpenClawSessionLabel(metaLabel, "id"); got != "Chat" {
		t.Fatalf("bestOpenClawSessionLabel(origin label) = %q", got)
	}
	metaLabel.Origin.Label = ""
	if got := bestOpenClawSessionLabel(metaLabel, "id"); got != "OpenAI long" {
		t.Fatalf("bestOpenClawSessionLabel(provider chatType) = %q", got)
	}
	if got := bestOpenClawSessionSource(metaLabel); got != "OpenAI · Desktop · long" {
		t.Fatalf("bestOpenClawSessionSource() = %q", got)
	}
	if bestOpenClawSessionSource(openClawSessionMeta{}) != "" {
		t.Fatal("bestOpenClawSessionSource(nil origin) should be empty")
	}

	if parseOpenClawTimestamp("bad").IsZero() == false {
		t.Fatal("parseOpenClawTimestamp(bad) should be zero")
	}
	if parseOpenClawTimestamp("").IsZero() == false {
		t.Fatal("parseOpenClawTimestamp(empty) should be zero")
	}
	if parseOpenClawTimestamp("2026-02-07T03:16:10.650Z").IsZero() {
		t.Fatal("parseOpenClawTimestamp(valid) should parse")
	}
	if unixMilli(0) != (time.Time{}) {
		t.Fatal("unixMilli(0) should be zero")
	}
	if unixMilli(1700000000000).IsZero() {
		t.Fatal("unixMilli(valid) should parse")
	}
	if trimForDisplay("abcdef", 3) != "abc" {
		t.Fatalf("trimForDisplay(limit=3) = %q", trimForDisplay("abcdef", 3))
	}
	if trimForDisplay("abcdef", 5) != "ab..." {
		t.Fatalf("trimForDisplay(limit=5) = %q", trimForDisplay("abcdef", 5))
	}
}
