package cleaner

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

func runAnalyze(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("analyze", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var assistant string
	var mode string
	var jsonOutput bool
	var beforeRaw string

	fs.StringVar(&assistant, "assistant", "", "start directly with one assistant: openclaw, ironclaw, ollama")
	fs.StringVar(&mode, "mode", "human", "analyze currently supports human mode only")
	fs.BoolVar(&jsonOutput, "json", false, "analyze currently supports human mode only")
	fs.StringVar(&beforeRaw, "before", "", "pre-filter OpenClaw conversations before YYYY-MM-DD")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	mode = normalizeMode(mode, jsonOutput)
	if jsonOutput || mode == "agent" || !isTerminal(os.Stdin) {
		fmt.Fprintln(stderr, "analyze currently supports interactive human mode only")
		return 2
	}

	var assistants []string
	var err error
	if strings.TrimSpace(assistant) == "" {
		assistants = defaultAssistants()
	} else {
		assistants, err = parseAssistantList(assistant)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
	}

	var before time.Time
	if strings.TrimSpace(beforeRaw) != "" {
		before, err = parseDateCutoff(beforeRaw)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
	}

	if err := runAnalyzeTUI(assistants, before, stdout, stderr); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

type assistantSummary struct {
	Assistant     string
	SessionCount  int
	LeftoverCount int
}

func assistantAnalyzeSummary(assistants []string) ([]assistantSummary, error) {
	out := make([]assistantSummary, 0, len(assistants))
	for _, assistant := range assistants {
		leftovers, err := discoverAssistantLeftovers(assistant)
		if err != nil {
			return nil, err
		}
		item := assistantSummary{
			Assistant:     assistant,
			LeftoverCount: len(leftovers),
		}
		if assistant == "openclaw" {
			sessions, err := discoverOpenClawSessions()
			if err != nil {
				return nil, err
			}
			item.SessionCount = len(sessions)
		}
		out = append(out, item)
	}
	return out, nil
}

func discoverAssistantLeftovers(assistant string) ([]Candidate, error) {
	candidates, err := discoverCandidates([]string{assistant})
	if err != nil {
		return nil, err
	}

	out := make([]Candidate, 0, len(candidates))
	for _, candidate := range candidates {
		if assistant == "openclaw" && candidate.Kind == "session_store" {
			continue
		}
		out = append(out, candidate)
	}
	return out, nil
}

func liveCandidates(candidates []Candidate) []Candidate {
	out := make([]Candidate, 0, len(candidates))
	for _, candidate := range candidates {
		if pathExists(candidate.Path) {
			candidate.SizeBytes = pathSize(candidate.Path)
			out = append(out, candidate)
		}
	}
	return out
}

func parseDateCutoff(raw string) (time.Time, error) {
	t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(raw), time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q, expected YYYY-MM-DD", raw)
	}
	return t, nil
}

func formatTokenCount(tokens int64) string {
	if tokens <= 0 {
		return "-"
	}
	if tokens >= 1000 {
		return fmt.Sprintf("%.1fK tok", float64(tokens)/1000)
	}
	return fmt.Sprintf("%d tok", tokens)
}

func formatSessionTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Local().Format("2006-01-02 15:04")
}
