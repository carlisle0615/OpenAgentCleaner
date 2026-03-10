package cleaner

import (
	"os"
	"strings"
	"testing"
)

func TestDefaultAssistants(t *testing.T) {
	got := defaultAssistants()
	want := []string{"openclaw", "ironclaw", "ollama", "codex", "codex-cli", "claudecode", "cursor", "antigravity"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("defaultAssistants() = %v, want %v", got, want)
	}
}

func TestParseAssistantList(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr string
	}{
		{name: "default", raw: "", want: "openclaw,ironclaw,ollama,codex,codex-cli,claudecode,cursor,antigravity"},
		{name: "dedupe and sort", raw: "ollama,openclaw,ollama", want: "ollama,openclaw"},
		{name: "trim and normalize", raw: " IronClaw , openclaw ", want: "ironclaw,openclaw"},
		{name: "new assistants", raw: "cursor,codex,codexcli,claude-code,antigravity", want: "antigravity,claudecode,codex,codex-cli,cursor"},
		{name: "unsupported", raw: "foo", wantErr: `unsupported assistant "foo"`},
		{name: "empty selection", raw: ",,", wantErr: "no assistants selected"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseAssistantList(tc.raw)
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Fatalf("parseAssistantList(%q) err = %v, want %q", tc.raw, err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseAssistantList(%q) err = %v", tc.raw, err)
			}
			if strings.Join(got, ",") != tc.want {
				t.Fatalf("parseAssistantList(%q) = %v, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestNormalizeMode(t *testing.T) {
	if got := normalizeMode("", true); got != "agent" {
		t.Fatalf("normalizeMode json = %q, want agent", got)
	}
	if got := normalizeMode("agent", false); got != "agent" {
		t.Fatalf("normalizeMode(agent) = %q", got)
	}
	if got := normalizeMode("human", false); got != "human" {
		t.Fatalf("normalizeMode(human) = %q", got)
	}
	if got := normalizeMode("weird", false); got != "human" {
		t.Fatalf("normalizeMode(weird) = %q", got)
	}
	if got := normalizeMode("auto", false); got != "agent" {
		t.Fatalf("normalizeMode(auto) = %q, want agent in tests", got)
	}

	devNull, err := os.Open("/dev/null")
	if err != nil {
		t.Fatalf("open /dev/null: %v", err)
	}
	defer devNull.Close()
	withStdoutFile(t, devNull, func() {
		if got := normalizeMode("auto", false); got != "human" {
			t.Fatalf("normalizeMode(auto with char device) = %q, want human", got)
		}
	})
}

func TestIsTerminalNil(t *testing.T) {
	if isTerminal(nil) {
		t.Fatal("isTerminal(nil) = true, want false")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := map[int64]string{
		0:           "0 B",
		12:          "12 B",
		1024:        "1.0 KB",
		1536:        "1.5 KB",
		1024 * 1024: "1.0 MB",
	}
	for size, want := range tests {
		if got := formatBytes(size); got != want {
			t.Fatalf("formatBytes(%d) = %q, want %q", size, got, want)
		}
	}
}

func TestToJSON(t *testing.T) {
	if got := toJSON(map[string]string{"a": "b"}); !strings.Contains(got, `"a": "b"`) {
		t.Fatalf("toJSON(map) = %q", got)
	}
	if got := toJSON(make(chan int)); !strings.Contains(got, `"error"`) {
		t.Fatalf("toJSON(error) = %q", got)
	}
}

func TestCleanPath(t *testing.T) {
	if got := cleanPath(" /tmp/foo/../bar "); got != "/tmp/bar" {
		t.Fatalf("cleanPath() = %q", got)
	}
}

func TestAssistantFlagHelp(t *testing.T) {
	if got := assistantFlagHelp(); !strings.Contains(got, "claudecode") || !strings.Contains(got, "cursor") {
		t.Fatalf("assistantFlagHelp() = %q", got)
	}
}
