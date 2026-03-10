package sessionstore

func IgnoredCandidateKinds(assistant string) map[string]struct{} {
	switch assistant {
	case "codex", "codex-cli":
		return map[string]struct{}{
			"session_store":     {},
			"archived_sessions": {},
			"session_index":     {},
			"session_db":        {},
		}
	case "claudecode":
		return map[string]struct{}{
			"transcripts":      {},
			"project_sessions": {},
		}
	case "cursor":
		return map[string]struct{}{
			"global_chat_state":    {},
			"workspace_chat_state": {},
		}
	default:
		return nil
	}
}
