package cleaner

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	analyzePageMain   = "main"
	analyzePageDialog = "dialog"
)

type analyzeTUIView struct {
	app    *tview.Application
	pages  *tview.Pages
	header *tview.TextView
	left   *tview.TextView
	right  *tview.TextView
	footer *tview.TextView

	model analyzeModel
	err   error

	ignoreKeysUntil time.Time
	lastWidth       int
	lastHeight      int
}

func runAnalyzeTUI(assistants []string, before time.Time, stdout, stderr io.Writer) error {
	if !isTerminal(os.Stdin) {
		// Keep deterministic test behavior for non-TTY inputs.
		return runAnalyzeTUILegacy(assistants, before, stdout, stderr)
	}

	model, err := newAnalyzeModel(assistants, before)
	if err != nil {
		return err
	}

	view := newAnalyzeTUIView(model)
	view.refresh()
	if err := view.app.Run(); err != nil {
		return err
	}
	if view.err != nil {
		fmt.Fprintln(stderr, view.err)
	}
	return nil
}

func newAnalyzeTUIView(model analyzeModel) *analyzeTUIView {
	header := tview.NewTextView().SetDynamicColors(true)
	left := tview.NewTextView().SetDynamicColors(true)
	right := tview.NewTextView().SetDynamicColors(true)
	footer := tview.NewTextView().SetDynamicColors(true)

	left.SetBorder(true).SetTitle(" Items ")
	right.SetBorder(true).SetTitle(" Detail ")

	body := tview.NewFlex().
		AddItem(left, 0, 45, true).
		AddItem(right, 0, 55, false)
	root := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(header, 2, 0, false).
		AddItem(body, 0, 1, true).
		AddItem(footer, 1, 0, false)

	pages := tview.NewPages()
	pages.AddPage(analyzePageMain, root, true, true)

	view := &analyzeTUIView{
		app:    tview.NewApplication(),
		pages:  pages,
		header: header,
		left:   left,
		right:  right,
		footer: footer,
		model:  model,
		// Some terminals emit startup escape noise. Ignore early keybindings briefly.
		ignoreKeysUntil: time.Now().Add(300 * time.Millisecond),
	}
	view.applyTheme()
	view.app.SetRoot(pages, true)
	view.app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		width, height := screen.Size()
		if width > 0 {
			view.lastWidth = width
		}
		if height > 0 {
			view.lastHeight = height
		}
		return false
	})
	view.app.SetInputCapture(view.handleKey)
	return view
}

func (v *analyzeTUIView) handleKey(event *tcell.EventKey) *tcell.EventKey {
	if v.pages.HasPage(analyzePageDialog) {
		switch event.Key() {
		case tcell.KeyEsc:
			v.closeDialog("Cancelled.")
			return nil
		}
		switch event.Rune() {
		case 'q':
			v.closeDialog("Cancelled.")
			return nil
		}
		// Let active dialog widgets handle all other keys.
		return event
	}

	switch event.Key() {
	case tcell.KeyCtrlC:
		v.app.Stop()
		return nil
	case tcell.KeyUp:
		v.model.moveCursor(-1)
		v.refresh()
		return nil
	case tcell.KeyDown:
		v.model.moveCursor(1)
		v.refresh()
		return nil
	case tcell.KeyEnter:
		if err := v.model.activateSelection(); err != nil {
			v.model.status = "Error: " + err.Error()
			v.err = err
		}
		v.refresh()
		return nil
	case tcell.KeyEsc, tcell.KeyBackspace, tcell.KeyBackspace2:
		if v.model.atRoot() {
			v.app.Stop()
			return nil
		}
		v.model.navigateBack()
		v.refresh()
		return nil
	}

	if time.Now().Before(v.ignoreKeysUntil) {
		// Drop early startup runes to avoid accidental shortcut triggers.
		return nil
	}

	switch event.Rune() {
	case 'k':
		v.model.moveCursor(-1)
		v.refresh()
		return nil
	case 'j':
		v.model.moveCursor(1)
		v.refresh()
		return nil
	case 'q':
		if v.model.atRoot() {
			v.app.Stop()
			return nil
		}
		v.model.navigateBack()
		v.refresh()
		return nil
	case 'd':
		if err := v.model.prepareDeleteSelected(); err != nil {
			v.model.status = "Cannot delete: " + err.Error()
			v.refresh()
			return nil
		}
		v.showConfirmDialog(v.model.confirmTitle, v.model.confirmBody, func() error {
			err := v.model.executeConfirm()
			v.model.clearConfirm()
			return err
		})
		return nil
	case 'f':
		if v.model.screen != screenSessions {
			return event
		}
		v.showDateInputDialog(
			"Filter conversations before date (YYYY-MM-DD)",
			formatCutoffValue(v.model.sessionBefore),
			func(cutoff time.Time) error {
				v.model.sessionBefore = cutoff
				if err := v.model.reloadSessions(); err != nil {
					return err
				}
				v.model.status = "Showing conversations before " + cutoff.Format("2006-01-02") + "."
				return nil
			},
		)
		return nil
	case 'x':
		if v.model.screen != screenSessions || !assistantSupportsSessionDelete(v.model.activeAssistant) {
			return event
		}
		if v.model.sessionBefore.IsZero() {
			v.showDateInputDialog(
				"Delete conversations before date (YYYY-MM-DD)",
				"",
				func(cutoff time.Time) error {
					v.model.prepareDeleteSessionsBefore(cutoff)
					v.showConfirmDialog(v.model.confirmTitle, v.model.confirmBody, func() error {
						err := v.model.executeConfirm()
						v.model.clearConfirm()
						return err
					})
					return nil
				},
			)
			return nil
		}
		v.model.prepareDeleteSessionsBefore(v.model.sessionBefore)
		v.showConfirmDialog(v.model.confirmTitle, v.model.confirmBody, func() error {
			err := v.model.executeConfirm()
			v.model.clearConfirm()
			return err
		})
		return nil
	case 'c':
		if v.model.screen != screenSessions {
			return event
		}
		v.model.sessionBefore = time.Time{}
		if err := v.model.reloadSessions(); err != nil {
			v.model.status = "Error: " + err.Error()
			v.err = err
		} else {
			v.model.status = "Date filter cleared."
		}
		v.refresh()
		return nil
	}

	return event
}

func (v *analyzeTUIView) refresh() {
	if v.lastWidth > 0 {
		v.model.width = v.lastWidth
	}
	if v.lastHeight > 0 {
		v.model.height = v.lastHeight
	}

	if v.model.screen == screenSessions && len(v.model.sessions) > 0 && !v.model.hasSelectedSessionPreview() {
		v.model.loadSelectedSessionPreview()
	}

	headerLines := []string{"Analyze", ""}
	if v.model.activeAssistant != "" && v.model.screen != screenAssistants {
		headerLines = append(headerLines, "Assistant: "+displayAssistant(v.model.activeAssistant))
	}
	if !v.model.sessionBefore.IsZero() && v.model.screen == screenSessions {
		headerLines = append(headerLines, "Filter before: "+v.model.sessionBefore.Format("2006-01-02"))
	}
	if strings.TrimSpace(v.model.status) != "" {
		headerLines = append(headerLines, v.model.status)
	}
	v.header.SetText(strings.Join(headerLines, "\n"))

	left, right := v.renderBody()
	if v.model.screen == screenSessions || v.model.screen == screenSessionPreview {
		v.left.SetTitle(" Conversations ")
		v.right.SetTitle(" Detail ")
	} else if v.model.screen == screenCandidates {
		v.left.SetTitle(" Leftovers ")
		v.right.SetTitle(" Detail ")
	} else {
		v.left.SetTitle(" Items ")
		v.right.SetTitle(" Detail ")
	}
	v.left.SetText(left)
	v.right.SetText(right)
	v.footer.SetText(v.renderFooter())
}

func (v *analyzeTUIView) renderBody() (string, string) {
	switch v.model.screen {
	case screenAssistants:
		return renderAssistantListPlain(v.model), renderAssistantDetailPlain(v.model)
	case screenAssistantMenu:
		return renderAssistantMenuPlain(v.model), renderAssistantMenuDetailPlain(v.model)
	case screenSessions:
		return renderSessionsListPlain(v.model), renderSessionDetailPlain(v.model)
	case screenSessionPreview:
		return renderSessionsListPlain(v.model), renderSessionPreviewPlain(v.model)
	case screenCandidates:
		return renderCandidateListPlain(v.model), renderCandidateDetailPlain(v.model)
	default:
		return "", ""
	}
}

func (v *analyzeTUIView) renderFooter() string {
	switch v.model.screen {
	case screenSessionPreview:
		return "q back to conversations"
	case screenSessions:
		return "j/k or up/down navigate · enter preview · f filter · c clear · d delete · x bulk delete · q back"
	case screenCandidates:
		return "j/k or up/down navigate · d delete · q back"
	default:
		return "j/k or up/down navigate · enter select · q back"
	}
}

func (v *analyzeTUIView) showConfirmDialog(title, body string, onConfirm func() error) {
	modal := tview.NewModal().
		SetText(title + "\n\n" + body).
		AddButtons([]string{"Confirm", "Cancel"}).
		SetDoneFunc(func(index int, _ string) {
			v.pages.RemovePage(analyzePageDialog)
			if index == 0 {
				if err := onConfirm(); err != nil {
					v.err = err
					v.model.status = "Error: " + err.Error()
				}
			} else {
				v.model.status = "Deletion cancelled."
			}
			v.refresh()
			v.app.SetFocus(v.left)
		})
	modal.SetBackgroundColor(tcell.NewRGBColor(13, 17, 25))
	v.pages.AddPage(analyzePageDialog, modal, true, true)
	v.app.SetFocus(modal)
}

func (v *analyzeTUIView) showDateInputDialog(title, initial string, onSubmit func(time.Time) error) {
	input := tview.NewInputField().
		SetLabel("Date: ").
		SetText(initial)

	form := tview.NewForm().
		AddFormItem(input).
		AddButton("Apply", func() {
			cutoff, err := parseDateCutoff(input.GetText())
			if err != nil {
				v.model.status = err.Error()
				v.refresh()
				return
			}
			if err := onSubmit(cutoff); err != nil {
				v.err = err
				v.model.status = "Error: " + err.Error()
			}
			v.pages.RemovePage(analyzePageDialog)
			v.refresh()
			v.app.SetFocus(v.left)
		}).
		AddButton("Cancel", func() {
			v.closeDialog("Cancelled.")
		})
	form.SetBorder(true).SetTitle(" " + title + " ").SetTitleAlign(tview.AlignLeft)
	form.SetFieldBackgroundColor(tcell.NewRGBColor(19, 26, 41))
	form.SetButtonBackgroundColor(tcell.NewRGBColor(37, 52, 84))
	form.SetButtonTextColor(tcell.NewRGBColor(230, 233, 239))
	form.SetLabelColor(tcell.NewRGBColor(124, 211, 255))
	form.SetBorderColor(tcell.NewRGBColor(74, 85, 104))
	form.SetTitleColor(tcell.NewRGBColor(124, 211, 255))

	width := 68
	if len(title)+8 > width {
		width = len(title) + 8
	}
	dialog := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(form, width, 1, true).
		AddItem(nil, 0, 1, false)
	dialogRoot := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(dialog, 9, 1, true).
		AddItem(nil, 0, 1, false)

	v.pages.AddPage(analyzePageDialog, dialogRoot, true, true)
	v.app.SetFocus(input)
}

func (v *analyzeTUIView) closeDialog(status string) {
	v.pages.RemovePage(analyzePageDialog)
	v.model.status = status
	v.refresh()
	v.app.SetFocus(v.left)
}

func (v *analyzeTUIView) applyTheme() {
	primary := tcell.NewRGBColor(124, 211, 255)
	muted := tcell.NewRGBColor(148, 163, 184)
	normal := tcell.NewRGBColor(230, 233, 239)
	panel := tcell.NewRGBColor(13, 17, 25)
	border := tcell.NewRGBColor(74, 85, 104)

	v.header.SetTextColor(normal)
	v.left.SetTextColor(normal)
	v.right.SetTextColor(normal)
	v.footer.SetTextColor(muted)

	v.left.SetBackgroundColor(panel)
	v.right.SetBackgroundColor(panel)
	v.left.SetBorderColor(border)
	v.right.SetBorderColor(border)
	v.left.SetTitleColor(primary)
	v.right.SetTitleColor(primary)
}

func renderAssistantListPlain(m analyzeModel) string {
	lines := []string{"Assistants", ""}
	for i, item := range m.summaries {
		detail := fmt.Sprintf("%d items", item.LeftoverCount)
		if item.SessionCount > 0 {
			detail = fmt.Sprintf("%d conversations, %d others", item.SessionCount, item.LeftoverCount)
		}
		prefix := "  "
		if i == m.assistantIndex {
			prefix = "[#7cd3ff]>[-] "
		}
		lines = append(lines, fmt.Sprintf("%s%s  %s", prefix, displayAssistant(item.Assistant), detail))
	}
	return strings.Join(lines, "\n")
}

func renderAssistantDetailPlain(m analyzeModel) string {
	if len(m.summaries) == 0 {
		return "No assistants available."
	}
	item := m.summaries[m.assistantIndex]
	lines := []string{
		displayAssistant(item.Assistant),
		"",
		fmt.Sprintf("Conversations: %d", item.SessionCount),
		fmt.Sprintf("Other items:   %d", item.LeftoverCount),
	}
	return strings.Join(lines, "\n")
}

func renderAssistantMenuPlain(m analyzeModel) string {
	options := []string{"Conversations", "Other leftover items"}
	lines := []string{displayAssistant(m.activeAssistant) + " Categories", ""}
	for i, option := range options {
		prefix := "  "
		if i == m.assistantMenuIndex {
			prefix = "[#7cd3ff]>[-] "
		}
		lines = append(lines, prefix+option)
	}
	return strings.Join(lines, "\n")
}

func renderAssistantMenuDetailPlain(m analyzeModel) string {
	summary, _ := m.summaryForAssistant(m.activeAssistant)
	lines := []string{
		"Stats",
		"",
		fmt.Sprintf("Conversations: %d", summary.SessionCount),
		fmt.Sprintf("Other items:   %d", summary.LeftoverCount),
	}
	return strings.Join(lines, "\n")
}

func renderSessionsListPlain(m analyzeModel) string {
	lines := []string{"Conversations", ""}
	if len(m.sessions) == 0 {
		lines = append(lines, "No conversations match the current filter.")
		return strings.Join(lines, "\n")
	}

	labelWidth := 42
	start, end := visibleWindow(len(m.sessions), m.sessionIndex, m.sessionListVisibleCount())
	if start > 0 {
		lines = append(lines, fmt.Sprintf("... %d older conversation(s)", start))
	}
	for i := start; i < end; i++ {
		session := m.sessions[i]
		prefix := "  "
		if i == m.sessionIndex {
			prefix = "[#7cd3ff]>[-] "
		}
		lines = append(lines, fmt.Sprintf(
			"%s%s %8s %s",
			prefix,
			session.SortTime().Format("2006-01-02"),
			formatBytes(session.SizeBytes),
			trimForDisplay(session.ShortLabel(), labelWidth),
		))
	}
	if end < len(m.sessions) {
		lines = append(lines, fmt.Sprintf("... %d newer conversation(s)", len(m.sessions)-end))
	}
	return strings.Join(lines, "\n")
}

func renderSessionDetailPlain(m analyzeModel) string {
	if len(m.sessions) == 0 {
		return "Use f to set another date or c to clear the current filter."
	}
	session := m.sessions[m.sessionIndex]
	lines := []string{
		session.DisplayLabel(),
		"",
		fmt.Sprintf("Updated: %s", formatSessionTime(session.SortTime())),
		fmt.Sprintf("Started: %s", formatSessionTime(session.StartedAt)),
		fmt.Sprintf("Events:  %d", session.MessageCount),
		fmt.Sprintf("Tokens:  %s", formatTokenCount(session.TotalTokens)),
		fmt.Sprintf("Size:    %s", formatBytes(session.SizeBytes)),
	}
	if session.Subtitle != "" {
		lines = append(lines, fmt.Sprintf("Subtitle: %s", trimForDisplay(session.Subtitle, 64)))
	}
	if session.Source != "" {
		lines = append(lines, fmt.Sprintf("Source:   %s", trimForDisplay(session.Source, 64)))
	}
	lines = append(lines, "", "Session data", session.Path, "", "Preview", "")
	if !m.hasSelectedSessionPreview() {
		lines = append(lines, "No preview loaded.")
	} else {
		lines = append(lines, m.renderPreviewBlock(m.previewText, m.sessionPreviewLineLimit())...)
	}
	return strings.Join(lines, "\n")
}

func renderSessionPreviewPlain(m analyzeModel) string {
	if strings.TrimSpace(m.previewText) == "" {
		return "No preview loaded."
	}
	return strings.Join(m.renderPreviewBlock(m.previewText, m.height-8), "\n")
}

func renderCandidateListPlain(m analyzeModel) string {
	lines := []string{"Leftover Items", ""}
	if len(m.candidates) == 0 {
		return strings.Join(append(lines, "No leftover items found."), "\n")
	}
	start, end := visibleWindow(len(m.candidates), m.candidateIndex, m.sessionListVisibleCount())
	for i := start; i < end; i++ {
		c := m.candidates[i]
		prefix := "  "
		if i == m.candidateIndex {
			prefix = "[#7cd3ff]>[-] "
		}
		lines = append(lines, fmt.Sprintf("%s%-18s %-8s %8s", prefix, trimForDisplay(displayKind(c.Kind), 18), string(c.Safety), formatBytes(c.SizeBytes)))
	}
	return strings.Join(lines, "\n")
}

func renderCandidateDetailPlain(m analyzeModel) string {
	if len(m.candidates) == 0 {
		return "No leftover item selected."
	}
	c := m.candidates[m.candidateIndex]
	lines := []string{
		displayKind(c.Kind),
		"",
		fmt.Sprintf("Assistant: %s", displayAssistant(c.Assistant)),
		fmt.Sprintf("Safety:    %s", c.Safety),
		fmt.Sprintf("Size:      %s", formatBytes(c.SizeBytes)),
		"",
		"Path",
		c.Path,
		"",
		"Reason",
		c.Reason,
	}
	return strings.Join(lines, "\n")
}
