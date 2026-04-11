package main

import "github.com/spf13/cobra"

var evolveCmd = &cobra.Command{
	Use:   "evolve [goal]",
	Short: "Run the autonomous code-improvement loop",
	Long: `Run the v2 autonomous improvement loop.

This is the top-level operator surface for the old /evolve flow. It uses the
same engine as "ao rpi loop" and defaults to supervisor mode so each cycle gets
lease locking, compile producer cadence, quality gates, retries, and cleanup.

Examples:
  ao evolve                          # run until queue stable or stopped
  ao evolve --max-cycles 1           # one supervised autonomous cycle
  ao evolve "improve test coverage"  # run one explicit-goal cycle
  ao evolve --supervisor=false       # use raw rpi loop defaults`,
	Args: cobra.MaximumNArgs(1),
	RunE: runEvolve,
}

func init() {
	evolveCmd.GroupID = "workflow"
	addRPILoopFlags(evolveCmd)
	if flag := evolveCmd.Flags().Lookup("supervisor"); flag != nil {
		flag.DefValue = "true"
	}
	rootCmd.AddCommand(evolveCmd)
}

func runEvolve(cmd *cobra.Command, args []string) error {
	applyEvolveDefaults(cmd)
	return runRPILoop(cmd, args)
}

func applyEvolveDefaults(cmd *cobra.Command) {
	if cmd != nil && cmd.Flags().Changed("supervisor") {
		return
	}
	rpiSupervisor = true
}
