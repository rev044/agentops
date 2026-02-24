package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/boshu2/agentops/cli/internal/goals"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type pruneResult struct {
	StaleGoals []staleGoal `json:"stale_goals"`
	Removed    int         `json:"removed"`
	DryRun     bool        `json:"dry_run"`
}

type staleGoal struct {
	ID    string `json:"id"`
	Check string `json:"check"`
	Path  string `json:"missing_path"`
}

var goalsPruneCmd = &cobra.Command{
	Use:     "prune",
	Aliases: []string{"p"},
	Short:   "Remove goals referencing nonexistent files",
	GroupID: "management",
	RunE: func(cmd *cobra.Command, args []string) error {
		resolvedPath := goals.ResolveGoalsPath(goalsFile)

		gf, err := goals.LoadGoals(goalsFile)
		if err != nil {
			return fmt.Errorf("loading goals: %w", err)
		}

		// Find stale goals — those referencing file paths that don't exist
		var stale []staleGoal
		staleIDs := make(map[string]bool)

		for _, g := range gf.Goals {
			missingPath := findMissingPath(g.Check)
			if missingPath != "" {
				stale = append(stale, staleGoal{
					ID:    g.ID,
					Check: g.Check,
					Path:  missingPath,
				})
				staleIDs[g.ID] = true
			}
		}

		result := pruneResult{
			StaleGoals: stale,
			DryRun:     dryRun,
		}

		if dryRun || len(stale) == 0 {
			result.Removed = 0

			if goalsJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			if len(stale) == 0 {
				fmt.Println("No stale goals found.")
				return nil
			}

			fmt.Printf("Found %d stale goal(s):\n", len(stale))
			for _, s := range stale {
				fmt.Printf("  %s: %s (missing: %s)\n", s.ID, s.Check, s.Path)
			}
			fmt.Println("\nRun without --dry-run to remove them.")
			return nil
		}

		// Remove stale goals
		var kept []goals.Goal
		for _, g := range gf.Goals {
			if !staleIDs[g.ID] {
				kept = append(kept, g)
			}
		}
		gf.Goals = kept
		result.Removed = len(stale)

		// Write back based on format
		if gf.Format == "md" {
			content := goals.RenderGoalsMD(gf)
			outPath := resolvedPath
			if strings.ToLower(filepath.Ext(outPath)) != ".md" {
				outPath = filepath.Join(filepath.Dir(outPath), "GOALS.md")
			}
			if err := os.WriteFile(outPath, []byte(content), 0o644); err != nil {
				return fmt.Errorf("writing goals file: %w", err)
			}
		} else {
			data, err := yaml.Marshal(gf)
			if err != nil {
				return fmt.Errorf("marshaling goals: %w", err)
			}
			if err := os.WriteFile(resolvedPath, data, 0o644); err != nil {
				return fmt.Errorf("writing goals file: %w", err)
			}
		}

		if goalsJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(result)
		}

		fmt.Printf("Pruned %d stale goal(s) from %s\n", result.Removed, resolvedPath)
		for _, s := range stale {
			fmt.Printf("  removed: %s (missing: %s)\n", s.ID, s.Path)
		}
		return nil
	},
}

// findMissingPath checks if a goal's check command references a file path
// that doesn't exist. Returns the missing path, or "" if all paths exist.
func findMissingPath(check string) string {
	parts := strings.Fields(check)
	for _, part := range parts {
		// Check for path-like strings: starts with scripts/, contains a slash
		// and looks like a file path (has an extension or is a known directory prefix)
		if strings.HasPrefix(part, "scripts/") ||
			strings.HasPrefix(part, "./scripts/") ||
			strings.HasPrefix(part, "tests/") ||
			strings.HasPrefix(part, "./tests/") ||
			strings.HasPrefix(part, "hooks/") ||
			strings.HasPrefix(part, "./hooks/") {
			// Strip any trailing shell operators
			cleanPath := strings.TrimRight(part, ";|&")
			if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
				return cleanPath
			}
		}
		// Also check any part that contains a path separator and has a file extension
		if strings.Contains(part, "/") && filepath.Ext(part) != "" {
			cleanPath := strings.TrimRight(part, ";|&")
			if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
				return cleanPath
			}
		}
	}
	return ""
}

func init() {
	goalsCmd.AddCommand(goalsPruneCmd)
}
