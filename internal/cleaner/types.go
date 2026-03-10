package cleaner

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const appName = "OpenAgentCleaner"

var Version = "dev"

type Safety string

const (
	SafetySafe    Safety = "safe"
	SafetyConfirm Safety = "confirm"
	SafetyManual  Safety = "manual"
)

type Candidate struct {
	ID                   string   `json:"id"`
	Assistant            string   `json:"assistant"`
	Path                 string   `json:"path"`
	Kind                 string   `json:"kind"`
	Safety               Safety   `json:"safety"`
	Reason               string   `json:"reason"`
	Notes                []string `json:"notes,omitempty"`
	SizeBytes            int64    `json:"size_bytes"`
	Deletable            bool     `json:"deletable"`
	RequiresConfirmation bool     `json:"requires_confirmation"`
	Selected             bool     `json:"selected,omitempty"`
	Deleted              bool     `json:"deleted,omitempty"`
	Skipped              bool     `json:"skipped,omitempty"`
	Error                string   `json:"error,omitempty"`
}

type Summary struct {
	CandidateCount int   `json:"candidate_count"`
	SelectedCount  int   `json:"selected_count"`
	DeletedCount   int   `json:"deleted_count"`
	BytesFound     int64 `json:"bytes_found"`
	BytesDeleted   int64 `json:"bytes_deleted"`
}

type Report struct {
	Operation  string      `json:"operation"`
	Mode       string      `json:"mode"`
	DryRun     bool        `json:"dry_run"`
	Assistants []string    `json:"assistants"`
	Candidates []Candidate `json:"candidates"`
	Summary    Summary     `json:"summary"`
}

type options struct {
	Assistants   []string
	Mode         string
	JSON         bool
	Verbose      bool
	CandidateIDs []string
	Kinds        []string
	Safeties     []Safety
	Yes          bool
	DryRun       bool
}

var supportedAssistants = []string{
	"openclaw",
	"ironclaw",
	"ollama",
	"codex",
	"codex-cli",
	"claudecode",
	"cursor",
	"antigravity",
}

func defaultAssistants() []string {
	out := make([]string, 0, len(supportedAssistants))
	out = append(out, supportedAssistants...)
	return out
}

func assistantFlagHelp() string {
	return "comma-separated assistants: " + strings.Join(defaultAssistants(), ",")
}

func parseAssistantList(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return defaultAssistants(), nil
	}

	seen := map[string]struct{}{}
	out := make([]string, 0, len(supportedAssistants))
	for _, part := range strings.Split(raw, ",") {
		name := strings.ToLower(strings.TrimSpace(part))
		switch name {
		case "openclaw", "ironclaw", "ollama", "codex", "codex-cli", "codexcli", "claudecode", "claude-code", "cursor", "antigravity":
			if name == "codexcli" {
				name = "codex-cli"
			}
			if name == "claude-code" {
				name = "claudecode"
			}
			if _, ok := seen[name]; !ok {
				seen[name] = struct{}{}
				out = append(out, name)
			}
		case "":
		default:
			return nil, fmt.Errorf("unsupported assistant %q", name)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no assistants selected")
	}
	sort.Strings(out)
	return out, nil
}

func normalizeMode(raw string, jsonOutput bool) string {
	mode := strings.ToLower(strings.TrimSpace(raw))
	switch mode {
	case "", "auto":
		if jsonOutput || !isTerminal(os.Stdout) {
			return "agent"
		}
		return "human"
	case "agent", "human":
		return mode
	default:
		return "human"
	}
}

func isTerminal(file *os.File) bool {
	if file == nil {
		return false
	}
	info, err := file.Stat()
	return err == nil && (info.Mode()&os.ModeCharDevice) != 0
}

func formatBytes(size int64) string {
	if size <= 0 {
		return "0 B"
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	value := float64(size)
	unit := 0
	for value >= 1024 && unit < len(units)-1 {
		value /= 1024
		unit++
	}
	if unit == 0 {
		return fmt.Sprintf("%d %s", size, units[unit])
	}
	return fmt.Sprintf("%.1f %s", value, units[unit])
}

func toJSON(v any) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error":%q}`, err.Error())
	}
	return string(data)
}

func cleanPath(path string) string {
	return filepath.Clean(filepath.Clean(strings.TrimSpace(path)))
}

func candidateID(assistant, kind, path string) string {
	sum := sha1.Sum([]byte(assistant + "\x00" + kind + "\x00" + cleanPath(path)))
	return fmt.Sprintf("%s-%x", assistant, sum[:6])
}

func parseCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	out := make([]string, 0, 4)
	seen := map[string]struct{}{}
	for _, part := range strings.Split(raw, ",") {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func parseSafetyList(raw string) ([]Safety, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	items := parseCSV(strings.ToLower(raw))
	out := make([]Safety, 0, len(items))
	for _, item := range items {
		switch Safety(item) {
		case SafetySafe, SafetyConfirm:
			out = append(out, Safety(item))
		case SafetyManual:
			return nil, fmt.Errorf("safety %q is manual-only and cannot be deleted", item)
		default:
			return nil, fmt.Errorf("unsupported safety %q", item)
		}
	}
	return out, nil
}
