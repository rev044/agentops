package goals

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMeasureOne_Pass(t *testing.T) {
	goal := Goal{
		ID:     "test-pass",
		Check:  "exit 0",
		Weight: 5,
	}
	m := MeasureOne(goal, 5*time.Second)
	if m.Result != "pass" {
		t.Errorf("expected pass, got %q", m.Result)
	}
	if m.GoalID != "test-pass" {
		t.Errorf("GoalID = %q, want %q", m.GoalID, "test-pass")
	}
	if m.Weight != 5 {
		t.Errorf("Weight = %d, want 5", m.Weight)
	}
	if m.Duration < 0 {
		t.Errorf("Duration should be >= 0, got %f", m.Duration)
	}
}

func TestMeasureOne_Fail(t *testing.T) {
	goal := Goal{
		ID:     "test-fail",
		Check:  "exit 1",
		Weight: 3,
	}
	m := MeasureOne(goal, 5*time.Second)
	if m.Result != "fail" {
		t.Errorf("expected fail, got %q", m.Result)
	}
}

func TestMeasureOne_Timeout(t *testing.T) {
	goal := Goal{
		ID:     "test-timeout",
		Check:  "sleep 10",
		Weight: 1,
	}
	m := MeasureOne(goal, 100*time.Millisecond)
	if m.Result != "skip" {
		t.Errorf("expected skip on timeout, got %q", m.Result)
	}
}

func TestMeasureOne_OutputTruncated(t *testing.T) {
	goal := Goal{
		ID:     "test-truncate",
		Check:  "printf '%600s' | tr ' ' 'A'",
		Weight: 1,
	}
	m := MeasureOne(goal, 5*time.Second)
	if len([]rune(m.Output)) > 500 {
		t.Errorf("output should be truncated to 500 runes, got %d", len([]rune(m.Output)))
	}
}

func TestMeasureOne_ContinuousMetric_ParsesValue(t *testing.T) {
	threshold := 0.5
	goal := Goal{
		ID:     "test-continuous",
		Check:  "echo 0.75",
		Weight: 2,
		Continuous: &ContinuousMetric{
			Metric:    "my_metric",
			Threshold: threshold,
		},
	}
	m := MeasureOne(goal, 5*time.Second)
	if m.Value == nil {
		t.Fatal("expected Value to be set for continuous metric")
	}
	if *m.Value != 0.75 {
		t.Errorf("Value = %f, want 0.75", *m.Value)
	}
	if m.Threshold == nil {
		t.Fatal("expected Threshold to be set for continuous metric")
	}
	if *m.Threshold != threshold {
		t.Errorf("Threshold = %f, want %f", *m.Threshold, threshold)
	}
}

func TestMeasureOne_ContinuousMetric_NonNumericOutput(t *testing.T) {
	goal := Goal{
		ID:     "test-nonnumeric",
		Check:  "echo hello",
		Weight: 1,
		Continuous: &ContinuousMetric{
			Metric:    "my_metric",
			Threshold: 0.5,
		},
	}
	m := MeasureOne(goal, 5*time.Second)
	// Non-numeric output: Value should remain nil
	if m.Value != nil {
		t.Errorf("expected Value to be nil for non-numeric output, got %f", *m.Value)
	}
}

func TestMeasure_MetaGoalsRunFirst(t *testing.T) {
	gf := &GoalFile{
		Version: 2,
		Goals: []Goal{
			{ID: "non-meta-1", Check: "exit 0", Weight: 1, Type: GoalTypeHealth},
			{ID: "meta-1", Check: "exit 0", Weight: 1, Type: GoalTypeMeta},
			{ID: "non-meta-2", Check: "exit 0", Weight: 1, Type: GoalTypeQuality},
		},
	}
	snap := Measure(gf, 5*time.Second)
	if len(snap.Goals) != 3 {
		t.Fatalf("expected 3 measurements, got %d", len(snap.Goals))
	}
	// Meta goal should come first
	if snap.Goals[0].GoalID != "meta-1" {
		t.Errorf("expected meta-1 first, got %q", snap.Goals[0].GoalID)
	}
}

func TestMeasure_SummaryCorrect(t *testing.T) {
	gf := &GoalFile{
		Version: 2,
		Goals: []Goal{
			{ID: "pass-1", Check: "exit 0", Weight: 5, Type: GoalTypeHealth},
			{ID: "pass-2", Check: "exit 0", Weight: 3, Type: GoalTypeHealth},
			{ID: "fail-1", Check: "exit 1", Weight: 2, Type: GoalTypeHealth},
		},
	}
	snap := Measure(gf, 5*time.Second)
	if snap.Summary.Total != 3 {
		t.Errorf("Total = %d, want 3", snap.Summary.Total)
	}
	if snap.Summary.Passing != 2 {
		t.Errorf("Passing = %d, want 2", snap.Summary.Passing)
	}
	if snap.Summary.Failing != 1 {
		t.Errorf("Failing = %d, want 1", snap.Summary.Failing)
	}
	// Weighted score: (5+3)/(5+3+2) * 100 = 80
	expectedScore := 80.0
	if snap.Summary.Score != expectedScore {
		t.Errorf("Score = %f, want %f", snap.Summary.Score, expectedScore)
	}
}

func TestMeasure_SkippedGoalsExcludedFromScore(t *testing.T) {
	gf := &GoalFile{
		Version: 2,
		Goals: []Goal{
			{ID: "pass-1", Check: "exit 0", Weight: 5, Type: GoalTypeHealth},
			{ID: "skip-1", Check: "sleep 10", Weight: 10, Type: GoalTypeHealth},
		},
	}
	snap := Measure(gf, 50*time.Millisecond)
	if snap.Summary.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", snap.Summary.Skipped)
	}
	// Score should only count passing/failing, not skipped
	// pass-1 weight=5, skip-1 excluded; score = 5/5 * 100 = 100
	if snap.Summary.Score != 100.0 {
		t.Errorf("Score = %f, want 100.0 (skipped excluded)", snap.Summary.Score)
	}
}

func TestRequiresExclusiveExecution(t *testing.T) {
	tests := []struct {
		name string
		goal Goal
		want bool
	}{
		{
			name: "go test gate is exclusive",
			goal: Goal{Check: "cd cli && go test -race ./..."},
			want: true,
		},
		{
			name: "cmd ao coverage script is exclusive",
			goal: Goal{Check: "bash scripts/check-cmdao-coverage-floor.sh"},
			want: true,
		},
		{
			name: "shell lint is not exclusive",
			goal: Goal{Check: "find . -name '*.sh' -print0 | xargs -0 shellcheck"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := requiresExclusiveExecution(tt.goal); got != tt.want {
				t.Fatalf("requiresExclusiveExecution(%q) = %v, want %v", tt.goal.Check, got, tt.want)
			}
		})
	}
}

func TestRunGoals_ExclusiveChecksDoNotOverlap(t *testing.T) {
	gf := &GoalFile{
		Version: 2,
		Goals: []Goal{
			{ID: "exclusive-1", Check: "sleep 0.3 # go test", Weight: 1, Type: GoalTypeHealth},
			{ID: "normal-1", Check: "sleep 0.3", Weight: 1, Type: GoalTypeHealth},
		},
	}

	start := time.Now()
	snap := Measure(gf, 2*time.Second)
	elapsed := time.Since(start)

	if len(snap.Goals) != 2 {
		t.Fatalf("expected 2 goals, got %d", len(snap.Goals))
	}
	if elapsed < 500*time.Millisecond {
		t.Fatalf("expected exclusive goal to serialize execution, elapsed=%s", elapsed)
	}
}

func TestMeasure_EmptyGoals(t *testing.T) {
	gf := &GoalFile{Version: 2, Goals: []Goal{}}
	snap := Measure(gf, 5*time.Second)
	if snap.Summary.Total != 0 {
		t.Errorf("Total = %d, want 0", snap.Summary.Total)
	}
	if snap.Summary.Score != 0 {
		t.Errorf("Score = %f, want 0 for empty goals", snap.Summary.Score)
	}
	if snap.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}
}

func TestTruncateOutput_MultiByteRunes(t *testing.T) {
	// 501 copies of '世' (3 bytes each = 1503 bytes).
	// truncateOutput should limit to 500 runes, not 500 bytes.
	runes := make([]rune, 501)
	for i := range runes {
		runes[i] = '世'
	}
	input := []byte(string(runes))

	result := truncateOutput(input)
	runeCount := len([]rune(result))
	if runeCount > 500 {
		t.Errorf("expected <=500 runes, got %d", runeCount)
	}
	if runeCount < 500 {
		t.Errorf("expected exactly 500 runes (not fewer), got %d", runeCount)
	}
	// Verify the result is valid UTF-8 (no mid-rune truncation).
	for i, r := range result {
		if r == '\uFFFD' {
			t.Errorf("invalid UTF-8 at byte %d (replacement character found)", i)
			break
		}
	}
}

func TestGitSHA_OutsideGitRepo(t *testing.T) {
	// Exercise the gitSHA error path (line 120-122).
	// Change to a temp dir that is NOT a git repo, call gitSHA, then restore.
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(origDir) //nolint:errcheck // best effort restore
	}()

	sha := gitSHA()
	if sha != "" {
		t.Errorf("expected empty SHA outside git repo, got %q", sha)
	}
}

func TestGitSHAWithTimeout_WhenGitBlocks(t *testing.T) {
	origPath := os.Getenv("PATH")
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	tmpDir := t.TempDir()
	binDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}

	fakeGit := filepath.Join(binDir, "git")
	script := "#!/usr/bin/env bash\nsleep 10\n"
	if err := os.WriteFile(fakeGit, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.Setenv("PATH", binDir+string(os.PathListSeparator)+origPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Setenv("PATH", origPath)
		_ = os.Chdir(origDir)
	})

	start := time.Now()
	sha := gitSHAWithTimeout(50 * time.Millisecond)
	elapsed := time.Since(start)

	if sha != "" {
		t.Fatalf("expected empty SHA on timeout, got %q", sha)
	}
	if elapsed > time.Second {
		t.Fatalf("expected gitSHA timeout quickly, took %s", elapsed)
	}
}
