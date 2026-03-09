package cleaner

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type humanUI struct {
	w     io.Writer
	color bool
}

func newHumanUI(w io.Writer) humanUI {
	return humanUI{
		w:     w,
		color: supportsColor(w),
	}
}

func (ui humanUI) banner(title, subtitle string) {
	fmt.Fprintln(ui.w, ui.emphasis(title))
	if subtitle != "" {
		fmt.Fprintln(ui.w, subtitle)
	}
	fmt.Fprintln(ui.w)
}

func (ui humanUI) section(title, subtitle string) {
	fmt.Fprintln(ui.w, ui.emphasis(title))
	if subtitle != "" {
		fmt.Fprintln(ui.w, subtitle)
	}
}

func (ui humanUI) badgeOK(label string) string {
	return ui.badge(label, "32")
}

func (ui humanUI) badgeWarn(label string) string {
	return ui.badge(label, "33")
}

func (ui humanUI) badgeError(label string) string {
	return ui.badge(label, "31")
}

func (ui humanUI) badgeInfo(label string) string {
	return ui.badge(label, "36")
}

func (ui humanUI) badgeMuted(label string) string {
	return ui.badge(label, "2")
}

func (ui humanUI) statusBadge(status string) string {
	switch status {
	case "removed":
		return ui.badgeOK("Removed")
	case "planned":
		return ui.badgeInfo("Planned")
	case "selected":
		return ui.badgeInfo("Selected")
	case "skipped":
		return ui.badgeWarn("Skipped")
	case "error":
		return ui.badgeError("Error")
	default:
		return ui.badgeMuted("Found")
	}
}

func (ui humanUI) muted(text string) string {
	return ui.wrap(text, "2")
}

func (ui humanUI) emphasis(text string) string {
	return ui.wrap(text, "1")
}

func (ui humanUI) badge(label, code string) string {
	return ui.wrap("["+label+"]", code)
}

func (ui humanUI) wrap(text, code string) string {
	if !ui.color {
		return text
	}
	return "\033[" + code + "m" + text + "\033[0m"
}

func supportsColor(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if strings.EqualFold(os.Getenv("TERM"), "dumb") {
		return false
	}
	file, ok := w.(*os.File)
	return ok && isTerminal(file)
}
