package main

import "github.com/spf13/cobra"

var workCmd = &cobra.Command{
	Use:   "work",
	Short: "Workflow commands",
	Long:  "Commands for the RPI workflow: research, plan, implement, validate.",
}

func init() {
	workCmd.GroupID = "workflow"
	rootCmd.AddCommand(workCmd)

	// Deprecated top-level aliases for backward compatibility
	rootCmd.AddCommand(deprecatedAlias("rpi", "ao work rpi", rpiCmd))
	rootCmd.AddCommand(deprecatedAlias("ratchet", "ao work ratchet", ratchetCmd))
	rootCmd.AddCommand(deprecatedAlias("goals", "ao work goals", goalsCmd))
	rootCmd.AddCommand(deprecatedAlias("session", "ao work session", sessionCmd))
	rootCmd.AddCommand(deprecatedAlias("feedback-loop", "ao work feedback-loop", feedbackLoopCmd))
	rootCmd.AddCommand(deprecatedAlias("context", "ao work context", contextCmd))
	rootCmd.AddCommand(deprecatedAlias("task-sync", "ao work task-sync", taskSyncCmd))
	rootCmd.AddCommand(deprecatedAlias("task-feedback", "ao work task-feedback", taskFeedbackCmd))
	rootCmd.AddCommand(deprecatedAlias("task-status", "ao work task-status", taskStatusCmd))
}
