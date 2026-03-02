package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// --- Plan-file fallback helpers ---

const planFileEpicPrefix = "plan:"

// isPlanFileEpic returns true when the epic ID is a plan-file sentinel.
func isPlanFileEpic(epicID string) bool {
	return strings.HasPrefix(epicID, planFileEpicPrefix)
}

// planFileFromEpic extracts the plan file path from a plan-file epic sentinel.
func planFileFromEpic(epicID string) string {
	return strings.TrimPrefix(epicID, planFileEpicPrefix)
}

// discoverPlanFile scans .agents/plans/ for the most recently modified .md file
// and returns its path relative to cwd.
func discoverPlanFile(cwd string) (string, error) {
	plansDir := filepath.Join(cwd, ".agents", "plans")
	entries, err := os.ReadDir(plansDir)
	if err != nil {
		return "", fmt.Errorf("read plans directory: %w", err)
	}

	var latestPath string
	var latestMod time.Time
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if latestPath == "" || info.ModTime().After(latestMod) {
			latestPath = filepath.Join(".agents", "plans", entry.Name())
			latestMod = info.ModTime()
		}
	}

	if latestPath == "" {
		return "", fmt.Errorf("no .md files found in %s", plansDir)
	}

	return latestPath, nil
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
	var entries []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(data, &entries); err != nil {
		return "", fmt.Errorf("parse bd list JSON: %w", err)
	}
	for i := len(entries) - 1; i >= 0; i-- {
		epicID := strings.TrimSpace(entries[i].ID)
		if epicID != "" {
			return epicID, nil
		}
	}
	return "", fmt.Errorf("no epic found in bd list output")
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
	return parseIssueTypeFromShowJSON(out)
}

func parseIssueTypeFromShowJSON(data []byte) (bool, error) {
	var single map[string]any
	if err := json.Unmarshal(data, &single); err == nil {
		if isEpic, ok := issueTypeFromMap(single); ok {
			return isEpic, nil
		}
		return false, fmt.Errorf("bd show output missing issue type")
	}

	var multiple []map[string]any
	if err := json.Unmarshal(data, &multiple); err == nil {
		for _, entry := range multiple {
			if isEpic, ok := issueTypeFromMap(entry); ok {
				return isEpic, nil
			}
		}
		return false, fmt.Errorf("bd show array output missing issue type")
	}

	return false, fmt.Errorf("bd show output is not valid json")
}

func issueTypeFromMap(payload map[string]any) (bool, bool) {
	if payload == nil {
		return false, false
	}
	if rawEpic, ok := payload["epic"]; ok {
		if isEpic, ok := rawEpic.(bool); ok {
			return isEpic, true
		}
	}
	for _, key := range []string{"type", "issue_type", "kind"} {
		raw, ok := payload[key]
		if !ok {
			continue
		}
		if kind, ok := raw.(string); ok {
			return strings.EqualFold(strings.TrimSpace(kind), "epic"), true
		}
	}
	if nested, ok := payload["issue"].(map[string]any); ok {
		return issueTypeFromMap(nested)
	}
	return false, false
}

func parseLatestEpicIDFromText(output string) (string, error) {
	// Allow custom prefixes (bd-*, ag-*, etc.) and keep the match anchored
	// to issue-like tokens near the start of each line.
	idPattern := regexp.MustCompile(`^[a-z][a-z0-9]*-[a-z0-9][a-z0-9.]*$`)

	var latest string
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		limit := len(fields)
		if limit > 3 {
			limit = 3
		}
		for i := range limit {
			field := fields[i]
			token := strings.Trim(field, "[]()")
			if idPattern.MatchString(token) {
				latest = token
				break
			}
		}
	}
	if latest == "" {
		return "", fmt.Errorf("no epic found in bd list output")
	}
	return latest, nil
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
func parseFastPath(output string) bool {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	issueCount := 0
	blockedCount := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		issueCount++
		if strings.Contains(strings.ToLower(line), "blocked") {
			blockedCount++
		}
	}
	return issueCount <= 2 && blockedCount == 0
}

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
func parseCrankCompletion(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	total := 0
	closed := 0
	blocked := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		total++
		lower := strings.ToLower(line)
		if strings.Contains(lower, "closed") || strings.Contains(lower, "\u2713") {
			closed++
		}
		if strings.Contains(lower, "blocked") {
			blocked++
		}
	}

	if total == 0 {
		return "DONE"
	}
	if closed == total {
		return "DONE"
	}
	if blocked > 0 {
		return "BLOCKED"
	}
	return "PARTIAL"
}
