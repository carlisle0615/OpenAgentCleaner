package cleaner

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func discoverCandidates(assistants []string) ([]Candidate, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	all := make([]Candidate, 0, 48)
	for _, assistant := range assistants {
		verbosef("scanning leftover candidates for %s", displayAssistant(assistant))
		switch assistant {
		case "openclaw":
			all = append(all, discoverOpenClaw(home)...)
		case "ironclaw":
			all = append(all, discoverIronClaw(home)...)
		case "ollama":
			all = append(all, discoverOllama(home)...)
		case "codex":
			all = append(all, discoverCodexDesktop(home)...)
		case "codex-cli":
			all = append(all, discoverCodexCLI(home)...)
		case "claudecode":
			all = append(all, discoverClaudeCode(home)...)
		case "cursor":
			all = append(all, discoverCursor(home)...)
		case "antigravity":
			all = append(all, discoverAntigravity(home)...)
		}
	}

	all = dedupeCandidates(all)
	for i := range all {
		all[i].Path = cleanPath(all[i].Path)
		all[i].ID = candidateID(all[i].Assistant, all[i].Kind, all[i].Path)
		all[i].SizeBytes = pathSize(all[i].Path)
		all[i].Deletable = all[i].Safety != SafetyManual
		all[i].RequiresConfirmation = all[i].Safety == SafetyConfirm
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Assistant != all[j].Assistant {
			return all[i].Assistant < all[j].Assistant
		}
		if all[i].Safety != all[j].Safety {
			return all[i].Safety < all[j].Safety
		}
		return all[i].Path < all[j].Path
	})
	return all, nil
}

func appendCandidateIfExists(dst []Candidate, candidate Candidate) []Candidate {
	if pathExists(candidate.Path) {
		dst = append(dst, candidate)
	}
	return dst
}

func appendGlobCandidates(dst []Candidate, assistant, pattern, kind string, safety Safety, reason string) []Candidate {
	matches, _ := filepath.Glob(pattern)
	for _, match := range matches {
		if pathExists(match) {
			dst = append(dst, Candidate{
				Assistant: assistant,
				Path:      match,
				Kind:      kind,
				Safety:    safety,
				Reason:    reason,
			})
		}
	}
	return dst
}

func pathExists(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	_, err := os.Lstat(path)
	return err == nil
}

func pathSize(path string) int64 {
	info, err := os.Lstat(path)
	if err != nil {
		return 0
	}
	if !info.IsDir() {
		return info.Size()
	}

	var total int64
	filepath.WalkDir(path, func(walkPath string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		total += info.Size()
		return nil
	})
	return total
}

func dedupeCandidates(candidates []Candidate) []Candidate {
	seen := map[string]struct{}{}
	out := make([]Candidate, 0, len(candidates))
	for _, candidate := range candidates {
		key := candidate.Assistant + "\x00" + cleanPath(candidate.Path)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, candidate)
	}
	return out
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
