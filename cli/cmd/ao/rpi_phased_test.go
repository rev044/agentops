package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractCouncilVerdict(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
		wantErr  bool
	}{
		{
			name:     "PASS verdict",
			content:  "# Pre-Mortem\n\n## Council Verdict: PASS\n\nDetails here.",
			expected: "PASS",
		},
		{
			name:     "WARN verdict",
			content:  "## Council Verdict: WARN\n\nSome concerns.",
			expected: "WARN",
		},
		{
			name:     "FAIL verdict",
			content:  "## Council Verdict: FAIL\n\nCritical issues.",
			expected: "FAIL",
		},
		{
			name:    "no verdict",
			content: "# Report\n\nNo verdict line here.",
			wantErr: true,
		},
		{
			name:    "empty file",
			content: "",
			wantErr: true,
		},
		{
			name:     "verdict with extra whitespace",
			content:  "## Council Verdict:  PASS \n",
			expected: "PASS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "report.md")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			verdict, err := extractCouncilVerdict(path)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got verdict %q", verdict)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if verdict != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, verdict)
			}
		})
	}
}

func TestExtractCouncilFindings(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		max      int
		expected int
	}{
		{
			name:     "structured findings",
			content:  "FINDING: Missing auth | FIX: Add middleware | REF: auth.go:10\nFINDING: No tests | FIX: Add tests | REF: auth_test.go",
			max:      5,
			expected: 2,
		},
		{
			name:     "max limit applied",
			content:  "FINDING: A | FIX: B | REF: C\nFINDING: D | FIX: E | REF: F\nFINDING: G | FIX: H | REF: I",
			max:      2,
			expected: 2,
		},
		{
			name:     "fallback to markdown findings",
			content:  "## Shared Findings\n\n1. **Missing auth** — No middleware\n2. **No tests** — Zero coverage",
			max:      5,
			expected: 2,
		},
		{
			name:     "no findings",
			content:  "# Empty report",
			max:      5,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "report.md")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			findings, err := extractCouncilFindings(path, tt.max)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(findings) != tt.expected {
				t.Errorf("expected %d findings, got %d", tt.expected, len(findings))
			}
		})
	}
}

func TestBuildPromptForPhase(t *testing.T) {
	tests := []struct {
		name     string
		phase    int
		state    *phasedState
		contains string
	}{
		{
			name:     "research phase",
			phase:    1,
			state:    &phasedState{Goal: "add auth"},
			contains: `/research "add auth" --auto`,
		},
		{
			name:     "plan phase",
			phase:    2,
			state:    &phasedState{Goal: "add auth"},
			contains: `/plan "add auth" --auto`,
		},
		{
			name:     "pre-mortem normal",
			phase:    3,
			state:    &phasedState{},
			contains: "/pre-mortem",
		},
		{
			name:     "pre-mortem fast path",
			phase:    3,
			state:    &phasedState{FastPath: true},
			contains: "--quick",
		},
		{
			name:     "crank with epic",
			phase:    4,
			state:    &phasedState{EpicID: "ag-5k2"},
			contains: "/crank ag-5k2",
		},
		{
			name:     "crank with test-first",
			phase:    4,
			state:    &phasedState{EpicID: "ag-5k2", TestFirst: true},
			contains: "--test-first",
		},
		{
			name:     "vibe normal",
			phase:    5,
			state:    &phasedState{},
			contains: "/vibe recent",
		},
		{
			name:     "vibe fast path",
			phase:    5,
			state:    &phasedState{FastPath: true},
			contains: "/vibe --quick recent",
		},
		{
			name:     "post-mortem with epic",
			phase:    6,
			state:    &phasedState{EpicID: "ag-5k2"},
			contains: "/post-mortem ag-5k2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt, err := buildPromptForPhase("", tt.phase, tt.state, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !containsStr(prompt, tt.contains) {
				t.Errorf("prompt %q does not contain %q", prompt, tt.contains)
			}
		})
	}
}

func TestBuildPromptForPhase_Retry(t *testing.T) {
	state := &phasedState{Goal: "add auth", EpicID: "ag-5k2"}
	retryCtx := &retryContext{
		Attempt: 2,
		Findings: []finding{
			{Description: "Missing error handling", Fix: "Add try-catch", Ref: "auth.go:42"},
		},
		Verdict: "FAIL",
	}

	// Pre-mortem retry → re-plan
	prompt, err := buildRetryPrompt("", 3, state, retryCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !containsStr(prompt, "/plan") {
		t.Errorf("pre-mortem retry should invoke /plan, got: %q", prompt)
	}
	if !containsStr(prompt, "Missing error handling") {
		t.Errorf("retry prompt should contain finding description, got: %q", prompt)
	}

	// Vibe retry → re-crank
	prompt, err = buildRetryPrompt("", 5, state, retryCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !containsStr(prompt, "/crank") {
		t.Errorf("vibe retry should invoke /crank, got: %q", prompt)
	}
}

func TestPhasedState_SaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, ".agents", "rpi")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatal(err)
	}

	original := &phasedState{
		Goal:      "test goal",
		EpicID:    "ag-test",
		Phase:     3,
		Cycle:     1,
		FastPath:  true,
		TestFirst: false,
		Verdicts:  map[string]string{"pre_mortem": "PASS"},
		Attempts:  map[string]int{"phase_3": 1},
		StartedAt: "2026-02-14T12:00:00Z",
	}

	if err := savePhasedState(tmpDir, original); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := loadPhasedState(tmpDir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}

	if loaded.Goal != original.Goal {
		t.Errorf("goal: got %q, want %q", loaded.Goal, original.Goal)
	}
	if loaded.EpicID != original.EpicID {
		t.Errorf("epic_id: got %q, want %q", loaded.EpicID, original.EpicID)
	}
	if loaded.Phase != original.Phase {
		t.Errorf("phase: got %d, want %d", loaded.Phase, original.Phase)
	}
	if loaded.FastPath != original.FastPath {
		t.Errorf("fast_path: got %v, want %v", loaded.FastPath, original.FastPath)
	}
	if loaded.Verdicts["pre_mortem"] != "PASS" {
		t.Errorf("verdicts: got %v, want pre_mortem=PASS", loaded.Verdicts)
	}

	// Verify JSON round-trip
	data, _ := json.Marshal(original)
	var roundTrip phasedState
	if err := json.Unmarshal(data, &roundTrip); err != nil {
		t.Fatalf("round-trip unmarshal: %v", err)
	}
	if roundTrip.Goal != original.Goal {
		t.Errorf("round-trip goal mismatch")
	}
}

func TestPhaseNameToNum(t *testing.T) {
	tests := []struct {
		name     string
		expected int
	}{
		{"research", 1},
		{"plan", 2},
		{"pre-mortem", 3},
		{"premortem", 3},
		{"pre_mortem", 3},
		{"crank", 4},
		{"implement", 4},
		{"vibe", 5},
		{"validate", 5},
		{"post-mortem", 6},
		{"postmortem", 6},
		{"post_mortem", 6},
		{"unknown", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := phaseNameToNum(tt.name)
			if got != tt.expected {
				t.Errorf("phaseNameToNum(%q) = %d, want %d", tt.name, got, tt.expected)
			}
		})
	}
}

func TestFindLatestCouncilReport(t *testing.T) {
	tmpDir := t.TempDir()
	councilDir := filepath.Join(tmpDir, ".agents", "council")
	if err := os.MkdirAll(councilDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create two reports with different timestamps
	if err := os.WriteFile(filepath.Join(councilDir, "2026-02-13-pre-mortem-auth.md"), []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(councilDir, "2026-02-14-pre-mortem-auth.md"), []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}
	// Unrelated report
	if err := os.WriteFile(filepath.Join(councilDir, "2026-02-14-vibe-recent.md"), []byte("vibe"), 0644); err != nil {
		t.Fatal(err)
	}

	// Should find the latest pre-mortem report
	report, err := findLatestCouncilReport(tmpDir, "pre-mortem")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !containsStr(report, "2026-02-14-pre-mortem") {
		t.Errorf("expected latest report, got: %s", report)
	}

	// Should find vibe report
	report, err = findLatestCouncilReport(tmpDir, "vibe")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !containsStr(report, "vibe-recent") {
		t.Errorf("expected vibe report, got: %s", report)
	}

	// Should error on missing pattern
	_, err = findLatestCouncilReport(tmpDir, "nonexistent")
	if err == nil {
		t.Error("expected error for missing pattern")
	}
}

func TestParseFastPath(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected bool
	}{
		{"empty output (no issues)", "", true},
		{"one issue no blockers", "ag-001  open  Fix login bug", true},
		{"two issues no blockers", "ag-001  open  Fix login bug\nag-002  open  Add tests", true},
		{"three issues", "ag-001  open  Fix login\nag-002  open  Add tests\nag-003  open  Refactor", false},
		{"one blocked issue", "ag-001  blocked  Fix login bug", false},
		{"two issues one blocked", "ag-001  open  Fix login\nag-002  blocked  Add tests", false},
		{"whitespace only lines", "  \n  \n", true},
		{"mixed with empty lines", "ag-001  open  Fix login\n\nag-002  open  Add tests\n", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseFastPath(tt.output)
			if got != tt.expected {
				t.Errorf("parseFastPath(%q) = %v, want %v", tt.output, got, tt.expected)
			}
		})
	}
}

func TestParseCrankCompletion(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected string
	}{
		{"empty output", "", "DONE"},
		{"all closed", "ag-001  closed  Fix login\nag-002  ✓  Add tests", "DONE"},
		{"one blocked", "ag-001  closed  Fix login\nag-002  blocked  Add tests", "BLOCKED"},
		{"partial", "ag-001  closed  Fix login\nag-002  open  Add tests", "PARTIAL"},
		{"all open", "ag-001  open  Fix login\nag-002  open  Add tests", "PARTIAL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCrankCompletion(tt.output)
			if got != tt.expected {
				t.Errorf("parseCrankCompletion(%q) = %q, want %q", tt.output, got, tt.expected)
			}
		})
	}
}

func TestPhasedState_SchemaVersion(t *testing.T) {
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, ".agents", "rpi")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatal(err)
	}

	state := &phasedState{
		SchemaVersion: 1,
		Goal:          "test",
		Verdicts:      make(map[string]string),
		Attempts:      make(map[string]int),
	}

	if err := savePhasedState(tmpDir, state); err != nil {
		t.Fatalf("save error: %v", err)
	}

	// Verify JSON contains schema_version
	data, err := os.ReadFile(filepath.Join(stateDir, phasedStateFile))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if v, ok := raw["schema_version"]; !ok {
		t.Error("schema_version missing from JSON")
	} else if v.(float64) != 1 {
		t.Errorf("schema_version = %v, want 1", v)
	}

	loaded, err := loadPhasedState(tmpDir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.SchemaVersion != 1 {
		t.Errorf("loaded SchemaVersion = %d, want 1", loaded.SchemaVersion)
	}
}

func TestBuildPromptForPhase_Interactive(t *testing.T) {
	// Save and restore global flag
	orig := phasedInteractive
	defer func() { phasedInteractive = orig }()

	state := &phasedState{Goal: "add auth"}

	// Default (non-interactive) — should have --auto
	phasedInteractive = false
	prompt, err := buildPromptForPhase("", 1, state, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !containsStr(prompt, "--auto") {
		t.Errorf("non-interactive research prompt should contain --auto, got: %q", prompt)
	}

	// Interactive — should NOT have --auto
	phasedInteractive = true
	prompt, err = buildPromptForPhase("", 1, state, nil)
	if err != nil {
		t.Fatal(err)
	}
	if containsStr(prompt, "--auto") {
		t.Errorf("interactive research prompt should not contain --auto, got: %q", prompt)
	}

	// Plan phase too
	prompt, err = buildPromptForPhase("", 2, state, nil)
	if err != nil {
		t.Fatal(err)
	}
	if containsStr(prompt, "--auto") {
		t.Errorf("interactive plan prompt should not contain --auto, got: %q", prompt)
	}
}

func TestBuildPhaseContext(t *testing.T) {
	// With goal and verdicts
	state := &phasedState{
		Goal:   "add user authentication",
		EpicID: "ag-5k2",
		Verdicts: map[string]string{
			"pre_mortem": "WARN",
		},
	}

	ctx := buildPhaseContext("", state, 4)
	if !containsStr(ctx, "Goal: add user authentication") {
		t.Errorf("context should contain goal, got: %q", ctx)
	}
	if !containsStr(ctx, "pre-mortem verdict: WARN") {
		t.Errorf("context should contain verdict, got: %q", ctx)
	}
	if !containsStr(ctx, "RPI Context") {
		t.Errorf("context should have header, got: %q", ctx)
	}

	// Empty state
	emptyState := &phasedState{Verdicts: make(map[string]string)}
	ctx = buildPhaseContext("", emptyState, 3)
	if ctx != "" {
		t.Errorf("empty state should produce empty context, got: %q", ctx)
	}
}

func TestBuildPromptForPhase_WithContext(t *testing.T) {
	state := &phasedState{
		Goal:   "add auth",
		EpicID: "ag-5k2",
		Verdicts: map[string]string{
			"pre_mortem": "PASS",
		},
	}

	// Phase 4 (crank) should include context and summary instruction
	prompt, err := buildPromptForPhase("", 4, state, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !containsStr(prompt, "/crank ag-5k2") {
		t.Errorf("crank prompt missing command, got: %q", prompt)
	}
	if !containsStr(prompt, "Goal: add auth") {
		t.Errorf("crank prompt missing goal context, got: %q", prompt)
	}
	if !containsStr(prompt, "pre-mortem verdict: PASS") {
		t.Errorf("crank prompt missing verdict context, got: %q", prompt)
	}
	if !containsStr(prompt, "phase-4-summary.md") {
		t.Errorf("crank prompt missing summary instruction, got: %q", prompt)
	}

	// Phase 1 (research) should NOT include cross-phase context but SHOULD have summary instruction
	prompt, err = buildPromptForPhase("", 1, state, nil)
	if err != nil {
		t.Fatal(err)
	}
	if containsStr(prompt, "RPI Context") {
		t.Errorf("research prompt should not have context block, got: %q", prompt)
	}
	if !containsStr(prompt, "phase-1-summary.md") {
		t.Errorf("research prompt should have summary instruction, got: %q", prompt)
	}
}

func TestGeneratePhaseSummary(t *testing.T) {
	state := &phasedState{
		Goal:     "add auth",
		EpicID:   "ag-5k2",
		FastPath: true,
		Verdicts: map[string]string{
			"pre_mortem": "WARN",
			"vibe":       "PASS",
		},
	}

	// Phase 1: research
	s := generatePhaseSummary(state, 1)
	if !containsStr(s, "add auth") {
		t.Errorf("research summary missing goal, got: %q", s)
	}

	// Phase 2: plan
	s = generatePhaseSummary(state, 2)
	if !containsStr(s, "ag-5k2") {
		t.Errorf("plan summary missing epic, got: %q", s)
	}
	if !containsStr(s, "fast path") {
		t.Errorf("plan summary missing fast path, got: %q", s)
	}

	// Phase 3: pre-mortem
	s = generatePhaseSummary(state, 3)
	if !containsStr(s, "WARN") {
		t.Errorf("pre-mortem summary missing verdict, got: %q", s)
	}

	// Phase 5: vibe
	s = generatePhaseSummary(state, 5)
	if !containsStr(s, "PASS") {
		t.Errorf("vibe summary missing verdict, got: %q", s)
	}
}

func TestReadPhaseSummaries(t *testing.T) {
	tmpDir := t.TempDir()
	rpiDir := filepath.Join(tmpDir, ".agents", "rpi")
	if err := os.MkdirAll(rpiDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write summaries for phases 1 and 2
	if err := os.WriteFile(filepath.Join(rpiDir, "phase-1-summary.md"), []byte("Research found X and Y"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rpiDir, "phase-2-summary.md"), []byte("Plan created epic ag-test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Reading for phase 3 should get both
	result := readPhaseSummaries(tmpDir, 3)
	if !containsStr(result, "Research found X and Y") {
		t.Errorf("should include phase 1 summary, got: %q", result)
	}
	if !containsStr(result, "Plan created epic ag-test") {
		t.Errorf("should include phase 2 summary, got: %q", result)
	}

	// Reading for phase 1 should get nothing (no prior phases)
	result = readPhaseSummaries(tmpDir, 1)
	if result != "" {
		t.Errorf("phase 1 should have no prior summaries, got: %q", result)
	}

	// Reading for phase 2 should get only phase 1
	result = readPhaseSummaries(tmpDir, 2)
	if !containsStr(result, "Research found X and Y") {
		t.Errorf("should include phase 1 summary, got: %q", result)
	}
	if containsStr(result, "Plan created") {
		t.Errorf("should NOT include phase 2 summary, got: %q", result)
	}
}

func TestWritePhaseSummary(t *testing.T) {
	tmpDir := t.TempDir()
	state := &phasedState{
		Goal:     "add auth",
		EpicID:   "ag-5k2",
		Verdicts: map[string]string{"pre_mortem": "PASS"},
	}

	// Fallback: no existing summary → writes mechanical one
	writePhaseSummary(tmpDir, state, 3)

	path := filepath.Join(tmpDir, ".agents", "rpi", "phase-3-summary.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("summary file not written: %v", err)
	}
	if !containsStr(string(data), "PASS") {
		t.Errorf("summary should contain verdict, got: %q", string(data))
	}

	// Claude-written summary exists → don't overwrite
	richSummary := "Research found JWT is best approach because stateless and fits API."
	if err := os.WriteFile(path, []byte(richSummary), 0644); err != nil {
		t.Fatal(err)
	}
	writePhaseSummary(tmpDir, state, 3) // should not overwrite
	data, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != richSummary {
		t.Errorf("should not overwrite Claude summary, got: %q", string(data))
	}
}

func TestCleanPhaseSummaries(t *testing.T) {
	tmpDir := t.TempDir()
	rpiDir := filepath.Join(tmpDir, ".agents", "rpi")
	if err := os.MkdirAll(rpiDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create summaries
	for i := 1; i <= 3; i++ {
		path := filepath.Join(rpiDir, fmt.Sprintf("phase-%d-summary.md", i))
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	cleanPhaseSummaries(rpiDir)

	for i := 1; i <= 6; i++ {
		path := filepath.Join(rpiDir, fmt.Sprintf("phase-%d-summary.md", i))
		if _, err := os.Stat(path); err == nil {
			t.Errorf("phase-%d-summary.md should be deleted", i)
		}
	}
}

func TestContextDisciplineInPrompt(t *testing.T) {
	state := &phasedState{
		Goal:     "test goal",
		EpicID:   "ag-test",
		Verdicts: map[string]string{},
		Attempts: make(map[string]int),
	}

	// Every phase should contain context discipline
	for phaseNum := 1; phaseNum <= 6; phaseNum++ {
		prompt, err := buildPromptForPhase("", phaseNum, state, nil)
		if err != nil {
			t.Fatalf("phase %d: unexpected error: %v", phaseNum, err)
		}
		if !containsStr(prompt, "CONTEXT DISCIPLINE") {
			t.Errorf("phase %d: prompt should contain CONTEXT DISCIPLINE", phaseNum)
		}
		if !containsStr(prompt, "PHASE SUMMARY CONTRACT") {
			t.Errorf("phase %d: prompt should contain PHASE SUMMARY CONTRACT", phaseNum)
		}
		if !containsStr(prompt, "handoff") {
			t.Errorf("phase %d: prompt should mention handoff", phaseNum)
		}
		if !containsStr(prompt, "BUDGET") {
			t.Errorf("phase %d: prompt should contain BUDGET guidance", phaseNum)
		}
	}
}

func TestContextDiscipline_PhaseSpecificBudgets(t *testing.T) {
	// Verify each phase has a specific budget
	for phaseNum := 1; phaseNum <= 6; phaseNum++ {
		budget, ok := phaseContextBudgets[phaseNum]
		if !ok {
			t.Errorf("phase %d: no context budget defined", phaseNum)
		}
		if budget == "" {
			t.Errorf("phase %d: context budget is empty", phaseNum)
		}
	}

	// Phase 4 (crank) should have CRITICAL warning
	if !containsStr(phaseContextBudgets[4], "CRITICAL") {
		t.Error("phase 4 budget should contain CRITICAL warning")
	}
}

func TestContextDiscipline_PromptOrdering(t *testing.T) {
	state := &phasedState{
		Goal:     "test goal",
		EpicID:   "ag-test",
		Verdicts: map[string]string{"pre_mortem": "PASS"},
		Attempts: make(map[string]int),
	}

	// Phase 4: check that discipline comes before skill invocation
	prompt, err := buildPromptForPhase("", 4, state, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	disciplineIdx := strings.Index(prompt, "CONTEXT DISCIPLINE")
	summaryIdx := strings.Index(prompt, "PHASE SUMMARY CONTRACT")
	// Use LastIndex for /crank since budget text also mentions it
	crankIdx := strings.LastIndex(prompt, "/crank")

	if disciplineIdx < 0 {
		t.Fatal("CONTEXT DISCIPLINE not found in prompt")
	}
	if summaryIdx < 0 {
		t.Fatal("PHASE SUMMARY CONTRACT not found in prompt")
	}
	if crankIdx < 0 {
		t.Fatal("/crank not found in prompt")
	}

	// Discipline should come first, then summary, then skill invocation (last /crank)
	if disciplineIdx >= summaryIdx {
		t.Errorf("discipline (%d) should come before summary (%d)", disciplineIdx, summaryIdx)
	}
	if summaryIdx >= crankIdx {
		t.Errorf("summary (%d) should come before skill invocation (%d)", summaryIdx, crankIdx)
	}
}

func TestHandoffDetection(t *testing.T) {
	tmpDir := t.TempDir()
	rpiDir := filepath.Join(tmpDir, ".agents", "rpi")
	if err := os.MkdirAll(rpiDir, 0755); err != nil {
		t.Fatal(err)
	}

	// No handoff file → not detected
	if handoffDetected(tmpDir, 4) {
		t.Error("should not detect handoff when file doesn't exist")
	}

	// Write handoff file → detected
	handoffPath := filepath.Join(rpiDir, "phase-4-handoff.md")
	if err := os.WriteFile(handoffPath, []byte("# Handoff\nContext degraded."), 0644); err != nil {
		t.Fatal(err)
	}

	if !handoffDetected(tmpDir, 4) {
		t.Error("should detect handoff when file exists")
	}

	// Different phase → not detected
	if handoffDetected(tmpDir, 3) {
		t.Error("should not detect handoff for different phase")
	}
}

func TestCleanPhaseSummaries_AlsoRemovesHandoffs(t *testing.T) {
	tmpDir := t.TempDir()
	rpiDir := filepath.Join(tmpDir, ".agents", "rpi")
	if err := os.MkdirAll(rpiDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create summaries and handoffs
	for i := 1; i <= 3; i++ {
		summaryPath := filepath.Join(rpiDir, fmt.Sprintf("phase-%d-summary.md", i))
		if err := os.WriteFile(summaryPath, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
		handoffPath := filepath.Join(rpiDir, fmt.Sprintf("phase-%d-handoff.md", i))
		if err := os.WriteFile(handoffPath, []byte("handoff"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	cleanPhaseSummaries(rpiDir)

	for i := 1; i <= 6; i++ {
		summaryPath := filepath.Join(rpiDir, fmt.Sprintf("phase-%d-summary.md", i))
		if _, err := os.Stat(summaryPath); err == nil {
			t.Errorf("phase-%d-summary.md should be deleted", i)
		}
		handoffPath := filepath.Join(rpiDir, fmt.Sprintf("phase-%d-handoff.md", i))
		if _, err := os.Stat(handoffPath); err == nil {
			t.Errorf("phase-%d-handoff.md should be deleted", i)
		}
	}
}

func TestPromptBudgetEstimate(t *testing.T) {
	state := &phasedState{
		Goal:     "test goal with a reasonable description of what needs to happen",
		EpicID:   "ag-test",
		Verdicts: map[string]string{"pre_mortem": "PASS", "vibe": "WARN"},
		Attempts: make(map[string]int),
	}

	// Every phase prompt should stay under 5000 chars (without summaries on disk)
	for phaseNum := 1; phaseNum <= 6; phaseNum++ {
		prompt, err := buildPromptForPhase("", phaseNum, state, nil)
		if err != nil {
			t.Fatalf("phase %d: unexpected error: %v", phaseNum, err)
		}
		if len(prompt) > 5000 {
			t.Errorf("phase %d: prompt is %d chars (max 5000 without summaries)", phaseNum, len(prompt))
		}
	}
}

// containsStr is a helper to check substring presence.
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
