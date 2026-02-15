package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/cobra"
)

var (
	rpiMaxCycles int
)

func init() {
	loopCmd := &cobra.Command{
		Use:   "loop [goal]",
		Short: "Run continuous RPI cycles from next-work queue",
		Long: `Execute RPI cycles in a loop, consuming from next-work.jsonl.

Each cycle spawns a fresh Claude session (Ralph Wiggum pattern):
  1. Read unconsumed items from .agents/rpi/next-work.jsonl
  2. Pick highest-severity item as goal (or use explicit goal)
  3. Spawn: claude -p '/rpi "<goal>" --spawn-next'
  4. Wait for completion
  5. Re-read next-work.jsonl (post-mortem may have harvested new items)
  6. Repeat until queue empty or max-cycles reached

Examples:
  ao rpi loop                          # consume from queue until stable
  ao rpi loop "improve test coverage"  # run one cycle with explicit goal
  ao rpi loop --max-cycles 3           # cap at 3 iterations
  ao rpi loop --dry-run                # show what would run`,
		Args: cobra.MaximumNArgs(1),
		RunE: runRPILoop,
	}

	loopCmd.Flags().IntVar(&rpiMaxCycles, "max-cycles", 0, "Maximum cycles (0 = unlimited, stop when queue empty)")

	rpiCmd.AddCommand(loopCmd)
}

// nextWorkEntry represents one line in next-work.jsonl.
type nextWorkEntry struct {
	SourceEpic string         `json:"source_epic"`
	Timestamp  string         `json:"timestamp"`
	Items      []nextWorkItem `json:"items"`
	Consumed   bool           `json:"consumed"`
	ConsumedBy *string        `json:"consumed_by"`
	ConsumedAt *string        `json:"consumed_at"`
}

// nextWorkItem represents a single harvested work item.
type nextWorkItem struct {
	Title       string `json:"title"`
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Source      string `json:"source"`
	Description string `json:"description"`
	Evidence    string `json:"evidence,omitempty"`
	TargetRepo  string `json:"target_repo,omitempty"`
}

func runRPILoop(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	// Check claude is available
	if _, err := exec.LookPath("claude"); err != nil {
		return fmt.Errorf("claude CLI not found on PATH (required for spawning RPI sessions)")
	}

	// Parse explicit goal if provided
	explicitGoal := ""
	if len(args) > 0 {
		explicitGoal = args[0]
	}

	nextWorkPath := filepath.Join(cwd, ".agents", "rpi", "next-work.jsonl")

	cycle := 0
	for {
		cycle++

		if rpiMaxCycles > 0 && cycle > rpiMaxCycles {
			fmt.Printf("\nReached max cycles (%d). Stopping.\n", rpiMaxCycles)
			break
		}

		fmt.Printf("\n=== RPI Loop: Cycle %d ===\n", cycle)

		// Determine goal for this cycle
		goal := explicitGoal
		if goal == "" {
			// Read queue for unconsumed items
			items, err := readUnconsumedItems(nextWorkPath, "")
			if err != nil {
				VerbosePrintf("Warning: %v\n", err)
			}

			if len(items) == 0 {
				fmt.Println("No unconsumed work in queue. Flywheel stable.")
				break
			}

			goal = selectHighestSeverityItem(items)
			fmt.Printf("From queue: %s\n", goal)
		}

		if goal == "" {
			fmt.Println("No goal and empty queue. Nothing to do.")
			break
		}

		// Build the /rpi command
		rpiArg := fmt.Sprintf(`/rpi "%s" --spawn-next`, goal)

		if GetDryRun() {
			fmt.Printf("[dry-run] Would spawn: claude -p '%s'\n", rpiArg)
			if explicitGoal == "" {
				fmt.Println("[dry-run] Queue not consumed in dry-run. Showing first cycle only.")
			}
			break
		} else {
			fmt.Printf("Spawning: claude -p '%s'\n", rpiArg)
			start := time.Now()

			if err := spawnClaudeRPI(rpiArg); err != nil {
				fmt.Printf("Cycle %d failed: %v\n", cycle, err)
				fmt.Println("Stopping loop. Fix the issue and re-run ao rpi loop.")
				return err
			}

			elapsed := time.Since(start).Round(time.Second)
			fmt.Printf("Cycle %d completed in %s\n", cycle, elapsed)
		}

		// If explicit goal was provided, only run once
		if explicitGoal != "" {
			fmt.Println("Explicit goal completed.")
			break
		}
	}

	fmt.Printf("\nRPI loop finished after %d cycle(s).\n", cycle-1)
	return nil
}

// readUnconsumedItems reads next-work.jsonl and returns all unconsumed items
// across all entries, flattened. When repoFilter is non-empty, items with a
// non-empty TargetRepo that is neither "*" nor equal to repoFilter are skipped.
// Items without a TargetRepo (legacy) or with TargetRepo=="*" always pass.
func readUnconsumedItems(path string, repoFilter string) ([]nextWorkItem, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No file = no items
		}
		return nil, fmt.Errorf("open next-work.jsonl: %w", err)
	}
	defer f.Close()

	var items []nextWorkItem
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var entry nextWorkEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			VerbosePrintf("Skipping malformed line: %v\n", err)
			continue
		}

		if !entry.Consumed && len(entry.Items) > 0 {
			for _, item := range entry.Items {
				if repoFilter != "" && item.TargetRepo != "" && item.TargetRepo != "*" && item.TargetRepo != repoFilter {
					continue
				}
				items = append(items, item)
			}
		}
	}

	return items, scanner.Err()
}

// selectHighestSeverityItem returns the title of the highest-severity item.
// Severity order: high > medium > low.
func selectHighestSeverityItem(items []nextWorkItem) string {
	if len(items) == 0 {
		return ""
	}

	sort.Slice(items, func(i, j int) bool {
		return severityRank(items[i].Severity) > severityRank(items[j].Severity)
	})

	return items[0].Title
}

func severityRank(s string) int {
	switch s {
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

// spawnClaudeRPI spawns a fresh Claude session with /rpi.
func spawnClaudeRPI(rpiArg string) error {
	cmd := exec.Command("claude", "-p", rpiArg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
