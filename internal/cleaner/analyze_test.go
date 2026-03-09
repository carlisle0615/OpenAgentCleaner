package cleaner

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunAnalyzeValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    int
		wantErr string
	}{
		{name: "json", args: []string{"--json"}, want: 2, wantErr: "interactive human mode only"},
		{name: "agent", args: []string{"--mode", "agent"}, want: 2, wantErr: "interactive human mode only"},
		{name: "bad assistant", args: []string{"--assistant", "bad"}, want: 2, wantErr: "unsupported assistant"},
		{name: "bad date", args: []string{"--before", "2026/03/01"}, want: 2, wantErr: "expected YYYY-MM-DD"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := runAnalyze(tc.args, &stdout, &stderr)
			if code != tc.want {
				t.Fatalf("runAnalyze(%v) = %d, want %d", tc.args, code, tc.want)
			}
			if !strings.Contains(stderr.String(), tc.wantErr) {
				t.Fatalf("stderr = %q, want %q", stderr.String(), tc.wantErr)
			}
		})
	}
}

func TestAnalyzeHelpers(t *testing.T) {
	home := setTestHome(t)
	root := filepath.Join(home, ".openclaw")
	writeTestFile(t, filepath.Join(root, "logs", "app.log"), "log")
	writeTestFile(t, filepath.Join(root, "agents", "main", "sessions", "sessions.json"), `{"one":{"sessionId":"session-1","updatedAt":1700000000000}}`)
	writeTestFile(t, filepath.Join(root, "agents", "main", "sessions", "session-1.jsonl"), "{\"timestamp\":\"2026-02-07T03:16:10.650Z\"}\n")
	writeTestFile(t, filepath.Join(home, ".ollama", "logs", "app.log"), "log")
	writeTestFile(t, filepath.Join(root, "workspace", "notes.md"), "keep")

	summary, err := assistantAnalyzeSummary([]string{"openclaw", "ollama"})
	if err != nil {
		t.Fatalf("assistantAnalyzeSummary() err = %v", err)
	}
	if len(summary) != 2 || summary[0].SessionCount != 1 {
		t.Fatalf("assistantAnalyzeSummary() = %#v", summary)
	}

	leftovers, err := discoverAssistantLeftovers("openclaw")
	if err != nil {
		t.Fatalf("discoverAssistantLeftovers() err = %v", err)
	}
	for _, candidate := range leftovers {
		if candidate.Kind == "session_store" {
			t.Fatalf("discoverAssistantLeftovers() should filter session_store: %#v", leftovers)
		}
	}

	tempFile := filepath.Join(home, "temp.txt")
	writeTestFile(t, tempFile, "12345")
	live := liveCandidates([]Candidate{
		{Path: tempFile, Assistant: "ollama"},
		{Path: filepath.Join(home, "missing"), Assistant: "ollama"},
	})
	if len(live) != 1 || live[0].SizeBytes != 5 {
		t.Fatalf("liveCandidates() = %#v", live)
	}

	cutoff, err := parseDateCutoff("2026-03-01")
	if err != nil || cutoff.IsZero() {
		t.Fatalf("parseDateCutoff(valid) = %v, %v", cutoff, err)
	}
	if _, err := parseDateCutoff("bad"); err == nil {
		t.Fatal("parseDateCutoff(invalid) should fail")
	}

	if formatTokenCount(0) != "-" || formatTokenCount(999) != "999 tok" || formatTokenCount(1200) != "1.2K tok" {
		t.Fatalf("formatTokenCount outputs are wrong")
	}
	if formatSessionTime(time.Time{}) != "-" {
		t.Fatal("formatSessionTime(zero) should be -")
	}
	if got := formatSessionTime(time.Date(2026, 3, 1, 10, 30, 0, 0, time.UTC)); !strings.Contains(got, "2026-03-01") {
		t.Fatalf("formatSessionTime() = %q", got)
	}
}
