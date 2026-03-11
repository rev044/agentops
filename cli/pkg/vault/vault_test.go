package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectVault(t *testing.T) {
	// Create temp dir structure
	tmpDir := t.TempDir()

	// No vault case
	if got := DetectVault(tmpDir); got != "" {
		t.Errorf("DetectVault() = %q, want empty string", got)
	}

	// Create .obsidian directory to simulate vault
	vaultDir := filepath.Join(tmpDir, "my-vault")
	obsidianDir := filepath.Join(vaultDir, ".obsidian")
	if err := os.MkdirAll(obsidianDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Detect from vault root
	if got := DetectVault(vaultDir); got != vaultDir {
		t.Errorf("DetectVault(%q) = %q, want %q", vaultDir, got, vaultDir)
	}

	// Detect from subdirectory
	subDir := filepath.Join(vaultDir, "notes", "daily")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if got := DetectVault(subDir); got != vaultDir {
		t.Errorf("DetectVault(%q) = %q, want %q", subDir, got, vaultDir)
	}
}

func TestHasSmartConnections(t *testing.T) {
	tmpDir := t.TempDir()

	// No vault
	if HasSmartConnections(tmpDir) {
		t.Error("HasSmartConnections() = true, want false for non-vault")
	}

	// Empty string
	if HasSmartConnections("") {
		t.Error("HasSmartConnections(\"\") = true, want false")
	}

	// Vault without SC
	vaultDir := filepath.Join(tmpDir, "vault")
	obsidianDir := filepath.Join(vaultDir, ".obsidian")
	if err := os.MkdirAll(obsidianDir, 0755); err != nil {
		t.Fatal(err)
	}
	if HasSmartConnections(vaultDir) {
		t.Error("HasSmartConnections() = true, want false without SC plugin")
	}

	// Vault with SC
	scDir := filepath.Join(obsidianDir, "plugins", "smart-connections")
	if err := os.MkdirAll(scDir, 0755); err != nil {
		t.Fatal(err)
	}
	if !HasSmartConnections(vaultDir) {
		t.Error("HasSmartConnections() = false, want true with SC plugin")
	}
}

func TestDetectVault_EmptyString(t *testing.T) {
	// Empty string should use current working directory (os.Getwd)
	// and walk upward. We don't control cwd, but it should not panic.
	result := DetectVault("")
	// Result depends on whether we're inside an Obsidian vault.
	// Just verify no panic and it returns a string.
	_ = result
}

func TestIsInVault(t *testing.T) {
	tmpDir := t.TempDir()

	if IsInVault(tmpDir) {
		t.Error("IsInVault() = true, want false")
	}

	vaultDir := filepath.Join(tmpDir, "vault")
	if err := os.MkdirAll(filepath.Join(vaultDir, ".obsidian"), 0755); err != nil {
		t.Fatal(err)
	}

	if !IsInVault(vaultDir) {
		t.Error("IsInVault() = false, want true")
	}
}

func TestDetectVault_GetwdError(t *testing.T) {
	// Trigger os.Getwd() failure by chdir-ing to a deleted directory.
	// On macOS, Getwd may still succeed for deleted dirs (kernel caches vnode),
	// so this test verifies the behavior without guaranteeing the error branch.
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	tmpDir, err := os.MkdirTemp("", "vault-getwd-test")
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatal(err)
	}

	// Remove the directory while we're in it
	_ = os.RemoveAll(tmpDir)

	result := DetectVault("")

	// Restore cwd before assertions so test cleanup works
	if err := os.Chdir(origDir); err != nil {
		t.Fatalf("failed to restore working directory: %v", err)
	}

	// On macOS Getwd succeeds for deleted dirs, so result may not be "".
	// On Linux, Getwd fails and result should be "".
	// Either way, verify no panic occurred.
	_ = result
}

func TestDetectVault_WithNestedVault(t *testing.T) {
	tmpDir := t.TempDir()

	vaultDir := filepath.Join(tmpDir, "workspace", "notes")
	obsidianDir := filepath.Join(vaultDir, ".obsidian")
	subDir := filepath.Join(vaultDir, "daily", "2026", "03")
	if err := os.MkdirAll(obsidianDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}

	got := DetectVault(subDir)
	if got != vaultDir {
		t.Errorf("DetectVault(%q) = %q, want %q", subDir, got, vaultDir)
	}
}

func TestDetectVault_WalksToRoot(t *testing.T) {
	tmpDir := t.TempDir()
	deep := filepath.Join(tmpDir, "a", "b", "c", "d")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}

	result := DetectVault(deep)
	if result != "" {
		t.Errorf("DetectVault(%q) = %q, want empty string (no vault)", deep, result)
	}
}

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
