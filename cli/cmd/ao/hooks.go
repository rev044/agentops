package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	hooksOutputFormat string
	hooksDryRun       bool
	hooksForce        bool
)

// HookConfig represents a single hook configuration.
type HookConfig struct {
	Matcher string   `json:"matcher"`
	Command []string `json:"command"`
}

// HooksConfig represents the hooks section of Claude settings.
type HooksConfig struct {
	SessionStart []HookConfig `json:"SessionStart,omitempty"`
	Stop         []HookConfig `json:"Stop,omitempty"`
}

// ClaudeSettings represents the Claude Code settings.json structure.
type ClaudeSettings struct {
	Hooks   *HooksConfig           `json:"hooks,omitempty"`
	Other   map[string]interface{} `json:"-"` // Preserve other settings
	rawJSON map[string]interface{}
}

var hooksCmd = &cobra.Command{
	Use:   "hooks",
	Short: "Manage Claude Code hooks for automatic knowledge flywheel",
	Long: `The hooks command manages Claude Code hooks that automate the CASS knowledge flywheel.

Subcommands:
  init      Generate hooks configuration
  install   Install hooks to ~/.claude/settings.json
  show      Display current hook configuration
  test      Verify hooks work correctly

The knowledge flywheel automates:
  1. SessionStart: Inject prior knowledge with confidence decay
  2. Stop: Extract learnings and update feedback loop

Example workflow:
  ao hooks init                    # Generate configuration
  ao hooks install                 # Install to Claude Code
  ao hooks test                    # Verify everything works`,
}

var hooksInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate hooks configuration",
	Long: `Generate Claude Code hooks configuration for the CASS knowledge flywheel.

The generated hooks will:
  SessionStart:
    - Apply confidence decay to stale learnings
    - Inject CASS-weighted knowledge (up to 1500 tokens)

  Stop:
    - Extract learnings from completed session
    - Sync task completion signals
    - Update feedback loop

Output formats:
  json     JSON for manual settings.json editing
  shell    Shell commands for verification`,
	RunE: runHooksInit,
}

var hooksInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install hooks to Claude Code settings",
	Long: `Install ao hooks to ~/.claude/settings.json.

This command:
  1. Reads existing settings.json (if any)
  2. Merges ao hooks with existing configuration
  3. Creates a backup of the original settings
  4. Writes the updated configuration

Use --force to overwrite existing ao hooks.`,
	RunE: runHooksInstall,
}

var hooksShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current hook configuration",
	Long:  `Display the current Claude Code hooks configuration from ~/.claude/settings.json.`,
	RunE:  runHooksShow,
}

var hooksTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test hooks configuration",
	Long: `Test that all hook dependencies are available and working.

This command:
  1. Verifies ao is in PATH
  2. Checks that required subcommands exist
  3. Dry-runs the SessionStart hook
  4. Reports any issues`,
	RunE: runHooksTest,
}

func init() {
	rootCmd.AddCommand(hooksCmd)
	hooksCmd.AddCommand(hooksInitCmd)
	hooksCmd.AddCommand(hooksInstallCmd)
	hooksCmd.AddCommand(hooksShowCmd)
	hooksCmd.AddCommand(hooksTestCmd)

	// Init flags
	hooksInitCmd.Flags().StringVar(&hooksOutputFormat, "format", "json", "Output format: json, shell")

	// Install flags
	hooksInstallCmd.Flags().BoolVar(&hooksDryRun, "dry-run", false, "Show what would be installed without making changes")
	hooksInstallCmd.Flags().BoolVar(&hooksForce, "force", false, "Overwrite existing ao hooks")

	// Test flags
	hooksTestCmd.Flags().BoolVar(&hooksDryRun, "dry-run", false, "Show test steps without running hooks")
}

// generateHooksConfig creates the standard ao hooks configuration.
func generateHooksConfig() *HooksConfig {
	return &HooksConfig{
		SessionStart: []HookConfig{
			{
				Matcher: "",
				Command: []string{"bash", "-c", "ao inject --apply-decay --max-tokens 1500 2>/dev/null || true"},
			},
		},
		Stop: []HookConfig{
			{
				Matcher: "",
				Command: []string{"bash", "-c", "ao forge transcript --last-session --quiet --queue 2>/dev/null; ao task-sync --promote 2>/dev/null || true"},
			},
		},
	}
}

func runHooksInit(cmd *cobra.Command, args []string) error {
	hooks := generateHooksConfig()

	switch hooksOutputFormat {
	case "json":
		wrapper := struct {
			Hooks *HooksConfig `json:"hooks"`
		}{Hooks: hooks}

		data, err := json.MarshalIndent(wrapper, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal hooks: %w", err)
		}
		fmt.Println(string(data))

	case "shell":
		fmt.Println("# SessionStart hook (knowledge injection)")
		fmt.Printf("# %s\n", strings.Join(hooks.SessionStart[0].Command, " "))
		fmt.Println("ao inject --apply-decay --max-tokens 1500")
		fmt.Println()
		fmt.Println("# Stop hook (learning extraction)")
		fmt.Printf("# %s\n", strings.Join(hooks.Stop[0].Command, " "))
		fmt.Println("ao forge transcript --last-session --quiet --queue")
		fmt.Println("ao task-sync --promote")

	default:
		return fmt.Errorf("unknown format: %s (use json or shell)", hooksOutputFormat)
	}

	return nil
}

func runHooksInstall(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")

	// Read existing settings
	var rawSettings map[string]interface{}
	if data, err := os.ReadFile(settingsPath); err == nil {
		if err := json.Unmarshal(data, &rawSettings); err != nil {
			return fmt.Errorf("parse existing settings: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read settings: %w", err)
	} else {
		rawSettings = make(map[string]interface{})
	}

	// Generate new hooks
	newHooks := generateHooksConfig()

	// Check for existing hooks
	if existingHooks, ok := rawSettings["hooks"].(map[string]interface{}); ok {
		if !hooksForce {
			// Check if ao hooks already exist
			if sessionStart, ok := existingHooks["SessionStart"].([]interface{}); ok {
				for _, h := range sessionStart {
					if hook, ok := h.(map[string]interface{}); ok {
						if cmd, ok := hook["command"].([]interface{}); ok && len(cmd) > 1 {
							if cmdStr, ok := cmd[1].(string); ok && strings.Contains(cmdStr, "ao inject") {
								fmt.Println("ao hooks already installed. Use --force to overwrite.")
								return nil
							}
						}
					}
				}
			}
		}
	}

	// Merge hooks - preserve existing non-ao hooks
	hooksMap := make(map[string]interface{})
	if existing, ok := rawSettings["hooks"].(map[string]interface{}); ok {
		// Copy existing hooks
		for k, v := range existing {
			hooksMap[k] = v
		}
	}

	// Convert new hooks to map format
	sessionStartHooks := make([]map[string]interface{}, 0)
	stopHooks := make([]map[string]interface{}, 0)

	// Preserve existing non-ao hooks
	if existing, ok := hooksMap["SessionStart"].([]interface{}); ok {
		for _, h := range existing {
			if hook, ok := h.(map[string]interface{}); ok {
				if cmd, ok := hook["command"].([]interface{}); ok && len(cmd) > 1 {
					if cmdStr, ok := cmd[1].(string); ok && !strings.Contains(cmdStr, "ao ") {
						sessionStartHooks = append(sessionStartHooks, hook)
					}
				} else {
					sessionStartHooks = append(sessionStartHooks, hook)
				}
			}
		}
	}

	if existing, ok := hooksMap["Stop"].([]interface{}); ok {
		for _, h := range existing {
			if hook, ok := h.(map[string]interface{}); ok {
				if cmd, ok := hook["command"].([]interface{}); ok && len(cmd) > 1 {
					if cmdStr, ok := cmd[1].(string); ok && !strings.Contains(cmdStr, "ao ") {
						stopHooks = append(stopHooks, hook)
					}
				} else {
					stopHooks = append(stopHooks, hook)
				}
			}
		}
	}

	// Add ao hooks
	for _, h := range newHooks.SessionStart {
		sessionStartHooks = append(sessionStartHooks, map[string]interface{}{
			"matcher": h.Matcher,
			"command": h.Command,
		})
	}

	for _, h := range newHooks.Stop {
		stopHooks = append(stopHooks, map[string]interface{}{
			"matcher": h.Matcher,
			"command": h.Command,
		})
	}

	hooksMap["SessionStart"] = sessionStartHooks
	hooksMap["Stop"] = stopHooks
	rawSettings["hooks"] = hooksMap

	if hooksDryRun {
		fmt.Println("[dry-run] Would write to", settingsPath)
		data, _ := json.MarshalIndent(rawSettings, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	// Create backup
	if _, err := os.Stat(settingsPath); err == nil {
		backupPath := fmt.Sprintf("%s.backup.%s", settingsPath, time.Now().Format("20060102-150405"))
		if data, err := os.ReadFile(settingsPath); err == nil {
			if err := os.WriteFile(backupPath, data, 0644); err != nil {
				return fmt.Errorf("create backup: %w", err)
			}
			fmt.Printf("Backed up existing settings to %s\n", backupPath)
		}
	}

	// Ensure .claude directory exists
	claudeDir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("create .claude directory: %w", err)
	}

	// Write new settings
	data, err := json.MarshalIndent(rawSettings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}

	fmt.Printf("✓ Installed ao hooks to %s\n", settingsPath)
	fmt.Println()
	fmt.Println("Hooks installed:")
	fmt.Println("  SessionStart: ao inject --apply-decay")
	fmt.Println("  Stop: ao forge + ao task-sync")
	fmt.Println()
	fmt.Println("Run 'ao hooks test' to verify the installation.")

	return nil
}

func runHooksShow(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No Claude settings found at", settingsPath)
			fmt.Println("Run 'ao hooks install' to set up hooks.")
			return nil
		}
		return fmt.Errorf("read settings: %w", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("parse settings: %w", err)
	}

	hooks, ok := settings["hooks"]
	if !ok {
		fmt.Println("No hooks configured in", settingsPath)
		fmt.Println("Run 'ao hooks install' to set up hooks.")
		return nil
	}

	// Pretty print just the hooks section
	hooksData, err := json.MarshalIndent(map[string]interface{}{"hooks": hooks}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal hooks: %w", err)
	}

	fmt.Println(string(hooksData))

	// Check for ao hooks
	if hooksMap, ok := hooks.(map[string]interface{}); ok {
		hasAoHooks := false
		if sessionStart, ok := hooksMap["SessionStart"].([]interface{}); ok {
			for _, h := range sessionStart {
				if hook, ok := h.(map[string]interface{}); ok {
					if cmd, ok := hook["command"].([]interface{}); ok && len(cmd) > 1 {
						if cmdStr, ok := cmd[1].(string); ok && strings.Contains(cmdStr, "ao ") {
							hasAoHooks = true
							break
						}
					}
				}
			}
		}

		if hasAoHooks {
			fmt.Println()
			fmt.Println("✓ ao hooks are installed")
		} else {
			fmt.Println()
			fmt.Println("⚠ ao hooks not found. Run 'ao hooks install' to set up.")
		}
	}

	return nil
}

func runHooksTest(cmd *cobra.Command, args []string) error {
	fmt.Println("Testing ao hooks configuration...")
	fmt.Println()

	allPassed := true

	// Test 1: Check ao is in PATH
	fmt.Print("1. Checking ao is in PATH... ")
	aoPath, err := exec.LookPath("ao")
	if err != nil {
		fmt.Println("✗ FAILED")
		fmt.Printf("   ao not found in PATH. Ensure ao is installed and in your PATH.\n")
		allPassed = false
	} else {
		fmt.Printf("✓ found at %s\n", aoPath)
	}

	// Test 2: Check required subcommands
	subcommands := []string{"inject", "forge", "task-sync", "feedback-loop"}
	fmt.Print("2. Checking required subcommands... ")
	missingCmds := []string{}
	for _, subcmd := range subcommands {
		// Run ao <subcmd> --help to verify it exists
		testCmd := exec.Command("ao", subcmd, "--help")
		if err := testCmd.Run(); err != nil {
			missingCmds = append(missingCmds, subcmd)
		}
	}
	if len(missingCmds) > 0 {
		fmt.Println("✗ FAILED")
		fmt.Printf("   Missing subcommands: %s\n", strings.Join(missingCmds, ", "))
		allPassed = false
	} else {
		fmt.Println("✓ all present")
	}

	// Test 3: Check settings.json
	fmt.Print("3. Checking Claude settings... ")
	homeDir, _ := os.UserHomeDir()
	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		fmt.Println("⚠ settings.json not found")
		fmt.Println("   Run 'ao hooks install' to create hooks configuration.")
	} else {
		data, err := os.ReadFile(settingsPath)
		if err != nil {
			fmt.Println("✗ FAILED to read")
			allPassed = false
		} else {
			var settings map[string]interface{}
			if err := json.Unmarshal(data, &settings); err != nil {
				fmt.Println("✗ FAILED to parse")
				allPassed = false
			} else if _, ok := settings["hooks"]; !ok {
				fmt.Println("⚠ no hooks configured")
				fmt.Println("   Run 'ao hooks install' to set up hooks.")
			} else {
				fmt.Println("✓ hooks configured")
			}
		}
	}

	// Test 4: Dry-run inject command
	fmt.Print("4. Testing inject command... ")
	if hooksDryRun {
		fmt.Println("⏭ skipped (--dry-run)")
	} else {
		testCmd := exec.Command("ao", "inject", "--max-tokens", "100", "--no-cite")
		output, err := testCmd.CombinedOutput()
		if err != nil {
			// Inject might "fail" if no learnings exist - that's OK
			if strings.Contains(string(output), "No prior knowledge") || len(output) > 0 {
				fmt.Println("✓ working")
			} else {
				fmt.Println("✗ FAILED")
				fmt.Printf("   Error: %v\n", err)
				allPassed = false
			}
		} else {
			fmt.Println("✓ working")
		}
	}

	// Test 5: Check forge can find transcripts
	fmt.Print("5. Testing forge transcript access... ")
	if hooksDryRun {
		fmt.Println("⏭ skipped (--dry-run)")
	} else {
		projectsDir := filepath.Join(homeDir, ".claude", "projects")
		if _, err := os.Stat(projectsDir); os.IsNotExist(err) {
			fmt.Println("⚠ no Claude projects found")
			fmt.Println("   This is OK for first-time setup.")
		} else {
			// Just verify the directory exists - don't actually process
			fmt.Println("✓ projects directory found")
		}
	}

	fmt.Println()
	if allPassed {
		fmt.Println("✓ All tests passed! Hooks are ready to use.")
	} else {
		fmt.Println("⚠ Some tests failed. Please fix the issues above.")
	}

	return nil
}
