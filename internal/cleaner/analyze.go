package cleaner

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

func runAnalyze(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("analyze", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var assistant string
	var mode string
	var jsonOutput bool
	var beforeRaw string
	var verbose bool

	fs.StringVar(&assistant, "assistant", "", "start directly with one assistant from --assistants")
	fs.StringVar(&mode, "mode", "human", "analyze currently supports human mode only")
	fs.BoolVar(&jsonOutput, "json", false, "analyze currently supports human mode only")
	fs.StringVar(&beforeRaw, "before", "", "pre-filter assistant conversations before YYYY-MM-DD")
	fs.BoolVar(&verbose, "verbose", false, "show scan progress on stderr")
	fs.BoolVar(&verbose, "v", false, "show scan progress on stderr")

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
	setVerboseLogger(verbose, stderr)
	defer resetVerboseLogger()

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
	type result struct {
		index   int
		summary assistantSummary
		err     error
	}

	results := make([]assistantSummary, len(assistants))
	errCh := make(chan error, 1)
	var wg sync.WaitGroup

	for i, assistant := range assistants {
		wg.Add(1)
		go func(index int, assistant string) {
			defer wg.Done()

			leftovers, err := discoverAssistantLeftoversCached(assistant)
			if err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}

			item := assistantSummary{
				Assistant:     assistant,
				LeftoverCount: len(leftovers),
			}
			if assistantSupportsSessions(assistant) {
				sessions, err := discoverAssistantSessionsCached(assistant)
				if err != nil {
					select {
					case errCh <- err:
					default:
					}
					return
				}
				item.SessionCount = len(sessions)
			}
			verbosef("summary for %s: %d conversation(s), %d leftover item(s)", displayAssistant(assistant), item.SessionCount, item.LeftoverCount)
			results[index] = item
		}(i, assistant)
	}

	wg.Wait()
	select {
	case err := <-errCh:
		return nil, err
	default:
	}
	return results, nil
}

func discoverAssistantLeftovers(assistant string) ([]Candidate, error) {
	candidates, err := discoverCandidates([]string{assistant})
	if err != nil {
		return nil, err
	}

	out := make([]Candidate, 0, len(candidates))
	for _, candidate := range candidates {
		if ignoredKinds := sessionIgnoredCandidateKinds(assistant); ignoredKinds != nil {
			if _, ok := ignoredKinds[candidate.Kind]; ok {
				continue
			}
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
