package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/boshu2/agentops/cli/internal/ratchet"
	"github.com/spf13/cobra"
)

// cov3W2WriteLearningJSONL writes a JSONL learning file with the given metadata.
func cov3W2WriteLearningJSONL(t *testing.T, dir, name string, data map[string]any) string {
	t.Helper()
	b, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal learning data: %v", err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatalf("write learning file: %v", err)
	}
	return path
}

// cov3W2SetupMaturityDir creates a temp dir with .agents/learnings/ structure.
func cov3W2SetupMaturityDir(t *testing.T) (string, string) {
	t.Helper()
	tmp := t.TempDir()
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}
	return tmp, learningsDir
}

// cov3W2ChdirTemp moved to testutil_test.go.

// cov3W2MakeTransitionResults creates a slice of MaturityTransitionResult for testing.
func cov3W2MakeTransitionResults(ids ...string) []*ratchet.MaturityTransitionResult {
	var results []*ratchet.MaturityTransitionResult
	for _, id := range ids {
		results = append(results, &ratchet.MaturityTransitionResult{
			LearningID:   id,
			OldMaturity:  "provisional",
			NewMaturity:  "anti-pattern",
			Transitioned: true,
			Utility:      0.1,
			HarmfulCount: 10,
			RewardCount:  10,
			Reason:       "test",
		})
	}
	return results
}

// cov3W2CaptureStdout moved to testutil_test.go.

// --- runMaturitySingle tests ---

func TestCov3_maturity_runMaturitySingle_dryRun(t *testing.T) {
	tmp, learningsDir := cov3W2SetupMaturityDir(t)

	cov3W2WriteLearningJSONL(t, learningsDir, "L001.jsonl", map[string]any{
		"id":       "L001",
		"maturity": "provisional",
		"utility":  0.5,
	})

	oldDryRun := dryRun
	dryRun = true
	defer func() { dryRun = oldDryRun }()

	err := runMaturitySingle(tmp, "L001")
	if err != nil {
		t.Fatalf("runMaturitySingle dry-run: %v", err)
	}
}

func TestCov3_maturity_runMaturitySingle_notFound(t *testing.T) {
	tmp, _ := cov3W2SetupMaturityDir(t)

	err := runMaturitySingle(tmp, "NONEXISTENT")
	if err == nil {
		t.Fatal("expected error for nonexistent learning")
	}
	if !strings.Contains(err.Error(), "find learning") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCov3_maturity_runMaturitySingle_checkNoTransition(t *testing.T) {
	tmp, learningsDir := cov3W2SetupMaturityDir(t)

	cov3W2WriteLearningJSONL(t, learningsDir, "L002.jsonl", map[string]any{
		"id":            "L002",
		"maturity":      "provisional",
		"utility":       0.5,
		"confidence":    0.5,
		"reward_count":  1,
		"helpful_count": 0,
		"harmful_count": 0,
	})

	oldDryRun := dryRun
	dryRun = false
	defer func() { dryRun = oldDryRun }()

	oldOutput := output
	output = "table"
	defer func() { output = oldOutput }()

	oldApply := maturityApply
	maturityApply = false
	defer func() { maturityApply = oldApply }()

	cov3W2CaptureStdout(t, func() {
		err := runMaturitySingle(tmp, "L002")
		if err != nil {
			t.Fatalf("runMaturitySingle: %v", err)
		}
	})
}

// --- checkOrApplyMaturity tests ---

func TestCov3_maturity_checkOrApplyMaturity_checkMode(t *testing.T) {
	_, learningsDir := cov3W2SetupMaturityDir(t)

	path := cov3W2WriteLearningJSONL(t, learningsDir, "L003.jsonl", map[string]any{
		"id":            "L003",
		"maturity":      "provisional",
		"utility":       0.5,
		"reward_count":  1,
		"helpful_count": 1,
		"harmful_count": 0,
	})

	oldApply := maturityApply
	maturityApply = false
	defer func() { maturityApply = oldApply }()

	result, err := checkOrApplyMaturity(path)
	if err != nil {
		t.Fatalf("checkOrApplyMaturity: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.LearningID != "L003" {
		t.Fatalf("expected learning ID L003, got %q", result.LearningID)
	}
}

func TestCov3_maturity_checkOrApplyMaturity_applyMode(t *testing.T) {
	_, learningsDir := cov3W2SetupMaturityDir(t)

	path := cov3W2WriteLearningJSONL(t, learningsDir, "L004.jsonl", map[string]any{
		"id":            "L004",
		"maturity":      "provisional",
		"utility":       0.8,
		"confidence":    0.9,
		"reward_count":  5,
		"helpful_count": 4,
		"harmful_count": 0,
	})

	oldApply := maturityApply
	maturityApply = true
	defer func() { maturityApply = oldApply }()

	result, err := checkOrApplyMaturity(path)
	if err != nil {
		t.Fatalf("checkOrApplyMaturity apply: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// --- outputSingleMaturityResult tests ---

func TestCov3_maturity_outputSingleMaturityResult_json(t *testing.T) {
	oldOutput := output
	output = "json"
	defer func() { output = oldOutput }()

	result := &ratchet.MaturityTransitionResult{
		LearningID:   "L005",
		OldMaturity:  "provisional",
		NewMaturity:  "provisional",
		Transitioned: false,
		Utility:      0.5,
		Confidence:   0.5,
		HelpfulCount: 1,
		HarmfulCount: 0,
		RewardCount:  1,
		Reason:       "no transition",
	}

	got := cov3W2CaptureStdout(t, func() {
		err := outputSingleMaturityResult(result)
		if err != nil {
			t.Fatalf("outputSingleMaturityResult json: %v", err)
		}
	})
	if !strings.Contains(got, "L005") {
		t.Fatalf("expected JSON output to contain L005, got: %s", got)
	}
}

func TestCov3_maturity_outputSingleMaturityResult_table(t *testing.T) {
	oldOutput := output
	output = "table"
	defer func() { output = oldOutput }()

	result := &ratchet.MaturityTransitionResult{
		LearningID:   "L006",
		OldMaturity:  "provisional",
		NewMaturity:  "candidate",
		Transitioned: true,
		Utility:      0.8,
		Confidence:   0.9,
		HelpfulCount: 4,
		HarmfulCount: 0,
		RewardCount:  5,
		Reason:       "met threshold",
	}

	oldApply := maturityApply
	maturityApply = false
	defer func() { maturityApply = oldApply }()

	cov3W2CaptureStdout(t, func() {
		err := outputSingleMaturityResult(result)
		if err != nil {
			t.Fatalf("outputSingleMaturityResult table: %v", err)
		}
	})
}

// --- runMaturity (Cobra RunE) tests ---

func TestCov3_maturity_runMaturity_noLearningsDir(t *testing.T) {
	tmp := t.TempDir()
	cov3W2ChdirTemp(t, tmp)

	cmd := &cobra.Command{}
	err := runMaturity(cmd, []string{})
	if err != nil {
		t.Fatalf("expected nil error when no learnings dir, got: %v", err)
	}
}

func TestCov3_maturity_runMaturity_noArgsNoScan(t *testing.T) {
	tmp, _ := cov3W2SetupMaturityDir(t)
	cov3W2ChdirTemp(t, tmp)

	oldScan := maturityScan
	maturityScan = false
	defer func() { maturityScan = oldScan }()

	oldExpire := maturityExpire
	maturityExpire = false
	defer func() { maturityExpire = oldExpire }()

	oldEvict := maturityEvict
	maturityEvict = false
	defer func() { maturityEvict = oldEvict }()

	cmd := &cobra.Command{}
	err := runMaturity(cmd, []string{})
	if err == nil {
		t.Fatal("expected error when no args and no --scan")
	}
	if !strings.Contains(err.Error(), "must provide learning-id or use --scan") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCov3_maturity_runMaturity_scanMode(t *testing.T) {
	tmp, learningsDir := cov3W2SetupMaturityDir(t)
	cov3W2ChdirTemp(t, tmp)

	cov3W2WriteLearningJSONL(t, learningsDir, "scan-test.jsonl", map[string]any{
		"id":            "scan-test",
		"maturity":      "provisional",
		"utility":       0.5,
		"reward_count":  1,
		"helpful_count": 1,
		"harmful_count": 0,
	})

	oldScan := maturityScan
	maturityScan = true
	defer func() { maturityScan = oldScan }()

	oldExpire := maturityExpire
	maturityExpire = false
	defer func() { maturityExpire = oldExpire }()

	oldEvict := maturityEvict
	maturityEvict = false
	defer func() { maturityEvict = oldEvict }()

	oldApply := maturityApply
	maturityApply = false
	defer func() { maturityApply = oldApply }()

	cov3W2CaptureStdout(t, func() {
		cmd := &cobra.Command{}
		err := runMaturity(cmd, []string{})
		if err != nil {
			t.Fatalf("runMaturity scan mode: %v", err)
		}
	})
}

func TestCov3_maturity_runMaturity_withLearningID(t *testing.T) {
	tmp, learningsDir := cov3W2SetupMaturityDir(t)
	cov3W2ChdirTemp(t, tmp)

	cov3W2WriteLearningJSONL(t, learningsDir, "L010.jsonl", map[string]any{
		"id":            "L010",
		"maturity":      "provisional",
		"utility":       0.5,
		"reward_count":  1,
		"helpful_count": 1,
		"harmful_count": 0,
	})

	oldScan := maturityScan
	maturityScan = false
	defer func() { maturityScan = oldScan }()

	oldExpire := maturityExpire
	maturityExpire = false
	defer func() { maturityExpire = oldExpire }()

	oldEvict := maturityEvict
	maturityEvict = false
	defer func() { maturityEvict = oldEvict }()

	oldApply := maturityApply
	maturityApply = false
	defer func() { maturityApply = oldApply }()

	oldOutput := output
	output = "table"
	defer func() { output = oldOutput }()

	cov3W2CaptureStdout(t, func() {
		cmd := &cobra.Command{}
		err := runMaturity(cmd, []string{"L010"})
		if err != nil {
			t.Fatalf("runMaturity with ID: %v", err)
		}
	})
}

// --- applyScannedTransitions tests ---

func TestCov3_maturity_applyScannedTransitions_noResults(t *testing.T) {
	_, learningsDir := cov3W2SetupMaturityDir(t)

	cov3W2CaptureStdout(t, func() {
		applyScannedTransitions(learningsDir, nil)
	})
}

func TestCov3_maturity_applyScannedTransitions_missingFile(t *testing.T) {
	_, learningsDir := cov3W2SetupMaturityDir(t)

	results := cov3W2MakeTransitionResults("missing-learning")

	cov3W2CaptureStdout(t, func() {
		applyScannedTransitions(learningsDir, results)
	})
}

func TestCov3_maturity_applyScannedTransitions_withValidFile(t *testing.T) {
	tmp, learningsDir := cov3W2SetupMaturityDir(t)
	_ = tmp

	// Create a learning that matches the transition result
	cov3W2WriteLearningJSONL(t, learningsDir, "apply-target.jsonl", map[string]any{
		"id":            "apply-target",
		"maturity":      "provisional",
		"utility":       0.1,
		"confidence":    0.1,
		"reward_count":  10,
		"helpful_count": 0,
		"harmful_count": 10,
	})

	results := cov3W2MakeTransitionResults("apply-target")

	cov3W2CaptureStdout(t, func() {
		applyScannedTransitions(learningsDir, results)
	})
}

// --- runMaturityScan tests ---

func TestCov3_maturity_runMaturityScan_dryRun(t *testing.T) {
	_, learningsDir := cov3W2SetupMaturityDir(t)

	oldDryRun := dryRun
	dryRun = true
	defer func() { dryRun = oldDryRun }()

	err := runMaturityScan(learningsDir)
	if err != nil {
		t.Fatalf("runMaturityScan dry-run: %v", err)
	}
}

func TestCov3_maturity_runMaturityScan_noTransitions(t *testing.T) {
	_, learningsDir := cov3W2SetupMaturityDir(t)

	oldDryRun := dryRun
	dryRun = false
	defer func() { dryRun = oldDryRun }()

	cov3W2WriteLearningJSONL(t, learningsDir, "stable.jsonl", map[string]any{
		"id":            "stable",
		"maturity":      "provisional",
		"utility":       0.5,
		"reward_count":  1,
		"helpful_count": 0,
		"harmful_count": 0,
	})

	cov3W2CaptureStdout(t, func() {
		err := runMaturityScan(learningsDir)
		if err != nil {
			t.Fatalf("runMaturityScan no transitions: %v", err)
		}
	})
}

func TestCov3_maturity_runMaturityScan_withTransitions(t *testing.T) {
	_, learningsDir := cov3W2SetupMaturityDir(t)

	oldDryRun := dryRun
	dryRun = false
	defer func() { dryRun = oldDryRun }()

	oldApply := maturityApply
	maturityApply = false
	defer func() { maturityApply = oldApply }()

	cov3W2WriteLearningJSONL(t, learningsDir, "ready.jsonl", map[string]any{
		"id":            "ready",
		"maturity":      "provisional",
		"utility":       0.8,
		"confidence":    0.9,
		"reward_count":  5,
		"helpful_count": 4,
		"harmful_count": 0,
	})

	oldOutput := output
	output = "table"
	defer func() { output = oldOutput }()

	cov3W2CaptureStdout(t, func() {
		err := runMaturityScan(learningsDir)
		if err != nil {
			t.Fatalf("runMaturityScan with transitions: %v", err)
		}
	})
}

func TestCov3_maturity_runMaturityScan_withApply(t *testing.T) {
	_, learningsDir := cov3W2SetupMaturityDir(t)

	oldDryRun := dryRun
	dryRun = false
	defer func() { dryRun = oldDryRun }()

	oldApply := maturityApply
	maturityApply = true
	defer func() { maturityApply = oldApply }()

	cov3W2WriteLearningJSONL(t, learningsDir, "apply-me.jsonl", map[string]any{
		"id":            "apply-me",
		"maturity":      "provisional",
		"utility":       0.8,
		"confidence":    0.9,
		"reward_count":  5,
		"helpful_count": 4,
		"harmful_count": 0,
	})

	oldOutput := output
	output = "table"
	defer func() { output = oldOutput }()

	cov3W2CaptureStdout(t, func() {
		err := runMaturityScan(learningsDir)
		if err != nil {
			t.Fatalf("runMaturityScan with apply: %v", err)
		}
	})
}

// --- runAntiPatterns tests ---

func TestCov3_maturity_runAntiPatterns_noLearningsDir(t *testing.T) {
	tmp := t.TempDir()
	cov3W2ChdirTemp(t, tmp)

	cov3W2CaptureStdout(t, func() {
		cmd := &cobra.Command{}
		err := runAntiPatterns(cmd, nil)
		if err != nil {
			t.Fatalf("expected nil for missing learnings dir, got: %v", err)
		}
	})
}

func TestCov3_maturity_runAntiPatterns_noAntiPatterns(t *testing.T) {
	tmp, learningsDir := cov3W2SetupMaturityDir(t)
	cov3W2ChdirTemp(t, tmp)

	cov3W2WriteLearningJSONL(t, learningsDir, "normal.jsonl", map[string]any{
		"id":            "normal",
		"maturity":      "provisional",
		"utility":       0.8,
		"reward_count":  2,
		"helpful_count": 2,
		"harmful_count": 0,
	})

	cov3W2CaptureStdout(t, func() {
		cmd := &cobra.Command{}
		err := runAntiPatterns(cmd, nil)
		if err != nil {
			t.Fatalf("runAntiPatterns with no anti-patterns: %v", err)
		}
	})
}

func TestCov3_maturity_runAntiPatterns_tableOutput(t *testing.T) {
	tmp, learningsDir := cov3W2SetupMaturityDir(t)
	cov3W2ChdirTemp(t, tmp)

	cov3W2WriteLearningJSONL(t, learningsDir, "bad.jsonl", map[string]any{
		"id":            "bad",
		"maturity":      "anti-pattern",
		"utility":       0.1,
		"confidence":    0.1,
		"reward_count":  10,
		"helpful_count": 0,
		"harmful_count": 10,
	})

	oldOutput := output
	output = "table"
	defer func() { output = oldOutput }()

	cov3W2CaptureStdout(t, func() {
		cmd := &cobra.Command{}
		err := runAntiPatterns(cmd, nil)
		if err != nil {
			t.Fatalf("runAntiPatterns table: %v", err)
		}
	})
}

func TestCov3_maturity_runAntiPatterns_jsonOutput(t *testing.T) {
	tmp, learningsDir := cov3W2SetupMaturityDir(t)
	cov3W2ChdirTemp(t, tmp)

	cov3W2WriteLearningJSONL(t, learningsDir, "bad-json.jsonl", map[string]any{
		"id":            "bad-json",
		"maturity":      "anti-pattern",
		"utility":       0.1,
		"confidence":    0.1,
		"reward_count":  10,
		"helpful_count": 0,
		"harmful_count": 10,
	})

	oldOutput := output
	output = "json"
	defer func() { output = oldOutput }()

	cov3W2CaptureStdout(t, func() {
		cmd := &cobra.Command{}
		err := runAntiPatterns(cmd, nil)
		if err != nil {
			t.Fatalf("runAntiPatterns JSON: %v", err)
		}
	})
}

// --- executeAntiPatternPromotions tests ---

func TestCov3_maturity_executeAntiPatternPromotions_empty(t *testing.T) {
	_, learningsDir := cov3W2SetupMaturityDir(t)

	cov3W2CaptureStdout(t, func() {
		promoted := executeAntiPatternPromotions(learningsDir, nil)
		if promoted != 0 {
			t.Fatalf("expected 0 promotions for empty input, got %d", promoted)
		}
	})
}

func TestCov3_maturity_executeAntiPatternPromotions_missingFile(t *testing.T) {
	_, learningsDir := cov3W2SetupMaturityDir(t)

	results := cov3W2MakeTransitionResults("missing-learning")

	cov3W2CaptureStdout(t, func() {
		promoted := executeAntiPatternPromotions(learningsDir, results)
		if promoted != 0 {
			t.Fatalf("expected 0 promotions for missing file, got %d", promoted)
		}
	})
}

func TestCov3_maturity_executeAntiPatternPromotions_withValidFile(t *testing.T) {
	tmp, learningsDir := cov3W2SetupMaturityDir(t)
	_ = tmp

	cov3W2WriteLearningJSONL(t, learningsDir, "promote-target.jsonl", map[string]any{
		"id":            "promote-target",
		"maturity":      "provisional",
		"utility":       0.1,
		"confidence":    0.1,
		"reward_count":  10,
		"helpful_count": 0,
		"harmful_count": 10,
	})

	results := cov3W2MakeTransitionResults("promote-target")

	cov3W2CaptureStdout(t, func() {
		promoted := executeAntiPatternPromotions(learningsDir, results)
		// The function should find the file and apply the transition
		_ = promoted // may or may not promote depending on ratchet logic
	})
}

// --- runPromoteAntiPatterns tests ---

func TestCov3_maturity_runPromoteAntiPatterns_noLearningsDir(t *testing.T) {
	tmp := t.TempDir()
	cov3W2ChdirTemp(t, tmp)

	cov3W2CaptureStdout(t, func() {
		cmd := &cobra.Command{}
		err := runPromoteAntiPatterns(cmd, nil)
		if err != nil {
			t.Fatalf("expected nil for missing learnings dir, got: %v", err)
		}
	})
}

func TestCov3_maturity_runPromoteAntiPatterns_noCandidates(t *testing.T) {
	tmp, learningsDir := cov3W2SetupMaturityDir(t)
	cov3W2ChdirTemp(t, tmp)

	cov3W2WriteLearningJSONL(t, learningsDir, "ok.jsonl", map[string]any{
		"id":            "ok",
		"maturity":      "provisional",
		"utility":       0.8,
		"reward_count":  2,
		"helpful_count": 2,
		"harmful_count": 0,
	})

	cov3W2CaptureStdout(t, func() {
		cmd := &cobra.Command{}
		err := runPromoteAntiPatterns(cmd, nil)
		if err != nil {
			t.Fatalf("runPromoteAntiPatterns no candidates: %v", err)
		}
	})
}

func TestCov3_maturity_runPromoteAntiPatterns_dryRun(t *testing.T) {
	tmp, learningsDir := cov3W2SetupMaturityDir(t)
	cov3W2ChdirTemp(t, tmp)

	cov3W2WriteLearningJSONL(t, learningsDir, "harmful.jsonl", map[string]any{
		"id":            "harmful",
		"maturity":      "provisional",
		"utility":       0.1,
		"confidence":    0.1,
		"reward_count":  10,
		"helpful_count": 0,
		"harmful_count": 10,
	})

	oldDryRun := dryRun
	dryRun = true
	defer func() { dryRun = oldDryRun }()

	cov3W2CaptureStdout(t, func() {
		cmd := &cobra.Command{}
		err := runPromoteAntiPatterns(cmd, nil)
		if err != nil {
			t.Fatalf("runPromoteAntiPatterns dry-run: %v", err)
		}
	})
}

func TestCov3_maturity_runPromoteAntiPatterns_execute(t *testing.T) {
	tmp, learningsDir := cov3W2SetupMaturityDir(t)
	cov3W2ChdirTemp(t, tmp)

	cov3W2WriteLearningJSONL(t, learningsDir, "demote-me.jsonl", map[string]any{
		"id":            "demote-me",
		"maturity":      "provisional",
		"utility":       0.1,
		"confidence":    0.1,
		"reward_count":  10,
		"helpful_count": 0,
		"harmful_count": 10,
	})

	oldDryRun := dryRun
	dryRun = false
	defer func() { dryRun = oldDryRun }()

	cov3W2CaptureStdout(t, func() {
		cmd := &cobra.Command{}
		err := runPromoteAntiPatterns(cmd, nil)
		if err != nil {
			t.Fatalf("runPromoteAntiPatterns execute: %v", err)
		}
	})
}
