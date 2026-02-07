package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/boshu2/agentops/cli/internal/ratchet"
)

func init() {
	checkSubCmd := &cobra.Command{
		Use:   "check <step>",
		Short: "Check if step gate is met",
		Long: `Check if prerequisites are satisfied for a workflow step.

Returns exit code 0 if gate passes, 1 if not.

Steps: research, pre-mortem, plan, implement, crank, vibe, post-mortem
Aliases: premortem, postmortem, autopilot, validate, review

Examples:
  ao ratchet check research
  ao ratchet check plan
  ao ratchet check implement || echo "Run /plan first"`,
		Args: cobra.ExactArgs(1),
		RunE: runRatchetCheck,
	}
	ratchetCmd.AddCommand(checkSubCmd)
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
