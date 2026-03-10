package sessionstore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-runewidth"
)

var verboseState struct {
	mu      sync.Mutex
	enabled bool
	logger  func(string, ...any)
}

func SetVerboseLogger(enabled bool, logger func(string, ...any)) {
	verboseState.mu.Lock()
	defer verboseState.mu.Unlock()
	verboseState.enabled = enabled
	verboseState.logger = logger
}

func verbosef(format string, args ...any) {
	verboseState.mu.Lock()
	defer verboseState.mu.Unlock()
	if !verboseState.enabled || verboseState.logger == nil {
		return
	}
	verboseState.logger(format, args...)
}

func pathExists(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	_, err := os.Lstat(path)
	return err == nil
}

func cleanPath(path string) string {
	return filepath.Clean(filepath.Clean(strings.TrimSpace(path)))
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
	filepath.WalkDir(path, func(_ string, d os.DirEntry, walkErr error) error {
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

func trimForDisplay(value string, limit int) string {
	value = strings.TrimSpace(value)
	if limit <= 0 || value == "" {
		return ""
	}
	if runewidth.StringWidth(value) <= limit {
		return value
	}
	if limit <= 3 {
		return runewidth.Truncate(value, limit, "")
	}
	return runewidth.Truncate(value, limit, "...")
}

func parseOpenClawTimestamp(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if ts, err := time.Parse(layout, raw); err == nil {
			return ts
		}
	}
	return time.Time{}
}

func writeJSONAtomic(path string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	temp := path + ".tmp"
	if err := os.WriteFile(temp, append(data, '\n'), 0o600); err != nil {
		return err
	}
	return os.Rename(temp, path)
}

func deletePath(path string) error {
	path = cleanPath(path)
	if !filepath.IsAbs(path) {
		return fmt.Errorf("refusing to delete non-absolute path %q", path)
	}
	switch path {
	case "/":
		return fmt.Errorf("refusing to delete protected path %q", path)
	}
	home, _ := os.UserHomeDir()
	if path == home {
		return fmt.Errorf("refusing to delete home directory")
	}
	if !pathExists(path) {
		return nil
	}
	return os.RemoveAll(path)
}

func latestMatchingPath(pattern string) (string, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return "", err
	}
	sort.Strings(matches)
	return matches[len(matches)-1], nil
}

func firstNonEmptyLine(value string) string {
	for _, line := range strings.Split(value, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
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

func errUnexpectedSessionProviderData(assistant string, value any) error {
	return fmt.Errorf("%s session provider data type mismatch: %T", assistant, value)
}

func unixTimeAuto(raw int64) time.Time {
	if raw <= 0 {
		return time.Time{}
	}
	if raw > 1_000_000_000_000 {
		return time.UnixMilli(raw)
	}
	return time.Unix(raw, 0)
}

func displayAssistantName(assistant string) string {
	switch assistant {
	case "codex":
		return "Codex Desktop"
	case "codex-cli":
		return "Codex CLI"
	case "claudecode":
		return "Claude Code"
	case "cursor":
		return "Cursor"
	case "antigravity":
		return "Antigravity"
	default:
		return assistant
	}
}
