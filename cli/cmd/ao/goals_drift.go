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

var goalsDriftCmd = &cobra.Command{
	Use:     "drift",
	Aliases: []string{"d"},
	Short:   "Compare snapshots for regressions",
	GroupID: "analysis",
	RunE: func(cmd *cobra.Command, args []string) error {
		snapDir := ".agents/ao/goals/baselines"

		gf, err := goals.LoadGoals(resolveGoalsFile())
		if err != nil {
			return fmt.Errorf("loading goals: %w", err)
		}

		latest, err := goals.LoadLatestSnapshot(snapDir)
		if err != nil {
			// No snapshots — measure fresh and report no baseline
			timeout := time.Duration(goalsTimeout) * time.Second
			snap := goals.Measure(gf, timeout)
			if _, saveErr := goals.SaveSnapshot(snap, snapDir); saveErr != nil {
				fmt.Fprintf(os.Stderr, "warning: could not save snapshot: %v\n", saveErr)
			}
			fmt.Println("No baseline snapshot found. Created initial snapshot.")
			fmt.Printf("Score: %.1f%% (%d/%d passing)\n", snap.Summary.Score, snap.Summary.Passing, snap.Summary.Total)
			return nil
		}

		// Measure current state
		timeout := time.Duration(goalsTimeout) * time.Second
		current := goals.Measure(gf, timeout)
		if _, saveErr := goals.SaveSnapshot(current, snapDir); saveErr != nil {
			fmt.Fprintf(os.Stderr, "warning: could not save snapshot: %v\n", saveErr)
		}

		drifts := goals.ComputeDrift(latest, current)

		if goalsJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(drifts)
		}

		// Table output
		regressions := 0
		improvements := 0
		for _, d := range drifts {
			if d.Delta == "regressed" {
				regressions++
			}
			if d.Delta == "improved" {
				improvements++
			}
		}

		fmt.Printf("Drift: %d regressions, %d improvements, %d unchanged\n\n",
			regressions, improvements, len(drifts)-regressions-improvements)

		if regressions > 0 || improvements > 0 {
			tbl := formatter.NewTable(os.Stdout, "GOAL", "DELTA", "BEFORE", "AFTER")
			tbl.SetMaxWidth(0, 30)
			for _, d := range drifts {
				if d.Delta == "unchanged" {
					continue
				}
				tbl.AddRow(d.GoalID, d.Delta, d.Before, fmt.Sprintf("-> %s", d.After))
			}
			_ = tbl.Render()
			fmt.Println()
		}

		// Score comparison
		fmt.Printf("Baseline: %.1f%% -> Current: %.1f%%\n", latest.Summary.Score, current.Summary.Score)

		return nil
	},
}

func init() {
	goalsCmd.AddCommand(goalsDriftCmd)
}
