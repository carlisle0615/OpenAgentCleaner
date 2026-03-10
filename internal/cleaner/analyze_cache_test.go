package cleaner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeDiscoveryCacheReuseAndInvalidation(t *testing.T) {
	home := setTestHome(t)
	resetAnalyzeDiscoveryCache()
	t.Cleanup(resetAnalyzeDiscoveryCache)

	openclawRoot := filepath.Join(home, ".openclaw")
	sessionsDir := filepath.Join(openclawRoot, "agents", "main", "sessions")
	metadataPath := filepath.Join(sessionsDir, "sessions.json")
	transcriptPath := filepath.Join(sessionsDir, "session-1.jsonl")
	writeTestFile(t, metadataPath, `{"one":{"sessionId":"session-1","updatedAt":1700000000000}}`)
	writeTestFile(t, transcriptPath, "{\"timestamp\":\"2026-02-07T03:16:10.650Z\"}\n")

	ollamaLogPath := filepath.Join(home, ".ollama", "logs", "server.log")
	writeTestFile(t, ollamaLogPath, "log")

	sessions, err := discoverAssistantSessionsCached("openclaw")
	if err != nil {
		t.Fatalf("discoverAssistantSessionsCached(openclaw) err = %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("discoverAssistantSessionsCached(openclaw) len = %d", len(sessions))
	}

	leftovers, err := discoverAssistantLeftoversCached("ollama")
	if err != nil {
		t.Fatalf("discoverAssistantLeftoversCached(ollama) err = %v", err)
	}
	initialLeftovers := len(leftovers)
	if initialLeftovers == 0 {
		t.Fatalf("discoverAssistantLeftoversCached(ollama) len = %d", len(leftovers))
	}

	if err := os.Remove(transcriptPath); err != nil {
		t.Fatalf("remove transcript err = %v", err)
	}
	if err := os.Remove(ollamaLogPath); err != nil {
		t.Fatalf("remove ollama log err = %v", err)
	}
	if err := os.Remove(filepath.Dir(ollamaLogPath)); err != nil {
		t.Fatalf("remove ollama logs dir err = %v", err)
	}

	sessions, err = discoverAssistantSessionsCached("openclaw")
	if err != nil {
		t.Fatalf("discoverAssistantSessionsCached(openclaw) second err = %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("cached sessions len = %d", len(sessions))
	}

	leftovers, err = discoverAssistantLeftoversCached("ollama")
	if err != nil {
		t.Fatalf("discoverAssistantLeftoversCached(ollama) second err = %v", err)
	}
	if len(leftovers) != initialLeftovers {
		t.Fatalf("cached leftovers len = %d", len(leftovers))
	}

	invalidateAssistantSessionsCache("openclaw")
	invalidateAssistantLeftoversCache("ollama")

	sessions, err = discoverAssistantSessionsCached("openclaw")
	if err != nil {
		t.Fatalf("discoverAssistantSessionsCached(openclaw) after invalidate err = %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("sessions after invalidate len = %d", len(sessions))
	}

	leftovers, err = discoverAssistantLeftoversCached("ollama")
	if err != nil {
		t.Fatalf("discoverAssistantLeftoversCached(ollama) after invalidate err = %v", err)
	}
	if len(leftovers) >= initialLeftovers {
		t.Fatalf("leftovers after invalidate len = %d", len(leftovers))
	}
}
