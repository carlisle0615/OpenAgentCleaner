package sessionstore

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type cursorComposerData struct {
	ComposerID                  string `json:"composerId"`
	Name                        string `json:"name"`
	Subtitle                    string `json:"subtitle"`
	CreatedAt                   int64  `json:"createdAt"`
	LastUpdatedAt               int64  `json:"lastUpdatedAt"`
	IsArchived                  bool   `json:"isArchived"`
	FullConversationHeadersOnly []struct {
		BubbleID string `json:"bubbleId"`
		Type     int    `json:"type"`
	} `json:"fullConversationHeadersOnly"`
}

type cursorBubbleData struct {
	Type  int    `json:"type"`
	Text  string `json:"text"`
	Token struct {
		InputTokens  int64 `json:"inputTokens"`
		OutputTokens int64 `json:"outputTokens"`
	} `json:"tokenCount"`
}

type cursorSession struct {
	DBPath        string
	ComposerKey   string
	ComposerID    string
	Name          string
	Subtitle      string
	CreatedAt     int64
	LastUpdatedAt int64
	BubbleIDs     []string
	MessageCount  int
	SizeBytes     int64
	IsArchived    bool
}

func DiscoverCursorConversationSessions() ([]ConversationSession, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dbPath := filepath.Join(home, "Library", "Application Support", "Cursor", "User", "globalStorage", "state.vscdb")
	if !pathExists(dbPath) {
		return nil, nil
	}
	verbosef("reading Cursor session state from %s", dbPath)

	db, err := OpenSQLiteDB(dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	composerLower, composerUpper := sqliteKeyPrefixRange("composerData:")
	rows, err := db.Query(`
		SELECT key, CAST(value AS TEXT)
		FROM cursorDiskKV
		WHERE key >= ? AND key < ?
	`, composerLower, composerUpper)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := []ConversationSession{}
	for rows.Next() {
		var key string
		var raw sql.NullString
		if err := rows.Scan(&key, &raw); err != nil {
			return nil, err
		}
		if !raw.Valid || strings.TrimSpace(raw.String) == "" {
			continue
		}
		var composer cursorComposerData
		if err := json.Unmarshal([]byte(raw.String), &composer); err != nil {
			continue
		}
		if composer.ComposerID == "" {
			continue
		}
		stored := cursorSession{
			DBPath:        dbPath,
			ComposerKey:   key,
			ComposerID:    composer.ComposerID,
			Name:          strings.TrimSpace(composer.Name),
			Subtitle:      strings.TrimSpace(composer.Subtitle),
			CreatedAt:     composer.CreatedAt,
			LastUpdatedAt: composer.LastUpdatedAt,
			IsArchived:    composer.IsArchived,
		}
		for _, header := range composer.FullConversationHeadersOnly {
			if header.BubbleID == "" {
				continue
			}
			stored.BubbleIDs = append(stored.BubbleIDs, header.BubbleID)
		}
		stored.MessageCount = len(stored.BubbleIDs)
		stored.SizeBytes, _ = cursorSessionSize(db, stored.ComposerID)

		title := stored.Name
		if title == "" {
			title, _ = firstCursorBubbleText(db, stored.ComposerID, stored.BubbleIDs)
		}
		if title == "" {
			title = stored.ComposerID
		}

		sessions = append(sessions, ConversationSession{
			Assistant:    "cursor",
			ID:           stored.ComposerID,
			Title:        title,
			Subtitle:     stored.Subtitle,
			Source:       "Cursor Composer",
			Path:         dbPath,
			StartedAt:    unixTimeAuto(stored.CreatedAt),
			UpdatedAt:    unixTimeAuto(stored.LastUpdatedAt),
			SizeBytes:    stored.SizeBytes,
			MessageCount: stored.MessageCount,
			Deletable:    true,
			ProviderData: stored,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].SortTime().After(sessions[j].SortTime())
	})
	return sessions, nil
}

func cursorSessionSize(db *sql.DB, composerID string) (int64, error) {
	bubblePrefix := "bubbleId:" + composerID + ":"
	contextPrefix := "messageRequestContext:" + composerID + ":"
	bubbleLower, bubbleUpper := sqliteKeyPrefixRange(bubblePrefix)
	contextLower, contextUpper := sqliteKeyPrefixRange(contextPrefix)
	var total int64
	if err := db.QueryRow(`
		SELECT COALESCE(SUM(LENGTH(value)), 0)
		FROM cursorDiskKV
		WHERE key = ?
			OR (key >= ? AND key < ?)
			OR (key >= ? AND key < ?)
	`,
		"composerData:"+composerID,
		bubbleLower, bubbleUpper,
		contextLower, contextUpper,
	).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func firstCursorBubbleText(db *sql.DB, composerID string, bubbleIDs []string) (string, error) {
	for _, bubbleID := range bubbleIDs {
		text, _, err := readCursorBubble(db, composerID, bubbleID)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(text) != "" {
			return firstNonEmptyLine(text), nil
		}
	}
	return "", nil
}

func PreviewCursorConversationSession(session ConversationSession) (string, error) {
	stored, ok := session.ProviderData.(cursorSession)
	if !ok {
		return "", errUnexpectedSessionProviderData("cursor", session.ProviderData)
	}
	db, err := OpenSQLiteDB(stored.DBPath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	messages := []string{}
	for _, bubbleID := range stored.BubbleIDs {
		text, role, err := readCursorBubble(db, stored.ComposerID, bubbleID)
		if err != nil || strings.TrimSpace(text) == "" {
			continue
		}
		messages = append(messages, fmt.Sprintf("== %s ==\n%s", role, text))
	}
	if len(messages) == 0 {
		return "No text messages found in this conversation.", nil
	}
	if len(messages) > 6 {
		messages = messages[len(messages)-6:]
	}
	return strings.Join(messages, "\n\n"), nil
}

func readCursorBubble(db *sql.DB, composerID, bubbleID string) (string, string, error) {
	var raw sql.NullString
	err := db.QueryRow(`SELECT CAST(value AS TEXT) FROM cursorDiskKV WHERE key = ?`, "bubbleId:"+composerID+":"+bubbleID).Scan(&raw)
	if err == sql.ErrNoRows {
		return "", "", nil
	}
	if err != nil {
		return "", "", err
	}
	if !raw.Valid || strings.TrimSpace(raw.String) == "" {
		return "", "", nil
	}
	var bubble cursorBubbleData
	if err := json.Unmarshal([]byte(raw.String), &bubble); err != nil {
		return "", "", err
	}
	role := "Assistant"
	if bubble.Type == 1 {
		role = "User"
	}
	return strings.TrimSpace(bubble.Text), role, nil
}

func DeleteCursorConversationSessions(sessions []ConversationSession) error {
	grouped := map[string][]cursorSession{}
	for _, session := range sessions {
		stored, ok := session.ProviderData.(cursorSession)
		if !ok {
			return errUnexpectedSessionProviderData("cursor", session.ProviderData)
		}
		grouped[stored.DBPath] = append(grouped[stored.DBPath], stored)
	}

	for dbPath, batch := range grouped {
		db, err := OpenSQLiteDB(dbPath)
		if err != nil {
			return err
		}
		tx, err := db.Begin()
		if err != nil {
			db.Close()
			return err
		}
		for _, item := range batch {
			bubblePrefix := "bubbleId:" + item.ComposerID + ":"
			contextPrefix := "messageRequestContext:" + item.ComposerID + ":"
			bubbleLower, bubbleUpper := sqliteKeyPrefixRange(bubblePrefix)
			contextLower, contextUpper := sqliteKeyPrefixRange(contextPrefix)
			if _, err := tx.Exec(`
				DELETE FROM cursorDiskKV
				WHERE key = ?
					OR (key >= ? AND key < ?)
					OR (key >= ? AND key < ?)
			`,
				"composerData:"+item.ComposerID,
				bubbleLower, bubbleUpper,
				contextLower, contextUpper,
			); err != nil {
				tx.Rollback()
				db.Close()
				return fmt.Errorf("delete cursor session %s: %w", item.ComposerID, err)
			}
		}
		if err := tx.Commit(); err != nil {
			db.Close()
			return err
		}
		db.Close()
	}
	return nil
}

func sqliteKeyPrefixRange(prefix string) (string, string) {
	return prefix, prefix + "\xff"
}
