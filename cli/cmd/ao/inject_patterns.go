package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// collectPatterns finds patterns from .agents/patterns/
func collectPatterns(cwd, query string, limit int) ([]pattern, error) {
	patternsDir := filepath.Join(cwd, ".agents", "patterns")
	if _, err := os.Stat(patternsDir); os.IsNotExist(err) {
		// Try rig root
		patternsDir = findAgentsSubdir(cwd, "patterns")
		if patternsDir == "" {
			return nil, nil
		}
	}

	files, err := filepath.Glob(filepath.Join(patternsDir, "*.md"))
	if err != nil {
		return nil, err
	}

	// Sort by modification time
	sort.Slice(files, func(i, j int) bool {
		infoI, _ := os.Stat(files[i])
		infoJ, _ := os.Stat(files[j])
		if infoI == nil || infoJ == nil {
			return false
		}
		return infoI.ModTime().After(infoJ.ModTime())
	})

	var patterns []pattern
	queryLower := strings.ToLower(query)

	for _, file := range files {
		if len(patterns) >= limit {
			break
		}

		p, err := parsePatternFile(file)
		if err != nil {
			continue
		}

		// Filter by query
		if query != "" {
			content := strings.ToLower(p.Name + " " + p.Description)
			if !strings.Contains(content, queryLower) {
				continue
			}
		}

		patterns = append(patterns, p)
	}

	return patterns, nil
}

// parsePatternFile extracts pattern info from a markdown file
func parsePatternFile(path string) (pattern, error) {
	p := pattern{
		Name:     strings.TrimSuffix(filepath.Base(path), ".md"),
		FilePath: path,
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return p, err
	}

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Extract name from title
		if strings.HasPrefix(line, "# ") {
			p.Name = strings.TrimPrefix(line, "# ")
			continue
		}

		// First paragraph as description
		if p.Description == "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "---") && line != "" {
			desc := line
			for j := i + 1; j < len(lines) && j < i+2; j++ {
				nextLine := strings.TrimSpace(lines[j])
				if nextLine == "" || strings.HasPrefix(nextLine, "#") {
					break
				}
				desc += " " + nextLine
			}
			p.Description = truncateText(desc, 150)
			break
		}
	}

	return p, nil
}
