package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// curate.go — curateParseFrontmatter (0% → higher)
// ---------------------------------------------------------------------------

func TestCov16_curateParseFrontmatter_noFrontmatter(t *testing.T) {
	data := "Just plain content without frontmatter"
	fm, body := curateParseFrontmatter(data)
	if len(fm) != 0 {
		t.Errorf("expected empty fm for no frontmatter, got %v", fm)
	}
	if body != data {
		t.Errorf("body should be full content when no frontmatter")
	}
}

func TestCov16_curateParseFrontmatter_shortString(t *testing.T) {
	// Less than 3 lines → no frontmatter
	fm, body := curateParseFrontmatter("one line")
	_ = fm
	_ = body
}

func TestCov16_curateParseFrontmatter_withValidFrontmatter(t *testing.T) {
	data := "---\ntype: learning\ndate: 2026-01-01\n---\n# Content here\n\nBody text."
	fm, body := curateParseFrontmatter(data)
	if fm["type"] != "learning" {
		t.Errorf("fm type: got %v, want 'learning'", fm["type"])
	}
	if !strings.Contains(body, "Content here") {
		t.Errorf("body should contain 'Content here', got %q", body)
	}
}

func TestCov16_curateParseFrontmatter_noClosingDelimiter(t *testing.T) {
	// Opening --- but no closing ---
	data := "---\ntype: learning\nno closing delimiter here"
	fm, body := curateParseFrontmatter(data)
	if len(fm) != 0 {
		t.Errorf("expected empty fm when no closing delimiter, got %v", fm)
	}
	_ = body
}

func TestCov16_curateParseFrontmatter_emptyBody(t *testing.T) {
	data := "---\ntype: learning\n---\n"
	fm, body := curateParseFrontmatter(data)
	if fm["type"] != "learning" {
		t.Errorf("fm type: got %v, want 'learning'", fm["type"])
	}
	_ = body
}

// ---------------------------------------------------------------------------
// curate.go — curateFrontmatterString (0% → higher)
// ---------------------------------------------------------------------------

func TestCov16_curateFrontmatterString_missingKey(t *testing.T) {
	fm := map[string]any{"other": "value"}
	result := curateFrontmatterString(fm, "missing")
	if result != "" {
		t.Errorf("expected empty string for missing key, got %q", result)
	}
}

func TestCov16_curateFrontmatterString_nilValue(t *testing.T) {
	fm := map[string]any{"key": nil}
	result := curateFrontmatterString(fm, "key")
	if result != "" {
		t.Errorf("expected empty string for nil value, got %q", result)
	}
}

func TestCov16_curateFrontmatterString_stringValue(t *testing.T) {
	fm := map[string]any{"type": "  learning  "}
	result := curateFrontmatterString(fm, "type")
	if result != "learning" {
		t.Errorf("expected 'learning', got %q", result)
	}
}

func TestCov16_curateFrontmatterString_timeValue(t *testing.T) {
	ts := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	fm := map[string]any{"date": ts}
	result := curateFrontmatterString(fm, "date")
	if result != "2026-01-15" {
		t.Errorf("expected '2026-01-15', got %q", result)
	}
}

func TestCov16_curateFrontmatterString_intValue(t *testing.T) {
	fm := map[string]any{"count": 42}
	result := curateFrontmatterString(fm, "count")
	if result != "42" {
		t.Errorf("expected '42', got %q", result)
	}
}

// ---------------------------------------------------------------------------
// curate.go — resolveCurateGoalsFile (0% → higher)
// ---------------------------------------------------------------------------

func TestCov16_resolveCurateGoalsFile_noFile(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	_, err := resolveCurateGoalsFile()
	if err == nil {
		t.Error("expected error for no goals file, got nil")
	}
}

func TestCov16_resolveCurateGoalsFile_withGoalsMD(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "GOALS.md"), []byte("# Goals\n"), 0644); err != nil {
		t.Fatalf("write GOALS.md: %v", err)
	}
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	path, err := resolveCurateGoalsFile()
	if err != nil {
		t.Fatalf("resolveCurateGoalsFile: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty path")
	}
}

// ---------------------------------------------------------------------------
// curate.go — generateArtifactID (0% → higher)
// ---------------------------------------------------------------------------

func TestCov16_generateArtifactID_allTypes(t *testing.T) {
	types := []struct {
		atype    string
		expected string
	}{
		{"learning", "learn"},
		{"decision", "decis"},
		{"failure", "fail"},
		{"pattern", "patt"},
	}
	for _, tc := range types {
		id := generateArtifactID(tc.atype, "2026-01-01", "test content")
		if !strings.HasPrefix(id, tc.expected) {
			t.Errorf("generateArtifactID %s: expected prefix %q, got %q", tc.atype, tc.expected, id)
		}
	}
}

func TestCov16_generateArtifactID_unknownType(t *testing.T) {
	id := generateArtifactID("unknown-type", "2026-01-01", "test")
	// Unknown type → empty prefix but still generates ID with date and hash
	_ = id // no panic
}

// ---------------------------------------------------------------------------
// curate.go — curateArtifactDir (0% → higher)
// ---------------------------------------------------------------------------

func TestCov16_curateArtifactDir(t *testing.T) {
	if dir := curateArtifactDir("pattern"); dir != ".agents/patterns" {
		t.Errorf("pattern dir: got %q, want '.agents/patterns'", dir)
	}
	if dir := curateArtifactDir("learning"); dir != ".agents/learnings" {
		t.Errorf("learning dir: got %q, want '.agents/learnings'", dir)
	}
	if dir := curateArtifactDir("decision"); dir != ".agents/learnings" {
		t.Errorf("decision dir: got %q, want '.agents/learnings'", dir)
	}
}

// ---------------------------------------------------------------------------
// curate.go — countArtifactsInDir (0% → higher)
// ---------------------------------------------------------------------------

func TestCov16_countArtifactsInDir_missingDir(t *testing.T) {
	counts, _ := countArtifactsInDir("/tmp/nonexistent-curate-dir-xyz123")
	if len(counts) != 0 {
		t.Errorf("expected empty counts for missing dir, got %v", counts)
	}
}

func TestCov16_countArtifactsInDir_emptyDir(t *testing.T) {
	tmp := t.TempDir()
	counts, latest := countArtifactsInDir(tmp)
	if len(counts) != 0 {
		t.Errorf("expected empty counts, got %v", counts)
	}
	if !latest.IsZero() {
		t.Error("expected zero time for empty dir")
	}
}

func TestCov16_countArtifactsInDir_withArtifacts(t *testing.T) {
	tmp := t.TempDir()

	now := time.Now().UTC()
	artifact := curateArtifact{
		ID:            "learn-2026-01-01-abcdef01",
		Type:          "learning",
		Content:       "Some content",
		Date:          "2026-01-01",
		SchemaVersion: 1,
		CuratedAt:     now.Format(time.RFC3339),
		Path:          "/path/to/artifact",
	}
	data, _ := json.MarshalIndent(artifact, "", "  ")
	if err := os.WriteFile(filepath.Join(tmp, "learn-2026-01-01-abcdef01.json"), data, 0644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	// Also write a non-json file (should be skipped)
	if err := os.WriteFile(filepath.Join(tmp, "readme.txt"), []byte("not json"), 0644); err != nil {
		t.Fatalf("write txt: %v", err)
	}

	counts, latest := countArtifactsInDir(tmp)
	if counts["learning"] != 1 {
		t.Errorf("expected 1 learning artifact, got %d", counts["learning"])
	}
	if latest.IsZero() {
		t.Error("expected non-zero latest time")
	}
}

func TestCov16_countArtifactsInDir_invalidJson(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "bad.json"), []byte("not valid json"), 0644); err != nil {
		t.Fatalf("write bad.json: %v", err)
	}
	counts, _ := countArtifactsInDir(tmp)
	if len(counts) != 0 {
		t.Errorf("expected empty counts for invalid JSON, got %v", counts)
	}
}

// ---------------------------------------------------------------------------
// curate.go — curateOutWriter (0% → higher)
// ---------------------------------------------------------------------------

func TestCov16_curateOutWriter(t *testing.T) {
	// nil cmd → os.Stdout
	w := curateOutWriter(nil)
	if w != os.Stdout {
		t.Error("curateOutWriter(nil) should return os.Stdout")
	}

	// non-nil cmd → cmd.OutOrStdout()
	cmd := &cobra.Command{}
	w = curateOutWriter(cmd)
	if w == nil {
		t.Error("curateOutWriter(cmd) should not be nil")
	}
}

// ---------------------------------------------------------------------------
// curate.go — runCurateCatalog error paths (0% → higher)
// ---------------------------------------------------------------------------

func TestCov16_runCurateCatalog_missingFile(t *testing.T) {
	cmd := &cobra.Command{}
	err := runCurateCatalog(cmd, []string{"/tmp/nonexistent-artifact-xyz123.md"})
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestCov16_runCurateCatalog_missingType(t *testing.T) {
	tmp := t.TempDir()
	artifactPath := filepath.Join(tmp, "artifact.md")
	// No type field in frontmatter
	content := "---\ndate: 2026-01-01\n---\n# Content\n\nSome learning content here.\n"
	if err := os.WriteFile(artifactPath, []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cmd := &cobra.Command{}
	err := runCurateCatalog(cmd, []string{artifactPath})
	if err == nil {
		t.Fatal("expected error for missing type, got nil")
	}
	if !strings.Contains(err.Error(), "type") {
		t.Errorf("expected 'type' in error, got: %v", err)
	}
}

func TestCov16_runCurateCatalog_unknownType(t *testing.T) {
	tmp := t.TempDir()
	artifactPath := filepath.Join(tmp, "artifact.md")
	content := "---\ntype: invalid-type-xyz\ndate: 2026-01-01\n---\n# Content\n\nSome content.\n"
	if err := os.WriteFile(artifactPath, []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cmd := &cobra.Command{}
	err := runCurateCatalog(cmd, []string{artifactPath})
	if err == nil {
		t.Fatal("expected error for unknown type, got nil")
	}
}

func TestCov16_runCurateCatalog_noContent(t *testing.T) {
	tmp := t.TempDir()
	artifactPath := filepath.Join(tmp, "artifact.md")
	// Valid type but no body and no content field
	content := "---\ntype: learning\ndate: 2026-01-01\n---\n"
	if err := os.WriteFile(artifactPath, []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cmd := &cobra.Command{}
	err := runCurateCatalog(cmd, []string{artifactPath})
	if err == nil {
		t.Fatal("expected error for no content, got nil")
	}
	if !strings.Contains(err.Error(), "content") {
		t.Errorf("expected 'content' in error, got: %v", err)
	}
}

func TestCov16_runCurateCatalog_validArtifact(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	artifactPath := filepath.Join(tmp, "learning.md")
	content := "---\ntype: learning\ndate: 2026-01-01\n---\n# A great learning\n\nThis is the content of the learning artifact.\n"
	if err := os.WriteFile(artifactPath, []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	origOutput := output
	defer func() { output = origOutput }()
	output = "" // text mode

	cmd := &cobra.Command{}
	err := runCurateCatalog(cmd, []string{artifactPath})
	if err != nil {
		t.Fatalf("runCurateCatalog valid artifact: %v", err)
	}
}

func TestCov16_runCurateCatalog_validArtifact_jsonOutput(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	artifactPath := filepath.Join(tmp, "decision.md")
	content := "---\ntype: decision\ndate: 2026-01-01\n---\n# A decision\n\nWe decided to use X over Y.\n"
	if err := os.WriteFile(artifactPath, []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	origOutput := output
	defer func() { output = origOutput }()
	output = "json"

	cmd := &cobra.Command{}
	err := runCurateCatalog(cmd, []string{artifactPath})
	if err != nil {
		t.Fatalf("runCurateCatalog json: %v", err)
	}
}

// ---------------------------------------------------------------------------
// curate.go — runCurateStatus (0% → higher)
// ---------------------------------------------------------------------------

func TestCov16_runCurateStatus_emptyDirs(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	cmd := &cobra.Command{}
	err := runCurateStatus(cmd, nil)
	if err != nil {
		t.Fatalf("runCurateStatus empty dirs: %v", err)
	}
}

func TestCov16_runCurateStatus_jsonOutput(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	origOutput := output
	defer func() { output = origOutput }()
	output = "json"

	cmd := &cobra.Command{}
	err := runCurateStatus(cmd, nil)
	if err != nil {
		t.Fatalf("runCurateStatus json: %v", err)
	}
}

func TestCov16_runCurateStatus_withArtifacts(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// Create .agents/learnings/ with an artifact
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	now := time.Now().UTC()
	artifact := curateArtifact{
		ID:            "learn-2026-01-01-abcd1234",
		Type:          "learning",
		Content:       "Test learning",
		Date:          "2026-01-01",
		SchemaVersion: 1,
		CuratedAt:     now.Format(time.RFC3339),
	}
	data, _ := json.MarshalIndent(artifact, "", "  ")
	if err := os.WriteFile(filepath.Join(learningsDir, "learn-2026-01-01-abcd1234.json"), data, 0644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	origOutput := output
	defer func() { output = origOutput }()
	output = "" // text mode

	cmd := &cobra.Command{}
	err := runCurateStatus(cmd, nil)
	if err != nil {
		t.Fatalf("runCurateStatus with artifacts: %v", err)
	}
}

// ---------------------------------------------------------------------------
// curate.go — runCurateVerify (0% → higher)
// ---------------------------------------------------------------------------

func TestCov16_runCurateVerify_noGoalsFile(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	origOutput := output
	defer func() { output = origOutput }()
	output = "" // text mode

	cmd := &cobra.Command{}
	err := runCurateVerify(cmd, nil)
	if err != nil {
		t.Fatalf("runCurateVerify no goals file: %v", err)
	}
}

func TestCov16_runCurateVerify_noGoalsFile_json(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	origOutput := output
	defer func() { output = origOutput }()
	output = "json"

	cmd := &cobra.Command{}
	err := runCurateVerify(cmd, nil)
	if err != nil {
		t.Fatalf("runCurateVerify no goals file json: %v", err)
	}
}

// ---------------------------------------------------------------------------
// curate.go — countArtifactsSince (0% → higher)
// ---------------------------------------------------------------------------

func TestCov16_countArtifactsSince_noArtifacts(t *testing.T) {
	tmp := t.TempDir()
	count := countArtifactsSince(filepath.Join(tmp, "learnings"), filepath.Join(tmp, "patterns"), time.Now().Add(-time.Hour))
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestCov16_countArtifactsSince_withRecentArtifacts(t *testing.T) {
	tmp := t.TempDir()
	learningsDir := filepath.Join(tmp, "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	now := time.Now().UTC()
	artifact := curateArtifact{
		ID:        "learn-test-001",
		Type:      "learning",
		Content:   "test",
		Date:      "2026-01-01",
		CuratedAt: now.Format(time.RFC3339),
	}
	data, _ := json.MarshalIndent(artifact, "", "  ")
	if err := os.WriteFile(filepath.Join(learningsDir, "learn-test-001.json"), data, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Count since 1 hour ago → should find the just-created artifact
	count := countArtifactsSince(learningsDir, filepath.Join(tmp, "patterns"), now.Add(-time.Hour))
	if count != 1 {
		t.Errorf("expected 1 recent artifact, got %d", count)
	}
}
