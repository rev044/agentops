package quality

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ParseUtilityFromMarkdown extracts utility from markdown front matter.
func ParseUtilityFromMarkdown(path string) float64 {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return 0
	}
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "---" {
			break
		}
		if strings.HasPrefix(line, "utility:") {
			var utility float64
			if _, parseErr := fmt.Sscanf(line, "utility: %f", &utility); parseErr == nil {
				return utility
			}
		}
	}
	return 0
}

// ParseUtilityFromJSONL extracts utility from the first line of a JSONL file.
func ParseUtilityFromJSONL(path string) float64 {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer func() {
		_ = f.Close()
	}()
	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		var data map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &data); err == nil {
			if utility, ok := data["utility"].(float64); ok {
				return utility
			}
		}
	}
	return 0
}

// ParseUtilityFromFile extracts utility value from JSONL or markdown front matter.
func ParseUtilityFromFile(path string) float64 {
	if strings.HasSuffix(path, ".md") {
		return ParseUtilityFromMarkdown(path)
	}
	return ParseUtilityFromJSONL(path)
}

// CollectUtilityValuesFromDir walks a directory and collects utility values from files.
func CollectUtilityValuesFromDir(dir string) []float64 {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	var values []float64
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".jsonl") && !strings.HasSuffix(path, ".md") {
			return nil
		}
		if u := ParseUtilityFromFile(path); u > 0 {
			values = append(values, u)
		}
		return nil
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to walk %s: %v\n", dir, err)
	}
	return values
}

// ComputeUtilityMetrics calculates MemRL utility statistics from a list of dirs.
func ComputeUtilityMetrics(dirs []string) UtilityStats {
	var utilities []float64
	for _, dir := range dirs {
		utilities = append(utilities, CollectUtilityValuesFromDir(dir)...)
	}
	return ComputeUtilityStats(utilities)
}

// RetroHasLearnings checks whether a retro markdown file contains a learnings section.
func RetroHasLearnings(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	text := string(content)
	return strings.Contains(text, "## Learnings") ||
		strings.Contains(text, "## Key Learnings") ||
		strings.Contains(text, "### Learnings")
}

// CountRetros counts retro artifacts and how many have associated learnings since a time.
func CountRetros(baseDir string, since time.Time) (total int, withLearnings int, err error) {
	retrosDir := filepath.Join(baseDir, ".agents", "retro")
	if _, statErr := os.Stat(retrosDir); os.IsNotExist(statErr) {
		return 0, 0, nil
	}

	if walkErr := filepath.Walk(retrosDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") || !info.ModTime().After(since) {
			return nil
		}
		total++
		if RetroHasLearnings(path) {
			withLearnings++
		}
		return nil
	}); walkErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to walk %s: %v\n", retrosDir, walkErr)
	}

	return total, withLearnings, nil
}

// CountNewArtifactsInDir counts artifacts created after a time in a specific directory.
func CountNewArtifactsInDir(dir string, since time.Time) (int, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return 0, nil
	}

	count := 0
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if info.ModTime().After(since) {
			count++
		}
		return nil
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to walk %s: %v\n", dir, err)
	}

	return count, nil
}

// CountNewArtifactsInDirs counts artifacts created since a time across multiple dirs.
func CountNewArtifactsInDirs(dirs []string, since time.Time) (int, error) {
	count := 0
	for _, dir := range dirs {
		c, _ := CountNewArtifactsInDir(dir, since)
		count += c
	}
	return count, nil
}

// IsKnowledgeFile returns true if path ends with .md or .jsonl.
func IsKnowledgeFile(path string) bool {
	return strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".jsonl")
}
