package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// processAllIndexDirs (0%)
// ---------------------------------------------------------------------------

func TestIndexCov_ProcessAllIndexDirs_Empty(t *testing.T) {
	tmp := t.TempDir()
	results, stale := processAllIndexDirs(tmp, []string{".agents/learnings"}, false, true)
	// Directory doesn't exist, should skip it
	if len(results) != 0 {
		t.Errorf("expected 0 results for nonexistent dir, got %d", len(results))
	}
	if stale {
		t.Error("expected not stale for skipped dirs")
	}
}

func TestIndexCov_ProcessAllIndexDirs_WithFiles(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "2026-01-01-test.md"), []byte("---\nsummary: test\n---\n# Test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Write mode (not check, not quiet)
	results, stale := processAllIndexDirs(tmp, []string{".agents/learnings"}, false, false)
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if stale {
		t.Error("expected not stale in write mode")
	}
	if len(results) > 0 && !results[0].Written {
		t.Error("expected Written=true in write mode")
	}
}

func TestIndexCov_ProcessAllIndexDirs_CheckMode(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "2026-01-01-test.md"), []byte("---\nsummary: test\n---\n# Test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// First, write the index
	processAllIndexDirs(tmp, []string{".agents/learnings"}, false, true)

	// Now check mode — should be current
	results, stale := processAllIndexDirs(tmp, []string{".agents/learnings"}, true, false)
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if stale {
		t.Error("expected not stale when index was just written")
	}
}

func TestIndexCov_ProcessAllIndexDirs_CheckModeStale(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "2026-01-01-test.md"), []byte("---\nsummary: test\n---\n# Test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Write the index first
	processAllIndexDirs(tmp, []string{".agents/learnings"}, false, true)

	// Add a new file to make it stale
	if err := os.WriteFile(filepath.Join(dir, "2026-02-01-new.md"), []byte("---\nsummary: new\n---\n# New\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Check mode should detect staleness
	_, stale := processAllIndexDirs(tmp, []string{".agents/learnings"}, true, false)
	if !stale {
		t.Error("expected stale when new file was added after index")
	}
}

// ---------------------------------------------------------------------------
// processIndexDir (0%)
// ---------------------------------------------------------------------------

func TestIndexCov_ProcessIndexDir_WriteMode(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "learnings")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}

	entries := []indexEntry{
		{Filename: "test.md", Date: "2026-01-01", Summary: "Test", Tags: "alpha"},
	}

	origDryRun := dryRun
	dryRun = false
	t.Cleanup(func() { dryRun = origDryRun })

	result, isStale := processIndexDir(false, false, ".agents/learnings", dir, entries)
	if isStale {
		t.Error("expected not stale in write mode")
	}
	if !result.Written {
		t.Error("expected Written=true")
	}
	if result.Error != "" {
		t.Errorf("unexpected error: %s", result.Error)
	}
}

func TestIndexCov_ProcessIndexDir_CheckMode_Current(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "learnings")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}

	entries := []indexEntry{
		{Filename: "test.md", Date: "2026-01-01", Summary: "Test", Tags: "alpha"},
	}

	// Write the index first
	if err := writeIndex(dir, ".agents/learnings", entries, false); err != nil {
		t.Fatal(err)
	}

	result, isStale := processIndexDir(true, false, ".agents/learnings", dir, entries)
	if isStale {
		t.Error("expected current index not stale")
	}
	if result.Written {
		t.Error("expected Written=false in check mode")
	}
}

func TestIndexCov_ProcessIndexDir_CheckMode_Stale(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "learnings")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write index with just one entry
	initialEntries := []indexEntry{
		{Filename: "test.md", Date: "2026-01-01", Summary: "Test"},
	}
	if err := writeIndex(dir, ".agents/learnings", initialEntries, false); err != nil {
		t.Fatal(err)
	}

	// Check with more entries (stale)
	newEntries := []indexEntry{
		{Filename: "test.md", Date: "2026-01-01", Summary: "Test"},
		{Filename: "new.md", Date: "2026-02-01", Summary: "New"},
	}

	_, isStale := processIndexDir(true, false, ".agents/learnings", dir, newEntries)
	if !isStale {
		t.Error("expected stale when entries don't match")
	}
}

func TestIndexCov_ProcessIndexDir_Quiet(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "learnings")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}

	entries := []indexEntry{
		{Filename: "test.md", Date: "2026-01-01", Summary: "Test"},
	}

	origDryRun := dryRun
	dryRun = false
	t.Cleanup(func() { dryRun = origDryRun })

	// quiet mode should not panic
	result, _ := processIndexDir(false, true, ".agents/learnings", dir, entries)
	if result.Error != "" {
		t.Errorf("unexpected error: %s", result.Error)
	}
}

// ---------------------------------------------------------------------------
// runIndex (0%) — exercise through cobra
// ---------------------------------------------------------------------------

func TestIndexCov_RunIndex_WriteMode(t *testing.T) {
	tmp := t.TempDir()
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prevWD) })

	dir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "2026-01-01-test.md"), []byte("---\nsummary: test\n---\n# Test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	origDryRun := dryRun
	origOutput := output
	dryRun = false
	output = "table"
	t.Cleanup(func() {
		dryRun = origDryRun
		output = origOutput
	})

	// Reset flags on the cobra command
	_ = indexCmd.Flags().Set("check", "false")
	_ = indexCmd.Flags().Set("json", "false")
	_ = indexCmd.Flags().Set("dir", "")
	_ = indexCmd.Flags().Set("quiet", "true")

	err = runIndex(indexCmd, nil)
	if err != nil {
		t.Fatalf("runIndex() error = %v", err)
	}
}

func TestIndexCov_RunIndex_CheckMode_Stale(t *testing.T) {
	tmp := t.TempDir()
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prevWD) })

	dir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "2026-01-01-test.md"), []byte("---\nsummary: test\n---\n# Test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Don't write INDEX.md → stale

	origOutput := output
	output = "table"
	t.Cleanup(func() { output = origOutput })

	_ = indexCmd.Flags().Set("check", "true")
	_ = indexCmd.Flags().Set("json", "false")
	_ = indexCmd.Flags().Set("dir", ".agents/learnings")
	_ = indexCmd.Flags().Set("quiet", "false")

	err = runIndex(indexCmd, nil)
	if err == nil {
		t.Error("expected error for stale INDEX.md")
	}
	if err != nil && !strings.Contains(err.Error(), "stale") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestIndexCov_RunIndex_JSONOutput(t *testing.T) {
	tmp := t.TempDir()
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prevWD) })

	dir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "2026-01-01-test.md"), []byte("---\nsummary: test\n---\n# Test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	origDryRun := dryRun
	origOutput := output
	dryRun = false
	output = "json"
	t.Cleanup(func() {
		dryRun = origDryRun
		output = origOutput
	})

	_ = indexCmd.Flags().Set("check", "false")
	_ = indexCmd.Flags().Set("json", "true")
	_ = indexCmd.Flags().Set("dir", ".agents/learnings")
	_ = indexCmd.Flags().Set("quiet", "true")

	// Capture stdout
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runIndex(indexCmd, nil)

	_ = w.Close()
	os.Stdout = origStdout

	if err != nil {
		t.Fatalf("runIndex() error = %v", err)
	}

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	out := string(buf[:n])
	if !strings.Contains(out, "learnings") {
		t.Errorf("expected 'learnings' in JSON output, got: %s", out[:min(len(out), 200)])
	}
}

func TestIndexCov_RunIndex_SingleDir(t *testing.T) {
	tmp := t.TempDir()
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prevWD) })

	dir := filepath.Join(tmp, ".agents", "patterns")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "2026-01-01-retry.md"), []byte("---\nsummary: retry pattern\n---\n# Retry\n"), 0644); err != nil {
		t.Fatal(err)
	}

	origDryRun := dryRun
	origOutput := output
	dryRun = false
	output = "table"
	t.Cleanup(func() {
		dryRun = origDryRun
		output = origOutput
	})

	_ = indexCmd.Flags().Set("check", "false")
	_ = indexCmd.Flags().Set("json", "false")
	_ = indexCmd.Flags().Set("dir", ".agents/patterns")
	_ = indexCmd.Flags().Set("quiet", "true")

	err = runIndex(indexCmd, nil)
	if err != nil {
		t.Fatalf("runIndex() error = %v", err)
	}

	// Verify INDEX.md was written
	content, err := os.ReadFile(filepath.Join(dir, "INDEX.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "retry") {
		t.Error("expected INDEX.md to contain 'retry'")
	}
}

// ---------------------------------------------------------------------------
// scanAndSortDir — nonexistent dir
// ---------------------------------------------------------------------------

func TestIndexCov_ScanAndSortDir_Nonexistent(t *testing.T) {
	tmp := t.TempDir()
	_, entries, ok := scanAndSortDir(tmp, ".agents/nonexistent")
	if ok {
		t.Error("expected ok=false for nonexistent dir")
	}
	if entries != nil {
		t.Error("expected nil entries for nonexistent dir")
	}
}

func TestIndexCov_ScanAndSortDir_WithFiles(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "2026-01-01-a.md"), []byte("---\nsummary: A\n---\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "2026-02-01-b.md"), []byte("---\nsummary: B\n---\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, entries, ok := scanAndSortDir(tmp, ".agents/learnings")
	if !ok {
		t.Error("expected ok=true")
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
	// Should be sorted by date descending (2026-02-01 first)
	if len(entries) >= 2 && entries[0].Date < entries[1].Date {
		t.Errorf("expected descending date sort, got %s before %s", entries[0].Date, entries[1].Date)
	}
}

// ---------------------------------------------------------------------------
// writeIndex — dry-run
// ---------------------------------------------------------------------------

func TestIndexCov_WriteIndex_DryRun(t *testing.T) {
	tmp := t.TempDir()
	entries := []indexEntry{
		{Filename: "test.md", Date: "2026-01-01", Summary: "Test"},
	}

	err := writeIndex(tmp, ".agents/learnings", entries, true)
	if err != nil {
		t.Fatalf("writeIndex dry-run error = %v", err)
	}

	// Verify INDEX.md was NOT written
	_, err = os.Stat(filepath.Join(tmp, "INDEX.md"))
	if err == nil {
		t.Error("expected INDEX.md not written in dry-run mode")
	}
}

// ---------------------------------------------------------------------------
// parseIndexTableRows — various inputs
// ---------------------------------------------------------------------------

func TestIndexCov_ParseIndexTableRows_Empty(t *testing.T) {
	result := parseIndexTableRows([]byte(""))
	if len(result) != 0 {
		t.Errorf("expected 0 files, got %d", len(result))
	}
}

func TestIndexCov_ParseIndexTableRows_ValidTable(t *testing.T) {
	table := `# Index: Learnings

| File | Date | Summary | Tags |
|------|------|---------|------|
| test.md | 2026-01-01 | Test | alpha |
| foo.md | 2026-02-01 | Foo | beta |
`
	result := parseIndexTableRows([]byte(table))
	if len(result) != 2 {
		t.Errorf("expected 2 files, got %d", len(result))
	}
	if !result["test.md"] {
		t.Error("expected test.md in results")
	}
	if !result["foo.md"] {
		t.Error("expected foo.md in results")
	}
}

// ---------------------------------------------------------------------------
// buildIndexDiffMessage
// ---------------------------------------------------------------------------

func TestIndexCov_BuildIndexDiffMessage_MissingAndExtra(t *testing.T) {
	msg := buildIndexDiffMessage(".agents/learnings", []string{"new.md"}, []string{"old.md"})
	if !strings.Contains(msg, "missing=[new.md]") {
		t.Errorf("expected 'missing=[new.md]' in message, got: %s", msg)
	}
	if !strings.Contains(msg, "extra=[old.md]") {
		t.Errorf("expected 'extra=[old.md]' in message, got: %s", msg)
	}
}
