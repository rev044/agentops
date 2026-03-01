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

var goalsMetaCmd = &cobra.Command{
	Use:     "meta",
	Short:   "Run and report meta-goals only",
	GroupID: "management",
	RunE: func(cmd *cobra.Command, args []string) error {
		gf, err := goals.LoadGoals(resolveGoalsFile())
		if err != nil {
			return fmt.Errorf("loading goals: %w", err)
		}

		// Filter to meta-goals only.
		var metaGoals []goals.Goal
		for _, g := range gf.Goals {
			if g.Type == goals.GoalTypeMeta {
				metaGoals = append(metaGoals, g)
			}
		}

		if len(metaGoals) == 0 {
			fmt.Println("No meta-goals found (type: meta)")
			return nil
		}

		// Build a filtered GoalFile.
		metaGF := &goals.GoalFile{
			Version: gf.Version,
			Mission: gf.Mission,
			Goals:   metaGoals,
		}

		timeout := time.Duration(goalsTimeout) * time.Second
		snap := goals.Measure(metaGF, timeout)

		if goalsJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(snap)
		}

		// Table output.
		fmt.Printf("Meta-Goals: %d total\n\n", len(metaGoals))
		tbl := formatter.NewTable(os.Stdout, "GOAL", "RESULT", "DURATION")
		tbl.SetMaxWidth(0, 30)
		for _, m := range snap.Goals {
			tbl.AddRow(m.GoalID, m.Result, fmt.Sprintf("%.1fs", m.Duration))
		}
		tbl.Render()
		fmt.Println()

		if snap.Summary.Failing > 0 {
			fmt.Printf("META-HEALTH: DEGRADED (%d/%d failing)\n", snap.Summary.Failing, snap.Summary.Total)
			return fmt.Errorf("meta-goal failures detected")
		}

		fmt.Printf("META-HEALTH: OK (%d/%d passing)\n", snap.Summary.Passing, snap.Summary.Total)
		return nil
	},
}

func init() {
	goalsCmd.AddCommand(goalsMetaCmd)
}
