package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/boshu2/agentops/cli/internal/storage"
)

var doctorJSON bool

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check AgentOps health",
	Long: `Run health checks on your AgentOps installation.

Validates that all required components are present and configured.
Optional components are reported as warnings but do not cause failure.

Examples:
  ao doctor
  ao doctor --json`,
	RunE: runDoctor,
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorJSON, "json", false, "Output results as JSON")
	rootCmd.AddCommand(doctorCmd)
}

type doctorCheck struct {
	Name     string `json:"name"`
	Status   string `json:"status"` // "pass", "warn", "fail"
	Detail   string `json:"detail"`
	Required bool   `json:"required"`
}

type doctorOutput struct {
	Checks  []doctorCheck `json:"checks"`
	Result  string        `json:"result"` // "HEALTHY", "DEGRADED", "UNHEALTHY"
	Summary string        `json:"summary"`
}

func runDoctor(cmd *cobra.Command, args []string) error {
	var checks []doctorCheck

	// 1. ao CLI version
	checks = append(checks, doctorCheck{
		Name:     "ao CLI",
		Status:   "pass",
		Detail:   fmt.Sprintf("v%s", version),
		Required: true,
	})

	// 2. Hooks installed
	checks = append(checks, checkHooksInstalled())

	// 3. .agents/ao directory
	checks = append(checks, checkKnowledgeBase())

	// 4. Plugin/skills presence
	checks = append(checks, checkSkills())

	// 5. Codex CLI (optional)
	checks = append(checks, checkOptionalCLI("codex", "needed for --mixed council"))

	// 6. bd CLI (optional)
	checks = append(checks, checkOptionalCLI("bd", "needed for issue tracking"))

	// 7. Knowledge pool health (optional)
	checks = append(checks, checkKnowledgePool())

	// Compute result
	output := computeResult(checks)

	if doctorJSON {
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	// Table output
	fmt.Println("AgentOps Health Check")
	fmt.Println("=====================")
	for _, c := range output.Checks {
		var icon string
		switch c.Status {
		case "pass":
			icon = "✓"
		case "warn":
			icon = "⚠"
		case "fail":
			icon = "✗"
		}
		fmt.Printf("%s %s: %s\n", icon, c.Name, c.Detail)
	}
	fmt.Println()
	fmt.Printf("Result: %s (%s)\n", output.Result, output.Summary)

	// Exit non-zero if any required checks failed
	for _, c := range output.Checks {
		if c.Required && c.Status == "fail" {
			os.Exit(1)
		}
	}

	return nil
}

func checkHooksInstalled() doctorCheck {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return doctorCheck{Name: "Hooks installed", Status: "fail", Detail: "cannot determine home directory", Required: true}
	}

	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return doctorCheck{Name: "Hooks installed", Status: "fail", Detail: "settings.json not found", Required: true}
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return doctorCheck{Name: "Hooks installed", Status: "fail", Detail: "settings.json parse error", Required: true}
	}

	hooksRaw, ok := settings["hooks"]
	if !ok {
		return doctorCheck{Name: "Hooks installed", Status: "fail", Detail: "no hooks configured", Required: true}
	}

	// Count hooks
	hookCount := 0
	if hooksMap, ok := hooksRaw.(map[string]interface{}); ok {
		for _, v := range hooksMap {
			if arr, ok := v.([]interface{}); ok {
				hookCount += len(arr)
			}
		}
	}

	return doctorCheck{
		Name:     "Hooks installed",
		Status:   "pass",
		Detail:   fmt.Sprintf("%d hooks configured", hookCount),
		Required: true,
	}
}

func checkKnowledgeBase() doctorCheck {
	cwd, err := os.Getwd()
	if err != nil {
		return doctorCheck{Name: "Knowledge base", Status: "fail", Detail: "cannot determine working directory", Required: true}
	}

	baseDir := filepath.Join(cwd, storage.DefaultBaseDir)
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return doctorCheck{Name: "Knowledge base", Status: "fail", Detail: ".agents/ao not initialized", Required: true}
	}

	return doctorCheck{Name: "Knowledge base", Status: "pass", Detail: ".agents/ao initialized", Required: true}
}

func checkSkills() doctorCheck {
	cwd, err := os.Getwd()
	if err != nil {
		return doctorCheck{Name: "Plugin", Status: "fail", Detail: "cannot determine working directory", Required: true}
	}

	skillsDir := filepath.Join(cwd, "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return doctorCheck{Name: "Plugin", Status: "fail", Detail: "skills directory not found", Required: true}
	}

	count := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		skillFile := filepath.Join(skillsDir, e.Name(), "SKILL.md")
		if _, err := os.Stat(skillFile); err == nil {
			count++
		}
	}

	if count == 0 {
		return doctorCheck{Name: "Plugin", Status: "fail", Detail: "no skills found", Required: true}
	}

	return doctorCheck{
		Name:     "Plugin",
		Status:   "pass",
		Detail:   fmt.Sprintf("%d skills found", count),
		Required: true,
	}
}

func checkOptionalCLI(name string, reason string) doctorCheck {
	_, err := exec.LookPath(name)
	if err != nil {
		return doctorCheck{
			Name:     strings.Title(name) + " CLI",
			Status:   "warn",
			Detail:   fmt.Sprintf("not found (optional — %s)", reason),
			Required: false,
		}
	}

	return doctorCheck{
		Name:     strings.Title(name) + " CLI",
		Status:   "pass",
		Detail:   "available",
		Required: false,
	}
}

func checkKnowledgePool() doctorCheck {
	cwd, err := os.Getwd()
	if err != nil {
		return doctorCheck{Name: "Knowledge pool", Status: "warn", Detail: "cannot determine working directory", Required: false}
	}

	learningsDir := filepath.Join(cwd, ".agents", "learnings")
	patternsDir := filepath.Join(cwd, ".agents", "patterns")

	learnings := countFiles(learningsDir)
	patterns := countFiles(patternsDir)

	if learnings == 0 && patterns == 0 {
		return doctorCheck{Name: "Knowledge pool", Status: "warn", Detail: "empty (no learnings or patterns yet)", Required: false}
	}

	return doctorCheck{
		Name:     "Knowledge pool",
		Status:   "pass",
		Detail:   fmt.Sprintf("%d learnings, %d patterns", learnings, patterns),
		Required: false,
	}
}

func countFiles(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() {
			count++
		}
	}
	return count
}

func computeResult(checks []doctorCheck) doctorOutput {
	fails := 0
	warns := 0
	for _, c := range checks {
		switch c.Status {
		case "fail":
			fails++
		case "warn":
			warns++
		}
	}

	result := "HEALTHY"
	var parts []string

	if fails > 0 {
		result = "UNHEALTHY"
		parts = append(parts, fmt.Sprintf("%d failed", fails))
	}
	if warns > 0 {
		if result == "HEALTHY" {
			result = "HEALTHY"
		}
		parts = append(parts, fmt.Sprintf("%d optional warnings", warns))
	}

	summary := "all checks passed"
	if len(parts) > 0 {
		summary = strings.Join(parts, ", ")
	}

	return doctorOutput{
		Checks:  checks,
		Result:  result,
		Summary: summary,
	}
}
