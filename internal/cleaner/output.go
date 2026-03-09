package cleaner

import (
	"fmt"
	"io"
)

func printHumanReport(w io.Writer, report Report, isClean bool) {
	if isClean {
		fmt.Fprintf(w, "Operation: %s", report.Operation)
		if report.DryRun {
			fmt.Fprint(w, " (dry-run)")
		}
		fmt.Fprintln(w)
	} else {
		fmt.Fprintln(w, "Operation: scan")
	}
	fmt.Fprintf(w, "Mode: %s\n", report.Mode)
	fmt.Fprintf(w, "Assistants: %v\n\n", report.Assistants)

	current := ""
	for _, candidate := range report.Candidates {
		if candidate.Assistant != current {
			current = candidate.Assistant
			fmt.Fprintf(w, "%s\n", current)
		}

		status := "found"
		switch {
		case candidate.Deleted:
			status = "deleted"
		case candidate.Selected && report.DryRun:
			status = "planned"
		case candidate.Selected && candidate.Skipped:
			status = "skipped"
		case candidate.Error != "":
			status = "error"
		}

		fmt.Fprintf(w, "  - %-8s %-14s %-7s %8s  %s\n", status, candidate.Kind, candidate.Safety, formatBytes(candidate.SizeBytes), candidate.Path)
		fmt.Fprintf(w, "    %s\n", candidate.Reason)
		for _, note := range candidate.Notes {
			fmt.Fprintf(w, "    note: %s\n", note)
		}
		if candidate.Error != "" {
			fmt.Fprintf(w, "    error: %s\n", candidate.Error)
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintf(w, "Summary: %d candidates, %s total\n", report.Summary.CandidateCount, formatBytes(report.Summary.BytesFound))
	if isClean {
		fmt.Fprintf(w, "Selected: %d, deleted: %d, reclaimed: %s\n", report.Summary.SelectedCount, report.Summary.DeletedCount, formatBytes(report.Summary.BytesDeleted))
		fmt.Fprintln(w, "Manual items are listed for review only and are never auto-deleted.")
	}
}
