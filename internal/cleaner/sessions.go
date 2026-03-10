package cleaner

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/carlisle0615/OpenAgentCleaner/internal/cleaner/sessionstore"
)

type ConversationSession = sessionstore.ConversationSession

func assistantSupportsSessions(assistant string) bool {
	switch assistant {
	case "openclaw", "codex", "codex-cli", "claudecode", "cursor", "antigravity":
		return true
	default:
		return false
	}
}

func assistantSupportsSessionDelete(assistant string) bool {
	switch assistant {
	case "openclaw", "codex", "codex-cli", "claudecode", "cursor":
		return true
	default:
		return false
	}
}

func discoverAssistantSessions(assistant string) ([]ConversationSession, error) {
	if !assistantSupportsSessions(assistant) {
		return nil, nil
	}
	verbosef("scanning conversation sessions for %s", displayAssistant(assistant))
	switch assistant {
	case "openclaw":
		return discoverOpenClawConversationSessions()
	case "codex":
		return sessionstore.DiscoverCodexConversationSessions("codex")
	case "codex-cli":
		return sessionstore.DiscoverCodexConversationSessions("codex-cli")
	case "claudecode":
		return sessionstore.DiscoverClaudeCodeConversationSessions()
	case "cursor":
		return sessionstore.DiscoverCursorConversationSessions()
	case "antigravity":
		return sessionstore.DiscoverAntigravityConversationSessions()
	default:
		return nil, nil
	}
}

func previewConversationSession(session ConversationSession) (string, error) {
	switch session.Assistant {
	case "openclaw":
		return previewOpenClawConversationSession(session)
	case "codex", "codex-cli":
		return sessionstore.PreviewCodexConversationSession(session)
	case "claudecode":
		return sessionstore.PreviewClaudeCodeConversationSession(session)
	case "cursor":
		return sessionstore.PreviewCursorConversationSession(session)
	case "antigravity":
		return sessionstore.PreviewAntigravityConversationSession(session)
	default:
		return "", fmt.Errorf("%s sessions do not support previews", displayAssistant(session.Assistant))
	}
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
		switch assistant {
		case "openclaw":
			if err := deleteOpenClawConversationSessions(batch); err != nil {
				return err
			}
		case "codex", "codex-cli":
			if err := sessionstore.DeleteCodexConversationSessions(batch); err != nil {
				return err
			}
		case "claudecode":
			if err := sessionstore.DeleteClaudeCodeConversationSessions(batch); err != nil {
				return err
			}
		case "cursor":
			if err := sessionstore.DeleteCursorConversationSessions(batch); err != nil {
				return err
			}
		default:
			return fmt.Errorf("%s sessions do not support deletion", displayAssistant(assistant))
		}
	}
	return nil
}

func sessionIgnoredCandidateKinds(assistant string) map[string]struct{} {
	switch assistant {
	case "openclaw":
		return map[string]struct{}{
			"session_store": {},
		}
	default:
		return sessionstore.IgnoredCandidateKinds(assistant)
	}
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

func openSQLiteDB(path string) (*sql.DB, error) {
	return sessionstore.OpenSQLiteDB(path)
}

func classifyCodexAssistantFromSource(source string) (string, bool) {
	return sessionstore.ClassifyCodexAssistantFromSource(source)
}
