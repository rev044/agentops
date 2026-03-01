package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/boshu2/agentops/cli/internal/goals"
	"github.com/spf13/cobra"
)

// Valid steer values for directives.
var validSteers = map[string]bool{
	"increase": true,
	"decrease": true,
	"hold":     true,
	"explore":  true,
}

var goalsSteerCmd = &cobra.Command{
	Use:     "steer",
	Short:   "Manage directives",
	GroupID: "management",
}

// --- steer add ---

var (
	steerAddDescription string
	steerAddSteer       string
)

var goalsSteerAddCmd = &cobra.Command{
	Use:   "add <title>",
	Short: "Add a new directive",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]

		if !validSteers[steerAddSteer] {
			return fmt.Errorf("invalid steer value %q (valid: increase, decrease, hold, explore)", steerAddSteer)
		}

		gf, resolvedPath, err := loadMDGoals()
		if err != nil {
			return err
		}

		// Determine next directive number
		maxNum := 0
		for _, d := range gf.Directives {
			if d.Number > maxNum {
				maxNum = d.Number
			}
		}

		newDirective := goals.Directive{
			Number:      maxNum + 1,
			Title:       title,
			Description: steerAddDescription,
			Steer:       steerAddSteer,
		}

		gf.Directives = append(gf.Directives, newDirective)

		if goalsJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(newDirective)
		}

		if dryRun {
			fmt.Printf("Would add directive #%d: %s\n", newDirective.Number, newDirective.Title)
			return nil
		}

		if err := writeMDGoals(gf, resolvedPath); err != nil {
			return err
		}

		fmt.Printf("Added directive #%d: %s (steer: %s)\n", newDirective.Number, newDirective.Title, newDirective.Steer)
		return nil
	},
}

// --- steer remove ---

var goalsSteerRemoveCmd = &cobra.Command{
	Use:   "remove <number>",
	Short: "Remove a directive by number",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		num, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("directive number must be an integer: %w", err)
		}

		gf, resolvedPath, err := loadMDGoals()
		if err != nil {
			return err
		}

		// Find and remove the directive
		found := false
		var remaining []goals.Directive
		for _, d := range gf.Directives {
			if d.Number == num {
				found = true
				continue
			}
			remaining = append(remaining, d)
		}

		if !found {
			return fmt.Errorf("directive #%d not found", num)
		}

		// Renumber remaining directives sequentially
		for i := range remaining {
			remaining[i].Number = i + 1
		}
		gf.Directives = remaining

		if goalsJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(gf.Directives)
		}

		if dryRun {
			fmt.Printf("Would remove directive #%d and renumber %d remaining\n", num, len(remaining))
			return nil
		}

		if err := writeMDGoals(gf, resolvedPath); err != nil {
			return err
		}

		fmt.Printf("Removed directive #%d, renumbered %d remaining\n", num, len(remaining))
		return nil
	},
}

// --- steer prioritize ---

var goalsSteerPrioritizeCmd = &cobra.Command{
	Use:   "prioritize <number> <new-position>",
	Short: "Move a directive to a new position",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		num, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("directive number must be an integer: %w", err)
		}
		newPos, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("new position must be an integer: %w", err)
		}

		gf, resolvedPath, err := loadMDGoals()
		if err != nil {
			return err
		}

		if len(gf.Directives) == 0 {
			return fmt.Errorf("no directives to prioritize")
		}

		if newPos < 1 || newPos > len(gf.Directives) {
			return fmt.Errorf("new position must be between 1 and %d", len(gf.Directives))
		}

		// Find the directive to move
		srcIdx := -1
		for i, d := range gf.Directives {
			if d.Number == num {
				srcIdx = i
				break
			}
		}
		if srcIdx < 0 {
			return fmt.Errorf("directive #%d not found", num)
		}

		// Remove from current position
		moving := gf.Directives[srcIdx]
		directives := make([]goals.Directive, 0, len(gf.Directives))
		directives = append(directives, gf.Directives[:srcIdx]...)
		directives = append(directives, gf.Directives[srcIdx+1:]...)

		// Insert at new position (1-indexed)
		insertIdx := newPos - 1
		if insertIdx > len(directives) {
			insertIdx = len(directives)
		}

		result := make([]goals.Directive, 0, len(gf.Directives))
		result = append(result, directives[:insertIdx]...)
		result = append(result, moving)
		result = append(result, directives[insertIdx:]...)

		// Renumber
		for i := range result {
			result[i].Number = i + 1
		}
		gf.Directives = result

		if goalsJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(gf.Directives)
		}

		if dryRun {
			fmt.Printf("Would move directive %q to position %d\n", moving.Title, newPos)
			return nil
		}

		if err := writeMDGoals(gf, resolvedPath); err != nil {
			return err
		}

		fmt.Printf("Moved directive %q to position %d\n", moving.Title, newPos)
		return nil
	},
}

// loadMDGoals loads goals and validates the format is markdown.
func loadMDGoals() (*goals.GoalFile, string, error) {
	resolved := resolveGoalsFile()
	resolvedPath := goals.ResolveGoalsPath(resolved)
	gf, err := goals.LoadGoals(resolved)
	if err != nil {
		return nil, "", fmt.Errorf("loading goals: %w", err)
	}

	if gf.Format != "md" {
		return nil, "", fmt.Errorf("directives require GOALS.md format; run 'ao goals migrate --to-md'")
	}

	return gf, resolvedPath, nil
}

// writeMDGoals renders and writes a GoalFile back to disk as GOALS.md.
func writeMDGoals(gf *goals.GoalFile, path string) error {
	content := goals.RenderGoalsMD(gf)

	// Ensure path ends in .md
	if strings.ToLower(filepath.Ext(path)) != ".md" {
		path = filepath.Join(filepath.Dir(path), "GOALS.md")
	}

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("writing goals file: %w", err)
	}
	return nil
}

func init() {
	// steer add flags
	goalsSteerAddCmd.Flags().StringVar(&steerAddDescription, "description", "", "Directive description (required)")
	_ = goalsSteerAddCmd.MarkFlagRequired("description")
	goalsSteerAddCmd.Flags().StringVar(&steerAddSteer, "steer", "increase", "Steer direction (increase, decrease, hold, explore)")

	// Register sub-subcommands
	goalsSteerCmd.AddCommand(goalsSteerAddCmd)
	goalsSteerCmd.AddCommand(goalsSteerRemoveCmd)
	goalsSteerCmd.AddCommand(goalsSteerPrioritizeCmd)

	// Register steer under goals
	goalsCmd.AddCommand(goalsSteerCmd)
}
