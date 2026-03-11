package vault

import (
	"os"
	"path/filepath"
	"testing"
)

// TestExtra_DetectVault_EmptyStartDir covers the empty-string branch
// where os.Getwd is called internally.
func TestExtra_DetectVault_EmptyStartDir(t *testing.T) {
	// With empty string, DetectVault uses cwd. Should not panic.
	result := DetectVault("")
	// We just verify it returns without error; result depends on cwd.
	_ = result
}

// TestExtra_DetectVault_NoVaultFound covers walking to root without finding .obsidian.
func TestExtra_DetectVault_NoVaultFound(t *testing.T) {
	tmp := t.TempDir()
	result := DetectVault(tmp)
	if result != "" {
		t.Errorf("expected empty string for dir with no vault, got %q", result)
	}
}

// TestExtra_DetectVault_FoundVault covers successful vault detection.
func TestExtra_DetectVault_FoundVault(t *testing.T) {
	tmp := t.TempDir()
	obsDir := filepath.Join(tmp, ".obsidian")
	if err := os.MkdirAll(obsDir, 0700); err != nil {
		t.Fatal(err)
	}
	// Start from a subdirectory.
	sub := filepath.Join(tmp, "notes", "daily")
	if err := os.MkdirAll(sub, 0700); err != nil {
		t.Fatal(err)
	}

	result := DetectVault(sub)
	if result != tmp {
		t.Errorf("expected %q, got %q", tmp, result)
	}
}

// TestExtra_DetectVault_GetWdError covers the Getwd error branch.
// We simulate by changing to a removed directory.
func TestExtra_DetectVault_GetWdError(t *testing.T) {
	tmp := t.TempDir()
	sub := filepath.Join(tmp, "removeme")
	if err := os.MkdirAll(sub, 0700); err != nil {
		t.Fatal(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(sub); err != nil {
		t.Fatal(err)
	}
	// Remove the directory while we're in it.
	if err := os.RemoveAll(sub); err != nil {
		t.Fatal(err)
	}

	result := DetectVault("")
	// On macOS/Linux, Getwd on removed dir may still work, so we just
	// verify it doesn't panic and returns a string.
	_ = result
}
