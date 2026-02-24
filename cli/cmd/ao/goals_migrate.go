package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/boshu2/agentops/cli/internal/goals"
)

var migrateToMD bool

func init() {
	migrateCmd := &cobra.Command{
		Use:     "migrate",
		Short:   "Migrate goals to latest format",
		Aliases: []string{"mg"},
		GroupID: "management",
		Long: `Migrate goals between formats.

Without flags, migrates GOALS.yaml from version 1 to version 2:
  - Sets version to 2
  - Adds mission field if missing
  - Sets goal type to "health" for goals without a type
  - Backs up original to GOALS.yaml.v1.bak

With --to-md, converts GOALS.yaml to GOALS.md (version 4):
  - Carries over mission and all gates
  - Groups goals by pillar to generate directives
  - Adds default north/anti stars
  - Preserves original YAML file

Examples:
  ao goals migrate                       # v1 YAML → v2 YAML
  ao goals migrate --to-md               # YAML → GOALS.md
  ao goals migrate --to-md --file g.yaml # Custom source file`,
		RunE: runGoalsMigrate,
	}
	migrateCmd.Flags().BoolVar(&migrateToMD, "to-md", false, "Convert GOALS.yaml to GOALS.md format")
	goalsCmd.AddCommand(migrateCmd)
}

func runGoalsMigrate(cmd *cobra.Command, args []string) error {
	path := goalsFile

	// Read and parse the file (LoadGoals now accepts v1)
	gf, err := goals.LoadGoals(path)
	if err != nil {
		return fmt.Errorf("load goals: %w", err)
	}

	if migrateToMD {
		if gf.Format == "md" {
			fmt.Println("Already in GOALS.md format — no migration needed.")
			return nil
		}
		gf.Format = "md"
		gf.Version = 4
		if gf.Mission == "" {
			gf.Mission = "Project fitness goals"
		}
		// Generate directives from pillar groups
		if len(gf.Directives) == 0 {
			gf.Directives = directivesFromPillars(gf.Goals)
		}
		// Add default north/anti stars if empty
		if len(gf.NorthStars) == 0 {
			gf.NorthStars = []string{
				"Every check passes before changes reach users",
				"Validation catches regressions automatically",
			}
		}
		if len(gf.AntiStars) == 0 {
			gf.AntiStars = []string{
				"Untested changes reaching main",
				"Goals that are trivially true or test implementation details",
			}
		}
		content := goals.RenderGoalsMD(gf)
		mdPath := filepath.Join(filepath.Dir(path), "GOALS.md")
		if err := os.WriteFile(mdPath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("writing GOALS.md: %w", err)
		}
		fmt.Printf("Migrated %s → %s (GOALS.md format, version 4)\n", path, mdPath)
		fmt.Println("Original YAML file preserved. Delete it manually when ready.")
		return nil
	}

	if gf.Version >= 2 {
		fmt.Printf("%s is already version %d — no migration needed.\n", path, gf.Version)
		return nil
	}

	// Backup original
	backupPath := path + ".v1.bak"
	original, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read original for backup: %w", err)
	}
	if err := os.WriteFile(backupPath, original, 0o644); err != nil {
		return fmt.Errorf("write backup: %w", err)
	}
	fmt.Printf("Backed up original to %s\n", backupPath)

	// Apply migration
	goals.MigrateV1ToV2(gf)

	// Write migrated file
	out, err := yaml.Marshal(gf)
	if err != nil {
		return fmt.Errorf("marshal migrated goals: %w", err)
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return fmt.Errorf("write migrated goals: %w", err)
	}

	fmt.Printf("Migrated %s from version 1 to version 2.\n", path)
	return nil
}

// directivesFromPillars generates directives from existing goal pillar groupings.
// Each unique pillar becomes a directive. Goals without a pillar are skipped.
// If no pillars exist, returns a single generic directive.
func directivesFromPillars(gs []goals.Goal) []goals.Directive {
	seen := map[string]bool{}
	var pillars []string
	for _, g := range gs {
		p := g.Pillar
		if p == "" {
			continue
		}
		if !seen[p] {
			seen[p] = true
			pillars = append(pillars, p)
		}
	}
	if len(pillars) == 0 {
		return []goals.Directive{
			{Number: 1, Title: "Improve project quality", Description: "Focus on the highest-impact improvements.", Steer: "increase"},
		}
	}
	dirs := make([]goals.Directive, len(pillars))
	for i, p := range pillars {
		dirs[i] = goals.Directive{
			Number:      i + 1,
			Title:       "Strengthen " + p,
			Description: fmt.Sprintf("Improve goals in the %s pillar.", p),
			Steer:       "increase",
		}
	}
	return dirs
}
