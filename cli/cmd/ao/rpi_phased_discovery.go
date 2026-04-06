package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/boshu2/agentops/cli/internal/rpi"
)

// --- Plan-file fallback helpers ---

// isPlanFileEpic returns true when the epic ID is a plan-file sentinel.
func isPlanFileEpic(epicID string) bool { return rpi.IsPlanFileEpic(epicID) }

// planFileFromEpic extracts the plan file path from a plan-file epic sentinel.
func planFileFromEpic(epicID string) string { return rpi.PlanFileFromEpic(epicID) }

// discoverPlanFile scans .agents/plans/ for the most recently modified .md file.
func discoverPlanFile(cwd string) (string, error) { return rpi.DiscoverPlanFile(cwd) }

// issueTypeFromMap wraps rpi.IssueTypeFromMap.
func issueTypeFromMap(payload map[string]any) (bool, bool) { return rpi.IssueTypeFromMap(payload) }

// parseIssueTypeFromShowJSON wraps rpi.ParseIssueTypeFromShowJSON.
func parseIssueTypeFromShowJSON(data []byte) (bool, error) {
	return rpi.ParseIssueTypeFromShowJSON(data)
}

// --- Epic and completion helpers ---

// extractEpicID finds the most recently created open epic ID via bd CLI.
// bd list returns epics in creation order; we take the LAST match so that
// the epic just created by the plan phase is selected over older ones.
func extractEpicID(bdCommand string) (string, error) {
	command := effectiveBDCommand(bdCommand)

	// Prefer JSON output for prefix-agnostic parsing.
	cmd := exec.Command(command, "list", "--type", "epic", "--status", "open", "--json")
	out, err := cmd.Output()
	if err == nil {
		epicID, parseErr := parseLatestEpicIDFromJSON(out)
		if parseErr == nil {
			return epicID, nil
		}
		VerbosePrintf("Warning: could not parse bd JSON epic list (falling back to text): %v\n", parseErr)
	} else {
		VerbosePrintf("Warning: bd list --json failed (falling back to text): %v\n", err)
	}

	// Fallback for older bd builds that do not support JSON output.
	cmd = exec.Command(command, "list", "--type", "epic", "--status", "open")
	out, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("bd list: %w", err)
	}
	return parseLatestEpicIDFromText(string(out))
}

// extractAnyOpenIssueID finds the most recently created open issue, preferring epics.
// It first tries --type epic to avoid selecting a non-epic issue that would cause
// checkCrankCompletion to return false DONE (bd children returns empty for non-epics).
// Falls back to any open issue when no epic exists (e.g., small-scope work created as a task).
func extractAnyOpenIssueID(bdCommand string) (string, error) {
	command := effectiveBDCommand(bdCommand)

	// Prefer epic-type issues to avoid false DONE from empty bd children output.
	cmd := exec.Command(command, "list", "--type", "epic", "--status", "open", "--json")
	out, err := cmd.Output()
	if err == nil {
		if id, parseErr := parseLatestEpicIDFromJSON(out); parseErr == nil {
			return id, nil
		}
	}

	// Fallback: any open issue (handles small-scope tasks that aren't epics).
	cmd = exec.Command(command, "list", "--status", "open", "--json")
	out, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("bd list (any type): %w", err)
	}
	return parseLatestEpicIDFromJSON(out)
}

func parseLatestEpicIDFromJSON(data []byte) (string, error) {
	return rpi.ParseLatestEpicIDFromJSON(data)
}

func isEpicIssue(issueID, bdCommand string) (bool, error) {
	if strings.TrimSpace(issueID) == "" {
		return false, fmt.Errorf("empty issue id")
	}
	cmd := exec.Command(effectiveBDCommand(bdCommand), "show", issueID, "--json")
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("bd show: %w", err)
	}
	return rpi.ParseIssueTypeFromShowJSON(out)
}

func parseLatestEpicIDFromText(output string) (string, error) {
	return rpi.ParseLatestEpicIDFromText(output)
}

// detectFastPath checks if an epic is a micro-epic (<=2 issues, no blockers).
func detectFastPath(epicID string, bdCommand string) (bool, error) {
	cmd := exec.Command(effectiveBDCommand(bdCommand), "children", epicID)
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("bd children: %w", err)
	}
	return parseFastPath(string(out)), nil
}

// parseFastPath determines if bd children output indicates a micro-epic.
func parseFastPath(output string) bool { return rpi.ParseFastPath(output) }

// checkCrankCompletion checks epic completion via bd children statuses.
// Returns "DONE", "BLOCKED", or "PARTIAL".
func checkCrankCompletion(epicID string, bdCommand string) (string, error) {
	cmd := exec.Command(effectiveBDCommand(bdCommand), "children", epicID)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("bd children: %w", err)
	}
	return parseCrankCompletion(string(out)), nil
}

// parseCrankCompletion determines completion status from bd children output.
func parseCrankCompletion(output string) string { return rpi.ParseCrankCompletion(output) }
