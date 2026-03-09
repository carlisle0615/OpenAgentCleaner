package cleaner

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

type bucketStats struct {
	Count int
	Bytes int64
}

type assistantGroup struct {
	Name       string
	Candidates []Candidate
}

func printHumanReport(w io.Writer, report Report, isClean bool) {
	ui := newHumanUI(w)
	cmd := commandName()
	safeStats, confirmStats, manualStats := summarizeBySafety(report.Candidates)
	groups := groupCandidatesByAssistant(report.Candidates)

	title := "Scan complete"
	subtitle := "Nothing has been deleted."
	if isClean && report.DryRun {
		title = "Cleanup preview"
		subtitle = "This is a dry run. Review the plan before deleting anything."
	} else if isClean && report.Summary.SelectedCount == 0 && report.Summary.DeletedCount == 0 {
		title = "Nothing removed"
		subtitle = "No files were selected, or the cleanup was cancelled."
	} else if isClean {
		title = "Cleanup finished"
		subtitle = "Review the results below."
	}

	ui.banner(title, subtitle)

	if len(report.Candidates) == 0 {
		fmt.Fprintf(w, "%s No leftover files were found for %s.\n\n", ui.badgeOK("Clean"), strings.Join(report.Assistants, ", "))
		fmt.Fprintf(w, "Tip: run `%s scan --mode agent --json` if you want machine-readable output.\n", cmd)
		return
	}

	printOverview(w, ui, report, safeStats, confirmStats, manualStats, isClean)

	for _, group := range groups {
		printAssistantGroup(w, ui, group, report)
	}

	printNextSteps(w, ui, report, safeStats, confirmStats, manualStats, cmd, isClean)
}

func printOverview(w io.Writer, ui humanUI, report Report, safeStats, confirmStats, manualStats bucketStats, isClean bool) {
	fmt.Fprintf(w, "%s %d item(s) found across %d assistant(s)\n", ui.badgeInfo("Summary"), report.Summary.CandidateCount, countAssistantsWithCandidates(report.Candidates))
	fmt.Fprintf(w, "  Safe now:        %3d item(s)  %10s\n", safeStats.Count, formatBytes(safeStats.Bytes))
	fmt.Fprintf(w, "  Needs review:    %3d item(s)  %10s\n", confirmStats.Count, formatBytes(confirmStats.Bytes))
	fmt.Fprintf(w, "  Manual only:     %3d item(s)  %10s\n", manualStats.Count, formatBytes(manualStats.Bytes))
	fmt.Fprintln(w)
	if isClean {
		bytes := report.Summary.BytesDeleted
		label := "Reclaimed"
		if report.DryRun {
			label = "Planned reclaim"
			bytes = plannedBytes(report)
		}
		fmt.Fprintf(w, "%s %d selected, %d removed, %s %s\n\n", ui.badgeInfo("Result"), report.Summary.SelectedCount, report.Summary.DeletedCount, label, formatBytes(bytes))
	}
}

func printAssistantGroup(w io.Writer, ui humanUI, group assistantGroup, report Report) {
	sort.Slice(group.Candidates, func(i, j int) bool {
		if group.Candidates[i].Safety != group.Candidates[j].Safety {
			return group.Candidates[i].Safety < group.Candidates[j].Safety
		}
		return group.Candidates[i].Path < group.Candidates[j].Path
	})

	count := len(group.Candidates)
	var bytes int64
	for _, candidate := range group.Candidates {
		bytes += candidate.SizeBytes
	}

	ui.section(displayAssistant(group.Name), fmt.Sprintf("%d item(s), %s", count, formatBytes(bytes)))
	for _, candidate := range group.Candidates {
		printCandidate(w, ui, candidate, report)
	}
	fmt.Fprintln(w)
}

func printCandidate(w io.Writer, ui humanUI, candidate Candidate, report Report) {
	status := statusLabel(candidate, report)
	fmt.Fprintf(w, "  %s %-18s %-12s %8s\n", ui.statusBadge(status), displayKind(candidate.Kind), displaySafety(candidate.Safety), formatBytes(candidate.SizeBytes))
	fmt.Fprintf(w, "     %s\n", candidate.Reason)
	fmt.Fprintf(w, "     %s %s\n", ui.muted("Path:"), candidate.Path)
	for _, note := range candidate.Notes {
		fmt.Fprintf(w, "     %s %s\n", ui.muted("Note:"), note)
	}
	if candidate.Error != "" {
		fmt.Fprintf(w, "     %s %s\n", ui.badgeError("Error"), candidate.Error)
	}
}

func printNextSteps(w io.Writer, ui humanUI, report Report, safeStats, confirmStats, manualStats bucketStats, cmd string, isClean bool) {
	ui.section("What to do next", "")
	switch {
	case !isClean && safeStats.Count > 0:
		fmt.Fprintf(w, "  1. Run `%s clean --safety safe --dry-run` to preview the recommended cleanup.\n", cmd)
		fmt.Fprintf(w, "  2. Run `%s clean --safety safe` to remove the recommended leftovers.\n", cmd)
		if confirmStats.Count > 0 {
			fmt.Fprintf(w, "  3. Run `%s clean --kind models --assistants ollama --dry-run` or target a specific candidate ID if you want to remove review items such as models.\n", cmd)
		}
	case !isClean && confirmStats.Count > 0:
		fmt.Fprintf(w, "  1. Review the items above carefully.\n")
		fmt.Fprintf(w, "  2. Run `%s clean --kind <kind> --dry-run` or `%s clean --id <candidate-id> --dry-run` to preview explicit review-item cleanup.\n", cmd, cmd)
	case isClean && report.Summary.SelectedCount == 0 && report.Summary.DeletedCount == 0:
		fmt.Fprintf(w, "  1. No changes were applied.\n")
		fmt.Fprintf(w, "  2. Run `%s clean --safety safe --dry-run` if you want a preview first.\n", cmd)
		if confirmStats.Count > 0 {
			fmt.Fprintf(w, "  3. Use `%s clean --kind <kind>` or `%s clean --id <candidate-id>` when you intend to remove models, sessions, or settings.\n", cmd, cmd)
		}
	case isClean && report.DryRun:
		fmt.Fprintf(w, "  1. If the preview looks right, rerun the same command with `--yes` and without `--dry-run`.\n")
		fmt.Fprintf(w, "  2. Use the guided home screen with `%s` if you want a step-by-step flow.\n", cmd)
	default:
		fmt.Fprintf(w, "  1. Manual-only items still need your review.\n")
		if manualStats.Count > 0 {
			fmt.Fprintf(w, "  2. Keep the listed workspace or key files unless you intentionally want a full wipe.\n")
		}
	}
}

func summarizeBySafety(candidates []Candidate) (bucketStats, bucketStats, bucketStats) {
	var safeStats bucketStats
	var confirmStats bucketStats
	var manualStats bucketStats

	for _, candidate := range candidates {
		switch candidate.Safety {
		case SafetySafe:
			safeStats.Count++
			safeStats.Bytes += candidate.SizeBytes
		case SafetyConfirm:
			confirmStats.Count++
			confirmStats.Bytes += candidate.SizeBytes
		case SafetyManual:
			manualStats.Count++
			manualStats.Bytes += candidate.SizeBytes
		}
	}

	return safeStats, confirmStats, manualStats
}

func groupCandidatesByAssistant(candidates []Candidate) []assistantGroup {
	grouped := map[string][]Candidate{}
	order := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if _, ok := grouped[candidate.Assistant]; !ok {
			order = append(order, candidate.Assistant)
		}
		grouped[candidate.Assistant] = append(grouped[candidate.Assistant], candidate)
	}
	sort.Strings(order)

	out := make([]assistantGroup, 0, len(order))
	for _, assistant := range order {
		out = append(out, assistantGroup{
			Name:       assistant,
			Candidates: grouped[assistant],
		})
	}
	return out
}

func countAssistantsWithCandidates(candidates []Candidate) int {
	seen := map[string]struct{}{}
	for _, candidate := range candidates {
		seen[candidate.Assistant] = struct{}{}
	}
	return len(seen)
}

func statusLabel(candidate Candidate, report Report) string {
	switch {
	case candidate.Error != "":
		return "error"
	case candidate.Deleted:
		return "removed"
	case candidate.Selected && report.DryRun:
		return "planned"
	case candidate.Selected && candidate.Skipped:
		return "skipped"
	case candidate.Selected:
		return "selected"
	default:
		return "found"
	}
}

func plannedBytes(report Report) int64 {
	var bytes int64
	for _, candidate := range report.Candidates {
		if candidate.Selected {
			bytes += candidate.SizeBytes
		}
	}
	return bytes
}

func displayAssistant(name string) string {
	switch name {
	case "openclaw":
		return "OpenClaw"
	case "ironclaw":
		return "IronClaw"
	case "ollama":
		return "Ollama"
	default:
		return titleize(name)
	}
}

func displaySafety(safety Safety) string {
	switch safety {
	case SafetySafe:
		return "safe"
	case SafetyConfirm:
		return "review"
	case SafetyManual:
		return "manual"
	default:
		return string(safety)
	}
}

func displayKind(kind string) string {
	switch kind {
	case "gateway_logs", "logs", "launchd_logs":
		return "Logs"
	case "session_store":
		return "Sessions"
	case "config":
		return "Config"
	case "env_file":
		return "Env file"
	case "extensions":
		return "Extensions"
	case "workspace":
		return "Workspace"
	case "service_plist", "legacy_service_plist":
		return "Background service"
	case "state_root":
		return "State folder"
	case "database":
		return "Database"
	case "oauth_session":
		return "Signed-in session"
	case "mcp_config":
		return "MCP servers"
	case "legacy_settings":
		return "Legacy settings"
	case "legacy_bootstrap":
		return "Legacy bootstrap"
	case "channels":
		return "Channels"
	case "tools":
		return "Tools"
	case "repl_history":
		return "History"
	case "pairing_store":
		return "Pairing data"
	case "allow_from":
		return "Allow list"
	case "pairing_attempts":
		return "Pairing attempts"
	case "saved_state":
		return "Saved window state"
	case "cache":
		return "Cache"
	case "webkit_cache":
		return "Web cache"
	case "app_support":
		return "App support"
	case "models":
		return "Models"
	case "server_config":
		return "Server config"
	case "app_bundle":
		return "Mac app"
	case "cli_symlink":
		return "CLI link"
	case "auth_key":
		return "Auth key"
	default:
		return titleize(kind)
	}
}

func titleize(raw string) string {
	parts := strings.Fields(strings.NewReplacer("-", " ", "_", " ").Replace(raw))
	for i := range parts {
		if len(parts[i]) == 0 {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:])
	}
	return strings.Join(parts, " ")
}

func selectedHasConfirm(report Report) bool {
	for _, candidate := range report.Candidates {
		if candidate.Selected && candidate.Safety == SafetyConfirm {
			return true
		}
	}
	return false
}
