package sessionstore

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type codexStoredSession struct {
	ThreadID         string
	RolloutPath      string
	DBPath           string
	SessionIndexPath string
	Assistant        string
	Title            string
	FirstUserMessage string
	Source           string
	Originator       string
	TokensUsed       int64
	StartedAt        time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type codexThreadRow struct {
	ID               string
	RolloutPath      string
	CreatedAt        int64
	UpdatedAt        int64
	Source           string
	Title            string
	TokensUsed       int64
	FirstUserMessage string
}

type codexRolloutEnvelope struct {
	Timestamp string          `json:"timestamp"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
}

type codexSessionMetaPayload struct {
	ID         string `json:"id"`
	Timestamp  string `json:"timestamp"`
	Originator string `json:"originator"`
	Source     string `json:"source"`
}

type codexResponseItemPayload struct {
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

type codexEventPayload struct {
	Type    string `json:"type"`
	Text    string `json:"text"`
	Message string `json:"message"`
}

func DiscoverCodexConversationSessions(targetAssistant string) ([]ConversationSession, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dbPath, err := latestMatchingPath(filepath.Join(home, ".codex", "state_*.sqlite"))
	if err != nil {
		return nil, err
	}
	if dbPath == "" || !pathExists(dbPath) {
		return nil, nil
	}
	verbosef("reading %s session database: %s", displayAssistantName(targetAssistant), dbPath)

	db, err := OpenSQLiteDB(dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query, args := codexThreadsQuery(targetAssistant)
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexPath := filepath.Join(home, ".codex", "session_index.jsonl")
	out := []ConversationSession{}
	for rows.Next() {
		var row codexThreadRow
		if err := rows.Scan(
			&row.ID,
			&row.RolloutPath,
			&row.CreatedAt,
			&row.UpdatedAt,
			&row.Source,
			&row.Title,
			&row.TokensUsed,
			&row.FirstUserMessage,
		); err != nil {
			return nil, err
		}
		if !pathExists(row.RolloutPath) {
			continue
		}

		assistant, known := ClassifyCodexAssistantFromSource(row.Source)
		rolloutMeta := codexStoredSession{}
		messageCount := 0
		if !known {
			var err error
			rolloutMeta, messageCount, err = scanCodexRolloutMetadata(row.RolloutPath)
			if err != nil {
				return nil, err
			}
			assistant = classifyCodexAssistant(row.Source, rolloutMeta.Originator)
		}
		if assistant != targetAssistant {
			continue
		}

		info, statErr := os.Stat(row.RolloutPath)
		modTime := time.Time{}
		if statErr == nil {
			modTime = info.ModTime()
		}
		startedAt := rolloutMeta.StartedAt
		if startedAt.IsZero() {
			startedAt = unixTimeAuto(row.CreatedAt)
		}
		updatedAt := unixTimeAuto(row.UpdatedAt)
		if updatedAt.IsZero() {
			updatedAt = modTime
		}

		stored := codexStoredSession{
			ThreadID:         row.ID,
			RolloutPath:      row.RolloutPath,
			DBPath:           dbPath,
			SessionIndexPath: indexPath,
			Assistant:        assistant,
			Title:            strings.TrimSpace(row.Title),
			FirstUserMessage: strings.TrimSpace(row.FirstUserMessage),
			Source:           row.Source,
			Originator:       rolloutMeta.Originator,
			TokensUsed:       row.TokensUsed,
			StartedAt:        startedAt,
			CreatedAt:        unixTimeAuto(row.CreatedAt),
			UpdatedAt:        updatedAt,
		}

		title := strings.TrimSpace(row.Title)
		if title == "" {
			title = firstNonEmptyLine(row.FirstUserMessage)
		}
		if title == "" {
			title = row.ID
		}

		out = append(out, ConversationSession{
			Assistant:    assistant,
			ID:           row.ID,
			Title:        title,
			Subtitle:     codexSessionSubtitle(stored),
			Source:       codexSessionSource(stored),
			Path:         row.RolloutPath,
			StartedAt:    startedAt,
			UpdatedAt:    updatedAt,
			SizeBytes:    pathSize(row.RolloutPath),
			MessageCount: messageCount,
			TotalTokens:  row.TokensUsed,
			Deletable:    true,
			ProviderData: stored,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].SortTime().After(out[j].SortTime())
	})
	return out, nil
}

func PreviewCodexConversationSession(session ConversationSession) (string, error) {
	stored, ok := session.ProviderData.(codexStoredSession)
	if !ok {
		return "", errUnexpectedSessionProviderData("codex", session.ProviderData)
	}
	return previewCodexRollout(stored)
}

func DeleteCodexConversationSessions(sessions []ConversationSession) error {
	grouped := map[string][]codexStoredSession{}
	for _, session := range sessions {
		stored, ok := session.ProviderData.(codexStoredSession)
		if !ok {
			return errUnexpectedSessionProviderData(session.Assistant, session.ProviderData)
		}
		grouped[stored.DBPath] = append(grouped[stored.DBPath], stored)
	}

	for dbPath, batch := range grouped {
		if err := deleteCodexConversationBatch(dbPath, batch); err != nil {
			return err
		}
	}
	return nil
}

func deleteCodexConversationBatch(dbPath string, batch []codexStoredSession) error {
	if len(batch) == 0 {
		return nil
	}

	ids := make([]string, 0, len(batch))
	rollouts := make([]string, 0, len(batch))
	indexPath := batch[0].SessionIndexPath
	for _, item := range batch {
		ids = append(ids, item.ThreadID)
		rollouts = append(rollouts, item.RolloutPath)
	}

	indexLines, err := filterCodexSessionIndex(indexPath, ids)
	if err != nil {
		return err
	}

	dbBackups, err := backupSQLiteFiles(dbPath)
	if err != nil {
		return err
	}
	indexBackups, err := backupFileIfExists(indexPath)
	if err != nil {
		_ = cleanupCopiedBackups(dbBackups)
		return err
	}
	stagedRollouts, err := stageDeletePaths(rollouts)
	if err != nil {
		_ = cleanupCopiedBackups(indexBackups)
		_ = cleanupCopiedBackups(dbBackups)
		return err
	}

	rollback := func(cause error) error {
		if restoreErr := restoreCopiedBackups(indexBackups); restoreErr != nil {
			return fmt.Errorf("%w; restore Codex session index: %v", cause, restoreErr)
		}
		if restoreErr := restoreCopiedBackups(dbBackups); restoreErr != nil {
			return fmt.Errorf("%w; restore Codex session database: %v", cause, restoreErr)
		}
		if restoreErr := restoreStagedDeletes(stagedRollouts); restoreErr != nil {
			return fmt.Errorf("%w; restore Codex rollout files: %v", cause, restoreErr)
		}
		_ = cleanupCopiedBackups(indexBackups)
		_ = cleanupCopiedBackups(dbBackups)
		return cause
	}

	db, err := OpenSQLiteDB(dbPath)
	if err != nil {
		_ = restoreStagedDeletes(stagedRollouts)
		_ = cleanupCopiedBackups(indexBackups)
		_ = cleanupCopiedBackups(dbBackups)
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		db.Close()
		_ = restoreStagedDeletes(stagedRollouts)
		_ = cleanupCopiedBackups(indexBackups)
		_ = cleanupCopiedBackups(dbBackups)
		return err
	}

	if err := deleteCodexLogs(tx, ids); err != nil {
		tx.Rollback()
		db.Close()
		return rollback(err)
	}
	if err := deleteCodexThreads(tx, ids); err != nil {
		tx.Rollback()
		db.Close()
		return rollback(err)
	}
	if err := tx.Commit(); err != nil {
		db.Close()
		return rollback(err)
	}
	db.Close()

	if err := writeJSONLLinesAtomic(indexPath, indexLines); err != nil {
		return rollback(err)
	}
	if err := cleanupCopiedBackups(indexBackups); err != nil {
		return err
	}
	if err := cleanupCopiedBackups(dbBackups); err != nil {
		return err
	}
	if err := cleanupStagedDeletes(stagedRollouts); err != nil {
		return err
	}
	return nil
}

func deleteCodexLogs(tx *sql.Tx, ids []string) error {
	query, args := sqlDeleteIn("DELETE FROM logs WHERE thread_id IN ", ids)
	if _, err := tx.Exec(query, args...); err != nil {
		return fmt.Errorf("delete codex logs: %w", err)
	}
	return nil
}

func deleteCodexThreads(tx *sql.Tx, ids []string) error {
	query, args := sqlDeleteIn("DELETE FROM threads WHERE id IN ", ids)
	if _, err := tx.Exec(query, args...); err != nil {
		return fmt.Errorf("delete codex threads: %w", err)
	}
	return nil
}

func filterCodexSessionIndex(path string, removeIDs []string) ([]string, error) {
	if !pathExists(path) {
		return nil, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	removeSet := map[string]struct{}{}
	for _, id := range removeIDs {
		removeSet[id] = struct{}{}
	}

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	lines := []string{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var item map[string]any
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			return nil, err
		}
		id, _ := item["id"].(string)
		if _, drop := removeSet[id]; drop {
			continue
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func writeJSONLLinesAtomic(path string, lines []string) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	var buf bytes.Buffer
	for _, line := range lines {
		buf.WriteString(strings.TrimRight(line, "\n"))
		buf.WriteByte('\n')
	}
	temp := path + ".tmp"
	if err := os.WriteFile(temp, buf.Bytes(), 0o600); err != nil {
		return err
	}
	return os.Rename(temp, path)
}

func previewCodexRollout(session codexStoredSession) (string, error) {
	file, err := os.Open(session.RolloutPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	messages := []string{}
	if session.FirstUserMessage != "" {
		messages = append(messages, "== User ==\n"+strings.TrimSpace(session.FirstUserMessage))
	}

	for scanner.Scan() {
		var env codexRolloutEnvelope
		if err := json.Unmarshal(scanner.Bytes(), &env); err != nil {
			continue
		}
		switch env.Type {
		case "response_item":
			var payload codexResponseItemPayload
			if err := json.Unmarshal(env.Payload, &payload); err != nil || payload.Type != "message" {
				continue
			}
			text := collectCodexResponseText(payload)
			if text == "" {
				continue
			}
			role := strings.Title(payload.Role)
			if role == "" {
				role = "Assistant"
			}
			messages = append(messages, "== "+role+" ==\n"+text)
		case "event_msg":
			var payload codexEventPayload
			if err := json.Unmarshal(env.Payload, &payload); err != nil {
				continue
			}
			if payload.Type == "agent_message" && strings.TrimSpace(payload.Message) != "" {
				messages = append(messages, "== Assistant ==\n"+strings.TrimSpace(payload.Message))
			}
		}
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

func scanCodexRolloutMetadata(path string) (codexStoredSession, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return codexStoredSession{}, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	meta := codexStoredSession{}
	var messages int
	for scanner.Scan() {
		var env codexRolloutEnvelope
		if err := json.Unmarshal(scanner.Bytes(), &env); err != nil {
			continue
		}
		if env.Type == "session_meta" {
			var payload codexSessionMetaPayload
			if err := json.Unmarshal(env.Payload, &payload); err == nil {
				meta.Originator = payload.Originator
				meta.StartedAt = parseOpenClawTimestamp(payload.Timestamp)
			}
			continue
		}
		if env.Type == "response_item" {
			var payload codexResponseItemPayload
			if err := json.Unmarshal(env.Payload, &payload); err == nil && payload.Type == "message" && collectCodexResponseText(payload) != "" {
				messages++
			}
			continue
		}
		if env.Type == "event_msg" {
			var payload codexEventPayload
			if err := json.Unmarshal(env.Payload, &payload); err == nil && payload.Type == "agent_message" && strings.TrimSpace(payload.Message) != "" {
				messages++
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return codexStoredSession{}, 0, err
	}
	return meta, messages, nil
}

func collectCodexResponseText(payload codexResponseItemPayload) string {
	parts := []string{}
	for _, item := range payload.Content {
		if strings.TrimSpace(item.Text) == "" {
			continue
		}
		if item.Type == "output_text" || item.Type == "text" || item.Type == "" {
			parts = append(parts, strings.TrimSpace(item.Text))
		}
	}
	return strings.Join(parts, "\n")
}

func classifyCodexAssistant(source, originator string) string {
	if assistant, ok := ClassifyCodexAssistantFromSource(source); ok {
		return assistant
	}
	normalizedOriginator := strings.ToLower(strings.TrimSpace(originator))
	if strings.Contains(normalizedOriginator, "desktop") ||
		strings.Contains(normalizedOriginator, "vscode") {
		return "codex"
	}
	return "codex-cli"
}

func ClassifyCodexAssistantFromSource(source string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case "vscode":
		return "codex", true
	case "cli", "exec", "mcp":
		return "codex-cli", true
	default:
		return "", false
	}
}

func codexThreadsQuery(targetAssistant string) (string, []any) {
	base := `
		SELECT id, rollout_path, created_at, updated_at, source, title, tokens_used, first_user_message
		FROM threads
	`
	switch targetAssistant {
	case "codex":
		return base + `
			WHERE source IN ('vscode', 'unknown')
			ORDER BY updated_at DESC, id DESC
		`, nil
	case "codex-cli":
		return base + `
			WHERE source != 'vscode'
			ORDER BY updated_at DESC, id DESC
		`, nil
	default:
		return base + `
			ORDER BY updated_at DESC, id DESC
		`, nil
	}
}

func codexSessionSubtitle(session codexStoredSession) string {
	switch {
	case session.Assistant == "codex" && session.Originator != "":
		return session.Originator
	case session.Source != "":
		return session.Source
	default:
		return ""
	}
}

func codexSessionSource(session codexStoredSession) string {
	parts := []string{}
	if session.Originator != "" {
		parts = append(parts, session.Originator)
	}
	if session.Source != "" && !strings.EqualFold(session.Source, session.Originator) {
		parts = append(parts, session.Source)
	}
	return strings.Join(uniqueStrings(parts), " · ")
}

func sqlDeleteIn(prefix string, ids []string) (string, []any) {
	holders := make([]string, 0, len(ids))
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		holders = append(holders, "?")
		args = append(args, id)
	}
	return prefix + "(" + strings.Join(holders, ",") + ")", args
}
