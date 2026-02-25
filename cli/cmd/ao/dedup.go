package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

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
	Long: `Scan learnings for near-duplicates using normalized content hashing.

Reads all learning files (.md and .jsonl), extracts body content,
normalizes (lowercase, collapse whitespace, strip markdown formatting),
and groups by SHA256 hash. Groups with more than one member are duplicates.

Examples:
  ao dedup
  ao dedup --json`,
	RunE: runDedup,
}

func init() {
	dedupCmd.GroupID = "knowledge"
	rootCmd.AddCommand(dedupCmd)
}

func runDedup(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	learningsDir := filepath.Join(cwd, ".agents", "learnings")
	if _, err := os.Stat(learningsDir); os.IsNotExist(err) {
		fmt.Println("No learnings directory found.")
		return nil
	}

	// Collect all learning files
	var files []string
	jsonlFiles, _ := filepath.Glob(filepath.Join(learningsDir, "*.jsonl"))
	mdFiles, _ := filepath.Glob(filepath.Join(learningsDir, "*.md"))
	files = append(files, jsonlFiles...)
	files = append(files, mdFiles...)

	if len(files) == 0 {
		fmt.Println("No learning files found.")
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
