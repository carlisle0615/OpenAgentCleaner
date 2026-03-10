package cleaner

import (
	"fmt"
	"sort"
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

type conversationProvider struct {
	discover              func() ([]ConversationSession, error)
	preview               func(ConversationSession) (string, error)
	delete                func([]ConversationSession) error
	ignoredCandidateKinds map[string]struct{}
}

func conversationProviderForAssistant(assistant string) (conversationProvider, bool) {
	switch assistant {
	case "openclaw":
		return newOpenClawConversationProvider(), true
	case "codex":
		return newCodexConversationProvider("codex"), true
	case "codex-cli":
		return newCodexConversationProvider("codex-cli"), true
	case "claudecode":
		return newClaudeCodeConversationProvider(), true
	case "cursor":
		return newCursorConversationProvider(), true
	case "antigravity":
		return newAntigravityConversationProvider(), true
	default:
		return conversationProvider{}, false
	}
}

func assistantSupportsSessions(assistant string) bool {
	_, ok := conversationProviderForAssistant(assistant)
	return ok
}

func assistantSupportsSessionDelete(assistant string) bool {
	provider, ok := conversationProviderForAssistant(assistant)
	return ok && provider.delete != nil
}

func discoverAssistantSessions(assistant string) ([]ConversationSession, error) {
	provider, ok := conversationProviderForAssistant(assistant)
	if !ok {
		return nil, nil
	}
	verbosef("scanning conversation sessions for %s", displayAssistant(assistant))
	sessions, err := provider.discover()
	if err != nil {
		return nil, err
	}
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].SortTime().After(sessions[j].SortTime())
	})
	return sessions, nil
}

func previewConversationSession(session ConversationSession) (string, error) {
	provider, ok := conversationProviderForAssistant(session.Assistant)
	if !ok || provider.preview == nil {
		return "", fmt.Errorf("%s sessions do not support previews", displayAssistant(session.Assistant))
	}
	return provider.preview(session)
}

func deleteConversationSessions(sessions []ConversationSession) error {
	grouped := map[string][]ConversationSession{}
	for _, session := range sessions {
		if !session.Deletable {
			label := session.DisplayLabel()
			if label == "" {
				label = session.ID
			}
			return fmt.Errorf("%s cannot be deleted safely: %s", label, session.DeleteExplanation())
		}
		grouped[session.Assistant] = append(grouped[session.Assistant], session)
	}

	for assistant, batch := range grouped {
		provider, ok := conversationProviderForAssistant(assistant)
		if !ok || provider.delete == nil {
			return fmt.Errorf("%s sessions do not support deletion", displayAssistant(assistant))
		}
		if err := provider.delete(batch); err != nil {
			return err
		}
	}
	return nil
}

func sessionIgnoredCandidateKinds(assistant string) map[string]struct{} {
	provider, ok := conversationProviderForAssistant(assistant)
	if !ok {
		return nil
	}
	return provider.ignoredCandidateKinds
}

func filterSessionsBefore(sessions []ConversationSession, cutoff time.Time) []ConversationSession {
	out := make([]ConversationSession, 0, len(sessions))
	for _, session := range sessions {
		if session.SortTime().Before(cutoff) {
			out = append(out, session)
		}
	}
	return out
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

func unixTimeAuto(raw int64) time.Time {
	if raw <= 0 {
		return time.Time{}
	}
	if raw > 1_000_000_000_000 {
		return time.UnixMilli(raw)
	}
	return time.Unix(raw, 0)
}
