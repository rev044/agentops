package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/boshu2/agentops/cli/internal/formatter"
	"github.com/boshu2/agentops/cli/internal/goals"
	"github.com/spf13/cobra"
)

var (
	goalsMeasureGoalID     string
	goalsMeasureDirectives bool
)

var goalsMeasureCmd = &cobra.Command{
	Use:     "measure",
	Aliases: []string{"m"},
	Short:   "Run goal checks and produce a snapshot",
	GroupID: "measurement",
	RunE: func(cmd *cobra.Command, args []string) error {
		gf, err := goals.LoadGoals(resolveGoalsFile())
		if err != nil {
			return fmt.Errorf("loading goals: %w", err)
		}

		// Early exit: output directives as JSON and skip gate checks.
		// Directives are only available in GOALS.md (version 4) format;
		// YAML files have no directives section and silently return empty.
		if goalsMeasureDirectives {
			if goalsMeasureGoalID != "" {
				return fmt.Errorf("--directives and --goal cannot be combined")
			}
			if gf.Format != "md" {
				fmt.Fprintln(os.Stderr, "Warning: --directives requires GOALS.md format. Run 'ao goals migrate --to-md' to convert.")
				return nil
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(gf.Directives)
		}

		if errs := goals.ValidateGoals(gf); len(errs) > 0 {
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "validation: %s\n", e)
			}
			return fmt.Errorf("%d validation errors", len(errs))
		}

		timeout := time.Duration(goalsTimeout) * time.Second

		// Filter to single goal if --goal specified
		if goalsMeasureGoalID != "" {
			var filtered []goals.Goal
			for _, g := range gf.Goals {
				if g.ID == goalsMeasureGoalID {
					filtered = append(filtered, g)
				}
			}
			if len(filtered) == 0 {
				return fmt.Errorf("goal %q not found", goalsMeasureGoalID)
			}
			gf.Goals = filtered
		}

		snap := goals.Measure(gf, timeout)

		// Save snapshot
		snapDir := ".agents/ao/goals/baselines"
		path, err := goals.SaveSnapshot(snap, snapDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not save snapshot: %v\n", err)
		} else if verbose {
			fmt.Fprintf(os.Stderr, "Snapshot saved: %s\n", path)
		}

		if goalsJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(snap)
		}

		// Table output
		tbl := formatter.NewTable(os.Stdout, "GOAL", "RESULT", "DURATION", "WEIGHT")
		tbl.SetMaxWidth(0, 30)
		for _, m := range snap.Goals {
			tbl.AddRow(m.GoalID, m.Result, fmt.Sprintf("%.1fs", m.Duration), fmt.Sprintf("%d", m.Weight))
		}
		tbl.Render()
		fmt.Println()
		fmt.Printf("Score: %.1f%% (%d/%d passing, %d skipped)\n",
			snap.Summary.Score, snap.Summary.Passing, snap.Summary.Total, snap.Summary.Skipped)

		return nil
	},
}

func init() {
	goalsMeasureCmd.Flags().StringVar(&goalsMeasureGoalID, "goal", "", "Measure a single goal by ID")
	goalsMeasureCmd.Flags().BoolVar(&goalsMeasureDirectives, "directives", false, "Output directives as JSON (skip gate checks)")
	goalsCmd.AddCommand(goalsMeasureCmd)
}
