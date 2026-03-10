package sessionstore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type antigravityTaskMetadata struct {
	Summary   string `json:"summary"`
	UpdatedAt string `json:"updatedAt"`
}

type antigravityTaskSession struct {
	ID               string
	BrainDir         string
	TaskPath         string
	ConversationPath string
	UpdatedAt        string
	Summary          string
}

func DiscoverAntigravityConversationSessions() ([]ConversationSession, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	brainRoot := filepath.Join(home, ".gemini", "antigravity", "brain")
	if !pathExists(brainRoot) {
		return nil, nil
	}
	verbosef("reading Antigravity task sessions from %s", brainRoot)

	dirs, err := filepath.Glob(filepath.Join(brainRoot, "*"))
	if err != nil {
		return nil, err
	}

	out := []ConversationSession{}
	for _, dir := range dirs {
		taskPath := filepath.Join(dir, "task.md")
		if !pathExists(taskPath) {
			continue
		}
		taskBytes, err := os.ReadFile(taskPath)
		if err != nil {
			return nil, err
		}
		info, err := os.Stat(taskPath)
		if err != nil {
			return nil, err
		}
		id := filepath.Base(dir)
		metadata := antigravityTaskMetadata{}
		metaPath := filepath.Join(dir, "task.md.metadata.json")
		if pathExists(metaPath) {
			if data, err := os.ReadFile(metaPath); err == nil {
				_ = json.Unmarshal(data, &metadata)
			}
		}
		conversationPath := filepath.Join(home, ".gemini", "antigravity", "conversations", id+".pb")
		title := antigravityTaskTitle(string(taskBytes))
		if title == "" {
			title = id
		}
		updatedAt := parseOpenClawTimestamp(metadata.UpdatedAt)
		if updatedAt.IsZero() {
			updatedAt = info.ModTime()
		}
		size := pathSize(dir)
		if pathExists(conversationPath) {
			size += pathSize(conversationPath)
		}
		stored := antigravityTaskSession{
			ID:               id,
			BrainDir:         dir,
			TaskPath:         taskPath,
			ConversationPath: conversationPath,
			UpdatedAt:        metadata.UpdatedAt,
			Summary:          metadata.Summary,
		}
		out = append(out, ConversationSession{
			Assistant:    "antigravity",
			ID:           id,
			Title:        title,
			Subtitle:     strings.TrimSpace(metadata.Summary),
			Source:       "Antigravity task artifacts",
			Path:         taskPath,
			UpdatedAt:    updatedAt,
			SizeBytes:    size,
			Deletable:    false,
			DeleteNote:   "Antigravity session indexes are protobuf-backed, so this view is preview-only until index cleanup is implemented",
			ProviderData: stored,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].SortTime().After(out[j].SortTime())
	})
	return out, nil
}

func PreviewAntigravityConversationSession(session ConversationSession) (string, error) {
	stored, ok := session.ProviderData.(antigravityTaskSession)
	if !ok {
		return "", errUnexpectedSessionProviderData("antigravity", session.ProviderData)
	}
	data, err := os.ReadFile(stored.TaskPath)
	if err != nil {
		return "", err
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) > 24 {
		lines = lines[:24]
	}
	if stored.Summary != "" {
		lines = append(lines, "", "Summary:", stored.Summary)
	}
	return strings.Join(lines, "\n"), nil
}

func antigravityTaskTitle(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimLeft(line, "#*-0123456789. ")
		if line != "" {
			return line
		}
	}
	return ""
}
