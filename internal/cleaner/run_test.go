package cleaner

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDispatch(t *testing.T) {
	home := setTestHome(t)
	writeTestFile(t, filepath.Join(home, ".ollama", "logs", "app.log"), "log")

	withFakeArgs(t, []string{"/tmp/oac"}, func() {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		if code := Run(nil, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "Start here:") {
			t.Fatalf("Run(nil) = %d, stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}

		stdout.Reset()
		stderr.Reset()
		if code := Run([]string{"version"}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "oac") {
			t.Fatalf("Run(version) = %d, stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}

		stdout.Reset()
		stderr.Reset()
		if code := Run([]string{"--help"}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "Core commands:") {
			t.Fatalf("Run(help) = %d, stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		for _, want := range []string{
			"Agent-first workflow:",
			"scan --mode agent --json",
			"clean --safety safe --dry-run --mode agent --json",
			"clean --id <candidate-id> --yes --mode agent --json",
			"analyze",
			"Human workflow:",
		} {
			if !strings.Contains(stdout.String(), want) {
				t.Fatalf("Run(help) missing %q in stdout=%q", want, stdout.String())
			}
		}

		stdout.Reset()
		stderr.Reset()
		if code := Run([]string{"scan", "--mode", "agent", "--json"}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), `"operation": "scan"`) {
			t.Fatalf("Run(scan) = %d, stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}

		stdout.Reset()
		stderr.Reset()
		if code := Run([]string{"clean", "--mode", "agent", "--safety", "safe", "--dry-run", "--json"}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), `"operation": "clean"`) {
			t.Fatalf("Run(clean dry-run) = %d, stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}

		stdout.Reset()
		stderr.Reset()
		withFakeStdin(t, "x\n", func() {
			if code := Run([]string{"analyze"}, &stdout, &stderr); code != 2 || !strings.Contains(stderr.String(), "interactive human mode only") {
				t.Fatalf("Run(analyze) = %d, stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
		})

		stdout.Reset()
		stderr.Reset()
		if code := Run([]string{"unknown"}, &stdout, &stderr); code != 2 || !strings.Contains(stderr.String(), "unknown command") {
			t.Fatalf("Run(unknown) = %d, stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
}

func TestRunHomeMenu(t *testing.T) {
	home := setTestHome(t)
	writeTestFile(t, filepath.Join(home, ".ollama", "logs", "app.log"), "log")

	withFakeStdin(t, "6\n7\n", func() {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		if code := runHomeMenu(&stdout, &stderr); code != 0 {
			t.Fatalf("runHomeMenu(help->quit) = %d", code)
		}
		if !strings.Contains(stdout.String(), "Show command help") || !strings.Contains(stdout.String(), "Agent-first workflow:") {
			t.Fatalf("runHomeMenu output = %q", stdout.String())
		}
	})

	withFakeStdin(t, "oops\n7\n", func() {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		if code := runHomeMenu(&stdout, &stderr); code != 0 {
			t.Fatalf("runHomeMenu(invalid->quit) = %d", code)
		}
		if !strings.Contains(stderr.String(), "unknown action") {
			t.Fatalf("stderr = %q", stderr.String())
		}
	})

	withFakeStdin(t, "1\n", func() {
		var stdout bytes.Buffer
		if code := runHomeMenu(&stdout, &bytes.Buffer{}); code != 0 || !strings.Contains(stdout.String(), `"operation": "scan"`) {
			t.Fatalf("runHomeMenu(scan) = %d, stdout=%q", code, stdout.String())
		}
	})

	withFakeStdin(t, "3\n", func() {
		var stdout bytes.Buffer
		if code := runHomeMenu(&stdout, &bytes.Buffer{}); code != 0 || !strings.Contains(stdout.String(), `"operation": "clean"`) {
			t.Fatalf("runHomeMenu(preview) = %d, stdout=%q", code, stdout.String())
		}
	})
}

func TestRunScanAndRunCleanValidation(t *testing.T) {
	home := setTestHome(t)
	writeTestFile(t, filepath.Join(home, ".ollama", "logs", "app.log"), "log")

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if code := runScan([]string{"--assistants", "bad"}, &stdout, &stderr); code != 2 || !strings.Contains(stderr.String(), "unsupported assistant") {
		t.Fatalf("runScan invalid = %d stderr=%q", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := runClean([]string{"--assistants", "bad"}, &stdout, &stderr); code != 2 || !strings.Contains(stderr.String(), "unsupported assistant") {
		t.Fatalf("runClean invalid = %d stderr=%q", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := runClean([]string{"--mode", "agent"}, &stdout, &stderr); code != 1 || !strings.Contains(stderr.String(), "explicit selector") {
		t.Fatalf("runClean noninteractive = %d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}

	devNull, err := os.Open("/dev/null")
	if err != nil {
		t.Fatalf("open /dev/null: %v", err)
	}
	defer devNull.Close()

	withStdoutFile(t, devNull, func() {
		stdout.Reset()
		stderr.Reset()
		if code := runScan([]string{}, &stdout, &stderr); code != 0 {
			t.Fatalf("runScan human auto = %d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})

	stdout.Reset()
	stderr.Reset()
	if code := runScan([]string{"--assistants", "ollama", "--verbose"}, &stdout, &stderr); code != 0 {
		t.Fatalf("runScan verbose = %d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "[verbose] scanning leftover candidates for Ollama") {
		t.Fatalf("runScan verbose stderr = %q", stderr.String())
	}
}

func TestScanAndCleanReport(t *testing.T) {
	home := setTestHome(t)
	safePath := filepath.Join(home, ".ollama", "logs", "app.log")
	confirmPath := filepath.Join(home, ".ollama", "server.json")
	writeTestFile(t, safePath, "safe")
	writeTestFile(t, confirmPath, "confirm")

	report, err := scanReport(options{
		Assistants: []string{"ollama"},
		Mode:       "agent",
		JSON:       true,
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("scanReport() err = %v", err)
	}
	if report.Operation != "scan" || report.Mode != "agent" || report.Summary.CandidateCount == 0 {
		t.Fatalf("scanReport() = %#v", report)
	}

	report, err = cleanReport(options{
		Assistants: []string{"ollama"},
		Mode:       "agent",
		Safeties:   []Safety{SafetySafe},
		DryRun:     true,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("cleanReport(dry-run) err = %v", err)
	}
	if report.Summary.SelectedCount == 0 || report.Summary.DeletedCount != 0 {
		t.Fatalf("cleanReport(dry-run) = %#v", report.Summary)
	}

	report, err = cleanReport(options{
		Assistants: []string{"ollama"},
		Mode:       "agent",
		Safeties:   []Safety{SafetySafe},
		Yes:        true,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("cleanReport(delete safe) err = %v", err)
	}
	var deletedSafe bool
	for _, candidate := range report.Candidates {
		if candidate.Kind == "logs" && candidate.Deleted {
			deletedSafe = true
		}
	}
	if !deletedSafe {
		t.Fatalf("cleanReport(delete) candidates = %#v", report.Candidates)
	}
	if pathExists(safePath) {
		t.Fatal("safePath should be deleted")
	}
	if !pathExists(confirmPath) {
		t.Fatal("confirmPath should remain because it was not explicitly targeted")
	}

	report, err = cleanReport(options{
		Assistants: []string{"ollama"},
		Mode:       "agent",
		Kinds:      []string{"config"},
		Yes:        true,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("cleanReport(delete confirm) err = %v", err)
	}
	if pathExists(confirmPath) {
		t.Fatal("confirmPath should be deleted")
	}

	report, err = cleanReport(options{
		Assistants: []string{"ironclaw"},
		Mode:       "agent",
		DryRun:     true,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("cleanReport(empty) err = %v", err)
	}
	if report.Summary.CandidateCount != 0 {
		t.Fatalf("cleanReport(empty) = %#v", report.Summary)
	}

}

func TestSelectionConfirmationAndDeleteHelpers(t *testing.T) {
	candidates := []Candidate{
		{ID: candidateID("ollama", "logs", "/tmp/a"), Assistant: "ollama", Path: "/tmp/a", Safety: SafetySafe, SizeBytes: 10, Kind: "logs", Deletable: true},
		{ID: candidateID("ollama", "config", "/tmp/b"), Assistant: "ollama", Path: "/tmp/b", Safety: SafetyConfirm, SizeBytes: 20, Kind: "config", Deletable: true, RequiresConfirmation: true},
		{ID: candidateID("openclaw", "workspace", "/tmp/c"), Assistant: "openclaw", Path: "/tmp/c", Safety: SafetyManual, SizeBytes: 30, Kind: "workspace"},
	}
	eligible, explicit, err := selectCleanCandidates(candidates, options{Safeties: []Safety{SafetySafe}})
	if err != nil || !explicit || len(eligible) != 1 || eligible[0] != 0 {
		t.Fatalf("selectCleanCandidates(safe) = %v, %v, %v", eligible, explicit, err)
	}
	eligible, explicit, err = selectCleanCandidates(candidates, options{})
	if err != nil || explicit || len(eligible) != 2 {
		t.Fatalf("selectCleanCandidates(default human) = %v, %v, %v", eligible, explicit, err)
	}
	eligible, explicit, err = selectCleanCandidates(candidates, options{CandidateIDs: []string{candidates[1].ID}})
	if err != nil || !explicit || len(eligible) != 1 || eligible[0] != 1 {
		t.Fatalf("selectCleanCandidates(id) = %v, %v, %v", eligible, explicit, err)
	}
	if _, _, err := selectCleanCandidates(candidates, options{CandidateIDs: []string{"missing"}}); err == nil {
		t.Fatal("selectCleanCandidates(missing id) should fail")
	}

	withFakeStdin(t, "\n", func() {
		got, err := promptSelection(&bytes.Buffer{}, &bytes.Buffer{}, candidates, []int{0, 1}, false)
		if err != nil || len(got) != 1 || got[0] != 0 {
			t.Fatalf("promptSelection(default safe) = %v, %v", got, err)
		}
	})
	withFakeStdin(t, "safe\n", func() {
		got, err := promptSelection(&bytes.Buffer{}, &bytes.Buffer{}, candidates, []int{0, 1}, false)
		if err != nil || len(got) != 1 || got[0] != 0 {
			t.Fatalf("promptSelection(safe) = %v, %v", got, err)
		}
	})
	withFakeStdin(t, "none\n", func() {
		got, err := promptSelection(&bytes.Buffer{}, &bytes.Buffer{}, candidates, []int{0, 1}, false)
		if err != nil || got != nil {
			t.Fatalf("promptSelection(none) = %v, %v", got, err)
		}
	})
	withFakeStdin(t, "all\n", func() {
		got, err := promptSelection(&bytes.Buffer{}, &bytes.Buffer{}, candidates, []int{0, 1}, true)
		if err != nil || len(got) != 2 {
			t.Fatalf("promptSelection(all) = %v, %v", got, err)
		}
	})
	withFakeStdin(t, "2,1\n", func() {
		got, err := promptSelection(&bytes.Buffer{}, &bytes.Buffer{}, candidates, []int{0, 1}, true)
		if err != nil || len(got) != 2 || got[0] != 1 || got[1] != 0 {
			t.Fatalf("promptSelection(explicit) = %v, %v", got, err)
		}
	})
	withFakeStdin(t, "9\n", func() {
		if _, err := promptSelection(&bytes.Buffer{}, &bytes.Buffer{}, candidates, []int{0, 1}, true); err == nil {
			t.Fatal("promptSelection(invalid) should fail")
		}
	})

	withFakeStdin(t, "yes\n", func() {
		if !confirmDeletion(&bytes.Buffer{}, &bytes.Buffer{}, candidates, map[int]struct{}{0: {}}) {
			t.Fatal("confirmDeletion(safe yes) should succeed")
		}
	})
	withFakeStdin(t, "delete\n", func() {
		if !confirmDeletion(&bytes.Buffer{}, &bytes.Buffer{}, candidates, map[int]struct{}{1: {}}) {
			t.Fatal("confirmDeletion(confirm delete) should succeed")
		}
	})
	withFakeStdin(t, "no\n", func() {
		if confirmDeletion(&bytes.Buffer{}, &bytes.Buffer{}, candidates, map[int]struct{}{0: {}}) {
			t.Fatal("confirmDeletion(no) should fail")
		}
	})

	home := setTestHome(t)
	file := filepath.Join(home, "delete-me.txt")
	writeTestFile(t, file, "bye")

	if err := deletePath("relative"); err == nil {
		t.Fatal("deletePath(relative) should fail")
	}
	if err := deletePath("/"); err == nil {
		t.Fatal("deletePath(/) should fail")
	}
	if err := deletePath(home); err == nil {
		t.Fatal("deletePath(home) should fail")
	}
	if err := deletePath(filepath.Join(home, "missing")); err != nil {
		t.Fatalf("deletePath(missing) err = %v", err)
	}
	if err := deletePath(file); err != nil {
		t.Fatalf("deletePath(file) err = %v", err)
	}
	if pathExists(file) {
		t.Fatal("deletePath(file) should remove file")
	}

	summary := summarize([]Candidate{
		{SizeBytes: 10, Selected: true, Deleted: true},
		{SizeBytes: 20, Selected: true},
		{SizeBytes: 30},
	})
	if summary.CandidateCount != 3 || summary.SelectedCount != 2 || summary.DeletedCount != 1 || summary.BytesFound != 60 || summary.BytesDeleted != 10 {
		t.Fatalf("summarize() = %#v", summary)
	}

	withFakeArgs(t, []string{"/tmp/my-oac"}, func() {
		if commandName() != "my-oac" {
			t.Fatalf("commandName() = %q", commandName())
		}
	})
	withFakeArgs(t, []string{"OpenAgentCleaner"}, func() {
		if commandName() != "oac" {
			t.Fatalf("commandName(appName) = %q", commandName())
		}
	})

	if isInteractiveSession() {
		t.Fatal("isInteractiveSession() should be false in tests")
	}

	filtered := filterIndexesBySafety(candidates, []int{0, 1, 2}, SafetyConfirm)
	if len(filtered) != 1 || filtered[0] != 1 {
		t.Fatalf("filterIndexesBySafety() = %v", filtered)
	}
}
