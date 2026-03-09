package cleaner

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestHumanUIAndColorSupport(t *testing.T) {
	if supportsColor(&bytes.Buffer{}) {
		t.Fatal("supportsColor(buffer) should be false")
	}

	t.Setenv("NO_COLOR", "1")
	if supportsColor(os.Stdout) {
		t.Fatal("supportsColor(NO_COLOR) should be false")
	}

	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "dumb")
	if supportsColor(os.Stdout) {
		t.Fatal("supportsColor(TERM=dumb) should be false")
	}
	t.Setenv("TERM", "xterm-256color")

	ui := newHumanUI(&bytes.Buffer{})
	if ui.color {
		t.Fatal("newHumanUI(buffer) should disable color")
	}

	devNull, err := os.Open("/dev/null")
	if err != nil {
		t.Fatalf("open /dev/null: %v", err)
	}
	defer devNull.Close()
	if !supportsColor(devNull) {
		t.Fatal("supportsColor(char device) should be true")
	}

	colorUI := humanUI{w: devNull, color: true}
	if !strings.Contains(colorUI.wrap("text", "31"), "\033[31m") {
		t.Fatalf("wrap(color) = %q", colorUI.wrap("text", "31"))
	}

	if ui.badgeOK("OK") != "[OK]" || ui.badgeWarn("WARN") != "[WARN]" || ui.badgeError("ERR") != "[ERR]" {
		t.Fatal("badge helpers should render plain text without color")
	}
	if ui.badgeInfo("INFO") != "[INFO]" || ui.badgeMuted("MUTED") != "[MUTED]" {
		t.Fatal("badge variants should render plain text")
	}
	if ui.statusBadge("removed") != "[Removed]" || ui.statusBadge("planned") != "[Planned]" || ui.statusBadge("selected") != "[Selected]" || ui.statusBadge("skipped") != "[Skipped]" || ui.statusBadge("error") != "[Error]" || ui.statusBadge("found") != "[Found]" {
		t.Fatal("statusBadge() mismatch")
	}
	if ui.muted("text") != "text" || ui.emphasis("text") != "text" || ui.wrap("text", "31") != "text" {
		t.Fatal("plain-text wrappers should be no-op")
	}

	var out bytes.Buffer
	ui = newHumanUI(&out)
	ui.banner("Title", "Subtitle")
	ui.section("Section", "More")
	text := out.String()
	if !strings.Contains(text, "Title") || !strings.Contains(text, "Subtitle") || !strings.Contains(text, "Section") || !strings.Contains(text, "More") {
		t.Fatalf("banner/section output = %q", text)
	}
}
