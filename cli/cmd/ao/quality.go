package main

import "github.com/spf13/cobra"

var qualityCmd = &cobra.Command{
	Use:   "quality",
	Short: "Quality and validation commands",
	Long:  "Commands for quality gates, metrics, and the knowledge pool.",
}

func init() {
	qualityCmd.GroupID = "core"
	rootCmd.AddCommand(qualityCmd)

	// Deprecation aliases for visible commands that moved under "ao quality"
	rootCmd.AddCommand(deprecatedAlias("metrics", "ao quality metrics", metricsCmd))
	rootCmd.AddCommand(deprecatedAlias("flywheel", "ao quality flywheel", flywheelCmd))
	rootCmd.AddCommand(deprecatedAlias("pool", "ao quality pool", poolCmd))
	rootCmd.AddCommand(deprecatedAlias("gate", "ao quality gate", gateCmd))
	rootCmd.AddCommand(deprecatedAlias("maturity", "ao quality maturity", maturityCmd))
	rootCmd.AddCommand(deprecatedAlias("contradict", "ao quality contradict", contradictCmd))
	rootCmd.AddCommand(deprecatedAlias("dedup", "ao quality dedup", dedupCmd))
	rootCmd.AddCommand(deprecatedAlias("constraint", "ao quality constraint", constraintCmd))
	rootCmd.AddCommand(deprecatedAlias("vibe-check", "ao quality vibe-check", vibeCheckCmd))
	rootCmd.AddCommand(deprecatedAlias("badge", "ao quality badge", badgeCmd))
	rootCmd.AddCommand(deprecatedAlias("anti-patterns", "ao quality anti-patterns", antiPatternCmd))
	rootCmd.AddCommand(deprecatedAlias("curate", "ao quality curate", curateCmd))
}
