package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/boshu2/agentops/cli/internal/pool"
	"github.com/boshu2/agentops/cli/internal/types"
)

var (
	poolTier      string
	poolStatus    string
	poolLimit     int
	poolReason    string
	poolThreshold string
)

var poolCmd = &cobra.Command{
	Use:   "pool",
	Short: "Manage quality pools",
	Long: `Manage knowledge candidates in quality pools.

Pools organize candidates by their processing status:
  pending    Awaiting initial scoring
  staged     Ready for promotion to Athena
  promoted   Successfully stored in Athena
  rejected   Rejected during review

Examples:
  ao pool list --tier=gold
  ao pool show <candidate-id>
  ao pool stage <candidate-id>
  ao pool promote <candidate-id>`,
}

var poolListCmd = &cobra.Command{
	Use:   "list",
	Short: "List candidates in pools",
	Long: `List knowledge candidates filtered by tier and/or status.

Examples:
  ao pool list
  ao pool list --tier=gold
  ao pool list --status=pending
  ao pool list --tier=bronze --status=staged`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if GetDryRun() {
			fmt.Printf("[dry-run] Would list pool entries")
			if poolTier != "" {
				fmt.Printf(" with tier=%s", poolTier)
			}
			if poolStatus != "" {
				fmt.Printf(" with status=%s", poolStatus)
			}
			fmt.Println()
			return nil
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}

		p := pool.NewPool(cwd)

		opts := pool.ListOptions{
			Limit: poolLimit,
		}

		if poolTier != "" {
			opts.Tier = types.Tier(poolTier)
		}
		if poolStatus != "" {
			opts.Status = types.PoolStatus(poolStatus)
		}

		entries, err := p.List(opts)
		if err != nil {
			return fmt.Errorf("list pool: %w", err)
		}

		return outputPoolList(entries)
	},
}

func outputPoolList(entries []pool.PoolEntry) error {
	switch GetOutput() {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(entries)

	case "yaml":
		enc := yaml.NewEncoder(os.Stdout)
		return enc.Encode(entries)

	default: // table
		if len(entries) == 0 {
			fmt.Println("No pool entries found")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		//nolint:errcheck // CLI tabwriter output to stdout, errors unlikely and non-recoverable
		fmt.Fprintln(w, "ID\tTIER\tSTATUS\tAGE\tUTILITY\tCONFIDENCE")
		//nolint:errcheck // CLI tabwriter output to stdout
		fmt.Fprintln(w, "--\t----\t------\t---\t-------\t----------")

		for _, e := range entries {
			//nolint:errcheck // CLI tabwriter output to stdout
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%.2f\t%.2f\n",
				truncateID(e.Candidate.ID, 12),
				e.Candidate.Tier,
				e.Status,
				e.AgeString,
				e.Candidate.Utility,
				e.Candidate.Confidence,
			)
		}

		return w.Flush()
	}
}

var poolShowCmd = &cobra.Command{
	Use:   "show <candidate-id>",
	Short: "Show candidate details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		candidateID := args[0]

		if GetDryRun() {
			fmt.Printf("[dry-run] Would show candidate %s\n", candidateID)
			return nil
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}

		p := pool.NewPool(cwd)

		entry, err := p.Get(candidateID)
		if err != nil {
			return fmt.Errorf("get candidate: %w", err)
		}

		return outputPoolShow(entry)
	},
}

func outputPoolShow(entry *pool.PoolEntry) error {
	switch GetOutput() {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(entry)

	case "yaml":
		enc := yaml.NewEncoder(os.Stdout)
		return enc.Encode(entry)

	default: // detailed text
		fmt.Printf("Candidate: %s\n", entry.Candidate.ID)
		fmt.Printf("============%s\n", repeat("=", len(entry.Candidate.ID)))
		fmt.Println()

		fmt.Printf("Type:      %s\n", entry.Candidate.Type)
		fmt.Printf("Tier:      %s\n", entry.Candidate.Tier)
		fmt.Printf("Status:    %s\n", entry.Status)
		fmt.Printf("Age:       %s\n", entry.AgeString)
		fmt.Println()

		fmt.Println("MemRL Metrics:")
		fmt.Printf("  Utility:    %.3f\n", entry.Candidate.Utility)
		fmt.Printf("  Confidence: %.3f\n", entry.Candidate.Confidence)
		fmt.Printf("  Maturity:   %s\n", entry.Candidate.Maturity)
		fmt.Printf("  Rewards:    %d\n", entry.Candidate.RewardCount)
		fmt.Println()

		fmt.Println("Scoring:")
		fmt.Printf("  Raw Score:  %.3f\n", entry.ScoringResult.RawScore)
		fmt.Printf("  Rubric:\n")
		fmt.Printf("    Specificity:   %.2f\n", entry.ScoringResult.Rubric.Specificity)
		fmt.Printf("    Actionability: %.2f\n", entry.ScoringResult.Rubric.Actionability)
		fmt.Printf("    Novelty:       %.2f\n", entry.ScoringResult.Rubric.Novelty)
		fmt.Printf("    Context:       %.2f\n", entry.ScoringResult.Rubric.Context)
		fmt.Printf("    Confidence:    %.2f\n", entry.ScoringResult.Rubric.Confidence)
		fmt.Println()

		fmt.Println("Provenance:")
		fmt.Printf("  Session:    %s\n", entry.Candidate.Source.SessionID)
		fmt.Printf("  Transcript: %s\n", entry.Candidate.Source.TranscriptPath)
		fmt.Printf("  Message:    %d\n", entry.Candidate.Source.MessageIndex)
		fmt.Println()

		fmt.Println("Content:")
		fmt.Println("---")
		fmt.Println(entry.Candidate.Content)
		fmt.Println("---")

		if entry.HumanReview != nil && entry.HumanReview.Reviewed {
			fmt.Println()
			fmt.Println("Human Review:")
			fmt.Printf("  Approved:   %v\n", entry.HumanReview.Approved)
			fmt.Printf("  Reviewer:   %s\n", entry.HumanReview.Reviewer)
			fmt.Printf("  Notes:      %s\n", entry.HumanReview.Notes)
		}

		return nil
	}
}

var poolStageCmd = &cobra.Command{
	Use:   "stage <candidate-id>",
	Short: "Stage candidate for promotion",
	Long: `Move a candidate from pending to staged status.

Validates that the candidate meets the minimum tier threshold (default: bronze).
Staged candidates are ready for promotion to the knowledge base.

Examples:
  ao pool stage cand-abc123
  ao pool stage cand-abc123 --min-tier=silver`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		candidateID := args[0]

		if GetDryRun() {
			fmt.Printf("[dry-run] Would stage candidate %s\n", candidateID)
			return nil
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}

		p := pool.NewPool(cwd)

		minTier := types.TierBronze
		if poolTier != "" {
			minTier = types.Tier(poolTier)
		}

		if err := p.Stage(candidateID, minTier); err != nil {
			return fmt.Errorf("stage candidate: %w", err)
		}

		fmt.Printf("Staged: %s\n", candidateID)
		return nil
	},
}

var poolPromoteCmd = &cobra.Command{
	Use:   "promote <candidate-id>",
	Short: "Promote candidate to knowledge base",
	Long: `Move a staged candidate to the knowledge base (.agents/learnings/ or .agents/patterns/).

Locks the artifact with the ratchet and records the promotion in chain.jsonl.

Examples:
  ao pool promote cand-abc123`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		candidateID := args[0]

		if GetDryRun() {
			fmt.Printf("[dry-run] Would promote candidate %s\n", candidateID)
			return nil
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}

		p := pool.NewPool(cwd)

		artifactPath, err := p.Promote(candidateID)
		if err != nil {
			return fmt.Errorf("promote candidate: %w", err)
		}

		fmt.Printf("Promoted: %s\n", candidateID)
		fmt.Printf("Artifact: %s\n", artifactPath)

		// Optionally lock with ratchet
		VerbosePrintf("Run 'ao ratchet record promotion --output %s' to lock\n", artifactPath)

		return nil
	},
}

var poolRejectCmd = &cobra.Command{
	Use:   "reject <candidate-id>",
	Short: "Reject candidate",
	Long: `Mark a candidate as rejected and move to rejected directory.

A reason must be provided for audit purposes.

Examples:
  ao pool reject cand-abc123 --reason="Too vague, lacks specificity"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		candidateID := args[0]

		if poolReason == "" {
			return fmt.Errorf("--reason is required for rejection")
		}

		if GetDryRun() {
			fmt.Printf("[dry-run] Would reject candidate %s with reason: %s\n", candidateID, poolReason)
			return nil
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}

		p := pool.NewPool(cwd)

		// Get reviewer from system user (not spoofable via env)
		reviewer := GetCurrentUser()

		if err := p.Reject(candidateID, poolReason, reviewer); err != nil {
			return fmt.Errorf("reject candidate: %w", err)
		}

		fmt.Printf("Rejected: %s\n", candidateID)
		fmt.Printf("Reason: %s\n", poolReason)

		return nil
	},
}

var poolAutoPromoteCmd = &cobra.Command{
	Use:   "auto-promote",
	Short: "Auto-promote silver candidates older than threshold",
	Long: `Automatically approve silver-tier candidates that have been pending
for longer than the specified threshold.

This is a bulk operation - use with caution. The threshold must be at least
1 hour to prevent accidental mass approval of recently added candidates.

Examples:
  ao pool auto-promote --threshold=24h
  ao pool auto-promote --threshold=48h --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if poolThreshold == "" {
			return fmt.Errorf("--threshold is required")
		}

		threshold, err := time.ParseDuration(poolThreshold)
		if err != nil {
			return fmt.Errorf("invalid threshold: %w", err)
		}

		if GetDryRun() {
			fmt.Printf("[dry-run] Would auto-promote silver candidates older than %s\n", threshold)
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}

		p := pool.NewPool(cwd)
		reviewer := GetCurrentUser()

		approved, err := p.BulkApprove(threshold, reviewer, GetDryRun())
		if err != nil {
			return fmt.Errorf("auto-promote: %w", err)
		}

		if len(approved) == 0 {
			fmt.Println("No candidates eligible for auto-promotion")
			return nil
		}

		if GetDryRun() {
			fmt.Printf("Would auto-promote %d candidates:\n", len(approved))
		} else {
			fmt.Printf("Auto-promoted %d candidates:\n", len(approved))
		}
		for _, id := range approved {
			fmt.Printf("  - %s\n", id)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(poolCmd)

	// Add subcommands
	poolCmd.AddCommand(poolListCmd)
	poolCmd.AddCommand(poolShowCmd)
	poolCmd.AddCommand(poolStageCmd)
	poolCmd.AddCommand(poolPromoteCmd)
	poolCmd.AddCommand(poolRejectCmd)
	poolCmd.AddCommand(poolAutoPromoteCmd)

	// Add flags to list command
	poolListCmd.Flags().StringVar(&poolTier, "tier", "", "Filter by tier (gold, silver, bronze)")
	poolListCmd.Flags().StringVar(&poolStatus, "status", "", "Filter by status (pending, staged, promoted, rejected)")
	poolListCmd.Flags().IntVar(&poolLimit, "limit", 0, "Limit number of results")

	// Add flags to stage command
	poolStageCmd.Flags().StringVar(&poolTier, "min-tier", "", "Minimum tier threshold (default: bronze)")

	// Add flags to reject command
	poolRejectCmd.Flags().StringVar(&poolReason, "reason", "", "Reason for rejection (required)")
	_ = poolRejectCmd.MarkFlagRequired("reason") //nolint:errcheck

	// Add flags to auto-promote command
	poolAutoPromoteCmd.Flags().StringVar(&poolThreshold, "threshold", "", "Minimum age for auto-promotion (e.g., 24h)")
	_ = poolAutoPromoteCmd.MarkFlagRequired("threshold") //nolint:errcheck
}

// truncateID shortens an ID for display.
func truncateID(id string, max int) string {
	if len(id) <= max {
		return id
	}
	return id[:max-3] + "..."
}

// repeat returns a string repeated n times.
func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
