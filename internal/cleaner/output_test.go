package cleaner

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintHumanReportAndHelpers(t *testing.T) {
	withFakeArgs(t, []string{"/tmp/oac"}, func() {
		var out bytes.Buffer
		printHumanReport(&out, Report{
			Assistants: []string{"openclaw"},
		}, false)
		if !strings.Contains(out.String(), "No leftover files were found") {
			t.Fatalf("printHumanReport(empty) = %q", out.String())
		}

		report := Report{
			Operation:  "clean",
			Assistants: []string{"openclaw", "ollama"},
			DryRun:     true,
			Candidates: []Candidate{
				{Assistant: "openclaw", Kind: "gateway_logs", Safety: SafetySafe, Reason: "logs", Path: "/tmp/a", SizeBytes: 10, Selected: true, Skipped: true},
				{Assistant: "openclaw", Kind: "workspace", Safety: SafetyManual, Reason: "workspace", Path: "/tmp/b", SizeBytes: 20},
				{Assistant: "ollama", Kind: "models", Safety: SafetyConfirm, Reason: "models", Path: "/tmp/c", SizeBytes: 30, Selected: true},
				{Assistant: "ollama", Kind: "cache", Safety: SafetySafe, Reason: "cache", Path: "/tmp/d", SizeBytes: 40, Error: "boom"},
			},
			Summary: Summary{
				CandidateCount: 4,
				SelectedCount:  2,
			},
		}
		report.Summary.BytesFound = 100

		out.Reset()
		printHumanReport(&out, report, true)
		text := out.String()
		for _, want := range []string{
			"Cleanup preview",
			"Summary",
			"OpenClaw",
			"Ollama",
			"Logs",
			"Models",
			"What to do next",
			"Error",
		} {
			if !strings.Contains(text, want) {
				t.Fatalf("printHumanReport(clean) missing %q\n%s", want, text)
			}
		}

		out.Reset()
		report.DryRun = false
		report.Summary.SelectedCount = 0
		report.Summary.DeletedCount = 0
		printHumanReport(&out, report, true)
		if !strings.Contains(out.String(), "Nothing removed") {
			t.Fatalf("printHumanReport(cancelled) = %q", out.String())
		}

		out.Reset()
		report.Summary.SelectedCount = 1
		report.Summary.DeletedCount = 1
		report.Candidates = []Candidate{
			{Assistant: "openclaw", Kind: "workspace", Safety: SafetyManual, Reason: "workspace", Path: "/tmp/b", SizeBytes: 20},
		}
		printHumanReport(&out, report, true)
		if !strings.Contains(out.String(), "Manual-only items still need your review.") {
			t.Fatalf("printHumanReport(manual left) = %q", out.String())
		}

		out.Reset()
		printHumanReport(&out, Report{
			Assistants: []string{"ollama"},
			Candidates: []Candidate{
				{Assistant: "ollama", Kind: "models", Safety: SafetyConfirm, Reason: "models", Path: "/tmp/c", SizeBytes: 30},
			},
			Summary: Summary{CandidateCount: 1, BytesFound: 30},
		}, false)
		if !strings.Contains(out.String(), "Review the items above carefully.") {
			t.Fatalf("printHumanReport(confirm only) = %q", out.String())
		}
	})

	safe, confirm, manual := summarizeBySafety([]Candidate{
		{Safety: SafetySafe, SizeBytes: 1},
		{Safety: SafetyConfirm, SizeBytes: 2},
		{Safety: SafetyManual, SizeBytes: 3},
	})
	if safe.Count != 1 || confirm.Bytes != 2 || manual.Bytes != 3 {
		t.Fatalf("summarizeBySafety() = %#v %#v %#v", safe, confirm, manual)
	}

	groups := groupCandidatesByAssistant([]Candidate{
		{Assistant: "ollama"},
		{Assistant: "openclaw"},
		{Assistant: "ollama"},
	})
	if len(groups) != 2 || groups[0].Name != "ollama" || groups[1].Name != "openclaw" {
		t.Fatalf("groupCandidatesByAssistant() = %#v", groups)
	}
	if countAssistantsWithCandidates([]Candidate{{Assistant: "a"}, {Assistant: "a"}, {Assistant: "b"}}) != 2 {
		t.Fatal("countAssistantsWithCandidates() should count unique assistants")
	}

	report := Report{
		DryRun: true,
		Candidates: []Candidate{
			{Selected: true, Safety: SafetyConfirm, SizeBytes: 10},
			{Deleted: true, SizeBytes: 20},
			{Error: "boom", SizeBytes: 30},
		},
	}
	if statusLabel(report.Candidates[0], report) != "planned" {
		t.Fatal("statusLabel(planned) mismatch")
	}
	report.DryRun = false
	report.Candidates[0].Skipped = true
	if statusLabel(report.Candidates[0], report) != "skipped" {
		t.Fatal("statusLabel(skipped) mismatch")
	}
	if statusLabel(report.Candidates[1], report) != "removed" {
		t.Fatal("statusLabel(removed) mismatch")
	}
	if statusLabel(report.Candidates[2], report) != "error" {
		t.Fatal("statusLabel(error) mismatch")
	}
	if plannedBytes(Report{Candidates: []Candidate{{Selected: true, SizeBytes: 5}, {SizeBytes: 3}}}) != 5 {
		t.Fatal("plannedBytes() mismatch")
	}
	if !selectedHasConfirm(report) {
		t.Fatal("selectedHasConfirm() should see confirm candidate")
	}
	report.Candidates[0].Selected = false
	if selectedHasConfirm(report) {
		t.Fatal("selectedHasConfirm() should be false when no confirm is selected")
	}

	for kind, want := range map[string]string{
		"gateway_logs":            "Logs",
		"session_store":           "Sessions",
		"config":                  "Config",
		"env_file":                "Env file",
		"database":                "Database",
		"oauth_session":           "Signed-in session",
		"mcp_config":              "MCP servers",
		"legacy_settings":         "Legacy settings",
		"legacy_bootstrap":        "Legacy bootstrap",
		"channels":                "Channels",
		"tools":                   "Tools",
		"repl_history":            "History",
		"allow_from":              "Allow list",
		"service_plist":           "Background service",
		"saved_state":             "Saved window state",
		"webkit_cache":            "Web cache",
		"app_support":             "App support",
		"server_config":           "Server config",
		"app_bundle":              "Mac app",
		"cli_symlink":             "CLI link",
		"archived_sessions":       "Archived sessions",
		"session_index":           "Session index",
		"session_db":              "Session database",
		"desktop_session_storage": "App session storage",
		"desktop_local_storage":   "App local storage",
		"desktop_indexeddb":       "App IndexedDB",
		"transcripts":             "Transcripts",
		"project_sessions":        "Project sessions",
		"prompt_history":          "Prompt history",
		"global_chat_state":       "Global chat state",
		"workspace_chat_state":    "Workspace chat state",
		"unknown_kind":            "Unknown Kind",
		"pairing_store":           "Pairing data",
		"pairing_attempts":        "Pairing attempts",
	} {
		if got := displayKind(kind); got != want {
			t.Fatalf("displayKind(%q) = %q, want %q", kind, got, want)
		}
	}
	if displayAssistant("openclaw") != "OpenClaw" ||
		displayAssistant("codex") != "Codex Desktop" ||
		displayAssistant("codex-cli") != "Codex CLI" ||
		displayAssistant("claudecode") != "Claude Code" ||
		displayAssistant("cursor") != "Cursor" ||
		displayAssistant("antigravity") != "Antigravity" ||
		displayAssistant("other_tool") != "Other Tool" {
		t.Fatal("displayAssistant() mismatch")
	}
	if displaySafety(SafetySafe) != "safe" || displaySafety("custom") != "custom" {
		t.Fatal("displaySafety() mismatch")
	}
	if titleize("hello_world-test") != "Hello World Test" {
		t.Fatalf("titleize() = %q", titleize("hello_world-test"))
	}
}
