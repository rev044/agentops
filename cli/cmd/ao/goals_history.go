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

var goalsHistoryGoalID string
var goalsHistorySince string

var goalsHistoryCmd = &cobra.Command{
	Use:     "history",
	Aliases: []string{"h"},
	Short:   "Show goal measurement history",
	GroupID: "analysis",
	RunE: func(cmd *cobra.Command, args []string) error {
		historyPath := ".agents/ao/goals/history.jsonl"

		entries, err := goals.LoadHistory(historyPath)
		if err != nil {
			return fmt.Errorf("loading history: %w", err)
		}

		if len(entries) == 0 {
			fmt.Println("No history entries found. Run 'ao goals measure' first.")
			return nil
		}

		// Filter by --since
		if goalsHistorySince != "" {
			since, parseErr := time.Parse("2006-01-02", goalsHistorySince)
			if parseErr != nil {
				return fmt.Errorf("invalid --since date: %w", parseErr)
			}
			entries = goals.QueryHistory(entries, goalsHistoryGoalID, since)
		}

		if goalsJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(entries)
		}

		// Table output
		tbl := formatter.NewTable(os.Stdout, "TIMESTAMP", "PASS", "TOTAL", "SCORE", "GIT SHA")
		tbl.SetMaxWidth(0, 20)
		for _, e := range entries {
			tbl.AddRow(e.Timestamp, fmt.Sprintf("%d", e.GoalsPassing), fmt.Sprintf("%d", e.GoalsTotal), fmt.Sprintf("%.1f%%", e.Score), e.GitSHA)
		}
		if err := tbl.Render(); err != nil {
			return fmt.Errorf("rendering table: %w", err)
		}

		return nil
	},
}

func init() {
	goalsHistoryCmd.Flags().StringVar(&goalsHistoryGoalID, "goal", "", "Filter history to a specific goal")
	goalsHistoryCmd.Flags().StringVar(&goalsHistorySince, "since", "", "Show entries since date (YYYY-MM-DD)")
	goalsCmd.AddCommand(goalsHistoryCmd)
}
