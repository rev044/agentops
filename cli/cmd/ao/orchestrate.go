package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/boshu2/agentops/cli/internal/orchestrator"
)

var (
	orchestrateFiles  []string
	orchestratePlanID string
)

var orchestrateCmd = &cobra.Command{
	Use:   "orchestrate",
	Short: "Generate agent dispatch plan",
	Long: `Generate a dispatch plan for the /vibe skill to execute.

This implements the "bridge pattern" where the CLI generates the plan
and the skill (via Claude Code) invokes the agents.

The plan includes:
- Agent dispatches for each category (security, quality, architecture, etc)
- Prompts tailored to each agent's focus area
- Output paths for findings

Example:
  ao orchestrate --files "cli/cmd/ao/*.go"
  ao orchestrate --files "src/**/*.ts,src/**/*.tsx"

The plan is output as JSON for the skill to consume.`,
	RunE: runOrchestrate,
}

func init() {
	rootCmd.AddCommand(orchestrateCmd)

	orchestrateCmd.Flags().StringArrayVar(&orchestrateFiles, "files", nil, "Files to analyze (can be specified multiple times)")
	orchestrateCmd.Flags().StringVar(&orchestratePlanID, "plan-id", "", "Custom plan ID (auto-generated if not provided)")
}

func runOrchestrate(cmd *cobra.Command, args []string) error {
	if len(orchestrateFiles) == 0 && len(args) > 0 {
		// Accept files as positional args too
		orchestrateFiles = args
	}

	if len(orchestrateFiles) == 0 {
		return fmt.Errorf("at least one --files argument is required")
	}

	// Generate plan ID if not provided
	planID := orchestratePlanID
	if planID == "" {
		planID = fmt.Sprintf("vibe-%s", generateShortID())
	}

	// Get working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	// Create findings directory
	findingsDir := filepath.Join(cwd, ".agents", "ao", "findings", planID)
	if !GetDryRun() {
		if err := os.MkdirAll(findingsDir, 0700); err != nil {
			return fmt.Errorf("create findings directory: %w", err)
		}
	}

	// Build dispatch plan
	plan := buildDispatchPlan(planID, orchestrateFiles, findingsDir)

	// Output
	switch GetOutput() {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(plan)

	default:
		// Pretty print for human consumption
		printDispatchPlan(plan)
	}

	// Save plan to file for synthesis step
	if !GetDryRun() {
		planPath := filepath.Join(findingsDir, "plan.json")
		data, err := json.MarshalIndent(plan, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal plan: %w", err)
		}
		if err := os.WriteFile(planPath, data, 0600); err != nil {
			return fmt.Errorf("write plan: %w", err)
		}
		VerbosePrintf("Plan saved to: %s\n", planPath)
	}

	return nil
}

func buildDispatchPlan(planID string, files []string, findingsDir string) *orchestrator.DispatchPlan {
	config := orchestrator.DefaultDispatchConfig()

	// Create dispatches for each category
	var wave1 []orchestrator.AgentDispatch

	for _, pod := range orchestrator.PodCategories {
		agentConfig := orchestrator.GetAgentForCategory(pod.Category)
		prompt := buildAgentPrompt(pod.Category, files, pod.Prompt)

		dispatch := orchestrator.AgentDispatch{
			ID:           fmt.Sprintf("%s-%s", planID, pod.Category),
			Category:     pod.Category,
			SubagentType: agentConfig.SubagentType,
			Prompt:       prompt,
			Files:        files,
			OutputPath:   filepath.Join(findingsDir, pod.Category+".json"),
			Model:        config.Model,
		}

		wave1 = append(wave1, dispatch)
	}

	return &orchestrator.DispatchPlan{
		PlanID:      planID,
		Wave1:       wave1,
		FindingsDir: findingsDir,
		Created:     time.Now().Format(time.RFC3339),
		Config:      config,
	}
}

func buildAgentPrompt(category string, files []string, podPrompt string) string {
	focus := orchestrator.BuildPromptForCategory(category, files)

	filesStr := strings.Join(files, ", ")
	if len(filesStr) > 200 {
		filesStr = filesStr[:197] + "..."
	}

	return fmt.Sprintf(`You are a %s expert performing code validation.

FILES TO ANALYZE: %s

%s

%s

OUTPUT FORMAT:
Report your findings as a JSON object with this structure:
{
  "category": "%s",
  "findings": [
    {
      "id": "unique-id",
      "severity": "CRITICAL|HIGH|MEDIUM|LOW",
      "title": "Brief title",
      "description": "Detailed description",
      "files": ["affected/file.go"],
      "lines": [123],
      "recommendation": "How to fix"
    }
  ],
  "summary": "Brief summary of findings"
}

Focus on actionable, specific findings. Avoid false positives.
If you find no issues, return an empty findings array.`, category, filesStr, focus, podPrompt, category)
}

// generateShortID creates a short random ID.
func generateShortID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		// Fallback to time-based ID
		return fmt.Sprintf("%x", time.Now().UnixNano()%0xFFFFFFFF)
	}
	return hex.EncodeToString(b)
}

func printDispatchPlan(plan *orchestrator.DispatchPlan) {
	fmt.Println()
	fmt.Printf("Dispatch Plan: %s\n", plan.PlanID)
	fmt.Println("═══════════════════════════════════")
	fmt.Println()

	fmt.Printf("Findings Directory: %s\n", plan.FindingsDir)
	fmt.Printf("Created: %s\n", plan.Created)
	fmt.Println()

	fmt.Printf("Wave 1 Dispatches (%d agents):\n", len(plan.Wave1))
	fmt.Println("───────────────────────────────────")

	for _, dispatch := range plan.Wave1 {
		fmt.Printf("  [%s] %s\n", dispatch.Category, dispatch.SubagentType)
		fmt.Printf("    Output: %s\n", dispatch.OutputPath)
	}

	fmt.Println()
	fmt.Println("Next Steps:")
	fmt.Println("1. Invoke each agent using the Task tool with the prompts above")
	fmt.Println("2. Write findings to the specified output paths")
	fmt.Printf("3. Run 'ao synthesize --plan-id %s' to merge findings\n", plan.PlanID)
}
