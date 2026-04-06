package search

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ConstraintIndex represents the .agents/constraints/index.json schema.
type ConstraintIndex struct {
	SchemaVersion int               `json:"schema_version"`
	Constraints   []ConstraintEntry `json:"constraints"`
}

// ConstraintAppliesTo encodes scope filters for a constraint.
type ConstraintAppliesTo struct {
	Scope      string   `json:"scope,omitempty"`
	IssueTypes []string `json:"issue_types,omitempty"`
	PathGlobs  []string `json:"path_globs,omitempty"`
	Languages  []string `json:"languages,omitempty"`
}

// ConstraintDetector encodes how a constraint is detected.
type ConstraintDetector struct {
	Kind      string `json:"kind,omitempty"`
	Mode      string `json:"mode,omitempty"`
	Pattern   string `json:"pattern,omitempty"`
	Exclude   string `json:"exclude,omitempty"`
	Companion string `json:"companion,omitempty"`
	Command   string `json:"command,omitempty"`
	Message   string `json:"message,omitempty"`
}

// ConstraintEntry represents a single compiled constraint.
type ConstraintEntry struct {
	ID              string              `json:"id"`
	FindingID       string              `json:"finding_id,omitempty"`
	Title           string              `json:"title"`
	Source          string              `json:"source"`
	SourceArtifact  string              `json:"source_artifact,omitempty"`
	SourceType      string              `json:"source_type,omitempty"`
	CompilerTargets []string            `json:"compiler_targets,omitempty"`
	Detectability   string              `json:"detectability,omitempty"`
	Status          string              `json:"status"`
	CompiledAt      string              `json:"compiled_at"`
	ReviewFile      string              `json:"review_file,omitempty"`
	AppliesTo       ConstraintAppliesTo `json:"applies_to,omitempty"`
	Detector        ConstraintDetector  `json:"detector,omitempty"`
	File            string              `json:"file"`
}

// ConstraintIndexPath returns the canonical path to the index file.
func ConstraintIndexPath() string {
	return filepath.Join(".agents", "constraints", "index.json")
}

// ConstraintLockPath returns the canonical path to the compile lock file.
func ConstraintLockPath() string {
	return filepath.Join(".agents", "constraints", "compile.lock")
}

// LoadConstraintIndex reads and parses the constraint index.
func LoadConstraintIndex() (*ConstraintIndex, error) {
	path := ConstraintIndexPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no constraints found — run constraint-compiler.sh first")
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var idx ConstraintIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return &idx, nil
}

// WithConstraintLock acquires the compile lock and runs fn.
func WithConstraintLock(fn func() error) error {
	lockPath := ConstraintLockPath()
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return fmt.Errorf("create constraints dir: %w", err)
	}

	var lockFile *os.File
	var err error
	for attempt := 0; attempt < 20; attempt++ {
		lockFile, err = os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			break
		}
		if !os.IsExist(err) {
			return fmt.Errorf("acquire constraint lock: %w", err)
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err != nil {
		return fmt.Errorf("acquire constraint lock: %w", err)
	}
	defer func() {
		_ = lockFile.Close()
		_ = os.Remove(lockPath)
	}()

	return fn()
}

// SaveConstraintIndexUnlocked writes the index without acquiring the lock.
func SaveConstraintIndexUnlocked(idx *ConstraintIndex) error {
	path := ConstraintIndexPath()
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling index: %w", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create constraints dir: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "index.json.tmp.*")
	if err != nil {
		return fmt.Errorf("create temp constraint index: %w", err)
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("write temp constraint index: %w", err)
	}
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("chmod temp constraint index: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close temp constraint index: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename constraint index: %w", err)
	}
	return nil
}

// SaveConstraintIndex writes the constraint index back to disk under lock.
func SaveConstraintIndex(idx *ConstraintIndex) error {
	return WithConstraintLock(func() error {
		return SaveConstraintIndexUnlocked(idx)
	})
}

// FindConstraint locates a constraint by ID and returns its pointer.
func FindConstraint(idx *ConstraintIndex, id string) *ConstraintEntry {
	for i := range idx.Constraints {
		if idx.Constraints[i].ID == id {
			return &idx.Constraints[i]
		}
	}
	return nil
}

// FilterStaleConstraints returns active/draft constraints compiled before cutoff.
func FilterStaleConstraints(entries []ConstraintEntry, cutoff time.Time) []ConstraintEntry {
	stale := make([]ConstraintEntry, 0)
	for _, c := range entries {
		if c.Status == "retired" {
			continue
		}
		compiled, parseErr := time.Parse(time.RFC3339, c.CompiledAt)
		if parseErr != nil {
			compiled, parseErr = time.Parse("2006-01-02", c.CompiledAt)
			if parseErr != nil {
				continue
			}
		}
		if compiled.Before(cutoff) {
			stale = append(stale, c)
		}
	}
	return stale
}
