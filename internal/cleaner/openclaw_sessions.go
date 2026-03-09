package cleaner

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

type OpenClawSession struct {
	AgentID        string
	SessionKey     string
	SessionID      string
	MetadataPath   string
	TranscriptPath string
	DisplayName    string
	Source         string
	StartedAt      time.Time
	UpdatedAt      time.Time
	SizeBytes      int64
	MessageCount   int
	InputTokens    int64
	OutputTokens   int64
	TotalTokens    int64
}

type openClawSessionMeta struct {
	SessionID    string               `json:"sessionId"`
	DisplayName  string               `json:"displayName"`
	UpdatedAt    int64                `json:"updatedAt"`
	InputTokens  int64                `json:"inputTokens"`
	OutputTokens int64                `json:"outputTokens"`
	TotalTokens  int64                `json:"totalTokens"`
	Origin       *openClawSessionFrom `json:"origin,omitempty"`
}

type openClawSessionFrom struct {
	Label    string `json:"label"`
	Provider string `json:"provider"`
	Surface  string `json:"surface"`
	ChatType string `json:"chatType"`
}

type openClawSessionEvent struct {
	Timestamp string `json:"timestamp"`
}

type openClawMessageEvent struct {
	Type    string `json:"type"`
	Message struct {
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		ErrorMessage string `json:"errorMessage,omitempty"`
	} `json:"message"`
}

func discoverOpenClawSessions() ([]OpenClawSession, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var sessions []OpenClawSession
	for _, root := range openClawStateRoots(home) {
		agentRoots, err := filepath.Glob(filepath.Join(root, "agents", "*"))
		if err != nil {
			return nil, err
		}
		for _, agentRoot := range agentRoots {
			agentID := filepath.Base(agentRoot)
			sessionDir := filepath.Join(agentRoot, "sessions")
			metadataPath := filepath.Join(sessionDir, "sessions.json")
			if !pathExists(sessionDir) || !pathExists(metadataPath) {
				continue
			}

			metadata, err := readOpenClawSessionsMetadata(metadataPath)
			if err != nil {
				return nil, err
			}
			bySessionID := map[string]struct {
				Key  string
				Meta openClawSessionMeta
			}{}
			for key, meta := range metadata {
				if meta.SessionID == "" {
					continue
				}
				bySessionID[meta.SessionID] = struct {
					Key  string
					Meta openClawSessionMeta
				}{
					Key:  key,
					Meta: meta,
				}
			}

			transcripts, err := filepath.Glob(filepath.Join(sessionDir, "*.jsonl"))
			if err != nil {
				return nil, err
			}
			for _, transcript := range transcripts {
				sessionID := strings.TrimSuffix(filepath.Base(transcript), filepath.Ext(transcript))
				startedAt, lineCount, err := scanOpenClawTranscript(transcript)
				if err != nil {
					return nil, err
				}
				metaEntry, ok := bySessionID[sessionID]
				session := OpenClawSession{
					AgentID:        agentID,
					SessionID:      sessionID,
					MetadataPath:   metadataPath,
					TranscriptPath: transcript,
					StartedAt:      startedAt,
					SizeBytes:      pathSize(transcript),
					MessageCount:   lineCount,
				}
				if ok {
					session.SessionKey = metaEntry.Key
					session.InputTokens = metaEntry.Meta.InputTokens
					session.OutputTokens = metaEntry.Meta.OutputTokens
					session.TotalTokens = metaEntry.Meta.TotalTokens
					session.UpdatedAt = unixMilli(metaEntry.Meta.UpdatedAt)
					session.DisplayName = bestOpenClawSessionLabel(metaEntry.Meta, sessionID)
					session.Source = bestOpenClawSessionSource(metaEntry.Meta)
				} else {
					session.DisplayName = sessionID
				}
				if session.UpdatedAt.IsZero() {
					info, statErr := os.Stat(transcript)
					if statErr == nil {
						session.UpdatedAt = info.ModTime()
					}
				}
				sessions = append(sessions, session)
			}
		}
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].SortTime().After(sessions[j].SortTime())
	})
	return sessions, nil
}

func readOpenClawSessionsMetadata(path string) (map[string]openClawSessionMeta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	meta := map[string]openClawSessionMeta{}
	if len(strings.TrimSpace(string(data))) == 0 {
		return meta, nil
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return meta, nil
}

func scanOpenClawTranscript(path string) (time.Time, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return time.Time{}, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buffer := make([]byte, 0, 64*1024)
	scanner.Buffer(buffer, 2*1024*1024)

	var startedAt time.Time
	var lines int
	for scanner.Scan() {
		lines++
		if startedAt.IsZero() {
			var event openClawSessionEvent
			if err := json.Unmarshal(scanner.Bytes(), &event); err == nil && event.Timestamp != "" {
				startedAt = parseOpenClawTimestamp(event.Timestamp)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return time.Time{}, 0, err
	}
	return startedAt, lines, nil
}

func previewOpenClawSession(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buffer := make([]byte, 0, 64*1024)
	scanner.Buffer(buffer, 2*1024*1024)

	var messages []string
	for scanner.Scan() {
		var event openClawMessageEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err == nil && event.Type == "message" {
			role := event.Message.Role
			if role == "" {
				continue
			}

			var textParts []string
			if event.Message.ErrorMessage != "" {
				textParts = append(textParts, "[Error] "+event.Message.ErrorMessage)
			}
			for _, c := range event.Message.Content {
				if c.Type == "text" && strings.TrimSpace(c.Text) != "" {
					textParts = append(textParts, strings.TrimSpace(c.Text))
				}
			}

			content := strings.Join(textParts, "\n")
			if content != "" {
				prefix := "User"
				if role == "assistant" {
					prefix = "Assistant"
				} else if role == "system" {
					prefix = "System"
				}
				messages = append(messages, fmt.Sprintf("== %s ==\n%s", prefix, content))
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if len(messages) == 0 {
		return "No text messages found in this conversation.", nil
	}

	// Keep only the last 4 messages to avoid overwhelming the terminal
	if len(messages) > 4 {
		messages = messages[len(messages)-4:]
	}
	return strings.Join(messages, "\n\n"), nil
}

func deleteOpenClawSessions(sessions []OpenClawSession) error {
	type sessionKey struct {
		Key       string
		SessionID string
	}

	grouped := map[string][]OpenClawSession{}
	for _, session := range sessions {
		grouped[session.MetadataPath] = append(grouped[session.MetadataPath], session)
	}

	for metadataPath, group := range grouped {
		deleted := make([]sessionKey, 0, len(group))
		for _, session := range group {
			if err := deletePath(session.TranscriptPath); err != nil {
				return err
			}
			deleted = append(deleted, sessionKey{
				Key:       session.SessionKey,
				SessionID: session.SessionID,
			})
		}

		if metadataPath == "" || !pathExists(metadataPath) || len(deleted) == 0 {
			continue
		}

		meta, err := readOpenClawSessionsMetadata(metadataPath)
		if err != nil {
			return err
		}
		removeKeys := map[string]struct{}{}
		removeIDs := map[string]struct{}{}
		for _, item := range deleted {
			if item.Key != "" {
				removeKeys[item.Key] = struct{}{}
			}
			if item.SessionID != "" {
				removeIDs[item.SessionID] = struct{}{}
			}
		}
		for key, value := range meta {
			if _, ok := removeKeys[key]; ok {
				delete(meta, key)
				continue
			}
			if _, ok := removeIDs[value.SessionID]; ok {
				delete(meta, key)
			}
		}
		if err := writeJSONAtomic(metadataPath, meta); err != nil {
			return err
		}
	}
	return nil
}

func writeJSONAtomic(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	temp := path + ".tmp"
	if err := os.WriteFile(temp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(temp, path)
}

func filterSessionsBefore(sessions []OpenClawSession, cutoff time.Time) []OpenClawSession {
	out := make([]OpenClawSession, 0, len(sessions))
	for _, session := range sessions {
		if session.SortTime().Before(cutoff) {
			out = append(out, session)
		}
	}
	return out
}

func (s OpenClawSession) SortTime() time.Time {
	if !s.UpdatedAt.IsZero() {
		return s.UpdatedAt
	}
	return s.StartedAt
}

func (s OpenClawSession) ShortLabel() string {
	label := s.DisplayLabel()
	if s.AgentID != "" && s.AgentID != "main" {
		label = s.AgentID + " · " + label
	}
	return trimForDisplay(label, 64)
}

func (s OpenClawSession) DisplayLabel() string {
	if strings.TrimSpace(s.DisplayName) != "" && s.DisplayName != s.SessionID {
		return s.DisplayName
	}
	if strings.TrimSpace(s.Source) != "" {
		return s.Source
	}
	return s.SessionID
}

func bestOpenClawSessionLabel(meta openClawSessionMeta, sessionID string) string {
	if strings.TrimSpace(meta.DisplayName) != "" {
		return meta.DisplayName
	}
	if meta.Origin != nil && strings.TrimSpace(meta.Origin.Label) != "" {
		return meta.Origin.Label
	}
	if meta.Origin != nil {
		parts := []string{}
		if meta.Origin.Provider != "" {
			parts = append(parts, meta.Origin.Provider)
		}
		if meta.Origin.ChatType != "" {
			parts = append(parts, meta.Origin.ChatType)
		}
		if len(parts) > 0 {
			return strings.Join(parts, " ")
		}
	}
	return sessionID
}

func bestOpenClawSessionSource(meta openClawSessionMeta) string {
	if meta.Origin == nil {
		return ""
	}
	parts := []string{}
	if meta.Origin.Provider != "" {
		parts = append(parts, meta.Origin.Provider)
	}
	if meta.Origin.Surface != "" && meta.Origin.Surface != meta.Origin.Provider {
		parts = append(parts, meta.Origin.Surface)
	}
	if meta.Origin.ChatType != "" {
		parts = append(parts, meta.Origin.ChatType)
	}
	if meta.Origin.Label != "" {
		parts = append(parts, meta.Origin.Label)
	}
	return strings.Join(parts, " · ")
}

func parseOpenClawTimestamp(raw string) time.Time {
	if raw == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339Nano, raw)
	if err == nil {
		return t
	}
	return time.Time{}
}

func unixMilli(ms int64) time.Time {
	if ms <= 0 {
		return time.Time{}
	}
	return time.UnixMilli(ms)
}

func trimForDisplay(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) <= limit {
		return value
	}
	if limit <= 3 {
		return value[:limit]
	}
	return value[:limit-3] + "..."
}
