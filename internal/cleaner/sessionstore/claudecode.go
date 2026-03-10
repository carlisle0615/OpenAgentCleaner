package sessionstore

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type claudeCodeSession struct {
	SessionID  string
	Path       string
	ProjectDir string
	Title      string
	StartedAt  time.Time
	UpdatedAt  time.Time
	Messages   int
}

type claudeCodeEvent struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	SessionID string `json:"sessionId"`
	Message   struct {
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"message"`
}

func DiscoverClaudeCodeConversationSessions() ([]ConversationSession, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	globs := []string{
		filepath.Join(home, ".claude", "transcripts", "*.jsonl"),
		filepath.Join(home, ".claude", "projects", "*", "*.jsonl"),
	}
	verbosef("reading Claude Code transcripts from %s", filepath.Join(home, ".claude"))

	out := []ConversationSession{}
	for _, pattern := range globs {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		sort.Strings(matches)
		for _, path := range matches {
			if strings.HasSuffix(path, "history.jsonl") {
				continue
			}
			info, err := os.Stat(path)
			if err != nil {
				return nil, err
			}
			session, err := scanClaudeCodeSession(path)
			if err != nil {
				return nil, err
			}
			if session.UpdatedAt.IsZero() {
				session.UpdatedAt = info.ModTime()
			}
			projectDir := filepath.Dir(path)
			if filepath.Base(projectDir) == "transcripts" {
				projectDir = ""
			}
			session.ProjectDir = projectDir
			out = append(out, ConversationSession{
				Assistant:    "claudecode",
				ID:           session.SessionID,
				Title:        session.Title,
				Subtitle:     claudeCodeSubtitle(projectDir),
				Source:       claudeCodeSource(projectDir),
				Path:         path,
				StartedAt:    session.StartedAt,
				UpdatedAt:    session.UpdatedAt,
				SizeBytes:    info.Size(),
				MessageCount: session.Messages,
				Deletable:    true,
				ProviderData: session,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].SortTime().After(out[j].SortTime())
	})
	return out, nil
}

func scanClaudeCodeSession(path string) (claudeCodeSession, error) {
	file, err := os.Open(path)
	if err != nil {
		return claudeCodeSession{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	session := claudeCodeSession{
		SessionID: strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
		Path:      path,
		Title:     strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
	}
	for scanner.Scan() {
		var event claudeCodeEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}
		ts := parseOpenClawTimestamp(event.Timestamp)
		if session.StartedAt.IsZero() && !ts.IsZero() {
			session.StartedAt = ts
		}
		if !ts.IsZero() {
			session.UpdatedAt = ts
		}
		if event.SessionID != "" {
			session.SessionID = event.SessionID
		}
		if event.Type != "user" && event.Type != "assistant" {
			continue
		}
		text := collectClaudeCodeText(event.Message.Content)
		if text == "" {
			continue
		}
		if session.Title == strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)) && event.Type == "user" {
			session.Title = firstNonEmptyLine(text)
		}
		session.Messages++
	}
	if err := scanner.Err(); err != nil {
		return claudeCodeSession{}, err
	}
	return session, nil
}

func PreviewClaudeCodeConversationSession(session ConversationSession) (string, error) {
	stored, ok := session.ProviderData.(claudeCodeSession)
	if !ok {
		return "", errUnexpectedSessionProviderData("claudecode", session.ProviderData)
	}
	file, err := os.Open(stored.Path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	messages := []string{}
	for scanner.Scan() {
		var event claudeCodeEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}
		if event.Type != "user" && event.Type != "assistant" {
			continue
		}
		text := collectClaudeCodeText(event.Message.Content)
		if text == "" {
			continue
		}
		role := "User"
		if event.Type == "assistant" {
			role = "Assistant"
		}
		messages = append(messages, fmt.Sprintf("== %s ==\n%s", role, text))
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if len(messages) == 0 {
		return "No text messages found in this conversation.", nil
	}
	if len(messages) > 6 {
		messages = messages[len(messages)-6:]
	}
	return strings.Join(messages, "\n\n"), nil
}

func DeleteClaudeCodeConversationSessions(sessions []ConversationSession) error {
	for _, session := range sessions {
		stored, ok := session.ProviderData.(claudeCodeSession)
		if !ok {
			return errUnexpectedSessionProviderData("claudecode", session.ProviderData)
		}

		indexPath := ""
		if stored.ProjectDir != "" {
			indexPath = filepath.Join(stored.ProjectDir, "sessions-index.json")
		}

		staged, err := stageDeletePaths([]string{stored.Path})
		if err != nil {
			return err
		}
		indexBackup, err := backupFileIfExists(indexPath)
		if err != nil {
			_ = restoreStagedDeletes(staged)
			return err
		}

		rollback := func(cause error) error {
			if restoreErr := restoreCopiedBackups(indexBackup); restoreErr != nil {
				return fmt.Errorf("%w; restore Claude Code session index: %v", cause, restoreErr)
			}
			if restoreErr := restoreStagedDeletes(staged); restoreErr != nil {
				return fmt.Errorf("%w; restore Claude Code transcript: %v", cause, restoreErr)
			}
			_ = cleanupCopiedBackups(indexBackup)
			return cause
		}

		if indexPath != "" {
			if err := removeClaudeCodeProjectIndexEntry(indexPath, stored); err != nil {
				return rollback(err)
			}
		}
		if err := cleanupCopiedBackups(indexBackup); err != nil {
			return err
		}
		if err := cleanupStagedDeletes(staged); err != nil {
			return err
		}
	}
	return nil
}

func removeClaudeCodeProjectIndexEntry(path string, session claudeCodeSession) error {
	if !pathExists(path) {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	switch entries := payload["entries"].(type) {
	case []any:
		filtered := make([]any, 0, len(entries))
		for _, entry := range entries {
			if shouldKeepClaudeCodeIndexEntry(entry, session) {
				filtered = append(filtered, entry)
			}
		}
		payload["entries"] = filtered
	case map[string]any:
		for key, entry := range entries {
			if !shouldKeepClaudeCodeIndexEntry(entry, session) || key == session.SessionID {
				delete(entries, key)
			}
		}
		payload["entries"] = entries
	}

	return writeJSONAtomic(path, payload)
}

func shouldKeepClaudeCodeIndexEntry(entry any, session claudeCodeSession) bool {
	removeValues := []string{
		session.SessionID,
		filepath.Base(session.Path),
		session.Path,
	}
	switch value := entry.(type) {
	case map[string]any:
		for _, candidate := range removeValues {
			if candidate == "" {
				continue
			}
			if mapContainsStringValue(value, candidate) {
				return false
			}
		}
	}
	return true
}

func mapContainsStringValue(value map[string]any, needle string) bool {
	for _, item := range value {
		switch typed := item.(type) {
		case string:
			if typed == needle {
				return true
			}
		case map[string]any:
			if mapContainsStringValue(typed, needle) {
				return true
			}
		case []any:
			for _, child := range typed {
				if childMap, ok := child.(map[string]any); ok && mapContainsStringValue(childMap, needle) {
					return true
				}
				if childValue, ok := child.(string); ok && childValue == needle {
					return true
				}
			}
		}
	}
	return false
}

func collectClaudeCodeText(content []struct {
	Type string `json:"type"`
	Text string `json:"text"`
}) string {
	parts := []string{}
	for _, item := range content {
		if item.Type == "text" && strings.TrimSpace(item.Text) != "" {
			parts = append(parts, strings.TrimSpace(item.Text))
		}
	}
	return strings.Join(parts, "\n")
}

func claudeCodeSource(projectDir string) string {
	if projectDir == "" {
		return "Global transcript"
	}
	return "Project transcript"
}

func claudeCodeSubtitle(projectDir string) string {
	if projectDir == "" {
		return ""
	}
	return filepath.Base(projectDir)
}
