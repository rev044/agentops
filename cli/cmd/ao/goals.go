package main

import "github.com/spf13/cobra"

var goalsCmd = &cobra.Command{
	Use:   "goals",
	Short: "Fitness goal measurement and validation",
	Long: `Track, measure, and validate project fitness goals.

Supports both GOALS.yaml (versions 1-3) and GOALS.md (version 4) formats.
When both exist, GOALS.md takes precedence.

Measurement:
  measure (m)   Run goal checks and produce a snapshot
  validate (v)  Validate goals structure and wiring

Analysis:
  drift (d)     Compare snapshots for regressions
  history (h)   Show goal measurement history
  export (e)    Export latest snapshot as JSON

Management:
  init          Bootstrap a new GOALS.md interactively
  add (a)       Add a new goal
  steer         Manage directives (add/remove/prioritize)
  prune (p)     Remove stale gates
  migrate (mg)  Migrate between formats
  meta          Run and report meta-goals only`,
}

// Shared flags
var (
	goalsFile    string // --file, default "GOALS.yaml"
	goalsJSON    bool   // --json
	goalsTimeout int    // --timeout in seconds, default 30
)

func init() {
	goalsCmd.AddGroup(
		&cobra.Group{ID: "measurement", Title: "Measurement:"},
		&cobra.Group{ID: "analysis", Title: "Analysis:"},
		&cobra.Group{ID: "management", Title: "Management:"},
	)
	goalsCmd.PersistentFlags().StringVar(&goalsFile, "file", "GOALS.yaml", "Path to goals file")
	goalsCmd.PersistentFlags().BoolVar(&goalsJSON, "json", false, "Output as JSON")
	goalsCmd.PersistentFlags().IntVar(&goalsTimeout, "timeout", 30, "Check timeout in seconds")
	goalsCmd.GroupID = "workflow"
	rootCmd.AddCommand(goalsCmd)
}
