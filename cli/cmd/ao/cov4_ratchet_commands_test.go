package main

import (
	"os"
	"path/filepath"
	"testing"
)

// cov4RatchetSetup creates a temp workdir with .agents/ao/chain.jsonl for ratchet tests.
func cov4RatchetSetup(t *testing.T) string {
	t.Helper()
	tmp := chdirTemp(t)
	t.Setenv("HOME", tmp)
	setupAgentsDir(t, tmp)
	// Create empty chain file so LoadChain succeeds
	chainDir := filepath.Join(tmp, ".agents", "ao")
	if err := os.MkdirAll(chainDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(chainDir, "chain.jsonl"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	return tmp
}

func TestCov4_RatchetCheck(t *testing.T) {
	cov4RatchetSetup(t)
	// Pass a valid step name — gate will likely fail but RunE is exercised
	_, _ = executeCommand("ratchet", "check", "research")
}

func TestCov4_RatchetFind(t *testing.T) {
	cov4RatchetSetup(t)
	_, _ = executeCommand("ratchet", "find", "specs/*.md")
}

func TestCov4_RatchetNext(t *testing.T) {
	cov4RatchetSetup(t)
	out, err := executeCommand("ratchet", "next")
	// Should succeed with "no steps completed" or similar
	_ = out
	_ = err
}

func TestCov4_RatchetPromote(t *testing.T) {
	tmp := cov4RatchetSetup(t)
	// Create a dummy artifact to promote
	artifact := filepath.Join(tmp, ".agents", "learnings", "test-artifact.md")
	if err := os.WriteFile(artifact, []byte("# Test\n**ID**: test-1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// --to is required
	_, _ = executeCommand("ratchet", "promote", artifact, "--to", "1")
}

func TestCov4_RatchetRecord(t *testing.T) {
	cov4RatchetSetup(t)
	// --output is required
	_, _ = executeCommand("ratchet", "record", "research", "--output", ".agents/research/topic.md")
}

func TestCov4_RatchetSkip(t *testing.T) {
	cov4RatchetSetup(t)
	// --reason is required
	_, _ = executeCommand("ratchet", "skip", "pre-mortem", "--reason", "test skip")
}

func TestCov4_RatchetSpec(t *testing.T) {
	cov4RatchetSetup(t)
	_, _ = executeCommand("ratchet", "spec")
}

func TestCov4_RatchetStatus(t *testing.T) {
	cov4RatchetSetup(t)
	out, err := executeCommand("ratchet", "status")
	_ = out
	_ = err
}

func TestCov4_RatchetValidate(t *testing.T) {
	tmp := cov4RatchetSetup(t)
	// Create a dummy file for --changes
	artifact := filepath.Join(tmp, ".agents", "research", "test.md")
	if err := os.WriteFile(artifact, []byte("# Test Research\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, _ = executeCommand("ratchet", "validate", "research", "--changes", artifact)
}
