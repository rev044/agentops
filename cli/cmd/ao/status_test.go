package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTruncateStatus(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncated", "hello world this is long", 10, "hello w..."},
		{"empty string", "", 10, ""},
		{"with newline", "first line\nsecond line", 60, "first line"},
		{"newline only", "\nsecond line", 60, ""},
		{"maxLen 4", "hello", 4, "h..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateStatus(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateStatus(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestFirstLine(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"single line", "hello", "hello"},
		{"multi line", "first\nsecond\nthird", "first"},
		{"empty string", "", ""},
		{"starts with newline", "\nfirst", ""},
		{"trailing newline", "hello\n", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := firstLine(tt.input)
			if got != tt.want {
				t.Errorf("firstLine(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFindLastForgeTime(t *testing.T) {
	t.Run("finds most recent file", func(t *testing.T) {
		tmp := t.TempDir()
		retrosDir := filepath.Join(tmp, ".agents", "retros")
		learningsDir := filepath.Join(tmp, ".agents", "learnings")
		if err := os.MkdirAll(retrosDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(learningsDir, 0755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(retrosDir, "retro-1.md"), []byte("retro"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(learningsDir, "L1.md"), []byte("learning"), 0644); err != nil {
			t.Fatal(err)
		}

		result := findLastForgeTime(tmp)
		if result.IsZero() {
			t.Error("expected non-zero time")
		}
		// Should be very recent (within last minute)
		if time.Since(result) > time.Minute {
			t.Errorf("last forge time too old: %v", result)
		}
	})

	t.Run("empty dirs return zero", func(t *testing.T) {
		tmp := t.TempDir()
		if err := os.MkdirAll(filepath.Join(tmp, ".agents", "retros"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(tmp, ".agents", "learnings"), 0755); err != nil {
			t.Fatal(err)
		}

		result := findLastForgeTime(tmp)
		if !result.IsZero() {
			t.Errorf("expected zero time, got %v", result)
		}
	})

	t.Run("nonexistent dirs return zero", func(t *testing.T) {
		tmp := t.TempDir()
		result := findLastForgeTime(tmp)
		if !result.IsZero() {
			t.Errorf("expected zero time, got %v", result)
		}
	})

	t.Run("ignores subdirectories", func(t *testing.T) {
		tmp := t.TempDir()
		retrosDir := filepath.Join(tmp, ".agents", "retros")
		if err := os.MkdirAll(filepath.Join(retrosDir, "subdir"), 0755); err != nil {
			t.Fatal(err)
		}

		result := findLastForgeTime(tmp)
		if !result.IsZero() {
			t.Errorf("expected zero time (dirs should be ignored), got %v", result)
		}
	})
}

func TestFormatDurationBrief(t *testing.T) {
	tests := []struct {
		name  string
		input time.Duration
		want  string
	}{
		{"sub-minute", 30 * time.Second, "<1m"},
		{"minutes", 45 * time.Minute, "45m"},
		{"hours", 5 * time.Hour, "5h"},
		{"days", 3 * 24 * time.Hour, "3d"},
		{"weeks", 45 * 24 * time.Hour, "6w"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDurationBrief(tt.input)
			if got != tt.want {
				t.Errorf("formatDurationBrief(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestLoadFlywheelBrief_IncludesStigmergicScorecard(t *testing.T) {
	tmp := t.TempDir()
	for _, rel := range []string{
		filepath.Join(".agents", "findings"),
		filepath.Join(".agents", "planning-rules"),
		filepath.Join(".agents", "pre-mortem-checks"),
		filepath.Join(".agents", "rpi"),
	} {
		if err := os.MkdirAll(filepath.Join(tmp, rel), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(tmp, ".agents", "findings", "f-1.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".agents", "planning-rules", "f-1.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".agents", "pre-mortem-checks", "f-1.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	queue := `{"source_epic":"ag-h83","timestamp":"2026-03-11T17:00:00Z","items":[{"title":"High one","type":"task","severity":"high","source":"council-finding","description":"d1","target_repo":"agentops","consumed":false}],"consumed":false,"claim_status":"available","claimed_by":null,"claimed_at":null,"consumed_by":null,"consumed_at":null}
`
	if err := os.WriteFile(filepath.Join(tmp, ".agents", "rpi", "next-work.jsonl"), []byte(queue), 0o644); err != nil {
		t.Fatal(err)
	}

	brief := loadFlywheelBrief(tmp)
	if brief == nil {
		t.Fatal("expected flywheel brief")
	}
	if brief.PromotedFindings != 1 || brief.PlanningRules != 1 || brief.PreMortemChecks != 1 {
		t.Fatalf("brief signal counts = %+v, want 1/1/1", brief)
	}
	if brief.UnconsumedItems != 1 || brief.HighSeverityUnconsumed != 1 {
		t.Fatalf("brief backlog counts = %+v, want 1/1", brief)
	}
}

func TestPrintFlywheelHealth_IncludesBacklogLine(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	printFlywheelHealth(&flywheelBrief{
		Status:                 "COMPOUNDING",
		TotalArtifacts:         10,
		Velocity:               1.2,
		NewArtifacts:           3,
		StaleArtifacts:         1,
		PromotedFindings:       2,
		PlanningRules:          2,
		PreMortemChecks:        1,
		UnconsumedItems:        7,
		HighSeverityUnconsumed: 3,
	})

	_ = w.Close()
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "Backlog:") {
		t.Fatalf("expected backlog line, got: %q", got)
	}
}
