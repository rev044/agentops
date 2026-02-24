package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestConstraintIndexPath(t *testing.T) {
	want := filepath.Join(".agents", "constraints", "index.json")
	if got := constraintIndexPath(); got != want {
		t.Fatalf("constraintIndexPath() = %q, want %q", got, want)
	}
}

func TestConstraintLoadSaveFindRoundTrip(t *testing.T) {
	wd := t.TempDir()
	oldWD := chdirTo(t, wd)
	t.Cleanup(func() { _ = os.Chdir(oldWD) })
	mkdirConstraintsDir(t)

	idx := &constraintIndex{
		SchemaVersion: 1,
		Constraints: []constraintEntry{
			{
				ID:         "c-1",
				Title:      "No eval",
				Status:     "draft",
				CompiledAt: "2026-01-01",
				File:       "notes.md",
			},
		},
	}

	if err := saveConstraintIndex(idx); err != nil {
		t.Fatalf("saveConstraintIndex: %v", err)
	}

	got, err := loadConstraintIndex()
	if err != nil {
		t.Fatalf("loadConstraintIndex: %v", err)
	}
	if got.SchemaVersion != 1 {
		t.Fatalf("SchemaVersion = %d, want 1", got.SchemaVersion)
	}
	if gotEntry := findConstraint(got, "c-1"); gotEntry == nil || gotEntry.ID != "c-1" {
		t.Fatalf("findConstraint() = %v, want entry c-1", gotEntry)
	}
}

func TestConstraintLoadMissingIndex(t *testing.T) {
	wd := t.TempDir()
	oldWD := chdirTo(t, wd)
	t.Cleanup(func() { _ = os.Chdir(oldWD) })

	if _, err := loadConstraintIndex(); err == nil {
		t.Fatal("expected loadConstraintIndex to fail when index file is missing")
	}
}

func TestConstraintActivateDraft(t *testing.T) {
	wd := t.TempDir()
	oldWD := chdirTo(t, wd)
	t.Cleanup(func() { _ = os.Chdir(oldWD) })
	mkdirConstraintsDir(t)

	idx := &constraintIndex{
		SchemaVersion: 1,
		Constraints: []constraintEntry{
			{ID: "c-1", Status: "draft", CompiledAt: time.Now().Format(time.RFC3339)},
		},
	}
	if err := saveConstraintIndex(idx); err != nil {
		t.Fatalf("saveConstraintIndex: %v", err)
	}

	oldJSON := constraintJSON
	constraintJSON = false
	t.Cleanup(func() { constraintJSON = oldJSON })

	stdout, err := captureStdout(t, func() error {
		return constraintActivateCmd.RunE(constraintActivateCmd, []string{"c-1"})
	})
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !strings.Contains(stdout, "Constraint \"c-1\" activated") {
		t.Fatalf("expected activation output, got: %q", stdout)
	}

	reloaded, err := loadConstraintIndex()
	if err != nil {
		t.Fatalf("loadConstraintIndex: %v", err)
	}
	if got := findConstraint(reloaded, "c-1").Status; got != "active" {
		t.Fatalf("constraint status = %q, want active", got)
	}
}

func TestConstraintRetireActive(t *testing.T) {
	wd := t.TempDir()
	oldWD := chdirTo(t, wd)
	t.Cleanup(func() { _ = os.Chdir(oldWD) })
	mkdirConstraintsDir(t)

	idx := &constraintIndex{
		SchemaVersion: 1,
		Constraints: []constraintEntry{
			{ID: "c-1", Status: "active", CompiledAt: time.Now().Format(time.RFC3339)},
		},
	}
	if err := saveConstraintIndex(idx); err != nil {
		t.Fatalf("saveConstraintIndex: %v", err)
	}

	oldJSON := constraintJSON
	constraintJSON = false
	t.Cleanup(func() { constraintJSON = oldJSON })

	stdout, err := captureStdout(t, func() error {
		return constraintRetireCmd.RunE(constraintRetireCmd, []string{"c-1"})
	})
	if err != nil {
		t.Fatalf("retire: %v", err)
	}
	if !strings.Contains(stdout, "Constraint \"c-1\" retired") {
		t.Fatalf("expected retirement output, got: %q", stdout)
	}

	reloaded, err := loadConstraintIndex()
	if err != nil {
		t.Fatalf("loadConstraintIndex: %v", err)
	}
	if got := findConstraint(reloaded, "c-1").Status; got != "retired" {
		t.Fatalf("constraint status = %q, want retired", got)
	}
}

func TestConstraintListAndReviewNoConstraints(t *testing.T) {
	wd := t.TempDir()
	oldWD := chdirTo(t, wd)
	t.Cleanup(func() { _ = os.Chdir(oldWD) })
	mkdirConstraintsDir(t)

	empty := &constraintIndex{SchemaVersion: 1}
	if err := saveConstraintIndex(empty); err != nil {
		t.Fatalf("saveConstraintIndex: %v", err)
	}

	oldJSON := constraintJSON
	constraintJSON = false
	t.Cleanup(func() { constraintJSON = oldJSON })

	listOut, err := captureStdout(t, func() error {
		return constraintListCmd.RunE(constraintListCmd, nil)
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(listOut, "No constraints found.") {
		t.Fatalf("expected no constraints message, got: %q", listOut)
	}

	reviewOut, err := captureStdout(t, func() error {
		return constraintReviewCmd.RunE(constraintReviewCmd, nil)
	})
	if err != nil {
		t.Fatalf("review: %v", err)
	}
	if !strings.Contains(reviewOut, "No constraints need review.") {
		t.Fatalf("expected no review candidates message, got: %q", reviewOut)
	}
}

func TestConstraintActivateJSONAndReviewStale(t *testing.T) {
	wd := t.TempDir()
	oldWD := chdirTo(t, wd)
	t.Cleanup(func() { _ = os.Chdir(oldWD) })
	mkdirConstraintsDir(t)

	oldJSON := constraintJSON
	constraintJSON = true
	t.Cleanup(func() { constraintJSON = oldJSON })

	stale := &constraintIndex{
		SchemaVersion: 1,
		Constraints: []constraintEntry{
			{
				ID:         "old-c1",
				Title:      "Old constraint",
				Status:     "draft",
				CompiledAt: time.Now().AddDate(0, 0, -91).Format(time.RFC3339),
			},
		},
	}
	if err := saveConstraintIndex(stale); err != nil {
		t.Fatalf("saveConstraintIndex: %v", err)
	}

	reviewOut, err := captureStdout(t, func() error {
		return constraintReviewCmd.RunE(constraintReviewCmd, nil)
	})
	if err != nil {
		t.Fatalf("review: %v", err)
	}
	if !json.Valid([]byte(reviewOut)) {
		t.Fatalf("expected JSON output for review, got: %q", reviewOut)
	}

	stdout, err := captureStdout(t, func() error {
		return constraintActivateCmd.RunE(constraintActivateCmd, []string{"old-c1"})
	})
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("expected activation JSON: %v", err)
	}
	if got["id"] != "old-c1" || got["status"] != "active" {
		t.Fatalf("unexpected activation payload: %v", got)
	}
}

func TestConstraintActivateRejectsNonDraft(t *testing.T) {
	wd := t.TempDir()
	oldWD := chdirTo(t, wd)
	t.Cleanup(func() { _ = os.Chdir(oldWD) })
	mkdirConstraintsDir(t)

	idx := &constraintIndex{
		SchemaVersion: 1,
		Constraints: []constraintEntry{
			{ID: "c-1", Status: "active", CompiledAt: time.Now().Format(time.RFC3339)},
		},
	}
	if err := saveConstraintIndex(idx); err != nil {
		t.Fatalf("saveConstraintIndex: %v", err)
	}

	if err := constraintActivateCmd.RunE(constraintActivateCmd, []string{"c-1"}); err == nil {
		t.Fatal("expected activate to reject non-draft constraint")
	}
}

func TestConstraintRetireRejectsMissingID(t *testing.T) {
	wd := t.TempDir()
	oldWD := chdirTo(t, wd)
	t.Cleanup(func() { _ = os.Chdir(oldWD) })
	mkdirConstraintsDir(t)

	if err := os.WriteFile(constraintIndexPath(), []byte(`{"schema_version":1,"constraints":[{"id":"c-1","status":"draft","compiled_at":"2026-01-01"}]`), 0644); err != nil {
		t.Fatalf("write index: %v", err)
	}
	if err := constraintRetireCmd.RunE(constraintRetireCmd, []string{"unknown"}); err == nil {
		t.Fatal("expected retire missing ID to fail")
	}
}

func chdirTo(t *testing.T, wd string) string {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(wd); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	return prev
}

func TestConstraintReviewJSONRoundTrip(t *testing.T) {
	wd := t.TempDir()
	oldWD := chdirTo(t, wd)
	t.Cleanup(func() { _ = os.Chdir(oldWD) })
	mkdirConstraintsDir(t)

	oldJSON := constraintJSON
	constraintJSON = true
	t.Cleanup(func() { constraintJSON = oldJSON })

	idx := &constraintIndex{
		SchemaVersion: 1,
		Constraints: []constraintEntry{
			{
				ID:     "c-1",
				Title:  "Recent",
				Status: "draft",
				CompiledAt: time.Now().Format(time.RFC3339),
			},
			{
				ID:     "c-2",
				Title:  "Old",
				Status: "active",
				CompiledAt: time.Now().AddDate(0, 0, -100).Format(time.RFC3339),
			},
		},
	}
	if err := saveConstraintIndex(idx); err != nil {
		t.Fatalf("saveConstraintIndex: %v", err)
	}

	out, err := captureStdout(t, func() error {
		return constraintListCmd.RunE(constraintListCmd, nil)
	})
	if err != nil {
		t.Fatalf("list json: %v", err)
	}
	var got []constraintEntry
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("list json unmarshal: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 constraints in list JSON, got %d", len(got))
	}

	reviewOut, err := captureStdout(t, func() error {
		return constraintReviewCmd.RunE(constraintReviewCmd, nil)
	})
	if err != nil {
		t.Fatalf("review json: %v", err)
	}
	var review []constraintEntry
	if err := json.Unmarshal([]byte(reviewOut), &review); err != nil {
		t.Fatalf("review json unmarshal: %v", err)
	}
	if len(review) != 1 {
		t.Fatalf("expected 1 stale constraint, got %d", len(review))
	}
}

func mkdirConstraintsDir(t *testing.T) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(".agents", "constraints"), 0o755); err != nil {
		t.Fatalf("mkdir .agents/constraints: %v", err)
	}
}
