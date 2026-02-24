package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/boshu2/agentops/cli/embedded"
	"github.com/boshu2/agentops/cli/internal/goals"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var goalsInitNonInteractive bool
var goalsInitTemplate string

// validTemplateNames lists the recognised --template values.
var validTemplateNames = []string{"go-cli", "python-lib", "web-app", "rust-cli", "generic"}

// goalTemplate is the YAML structure of an embedded template file.
type goalTemplate struct {
	Name        string             `yaml:"name"`
	Description string             `yaml:"description"`
	Directives  []string           `yaml:"directives"`
	Gates       []goalTemplateGate `yaml:"gates"`
}

// goalTemplateGate mirrors a single gate entry in a template YAML file.
type goalTemplateGate struct {
	ID          string `yaml:"id"`
	Description string `yaml:"description"`
	Check       string `yaml:"check"`
	Weight      int    `yaml:"weight"`
	Type        string `yaml:"type"`
}

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

		// Resolve the template name: explicit flag > auto-detect > ""
		projectRoot := filepath.Dir(resolvedPath)
		tmplName := goalsInitTemplate
		if tmplName == "" {
			tmplName = autoDetectTemplate(projectRoot)
		}

		// Load template gates (if any template was resolved).
		var tmpl *goalTemplate
		if tmplName != "" {
			var err error
			tmpl, err = loadTemplate(tmplName)
			if err != nil {
				return fmt.Errorf("loading template %q: %w", tmplName, err)
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

		// Populate gates: template gates take priority when a template is
		// loaded; otherwise fall back to the existing auto-detect logic.
		if tmpl != nil {
			gf.Goals = append(gf.Goals, templateGatesToGoals(tmpl)...)
		} else {
			detectedGoals := detectGates(projectRoot)
			gf.Goals = append(gf.Goals, detectedGoals...)
		}

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

// buildInteractiveGoalFile prompts the user for goal file fields via the given
// io.Reader. Accepting io.Reader (rather than *os.File) enables testing with
// strings.NewReader or bytes.Buffer without requiring real file descriptors.
func buildInteractiveGoalFile(r io.Reader) (*goals.GoalFile, error) {
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

// detectGates checks for common project files relative to projectRoot and
// returns matching gate goals. The projectRoot is derived from the resolved
// goals file path so that detection works correctly regardless of the current
// working directory.
func detectGates(projectRoot string) []goals.Goal {
	var detected []goals.Goal

	// stat is a helper that checks for a file relative to the project root.
	stat := func(rel string) bool {
		_, err := os.Stat(filepath.Join(projectRoot, rel))
		return err == nil
	}

	switch {
	case stat("cli/go.mod"):
		// Nested Go project in cli/ subdirectory
		detected = append(detected, goals.Goal{
			ID:          "go-build",
			Description: "Go project builds cleanly",
			Check:       "cd cli && go build ./...",
			Weight:      5,
			Type:        goals.GoalTypeHealth,
		})
		detected = append(detected, goals.Goal{
			ID:          "go-test",
			Description: "Go tests pass",
			Check:       "cd cli && go test ./...",
			Weight:      5,
			Type:        goals.GoalTypeHealth,
		})
	case stat("go.mod"):
		// Root-level Go project
		detected = append(detected, goals.Goal{
			ID:          "go-build",
			Description: "Go project builds cleanly",
			Check:       "go build ./...",
			Weight:      5,
			Type:        goals.GoalTypeHealth,
		})
		detected = append(detected, goals.Goal{
			ID:          "go-test",
			Description: "Go tests pass",
			Check:       "go test ./...",
			Weight:      5,
			Type:        goals.GoalTypeHealth,
		})
	}

	if stat("package.json") {
		detected = append(detected, goals.Goal{
			ID:          "npm-test",
			Description: "npm tests pass",
			Check:       "npm test",
			Weight:      5,
			Type:        goals.GoalTypeHealth,
		})
	}

	if stat("Cargo.toml") {
		detected = append(detected, goals.Goal{
			ID:          "cargo-test",
			Description: "Cargo tests pass",
			Check:       "cargo test",
			Weight:      5,
			Type:        goals.GoalTypeHealth,
		})
	}

	if stat("pyproject.toml") {
		detected = append(detected, goals.Goal{
			ID:          "python-test",
			Description: "Python tests pass",
			Check:       "pytest",
			Weight:      5,
			Type:        goals.GoalTypeHealth,
		})
	}

	if stat("Makefile") {
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

// loadTemplate reads a named template from the embedded TemplatesFS.
func loadTemplate(name string) (*goalTemplate, error) {
	path := filepath.Join("templates", name+".yaml")
	data, err := fs.ReadFile(embedded.TemplatesFS, path)
	if err != nil {
		return nil, fmt.Errorf("template %q not found: %w", name, err)
	}
	var tmpl goalTemplate
	if err := yaml.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("parsing template %q: %w", name, err)
	}
	return &tmpl, nil
}

// templateGatesToGoals converts template gates into goals.Goal values.
func templateGatesToGoals(tmpl *goalTemplate) []goals.Goal {
	out := make([]goals.Goal, 0, len(tmpl.Gates))
	for _, g := range tmpl.Gates {
		out = append(out, goals.Goal{
			ID:          g.ID,
			Description: g.Description,
			Check:       g.Check,
			Weight:      g.Weight,
			Type:        goals.GoalType(g.Type),
		})
	}
	return out
}

// autoDetectTemplate chooses a template name based on project marker files.
// Returns "" if no recognisable project type is found.
func autoDetectTemplate(projectRoot string) string {
	detected := detectTemplateFromProjectRoot(projectRoot)
	if detected == "generic" {
		return ""
	}
	return detected
}

func init() {
	goalsInitCmd.Flags().BoolVar(&goalsInitNonInteractive, "non-interactive", false, "Use defaults without prompting")
	goalsInitCmd.Flags().StringVar(&goalsInitTemplate, "template", "", "Goal template (go-cli, python-lib, web-app, rust-cli, generic)")
	goalsCmd.AddCommand(goalsInitCmd)
}
