package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	dryRun  bool
	verbose bool
	output  string
	cfgFile string
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "ol",
	Short: "Olympus Knowledge Compounding CLI",
	Long: `ol is the CLI for Olympus, a knowledge compounding workflow system.

"Problem in. Value out. Intelligence compounds."

Get Started:
  demo         Interactive demo (see value in 5 minutes)
  quick-start  Set up Olympus in your project

Core Commands:
  forge        Extract knowledge from transcripts
  pool         Manage quality pools
  gate         Human review gates
  trace        Track knowledge provenance
  status       Show current state
  version      Show version information

The Knowledge Flywheel:
  Sessions compound via .agents/ + Smart Connections.
  Others start fresh. You get smarter every session.`,
	SilenceUsage: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags available to all commands
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would happen without executing")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "table", "Output format (json, table, yaml)")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default: ~/.olympus/config.yaml)")
}

// GetDryRun returns the dry-run flag value for use by subcommands.
func GetDryRun() bool {
	return dryRun
}

// GetVerbose returns the verbose flag value for use by subcommands.
func GetVerbose() bool {
	return verbose
}

// GetOutput returns the output format for use by subcommands.
func GetOutput() string {
	return output
}

// GetConfigFile returns the config file path for use by subcommands.
func GetConfigFile() string {
	return cfgFile
}

// VerbosePrintf prints only when verbose mode is enabled.
func VerbosePrintf(format string, args ...interface{}) {
	if verbose {
		fmt.Printf(format, args...)
	}
}
