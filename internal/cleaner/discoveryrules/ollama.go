package discoveryrules

import (
	"os"
	"path/filepath"
)

func DiscoverOllama(home string) []Candidate {
	baseDir := filepath.Join(home, ".ollama")
	modelsDir := cleanPath(os.Getenv("OLLAMA_MODELS"))
	if modelsDir == "." || modelsDir == "" {
		modelsDir = filepath.Join(baseDir, "models")
	}

	out := []Candidate{}
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ollama",
		Path:      filepath.Join(baseDir, "logs"),
		Kind:      "logs",
		Safety:    SafetySafe,
		Reason:    "Ollama app/server logs are disposable and recreated automatically.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ollama",
		Path:      filepath.Join(home, "Library", "Saved Application State", "com.electron.ollama.savedState"),
		Kind:      "saved_state",
		Safety:    SafetySafe,
		Reason:    "macOS saved window state can be safely removed.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ollama",
		Path:      filepath.Join(home, "Library", "Caches", "com.electron.ollama"),
		Kind:      "cache",
		Safety:    SafetySafe,
		Reason:    "Electron cache is disposable.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ollama",
		Path:      filepath.Join(home, "Library", "Caches", "ollama"),
		Kind:      "cache",
		Safety:    SafetySafe,
		Reason:    "App cache is disposable.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ollama",
		Path:      filepath.Join(home, "Library", "WebKit", "com.electron.ollama"),
		Kind:      "webkit_cache",
		Safety:    SafetySafe,
		Reason:    "Embedded WebKit cache is disposable.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ollama",
		Path:      filepath.Join(home, "Library", "Application Support", "Ollama"),
		Kind:      "app_support",
		Safety:    SafetyConfirm,
		Reason:    "Application Support may contain local UI state and account metadata.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ollama",
		Path:      modelsDir,
		Kind:      "models",
		Safety:    SafetyConfirm,
		Reason:    "Model blobs and manifests are large and valuable; require explicit confirmation.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ollama",
		Path:      filepath.Join(baseDir, "server.json"),
		Kind:      "config",
		Safety:    SafetyConfirm,
		Reason:    "Ollama configuration file should only be removed intentionally.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ollama",
		Path:      filepath.Join(baseDir, "id_ed25519"),
		Kind:      "auth_key",
		Safety:    SafetyManual,
		Reason:    "Private key can affect cloud publishing/auth flows and is not auto-removed.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ollama",
		Path:      filepath.Join(baseDir, "id_ed25519.pub"),
		Kind:      "auth_key",
		Safety:    SafetyManual,
		Reason:    "Public key is paired with Ollama auth state; remove manually if needed.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ollama",
		Path:      "/Applications/Ollama.app",
		Kind:      "app_bundle",
		Safety:    SafetyConfirm,
		Reason:    "Installed app bundle. Include this only when you want full product removal, not just leftovers.",
	})
	out = appendCandidateIfExists(out, Candidate{
		Assistant: "ollama",
		Path:      "/usr/local/bin/ollama",
		Kind:      "cli_symlink",
		Safety:    SafetyConfirm,
		Reason:    "CLI entrypoint is part of the installation, not just cached residue.",
	})
	return out
}
