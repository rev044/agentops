package main

import "github.com/spf13/cobra"

var knowCmd = &cobra.Command{
	Use:   "know",
	Short: "Knowledge management commands",
	Long:  "Commands for the knowledge flywheel: forge, search, inject, lookup.",
}

func init() {
	knowCmd.GroupID = "knowledge"
	rootCmd.AddCommand(knowCmd)

	// Deprecation aliases for visible commands that moved under "ao know"
	rootCmd.AddCommand(deprecatedAlias("forge", "ao know forge", forgeCmd))
	rootCmd.AddCommand(deprecatedAlias("inject", "ao know inject", injectCmd))
	rootCmd.AddCommand(deprecatedAlias("search", "ao know search", searchCmd))
	rootCmd.AddCommand(deprecatedAlias("lookup", "ao know lookup", lookupCmd))
	rootCmd.AddCommand(deprecatedAlias("trace", "ao know trace", traceCmd))
	// Note: "extract" alias omitted because extractCmd already exists on rootCmd
}
