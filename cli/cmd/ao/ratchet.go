package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/boshu2/agentops/cli/internal/ratchet"
)

var ratchetCmd = &cobra.Command{
	Use:   "ratchet",
	Short: "Brownian Ratchet workflow tracking",
	Long: `Track progress through the RPI (Research-Plan-Implement) workflow.

The Brownian Ratchet ensures progress can't be lost:
  Chaos × Filter → Ratchet = Progress

Commands:
  status     Show current ratchet chain state
  check      Check if a step's gate is met
  record     Record step completion
  skip       Record intentional skip
  validate   Validate step requirements
  trace      Trace provenance backward
  spec       Get current spec path
  find       Search for artifacts across locations
  promote    Record tier promotion
  migrate    Migrate legacy chain format

The ratchet chain is stored in .agents/ao/chain.jsonl`,
}

// Ratchet command flags
var (
	ratchetEpicID      string
	ratchetChainID     string
	ratchetInput       string
	ratchetOutput      string
	ratchetReason      string
	ratchetTier        int
	ratchetLock        bool
	ratchetFiles       []string
	ratchetLenient     bool
	ratchetLenientDays int
)

// ratchetStepInfo holds step information for status output.
type ratchetStepInfo struct {
	Step     ratchet.Step       `json:"step"`
	Status   ratchet.StepStatus `json:"status"`
	Output   string             `json:"output,omitempty"`
	Input    string             `json:"input,omitempty"`
	Time     string             `json:"time,omitempty"`
	Location string             `json:"location,omitempty"`
}

// ratchetStatusOutput holds the full status output structure.
type ratchetStatusOutput struct {
	ChainID string            `json:"chain_id"`
	Started string            `json:"started"`
	EpicID  string            `json:"epic_id,omitempty"`
	Steps   []ratchetStepInfo `json:"steps"`
	Path    string            `json:"path"`
}

func init() {
	rootCmd.AddCommand(ratchetCmd)

	// status subcommand
	statusSubCmd := &cobra.Command{
		Use:   "status",
		Short: "Show ratchet chain state",
		Long: `Display the current state of the ratchet chain.

Shows all steps and their status (pending, in_progress, locked, skipped).

Examples:
  ao ratchet status
  ao ratchet status --epic ol-0001
  ao ratchet status -o json`,
		RunE: runRatchetStatus,
	}
	statusSubCmd.Flags().StringVar(&ratchetEpicID, "epic", "", "Filter by epic ID")
	statusSubCmd.Flags().StringVar(&ratchetChainID, "chain", "", "Filter by chain ID")
	ratchetCmd.AddCommand(statusSubCmd)

	// check subcommand
	checkSubCmd := &cobra.Command{
		Use:   "check <step>",
		Short: "Check if step gate is met",
		Long: `Check if prerequisites are satisfied for a workflow step.

Returns exit code 0 if gate passes, 1 if not.

Steps: research, pre-mortem, plan, formulate, implement, crank, vibe, post-mortem
Aliases: premortem, postmortem, autopilot, validate, review

Examples:
  ao ratchet check research
  ao ratchet check plan
  ao ratchet check implement || echo "Run /plan first"`,
		Args: cobra.ExactArgs(1),
		RunE: runRatchetCheck,
	}
	ratchetCmd.AddCommand(checkSubCmd)

	// record subcommand
	recordSubCmd := &cobra.Command{
		Use:   "record <step>",
		Short: "Record step completion",
		Long: `Record that a workflow step has been completed.

This locks progress - the ratchet engages.

Examples:
  ao ratchet record research --output .agents/research/topic.md
  ao ratchet record plan --input .agents/specs/spec-v2.md --output epic:ol-0001
  ao ratchet record implement --output issue:ol-0002 --tier 1`,
		Args: cobra.ExactArgs(1),
		RunE: runRatchetRecord,
	}
	recordSubCmd.Flags().StringVar(&ratchetInput, "input", "", "Input artifact path")
	recordSubCmd.Flags().StringVar(&ratchetOutput, "output", "", "Output artifact path (required)")
	recordSubCmd.Flags().IntVar(&ratchetTier, "tier", -1, "Quality tier (0-4)")
	recordSubCmd.Flags().BoolVar(&ratchetLock, "lock", true, "Lock the step (engage ratchet)")
	_ = recordSubCmd.MarkFlagRequired("output") //nolint:errcheck
	ratchetCmd.AddCommand(recordSubCmd)

	// skip subcommand
	skipSubCmd := &cobra.Command{
		Use:   "skip <step>",
		Short: "Record intentional skip",
		Long: `Record that a step was intentionally skipped.

Use this for valid workflow variations (e.g., skipping pre-mortem for bug fixes).

Examples:
  ao ratchet skip pre-mortem --reason "Bug fix, no spec needed"
  ao ratchet skip research --reason "Existing knowledge sufficient"`,
		Args: cobra.ExactArgs(1),
		RunE: runRatchetSkip,
	}
	skipSubCmd.Flags().StringVar(&ratchetReason, "reason", "", "Reason for skipping (required)")
	_ = skipSubCmd.MarkFlagRequired("reason") //nolint:errcheck
	ratchetCmd.AddCommand(skipSubCmd)

	// validate subcommand
	validateSubCmd := &cobra.Command{
		Use:   "validate <step>",
		Short: "Validate step requirements",
		Long: `Validate that an artifact meets quality requirements.

Checks for required sections, formatting, and tier criteria.

Legacy artifacts without schema_version can use --lenient mode (expires in 90 days by default).
Default mode is STRICT (requires explicit --lenient flag).

Examples:
  ao ratchet validate research --changes .agents/research/topic.md
  ao ratchet validate plan --changes epic:ol-0001
  ao ratchet validate research --changes old.md --lenient
  ao ratchet validate research --changes old.md --lenient --lenient-expiry 180`,
		Args: cobra.ExactArgs(1),
		RunE: runRatchetValidate,
	}
	validateSubCmd.Flags().StringSliceVar(&ratchetFiles, "changes", nil, "Files to validate")
	validateSubCmd.Flags().BoolVar(&ratchetLenient, "lenient", false, "Allow legacy artifacts without schema_version (expires in 90 days)")
	validateSubCmd.Flags().IntVar(&ratchetLenientDays, "lenient-expiry", 90, "Days until lenient bypass expires")
	ratchetCmd.AddCommand(validateSubCmd)

	// trace subcommand
	traceSubCmd := &cobra.Command{
		Use:   "trace <artifact>",
		Short: "Trace provenance backward",
		Long: `Trace an artifact back through the ratchet chain.

Shows the provenance chain from output to input.

Examples:
  ao ratchet trace .agents/retros/2025-01-24-topic.md
  ao ratchet trace epic:ol-0001`,
		Args: cobra.ExactArgs(1),
		RunE: runRatchetTrace,
	}
	ratchetCmd.AddCommand(traceSubCmd)

	// spec subcommand
	specSubCmd := &cobra.Command{
		Use:   "spec",
		Short: "Get current spec path",
		Long: `Find and output the current spec artifact path.

Searches for specs in priority order: crew → rig → town.

Examples:
  ao ratchet spec
  SPEC=$(ol ratchet spec) && echo $SPEC`,
		RunE: runRatchetSpec,
	}
	ratchetCmd.AddCommand(specSubCmd)

	// find subcommand
	findSubCmd := &cobra.Command{
		Use:   "find <pattern>",
		Short: "Search for artifacts",
		Long: `Search for artifacts across all locations.

Searches in order: crew → rig → town → plugins.
Warns about duplicates found in multiple locations.

Examples:
  ao ratchet find "research/*.md"
  ao ratchet find "specs/*-v2.md"
  ao ratchet find "learnings/*.md" -o json`,
		Args: cobra.ExactArgs(1),
		RunE: runRatchetFind,
	}
	ratchetCmd.AddCommand(findSubCmd)

	// promote subcommand
	promoteSubCmd := &cobra.Command{
		Use:   "promote <artifact>",
		Short: "Record tier promotion",
		Long: `Record promotion of an artifact to a higher tier.

Validates promotion requirements before recording.

Tiers:
  0: Observation (.agents/candidates/)
  1: Learning (.agents/learnings/) - requires 2+ citations
  2: Pattern (.agents/patterns/) - requires 3+ sessions
  3: Skill (plugins/*/skills/) - requires SKILL.md format
  4: Core (CLAUDE.md) - requires 10+ uses

Examples:
  ao ratchet promote .agents/candidates/insight.md --to 1
  ao ratchet promote .agents/learnings/pattern.md --to 2`,
		Args: cobra.ExactArgs(1),
		RunE: runRatchetPromote,
	}
	promoteSubCmd.Flags().IntVar(&ratchetTier, "to", -1, "Target tier (0-4, required)")
	_ = promoteSubCmd.MarkFlagRequired("to") //nolint:errcheck
	ratchetCmd.AddCommand(promoteSubCmd)

	// migrate subcommand
	migrateSubCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate legacy chain",
		Long: `Migrate chain from legacy YAML format to JSONL.

Reads from .agents/provenance/chain.yaml
Writes to .agents/ao/chain.jsonl

Examples:
  ao ratchet migrate
  ao ratchet migrate --dry-run`,
		RunE: runRatchetMigrate,
	}
	ratchetCmd.AddCommand(migrateSubCmd)
}

// runRatchetStatus displays the ratchet chain state.
func runRatchetStatus(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	chain, err := ratchet.LoadChain(cwd)
	if err != nil {
		return fmt.Errorf("load chain: %w", err)
	}

	// Get status for all steps
	allStatus := chain.GetAllStatus()

	// Build output structure
	output := ratchetStatusOutput{
		ChainID: chain.ID,
		Started: chain.Started.Format(time.RFC3339),
		EpicID:  chain.EpicID,
		Path:    chain.Path(),
		Steps:   make([]ratchetStepInfo, 0),
	}

	for _, step := range ratchet.AllSteps() {
		info := ratchetStepInfo{
			Step:   step,
			Status: allStatus[step],
		}

		// Get details from latest entry
		if entry := chain.GetLatest(step); entry != nil {
			info.Output = entry.Output
			info.Input = entry.Input
			info.Time = entry.Timestamp.Format(time.RFC3339)
			info.Location = entry.Location
		}

		output.Steps = append(output.Steps, info)
	}

	return outputRatchetStatus(&output)
}

func outputRatchetStatus(data *ratchetStatusOutput) error {
	switch GetOutput() {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)

	case "yaml":
		enc := yaml.NewEncoder(os.Stdout)
		return enc.Encode(data)

	default: // table
		fmt.Println("Ratchet Chain Status")
		fmt.Println("====================")
		fmt.Printf("Chain: %s\n", data.ChainID)
		fmt.Printf("Started: %s\n", data.Started)
		if data.EpicID != "" {
			fmt.Printf("Epic: %s\n", data.EpicID)
		}
		fmt.Println()

		fmt.Printf("%-15s %-12s %-40s\n", "STEP", "STATUS", "OUTPUT")
		fmt.Printf("%-15s %-12s %-40s\n", "----", "------", "------")

		for _, s := range data.Steps {
			icon := statusIcon(s.Status)
			out := truncate(s.Output, 40)
			fmt.Printf("%-15s %s %-10s %-40s\n", s.Step, icon, s.Status, out)
		}

		fmt.Printf("\nPath: %s\n", data.Path)
		return nil
	}
}

func statusIcon(status ratchet.StepStatus) string {
	switch status {
	case ratchet.StatusLocked:
		return "✓"
	case ratchet.StatusSkipped:
		return "⊘"
	case ratchet.StatusInProgress:
		return "◐"
	default:
		return "○"
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// runRatchetCheck validates a step gate.
func runRatchetCheck(cmd *cobra.Command, args []string) error {
	stepName := args[0]
	step := ratchet.ParseStep(stepName)
	if step == "" {
		return fmt.Errorf("unknown step: %s", stepName)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	checker, err := ratchet.NewGateChecker(cwd)
	if err != nil {
		return fmt.Errorf("create gate checker: %w", err)
	}

	result, err := checker.Check(step)
	if err != nil {
		return fmt.Errorf("check gate: %w", err)
	}

	// Output result
	switch GetOutput() {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)

	default:
		if result.Passed {
			fmt.Printf("GATE PASSED: %s\n", result.Message)
			if result.Input != "" {
				fmt.Printf("Input: %s (%s)\n", result.Input, result.Location)
			}
		} else {
			fmt.Printf("GATE FAILED: %s\n", result.Message)
			os.Exit(1)
		}
	}

	return nil
}

// runRatchetRecord records step completion.
func runRatchetRecord(cmd *cobra.Command, args []string) error {
	stepName := args[0]
	step := ratchet.ParseStep(stepName)
	if step == "" {
		return fmt.Errorf("unknown step: %s", stepName)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	if GetDryRun() {
		fmt.Printf("Would record step: %s\n", step)
		fmt.Printf("  Input: %s\n", ratchetInput)
		fmt.Printf("  Output: %s\n", ratchetOutput)
		fmt.Printf("  Locked: %v\n", ratchetLock)
		return nil
	}

	chain, err := ratchet.LoadChain(cwd)
	if err != nil {
		return fmt.Errorf("load chain: %w", err)
	}

	entry := ratchet.ChainEntry{
		Step:      step,
		Timestamp: time.Now(),
		Input:     ratchetInput,
		Output:    ratchetOutput,
		Locked:    ratchetLock,
	}

	if ratchetTier >= 0 && ratchetTier <= 4 {
		tier := ratchet.Tier(ratchetTier)
		entry.Tier = &tier
	}

	if err := chain.Append(entry); err != nil {
		return fmt.Errorf("record entry: %w", err)
	}

	fmt.Printf("Recorded: %s → %s\n", step, ratchetOutput)
	if ratchetLock {
		fmt.Println("Ratchet engaged ✓")
	}

	return nil
}

// runRatchetSkip records an intentional skip.
func runRatchetSkip(cmd *cobra.Command, args []string) error {
	stepName := args[0]
	step := ratchet.ParseStep(stepName)
	if step == "" {
		return fmt.Errorf("unknown step: %s", stepName)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	if GetDryRun() {
		fmt.Printf("Would skip step: %s\n", step)
		fmt.Printf("  Reason: %s\n", ratchetReason)
		return nil
	}

	chain, err := ratchet.LoadChain(cwd)
	if err != nil {
		return fmt.Errorf("load chain: %w", err)
	}

	entry := ratchet.ChainEntry{
		Step:      step,
		Timestamp: time.Now(),
		Skipped:   true,
		Reason:    ratchetReason,
		Locked:    true, // Skips are also locked
	}

	if err := chain.Append(entry); err != nil {
		return fmt.Errorf("record skip: %w", err)
	}

	fmt.Printf("Skipped: %s (reason: %s)\n", step, ratchetReason)

	return nil
}

// runRatchetValidate validates step requirements.
func runRatchetValidate(cmd *cobra.Command, args []string) error {
	stepName := args[0]
	step := ratchet.ParseStep(stepName)
	if step == "" {
		return fmt.Errorf("unknown step: %s", stepName)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	validator, err := ratchet.NewValidator(cwd)
	if err != nil {
		return fmt.Errorf("create validator: %w", err)
	}

	files := resolveValidationFiles(cwd, step)
	if len(files) == 0 {
		return fmt.Errorf("no files to validate (use --changes or ensure output exists)")
	}

	opts := buildValidateOptions()

	allValid := true
	for _, file := range files {
		result, err := validator.ValidateWithOptions(step, file, opts)
		if err != nil {
			return fmt.Errorf("validate %s: %w", file, err)
		}

		if GetOutput() == "json" {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(result)
		} else {
			formatValidationResult(file, result, &allValid)
		}
	}

	if !allValid {
		os.Exit(1)
	}

	return nil
}

// resolveValidationFiles determines which files to validate.
// Uses explicit --changes files if provided, otherwise locates expected output.
func resolveValidationFiles(cwd string, step ratchet.Step) []string {
	if len(ratchetFiles) > 0 {
		return ratchetFiles
	}

	locator, _ := ratchet.NewLocator(cwd)
	pattern := ratchet.GetExpectedOutput(step)
	if strings.HasPrefix(pattern, "epic:") || strings.HasPrefix(pattern, "issue:") {
		return nil
	}

	if path, _, err := locator.FindFirst(pattern); err == nil {
		return []string{path}
	}
	return nil
}

// buildValidateOptions creates validation options from command flags.
func buildValidateOptions() *ratchet.ValidateOptions {
	opts := &ratchet.ValidateOptions{
		Lenient: ratchetLenient,
	}
	if ratchetLenient && ratchetLenientDays > 0 {
		expiryTime := time.Now().AddDate(0, 0, ratchetLenientDays)
		opts.LenientExpiryDate = &expiryTime
	}
	return opts
}

// formatValidationResult prints a single validation result in text format.
func formatValidationResult(file string, result *ratchet.ValidationResult, allValid *bool) {
	fmt.Printf("Validation: %s\n", file)
	if result.Valid {
		fmt.Printf("  Status: VALID ✓\n")
	} else {
		fmt.Printf("  Status: INVALID ✗\n")
		*allValid = false
	}

	if result.Lenient {
		fmt.Printf("  Mode: LENIENT (legacy bypass)\n")
		if result.LenientExpiryDate != nil {
			fmt.Printf("  Expires: %s\n", *result.LenientExpiryDate)
		}
		if result.LenientExpiringSoon {
			fmt.Printf("  ⚠️  Expiring soon - migration required\n")
		}
	}

	if len(result.Issues) > 0 {
		fmt.Println("  Issues:")
		for _, issue := range result.Issues {
			fmt.Printf("    - %s\n", issue)
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Println("  Warnings:")
		for _, warn := range result.Warnings {
			fmt.Printf("    - %s\n", warn)
		}
	}

	if result.Tier != nil {
		fmt.Printf("  Tier: %d (%s)\n", *result.Tier, result.Tier.String())
	}
}

// runRatchetTrace traces provenance for an artifact.
func runRatchetTrace(cmd *cobra.Command, args []string) error {
	artifact := args[0]

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	chain, err := ratchet.LoadChain(cwd)
	if err != nil {
		return fmt.Errorf("load chain: %w", err)
	}

	// Find all entries that reference this artifact
	type traceEntry struct {
		Step   ratchet.Step `json:"step"`
		Input  string       `json:"input"`
		Output string       `json:"output"`
		Time   string       `json:"time"`
	}

	trace := struct {
		Artifact string       `json:"artifact"`
		Chain    []traceEntry `json:"chain"`
	}{
		Artifact: artifact,
		Chain:    []traceEntry{},
	}

	// Walk backward through chain
	current := artifact
	for i := len(chain.Entries) - 1; i >= 0; i-- {
		entry := chain.Entries[i]
		if entry.Output == current || strings.HasSuffix(entry.Output, current) {
			trace.Chain = append([]traceEntry{{
				Step:   entry.Step,
				Input:  entry.Input,
				Output: entry.Output,
				Time:   entry.Timestamp.Format(time.RFC3339),
			}}, trace.Chain...)
			current = entry.Input
		}
	}

	switch GetOutput() {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(trace)

	default:
		fmt.Printf("Provenance Trace: %s\n", artifact)
		fmt.Println("=" + strings.Repeat("=", len(artifact)+18))

		if len(trace.Chain) == 0 {
			fmt.Println("No provenance chain found")
			return nil
		}

		for i, entry := range trace.Chain {
			if i > 0 {
				fmt.Println("  ↓")
			}
			fmt.Printf("%d. %s\n", i+1, entry.Step)
			if entry.Input != "" {
				fmt.Printf("   Input:  %s\n", entry.Input)
			}
			fmt.Printf("   Output: %s\n", entry.Output)
			fmt.Printf("   Time:   %s\n", entry.Time)
		}
	}

	return nil
}

// runRatchetSpec finds the current spec path.
func runRatchetSpec(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	locator, err := ratchet.NewLocator(cwd)
	if err != nil {
		return fmt.Errorf("create locator: %w", err)
	}

	// Search for specs in order
	patterns := []string{
		"specs/*-v*.md",
		"synthesis/*.md",
	}

	for _, pattern := range patterns {
		path, loc, err := locator.FindFirst(pattern)
		if err == nil {
			switch GetOutput() {
			case "json":
				result := map[string]string{
					"path":     path,
					"location": string(loc),
				}
				enc := json.NewEncoder(os.Stdout)
				return enc.Encode(result)

			default:
				fmt.Println(path)
			}
			return nil
		}
	}

	fmt.Fprintln(os.Stderr, "No spec found")
	os.Exit(1)
	return nil
}

// runRatchetFind searches for artifacts.
func runRatchetFind(cmd *cobra.Command, args []string) error {
	pattern := args[0]

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	locator, err := ratchet.NewLocator(cwd)
	if err != nil {
		return fmt.Errorf("create locator: %w", err)
	}

	result, err := locator.Find(pattern)
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}

	switch GetOutput() {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)

	default:
		if len(result.Matches) == 0 {
			fmt.Println("No matches found")
			return nil
		}

		fmt.Printf("Found %d match(es) for: %s\n\n", len(result.Matches), pattern)

		for _, match := range result.Matches {
			fmt.Printf("[%s] %s\n", match.Location, match.Path)
		}

		if len(result.Warnings) > 0 {
			fmt.Println("\nWarnings:")
			for _, warn := range result.Warnings {
				fmt.Printf("  ! %s\n", warn)
			}
		}
	}

	return nil
}

// runRatchetPromote records tier promotion.
func runRatchetPromote(cmd *cobra.Command, args []string) error {
	artifact := args[0]
	targetTier := ratchet.Tier(ratchetTier)

	if targetTier < 0 || targetTier > 4 {
		return fmt.Errorf("tier must be 0-4, got %d", ratchetTier)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	// Validate promotion requirements
	validator, err := ratchet.NewValidator(cwd)
	if err != nil {
		return fmt.Errorf("create validator: %w", err)
	}

	result, err := validator.ValidateForPromotion(artifact, targetTier)
	if err != nil {
		return fmt.Errorf("validate promotion: %w", err)
	}

	if !result.Valid {
		fmt.Println("Promotion blocked:")
		for _, issue := range result.Issues {
			fmt.Printf("  - %s\n", issue)
		}
		os.Exit(1)
	}

	if GetDryRun() {
		fmt.Printf("Would promote %s to tier %d (%s)\n", artifact, targetTier, targetTier.String())
		return nil
	}

	// Record in chain
	chain, err := ratchet.LoadChain(cwd)
	if err != nil {
		return fmt.Errorf("load chain: %w", err)
	}

	entry := ratchet.ChainEntry{
		Step:      ratchet.Step("promotion"),
		Timestamp: time.Now(),
		Input:     artifact,
		Output:    targetTier.Location(),
		Tier:      &targetTier,
		Locked:    true,
	}

	if err := chain.Append(entry); err != nil {
		return fmt.Errorf("record promotion: %w", err)
	}

	fmt.Printf("Promoted: %s → tier %d (%s)\n", artifact, targetTier, targetTier.String())

	return nil
}

// runRatchetMigrate migrates legacy chain format.
func runRatchetMigrate(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	if GetDryRun() {
		fmt.Println("Would migrate chain from:")
		fmt.Println("  .agents/provenance/chain.yaml")
		fmt.Println("To:")
		fmt.Println("  .agents/ao/chain.jsonl")
		return nil
	}

	if err := ratchet.MigrateChain(cwd); err != nil {
		return fmt.Errorf("migrate chain: %w", err)
	}

	fmt.Println("Migration complete ✓")
	return nil
}

// ol-a46.1.3: Artifact schema version migration
func init() {
	migrateArtifactsCmd := &cobra.Command{
		Use:   "migrate-artifacts [path]",
		Short: "Add schema_version to artifacts (ol-a46.1.3)",
		Long: `Add schema_version: 1 to existing .agents/ artifacts.

Scans markdown files and adds **Schema Version:** 1 if missing.
Non-destructive: only adds the field, doesn't modify existing content.

Examples:
  ao ratchet migrate-artifacts .agents/
  ao ratchet migrate-artifacts .agents/learnings/
  ao ratchet migrate-artifacts --dry-run`,
		RunE: runMigrateArtifacts,
	}
	ratchetCmd.AddCommand(migrateArtifactsCmd)
}

func runMigrateArtifacts(cmd *cobra.Command, args []string) error {
	path := ".agents"
	if len(args) > 0 {
		path = args[0]
	}

	migrated := 0
	skipped := 0
	errors := 0

	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !shouldMigrateFile(p, info) {
			return nil
		}

		result := migrateFile(p, info)
		switch result {
		case migrateResultSuccess:
			migrated++
		case migrateResultSkipped:
			skipped++
		case migrateResultError:
			errors++
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("walk path: %w", err)
	}

	fmt.Printf("\nSummary: %d migrated, %d skipped, %d errors\n", migrated, skipped, errors)
	return nil
}

// migrateResult represents the outcome of migrating a single file.
type migrateResult int

const (
	migrateResultSuccess migrateResult = iota
	migrateResultSkipped
	migrateResultError
)

// shouldMigrateFile checks if a file is a markdown file eligible for migration.
func shouldMigrateFile(path string, info os.FileInfo) bool {
	return !info.IsDir() && strings.HasSuffix(path, ".md")
}

// findSchemaInsertPoint locates where to insert schema_version in the file.
// Returns -1 if no suitable insertion point is found.
func findSchemaInsertPoint(lines []string) int {
	insertIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "**Date:**") || strings.HasPrefix(line, "**Epic:**") {
			return i + 1
		}
		if strings.HasPrefix(line, "# ") && insertIdx == -1 {
			insertIdx = i + 1
		}
	}
	if insertIdx >= len(lines) {
		return -1
	}
	return insertIdx
}

// migrateFile reads a file, adds schema_version if missing, and writes it back.
func migrateFile(path string, info os.FileInfo) migrateResult {
	content, err := os.ReadFile(path)
	if err != nil {
		return migrateResultError
	}

	text := string(content)

	// Already has schema version
	if strings.Contains(text, "Schema Version:") || strings.Contains(text, "schema_version:") {
		return migrateResultSkipped
	}

	lines := strings.Split(text, "\n")
	insertIdx := findSchemaInsertPoint(lines)
	if insertIdx == -1 {
		return migrateResultSkipped
	}

	// Insert schema version
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertIdx]...)
	newLines = append(newLines, "**Schema Version:** 1")
	newLines = append(newLines, lines[insertIdx:]...)
	newContent := strings.Join(newLines, "\n")

	if GetDryRun() {
		fmt.Printf("Would add schema_version to: %s\n", path)
		return migrateResultSuccess
	}

	if err := os.WriteFile(path, []byte(newContent), info.Mode()); err != nil {
		return migrateResultError
	}
	fmt.Printf("✓ Migrated: %s\n", path)
	return migrateResultSuccess
}
