package main

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestDeprecatedAlias(t *testing.T) {
	// Create a target command
	targetRan := false
	target := &cobra.Command{
		Use:   "new-cmd",
		Short: "The new command",
		RunE: func(cmd *cobra.Command, args []string) error {
			targetRan = true
			return nil
		},
	}

	alias := deprecatedAlias("old-cmd", "ao group new-cmd", target)

	// Verify alias properties
	if alias.Use != "old-cmd" {
		t.Errorf("expected Use='old-cmd', got '%s'", alias.Use)
	}
	if !alias.Hidden {
		t.Error("expected alias to be hidden")
	}

	// Verify deprecation message in Short
	if alias.Short != "DEPRECATED: use 'ao group new-cmd' instead" {
		t.Errorf("unexpected Short: %s", alias.Short)
	}

	// Execute the alias and check it forwards to target
	err := alias.RunE(alias, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !targetRan {
		t.Error("expected target command to run")
	}
}

func TestDeprecatedAliasWithRunTarget(t *testing.T) {
	// Test with Run (not RunE) target
	targetRan := false
	target := &cobra.Command{
		Use:   "run-cmd",
		Short: "Run target",
		Run: func(cmd *cobra.Command, args []string) {
			targetRan = true
		},
	}

	alias := deprecatedAlias("old-run", "ao group run-cmd", target)
	err := alias.RunE(alias, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !targetRan {
		t.Error("expected target Run to execute")
	}
}
