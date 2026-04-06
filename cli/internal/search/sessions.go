package search

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// CollectSessionFiles gathers .jsonl and .md files from the sessions directory,
// sorted by modification time (newest first).
// When both .jsonl and .md exist for the same stem, only the .jsonl is kept.
func CollectSessionFiles(sessionsDir string) ([]string, error) {
	jsonlFiles, err := filepath.Glob(filepath.Join(sessionsDir, "*.jsonl"))
	if err != nil {
		return nil, err
	}
	mdFiles, _ := filepath.Glob(filepath.Join(sessionsDir, "*.md"))

	stemSet := make(map[string]bool, len(jsonlFiles))
	for _, f := range jsonlFiles {
		stemSet[strings.TrimSuffix(f, ".jsonl")] = true
	}
	files := append([]string(nil), jsonlFiles...)
	for _, f := range mdFiles {
		stem := strings.TrimSuffix(f, ".md")
		if !stemSet[stem] {
			files = append(files, f)
		}
	}

	slices.SortFunc(files, func(a, b string) int {
		infoA, _ := os.Stat(a)
		infoB, _ := os.Stat(b)
		if infoA == nil || infoB == nil {
			return 0
		}
		return infoB.ModTime().Compare(infoA.ModTime())
	})
	return files, nil
}

// ParseJSONLSessionSummary reads the first line of a JSONL file and returns
// the truncated "summary" field value.
func ParseJSONLSessionSummary(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = f.Close()
	}()

	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		var data map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &data); err == nil {
			if summary, ok := data["summary"].(string); ok {
				return TruncateText(summary, 150), nil
			}
		}
	}
	return "", nil
}

// ParseMarkdownSessionSummary extracts the first content paragraph from a
// markdown file, skipping YAML frontmatter blocks and headings.
func ParseMarkdownSessionSummary(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	inFrontmatter := false
	frontmatterDone := false
	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			if !inFrontmatter && !frontmatterDone {
				inFrontmatter = true
				continue
			}
			if inFrontmatter {
				inFrontmatter = false
				frontmatterDone = true
				continue
			}
		}
		if inFrontmatter {
			continue
		}
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			return TruncateText(trimmed, 150), nil
		}
	}
	return "", nil
}

// ParseSessionFile extracts session summary from a file.
func ParseSessionFile(path string) (Session, error) {
	s := Session{Path: path}

	info, err := os.Stat(path)
	if err != nil {
		return s, err
	}
	s.Date = info.ModTime().Format("2006-01-02")

	if strings.HasSuffix(path, ".jsonl") {
		s.Summary, err = ParseJSONLSessionSummary(path)
	} else {
		s.Summary, err = ParseMarkdownSessionSummary(path)
	}

	return s, err
}
