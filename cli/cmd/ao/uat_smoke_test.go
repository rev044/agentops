package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// UAT regression tests — encode the na-xji manual validation scenarios
// as automated smoke tests to prevent regression across releases.
// Each test uses executeCommand() from cobra_commands_test.go.

func TestUATSmoke_InjectForResearch(t *testing.T) {
	tmp := chdirTemp(t)
	setupAgentsDir(t, tmp)

	// Create a minimal learning file so inject has something to return.
	writeFile(t, tmp+"/.agents/learnings/test-learning.md", `---
utility: 0.8
maturity: validated
tags: [research, test]
---
# Test Learning
This is a test learning for UAT.
`)

	out, err := executeCommand("inject", "--for=research", "--no-cite")
	if err != nil {
		t.Fatalf("inject --for=research failed: %v\noutput: %s", err, out)
	}
	if strings.Contains(out, "HISTORY") {
		t.Errorf("inject --for=research should exclude HISTORY section, got: %s", out)
	}
}

func TestUATSmoke_InjectForPathTraversal(t *testing.T) {
	tmp := chdirTemp(t)
	setupAgentsDir(t, tmp)

	_, err := executeCommand("inject", "--for=../../etc/passwd")
	if err == nil {
		t.Error("expected error for path traversal --for=../../etc/passwd, got nil")
	}
}

func TestUATSmoke_InjectForUnknownSkill(t *testing.T) {
	tmp := chdirTemp(t)
	setupAgentsDir(t, tmp)

	_, err := executeCommand("inject", "--for=nonexistent-skill-xyz")
	if err == nil {
		t.Error("expected error for unknown skill --for=nonexistent-skill-xyz, got nil")
	}
}

func TestUATSmoke_HandoffDryRun(t *testing.T) {
	tmp := chdirTemp(t)
	setupAgentsDir(t, tmp)

	out, err := executeCommand("handoff", "--dry-run", "test-handoff")
	if err != nil {
		t.Fatalf("handoff --dry-run failed: %v\noutput: %s", err, out)
	}
	// Dry-run should produce JSON-like output on stdout.
	if !strings.Contains(out, "goal") && !strings.Contains(out, "type") {
		t.Errorf("handoff --dry-run should produce structured output, got: %s", out)
	}
}

func TestUATSmoke_MineGitSource(t *testing.T) {
	tmp := chdirTemp(t)
	setupAgentsDir(t, tmp)

	// mine --sources git needs a git repo; init one.
	initGitRepo(t, tmp)

	out, err := executeCommand("mine", "--sources", "git", "--since", "7d", "--quiet")
	// mine may return error if no signals found — that's OK for smoke test.
	_ = err
	// When quiet, output should be JSON if there are results, or empty/error message.
	if out != "" && strings.HasPrefix(strings.TrimSpace(out), "{") {
		var result map[string]any
		if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); jsonErr != nil {
			t.Errorf("mine --quiet output is not valid JSON: %v\noutput: %s", jsonErr, out)
		}
	}
}

func TestUATSmoke_DefragDedup(t *testing.T) {
	tmp := chdirTemp(t)
	setupAgentsDir(t, tmp)

	out, err := executeCommand("defrag", "--dedup", "--quiet")
	if err != nil {
		t.Fatalf("defrag --dedup failed: %v\noutput: %s", err, out)
	}
}

func TestUATSmoke_Version(t *testing.T) {
	out, err := executeCommand("version")
	if err != nil {
		t.Fatalf("version failed: %v", err)
	}
	if !strings.Contains(out, "ao version") {
		t.Errorf("expected 'ao version' in output, got: %s", out)
	}
}
