package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComputeResult(t *testing.T) {
	tests := []struct {
		name       string
		checks     []doctorCheck
		wantResult string
		wantFails  bool
	}{
		{
			name: "all pass",
			checks: []doctorCheck{
				{Name: "a", Status: "pass", Required: true},
				{Name: "b", Status: "pass", Required: true},
			},
			wantResult: "HEALTHY",
			wantFails:  false,
		},
		{
			name: "one failure",
			checks: []doctorCheck{
				{Name: "a", Status: "pass", Required: true},
				{Name: "b", Status: "fail", Required: true},
			},
			wantResult: "UNHEALTHY",
			wantFails:  true,
		},
		{
			name: "warnings only",
			checks: []doctorCheck{
				{Name: "a", Status: "pass", Required: true},
				{Name: "b", Status: "warn", Required: false},
			},
			wantResult: "HEALTHY",
			wantFails:  false,
		},
		{
			name: "mixed failures and warnings",
			checks: []doctorCheck{
				{Name: "a", Status: "fail", Required: true},
				{Name: "b", Status: "warn", Required: false},
				{Name: "c", Status: "pass", Required: true},
			},
			wantResult: "UNHEALTHY",
			wantFails:  true,
		},
		{
			name:       "empty checks",
			checks:     []doctorCheck{},
			wantResult: "HEALTHY",
			wantFails:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := computeResult(tt.checks)
			if output.Result != tt.wantResult {
				t.Errorf("computeResult() result = %q, want %q", output.Result, tt.wantResult)
			}
			if tt.wantFails && output.Summary == "all checks passed" {
				t.Error("expected failure in summary")
			}
			if !tt.wantFails && len(tt.checks) > 0 && !hasWarns(tt.checks) && output.Summary != "all checks passed" {
				t.Errorf("expected 'all checks passed', got %q", output.Summary)
			}
		})
	}
}

func hasWarns(checks []doctorCheck) bool {
	for _, c := range checks {
		if c.Status == "warn" {
			return true
		}
	}
	return false
}

func TestCountFiles(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("empty directory", func(t *testing.T) {
		got := countFiles(tmpDir)
		if got != 0 {
			t.Errorf("countFiles(empty) = %d, want 0", got)
		}
	})

	t.Run("with files", func(t *testing.T) {
		os.WriteFile(filepath.Join(tmpDir, "a.md"), []byte("test"), 0644)
		os.WriteFile(filepath.Join(tmpDir, "b.md"), []byte("test"), 0644)
		os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)

		got := countFiles(tmpDir)
		if got != 2 {
			t.Errorf("countFiles() = %d, want 2 (should not count directories)", got)
		}
	})

	t.Run("nonexistent directory", func(t *testing.T) {
		got := countFiles(filepath.Join(tmpDir, "nonexistent"))
		if got != 0 {
			t.Errorf("countFiles(nonexistent) = %d, want 0", got)
		}
	})
}
