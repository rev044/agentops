package main

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestEvolveCommandRegisteredOnRoot(t *testing.T) {
	if evolveCmd == nil {
		t.Fatal("evolveCmd should not be nil")
	}
	if evolveCmd.Use != "evolve [goal]" {
		t.Errorf("evolveCmd.Use = %q, want %q", evolveCmd.Use, "evolve [goal]")
	}
	if evolveCmd.GroupID != "workflow" {
		t.Errorf("evolveCmd.GroupID = %q, want workflow", evolveCmd.GroupID)
	}

	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "evolve" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("evolveCmd should be registered on rootCmd")
	}
}

func TestEvolveCommandReusesRPILoopFlags(t *testing.T) {
	for _, flag := range []string{
		"max-cycles",
		"supervisor",
		"compile",
		"gate-policy",
		"landing-policy",
		"kill-switch-path",
	} {
		if evolveCmd.Flags().Lookup(flag) == nil {
			t.Fatalf("evolve command should expose --%s", flag)
		}
	}
	if got := evolveCmd.Flags().Lookup("supervisor").DefValue; got != "true" {
		t.Fatalf("evolve --supervisor help default = %q, want true", got)
	}
}

func TestApplyEvolveDefaultsEnablesSupervisor(t *testing.T) {
	prev := snapshotLoopSupervisorGlobals()
	defer restoreLoopSupervisorGlobals(prev)

	rpiSupervisor = false
	cmd := newEvolveDefaultsTestCommand()

	applyEvolveDefaults(cmd)

	if !rpiSupervisor {
		t.Fatal("evolve should default to supervisor mode")
	}
}

func TestApplyEvolveDefaultsRespectsExplicitSupervisorFalse(t *testing.T) {
	prev := snapshotLoopSupervisorGlobals()
	defer restoreLoopSupervisorGlobals(prev)

	rpiSupervisor = false
	cmd := newEvolveDefaultsTestCommand()
	if err := cmd.ParseFlags([]string{"--supervisor=false"}); err != nil {
		t.Fatalf("parse flags: %v", err)
	}

	applyEvolveDefaults(cmd)

	if rpiSupervisor {
		t.Fatal("explicit --supervisor=false should not be overridden")
	}
}

func newEvolveDefaultsTestCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "evolve"}
	cmd.Flags().BoolVar(&rpiSupervisor, "supervisor", false, "")
	return cmd
}
