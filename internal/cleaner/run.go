package cleaner

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		if isInteractiveSession() {
			return runHomeMenu(stdout, stderr)
		}
		printRootHelp(stdout)
		return 0
	}

	switch args[0] {
	case "scan":
		return runScan(args[1:], stdout, stderr)
	case "clean":
		return runClean(args[1:], stdout, stderr)
	case "-h", "--help", "help":
		printRootHelp(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", args[0])
		printRootHelp(stderr)
		return 2
	}
}

func runHomeMenu(stdout, stderr io.Writer) int {
	reader := bufio.NewReader(os.Stdin)

	for {
		cmd := commandName()
		fmt.Fprintf(stdout, "%s\n\n", cmd)
		fmt.Fprintln(stdout, "1. Scan all supported assistants")
		fmt.Fprintln(stdout, "2. Clean safe items")
		fmt.Fprintln(stdout, "3. Clean safe and confirm items")
		fmt.Fprintln(stdout, "4. Show help")
		fmt.Fprintln(stdout, "5. Quit")
		fmt.Fprint(stdout, "\nSelect action [1-5]: ")

		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}

		switch strings.ToLower(strings.TrimSpace(line)) {
		case "1", "scan", "s":
			return runScan(nil, stdout, stderr)
		case "2", "clean", "c":
			return runClean(nil, stdout, stderr)
		case "3", "clean-all", "confirm", "a":
			return runClean([]string{"--include-confirm"}, stdout, stderr)
		case "4", "help", "h":
			printRootHelp(stdout)
			fmt.Fprintln(stdout)
		case "5", "quit", "q", "exit":
			return 0
		default:
			fmt.Fprintf(stderr, "unknown action %q\n\n", strings.TrimSpace(line))
		}
	}
}

func runScan(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var rawAssistants string
	var mode string
	var jsonOutput bool

	fs.StringVar(&rawAssistants, "assistants", strings.Join(defaultAssistants(), ","), "comma-separated assistants: openclaw,ironclaw,ollama")
	fs.StringVar(&mode, "mode", "auto", "output mode: auto, human, agent")
	fs.BoolVar(&jsonOutput, "json", false, "emit JSON")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	assistants, err := parseAssistantList(rawAssistants)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	report, err := scanReport(options{
		Assistants: assistants,
		Mode:       mode,
		JSON:       jsonOutput,
		DryRun:     true,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	if report.Mode == "agent" || jsonOutput {
		fmt.Fprintln(stdout, toJSON(report))
		return 0
	}
	printHumanReport(stdout, report, false)
	return 0
}

func runClean(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("clean", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var rawAssistants string
	var mode string
	var jsonOutput bool
	var includeConfirm bool
	var yes bool
	var dryRun bool

	fs.StringVar(&rawAssistants, "assistants", strings.Join(defaultAssistants(), ","), "comma-separated assistants: openclaw,ironclaw,ollama")
	fs.StringVar(&mode, "mode", "auto", "output mode: auto, human, agent")
	fs.BoolVar(&jsonOutput, "json", false, "emit JSON")
	fs.BoolVar(&includeConfirm, "include-confirm", false, "include items that require explicit confirmation")
	fs.BoolVar(&yes, "yes", false, "skip prompts and delete all eligible items")
	fs.BoolVar(&dryRun, "dry-run", false, "preview cleanup without deleting")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	assistants, err := parseAssistantList(rawAssistants)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	report, err := cleanReport(options{
		Assistants:     assistants,
		Mode:           mode,
		JSON:           jsonOutput,
		IncludeConfirm: includeConfirm,
		Yes:            yes,
		DryRun:         dryRun,
	}, stdout, stderr)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	if report.Mode == "agent" || jsonOutput {
		fmt.Fprintln(stdout, toJSON(report))
		return 0
	}
	printHumanReport(stdout, report, true)
	return 0
}

func scanReport(opts options) (Report, error) {
	mode := normalizeMode(opts.Mode, opts.JSON)
	candidates, err := discoverCandidates(opts.Assistants)
	if err != nil {
		return Report{}, err
	}
	report := Report{
		Operation:  "scan",
		Mode:       mode,
		DryRun:     true,
		Assistants: opts.Assistants,
		Candidates: candidates,
	}
	report.Summary = summarize(report.Candidates)
	return report, nil
}

func cleanReport(opts options, stdout, stderr io.Writer) (Report, error) {
	mode := normalizeMode(opts.Mode, opts.JSON)
	report, err := scanReport(opts)
	if err != nil {
		return Report{}, err
	}
	report.Operation = "clean"
	report.Mode = mode
	report.DryRun = opts.DryRun

	eligibleIndexes := eligibleCandidates(report.Candidates, opts.IncludeConfirm)
	if len(eligibleIndexes) == 0 {
		return report, nil
	}

	selected := map[int]struct{}{}
	switch {
	case opts.DryRun:
		for _, idx := range eligibleIndexes {
			selected[idx] = struct{}{}
			report.Candidates[idx].Selected = true
			report.Candidates[idx].Skipped = true
		}
	case mode == "agent" || !isTerminal(os.Stdin):
		if !opts.Yes {
			return Report{}, errors.New("non-interactive cleanup requires --yes or --dry-run")
		}
		for _, idx := range eligibleIndexes {
			selected[idx] = struct{}{}
			report.Candidates[idx].Selected = true
		}
	default:
		chosen, err := promptSelection(stdout, stderr, report.Candidates, eligibleIndexes)
		if err != nil {
			return Report{}, err
		}
		for _, idx := range chosen {
			selected[idx] = struct{}{}
			report.Candidates[idx].Selected = true
		}
		if len(selected) == 0 {
			report.Summary = summarize(report.Candidates)
			return report, nil
		}
		if !confirmDeletion(stdout, stderr, report.Candidates, selected) {
			for idx := range selected {
				report.Candidates[idx].Skipped = true
			}
			report.Summary = summarize(report.Candidates)
			return report, nil
		}
	}

	if !opts.DryRun {
		for idx := range selected {
			if err := deletePath(report.Candidates[idx].Path); err != nil {
				report.Candidates[idx].Error = err.Error()
				continue
			}
			report.Candidates[idx].Deleted = true
		}
	}

	report.Summary = summarize(report.Candidates)
	return report, nil
}

func eligibleCandidates(candidates []Candidate, includeConfirm bool) []int {
	indexes := make([]int, 0, len(candidates))
	for i, candidate := range candidates {
		switch candidate.Safety {
		case SafetySafe:
			indexes = append(indexes, i)
		case SafetyConfirm:
			if includeConfirm {
				indexes = append(indexes, i)
			}
		}
	}
	return indexes
}

func promptSelection(stdout, stderr io.Writer, candidates []Candidate, eligible []int) ([]int, error) {
	fmt.Fprintln(stdout, "Eligible cleanup targets:")
	for i, idx := range eligible {
		candidate := candidates[idx]
		fmt.Fprintf(stdout, "  [%d] %-10s %-14s %-7s %8s  %s\n", i+1, candidate.Assistant, candidate.Kind, candidate.Safety, formatBytes(candidate.SizeBytes), candidate.Path)
	}
	fmt.Fprintln(stdout, "\nEnter indexes to delete (`all`, `none`, or comma-separated values):")

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	line = strings.ToLower(strings.TrimSpace(line))
	switch line {
	case "", "none", "n":
		return nil, nil
	case "all", "a":
		return eligible, nil
	}

	parts := strings.Split(line, ",")
	chosen := make([]int, 0, len(parts))
	seen := map[int]struct{}{}
	for _, part := range parts {
		value, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || value < 1 || value > len(eligible) {
			return nil, fmt.Errorf("invalid selection %q", part)
		}
		idx := eligible[value-1]
		if _, ok := seen[idx]; ok {
			continue
		}
		seen[idx] = struct{}{}
		chosen = append(chosen, idx)
	}
	return chosen, nil
}

func confirmDeletion(stdout, stderr io.Writer, candidates []Candidate, selected map[int]struct{}) bool {
	var count int
	var bytes int64
	for idx := range selected {
		count++
		bytes += candidates[idx].SizeBytes
	}
	fmt.Fprintf(stdout, "Delete %d target(s), reclaiming about %s? [y/N]: ", count, formatBytes(bytes))
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintln(stderr, err)
		return false
	}
	line = strings.ToLower(strings.TrimSpace(line))
	return line == "y" || line == "yes"
}

func deletePath(path string) error {
	path = cleanPath(path)
	if !filepath.IsAbs(path) {
		return fmt.Errorf("refusing to delete non-absolute path %q", path)
	}
	forbidden := map[string]struct{}{
		"/": {},
	}
	if _, ok := forbidden[path]; ok {
		return fmt.Errorf("refusing to delete protected path %q", path)
	}
	home, _ := os.UserHomeDir()
	if path == home {
		return fmt.Errorf("refusing to delete home directory")
	}
	if !pathExists(path) {
		return nil
	}
	return os.RemoveAll(path)
}

func summarize(candidates []Candidate) Summary {
	var summary Summary
	for _, candidate := range candidates {
		summary.CandidateCount++
		summary.BytesFound += candidate.SizeBytes
		if candidate.Selected {
			summary.SelectedCount++
		}
		if candidate.Deleted {
			summary.DeletedCount++
			summary.BytesDeleted += candidate.SizeBytes
		}
	}
	return summary
}

func printRootHelp(w io.Writer) {
	cmd := commandName()
	fmt.Fprintln(w, "OpenAgentCleaner")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintf(w, "  %s scan  [--assistants openclaw,ironclaw,ollama] [--mode auto|human|agent] [--json]\n", cmd)
	fmt.Fprintf(w, "  %s clean [--assistants openclaw,ironclaw,ollama] [--include-confirm] [--yes] [--dry-run] [--mode auto|human|agent] [--json]\n", cmd)
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Notes:")
	fmt.Fprintf(w, "  - Run `%s` with no arguments for the interactive menu.\n", cmd)
	fmt.Fprintln(w, "  - `scan` only discovers residue and classifications.")
	fmt.Fprintln(w, "  - `clean` removes only `safe` items by default.")
	fmt.Fprintln(w, "  - `--include-confirm` adds items that require explicit user intent.")
	fmt.Fprintln(w, "  - `manual` items are never auto-deleted in this version.")
}

func isInteractiveSession() bool {
	return isTerminal(os.Stdin) && isTerminal(os.Stdout)
}

func commandName() string {
	name := filepath.Base(os.Args[0])
	switch name {
	case "", ".", "OpenAgentCleaner":
		return "oac"
	default:
		return name
	}
}
