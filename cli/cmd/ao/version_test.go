package main

import (
	"testing"
)

func TestVersion_CommandExists(t *testing.T) {
	if versionCmd == nil {
		t.Fatal("versionCmd should not be nil")
	}
	if versionCmd.Use != "version" {
		t.Errorf("versionCmd.Use = %q, want %q", versionCmd.Use, "version")
	}
	if versionCmd.GroupID != "core" {
		t.Errorf("versionCmd.GroupID = %q, want %q", versionCmd.GroupID, "core")
	}
}

func TestVersion_VersionVariableHasDefault(t *testing.T) {
	// The version variable should have a default value at build time.
	// In test context it will be "dev".
	if version == "" {
		t.Error("version should not be empty")
	}
}

func TestVersion_RegisteredOnRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "version" {
			found = true
			break
		}
	}
	if !found {
		t.Error("versionCmd should be registered on rootCmd")
	}
}
