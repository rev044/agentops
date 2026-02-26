package main

import "github.com/spf13/cobra"

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Getting started commands",
	Long:  "Commands for onboarding and initial setup.",
}

func init() {
	startCmd.GroupID = "start"
	rootCmd.AddCommand(startCmd)

	// Deprecated top-level aliases for backward compatibility
	rootCmd.AddCommand(deprecatedAlias("demo", "ao start demo", demoCmd))
	rootCmd.AddCommand(deprecatedAlias("init", "ao start init", initCmd))
	rootCmd.AddCommand(deprecatedAlias("seed", "ao start seed", seedCmd))
	rootCmd.AddCommand(deprecatedAlias("quick-start", "ao start quick-start", quickstartCmd))
}
