package sessionstore

import (
	"strings"
	"time"
)

type ConversationSession struct {
	Assistant    string
	ID           string
	Title        string
	Subtitle     string
	Source       string
	Path         string
	StartedAt    time.Time
	UpdatedAt    time.Time
	SizeBytes    int64
	MessageCount int
	InputTokens  int64
	OutputTokens int64
	TotalTokens  int64
	Deletable    bool
	DeleteNote   string
	ProviderData any
}

func (s ConversationSession) SortTime() time.Time {
	if !s.UpdatedAt.IsZero() {
		return s.UpdatedAt
	}
	return s.StartedAt
}

func (s ConversationSession) DisplayLabel() string {
	switch {
	case strings.TrimSpace(s.Title) != "":
		return s.Title
	case strings.TrimSpace(s.Subtitle) != "":
		return s.Subtitle
	case strings.TrimSpace(s.Source) != "":
		return s.Source
	default:
		return s.ID
	}
}

func (s ConversationSession) ShortLabel() string {
	label := s.DisplayLabel()
	if strings.TrimSpace(s.Subtitle) != "" && s.Subtitle != s.Title {
		label = label + " · " + s.Subtitle
	}
	return trimForDisplay(label, 64)
}

func (s ConversationSession) DeleteExplanation() string {
	if strings.TrimSpace(s.DeleteNote) == "" {
		return "index cleanup is not implemented for this session format"
	}
	return s.DeleteNote
}
