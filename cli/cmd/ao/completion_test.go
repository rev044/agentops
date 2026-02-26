package main

import (
	"testing"
)

func TestCompletion_CommandExists(t *testing.T) {
	if completionCmd == nil {
		t.Fatal("completionCmd should not be nil")
	}
	if completionCmd.Use != "completion [bash|zsh|fish]" {
		t.Errorf("completionCmd.Use = %q, want %q", completionCmd.Use, "completion [bash|zsh|fish]")
	}
	if completionCmd.GroupID != "config" {
		t.Errorf("completionCmd.GroupID = %q, want %q", completionCmd.GroupID, "config")
	}
}

func TestCompletion_ValidArgs(t *testing.T) {
	expected := []string{"bash", "zsh", "fish"}
	if len(completionCmd.ValidArgs) != len(expected) {
		t.Fatalf("ValidArgs length = %d, want %d", len(completionCmd.ValidArgs), len(expected))
	}
	for i, got := range completionCmd.ValidArgs {
		if got != expected[i] {
			t.Errorf("ValidArgs[%d] = %q, want %q", i, got, expected[i])
		}
	}
}

func TestCompletion_RegisteredOnRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "completion" {
			found = true
			break
		}
	}
	if !found {
		t.Error("completionCmd should be registered on rootCmd")
	}
}
