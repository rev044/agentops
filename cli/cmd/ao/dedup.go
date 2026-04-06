package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/boshu2/agentops/cli/internal/lifecycle"
	"github.com/spf13/cobra"
)

var dedupMerge bool

// DedupGroup is an alias for the lifecycle type, kept for cmd/ao compatibility.
type DedupGroup = lifecycle.DedupGroup

// DedupResult is an alias for the lifecycle type, kept for cmd/ao compatibility.
type DedupResult = lifecycle.DedupResult

var dedupCmd = &cobra.Command{
	Use:   "dedup",
	Short: "Detect near-duplicate learnings",
	Long: `Scan learnings and patterns for near-duplicates using normalized content hashing.

Reads all files (.md and .jsonl) from .agents/learnings/ and .agents/patterns/,
extracts body content, normalizes (lowercase, collapse whitespace, strip markdown
formatting), and groups by SHA256 hash. Groups with more than one member are
duplicates. The patterns directory is optional — if it does not exist, only
learnings are scanned.

With --merge, automatically resolves each duplicate group by keeping the file
with the highest utility (from YAML frontmatter or JSON) and archiving the
rest to .agents/archive/dedup/. Files without a utility field default to 0.5.

Examples:
  ao dedup
  ao dedup --json
  ao dedup --merge`,
	RunE: runDedup,
}

func init() {
	dedupCmd.Flags().BoolVar(&dedupMerge, "merge", false, "Auto-resolve duplicates: keep highest utility, archive the rest")
	dedupCmd.GroupID = "core"
	rootCmd.AddCommand(dedupCmd)
}

// Thin wrappers preserved for tests.
func collectDedupFiles(cwd string) ([]string, error) { return lifecycle.CollectDedupFiles(cwd) }
func groupByContentHash(files []string) map[string][]string {
	return lifecycle.GroupByContentHash(files)
}
func mergeDedupGroups(hashToFiles map[string][]string, cwd string, dryRun bool) error {
	return lifecycle.MergeDedupGroups(hashToFiles, cwd, dryRun)
}
func buildDedupResult(hashToFiles map[string][]string, totalFiles int, cwd string) DedupResult {
	return lifecycle.BuildDedupResult(hashToFiles, totalFiles, cwd)
}
func pickHighestUtility(files []string) (string, []string) {
	return lifecycle.PickHighestUtility(files)
}
func readUtilityFromFile(path string) float64 { return lifecycle.ReadUtilityFromFile(path) }
func readUtilityFromFrontmatter(text string, defaultVal float64) float64 {
	return lifecycle.ReadUtilityFromFrontmatter(text, defaultVal)
}
func readUtilityFromJSONL(text string, defaultVal float64) float64 {
	return lifecycle.ReadUtilityFromJSONL(text, defaultVal)
}
func extractLearningBody(path string) string  { return lifecycle.ExtractLearningBody(path) }
func extractMarkdownBody(text string) string  { return lifecycle.ExtractMarkdownBody(text) }
func extractJSONLBody(text string) string     { return lifecycle.ExtractJSONLBody(text) }
func hashNormalizedContent(body string) string {
	return lifecycle.HashNormalizedContent(body)
}

func runDedup(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	files, err := lifecycle.CollectDedupFiles(cwd)
	if err != nil {
		return err
	}
	if files == nil {
		fmt.Println("No learnings or patterns directory found.")
		return nil
	}
	if len(files) == 0 {
		fmt.Println("No learning or pattern files found.")
		return nil
	}

	hashToFiles := lifecycle.GroupByContentHash(files)
	result := lifecycle.BuildDedupResult(hashToFiles, len(files), cwd)

	if dedupMerge && result.DuplicateGroups > 0 {
		return lifecycle.MergeDedupGroups(hashToFiles, cwd, GetDryRun())
	}

	if GetOutput() == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	fmt.Printf("Dedup Scan Results\n")
	fmt.Printf("==================\n")
	fmt.Printf("Total files:       %d\n", result.TotalFiles)
	fmt.Printf("Unique content:    %d\n", result.UniqueContent)
	fmt.Printf("Duplicate groups:  %d\n", result.DuplicateGroups)
	fmt.Printf("Duplicate files:   %d\n", result.DuplicateFiles)

	if result.DuplicateGroups > 0 {
		fmt.Println("\nDuplicate Groups:")
		for _, g := range result.Groups {
			fmt.Printf("\n  Hash: %s (%d files)\n", g.Hash, g.Count)
			for _, f := range g.Files {
				fmt.Printf("    - %s\n", f)
			}
		}
	} else {
		fmt.Println("\nNo duplicates found.")
	}

	return nil
}
