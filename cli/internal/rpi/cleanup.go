package rpi

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// StaleRunEntry describes a stale RPI run discovered during cleanup scanning.
type StaleRunEntry struct {
	RunID        string
	Root         string
	StatePath    string
	Reason       string
	WorktreePath string
	Terminal     string
}

// CheckTerminalRunStale returns a StaleRunEntry for a terminal run that qualifies
// for cleanup (non-completed terminal status with an existing worktree).
// minAge filters runs younger than the threshold; now is the reference time.
func CheckTerminalRunStale(
	runID, root, statePath string,
	terminalStatus, terminalReason, terminatedAt, startedAt, worktreePath string,
	minAge time.Duration, now time.Time,
) (StaleRunEntry, bool) {
	if terminalStatus == "completed" {
		return StaleRunEntry{}, false
	}
	if worktreePath == "" {
		return StaleRunEntry{}, false
	}
	if _, statErr := os.Stat(worktreePath); statErr != nil {
		return StaleRunEntry{}, false
	}
	if minAge > 0 {
		candidateAt := cmp.Or(terminatedAt, startedAt)
		parsedAt, parseErr := time.Parse(time.RFC3339, candidateAt)
		if parseErr != nil || now.Sub(parsedAt) < minAge {
			return StaleRunEntry{}, false
		}
	}
	reason := cmp.Or(terminalReason, "terminal status: "+terminalStatus)
	return StaleRunEntry{
		RunID:        runID,
		Root:         root,
		StatePath:    statePath,
		Reason:       reason,
		WorktreePath: worktreePath,
		Terminal:     terminalStatus,
	}, true
}

// UpdateFlatStateIfMatches updates the flat (root-level) state file with stale
// metadata when its run_id matches the given runID.
func UpdateFlatStateIfMatches(flatPath, runID, reason, terminatedAt string) {
	flatData, fErr := os.ReadFile(flatPath)
	if fErr != nil {
		return
	}
	var flatRaw map[string]any
	if json.Unmarshal(flatData, &flatRaw) != nil {
		return
	}
	if flatRunID, ok := flatRaw["run_id"].(string); !ok || flatRunID != runID {
		return
	}
	flatRaw["terminal_status"] = "stale"
	flatRaw["terminal_reason"] = reason
	flatRaw["terminated_at"] = terminatedAt
	if flatUpdated, mErr := json.MarshalIndent(flatRaw, "", "  "); mErr == nil {
		flatUpdated = append(flatUpdated, '\n')
		_ = WritePhasedStateAtomic(flatPath, flatUpdated)
	}
}

// MarkRunStaleInState writes terminal stale metadata to a run's state file
// and updates the flat state file if it references the same run.
func MarkRunStaleInState(sr StaleRunEntry, rootAgentsDir string) error {
	data, err := os.ReadFile(sr.StatePath)
	if err != nil {
		return fmt.Errorf("read state: %w", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("unmarshal state: %w", err)
	}

	raw["terminal_status"] = "stale"
	raw["terminal_reason"] = sr.Reason
	raw["terminated_at"] = time.Now().Format(time.RFC3339)

	updated, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	updated = append(updated, '\n')

	if err := WritePhasedStateAtomic(sr.StatePath, updated); err != nil {
		return fmt.Errorf("write state: %w", err)
	}

	flatPath := filepath.Join(rootAgentsDir, ".agents", "rpi", PhasedStateFile)
	UpdateFlatStateIfMatches(flatPath, sr.RunID, sr.Reason, raw["terminated_at"].(string))

	return nil
}

// PatchStateWithCancelFields writes interrupted terminal metadata to a state file.
func PatchStateWithCancelFields(path, reason, now string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	raw["terminal_status"] = "interrupted"
	raw["terminal_reason"] = reason
	raw["terminated_at"] = now
	updated, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	updated = append(updated, '\n')
	return WritePhasedStateAtomic(path, updated)
}

// ResolveCleanupRepoRoot picks a controller worktree root to execute
// `git worktree remove` against. It prefers a sibling worktree in the same
// parent directory as targetWorktree, avoiding attempts to remove a worktree
// from within itself. The roots parameter is the set of search roots (from
// collectSearchRoots in cmd/ao).
func ResolveCleanupRepoRoot(cwd, targetWorktree string, roots []string) string {
	target := filepath.Clean(targetWorktree)
	targetParent := filepath.Dir(target)

	for _, root := range roots {
		cleanRoot := filepath.Clean(root)
		if cleanRoot == target {
			continue
		}
		if filepath.Dir(cleanRoot) == targetParent {
			return cleanRoot
		}
	}
	return cwd
}

// ValidateWorktreeSibling checks that worktreePath is a sibling of repoRoot
// (same parent directory) and is not the repo root itself. This prevents
// corrupted state files from directing os.RemoveAll at unrelated paths.
func ValidateWorktreeSibling(repoRoot, worktreePath string) error {
	repoParent := filepath.Dir(filepath.Clean(repoRoot))
	wtParent := filepath.Dir(filepath.Clean(worktreePath))
	if wtParent != repoParent {
		return fmt.Errorf("worktree path %q is not a sibling of repo %q; refusing removal", worktreePath, repoRoot)
	}
	cleanWT := filepath.Clean(worktreePath)
	if cleanWT == filepath.Clean(repoRoot) {
		return fmt.Errorf("worktree path %q is the repo root; refusing removal", worktreePath)
	}
	return nil
}
