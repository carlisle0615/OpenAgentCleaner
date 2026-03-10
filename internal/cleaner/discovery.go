package cleaner

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/carlisle0615/OpenAgentCleaner/internal/cleaner/discoveryrules"
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
			all = append(all, convertDiscoveryCandidates(discoveryrules.DiscoverOpenClaw(home))...)
		case "ironclaw":
			all = append(all, convertDiscoveryCandidates(discoveryrules.DiscoverIronClaw(home))...)
		case "ollama":
			all = append(all, convertDiscoveryCandidates(discoveryrules.DiscoverOllama(home))...)
		case "codex":
			all = append(all, convertDiscoveryCandidates(discoveryrules.DiscoverCodexDesktop(home))...)
		case "codex-cli":
			all = append(all, convertDiscoveryCandidates(discoveryrules.DiscoverCodexCLI(home))...)
		case "claudecode":
			all = append(all, convertDiscoveryCandidates(discoveryrules.DiscoverClaudeCode(home))...)
		case "cursor":
			all = append(all, convertDiscoveryCandidates(discoveryrules.DiscoverCursor(home))...)
		case "antigravity":
			all = append(all, convertDiscoveryCandidates(discoveryrules.DiscoverAntigravity(home))...)
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

func convertDiscoveryCandidates(items []discoveryrules.Candidate) []Candidate {
	out := make([]Candidate, 0, len(items))
	for _, item := range items {
		out = append(out, Candidate{
			Assistant: item.Assistant,
			Path:      item.Path,
			Kind:      item.Kind,
			Safety:    Safety(item.Safety),
			Reason:    item.Reason,
			Notes:     append([]string(nil), item.Notes...),
		})
	}
	return out
}

func discoverOpenClaw(home string) []Candidate {
	return convertDiscoveryCandidates(discoveryrules.DiscoverOpenClaw(home))
}

func discoverIronClaw(home string) []Candidate {
	return convertDiscoveryCandidates(discoveryrules.DiscoverIronClaw(home))
}

func discoverOllama(home string) []Candidate {
	return convertDiscoveryCandidates(discoveryrules.DiscoverOllama(home))
}

func discoverCodexDesktop(home string) []Candidate {
	return convertDiscoveryCandidates(discoveryrules.DiscoverCodexDesktop(home))
}

func discoverCodexCLI(home string) []Candidate {
	return convertDiscoveryCandidates(discoveryrules.DiscoverCodexCLI(home))
}

func discoverClaudeCode(home string) []Candidate {
	return convertDiscoveryCandidates(discoveryrules.DiscoverClaudeCode(home))
}

func discoverCursor(home string) []Candidate {
	return convertDiscoveryCandidates(discoveryrules.DiscoverCursor(home))
}

func discoverAntigravity(home string) []Candidate {
	return convertDiscoveryCandidates(discoveryrules.DiscoverAntigravity(home))
}

func isOpenClawStateRoot(path string) bool {
	return discoveryrules.IsOpenClawStateRoot(path)
}

func openClawStateRoots(home string) []string {
	return discoveryrules.OpenClawStateRoots(home)
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
