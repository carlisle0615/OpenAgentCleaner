package cleaner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscoverCandidatesAndHelpers(t *testing.T) {
	home := setTestHome(t)

	openclawRoot := filepath.Join(home, ".openclaw")
	openclawAlt := filepath.Join(home, ".openclaw-alt")
	openclawConfig := filepath.Join(home, "custom", "openclaw.json")
	t.Setenv("OPENCLAW_CONFIG_PATH", openclawConfig)
	writeTestFile(t, filepath.Join(openclawRoot, "logs", "app.log"), "log")
	writeTestFile(t, filepath.Join(openclawRoot, "agents", "main", "sessions", "sessions.json"), "{}\n")
	writeTestFile(t, filepath.Join(openclawRoot, ".env"), "OPENCLAW=1\n")
	writeTestFile(t, filepath.Join(openclawRoot, "extensions", "plugin.txt"), "plugin")
	writeTestFile(t, filepath.Join(openclawRoot, "workspace", "notes.md"), "notes")
	writeTestFile(t, openclawConfig, "{}\n")
	writeTestFile(t, filepath.Join(openclawAlt, "workspace", "agents.md"), "alt")
	writeTestFile(t, filepath.Join(home, "Library", "LaunchAgents", "ai.openclaw.agent.plist"), "plist")
	writeTestFile(t, filepath.Join(home, "Library", "LaunchAgents", "bot.molt.agent.plist"), "plist")
	writeTestFile(t, filepath.Join(home, "Library", "LaunchAgents", "com.openclaw.legacy.plist"), "plist")

	tmpLogDir := filepath.Join("/tmp", "openclaw")
	if err := os.MkdirAll(tmpLogDir, 0o755); err != nil {
		t.Fatalf("mkdir tmp logs: %v", err)
	}
	tmpLog := filepath.Join(tmpLogDir, "test-openclaw.log")
	if err := os.WriteFile(tmpLog, []byte("tmp"), 0o644); err != nil {
		t.Fatalf("write tmp log: %v", err)
	}
	defer os.Remove(tmpLog)

	ironclawRoot := filepath.Join(home, "iron")
	t.Setenv("IRONCLAW_BASE_DIR", ironclawRoot)
	for _, path := range []string{
		"logs/daemon.log",
		".env",
		"ironclaw.db",
		"config.toml",
		"session.json",
		"mcp-servers.json",
		"settings.json",
		"bootstrap.json",
		"channels/ch1",
		"tools/tool1",
		"history/repl.txt",
		"device-pairing.json",
		"device-allowFrom.json",
		"device-approve-attempts.json",
	} {
		writeTestFile(t, filepath.Join(ironclawRoot, path), "x")
	}
	writeTestFile(t, filepath.Join(home, "Library", "LaunchAgents", "com.ironclaw.daemon.plist"), "plist")

	ollamaModels := filepath.Join(home, "ollama-models")
	t.Setenv("OLLAMA_MODELS", ollamaModels)
	for _, path := range []string{
		".ollama/logs/server.log",
		"Library/Saved Application State/com.electron.ollama.savedState/window",
		"Library/Caches/com.electron.ollama/cache.db",
		"Library/Caches/ollama/cache.db",
		"Library/WebKit/com.electron.ollama/index.db",
		"Library/Application Support/Ollama/state.json",
		".ollama/server.json",
		".ollama/id_ed25519",
		".ollama/id_ed25519.pub",
	} {
		writeTestFile(t, filepath.Join(home, path), "y")
	}
	writeTestFile(t, filepath.Join(ollamaModels, "manifests", "one"), "model")

	if !pathExists(filepath.Join(home, ".ollama")) {
		t.Fatal("pathExists should see created directory")
	}
	if pathExists(" ") {
		t.Fatal("pathExists should reject blank")
	}

	dirSize := pathSize(filepath.Join(home, ".ollama"))
	if dirSize <= 0 {
		t.Fatalf("pathSize(dir) = %d, want > 0", dirSize)
	}
	if pathSize(filepath.Join(home, "missing")) != 0 {
		t.Fatal("pathSize(missing) should be 0")
	}

	openclaw := discoverOpenClaw(home)
	if len(openclaw) == 0 {
		t.Fatal("discoverOpenClaw() returned no candidates")
	}
	ironclaw := discoverIronClaw(home)
	if len(ironclaw) == 0 {
		t.Fatal("discoverIronClaw() returned no candidates")
	}
	ollama := discoverOllama(home)
	if len(ollama) == 0 {
		t.Fatal("discoverOllama() returned no candidates")
	}

	all, err := discoverCandidates([]string{"openclaw", "ironclaw", "ollama"})
	if err != nil {
		t.Fatalf("discoverCandidates() err = %v", err)
	}
	if len(all) == 0 {
		t.Fatal("discoverCandidates() returned no candidates")
	}

	joined := make([]string, 0, len(all))
	for _, candidate := range all {
		joined = append(joined, candidate.Assistant+":"+candidate.Kind+":"+string(candidate.Safety))
	}
	got := strings.Join(joined, "\n")
	for _, want := range []string{
		"openclaw:gateway_logs:safe",
		"openclaw:workspace:manual",
		"ironclaw:local_database:confirm",
		"ironclaw:daemon_logs:safe",
		"ollama:cache:safe",
		"ollama:models:confirm",
		"ollama:auth_key:manual",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("discoverCandidates() missing %q\n%s", want, got)
		}
	}
}

func TestAppendCandidateIfExistsAndGlob(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "item.txt")
	writeTestFile(t, file, "x")

	var candidates []Candidate
	candidates = appendCandidateIfExists(candidates, Candidate{Path: file, Assistant: "openclaw"})
	candidates = appendCandidateIfExists(candidates, Candidate{Path: filepath.Join(root, "missing"), Assistant: "openclaw"})
	if len(candidates) != 1 {
		t.Fatalf("appendCandidateIfExists() len = %d, want 1", len(candidates))
	}

	candidates = appendGlobCandidates(nil, "ollama", filepath.Join(root, "*.txt"), "cache", SafetySafe, "reason")
	if len(candidates) != 1 || candidates[0].Kind != "cache" {
		t.Fatalf("appendGlobCandidates() = %#v", candidates)
	}
}

func TestDedupeUniqueAndStateRoots(t *testing.T) {
	home := setTestHome(t)
	mainRoot := filepath.Join(home, ".openclaw")
	altRoot := filepath.Join(home, ".openclaw-second")
	writeTestFile(t, filepath.Join(mainRoot, "openclaw.json"), "{}")
	writeTestFile(t, filepath.Join(altRoot, "workspace", "notes.md"), "x")

	deduped := dedupeCandidates([]Candidate{
		{Assistant: "openclaw", Path: filepath.Join(mainRoot, "logs", "..", "logs")},
		{Assistant: "openclaw", Path: filepath.Join(mainRoot, "logs")},
		{Assistant: "ollama", Path: filepath.Join(mainRoot, "logs")},
	})
	if len(deduped) != 2 {
		t.Fatalf("dedupeCandidates() len = %d, want 2", len(deduped))
	}

	unique := uniqueStrings([]string{"b", "", "a", "b"})
	if strings.Join(unique, ",") != "a,b" {
		t.Fatalf("uniqueStrings() = %v", unique)
	}

	if !isOpenClawStateRoot(mainRoot) {
		t.Fatal("isOpenClawStateRoot(mainRoot) = false")
	}
	if isOpenClawStateRoot(filepath.Join(home, "missing")) {
		t.Fatal("isOpenClawStateRoot(missing) = true")
	}

	t.Setenv("OPENCLAW_STATE_DIR", mainRoot)
	roots := openClawStateRoots(home)
	if strings.Join(roots, ",") != strings.Join([]string{mainRoot, altRoot}, ",") {
		t.Fatalf("openClawStateRoots(explicit) = %v", roots)
	}

	t.Setenv("OPENCLAW_STATE_DIR", "")
	roots = openClawStateRoots(home)
	if strings.Join(roots, ",") != strings.Join([]string{mainRoot, altRoot}, ",") {
		t.Fatalf("openClawStateRoots() = %v", roots)
	}
}
