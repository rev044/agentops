package main

import (
	"github.com/spf13/cobra"
)

// settingsCmd is the parent for configuration and settings commands.
var settingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Configuration and settings commands",
	Long:  "Commands for managing configuration, hooks, memory, and plans.",
}

func init() {
	settingsCmd.GroupID = "config"
	rootCmd.AddCommand(settingsCmd)

	// Reparent subcommands under settings
	settingsCmd.AddCommand(configCmd)
	settingsCmd.AddCommand(plansCmd)
	settingsCmd.AddCommand(hooksCmd)
	settingsCmd.AddCommand(memoryCmd)
	settingsCmd.AddCommand(notebookCmd)
	settingsCmd.AddCommand(worktreeCmd)

	// Deprecation aliases for backward compatibility (visible commands only)
	rootCmd.AddCommand(deprecatedAlias("config", "ao settings config", configCmd))
	rootCmd.AddCommand(deprecatedAlias("plans", "ao settings plans", plansCmd))
	rootCmd.AddCommand(deprecatedAlias("hooks", "ao settings hooks", hooksCmd))
	rootCmd.AddCommand(deprecatedAlias("memory", "ao settings memory", memoryCmd))
	rootCmd.AddCommand(deprecatedAlias("notebook", "ao settings notebook", notebookCmd))
	rootCmd.AddCommand(deprecatedAlias("worktree", "ao settings worktree", worktreeCmd))
}
