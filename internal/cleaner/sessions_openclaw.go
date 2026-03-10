package cleaner

func discoverOpenClawConversationSessions() ([]ConversationSession, error) {
	sessions, err := discoverOpenClawSessions()
	if err != nil {
		return nil, err
	}
	out := make([]ConversationSession, 0, len(sessions))
	for _, session := range sessions {
		out = append(out, ConversationSession{
			Assistant:    "openclaw",
			ID:           session.SessionID,
			Title:        session.DisplayName,
			Subtitle:     session.AgentID,
			Source:       session.Source,
			Path:         session.TranscriptPath,
			StartedAt:    session.StartedAt,
			UpdatedAt:    session.UpdatedAt,
			SizeBytes:    session.SizeBytes,
			MessageCount: session.MessageCount,
			InputTokens:  session.InputTokens,
			OutputTokens: session.OutputTokens,
			TotalTokens:  session.TotalTokens,
			Deletable:    true,
			ProviderData: session,
		})
	}
	return out, nil
}

func previewOpenClawConversationSession(session ConversationSession) (string, error) {
	value, ok := session.ProviderData.(OpenClawSession)
	if !ok {
		return "", errUnexpectedSessionProviderData("openclaw", session.ProviderData)
	}
	return previewOpenClawSession(value.TranscriptPath)
}

func deleteOpenClawConversationSessions(sessions []ConversationSession) error {
	batch := make([]OpenClawSession, 0, len(sessions))
	for _, session := range sessions {
		value, ok := session.ProviderData.(OpenClawSession)
		if !ok {
			return errUnexpectedSessionProviderData("openclaw", session.ProviderData)
		}
		batch = append(batch, value)
	}
	return deleteOpenClawSessions(batch)
}
