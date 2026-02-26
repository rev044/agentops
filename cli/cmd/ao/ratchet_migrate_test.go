package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShouldMigrateFile(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		isDir bool
		want  bool
	}{
		{"markdown file", "test.md", false, true},
		{"nested markdown", "dir/sub/file.md", false, true},
		{"go file", "test.go", false, false},
		{"txt file", "test.txt", false, false},
		{"directory", "somedir", true, false},
		{"md directory", "test.md", true, false},
		{"json file", "data.json", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := fakeFileInfo{name: filepath.Base(tt.path), isDir: tt.isDir}
			got := shouldMigrateFile(tt.path, info)
			if got != tt.want {
				t.Errorf("shouldMigrateFile(%q, isDir=%v) = %v, want %v", tt.path, tt.isDir, got, tt.want)
			}
		})
	}
}

func TestFindSchemaInsertPoint(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  int
	}{
		{
			name:  "after Date field",
			lines: []string{"# Title", "**Date:** 2025-01-15", "some content"},
			want:  2,
		},
		{
			name:  "after Epic field",
			lines: []string{"# Title", "**Epic:** ol-0001", "some content"},
			want:  2,
		},
		{
			name:  "Date takes priority over heading",
			lines: []string{"# Title", "**Date:** 2025-01-15", "more"},
			want:  2,
		},
		{
			name:  "fallback to after heading",
			lines: []string{"# Title", "some content", "more content"},
			want:  1,
		},
		{
			name:  "no heading or date",
			lines: []string{"plain text", "more text"},
			want:  -1,
		},
		{
			name:  "empty lines",
			lines: []string{},
			want:  -1,
		},
		{
			name:  "heading at end only line",
			lines: []string{"# Title"},
			want:  -1,
		},
		{
			name:  "heading with content after",
			lines: []string{"# Title", "body"},
			want:  1,
		},
		{
			name:  "multiple headings uses first",
			lines: []string{"# First", "# Second", "body"},
			want:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findSchemaInsertPoint(tt.lines)
			if got != tt.want {
				t.Errorf("findSchemaInsertPoint() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestMigrateFile_AddsSchemaVersion(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.md")

	content := "# Test Article\n**Date:** 2025-01-15\nBody text here.\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	origDryRun := dryRun
	dryRun = false
	defer func() { dryRun = origDryRun }()

	result := migrateFile(path, info)
	if result != migrateResultSuccess {
		t.Fatalf("migrateFile() = %d, want migrateResultSuccess (%d)", result, migrateResultSuccess)
	}

	migrated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migrated file: %v", err)
	}

	if !strings.Contains(string(migrated), "**Schema Version:** 1") {
		t.Errorf("migrated file missing schema version field\nContent:\n%s", string(migrated))
	}
}

func TestMigrateFile_SkipsExistingSchemaVersion(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.md")

	content := "# Test\n**Schema Version:** 1\n**Date:** 2025-01-15\nBody.\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	result := migrateFile(path, info)
	if result != migrateResultSkipped {
		t.Errorf("migrateFile() = %d, want migrateResultSkipped (%d)", result, migrateResultSkipped)
	}
}

func TestMigrateFile_SkipsLowercaseSchemaVersion(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.md")

	content := "# Test\nschema_version: 1\n**Date:** 2025-01-15\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	result := migrateFile(path, info)
	if result != migrateResultSkipped {
		t.Errorf("migrateFile() = %d, want migrateResultSkipped", result)
	}
}

func TestMigrateFile_SkipsNoInsertPoint(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.md")

	content := "plain text with no heading or date field\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	result := migrateFile(path, info)
	if result != migrateResultSkipped {
		t.Errorf("migrateFile() = %d, want migrateResultSkipped", result)
	}
}

func TestMigrateFile_DryRunDoesNotWrite(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.md")

	content := "# Test\n**Date:** 2025-01-15\nBody.\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	origDryRun := dryRun
	dryRun = true
	defer func() { dryRun = origDryRun }()

	result := migrateFile(path, info)
	if result != migrateResultSuccess {
		t.Fatalf("migrateFile() dry-run = %d, want migrateResultSuccess", result)
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	if string(after) != content {
		t.Errorf("file was modified during dry-run\nBefore: %s\nAfter: %s", content, string(after))
	}
}

func TestMigrateFile_NonexistentFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.md")

	info := fakeFileInfo{name: "does-not-exist.md", isDir: false}

	result := migrateFile(path, info)
	if result != migrateResultError {
		t.Errorf("migrateFile(nonexistent) = %d, want migrateResultError (%d)", result, migrateResultError)
	}
}

func TestMigrateFile_InsertAfterEpicField(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.md")

	content := "# Test\n**Epic:** ol-0001\nBody text.\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	origDryRun := dryRun
	dryRun = false
	defer func() { dryRun = origDryRun }()

	result := migrateFile(path, info)
	if result != migrateResultSuccess {
		t.Fatalf("migrateFile() = %d, want migrateResultSuccess", result)
	}

	migrated, _ := os.ReadFile(path)
	if !strings.Contains(string(migrated), "**Schema Version:** 1") {
		t.Errorf("missing schema version after Epic field\nContent:\n%s", string(migrated))
	}
}

func TestMigrateResult_Constants(t *testing.T) {
	if migrateResultSuccess == migrateResultSkipped {
		t.Error("migrateResultSuccess should differ from migrateResultSkipped")
	}
	if migrateResultSuccess == migrateResultError {
		t.Error("migrateResultSuccess should differ from migrateResultError")
	}
	if migrateResultSkipped == migrateResultError {
		t.Error("migrateResultSkipped should differ from migrateResultError")
	}
}
