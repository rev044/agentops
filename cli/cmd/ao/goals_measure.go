package main

import (
	"time"

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
		return goals.RunMeasure(goals.MeasureOptions{
			GoalID:     goalsMeasureGoalID,
			Directives: goalsMeasureDirectives,
			GoalsFile:  resolveGoalsFile(),
			Timeout:    time.Duration(goalsTimeout) * time.Second,
			JSON:       goalsJSON,
			Verbose:    verbose,
		})
	},
}

func init() {
	goalsMeasureCmd.Flags().StringVar(&goalsMeasureGoalID, "goal", "", "Measure a single goal by ID")
	goalsMeasureCmd.Flags().BoolVar(&goalsMeasureDirectives, "directives", false, "Output directives as JSON (skip gate checks)")
	goalsCmd.AddCommand(goalsMeasureCmd)
}
