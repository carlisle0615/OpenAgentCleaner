package cleaner

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func testConversationSession(id, title, path string, startedAt time.Time) ConversationSession {
	return ConversationSession{
		Assistant:    "openclaw",
		ID:           id,
		Title:        title,
		Path:         path,
		StartedAt:    startedAt,
		UpdatedAt:    startedAt,
		Deletable:    true,
		ProviderData: OpenClawSession{SessionID: id, DisplayName: title, TranscriptPath: path, StartedAt: startedAt},
	}
}

func TestNewAnalyzeModelAndNavigation(t *testing.T) {
	home := setTestHome(t)
	openclawRoot := filepath.Join(home, ".openclaw")
	writeTestFile(t, filepath.Join(openclawRoot, "logs", "app.log"), "log")
	writeTestFile(t, filepath.Join(openclawRoot, "agents", "main", "sessions", "sessions.json"), `{"one":{"sessionId":"session-1","updatedAt":1700000000000}}`)
	writeTestFile(t, filepath.Join(openclawRoot, "agents", "main", "sessions", "session-1.jsonl"), "{\"timestamp\":\"2026-02-07T03:16:10.650Z\"}\n")
	writeTestFile(t, filepath.Join(home, ".ollama", "logs", "server.log"), "log")

	model, err := newAnalyzeModel([]string{"openclaw", "ollama"}, time.Time{})
	if err != nil {
		t.Fatalf("newAnalyzeModel(multi) err = %v", err)
	}
	if model.screen != screenAssistants || len(model.summaries) != 2 {
		t.Fatalf("newAnalyzeModel(multi) = %#v", model)
	}

	model.moveCursor(1)
	if model.assistantIndex != 1 {
		t.Fatalf("moveCursor(assistants) = %d", model.assistantIndex)
	}
	model.moveCursor(-5)
	if model.assistantIndex != 0 {
		t.Fatalf("moveCursor(clamped) = %d", model.assistantIndex)
	}
	if err := model.activateSelection(); err != nil {
		t.Fatalf("activateSelection(assistant) err = %v", err)
	}
	if model.screen != screenAssistantMenu {
		t.Fatalf("screen after select = %v", model.screen)
	}

	model.assistantMenuIndex = 1
	if err := model.activateSelection(); err != nil {
		t.Fatalf("activateSelection(openclaw menu) err = %v", err)
	}
	if model.screen != screenCandidates || model.activeAssistant != "openclaw" {
		t.Fatalf("activateSelection(openclaw menu) = %#v", model)
	}

	model.navigateBack()
	if model.screen != screenAssistantMenu {
		t.Fatalf("navigateBack() = %v", model.screen)
	}
	model.navigateBack()
	if model.screen != screenAssistants || !model.atRoot() {
		t.Fatalf("navigateBack(root) = %v", model.screen)
	}

	ollamaModel, err := newAnalyzeModel([]string{"ollama"}, time.Time{})
	if err != nil {
		t.Fatalf("newAnalyzeModel(ollama) err = %v", err)
	}
	if ollamaModel.screen != screenCandidates || !ollamaModel.atRoot() {
		t.Fatalf("newAnalyzeModel(ollama) = %#v", ollamaModel)
	}

	before := time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local)
	openclawModel, err := newAnalyzeModel([]string{"openclaw"}, before)
	if err != nil {
		t.Fatalf("newAnalyzeModel(openclaw) err = %v", err)
	}
	if openclawModel.screen != screenSessions || openclawModel.sessionBefore.IsZero() {
		t.Fatalf("newAnalyzeModel(openclaw before) = %#v", openclawModel)
	}
}

func TestAnalyzeModelKeyHandlingAndDialogs(t *testing.T) {
	model := analyzeModel{
		screen:          screenSessions,
		activeAssistant: "openclaw",
		sessions:        []ConversationSession{testConversationSession("a", "A", "/tmp/a", time.Unix(1, 0))},
		candidates:      []Candidate{{Path: "/tmp/a", Kind: "logs", Safety: SafetySafe, Reason: "r"}},
		styles:          newAnalyzeStyles(),
		sessionBefore:   time.Time{},
	}

	next, cmd := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	updated := next.(analyzeModel)
	if updated.width != 80 || updated.height != 24 || cmd != nil {
		t.Fatalf("Update(WindowSizeMsg) = %#v, %v", updated, cmd)
	}

	nextModel, cmd := updated.handleKey(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil || nextModel.screen != screenSessions {
		t.Fatalf("handleKey(ctrl+c) = %#v, %v", next, cmd)
	}

	model.screen = screenCandidates
	model.inputMode = inputSessionFilter
	model.inputValue = "2026-03-01"
	nextModel, _ = model.handleInputKey(tea.KeyMsg{Type: tea.KeyEnter})
	if nextModel.inputMode != inputNone {
		t.Fatalf("handleInputKey(enter) should close dialog: %#v", nextModel)
	}

	model.inputMode = inputSessionBulkDelete
	model.inputValue = "bad"
	nextModel, _ = model.handleInputKey(tea.KeyMsg{Type: tea.KeyEnter})
	if !strings.Contains(nextModel.status, "invalid date") {
		t.Fatalf("handleInputKey(bad) = %#v", nextModel)
	}

	model.inputMode = inputSessionFilter
	model.inputValue = "2026-03-01"
	nextModel, _ = model.handleInputKey(tea.KeyMsg{Type: tea.KeyBackspace})
	if nextModel.inputValue != "2026-03-0" {
		t.Fatalf("handleInputKey(backspace) = %q", nextModel.inputValue)
	}
	nextModel, _ = nextModel.handleInputKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1x")})
	if nextModel.inputValue != "2026-03-01" {
		t.Fatalf("handleInputKey(runes) = %q", nextModel.inputValue)
	}
	nextModel, _ = nextModel.handleInputKey(tea.KeyMsg{Type: tea.KeyEsc})
	if nextModel.inputMode != inputNone {
		t.Fatalf("handleInputKey(esc) = %#v", nextModel)
	}

	model.confirmMode = confirmDeleteCandidate
	nextModel, _ = model.handleConfirmKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	if nextModel.confirmMode != confirmNone || nextModel.status != "Deletion cancelled." {
		t.Fatalf("handleConfirmKey(cancel) = %#v", nextModel)
	}

	model.screen = screenAssistants
	model.summaries = []assistantSummary{{Assistant: "openclaw"}, {Assistant: "ollama"}}
	model.inputMode = inputNone
	model.confirmMode = confirmNone
	nextModel, _ = model.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	if nextModel.assistantIndex != 1 {
		t.Fatalf("handleKey(down) = %#v", nextModel)
	}
	nextModel, _ = nextModel.handleKey(tea.KeyMsg{Type: tea.KeyUp})
	if nextModel.assistantIndex != 0 {
		t.Fatalf("handleKey(up) = %#v", nextModel)
	}

	model = analyzeModel{
		screen:          screenSessions,
		activeAssistant: "openclaw",
		styles:          newAnalyzeStyles(),
		sessions:        []ConversationSession{testConversationSession("a", "A", "/tmp/a", time.Unix(1, 0))},
		candidates:      []Candidate{{Path: "/tmp/a", Kind: "logs", Safety: SafetySafe, Reason: "r"}},
		summaries:       []assistantSummary{{Assistant: "openclaw"}},
	}
	nextModel, cmd = model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd != nil {
		t.Fatalf("handleKey(q) should navigate back before root quit: %v", cmd)
	}
	nextModel, _ = model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	if nextModel.inputMode != inputSessionFilter {
		t.Fatalf("handleKey(f) = %#v", nextModel)
	}
	nextModel, _ = model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if nextModel.inputMode != inputSessionBulkDelete {
		t.Fatalf("handleKey(x) = %#v", nextModel)
	}
	model.sessionBefore = time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	nextModel, _ = model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if nextModel.confirmMode != confirmDeleteSessionsBefore {
		t.Fatalf("handleKey(x with cutoff) = %#v", nextModel)
	}
	nextModel, _ = model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	if !nextModel.sessionBefore.IsZero() || !strings.Contains(nextModel.status, "cleared") {
		t.Fatalf("handleKey(c) = %#v", nextModel)
	}

	model = analyzeModel{screen: screenAssistantMenu, styles: newAnalyzeStyles(), assistants: []string{"openclaw"}}
	nextModel, cmd = model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("handleKey(esc at root) should quit")
	}

	model = analyzeModel{screen: screenAssistants, summaries: []assistantSummary{{Assistant: "openclaw", LeftoverCount: 1}}, styles: newAnalyzeStyles()}
	nextModel, _ = model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if nextModel.screen != screenAssistantMenu {
		t.Fatalf("handleKey(enter) = %#v", nextModel)
	}
}

func TestAnalyzeDeleteFlowsAndReloads(t *testing.T) {
	home := setTestHome(t)
	openclawRoot := filepath.Join(home, ".openclaw")
	sessionsDir := filepath.Join(openclawRoot, "agents", "main", "sessions")
	metadataPath := filepath.Join(sessionsDir, "sessions.json")
	transcriptOld := filepath.Join(sessionsDir, "old.jsonl")
	transcriptNew := filepath.Join(sessionsDir, "new.jsonl")
	writeTestFile(t, metadataPath, `{
  "old": {"sessionId":"old","updatedAt":1700000000000},
  "new": {"sessionId":"new","updatedAt":1800000000000}
}`)
	writeTestFile(t, transcriptOld, "{\"timestamp\":\"2025-01-01T00:00:00Z\"}\n")
	writeTestFile(t, transcriptNew, "{\"timestamp\":\"2026-01-01T00:00:00Z\"}\n")

	candidatePath := filepath.Join(home, ".ollama", "logs", "app.log")
	writeTestFile(t, candidatePath, "log")

	model := analyzeModel{styles: newAnalyzeStyles(), activeAssistant: "openclaw", screen: screenSessions}
	if err := model.reloadSessions(); err != nil {
		t.Fatalf("reloadSessions() err = %v", err)
	}
	if len(model.sessions) != 2 {
		t.Fatalf("reloadSessions() = %#v", model.sessions)
	}
	if model.previewText != "" || model.previewSessionID != "" {
		t.Fatal("reloadSessions() should not preload session preview")
	}
	if len(model.sessionsSource()) != 2 {
		t.Fatalf("sessionsSource() should return sessions")
	}

	if err := model.activateSelection(); err != nil {
		t.Fatalf("activateSelection(session preview) err = %v", err)
	}
	if model.screen != screenSessionPreview {
		t.Fatalf("activateSelection(session preview) screen = %v", model.screen)
	}
	if model.previewText == "" || model.previewSessionID == "" {
		t.Fatal("activateSelection(session preview) should load preview on demand")
	}
	model.navigateBack()

	if err := model.prepareDeleteSelected(); err != nil {
		t.Fatalf("prepareDeleteSelected(session) err = %v", err)
	}
	if model.confirmMode != confirmDeleteSession {
		t.Fatalf("prepareDeleteSelected(session) = %#v", model)
	}
	if err := model.executeConfirm(); err != nil {
		t.Fatalf("executeConfirm(session) err = %v", err)
	}
	if pathExists(transcriptNew) {
		t.Fatal("session transcript should be deleted")
	}

	model.clearConfirm()
	model.sessionBefore = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	model.prepareDeleteSessionsBefore(model.sessionBefore)
	if model.confirmMode != confirmDeleteSessionsBefore {
		t.Fatalf("prepareDeleteSessionsBefore() = %#v", model)
	}
	if err := model.executeConfirm(); err != nil {
		t.Fatalf("executeConfirm(before) err = %v", err)
	}

	model.screen = screenCandidates
	model.activeAssistant = "ollama"
	if err := model.reloadCandidates("ollama"); err != nil {
		t.Fatalf("reloadCandidates() err = %v", err)
	}
	var foundLog bool
	for i, candidate := range model.candidates {
		if candidate.Path == filepath.Join(home, ".ollama", "logs") || candidate.Path == candidatePath || strings.Contains(candidate.Path, ".ollama/logs") {
			model.candidateIndex = i
			foundLog = true
			break
		}
	}
	if !foundLog {
		t.Fatalf("reloadCandidates() = %#v", model.candidates)
	}
	if err := model.prepareDeleteSelected(); err != nil {
		t.Fatalf("prepareDeleteSelected(candidate) err = %v", err)
	}
	if err := model.executeConfirm(); err != nil {
		t.Fatalf("executeConfirm(candidate) err = %v", err)
	}
	if pathExists(candidatePath) {
		t.Fatal("candidate path should be deleted")
	}

	model.candidates = nil
	if err := model.prepareDeleteSelected(); err == nil {
		t.Fatal("prepareDeleteSelected() should fail without candidate")
	}
	model.screen = screenAssistantMenu
	if err := model.prepareDeleteSelected(); err == nil {
		t.Fatal("prepareDeleteSelected() should fail on menu screen")
	}
}

func TestAnalyzeRenderingHelpers(t *testing.T) {
	model := analyzeModel{
		width:            100,
		height:           30,
		styles:           newAnalyzeStyles(),
		screen:           screenAssistants,
		summaries:        []assistantSummary{{Assistant: "openclaw", SessionCount: 2, LeftoverCount: 3}},
		candidates:       []Candidate{{Assistant: "ollama", Kind: "models", Safety: SafetyConfirm, Reason: "reason", Path: "/tmp/model", SizeBytes: 10}},
		sessions:         []ConversationSession{{Assistant: "openclaw", ID: "sid", Title: "Title", Subtitle: "worker", Path: "/tmp/sid", StartedAt: time.Unix(1, 0), UpdatedAt: time.Unix(1, 0), TotalTokens: 1000, MessageCount: 2, SizeBytes: 10, Deletable: true}},
		activeAssistant:  "openclaw",
		status:           "Removed one item.",
		previewText:      "== User ==\nHow do I fix this?\n\n== Assistant ==\nTry updating the test.",
		previewSessionID: "sid",
	}

	if model.Init() != nil {
		t.Fatal("Init() should return nil")
	}

	for _, rendered := range []string{
		func() string { left, right := model.renderBody(); return left + right }(),
		model.renderHeader(),
		model.renderFooter(),
		model.renderAssistantList(),
		model.renderAssistantDetail(),
		model.renderAssistantMenu(),
		model.renderAssistantMenuDetail(),
		model.renderSessionsList(),
		model.renderSessionDetail(),
		model.renderCandidateList(),
		model.renderCandidateDetail(),
		model.renderInputDialog(),
		model.renderConfirmDialog(),
		model.View(),
	} {
		if rendered == "" {
			t.Fatal("rendered output should not be empty")
		}
	}
	if !strings.Contains(model.renderSessionDetail(), "How do I fix this?") || !strings.Contains(model.renderSessionDetail(), "Try updating the test.") {
		t.Fatalf("renderSessionDetail() = %q", model.renderSessionDetail())
	}
	model.previewText = ""
	model.previewSessionID = ""
	if !strings.Contains(model.renderSessionDetail(), "Preview is loaded on demand") {
		t.Fatalf("renderSessionDetail() should show on-demand hint = %q", model.renderSessionDetail())
	}
	if !strings.Contains(strings.Join(wrapDisplayText("中文 mixed English content", 8), "\n"), "\n") {
		t.Fatal("wrapDisplayText() should wrap mixed-width content")
	}

	model.screen = screenCandidates
	model.inputMode = inputSessionFilter
	model.inputLabel = "Date"
	model.inputValue = "2026-03-01"
	model.confirmMode = confirmDeleteCandidate
	model.confirmTitle = "Delete item?"
	model.confirmBody = "Body"
	view := model.View()
	for _, want := range []string{"Delete item?", "Body"} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q\n%s", want, view)
		}
	}
	model.confirmMode = confirmNone
	if !strings.Contains(model.View(), "Date") {
		t.Fatalf("View() missing input dialog\n%s", model.View())
	}

	model.screen = screenAssistantMenu
	if left, right := model.renderBody(); left == "" || right == "" {
		t.Fatalf("renderBody(assistant menu) = %q %q", left, right)
	}
	model.screen = screenSessions
	if left, right := model.renderBody(); left == "" || right == "" {
		t.Fatalf("renderBody(sessions) = %q %q", left, right)
	}
	model.screen = screenCandidates
	model.activeAssistant = "ollama"
	model.candidates = []Candidate{
		{Assistant: "ollama", Kind: "auth_key", Safety: SafetyManual, Reason: "manual", Path: "/tmp/key", SizeBytes: 1},
		{Assistant: "ollama", Kind: "logs", Safety: SafetySafe, Reason: "safe", Path: "/tmp/log", SizeBytes: 1},
	}
	model.candidateIndex = 0
	if !strings.Contains(model.renderCandidateDetail(), "Manual inspection only.") {
		t.Fatalf("renderCandidateDetail(manual) = %q", model.renderCandidateDetail())
	}
	model.candidateIndex = 1
	if !strings.Contains(model.renderCandidateDetail(), "delete this item") {
		t.Fatalf("renderCandidateDetail(safe) = %q", model.renderCandidateDetail())
	}
	model.sessions = nil
	model.screen = screenSessions
	if !strings.Contains(model.renderSessionsList(), "No conversations match") || !strings.Contains(model.renderSessionDetail(), "Use `f`") {
		t.Fatalf("empty session renders are wrong")
	}

	model.status = "Error: boom"
	if !strings.Contains(model.renderStatus(), "boom") {
		t.Fatalf("renderStatus(error) = %q", model.renderStatus())
	}
	model.status = "Removed item"
	if !strings.Contains(model.renderStatus(), "Removed item") {
		t.Fatalf("renderStatus(ok) = %q", model.renderStatus())
	}
	model.status = "Working"
	if !strings.Contains(model.renderStatus(), "Working") {
		t.Fatalf("renderStatus(accent) = %q", model.renderStatus())
	}

	if model.renderSafetyBadge(SafetySafe) == "" || model.renderSafetyBadge(SafetyConfirm) == "" || model.renderSafetyBadge(SafetyManual) == "" || model.renderSafetyBadge("other") != "other" {
		t.Fatal("renderSafetyBadge() mismatch")
	}
	if formatCutoffValue(time.Time{}) != "" || formatCutoffValue(time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)) != "2026-03-01" {
		t.Fatal("formatCutoffValue() mismatch")
	}
	if got := overlayCenter("abc\ndef", "X", 10, 4); !strings.Contains(got, "X") {
		t.Fatalf("overlayCenter() = %q", got)
	}
	if padRight("abc", 5) != "abc  " {
		t.Fatalf("padRight() = %q", padRight("abc", 5))
	}
	if maxInt(1, 2) != 2 || maxInt(3, 2) != 3 {
		t.Fatal("maxInt() mismatch")
	}
	if clampIndex(-1, 2) != 0 || clampIndex(9, 2) != 1 || clampIndex(0, 0) != 0 {
		t.Fatal("clampIndex() mismatch")
	}
}

func TestRunAnalyzeTUIQuit(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Bubble Tea quit smoke test is only reliable on macOS in this project")
	}

	home := setTestHome(t)
	writeTestFile(t, filepath.Join(home, ".ollama", "logs", "server.log"), "log")

	withFakeStdin(t, "q", func() {
		if err := runAnalyzeTUI([]string{"ollama"}, time.Time{}, &strings.Builder{}, &strings.Builder{}); err != nil {
			t.Fatalf("runAnalyzeTUI() err = %v", err)
		}
	})
}
