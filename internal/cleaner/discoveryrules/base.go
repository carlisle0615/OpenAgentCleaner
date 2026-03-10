package discoveryrules

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Safety string

const (
	SafetySafe    Safety = "safe"
	SafetyConfirm Safety = "confirm"
	SafetyManual  Safety = "manual"
)

type Candidate struct {
	Assistant string
	Path      string
	Kind      string
	Safety    Safety
	Reason    string
	Notes     []string
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

func cleanPath(path string) string {
	return filepath.Clean(filepath.Clean(strings.TrimSpace(path)))
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
