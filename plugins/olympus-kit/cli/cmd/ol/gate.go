package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	gateNote   string
	gateReason string
)

var gateCmd = &cobra.Command{
	Use:   "gate",
	Short: "Human review gates",
	Long: `Manage human review gates for bronze-tier candidates.

Bronze-tier candidates (score 0.50-0.69) require human review
before promotion. The gate command provides the review interface.

Examples:
  ol gate pending
  ol gate approve <candidate-id>
  ol gate reject <candidate-id> --reason="Too vague"`,
}

var gatePendingCmd = &cobra.Command{
	Use:   "pending",
	Short: "List candidates pending review",
	RunE: func(cmd *cobra.Command, args []string) error {
		if GetDryRun() {
			fmt.Println("[dry-run] Would list pending gate reviews")
			return nil
		}

		// TODO: Implement gate pending
		fmt.Println("Gate pending not yet implemented")
		return nil
	},
}

var gateApproveCmd = &cobra.Command{
	Use:   "approve <candidate-id>",
	Short: "Approve candidate for promotion",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		candidateID := args[0]

		if GetDryRun() {
			fmt.Printf("[dry-run] Would approve candidate %s", candidateID)
			if gateNote != "" {
				fmt.Printf(" with note: %s", gateNote)
			}
			fmt.Println()
			return nil
		}

		// TODO: Implement gate approve
		fmt.Printf("Gate approve not yet implemented for %s\n", candidateID)
		return nil
	},
}

var gateRejectCmd = &cobra.Command{
	Use:   "reject <candidate-id>",
	Short: "Reject candidate",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		candidateID := args[0]

		if gateReason == "" {
			return fmt.Errorf("--reason is required for rejection")
		}

		if GetDryRun() {
			fmt.Printf("[dry-run] Would reject candidate %s with reason: %s\n", candidateID, gateReason)
			return nil
		}

		// TODO: Implement gate reject
		fmt.Printf("Gate reject not yet implemented for %s\n", candidateID)
		return nil
	},
}

var gateBulkApproveCmd = &cobra.Command{
	Use:   "bulk-approve",
	Short: "Bulk approve silver candidates",
	Long: `Approve all silver-tier candidates older than a threshold.

Silver candidates auto-promote after 24h if not rejected.
This command accelerates the process for reviewed batches.

Example:
  ol gate bulk-approve --tier=silver --older-than=24h`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if GetDryRun() {
			fmt.Println("[dry-run] Would bulk approve candidates")
			return nil
		}

		// TODO: Implement bulk approve
		fmt.Println("Gate bulk-approve not yet implemented")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(gateCmd)

	// Add subcommands
	gateCmd.AddCommand(gatePendingCmd)
	gateCmd.AddCommand(gateApproveCmd)
	gateCmd.AddCommand(gateRejectCmd)
	gateCmd.AddCommand(gateBulkApproveCmd)

	// Add flags
	gateApproveCmd.Flags().StringVar(&gateNote, "note", "", "Optional approval note")
	gateRejectCmd.Flags().StringVar(&gateReason, "reason", "", "Required rejection reason")
	_ = gateRejectCmd.MarkFlagRequired("reason") //nolint:errcheck
}
