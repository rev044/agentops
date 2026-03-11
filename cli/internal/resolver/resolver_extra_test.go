package resolver

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestExtra_ResolveAbsPathWithinRoot_NotOnDisk verifies that an absolute path
// within the root that doesn't exist on disk falls back to basename + extension strip.
func TestExtra_ResolveAbsPathWithinRoot_NotOnDisk(t *testing.T) {
	root := t.TempDir()
	learningsDir := filepath.Join(root, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create the file that the basename-based lookup should find.
	if err := os.WriteFile(filepath.Join(learningsDir, "my-learning.md"), []byte("# Found\n"), 0644); err != nil {
		t.Fatal(err)
	}

	r := NewFileResolver(root)

	// Absolute path within root but the file itself doesn't exist at that path.
	absNonexistent := filepath.Join(root, "subdir", "my-learning.md")
	path, err := r.Resolve(absNonexistent)
	if err != nil {
		t.Fatalf("Resolve(%q) unexpected error: %v", absNonexistent, err)
	}
	if filepath.Base(path) != "my-learning.md" {
		t.Errorf("Resolve() = %q, want base my-learning.md", path)
	}
}

// TestExtra_ResolveAbsPathWithinRoot_KnownExtStrip verifies that extension
// stripping works for each known extension (.jsonl, .md, .json).
func TestExtra_ResolveAbsPathWithinRoot_KnownExtStrip(t *testing.T) {
	root := t.TempDir()
	learningsDir := filepath.Join(root, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}
	// The actual file on disk is "report.md"
	if err := os.WriteFile(filepath.Join(learningsDir, "report.md"), []byte("# Report\n"), 0644); err != nil {
		t.Fatal(err)
	}

	r := NewFileResolver(root)

	// Pass an abs path with .md extension — it should strip .md, use "report" as ID.
	absPath := filepath.Join(root, "nonexistent", "report.md")
	path, err := r.Resolve(absPath)
	if err != nil {
		t.Fatalf("Resolve(%q) error: %v", absPath, err)
	}
	if filepath.Base(path) != "report.md" {
		t.Errorf("got base %q, want report.md", filepath.Base(path))
	}
}

// TestExtra_ResolveAbsPathWithinRoot_JsonlExtStrip verifies .jsonl extension strip.
func TestExtra_ResolveAbsPathWithinRoot_JsonlExtStrip(t *testing.T) {
	root := t.TempDir()
	learningsDir := filepath.Join(root, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(learningsDir, "data.jsonl"), []byte(`{"x":1}`+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	r := NewFileResolver(root)
	absPath := filepath.Join(root, "fake", "data.jsonl")
	path, err := r.Resolve(absPath)
	if err != nil {
		t.Fatalf("Resolve(%q) error: %v", absPath, err)
	}
	if filepath.Base(path) != "data.jsonl" {
		t.Errorf("got base %q, want data.jsonl", filepath.Base(path))
	}
}

// TestExtra_ResolveAbsPathWithinRoot_JsonExtStrip verifies .json extension strip.
func TestExtra_ResolveAbsPathWithinRoot_JsonExtStrip(t *testing.T) {
	root := t.TempDir()
	learningsDir := filepath.Join(root, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(learningsDir, "config.json"), []byte(`{}`+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	r := NewFileResolver(root)
	absPath := filepath.Join(root, "fake", "config.json")
	path, err := r.Resolve(absPath)
	if err != nil {
		t.Fatalf("Resolve(%q) error: %v", absPath, err)
	}
	if filepath.Base(path) != "config.json" {
		t.Errorf("got base %q, want config.json", filepath.Base(path))
	}
}

// TestExtra_ResolveAbsPathWithinRoot_UnknownExtNoStrip verifies that an
// unknown extension (.txt) is NOT stripped — the full basename is used as ID.
func TestExtra_ResolveAbsPathWithinRoot_UnknownExtNoStrip(t *testing.T) {
	root := t.TempDir()
	learningsDir := filepath.Join(root, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create a file whose name contains the full basename (with .txt) as substring.
	if err := os.WriteFile(filepath.Join(learningsDir, "notes.txt"), []byte("notes\n"), 0644); err != nil {
		t.Fatal(err)
	}

	r := NewFileResolver(root)
	absPath := filepath.Join(root, "gone", "notes.txt")
	path, err := r.Resolve(absPath)
	if err != nil {
		t.Fatalf("Resolve(%q) error: %v", absPath, err)
	}
	// probeDirect should find "notes.txt" directly since the ext isn't stripped.
	if filepath.Base(path) != "notes.txt" {
		t.Errorf("got base %q, want notes.txt", filepath.Base(path))
	}
}

// TestExtra_ResolveAbsPathOutsideRoot verifies that an absolute path outside
// the resolver root is not mangled — the code skips the basename fallback
// (filepath.Rel returns ".." prefix), so the full path is used as-is for search.
func TestExtra_ResolveAbsPathOutsideRoot(t *testing.T) {
	root := t.TempDir()
	other := t.TempDir()

	r := NewFileResolver(root)
	// An absolute path pointing completely outside root.
	absPath := filepath.Join(other, "something.md")
	_, err := r.Resolve(absPath)
	// This should fail with not-found because the raw abs path won't match anything.
	if err == nil {
		t.Fatal("expected error for abs path outside root, got nil")
	}
}

// TestExtra_ProbeDirectUnknownExt verifies probeDirect finds a file with an
// extension not in the extensions list (e.g., .txt, .yaml).
func TestExtra_ProbeDirectUnknownExt(t *testing.T) {
	root := t.TempDir()
	learningsDir := filepath.Join(root, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}
	// File with unknown extension — probeDirect should find it.
	target := filepath.Join(learningsDir, "my-notes.txt")
	if err := os.WriteFile(target, []byte("notes"), 0644); err != nil {
		t.Fatal(err)
	}

	got := probeDirect(learningsDir, "my-notes.txt")
	if got != target {
		t.Errorf("probeDirect() = %q, want %q", got, target)
	}
}

// TestExtra_ProbeDirectMiss verifies probeDirect returns empty for missing file.
func TestExtra_ProbeDirectMiss(t *testing.T) {
	dir := t.TempDir()
	got := probeDirect(dir, "nope.txt")
	if got != "" {
		t.Errorf("probeDirect() = %q, want empty", got)
	}
}

// TestExtra_ProbeLiteralSubstring_SkipsDir verifies that probeLiteralSubstring
// skips directories whose names contain the search substring.
func TestExtra_ProbeLiteralSubstring_SkipsDir(t *testing.T) {
	root := t.TempDir()
	// Create a subdirectory whose name contains the search term.
	if err := os.MkdirAll(filepath.Join(root, "target-data"), 0755); err != nil {
		t.Fatal(err)
	}

	got := probeLiteralSubstring(root, "target-data")
	if got != "" {
		t.Errorf("probeLiteralSubstring() should skip dirs, got %q", got)
	}
}

// TestExtra_ProbeLiteralSubstring_FindsFile verifies file match.
func TestExtra_ProbeLiteralSubstring_FindsFile(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "abc-target-xyz.md"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	got := probeLiteralSubstring(root, "target")
	if filepath.Base(got) != "abc-target-xyz.md" {
		t.Errorf("probeLiteralSubstring() = %q, want abc-target-xyz.md", got)
	}
}

// TestExtra_ProbeLiteralSubstring_NonexistentDir verifies ReadDir error path.
func TestExtra_ProbeLiteralSubstring_NonexistentDir(t *testing.T) {
	got := probeLiteralSubstring("/nonexistent-dir-xyzzy", "anything")
	if got != "" {
		t.Errorf("probeLiteralSubstring() on bad dir = %q, want empty", got)
	}
}

// TestExtra_ProbeFrontmatterID_NonexistentDir verifies ReadDir error returns empty.
func TestExtra_ProbeFrontmatterID_NonexistentDir(t *testing.T) {
	got := probeFrontmatterID("/nonexistent-dir-xyzzy", "any-id")
	if got != "" {
		t.Errorf("probeFrontmatterID() on bad dir = %q, want empty", got)
	}
}

// TestExtra_ProbeFrontmatterID_UnreadableFile verifies that an unreadable .md
// file causes readFrontmatterField to error and the entry is skipped.
func TestExtra_ProbeFrontmatterID_UnreadableFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod not reliable on Windows")
	}
	root := t.TempDir()
	mdFile := filepath.Join(root, "broken.md")
	if err := os.WriteFile(mdFile, []byte("---\nid: secret\n---\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Make it unreadable.
	if err := os.Chmod(mdFile, 0000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(mdFile, 0644) })

	got := probeFrontmatterID(root, "secret")
	if got != "" {
		t.Errorf("probeFrontmatterID() should skip unreadable file, got %q", got)
	}
}

// TestExtra_ProbeFrontmatterID_SkipsDirs verifies directory entries are skipped.
func TestExtra_ProbeFrontmatterID_SkipsDirs(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "subdir.md"), 0755); err != nil {
		t.Fatal(err)
	}
	got := probeFrontmatterID(root, "anything")
	if got != "" {
		t.Errorf("probeFrontmatterID() should skip dirs, got %q", got)
	}
}

// TestExtra_ProbeFrontmatterID_NoMatch verifies no match returns empty.
func TestExtra_ProbeFrontmatterID_NoMatch(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.md"), []byte("---\nid: alpha\n---\n"), 0644); err != nil {
		t.Fatal(err)
	}
	got := probeFrontmatterID(root, "beta")
	if got != "" {
		t.Errorf("probeFrontmatterID() = %q, want empty", got)
	}
}

// TestExtra_ReadFrontmatterField_FileNotFound verifies os.Open error.
func TestExtra_ReadFrontmatterField_FileNotFound(t *testing.T) {
	_, err := readFrontmatterField("/nonexistent-path/foo.md", "id")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

// TestExtra_ReadFrontmatterField_FieldAfterClosingDelimiter verifies that a
// field appearing AFTER the closing "---" is NOT returned.
func TestExtra_ReadFrontmatterField_FieldAfterClosingDelimiter(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.md")
	content := "---\ntitle: hello\n---\nid: should-not-match\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	val, err := readFrontmatterField(path, "id")
	if err != nil {
		t.Fatalf("readFrontmatterField() error: %v", err)
	}
	if val != "" {
		t.Errorf("readFrontmatterField() = %q, want empty (field after closing ---)", val)
	}
}

// TestExtra_ReadFrontmatterField_NoFrontmatter verifies file without frontmatter.
func TestExtra_ReadFrontmatterField_NoFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "plain.md")
	if err := os.WriteFile(path, []byte("# Just a heading\nSome text.\n"), 0644); err != nil {
		t.Fatal(err)
	}
	val, err := readFrontmatterField(path, "id")
	if err != nil {
		t.Fatalf("readFrontmatterField() error: %v", err)
	}
	if val != "" {
		t.Errorf("readFrontmatterField() = %q, want empty", val)
	}
}

// TestExtra_ReadFrontmatterField_Success verifies happy path extraction.
func TestExtra_ReadFrontmatterField_Success(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "good.md")
	content := "---\nid: my-learning-id\ntitle: Test\n---\n# Body\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	val, err := readFrontmatterField(path, "id")
	if err != nil {
		t.Fatalf("readFrontmatterField() error: %v", err)
	}
	if val != "my-learning-id" {
		t.Errorf("readFrontmatterField() = %q, want my-learning-id", val)
	}
}

// TestExtra_ReadFrontmatterField_QuotedValue verifies quote stripping.
func TestExtra_ReadFrontmatterField_QuotedValue(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "quoted.md")
	content := "---\nid: \"quoted-id\"\n---\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	val, err := readFrontmatterField(path, "id")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if val != "quoted-id" {
		t.Errorf("got %q, want quoted-id", val)
	}
}

// TestExtra_InSlice covers true, false, and nil slice cases.
func TestExtra_InSlice(t *testing.T) {
	tests := []struct {
		name   string
		needle string
		slice  []string
		want   bool
	}{
		{"found", "b", []string{"a", "b", "c"}, true},
		{"not found", "z", []string{"a", "b", "c"}, false},
		{"nil slice", "x", nil, false},
		{"empty slice", "x", []string{}, false},
		{"single match", "only", []string{"only"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inSlice(tt.needle, tt.slice)
			if got != tt.want {
				t.Errorf("inSlice(%q, %v) = %v, want %v", tt.needle, tt.slice, got, tt.want)
			}
		})
	}
}

// TestExtra_DiscoverAll_GlobError exercises the filepath.Glob error → continue
// path inside DiscoverAll's collect closure. Glob returns an error only for
// malformed patterns, and we can trigger this by using a directory path
// containing a malformed glob metacharacter. On most OSes, a literal '[' in
// the path will cause Glob to return ErrBadPattern.
func TestExtra_DiscoverAll_GlobError(t *testing.T) {
	// Create a root that has an .agents/learnings dir with a name containing '['
	// to cause filepath.Glob(filepath.Join(dir, "*.md")) to fail.
	root := t.TempDir()
	badDir := filepath.Join(root, ".agents", "learnings")
	if err := os.MkdirAll(badDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Place a valid file so we know the function doesn't crash.
	if err := os.WriteFile(filepath.Join(badDir, "ok.md"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	// Now create a SECOND root whose path contains '[' to trigger Glob error.
	badRoot := filepath.Join(t.TempDir(), "bad[dir")
	badLearnings := filepath.Join(badRoot, ".agents", "learnings")
	if err := os.MkdirAll(badLearnings, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(badLearnings, "test.md"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	r := NewFileResolverWithGlobal(root, []string{badLearnings})
	files, err := r.DiscoverAll()
	if err != nil {
		t.Fatalf("DiscoverAll() error: %v", err)
	}
	// The root's learnings should still be found despite the bad glob dir.
	found := false
	for _, f := range files {
		if filepath.Base(f) == "ok.md" {
			found = true
		}
	}
	if !found {
		t.Error("DiscoverAll() should still find ok.md despite glob error in other dir")
	}
}

// TestExtra_ProbeWithExtensions_Miss verifies probeWithExtensions returns
// empty when no file matches.
func TestExtra_ProbeWithExtensions_Miss(t *testing.T) {
	dir := t.TempDir()
	got := probeWithExtensions(dir, "nonexistent-id")
	if got != "" {
		t.Errorf("probeWithExtensions() = %q, want empty", got)
	}
}

// TestExtra_BuildAgentsDirs verifies the correct subdirs are constructed.
func TestExtra_BuildAgentsDirs(t *testing.T) {
	dirs := buildAgentsDirs("/fake/root")
	if len(dirs) != 3 {
		t.Fatalf("buildAgentsDirs() returned %d dirs, want 3", len(dirs))
	}
	expected := []string{
		filepath.Join("/fake/root", ".agents", "learnings"),
		filepath.Join("/fake/root", ".agents", "findings"),
		filepath.Join("/fake/root", ".agents", "patterns"),
	}
	for i, want := range expected {
		if dirs[i] != want {
			t.Errorf("dirs[%d] = %q, want %q", i, dirs[i], want)
		}
	}
}
