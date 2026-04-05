package resolver

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func setupTestLearnings(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	learningsDir := filepath.Join(root, ".agents", "learnings")
	patternsDir := filepath.Join(root, ".agents", "patterns")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(patternsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// JSONL learning
	if err := os.WriteFile(filepath.Join(learningsDir, "L001.jsonl"), []byte(`{"id":"L001","title":"test"}`+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Markdown learning
	if err := os.WriteFile(filepath.Join(learningsDir, "L002.md"), []byte("---\nid: L002\ntitle: test\n---\n# L002\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Learning with long filename for glob matching
	if err := os.WriteFile(filepath.Join(learningsDir, "learning-003.jsonl"), []byte(`{"id":"003","title":"three"}`+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Pattern file
	if err := os.WriteFile(filepath.Join(patternsDir, "retry-backoff.md"), []byte("# Retry Backoff\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Learning with frontmatter ID different from filename
	if err := os.WriteFile(filepath.Join(learningsDir, "some-file.md"), []byte("---\nid: learn-2026-02-21-backend-detection\ntitle: Backend Detection\n---\n# Content\n"), 0644); err != nil {
		t.Fatal(err)
	}

	return root
}

func TestFileResolver_Resolve(t *testing.T) {
	root := setupTestLearnings(t)
	r := NewFileResolver(root)

	tests := []struct {
		name     string
		id       string
		wantBase string
		wantErr  bool
	}{
		{
			name:     "resolve by ID with jsonl extension",
			id:       "L001",
			wantBase: "L001.jsonl",
		},
		{
			name:     "resolve by ID with md extension",
			id:       "L002",
			wantBase: "L002.md",
		},
		{
			name:     "resolve by filename without extension",
			id:       "learning-003",
			wantBase: "learning-003.jsonl",
		},
		{
			name:     "resolve by glob (partial ID)",
			id:       "003",
			wantBase: "learning-003.jsonl",
		},
		{
			name:     "resolve pattern by name",
			id:       "retry-backoff",
			wantBase: "retry-backoff.md",
		},
		{
			name:     "resolve by frontmatter ID",
			id:       "learn-2026-02-21-backend-detection",
			wantBase: "some-file.md",
		},
		{
			name:    "not found returns error",
			id:      "nonexistent-xyz-999",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := r.Resolve(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
				return
			}
			if tt.wantBase != "" && filepath.Base(path) != tt.wantBase {
				t.Errorf("Resolve(%q) = %q, want base %q", tt.id, path, tt.wantBase)
			}
		})
	}
}

func TestFileResolver_Resolve_LiteralMetacharacterSubstring(t *testing.T) {
	root := t.TempDir()
	learningsDir := filepath.Join(root, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{
		"prefix[abc]suffix.md",
		"alpha-star-target.md",
		"prefix*star.md",
		"alphaqmark-target.md",
		"prefix?mark.md",
	} {
		if err := os.WriteFile(filepath.Join(learningsDir, name), []byte("test\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	r := NewFileResolver(root)

	tests := []struct {
		name     string
		id       string
		wantBase string
	}{
		{
			name:     "unbalanced bracket ID is treated literally",
			id:       "[abc",
			wantBase: "prefix[abc]suffix.md",
		},
		{
			name:     "star ID is treated literally",
			id:       "*star",
			wantBase: "prefix*star.md",
		},
		{
			name:     "question ID is treated literally",
			id:       "?mark",
			wantBase: "prefix?mark.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := r.Resolve(tt.id)
			if err != nil {
				t.Fatalf("Resolve(%q) error = %v", tt.id, err)
			}
			if filepath.Base(path) != tt.wantBase {
				t.Errorf("Resolve(%q) = %q, want base %q", tt.id, path, tt.wantBase)
			}
		})
	}
}

func TestFileResolver_Resolve_PoolID(t *testing.T) {
	root := t.TempDir()
	learningsDir := filepath.Join(root, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}
	// File named after the learning ID (without pend- prefix)
	if err := os.WriteFile(filepath.Join(learningsDir, "fix-auth-bug.md"), []byte("# Fix Auth Bug\n"), 0644); err != nil {
		t.Fatal(err)
	}

	r := NewFileResolver(root)
	path, err := r.Resolve("pend-fix-auth-bug")
	if err != nil {
		t.Fatalf("Resolve(pend-fix-auth-bug) error = %v", err)
	}
	if filepath.Base(path) != "fix-auth-bug.md" {
		t.Errorf("Resolve(pend-fix-auth-bug) = %q, want base fix-auth-bug.md", path)
	}
}

func TestFileResolver_Resolve_AbsolutePath(t *testing.T) {
	root := t.TempDir()
	learningsDir := filepath.Join(root, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}
	absPath := filepath.Join(learningsDir, "L001.jsonl")
	if err := os.WriteFile(absPath, []byte(`{"id":"L001"}`+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	r := NewFileResolver(root)
	path, err := r.Resolve(absPath)
	if err != nil {
		t.Fatalf("Resolve(absolute path) error = %v", err)
	}
	if path != absPath {
		t.Errorf("Resolve(absolute path) = %q, want %q", path, absPath)
	}
}

func TestFileResolver_Resolve_ParentWalk(t *testing.T) {
	// Create a nested structure: root/.agents/learnings/L001.md
	// Then resolve from root/sub/dir
	root := t.TempDir()
	learningsDir := filepath.Join(root, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(learningsDir, "L001.md"), []byte("# L001\n"), 0644); err != nil {
		t.Fatal(err)
	}

	subDir := filepath.Join(root, "sub", "dir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	r := NewFileResolver(subDir)
	path, err := r.Resolve("L001")
	if err != nil {
		t.Fatalf("Resolve from subdir error = %v", err)
	}
	if filepath.Base(path) != "L001.md" {
		t.Errorf("Resolve from subdir = %q, want base L001.md", path)
	}
}

func TestFileResolver_Resolve_NotFoundError(t *testing.T) {
	root := t.TempDir()
	r := NewFileResolver(root)
	_, err := r.Resolve("NONEXISTENT")
	if err == nil {
		t.Fatal("expected error for nonexistent learning")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want containing 'not found'", err.Error())
	}
}

func TestFileResolver_ImplementsInterface(t *testing.T) {
	var _ LearningResolver = &FileResolver{}
}

func TestFileResolver_DiscoverAll(t *testing.T) {
	root := setupTestLearnings(t)
	r := NewFileResolver(root)

	files, err := r.DiscoverAll()
	if err != nil {
		t.Fatalf("DiscoverAll() error = %v", err)
	}

	// setupTestLearnings creates: L001.jsonl, L002.md, learning-003.jsonl, some-file.md in learnings/
	// and retry-backoff.md in patterns/
	if len(files) != 5 {
		t.Errorf("DiscoverAll() returned %d files, want 5", len(files))
		for _, f := range files {
			t.Logf("  %s", f)
		}
	}

	// Verify known files are present
	bases := make(map[string]bool)
	for _, f := range files {
		bases[filepath.Base(f)] = true
	}
	for _, want := range []string{"L001.jsonl", "L002.md", "learning-003.jsonl", "some-file.md", "retry-backoff.md"} {
		if !bases[want] {
			t.Errorf("DiscoverAll() missing %s", want)
		}
	}
}

func TestFileResolver_DiscoverAll_Empty(t *testing.T) {
	root := t.TempDir()
	r := NewFileResolver(root)

	files, err := r.DiscoverAll()
	if err != nil {
		t.Fatalf("DiscoverAll() error = %v", err)
	}
	if len(files) != 0 {
		t.Errorf("DiscoverAll() on empty dir returned %d files, want 0", len(files))
	}
}

func TestDiscoverAll_WithGlobalDirs(t *testing.T) {
	// Set up local repo with one learning
	root := t.TempDir()
	localDir := filepath.Join(root, ".agents", "learnings")
	if err := os.MkdirAll(localDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(localDir, "local.md"), []byte("# Local"), 0644); err != nil {
		t.Fatal(err)
	}

	// Set up global dir with one learning
	globalDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(globalDir, "global.md"), []byte("# Global"), 0644); err != nil {
		t.Fatal(err)
	}

	r := NewFileResolverWithGlobal(root, []string{globalDir})
	files, err := r.DiscoverAll()
	if err != nil {
		t.Fatalf("DiscoverAll() error = %v", err)
	}

	if len(files) != 2 {
		t.Errorf("DiscoverAll() returned %d files, want 2", len(files))
	}

	// Verify both local and global files found
	hasLocal, hasGlobal := false, false
	for _, f := range files {
		if filepath.Base(f) == "local.md" {
			hasLocal = true
		}
		if filepath.Base(f) == "global.md" {
			hasGlobal = true
		}
	}
	if !hasLocal {
		t.Error("DiscoverAll() missing local file")
	}
	if !hasGlobal {
		t.Error("DiscoverAll() missing global file")
	}
}

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
	if err := os.WriteFile(filepath.Join(learningsDir, "report.md"), []byte("# Report\n"), 0644); err != nil {
		t.Fatal(err)
	}

	r := NewFileResolver(root)

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
// unknown extension (.txt) is NOT stripped.
func TestExtra_ResolveAbsPathWithinRoot_UnknownExtNoStrip(t *testing.T) {
	root := t.TempDir()
	learningsDir := filepath.Join(root, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(learningsDir, "notes.txt"), []byte("notes\n"), 0644); err != nil {
		t.Fatal(err)
	}

	r := NewFileResolver(root)
	absPath := filepath.Join(root, "gone", "notes.txt")
	path, err := r.Resolve(absPath)
	if err != nil {
		t.Fatalf("Resolve(%q) error: %v", absPath, err)
	}
	if filepath.Base(path) != "notes.txt" {
		t.Errorf("got base %q, want notes.txt", filepath.Base(path))
	}
}

// TestExtra_ResolveAbsPathOutsideRoot verifies that an absolute path outside
// the resolver root is not mangled.
func TestExtra_ResolveAbsPathOutsideRoot(t *testing.T) {
	root := t.TempDir()
	other := t.TempDir()

	r := NewFileResolver(root)
	absPath := filepath.Join(other, "something.md")
	_, err := r.Resolve(absPath)
	if err == nil {
		t.Fatal("expected error for abs path outside root, got nil")
	}
}

// TestExtra_ProbeDirectUnknownExt verifies probeDirect finds a file with an
// extension not in the extensions list.
func TestExtra_ProbeDirectUnknownExt(t *testing.T) {
	root := t.TempDir()
	learningsDir := filepath.Join(root, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}
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
	if os.Getuid() == 0 {
		t.Skip("root bypasses filesystem permissions")
	}
	root := t.TempDir()
	mdFile := filepath.Join(root, "broken.md")
	if err := os.WriteFile(mdFile, []byte("---\nid: secret\n---\n"), 0644); err != nil {
		t.Fatal(err)
	}
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

// TestExtra_DiscoverAll_GlobError exercises the filepath.Glob error path.
func TestExtra_DiscoverAll_GlobError(t *testing.T) {
	root := t.TempDir()
	badDir := filepath.Join(root, ".agents", "learnings")
	if err := os.MkdirAll(badDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(badDir, "ok.md"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

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

func TestResolve_FallsBackToGlobal(t *testing.T) {
	// Local repo with no learnings
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".agents", "learnings"), 0755); err != nil {
		t.Fatal(err)
	}

	// Global dir with the learning we want
	globalDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(globalDir, "target.md"), []byte("# Target"), 0644); err != nil {
		t.Fatal(err)
	}

	r := NewFileResolverWithGlobal(root, []string{globalDir})
	path, err := r.Resolve("target")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if filepath.Base(path) != "target.md" {
		t.Errorf("Resolve() = %q, want target.md", path)
	}
}
