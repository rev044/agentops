package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var dedupMerge bool

// DedupGroup represents a set of duplicate learnings.
type DedupGroup struct {
	Hash  string   `json:"hash"`
	Count int      `json:"count"`
	Files []string `json:"files"`
}

// DedupResult is the output of the dedup scan.
type DedupResult struct {
	TotalFiles      int          `json:"total_files"`
	UniqueContent   int          `json:"unique_content"`
	DuplicateGroups int          `json:"duplicate_groups"`
	DuplicateFiles  int          `json:"duplicate_files"`
	Groups          []DedupGroup `json:"groups,omitempty"`
}

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
	qualityCmd.AddCommand(dedupCmd)
}

func runDedup(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	learningsDir := filepath.Join(cwd, ".agents", "learnings")
	patternsDir := filepath.Join(cwd, ".agents", "patterns")

	learningsExists := true
	if _, err := os.Stat(learningsDir); os.IsNotExist(err) {
		learningsExists = false
	}
	patternsExists := true
	if _, err := os.Stat(patternsDir); os.IsNotExist(err) {
		patternsExists = false
	}

	if !learningsExists && !patternsExists {
		fmt.Println("No learnings or patterns directory found.")
		return nil
	}

	// Collect all files from learnings and patterns directories
	var files []string
	for _, dir := range []string{learningsDir, patternsDir} {
		if _, statErr := os.Stat(dir); os.IsNotExist(statErr) {
			continue
		}
		jsonlFiles, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
		mdFiles, _ := filepath.Glob(filepath.Join(dir, "*.md"))
		files = append(files, jsonlFiles...)
		files = append(files, mdFiles...)
	}

	if len(files) == 0 {
		fmt.Println("No learning or pattern files found.")
		return nil
	}

	// Hash content and group
	hashToFiles := make(map[string][]string)
	for _, f := range files {
		body := extractLearningBody(f)
		if body == "" {
			continue
		}
		hash := hashNormalizedContent(body)
		hashToFiles[hash] = append(hashToFiles[hash], f)
	}

	// Build result
	result := DedupResult{
		TotalFiles:    len(files),
		UniqueContent: len(hashToFiles),
	}

	for hash, group := range hashToFiles {
		if len(group) > 1 {
			result.DuplicateGroups++
			result.DuplicateFiles += len(group)
			// Use relative paths for cleaner output
			relFiles := make([]string, len(group))
			for i, f := range group {
				rel, relErr := filepath.Rel(cwd, f)
				if relErr != nil {
					rel = f
				}
				relFiles[i] = rel
			}
			result.Groups = append(result.Groups, DedupGroup{
				Hash:  hash[:12], // Short hash for display
				Count: len(group),
				Files: relFiles,
			})
		}
	}

	// --merge: resolve duplicate groups by keeping highest utility
	if dedupMerge && result.DuplicateGroups > 0 {
		archiveDir := filepath.Join(cwd, ".agents", "archive", "dedup")

		if GetDryRun() {
			fmt.Println("Merge (dry-run):")
			for _, group := range hashToFiles {
				if len(group) <= 1 {
					continue
				}
				kept, archived := pickHighestUtility(group)
				keptRel, _ := filepath.Rel(cwd, kept)
				fmt.Printf("  Keep: %s (utility %.2f)\n", keptRel, readUtilityFromFile(kept))
				for _, a := range archived {
					aRel, _ := filepath.Rel(cwd, a)
					fmt.Printf("  [dry-run] Would archive: %s -> .agents/archive/dedup/%s\n", aRel, filepath.Base(a))
				}
			}
			return nil
		}

		if err := os.MkdirAll(archiveDir, 0o755); err != nil {
			return fmt.Errorf("create archive directory: %w", err)
		}

		fmt.Println("Merge Results:")
		for _, group := range hashToFiles {
			if len(group) <= 1 {
				continue
			}
			kept, archived := pickHighestUtility(group)
			keptRel, _ := filepath.Rel(cwd, kept)
			fmt.Printf("  Keep: %s (utility %.2f)\n", keptRel, readUtilityFromFile(kept))
			for _, a := range archived {
				aRel, _ := filepath.Rel(cwd, a)
				dst := filepath.Join(archiveDir, filepath.Base(a))
				if mvErr := os.Rename(a, dst); mvErr != nil {
					fmt.Fprintf(os.Stderr, "Error archiving %s: %v\n", aRel, mvErr)
					continue
				}
				fmt.Printf("  Archived: %s -> .agents/archive/dedup/%s\n", aRel, filepath.Base(a))
			}
		}
		return nil
	}

	// Output
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

// pickHighestUtility selects the file with the highest utility from a group.
// Returns the keeper and the list of files to archive.
func pickHighestUtility(files []string) (string, []string) {
	bestIdx := 0
	bestUtility := readUtilityFromFile(files[0])
	for i := 1; i < len(files); i++ {
		u := readUtilityFromFile(files[i])
		if u > bestUtility {
			bestUtility = u
			bestIdx = i
		}
	}
	kept := files[bestIdx]
	var archived []string
	for i, f := range files {
		if i != bestIdx {
			archived = append(archived, f)
		}
	}
	return kept, archived
}

// readUtilityFromFile reads the utility value from a file's metadata.
// For .md files, parses YAML frontmatter. For .jsonl files, parses the first JSON line.
// Returns 0.5 as default if utility is not found or cannot be parsed.
func readUtilityFromFile(path string) float64 {
	const defaultUtility = 0.5

	content, err := os.ReadFile(path)
	if err != nil {
		return defaultUtility
	}
	text := string(content)

	if strings.HasSuffix(path, ".jsonl") {
		return readUtilityFromJSONL(text, defaultUtility)
	}
	return readUtilityFromFrontmatter(text, defaultUtility)
}

// readUtilityFromFrontmatter parses YAML frontmatter for the utility field.
func readUtilityFromFrontmatter(text string, defaultVal float64) float64 {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return defaultVal
	}
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "---" {
			return defaultVal // reached end of frontmatter without finding utility
		}
		if strings.HasPrefix(line, "utility:") {
			valStr := strings.TrimSpace(strings.TrimPrefix(line, "utility:"))
			if u, parseErr := strconv.ParseFloat(valStr, 64); parseErr == nil {
				return u
			}
		}
	}
	return defaultVal
}

// readUtilityFromJSONL parses the first JSON line for the utility field.
func readUtilityFromJSONL(text string, defaultVal float64) float64 {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return defaultVal
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &data); err != nil {
		return defaultVal
	}
	switch v := data["utility"].(type) {
	case float64:
		return v
	case string:
		if u, parseErr := strconv.ParseFloat(v, 64); parseErr == nil {
			return u
		}
	}
	return defaultVal
}

// extractLearningBody extracts the body content from a learning file.
func extractLearningBody(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	text := string(content)

	if strings.HasSuffix(path, ".md") {
		return extractMarkdownBody(text)
	}
	return extractJSONLBody(text)
}

// extractMarkdownBody returns content after YAML frontmatter.
func extractMarkdownBody(text string) string {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return text // No frontmatter, entire content is body
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return strings.Join(lines[i+1:], "\n")
		}
	}
	return text // No closing ---, treat all as body
}

// extractJSONLBody extracts title or content from a JSONL first line.
func extractJSONLBody(text string) string {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return ""
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &data); err != nil {
		return ""
	}
	// Prefer content, fall back to title
	if content, ok := data["content"].(string); ok && content != "" {
		return content
	}
	if title, ok := data["title"].(string); ok {
		return title
	}
	return ""
}

// hashNormalizedContent normalizes and hashes content for dedup comparison.
func hashNormalizedContent(body string) string {
	// Normalize: lowercase, strip markdown, collapse whitespace
	s := strings.ToLower(strings.TrimSpace(body))
	// Strip common markdown formatting
	s = strings.ReplaceAll(s, "#", "")
	s = strings.ReplaceAll(s, "*", "")
	s = strings.ReplaceAll(s, "`", "")
	s = strings.ReplaceAll(s, "---", "")
	s = strings.Join(strings.Fields(s), " ")
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
