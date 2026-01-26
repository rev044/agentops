package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/boshu2/agentops/cli/internal/ratchet"
)

var (
	maturityApply bool
	maturityScan  bool
)

var maturityCmd = &cobra.Command{
	Use:   "maturity [learning-id]",
	Short: "Check and manage learning maturity levels",
	Long: `Check and manage CASS (Contextual Agent Session Search) maturity levels.

Learnings progress through maturity stages based on feedback:
  provisional  → Initial stage, needs positive feedback
  candidate    → Received positive feedback, being validated
  established  → Proven value through consistent positive feedback
  anti-pattern → Consistently harmful, surfaced as what NOT to do

Transition Rules:
  provisional → candidate:    utility >= 0.7 AND reward_count >= 3
  candidate → established:    utility >= 0.7 AND reward_count >= 5 AND helpful > harmful
  any → anti-pattern:         utility <= 0.2 AND harmful_count >= 5
  established → candidate:    utility < 0.5 (demotion)
  candidate → provisional:    utility < 0.3 (demotion)

Examples:
  ao maturity L001                    # Check maturity status of a learning
  ao maturity L001 --apply            # Check and apply transition if needed
  ao maturity --scan                  # Scan all learnings for pending transitions
  ao maturity --scan --apply          # Apply all pending transitions`,
	Args: cobra.MaximumNArgs(1),
	RunE: runMaturity,
}

func init() {
	rootCmd.AddCommand(maturityCmd)
	maturityCmd.Flags().BoolVar(&maturityApply, "apply", false, "Apply maturity transitions")
	maturityCmd.Flags().BoolVar(&maturityScan, "scan", false, "Scan all learnings for pending transitions")
}

func runMaturity(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	learningsDir := filepath.Join(cwd, ".agents", "learnings")
	if _, err := os.Stat(learningsDir); os.IsNotExist(err) {
		fmt.Println("No learnings directory found.")
		return nil
	}

	// Scan mode: check all learnings
	if maturityScan {
		return runMaturityScan(learningsDir)
	}

	// Single learning mode
	if len(args) == 0 {
		return fmt.Errorf("must provide learning-id or use --scan")
	}

	learningID := args[0]
	learningPath, err := findLearningFile(cwd, learningID)
	if err != nil {
		return fmt.Errorf("find learning: %w", err)
	}

	if GetDryRun() {
		fmt.Printf("[dry-run] Would check maturity for: %s\n", learningID)
		return nil
	}

	var result *ratchet.MaturityTransitionResult
	if maturityApply {
		result, err = ratchet.ApplyMaturityTransition(learningPath)
	} else {
		result, err = ratchet.CheckMaturityTransition(learningPath)
	}

	if err != nil {
		return fmt.Errorf("check maturity: %w", err)
	}

	// Output results
	if GetOutput() == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	displayMaturityResult(result, maturityApply)
	return nil
}

func runMaturityScan(learningsDir string) error {
	if GetDryRun() {
		fmt.Printf("[dry-run] Would scan learnings in: %s\n", learningsDir)
		return nil
	}

	// Get distribution first
	dist, err := ratchet.GetMaturityDistribution(learningsDir)
	if err != nil {
		return fmt.Errorf("get distribution: %w", err)
	}

	fmt.Println("=== Maturity Distribution ===")
	fmt.Printf("  Provisional:  %d\n", dist.Provisional)
	fmt.Printf("  Candidate:    %d\n", dist.Candidate)
	fmt.Printf("  Established:  %d\n", dist.Established)
	fmt.Printf("  Anti-Pattern: %d\n", dist.AntiPattern)
	fmt.Printf("  Total:        %d\n", dist.Total)
	fmt.Println()

	// Scan for pending transitions
	results, err := ratchet.ScanForMaturityTransitions(learningsDir)
	if err != nil {
		return fmt.Errorf("scan transitions: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("No pending maturity transitions found.")
		return nil
	}

	fmt.Printf("=== Pending Transitions (%d) ===\n", len(results))

	if GetOutput() == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	for _, r := range results {
		displayMaturityResult(r, false)
		fmt.Println()
	}

	// Apply transitions if requested
	if maturityApply {
		fmt.Println("=== Applying Transitions ===")
		applied := 0
		for _, r := range results {
			learningPath, err := findLearningFile(filepath.Dir(learningsDir), r.LearningID)
			if err != nil {
				VerbosePrintf("Warning: could not find %s: %v\n", r.LearningID, err)
				continue
			}

			result, err := ratchet.ApplyMaturityTransition(learningPath)
			if err != nil {
				VerbosePrintf("Warning: could not apply transition for %s: %v\n", r.LearningID, err)
				continue
			}

			if result.Transitioned {
				fmt.Printf("✓ %s: %s → %s\n", result.LearningID, result.OldMaturity, result.NewMaturity)
				applied++
			}
		}
		fmt.Printf("\nApplied %d transitions.\n", applied)
	}

	return nil
}

func displayMaturityResult(r *ratchet.MaturityTransitionResult, applied bool) {
	fmt.Printf("Learning: %s\n", r.LearningID)
	fmt.Printf("  Maturity:  %s", r.OldMaturity)
	if r.Transitioned {
		action := "→"
		if applied {
			action = "→✓"
		}
		fmt.Printf(" %s %s", action, r.NewMaturity)
	}
	fmt.Println()
	fmt.Printf("  Utility:   %.3f\n", r.Utility)
	fmt.Printf("  Confidence: %.3f\n", r.Confidence)
	fmt.Printf("  Feedback:  %d total (helpful: %d, harmful: %d)\n",
		r.RewardCount, r.HelpfulCount, r.HarmfulCount)
	fmt.Printf("  Reason:    %s\n", r.Reason)
}

// antiPatternCmd lists and manages anti-patterns.
var antiPatternCmd = &cobra.Command{
	Use:   "anti-patterns",
	Short: "List learnings marked as anti-patterns",
	Long: `List learnings that have been marked as anti-patterns.

Anti-patterns are learnings that have received consistent harmful feedback
(utility <= 0.2 and harmful_count >= 5). They are surfaced to agents as
examples of what NOT to do.

Examples:
  ao anti-patterns                    # List all anti-patterns
  ao anti-patterns --format json      # Output as JSON`,
	RunE: runAntiPatterns,
}

func init() {
	rootCmd.AddCommand(antiPatternCmd)
}

func runAntiPatterns(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	learningsDir := filepath.Join(cwd, ".agents", "learnings")
	if _, err := os.Stat(learningsDir); os.IsNotExist(err) {
		fmt.Println("No learnings directory found.")
		return nil
	}

	antiPatterns, err := ratchet.GetAntiPatterns(learningsDir)
	if err != nil {
		return fmt.Errorf("get anti-patterns: %w", err)
	}

	if len(antiPatterns) == 0 {
		fmt.Println("No anti-patterns found.")
		return nil
	}

	if GetOutput() == "json" {
		data, _ := json.MarshalIndent(antiPatterns, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("Found %d anti-pattern(s):\n\n", len(antiPatterns))
	for _, path := range antiPatterns {
		// Read summary from the file
		result, err := ratchet.CheckMaturityTransition(path)
		if err != nil {
			fmt.Printf("  • %s\n", filepath.Base(path))
			continue
		}

		fmt.Printf("  • %s\n", result.LearningID)
		fmt.Printf("    Utility: %.3f, Harmful: %d, Reason: %s\n",
			result.Utility, result.HarmfulCount, result.Reason)
	}

	return nil
}

// promoteAntiPatternsCmd explicitly promotes learnings to anti-pattern status.
var promoteAntiPatternsCmd = &cobra.Command{
	Use:   "promote-anti-patterns",
	Short: "Promote harmful learnings to anti-pattern status",
	Long: `Scan learnings and promote those meeting anti-pattern criteria.

A learning becomes an anti-pattern when:
  - utility <= 0.2 (consistently not helpful)
  - harmful_count >= 5 (multiple negative feedback events)

This is useful for batch processing to identify and mark anti-patterns
that should be surfaced as "what NOT to do".

Examples:
  ao promote-anti-patterns            # Scan and promote
  ao promote-anti-patterns --dry-run  # Preview without changing`,
	RunE: runPromoteAntiPatterns,
}

func init() {
	rootCmd.AddCommand(promoteAntiPatternsCmd)
}

func runPromoteAntiPatterns(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	learningsDir := filepath.Join(cwd, ".agents", "learnings")
	if _, err := os.Stat(learningsDir); os.IsNotExist(err) {
		fmt.Println("No learnings directory found.")
		return nil
	}

	// Scan for all transitions
	results, err := ratchet.ScanForMaturityTransitions(learningsDir)
	if err != nil {
		return fmt.Errorf("scan transitions: %w", err)
	}

	// Filter for anti-pattern promotions only
	var antiPatternPromotions []*ratchet.MaturityTransitionResult
	for _, r := range results {
		if r.NewMaturity == "anti-pattern" {
			antiPatternPromotions = append(antiPatternPromotions, r)
		}
	}

	if len(antiPatternPromotions) == 0 {
		fmt.Println("No learnings eligible for anti-pattern promotion.")
		return nil
	}

	fmt.Printf("Found %d learning(s) eligible for anti-pattern promotion:\n\n", len(antiPatternPromotions))

	for _, r := range antiPatternPromotions {
		fmt.Printf("  • %s (utility: %.3f, harmful: %d)\n",
			r.LearningID, r.Utility, r.HarmfulCount)
	}

	if GetDryRun() {
		fmt.Println("\n[dry-run] Would promote the above learnings to anti-pattern status.")
		return nil
	}

	fmt.Println("\nPromoting to anti-pattern status...")
	promoted := 0
	for _, r := range antiPatternPromotions {
		learningPath, err := findLearningFile(filepath.Dir(learningsDir), r.LearningID)
		if err != nil {
			VerbosePrintf("Warning: could not find %s: %v\n", r.LearningID, err)
			continue
		}

		result, err := ratchet.ApplyMaturityTransition(learningPath)
		if err != nil {
			VerbosePrintf("Warning: could not apply transition for %s: %v\n", r.LearningID, err)
			continue
		}

		if result.Transitioned && result.NewMaturity == "anti-pattern" {
			fmt.Printf("  ✓ %s → anti-pattern\n", result.LearningID)
			promoted++
		}
	}

	fmt.Printf("\nPromoted %d learning(s) to anti-pattern status.\n", promoted)
	return nil
}
