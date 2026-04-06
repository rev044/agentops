package rpi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// PlanFileEpicPrefix is the sentinel prefix for plan-file-based epics.
const PlanFileEpicPrefix = "plan:"

// IsPlanFileEpic returns true when the epic ID is a plan-file sentinel.
func IsPlanFileEpic(epicID string) bool {
	return strings.HasPrefix(epicID, PlanFileEpicPrefix)
}

// PlanFileFromEpic extracts the plan file path from a plan-file epic sentinel.
func PlanFileFromEpic(epicID string) string {
	return strings.TrimPrefix(epicID, PlanFileEpicPrefix)
}

// DiscoverPlanFile scans .agents/plans/ for the most recently modified .md file
// and returns its path relative to cwd.
func DiscoverPlanFile(cwd string) (string, error) {
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

// ParseLatestEpicIDFromJSON extracts the last non-empty ID from bd list --json output.
func ParseLatestEpicIDFromJSON(data []byte) (string, error) {
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

// ParseIssueTypeFromShowJSON determines if a bd show --json response is an epic.
func ParseIssueTypeFromShowJSON(data []byte) (bool, error) {
	var single map[string]any
	if err := json.Unmarshal(data, &single); err == nil {
		if isEpic, ok := IssueTypeFromMap(single); ok {
			return isEpic, nil
		}
		return false, fmt.Errorf("bd show output missing issue type")
	}

	var multiple []map[string]any
	if err := json.Unmarshal(data, &multiple); err == nil {
		for _, entry := range multiple {
			if isEpic, ok := IssueTypeFromMap(entry); ok {
				return isEpic, nil
			}
		}
		return false, fmt.Errorf("bd show array output missing issue type")
	}

	return false, fmt.Errorf("bd show output is not valid json")
}

// IssueTypeFromMap checks a JSON map for epic/type indicators.
func IssueTypeFromMap(payload map[string]any) (bool, bool) {
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
		return IssueTypeFromMap(nested)
	}
	return false, false
}

// ParseLatestEpicIDFromText extracts the last issue-like ID from bd list text output.
func ParseLatestEpicIDFromText(output string) (string, error) {
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

// ParseFastPath determines if bd children output indicates a micro-epic.
func ParseFastPath(output string) bool {
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

// ParseCrankCompletion determines completion status from bd children output.
// Returns "DONE", "BLOCKED", or "PARTIAL".
func ParseCrankCompletion(output string) string {
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
