package main

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// forge.go — runForgeTranscript dry-run / no-files (26.9% → higher)
// ---------------------------------------------------------------------------

func TestCov10_runForgeTranscript_dryRun(t *testing.T) {
	origDryRun := dryRun
	origQuiet := forgeQuiet
	origLastSession := forgeLastSession
	defer func() {
		dryRun = origDryRun
		forgeQuiet = origQuiet
		forgeLastSession = origLastSession
	}()
	dryRun = true
	forgeQuiet = false
	forgeLastSession = false

	cmd := &cobra.Command{}
	var buf strings.Builder
	cmd.SetOut(&buf)

	// Empty args → files=[], but dry-run fires before len==0 check
	err := runForgeTranscript(cmd, []string{})
	if err != nil {
		t.Fatalf("runForgeTranscript dry-run: %v", err)
	}
	if !strings.Contains(buf.String(), "dry-run") {
		t.Errorf("expected dry-run in output, got %q", buf.String())
	}
}

func TestCov10_runForgeTranscript_noFiles(t *testing.T) {
	origDryRun := dryRun
	origQuiet := forgeQuiet
	origLastSession := forgeLastSession
	defer func() {
		dryRun = origDryRun
		forgeQuiet = origQuiet
		forgeLastSession = origLastSession
	}()
	dryRun = false
	forgeQuiet = false
	forgeLastSession = false

	cmd := &cobra.Command{}

	// Pass a pattern that won't match anything → len(files)==0 → error
	err := runForgeTranscript(cmd, []string{"zzz_no_match_ever_xyz_*.impossible"})
	if err == nil {
		t.Fatal("expected error for no matching files")
	}
	if !strings.Contains(err.Error(), "no files found") {
		t.Errorf("expected 'no files found' error, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// forge.go — runForgeMarkdown dry-run / no-files (29.2% → higher)
// ---------------------------------------------------------------------------

func TestCov10_runForgeMarkdown_dryRun(t *testing.T) {
	origDryRun := dryRun
	origQuiet := forgeMdQuiet
	defer func() {
		dryRun = origDryRun
		forgeMdQuiet = origQuiet
	}()
	dryRun = true
	forgeMdQuiet = false

	cmd := &cobra.Command{}
	var buf strings.Builder
	cmd.SetOut(&buf)

	// Empty args → files=[], dry-run fires before len==0 check
	err := runForgeMarkdown(cmd, []string{})
	if err != nil {
		t.Fatalf("runForgeMarkdown dry-run: %v", err)
	}
	if !strings.Contains(buf.String(), "dry-run") {
		t.Errorf("expected dry-run in output, got %q", buf.String())
	}
}

func TestCov10_runForgeMarkdown_noFiles(t *testing.T) {
	origDryRun := dryRun
	origQuiet := forgeMdQuiet
	defer func() {
		dryRun = origDryRun
		forgeMdQuiet = origQuiet
	}()
	dryRun = false
	forgeMdQuiet = false

	cmd := &cobra.Command{}

	err := runForgeMarkdown(cmd, []string{"zzz_no_match_ever_xyz_*.impossible"})
	if err == nil {
		t.Fatal("expected error for no markdown files")
	}
	if !strings.Contains(err.Error(), "no markdown files found") {
		t.Errorf("expected 'no markdown files found' error, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// hooks.go — runHooksShow (28% → higher)
// Reads real ~/.claude/settings.json; succeed or return nil on missing file
// ---------------------------------------------------------------------------

func TestCov10_runHooksShow_liveSettings(t *testing.T) {
	cmd := &cobra.Command{}
	err := runHooksShow(cmd, nil)
	// Either succeeds (settings.json exists) or returns nil (hooksMap nil)
	if err != nil {
		t.Fatalf("runHooksShow unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// session_close.go — runSessionClose dry-run (33.3% → higher)
// Finds most recent transcript from ~/.claude/projects; dry-run returns nil
// ---------------------------------------------------------------------------

func TestCov10_runSessionClose_dryRun(t *testing.T) {
	origDryRun := dryRun
	origSessionID := sessionCloseSessionID
	defer func() {
		dryRun = origDryRun
		sessionCloseSessionID = origSessionID
	}()
	dryRun = true
	sessionCloseSessionID = "" // use most-recent transcript fallback

	cmd := &cobra.Command{}
	err := runSessionClose(cmd, nil)
	// Should succeed (finds real transcript) or fail gracefully
	if err != nil {
		// Acceptable failure: no transcript found in test env
		if !strings.Contains(err.Error(), "transcript") &&
			!strings.Contains(err.Error(), "no recent") &&
			!strings.Contains(err.Error(), "find transcript") {
			t.Fatalf("runSessionClose dry-run unexpected error: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// forge.go — noFilesError helper (covers quiet=true branch)
// ---------------------------------------------------------------------------

func TestCov10_noFilesError_quiet(t *testing.T) {
	err := noFilesError(true, "some message")
	if err != nil {
		t.Errorf("noFilesError(quiet=true) should return nil, got %v", err)
	}
}

func TestCov10_noFilesError_notQuiet(t *testing.T) {
	err := noFilesError(false, "some message")
	if err == nil {
		t.Error("noFilesError(quiet=false) should return error")
	}
}
