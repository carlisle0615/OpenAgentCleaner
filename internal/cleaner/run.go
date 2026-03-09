package cleaner

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
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
	case "analyze":
		return runAnalyze(args[1:], stdout, stderr)
	case "scan":
		return runScan(args[1:], stdout, stderr)
	case "clean":
		return runClean(args[1:], stdout, stderr)
	case "version", "--version", "-v":
		printVersion(stdout)
		return 0
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
	ui := newHumanUI(stdout)

	for {
		cmd := commandName()
		ui.banner(appName, "Clean leftover AI assistant files on your Mac without guessing what is safe to remove.")
		fmt.Fprintf(stdout, "%s First step: run a scan. A scan never deletes anything.\n\n", ui.badgeMuted("Recommended"))
		fmt.Fprintln(stdout, "1. Scan this Mac")
		fmt.Fprintln(stdout, "2. Analyze conversations and leftovers")
		fmt.Fprintln(stdout, "3. Preview cleanup for safe items")
		fmt.Fprintln(stdout, "4. Remove safe leftovers")
		fmt.Fprintln(stdout, "5. Guided full cleanup")
		fmt.Fprintln(stdout, "6. Show command help")
		fmt.Fprintln(stdout, "7. Quit")
		fmt.Fprintf(stdout, "\nChoose an option [1-7] (%s): ", cmd+" scan")

		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}

		switch strings.ToLower(strings.TrimSpace(line)) {
		case "", "1", "scan", "s":
			return runScan(nil, stdout, stderr)
		case "2", "analyze", "a":
			return runAnalyze(nil, stdout, stderr)
		case "3", "preview", "p":
			return runClean([]string{"--dry-run"}, stdout, stderr)
		case "4", "clean", "c":
			return runClean(nil, stdout, stderr)
		case "5", "clean-all", "confirm", "full":
			return runClean([]string{"--include-confirm"}, stdout, stderr)
		case "6", "help", "h":
			printRootHelp(stdout)
			fmt.Fprintln(stdout)
		case "7", "quit", "q", "exit":
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
		chosen, err := promptSelection(stdout, stderr, report.Candidates, eligibleIndexes, opts.IncludeConfirm)
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

func promptSelection(stdout, stderr io.Writer, candidates []Candidate, eligible []int, includeConfirm bool) ([]int, error) {
	ui := newHumanUI(stdout)
	safeOnly := filterIndexesBySafety(candidates, eligible, SafetySafe)
	confirmOnly := filterIndexesBySafety(candidates, eligible, SafetyConfirm)

	ui.section("Choose what to remove", "Nothing will be deleted until you confirm.")
	for i, idx := range eligible {
		candidate := candidates[idx]
		fmt.Fprintf(stdout, "  [%d] %-18s %-10s %8s  %s\n", i+1, displayKind(candidate.Kind), displaySafety(candidate.Safety), formatBytes(candidate.SizeBytes), candidate.Path)
	}

	fmt.Fprintln(stdout)
	if includeConfirm && len(confirmOnly) > 0 {
		fmt.Fprintf(stdout, "%s Press Enter to remove only the recommended safe items.\n", ui.badgeMuted("Recommended"))
		fmt.Fprintln(stdout, "Type `safe` for safe items only, `all` for everything listed, `none` to cancel, or `1,3` to choose specific items.")
	} else {
		fmt.Fprintf(stdout, "%s Press Enter to remove all safe items shown above.\n", ui.badgeMuted("Recommended"))
		fmt.Fprintln(stdout, "Type `all`, `none`, or `1,3` to choose specific items.")
	}
	fmt.Fprint(stdout, "Your choice: ")

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	line = strings.ToLower(strings.TrimSpace(line))
	switch line {
	case "":
		if includeConfirm && len(confirmOnly) > 0 {
			if len(safeOnly) == 0 {
				return nil, nil
			}
			return safeOnly, nil
		}
		return eligible, nil
	case "safe":
		return safeOnly, nil
	case "none", "n":
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
	var requiresStrongConfirm bool
	selectedIndexes := make([]int, 0, len(selected))
	for idx := range selected {
		count++
		bytes += candidates[idx].SizeBytes
		selectedIndexes = append(selectedIndexes, idx)
		if candidates[idx].Safety == SafetyConfirm {
			requiresStrongConfirm = true
		}
	}

	sort.Ints(selectedIndexes)
	ui := newHumanUI(stdout)
	ui.section("Final confirmation", "These files will be permanently removed from this Mac.")
	for _, idx := range selectedIndexes {
		candidate := candidates[idx]
		fmt.Fprintf(stdout, "  - %-18s %-10s %8s  %s\n", displayKind(candidate.Kind), displaySafety(candidate.Safety), formatBytes(candidate.SizeBytes), candidate.Path)
	}
	fmt.Fprintln(stdout)
	fmt.Fprintf(stdout, "This will remove %d item(s) and reclaim about %s.\n", count, formatBytes(bytes))
	confirmationWord := "yes"
	if requiresStrongConfirm {
		fmt.Fprintf(stdout, "%s Some selected items may contain models, sessions, or saved settings.\n", ui.badgeWarn("Review"))
		confirmationWord = "delete"
	}
	fmt.Fprintf(stdout, "Type %q to continue: ", confirmationWord)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintln(stderr, err)
		return false
	}
	line = strings.ToLower(strings.TrimSpace(line))
	if confirmationWord == "delete" {
		return line == "delete"
	}
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
	ui := newHumanUI(w)
	ui.banner(appName, "A guided macOS cleaner for leftover AI assistant files.")
	fmt.Fprintf(w, "Command: %s\nVersion: %s\n\n", cmd, Version)
	fmt.Fprintln(w, "Start here:")
	fmt.Fprintf(w, "  %s\n", cmd)
	fmt.Fprintln(w, "  Opens the guided home screen.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Core commands:")
	fmt.Fprintf(w, "  %s scan\n", cmd)
	fmt.Fprintln(w, "  Scan your Mac and show what can be cleaned.")
	fmt.Fprintf(w, "  %s analyze\n", cmd)
	fmt.Fprintln(w, "  Browse conversations and leftovers, then delete specific items interactively.")
	fmt.Fprintf(w, "  %s clean\n", cmd)
	fmt.Fprintln(w, "  Remove only safe leftovers.")
	fmt.Fprintf(w, "  %s clean --include-confirm\n", cmd)
	fmt.Fprintln(w, "  Include items that may contain sessions, settings, or models.")
	fmt.Fprintf(w, "  %s version\n", cmd)
	fmt.Fprintln(w, "  Print the installed version.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Automation examples:")
	fmt.Fprintf(w, "  %s scan --mode agent --json\n", cmd)
	fmt.Fprintf(w, "  %s clean --mode agent --yes --json\n", cmd)
	fmt.Fprintf(w, "  %s analyze --assistant openclaw\n", cmd)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Safety rules:")
	fmt.Fprintln(w, "  - `safe` items can be removed directly.")
	fmt.Fprintln(w, "  - `confirm` items require explicit intent.")
	fmt.Fprintln(w, "  - `manual` items are shown for review only and are never auto-deleted.")
}

func printVersion(w io.Writer) {
	fmt.Fprintf(w, "%s %s\n", commandName(), Version)
}

func isInteractiveSession() bool {
	return isTerminal(os.Stdin) && isTerminal(os.Stdout)
}

func commandName() string {
	name := filepath.Base(os.Args[0])
	switch name {
	case "", ".", appName:
		return "oac"
	default:
		return name
	}
}

func filterIndexesBySafety(candidates []Candidate, indexes []int, safety Safety) []int {
	filtered := make([]int, 0, len(indexes))
	for _, idx := range indexes {
		if candidates[idx].Safety == safety {
			filtered = append(filtered, idx)
		}
	}
	return filtered
}
