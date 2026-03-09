package cleaner

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type analyzeScreen int

const (
	screenAssistants analyzeScreen = iota
	screenOpenClawMenu
	screenSessions
	screenSessionPreview
	screenCandidates
)

type analyzeInputMode int

const (
	inputNone analyzeInputMode = iota
	inputSessionFilter
	inputSessionBulkDelete
)

type analyzeConfirmMode int

const (
	confirmNone analyzeConfirmMode = iota
	confirmDeleteSession
	confirmDeleteSessionsBefore
	confirmDeleteCandidate
)

type analyzeStyles struct {
	frame     lipgloss.Style
	header    lipgloss.Style
	subheader lipgloss.Style
	selected  lipgloss.Style
	muted     lipgloss.Style
	accent    lipgloss.Style
	ok        lipgloss.Style
	warn      lipgloss.Style
	error     lipgloss.Style
	footer    lipgloss.Style
	box       lipgloss.Style
	dialog    lipgloss.Style
	badge     lipgloss.Style
	danger    lipgloss.Style
	review    lipgloss.Style
	manual    lipgloss.Style
}

type analyzeModel struct {
	width           int
	height          int
	styles          analyzeStyles
	assistants      []string
	summaries       []assistantSummary
	screen          analyzeScreen
	activeAssistant string
	assistantIndex  int
	openclawIndex   int
	candidates      []Candidate
	candidateIndex  int
	sessions        []OpenClawSession
	sessionIndex    int
	sessionBefore   time.Time
	initialBefore   time.Time
	inputMode       analyzeInputMode
	inputLabel      string
	inputValue      string
	confirmMode     analyzeConfirmMode
	confirmTitle    string
	confirmBody     string
	confirmCutoff   time.Time
	confirmSession  OpenClawSession
	confirmItem     Candidate
	status          string
	lastErr         error
	previewText     string
}

func runAnalyzeTUI(assistants []string, before time.Time, stdout, stderr io.Writer) error {
	model, err := newAnalyzeModel(assistants, before)
	if err != nil {
		return err
	}

	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithOutput(stdout),
		tea.WithInput(os.Stdin),
	)
	finalModel, err := program.Run()
	if err != nil {
		return err
	}
	if final, ok := finalModel.(analyzeModel); ok && final.lastErr != nil {
		fmt.Fprintln(stderr, final.lastErr)
	}
	return nil
}

func newAnalyzeModel(assistants []string, before time.Time) (analyzeModel, error) {
	m := analyzeModel{
		styles:        newAnalyzeStyles(),
		assistants:    assistants,
		initialBefore: before,
		sessionBefore: before,
	}

	summaries, err := assistantAnalyzeSummary(assistants)
	if err != nil {
		return m, err
	}
	m.summaries = summaries

	if len(assistants) == 1 {
		assistant := assistants[0]
		m.activeAssistant = assistant
		switch assistant {
		case "openclaw":
			if !before.IsZero() {
				m.screen = screenSessions
				if err := m.reloadSessions(); err != nil {
					return m, err
				}
			} else {
				m.screen = screenOpenClawMenu
			}
		default:
			m.screen = screenCandidates
			if err := m.reloadCandidates(assistant); err != nil {
				return m, err
			}
		}
		return m, nil
	}

	m.screen = screenAssistants
	return m, nil
}

// Modern SaaS Design Colors
var (
	colorPrimary = lipgloss.AdaptiveColor{Light: "#5A29E4", Dark: "#7D56F4"} // Elegant Violet
	colorAccent  = lipgloss.AdaptiveColor{Light: "#D92662", Dark: "#F25D94"} // Vibrant Pink
	colorSuccess = lipgloss.AdaptiveColor{Light: "#2E8A4A", Dark: "#43BF6D"} // Soft Green
	colorWarn    = lipgloss.AdaptiveColor{Light: "#B6454F", Dark: "#E06C75"} // Soft Red/Yellow
	colorDanger  = lipgloss.AdaptiveColor{Light: "#D32F2F", Dark: "#FF4A4A"} // Soft Red
	colorFg      = lipgloss.AdaptiveColor{Light: "#333333", Dark: "#E5E5E5"} // Main Text
	colorSubdued = lipgloss.AdaptiveColor{Light: "#737373", Dark: "#6B7280"} // Muted text
	colorBorder  = lipgloss.AdaptiveColor{Light: "#D9D9D9", Dark: "#3B4048"} // Panel borders
	colorActive  = lipgloss.AdaptiveColor{Light: "#E0D8F9", Dark: "#3A2A68"} // Active selection background
	colorWhite   = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}
)

func newAnalyzeStyles() analyzeStyles {
	return analyzeStyles{
		frame:     lipgloss.NewStyle().Padding(1, 4).Foreground(colorFg),
		header:    lipgloss.NewStyle().Bold(true).Foreground(colorPrimary),
		subheader: lipgloss.NewStyle().Foreground(colorSubdued),
		// Molecular minimalism: No block background for selected items, just bold accent color
		selected: lipgloss.NewStyle().Foreground(colorAccent).Bold(true),
		muted:    lipgloss.NewStyle().Foreground(colorSubdued),
		accent:   lipgloss.NewStyle().Foreground(colorAccent),
		ok:       lipgloss.NewStyle().Foreground(colorSuccess).Bold(true),
		warn:     lipgloss.NewStyle().Foreground(colorWarn).Bold(true),
		error:    lipgloss.NewStyle().Foreground(colorDanger).Bold(true),
		// Border top for footer to separate from content naturally
		footer: lipgloss.NewStyle().Foreground(colorSubdued).BorderTop(true).BorderForeground(colorBorder).MarginTop(1).PaddingTop(1),
		// No RoundedBorder for Box anymore, simple padding
		box:    lipgloss.NewStyle().Padding(0, 2),
		dialog: lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(colorPrimary).Padding(1, 4).Width(60).Align(lipgloss.Center),
		badge:  lipgloss.NewStyle().Foreground(colorWhite).Background(colorSubdued).Padding(0, 1),
		danger: lipgloss.NewStyle().Foreground(colorDanger),
		review: lipgloss.NewStyle().Foreground(colorWarn),
		manual: lipgloss.NewStyle().Foreground(colorSubdued),
	}
}

func (m analyzeModel) Init() tea.Cmd {
	return nil
}

func (m analyzeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		next, cmd := m.handleKey(msg)
		return next, cmd
	}
	return m, nil
}

func (m analyzeModel) handleKey(msg tea.KeyMsg) (analyzeModel, tea.Cmd) {
	if m.inputMode != inputNone {
		return m.handleInputKey(msg)
	}
	if m.confirmMode != confirmNone {
		return m.handleConfirmKey(msg)
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		if m.atRoot() {
			return m, tea.Quit
		}
		m.navigateBack()
		return m, nil
	case "esc", "backspace":
		if m.atRoot() {
			return m, tea.Quit
		}
		m.navigateBack()
		return m, nil
	case "up", "k":
		m.moveCursor(-1)
		return m, nil
	case "down", "j":
		m.moveCursor(1)
		return m, nil
	case "enter":
		if err := m.activateSelection(); err != nil {
			m.lastErr = err
			m.status = "Error: " + err.Error()
		}
		return m, nil
	case "d":
		if err := m.prepareDeleteSelected(); err != nil {
			m.status = "Cannot delete: " + err.Error()
		}
		return m, nil
	case "f":
		if m.screen == screenSessions {
			m.inputMode = inputSessionFilter
			m.inputLabel = "Filter conversations updated before date (YYYY-MM-DD)"
			m.inputValue = formatCutoffValue(m.sessionBefore)
		}
		return m, nil
	case "x":
		if m.screen == screenSessions {
			if m.sessionBefore.IsZero() {
				m.inputMode = inputSessionBulkDelete
				m.inputLabel = "Delete conversations updated before date (YYYY-MM-DD)"
				m.inputValue = ""
				return m, nil
			}
			m.prepareDeleteSessionsBefore(m.sessionBefore)
		}
		return m, nil
	case "c":
		if m.screen == screenSessions {
			m.sessionBefore = time.Time{}
			if err := m.reloadSessions(); err != nil {
				m.lastErr = err
				m.status = "Error: " + err.Error()
			} else {
				m.status = "Date filter cleared."
			}
		}
		return m, nil
	}

	return m, nil
}

func (m analyzeModel) handleInputKey(msg tea.KeyMsg) (analyzeModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.inputMode = inputNone
		m.inputLabel = ""
		m.inputValue = ""
		return m, nil
	case "enter":
		cutoff, err := parseDateCutoff(m.inputValue)
		if err != nil {
			m.status = err.Error()
			return m, nil
		}
		switch m.inputMode {
		case inputSessionFilter:
			m.sessionBefore = cutoff
			if err := m.reloadSessions(); err != nil {
				m.lastErr = err
				m.status = "Error: " + err.Error()
			} else {
				m.status = "Showing conversations before " + cutoff.Format("2006-01-02") + "."
			}
		case inputSessionBulkDelete:
			m.prepareDeleteSessionsBefore(cutoff)
		}
		m.inputMode = inputNone
		m.inputLabel = ""
		m.inputValue = ""
		return m, nil
	case "backspace":
		if len(m.inputValue) > 0 {
			m.inputValue = m.inputValue[:len(m.inputValue)-1]
		}
		return m, nil
	}

	if msg.Type == tea.KeyRunes {
		for _, r := range msg.Runes {
			if (r >= '0' && r <= '9') || r == '-' {
				m.inputValue += string(r)
			}
		}
	}
	return m, nil
}

func (m analyzeModel) handleConfirmKey(msg tea.KeyMsg) (analyzeModel, tea.Cmd) {
	switch msg.String() {
	case "esc", "n":
		m.clearConfirm()
		m.status = "Deletion cancelled."
		return m, nil
	case "enter", "y":
		if err := m.executeConfirm(); err != nil {
			m.lastErr = err
			m.status = "Error: " + err.Error()
		}
		m.clearConfirm()
		return m, nil
	}
	return m, nil
}

func (m *analyzeModel) moveCursor(delta int) {
	switch m.screen {
	case screenAssistants:
		m.assistantIndex = clampIndex(m.assistantIndex+delta, len(m.summaries))
	case screenOpenClawMenu:
		m.openclawIndex = clampIndex(m.openclawIndex+delta, 2)
	case screenSessions:
		m.sessionIndex = clampIndex(m.sessionIndex+delta, len(m.sessions))
	case screenCandidates:
		m.candidateIndex = clampIndex(m.candidateIndex+delta, len(m.candidates))
	}
}

func (m *analyzeModel) activateSelection() error {
	switch m.screen {
	case screenAssistants:
		if len(m.summaries) == 0 {
			return nil
		}
		m.activeAssistant = m.summaries[m.assistantIndex].Assistant
		if m.activeAssistant == "openclaw" {
			m.screen = screenOpenClawMenu
			return nil
		}
		m.screen = screenCandidates
		return m.reloadCandidates(m.activeAssistant)
	case screenOpenClawMenu:
		if m.openclawIndex == 0 {
			m.screen = screenSessions
			return m.reloadSessions()
		}
		m.screen = screenCandidates
		return m.reloadCandidates("openclaw")
	case screenSessions:
		if len(m.sessions) == 0 {
			return nil
		}
		session := m.sessions[m.sessionIndex]
		preview, err := previewOpenClawSession(session.TranscriptPath)
		if err != nil {
			return err
		}
		m.previewText = preview
		m.screen = screenSessionPreview
		return nil
	}
	return nil
}

func (m *analyzeModel) prepareDeleteSelected() error {
	switch m.screen {
	case screenSessions:
		if len(m.sessions) == 0 {
			return fmt.Errorf("no conversation selected")
		}
		session := m.sessions[m.sessionIndex]
		m.confirmMode = confirmDeleteSession
		m.confirmSession = session
		m.confirmTitle = "Delete conversation?"
		m.confirmBody = fmt.Sprintf("%s\n%s\n%s", session.DisplayLabel(), formatSessionTime(session.SortTime()), session.TranscriptPath)
		return nil
	case screenCandidates:
		if len(m.candidates) == 0 {
			return fmt.Errorf("no item selected")
		}
		candidate := m.candidates[m.candidateIndex]
		if candidate.Safety == SafetyManual {
			return fmt.Errorf("manual items must stay review-only")
		}
		m.confirmMode = confirmDeleteCandidate
		m.confirmItem = candidate
		m.confirmTitle = "Delete item?"
		m.confirmBody = fmt.Sprintf("%s\n%s\n%s", displayKind(candidate.Kind), candidate.Path, candidate.Reason)
		return nil
	default:
		return fmt.Errorf("nothing to delete on this screen")
	}
}

func (m *analyzeModel) prepareDeleteSessionsBefore(cutoff time.Time) {
	matches := filterSessionsBefore(m.sessionsSource(), cutoff)
	var bytes int64
	for _, session := range matches {
		bytes += session.SizeBytes
	}
	m.confirmMode = confirmDeleteSessionsBefore
	m.confirmCutoff = cutoff
	m.confirmTitle = "Delete older conversations?"
	m.confirmBody = fmt.Sprintf("%d conversation(s) before %s\nAbout %s will be removed.", len(matches), cutoff.Format("2006-01-02"), formatBytes(bytes))
}

func (m *analyzeModel) executeConfirm() error {
	switch m.confirmMode {
	case confirmDeleteSession:
		if err := deleteOpenClawSessions([]OpenClawSession{m.confirmSession}); err != nil {
			return err
		}
		m.status = "Conversation deleted."
		return m.reloadSessions()
	case confirmDeleteSessionsBefore:
		matches := filterSessionsBefore(m.sessionsSource(), m.confirmCutoff)
		if len(matches) == 0 {
			m.status = "No conversations matched that date."
			return nil
		}
		if err := deleteOpenClawSessions(matches); err != nil {
			return err
		}
		m.sessionBefore = m.confirmCutoff
		m.status = fmt.Sprintf("Removed %d older conversation(s).", len(matches))
		return m.reloadSessions()
	case confirmDeleteCandidate:
		if err := deletePath(m.confirmItem.Path); err != nil {
			return err
		}
		m.status = "Item deleted."
		return m.reloadCandidates(m.activeAssistant)
	}
	return nil
}

func (m *analyzeModel) clearConfirm() {
	m.confirmMode = confirmNone
	m.confirmTitle = ""
	m.confirmBody = ""
	m.confirmCutoff = time.Time{}
	m.confirmSession = OpenClawSession{}
	m.confirmItem = Candidate{}
}

func (m *analyzeModel) navigateBack() {
	switch m.screen {
	case screenSessionPreview:
		m.screen = screenSessions
		return
	case screenCandidates:
		if m.activeAssistant == "openclaw" {
			m.screen = screenOpenClawMenu
			return
		}
		if len(m.assistants) > 1 {
			m.screen = screenAssistants
			return
		}
	case screenSessions:
		m.screen = screenOpenClawMenu
		return
	case screenOpenClawMenu:
		if len(m.assistants) > 1 {
			m.screen = screenAssistants
			return
		}
	}
}

func (m analyzeModel) atRoot() bool {
	switch m.screen {
	case screenAssistants:
		return true
	case screenCandidates:
		return len(m.assistants) == 1 && m.activeAssistant != "openclaw"
	case screenOpenClawMenu:
		return len(m.assistants) == 1
	default:
		return false
	}
}

func (m *analyzeModel) reloadCandidates(assistant string) error {
	candidates, err := discoverAssistantLeftovers(assistant)
	if err != nil {
		return err
	}
	m.activeAssistant = assistant
	m.candidates = liveCandidates(candidates)
	sort.Slice(m.candidates, func(i, j int) bool {
		if m.candidates[i].Safety != m.candidates[j].Safety {
			return m.candidates[i].Safety < m.candidates[j].Safety
		}
		return m.candidates[i].Path < m.candidates[j].Path
	})
	m.candidateIndex = clampIndex(m.candidateIndex, len(m.candidates))
	return nil
}

func (m *analyzeModel) reloadSessions() error {
	sessions, err := discoverOpenClawSessions()
	if err != nil {
		return err
	}
	if !m.sessionBefore.IsZero() {
		sessions = filterSessionsBefore(sessions, m.sessionBefore)
	}
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].SortTime().After(sessions[j].SortTime())
	})
	m.sessions = sessions
	m.sessionIndex = clampIndex(m.sessionIndex, len(m.sessions))
	return nil
}

func (m analyzeModel) sessionsSource() []OpenClawSession {
	sessions, err := discoverOpenClawSessions()
	if err != nil {
		return nil
	}
	return sessions
}

func clampIndex(index, length int) int {
	if length <= 0 {
		return 0
	}
	if index < 0 {
		return 0
	}
	if index >= length {
		return length - 1
	}
	return index
}

func (m analyzeModel) View() string {
	if m.width == 0 {
		m.width = 100
	}
	if m.height == 0 {
		m.height = 30
	}

	header := m.renderHeader()
	listPane, detailPane := m.renderBody()
	footer := m.renderFooter()

	contentWidth := m.width - 4
	listWidth := contentWidth / 2
	var body string
	if contentWidth < 80 {
		body = lipgloss.JoinVertical(lipgloss.Left,
			m.styles.box.Render(listPane),
			m.styles.box.Render(detailPane),
		)
	} else {
		// Use a simple flow with a soft right border for the list.
		listMin := 45
		if listWidth < listMin {
			listMin = listWidth
		}

		listPaneStyled := lipgloss.NewStyle().
			Width(listMin).
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(colorBorder).
			MarginRight(3).
			PaddingRight(1).
			Render(listPane)

		body = lipgloss.JoinHorizontal(lipgloss.Top,
			listPaneStyled,
			m.styles.box.Render(detailPane),
		)
	}

	view := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
	if m.inputMode != inputNone {
		view = overlayCenter(view, m.renderInputDialog(), m.width, m.height)
	}
	if m.confirmMode != confirmNone {
		view = overlayCenter(view, m.renderConfirmDialog(), m.width, m.height)
	}
	return m.styles.frame.Width(m.width).Height(m.height).Render(view)
}

func (m analyzeModel) renderHeader() string {
	breadcrumbs := []string{"Analyze"}
	switch m.screen {
	case screenOpenClawMenu, screenSessions, screenSessionPreview, screenCandidates:
		if m.activeAssistant != "" {
			breadcrumbs = append(breadcrumbs, displayAssistant(m.activeAssistant))
		}
	}
	if m.screen == screenSessions || m.screen == screenSessionPreview {
		breadcrumbs = append(breadcrumbs, "Conversations")
	}
	if m.screen == screenSessionPreview {
		breadcrumbs = append(breadcrumbs, "Preview")
	}
	if m.screen == screenCandidates {
		breadcrumbs = append(breadcrumbs, "Items")
	}

	titleParts := make([]string, len(breadcrumbs))
	for i, crumb := range breadcrumbs {
		if i == len(breadcrumbs)-1 {
			titleParts[i] = m.styles.header.Render(crumb)
		} else {
			titleParts[i] = m.styles.muted.Render(crumb)
		}
	}
	title := strings.Join(titleParts, m.styles.muted.Render(" › "))

	subtitle := m.styles.muted.Render("↑/↓ move  •  Enter open  •  d delete  •  q back")
	if m.screen == screenSessions && !m.sessionBefore.IsZero() {
		subtitle = m.styles.warn.Render(fmt.Sprintf("Filtering OpenClaw conversations before %s. Press c to clear.", m.sessionBefore.Format("2006-01-02")))
	}

	// Minimalistic header padding
	parts := []string{title, subtitle, ""}
	if m.status != "" {
		parts = append(parts, m.renderStatus(), "")
	}
	return strings.Join(parts, "\n")
}

func (m analyzeModel) renderBody() (string, string) {
	switch m.screen {
	case screenAssistants:
		return m.renderAssistantList(), m.renderAssistantDetail()
	case screenOpenClawMenu:
		return m.renderOpenClawMenu(), m.renderOpenClawDetail()
	case screenSessions:
		return m.renderSessionsList(), m.renderSessionDetail()
	case screenSessionPreview:
		return m.renderSessionsList(), m.renderSessionPreview()
	case screenCandidates:
		return m.renderCandidateList(), m.renderCandidateDetail()
	default:
		return "", ""
	}
}

func (m analyzeModel) renderFooter() string {
	keys := []string{"↑/↓ Navigate", "Enter Select", "q Back"}
	switch m.screen {
	case screenSessionPreview:
		keys = []string{"q Back to conversations"}
	case screenSessions:
		keys = append(keys, "d Delete item", "f Filter date", "x Bulk delete", "c Clear filter")
	case screenCandidates:
		keys = append(keys, "d Delete item")
	}
	return m.styles.footer.Render(strings.Join(keys, "   "))
}

func (m analyzeModel) renderAssistantList() string {
	lines := []string{m.styles.header.Render("Assistants"), ""}
	for i, item := range m.summaries {
		label := displayAssistant(item.Assistant)
		detail := fmt.Sprintf("%d items", item.LeftoverCount)
		if item.SessionCount > 0 {
			detail = fmt.Sprintf("%d conversations, %d others", item.SessionCount, item.LeftoverCount)
		}

		row := fmt.Sprintf("%-14s %s", label, m.styles.muted.Render(detail))
		if i == m.assistantIndex {
			row = m.styles.selected.Render(fmt.Sprintf("❯ %-12s %s", label, detail))
		} else {
			row = fmt.Sprintf("  %s", row)
		}
		lines = append(lines, row)
	}
	return strings.Join(lines, "\n")
}

func (m analyzeModel) renderAssistantDetail() string {
	if len(m.summaries) == 0 {
		return m.styles.muted.Render("No assistants available.")
	}
	item := m.summaries[m.assistantIndex]
	lines := []string{
		m.styles.header.Render(displayAssistant(item.Assistant)),
		"",
		fmt.Sprintf("%-16s %8d", m.styles.muted.Render("Conversations"), item.SessionCount),
		fmt.Sprintf("%-16s %8d", m.styles.muted.Render("Other items"), item.LeftoverCount),
		"",
		m.styles.accent.Render("Press Enter to inspect."),
	}
	if item.Assistant == "openclaw" {
		lines = append(lines, m.styles.muted.Render("OpenClaw supports date-based cleanup."))
	}
	return strings.Join(lines, "\n")
}

func (m analyzeModel) renderOpenClawMenu() string {
	options := []string{"Conversations", "Other leftover items"}
	lines := []string{m.styles.header.Render("OpenClaw Categories"), ""}
	for i, option := range options {
		line := option
		if i == m.openclawIndex {
			line = m.styles.selected.Render(fmt.Sprintf("❯ %s", line))
		} else {
			line = fmt.Sprintf("  %s", line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (m analyzeModel) renderOpenClawDetail() string {
	sessions, _ := discoverOpenClawSessions()
	leftovers, _ := discoverAssistantLeftovers("openclaw")
	detail := []string{
		m.styles.header.Render("Stats"),
		"",
		fmt.Sprintf("%-16s %8d", m.styles.muted.Render("Conversations"), len(sessions)),
		fmt.Sprintf("%-16s %8d", m.styles.muted.Render("Other items"), len(leftovers)),
		"",
		m.styles.accent.Render("Conversations view lets you:"),
		m.styles.muted.Render("• browse one conversation at a time"),
		m.styles.muted.Render("• filter by date"),
		m.styles.muted.Render("• delete everything before a chosen date"),
	}
	if !m.initialBefore.IsZero() {
		detail = append(detail, "", m.styles.warn.Render("Startup filter: before "+m.initialBefore.Format("2006-01-02")))
	}
	return strings.Join(detail, "\n")
}

func (m analyzeModel) renderSessionsList() string {
	lines := []string{m.styles.header.Render("Conversations"), ""}
	if len(m.sessions) == 0 {
		lines = append(lines, m.styles.muted.Render("No conversations match the current filter."))
		return strings.Join(lines, "\n")
	}
	for i, session := range m.sessions {
		dateStr := session.SortTime().Format("2006-01-02")
		sizeStr := formatBytes(session.SizeBytes)
		labelStr := trimForDisplay(session.ShortLabel(), 22)

		row := fmt.Sprintf("%-10s %-8s %s", m.styles.muted.Render(dateStr), sizeStr, m.styles.muted.Render(labelStr))
		if i == m.sessionIndex {
			row = m.styles.selected.Render(fmt.Sprintf("❯ %-8s %-8s %s", dateStr, sizeStr, labelStr))
		} else {
			row = fmt.Sprintf("  %s", row)
		}
		lines = append(lines, row)
	}
	return strings.Join(lines, "\n")
}

func (m analyzeModel) renderSessionDetail() string {
	if len(m.sessions) == 0 {
		return m.styles.muted.Render("  Use `f` to set another date or `c` to clear the current filter.")
	}
	session := m.sessions[m.sessionIndex]
	lines := []string{
		m.styles.header.Render(" 📄 " + session.DisplayLabel()),
		"",
		fmt.Sprintf("  %-18s %s", m.styles.muted.Render("Updated:"), formatSessionTime(session.SortTime())),
		fmt.Sprintf("  %-18s %s", m.styles.muted.Render("Started:"), formatSessionTime(session.StartedAt)),
		fmt.Sprintf("  %-18s %s", m.styles.muted.Render("Agent:"), session.AgentID),
		fmt.Sprintf("  %-18s %d", m.styles.muted.Render("Events:"), session.MessageCount),
		fmt.Sprintf("  %-18s %s", m.styles.muted.Render("Tokens:"), m.styles.accent.Render(formatTokenCount(session.TotalTokens))),
		fmt.Sprintf("  %-18s %s", m.styles.muted.Render("Size:"), formatBytes(session.SizeBytes)),
	}
	if session.Source != "" {
		lines = append(lines, fmt.Sprintf("  %-18s %s", m.styles.muted.Render("Source:"), trimForDisplay(session.Source, 48)))
	}
	lines = append(lines, "", m.styles.muted.Render("  Transcript"), "  "+session.TranscriptPath, "", m.styles.accent.Render("  Press d to delete only this conversation."))
	if m.sessionBefore.IsZero() {
		lines = append(lines, m.styles.warn.Render("  Press x to choose a cutoff date and bulk-delete older conversations."))
	} else {
		lines = append(lines, m.styles.warn.Render("  Press x to delete all conversations in the current filter."))
	}
	return strings.Join(lines, "\n")
}

func (m analyzeModel) renderSessionPreview() string {
	if m.previewText == "" {
		return m.styles.muted.Render("Unable to load preview data.")
	}

	lines := []string{
		m.styles.header.Render("Conversation Preview"),
		"",
	}

	// Truncate the preview text to fit reasonably well
	previewLines := strings.Split(m.previewText, "\n")
	maxLines := m.height - 12 // Leave space for headers, footers
	if maxLines < 10 {
		maxLines = 10
	}

	if len(previewLines) > maxLines {
		// Just take the tail part
		lines = append(lines, m.styles.muted.Render(fmt.Sprintf("... (%d truncated lines) ...", len(previewLines)-maxLines)))
		previewLines = previewLines[len(previewLines)-maxLines:]
	}

	for _, l := range previewLines {
		if strings.HasPrefix(l, "== User ==") || strings.HasPrefix(l, "== Assistant ==") || strings.HasPrefix(l, "== System ==") {
			lines = append(lines, m.styles.accent.Render(strings.TrimSpace(l)))
		} else {
			lines = append(lines, l)
		}
	}

	return strings.Join(lines, "\n")
}

func (m analyzeModel) renderCandidateList() string {
	lines := []string{m.styles.header.Render(displayAssistant(m.activeAssistant) + " items"), ""}
	if len(m.candidates) == 0 {
		lines = append(lines, m.styles.muted.Render("No matching items remain."))
		return strings.Join(lines, "\n")
	}
	for i, candidate := range m.candidates {
		safety := displaySafety(candidate.Safety)

		row := fmt.Sprintf("%-16s %-7s %8s", trimForDisplay(displayKind(candidate.Kind), 16), safety, formatBytes(candidate.SizeBytes))
		if i == m.candidateIndex {
			row = m.styles.selected.Render(fmt.Sprintf("❯ %s", row))
		} else {
			row = fmt.Sprintf("  %s", row)
		}
		lines = append(lines, row)
	}
	return strings.Join(lines, "\n")
}

func (m analyzeModel) renderCandidateDetail() string {
	if len(m.candidates) == 0 {
		return m.styles.muted.Render("Nothing to show.")
	}
	candidate := m.candidates[m.candidateIndex]
	lines := []string{
		m.styles.header.Render(displayKind(candidate.Kind)),
		"",
		fmt.Sprintf("%-10s %s", m.styles.muted.Render("Safety"), m.renderSafetyBadge(candidate.Safety)),
		fmt.Sprintf("%-10s %s", m.styles.muted.Render("Size"), formatBytes(candidate.SizeBytes)),
		"",
		candidate.Reason,
		"",
		m.styles.muted.Render("Path"),
		candidate.Path,
	}
	for _, note := range candidate.Notes {
		lines = append(lines, "", m.styles.muted.Render("Note"), note)
	}
	switch candidate.Safety {
	case SafetyManual:
		lines = append(lines, "", m.styles.manual.Render("Manual inspection only."))
	case SafetyConfirm:
		lines = append(lines, "", m.styles.accent.Render("Press d to delete this reviewed item."))
	default:
		lines = append(lines, "", m.styles.danger.Render("Press d to delete this item."))
	}
	return strings.Join(lines, "\n")
}

func (m analyzeModel) renderStatus() string {
	switch {
	case strings.HasPrefix(m.status, "Error:"):
		return m.styles.error.Render(m.status)
	case strings.Contains(strings.ToLower(m.status), "deleted"), strings.Contains(strings.ToLower(m.status), "removed"):
		return m.styles.ok.Render(m.status)
	default:
		return m.styles.accent.Render(m.status)
	}
}

func (m analyzeModel) renderSafetyBadge(safety Safety) string {
	switch safety {
	case SafetySafe:
		return m.styles.ok.Render("safe")
	case SafetyConfirm:
		return m.styles.review.Render("review")
	case SafetyManual:
		return m.styles.manual.Render("manual")
	default:
		return string(safety)
	}
}

func (m analyzeModel) renderInputDialog() string {
	lines := []string{
		m.styles.header.Render(m.inputLabel),
		"",
		m.styles.accent.Render("> " + m.inputValue),
		"",
		m.styles.footer.Render("Enter apply  •  Esc cancel"),
	}
	return m.styles.dialog.Render(strings.Join(lines, "\n"))
}

func (m analyzeModel) renderConfirmDialog() string {
	lines := []string{
		m.styles.header.Render(m.confirmTitle),
		"",
		m.confirmBody,
		"",
		m.styles.footer.Render("Enter or y confirm  •  Esc or n cancel"),
	}
	return m.styles.dialog.Render(strings.Join(lines, "\n"))
}

func formatCutoffValue(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func overlayCenter(background, modal string, width, height int) string {
	bgLines := strings.Split(background, "\n")
	modalLines := strings.Split(modal, "\n")

	top := maxInt((height-len(modalLines))/2, 0)
	left := maxInt((width-lipgloss.Width(modal))/2, 0)

	for i := range modalLines {
		row := top + i
		if row >= len(bgLines) {
			bgLines = append(bgLines, strings.Repeat(" ", width))
		}
		line := padRight(bgLines[row], width)
		modalLine := modalLines[i]
		if left+lipgloss.Width(modalLine) > len([]rune(line)) {
			line = padRight(line, left+lipgloss.Width(modalLine))
		}
		prefix := string([]rune(line)[:left])
		suffixStart := left + lipgloss.Width(modalLine)
		runes := []rune(line)
		suffix := ""
		if suffixStart < len(runes) {
			suffix = string(runes[suffixStart:])
		}
		bgLines[row] = prefix + modalLine + suffix
	}

	return strings.Join(bgLines, "\n")
}

func padRight(value string, width int) string {
	current := lipgloss.Width(value)
	if current >= width {
		return value
	}
	return value + strings.Repeat(" ", width-current)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
