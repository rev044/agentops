package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/boshu2/agentops/cli/internal/goals"
	"github.com/spf13/cobra"
)

var goalsInitNonInteractive bool

var goalsInitCmd = &cobra.Command{
	Use:     "init",
	Short:   "Bootstrap a new GOALS.md file",
	GroupID: "management",
	RunE: func(cmd *cobra.Command, args []string) error {
		resolvedPath := goals.ResolveGoalsPath(goalsFile)

		// Check if the resolved path already exists
		if _, err := os.Stat(resolvedPath); err == nil {
			return fmt.Errorf("goals file already exists: %s", resolvedPath)
		}

		// Also check the literal goalsFile path if different
		if resolvedPath != goalsFile {
			if _, err := os.Stat(goalsFile); err == nil {
				return fmt.Errorf("goals file already exists: %s", goalsFile)
			}
		}

		var gf *goals.GoalFile

		if goalsInitNonInteractive {
			gf = buildDefaultGoalFile()
		} else {
			var err error
			gf, err = buildInteractiveGoalFile(os.Stdin)
			if err != nil {
				return fmt.Errorf("reading input: %w", err)
			}
		}

		// Auto-detect gates
		detectedGoals := detectGates()
		gf.Goals = append(gf.Goals, detectedGoals...)

		if goalsJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(gf)
		}

		content := goals.RenderGoalsMD(gf)

		// Ensure the output path ends in .md
		outPath := resolvedPath
		if strings.ToLower(filepath.Ext(outPath)) != ".md" {
			outPath = filepath.Join(filepath.Dir(outPath), "GOALS.md")
		}

		if dryRun {
			fmt.Printf("Would write %s:\n\n%s", outPath, content)
			return nil
		}

		if err := os.WriteFile(outPath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("writing goals file: %w", err)
		}

		fmt.Printf("Created %s with %d gates\n", outPath, len(gf.Goals))
		return nil
	},
}

func buildDefaultGoalFile() *goals.GoalFile {
	dirName := filepath.Base(currentDir())

	return &goals.GoalFile{
		Version: 4,
		Format:  "md",
		Mission: fmt.Sprintf("Fitness goals for %s", dirName),
		NorthStars: []string{
			"All checks pass on every commit",
		},
		AntiStars: []string{
			"Untested changes reaching main",
		},
		Directives: []goals.Directive{
			{
				Number:      1,
				Title:       "Establish baseline",
				Description: "Get all gates passing and maintain a green baseline.",
				Steer:       "increase",
			},
		},
	}
}

func buildInteractiveGoalFile(r *os.File) (*goals.GoalFile, error) {
	scanner := bufio.NewScanner(r)

	mission, err := prompt(scanner, "Mission (one sentence): ")
	if err != nil {
		return nil, err
	}
	if mission == "" {
		mission = fmt.Sprintf("Fitness goals for %s", filepath.Base(currentDir()))
	}

	northRaw, err := prompt(scanner, "North stars (comma-separated): ")
	if err != nil {
		return nil, err
	}
	northStars := splitCommaSeparated(northRaw)
	if len(northStars) == 0 {
		northStars = []string{"All checks pass on every commit"}
	}

	antiRaw, err := prompt(scanner, "Anti stars (comma-separated): ")
	if err != nil {
		return nil, err
	}
	antiStars := splitCommaSeparated(antiRaw)
	if len(antiStars) == 0 {
		antiStars = []string{"Untested changes reaching main"}
	}

	dirTitle, err := prompt(scanner, "First directive title: ")
	if err != nil {
		return nil, err
	}
	if dirTitle == "" {
		dirTitle = "Establish baseline"
	}

	dirDesc, err := prompt(scanner, "First directive description: ")
	if err != nil {
		return nil, err
	}
	if dirDesc == "" {
		dirDesc = "Get all gates passing and maintain a green baseline."
	}

	return &goals.GoalFile{
		Version:    4,
		Format:     "md",
		Mission:    mission,
		NorthStars: northStars,
		AntiStars:  antiStars,
		Directives: []goals.Directive{
			{
				Number:      1,
				Title:       dirTitle,
				Description: dirDesc,
				Steer:       "increase",
			},
		},
	}, nil
}

// detectGates checks for common project files and returns matching gate goals.
func detectGates() []goals.Goal {
	var detected []goals.Goal

	if _, err := os.Stat("go.mod"); err == nil {
		detected = append(detected, goals.Goal{
			ID:          "go-build",
			Description: "Go project builds cleanly",
			Check:       "cd cli && make build",
			Weight:      5,
			Type:        goals.GoalTypeHealth,
		})
		detected = append(detected, goals.Goal{
			ID:          "go-test",
			Description: "Go tests pass",
			Check:       "cd cli && make test",
			Weight:      5,
			Type:        goals.GoalTypeHealth,
		})
	}

	// Check for cli/go.mod as well (nested Go project)
	if _, err := os.Stat("cli/go.mod"); err == nil {
		// Only add if we didn't already detect top-level go.mod
		if _, err := os.Stat("go.mod"); err != nil {
			detected = append(detected, goals.Goal{
				ID:          "go-build",
				Description: "Go project builds cleanly",
				Check:       "cd cli && make build",
				Weight:      5,
				Type:        goals.GoalTypeHealth,
			})
			detected = append(detected, goals.Goal{
				ID:          "go-test",
				Description: "Go tests pass",
				Check:       "cd cli && make test",
				Weight:      5,
				Type:        goals.GoalTypeHealth,
			})
		}
	}

	if _, err := os.Stat("package.json"); err == nil {
		detected = append(detected, goals.Goal{
			ID:          "npm-test",
			Description: "npm tests pass",
			Check:       "npm test",
			Weight:      5,
			Type:        goals.GoalTypeHealth,
		})
	}

	if _, err := os.Stat("Cargo.toml"); err == nil {
		detected = append(detected, goals.Goal{
			ID:          "cargo-test",
			Description: "Cargo tests pass",
			Check:       "cargo test",
			Weight:      5,
			Type:        goals.GoalTypeHealth,
		})
	}

	if _, err := os.Stat("pyproject.toml"); err == nil {
		detected = append(detected, goals.Goal{
			ID:          "python-test",
			Description: "Python tests pass",
			Check:       "pytest",
			Weight:      5,
			Type:        goals.GoalTypeHealth,
		})
	}

	if _, err := os.Stat("Makefile"); err == nil {
		detected = append(detected, goals.Goal{
			ID:          "make-build",
			Description: "Make build succeeds",
			Check:       "make build",
			Weight:      5,
			Type:        goals.GoalTypeHealth,
		})
	}

	return detected
}

func prompt(scanner *bufio.Scanner, msg string) (string, error) {
	fmt.Print(msg)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", nil
}

func splitCommaSeparated(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func currentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "project"
	}
	return dir
}

func init() {
	goalsInitCmd.Flags().BoolVar(&goalsInitNonInteractive, "non-interactive", false, "Use defaults without prompting")
	goalsCmd.AddCommand(goalsInitCmd)
}
