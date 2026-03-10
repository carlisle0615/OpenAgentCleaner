package cleaner

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestManualAnalyzeLatency(t *testing.T) {
	if os.Getenv("OAC_MANUAL_ANALYZE_LATENCY") != "1" {
		t.Skip("set OAC_MANUAL_ANALYZE_LATENCY=1 to run manual analyze latency diagnostics")
	}

	rawAssistants := strings.TrimSpace(os.Getenv("OAC_MANUAL_ASSISTANTS"))
	if rawAssistants == "" {
		rawAssistants = "codex,codex-cli,claudecode,cursor,antigravity,openclaw"
	}

	assistants, err := parseAssistantList(rawAssistants)
	if err != nil {
		t.Fatalf("parseAssistantList(%q) err = %v", rawAssistants, err)
	}

	t.Logf("manual analyze latency assistants=%s", strings.Join(assistants, ","))
	for _, assistant := range assistants {
		start := time.Now()
		model, err := newAnalyzeModel([]string{assistant}, time.Time{})
		initElapsed := time.Since(start)
		if err != nil {
			t.Fatalf("newAnalyzeModel(%s) err = %v", assistant, err)
		}

		t.Logf(
			"assistant=%s init=%s screen=%v sessions=%d leftovers=%d preview_loaded=%t",
			assistant,
			initElapsed.Round(time.Millisecond),
			model.screen,
			len(model.sessions),
			len(model.candidates),
			model.hasSelectedSessionPreview(),
		)

		if model.screen == screenAssistantMenu {
			model.assistantMenuIndex = 0
			start = time.Now()
			err = model.activateSelection()
			openElapsed := time.Since(start)
			if err != nil {
				t.Fatalf("activateSelection(%s conversations) err = %v", assistant, err)
			}
			t.Logf(
				"assistant=%s open_conversations=%s screen=%v sessions=%d preview_loaded=%t",
				assistant,
				openElapsed.Round(time.Millisecond),
				model.screen,
				len(model.sessions),
				model.hasSelectedSessionPreview(),
			)
		}

		if model.screen == screenSessions && len(model.sessions) > 0 {
			start = time.Now()
			err = model.activateSelection()
			openPreviewElapsed := time.Since(start)
			if err != nil {
				t.Fatalf("activateSelection(%s preview) err = %v", assistant, err)
			}
			t.Logf(
				"assistant=%s open_preview=%s screen=%v preview_loaded=%t preview_chars=%d",
				assistant,
				openPreviewElapsed.Round(time.Millisecond),
				model.screen,
				strings.TrimSpace(model.previewText) != "",
				len(model.previewText),
			)
		}
	}
}
