package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/boshu2/agentops/cli/internal/orchestrator"
)

var (
	synthesizePlanID string
)

var synthesizeCmd = &cobra.Command{
	Use:   "synthesize",
	Short: "Synthesize agent findings",
	Long: `Synthesize findings from multiple agents into a final verdict.

This implements Wave 2 and Wave 3 of the orchestrator pattern:
- Wave 2: Merge findings with deduplication
- Wave 3: Apply quorum rules and produce final verdict

The command reads findings from .agents/ao/findings/<plan-id>/ and produces
a final validation report.

Example:
  ao synthesize --plan-id vibe-abc12345

This is typically called by the /vibe skill after dispatching agents.`,
	RunE: runSynthesize,
}

func init() {
	rootCmd.AddCommand(synthesizeCmd)

	synthesizeCmd.Flags().StringVar(&synthesizePlanID, "plan-id", "", "Plan ID to synthesize (required)")
	_ = synthesizeCmd.MarkFlagRequired("plan-id")
}

// AgentFindings represents findings from a single agent.
type AgentFindings struct {
	Category string                `json:"category"`
	Findings []orchestrator.Finding `json:"findings"`
	Summary  string                `json:"summary,omitempty"`
}

// SynthesisResult holds the final synthesized result.
type SynthesisResult struct {
	PlanID        string                  `json:"plan_id"`
	Verdict       orchestrator.Severity   `json:"verdict"`
	Grade         string                  `json:"grade"`
	Findings      []orchestrator.Finding  `json:"findings"`
	CriticalCount int                     `json:"critical_count"`
	HighCount     int                     `json:"high_count"`
	MediumCount   int                     `json:"medium_count"`
	LowCount      int                     `json:"low_count"`
	Summary       string                  `json:"summary"`
	AgentReports  []AgentFindings         `json:"agent_reports"`
	CompletedAt   time.Time               `json:"completed_at"`
}

func runSynthesize(cmd *cobra.Command, args []string) error {
	if synthesizePlanID == "" {
		return fmt.Errorf("--plan-id is required")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	findingsDir := filepath.Join(cwd, ".agents", "ao", "findings", synthesizePlanID)

	// Check if findings directory exists
	if _, err := os.Stat(findingsDir); os.IsNotExist(err) {
		return fmt.Errorf("findings directory not found: %s\nRun 'ao orchestrate' first", findingsDir)
	}

	// Load all findings
	agentReports, err := loadAgentFindings(findingsDir)
	if err != nil {
		return fmt.Errorf("load findings: %w", err)
	}

	if len(agentReports) == 0 {
		return fmt.Errorf("no findings files found in %s", findingsDir)
	}

	// Synthesize
	result := synthesizeFindings(synthesizePlanID, agentReports)

	// Output
	switch GetOutput() {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)

	case "yaml":
		enc := yaml.NewEncoder(os.Stdout)
		return enc.Encode(result)

	default:
		printSynthesisResult(result)
	}

	// Save result
	if !GetDryRun() {
		resultPath := filepath.Join(findingsDir, "result.json")
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal result: %w", err)
		}
		if err := os.WriteFile(resultPath, data, 0600); err != nil {
			return fmt.Errorf("write result: %w", err)
		}
		VerbosePrintf("Result saved to: %s\n", resultPath)
	}

	return nil
}

func loadAgentFindings(dir string) ([]AgentFindings, error) {
	var reports []AgentFindings

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		// Skip the plan and result files
		if file.Name() == "plan.json" || file.Name() == "result.json" {
			continue
		}

		path := filepath.Join(dir, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			VerbosePrintf("Warning: failed to read %s: %v\n", path, err)
			continue
		}

		var report AgentFindings
		if err := json.Unmarshal(data, &report); err != nil {
			VerbosePrintf("Warning: failed to parse %s: %v\n", path, err)
			continue
		}

		// Infer category from filename if not set
		if report.Category == "" {
			report.Category = strings.TrimSuffix(file.Name(), ".json")
		}

		reports = append(reports, report)
	}

	return reports, nil
}

func synthesizeFindings(planID string, reports []AgentFindings) *SynthesisResult {
	// Collect all findings
	var allFindings []orchestrator.Finding
	for _, report := range reports {
		allFindings = append(allFindings, report.Findings...)
	}

	// Deduplicate findings by similarity
	merged := deduplicateFindings(allFindings)

	// Sort by severity (highest first)
	sort.Slice(merged, func(i, j int) bool {
		return orchestrator.SeverityOrder[merged[i].Severity] > orchestrator.SeverityOrder[merged[j].Severity]
	})

	// Count by severity
	var criticalCount, highCount, mediumCount, lowCount int
	for _, f := range merged {
		switch f.Severity {
		case orchestrator.SeverityCritical:
			criticalCount++
		case orchestrator.SeverityHigh:
			highCount++
		case orchestrator.SeverityMedium:
			mediumCount++
		case orchestrator.SeverityLow:
			lowCount++
		}
	}

	// Determine verdict (single-veto for CRITICAL)
	verdict := calculateVerdict(merged)
	grade := verdictToGrade(verdict, criticalCount, highCount)

	// Generate summary
	summary := synthesizeSummary(merged, verdict, criticalCount, highCount, mediumCount, lowCount)

	return &SynthesisResult{
		PlanID:        planID,
		Verdict:       verdict,
		Grade:         grade,
		Findings:      merged,
		CriticalCount: criticalCount,
		HighCount:     highCount,
		MediumCount:   mediumCount,
		LowCount:      lowCount,
		Summary:       summary,
		AgentReports:  reports,
		CompletedAt:   time.Now(),
	}
}

func deduplicateFindings(findings []orchestrator.Finding) []orchestrator.Finding {
	if len(findings) == 0 {
		return findings
	}

	// Group by normalized title and category
	groups := make(map[string][]orchestrator.Finding)
	for _, f := range findings {
		key := fmt.Sprintf("%s:%s", f.Category, normalizeTitle(f.Title))
		groups[key] = append(groups[key], f)
	}

	// Merge each group, keeping highest severity
	var merged []orchestrator.Finding
	for _, group := range groups {
		best := group[0]
		for _, f := range group[1:] {
			if orchestrator.SeverityOrder[f.Severity] > orchestrator.SeverityOrder[best.Severity] {
				best = f
			}
			// Merge files
			best.Files = mergeStringSlices(best.Files, f.Files)
		}
		merged = append(merged, best)
	}

	return merged
}

func normalizeTitle(title string) string {
	// Simple normalization - lowercase and remove punctuation
	return strings.ToLower(strings.TrimSpace(title))
}

func mergeStringSlices(a, b []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range a {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	for _, s := range b {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

func calculateVerdict(findings []orchestrator.Finding) orchestrator.Severity {
	if len(findings) == 0 {
		return orchestrator.SeverityPass
	}

	// Single-veto: any CRITICAL means CRITICAL verdict
	for _, f := range findings {
		if f.Severity == orchestrator.SeverityCritical {
			return orchestrator.SeverityCritical
		}
	}

	// Check for HIGH
	for _, f := range findings {
		if f.Severity == orchestrator.SeverityHigh {
			return orchestrator.SeverityHigh
		}
	}

	// Check for MEDIUM
	for _, f := range findings {
		if f.Severity == orchestrator.SeverityMedium {
			return orchestrator.SeverityMedium
		}
	}

	return orchestrator.SeverityLow
}

func verdictToGrade(verdict orchestrator.Severity, criticalCount, highCount int) string {
	switch verdict {
	case orchestrator.SeverityPass:
		return "A"
	case orchestrator.SeverityLow:
		return "A-"
	case orchestrator.SeverityMedium:
		return "B"
	case orchestrator.SeverityHigh:
		if highCount > 3 {
			return "D"
		}
		return "C"
	case orchestrator.SeverityCritical:
		if criticalCount > 1 {
			return "F"
		}
		return "D"
	default:
		return "?"
	}
}

func synthesizeSummary(findings []orchestrator.Finding, verdict orchestrator.Severity, critical, high, medium, low int) string {
	if len(findings) == 0 {
		return "No issues found. Code passed all validation checks."
	}

	var parts []string

	if critical > 0 {
		parts = append(parts, fmt.Sprintf("%d CRITICAL", critical))
	}
	if high > 0 {
		parts = append(parts, fmt.Sprintf("%d HIGH", high))
	}
	if medium > 0 {
		parts = append(parts, fmt.Sprintf("%d MEDIUM", medium))
	}
	if low > 0 {
		parts = append(parts, fmt.Sprintf("%d LOW", low))
	}

	return fmt.Sprintf("Found %s issues. Verdict: %s", strings.Join(parts, ", "), verdict)
}

func printSynthesisResult(result *SynthesisResult) {
	fmt.Println()
	fmt.Printf("Validation Result: %s\n", result.PlanID)
	fmt.Println("═══════════════════════════════════")
	fmt.Println()

	// Grade display
	gradeColor := ""
	switch result.Grade[0] {
	case 'A':
		gradeColor = "[PASS]"
	case 'B':
		gradeColor = "[PASS]"
	case 'C':
		gradeColor = "[WARN]"
	case 'D':
		gradeColor = "[FAIL]"
	case 'F':
		gradeColor = "[FAIL]"
	}
	fmt.Printf("Grade: %s %s\n", result.Grade, gradeColor)
	fmt.Printf("Verdict: %s\n", result.Verdict)
	fmt.Println()

	// Counts
	fmt.Println("Issue Counts:")
	fmt.Printf("  CRITICAL: %d\n", result.CriticalCount)
	fmt.Printf("  HIGH:     %d\n", result.HighCount)
	fmt.Printf("  MEDIUM:   %d\n", result.MediumCount)
	fmt.Printf("  LOW:      %d\n", result.LowCount)
	fmt.Println()

	// Top findings
	if len(result.Findings) > 0 {
		fmt.Println("Findings:")
		fmt.Println("───────────────────────────────────")
		maxShow := 10
		if len(result.Findings) < maxShow {
			maxShow = len(result.Findings)
		}
		for i := 0; i < maxShow; i++ {
			f := result.Findings[i]
			fmt.Printf("[%s] %s\n", f.Severity, f.Title)
			if f.Description != "" {
				desc := f.Description
				if len(desc) > 100 {
					desc = desc[:97] + "..."
				}
				fmt.Printf("  %s\n", desc)
			}
			if len(f.Files) > 0 {
				fmt.Printf("  Files: %s\n", strings.Join(f.Files, ", "))
			}
			fmt.Println()
		}
		if len(result.Findings) > maxShow {
			fmt.Printf("... and %d more findings\n", len(result.Findings)-maxShow)
		}
	}

	// Summary
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Println(result.Summary)

	// Agent reports
	fmt.Println()
	fmt.Printf("Agent Reports: %d agents contributed\n", len(result.AgentReports))
	for _, report := range result.AgentReports {
		fmt.Printf("  - %s: %d findings\n", report.Category, len(report.Findings))
	}
}
