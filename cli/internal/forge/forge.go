// Package forge provides pure helpers for transcript and markdown knowledge extraction.
package forge

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	// SnippetMaxLength is the maximum length for extracted text snippets.
	SnippetMaxLength = 200
	// SummaryMaxLength is the maximum length for session summaries.
	SummaryMaxLength = 100
	// CharsPerToken is the rough estimate of characters per token.
	CharsPerToken = 4
)

// IssueIDPattern matches beads issue IDs like "ol-0001", "at-v123", "gt-abc-def".
var IssueIDPattern = regexp.MustCompile(`\b([a-z]{2,3})-([a-z0-9]{3,7}(?:-[a-z0-9]+)?)\b`)

// SessionUUIDPattern matches UUIDs in transcript file names.
var SessionUUIDPattern = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)

// SessionClaudeTranscriptPattern matches Claude session IDs like "ses_abc".
var SessionClaudeTranscriptPattern = regexp.MustCompile(`(ses_[A-Za-z0-9]+)`)

// CollectFilesFromPatterns expands glob patterns into a file list, optionally filtered.
func CollectFilesFromPatterns(patterns []string, matchFilter func(string) bool) ([]string, error) {
	var files []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, &PatternError{Pattern: pattern, Err: err}
		}
		if len(matches) == 0 {
			if _, err := os.Stat(pattern); err == nil {
				files = append(files, pattern)
			}
			continue
		}
		for _, match := range matches {
			if matchFilter == nil || matchFilter(match) {
				files = append(files, match)
			}
		}
	}
	return files, nil
}

// PatternError wraps a glob pattern parse error.
type PatternError struct {
	Pattern string
	Err     error
}

func (e *PatternError) Error() string { return "invalid pattern " + e.Pattern + ": " + e.Err.Error() }
func (e *PatternError) Unwrap() error { return e.Err }

// DetectSessionTypeFromContent infers session type from forged content.
func DetectSessionTypeFromContent(summary string, knowledge, decisions []string) string {
	combined := strings.ToLower(summary)
	for _, k := range knowledge {
		combined += " " + strings.ToLower(k)
	}
	for _, d := range decisions {
		combined += " " + strings.ToLower(d)
	}
	switch {
	case strings.Contains(combined, "career") || strings.Contains(combined, "interview") ||
		strings.Contains(combined, "resume") || strings.Contains(combined, "salary"):
		return "career"
	case strings.Contains(combined, "debug") || strings.Contains(combined, "stack trace") ||
		strings.Contains(combined, "broken") || strings.Contains(combined, "error log"):
		return "debug"
	case strings.Contains(combined, "brainstorm") || strings.Contains(combined, "what if") ||
		strings.Contains(combined, "option a vs"):
		return "brainstorm"
	case strings.Contains(combined, "research") || strings.Contains(combined, "explore"):
		return "research"
	case strings.Contains(combined, "go test") || strings.Contains(combined, "git commit") ||
		strings.Contains(combined, "implement") || strings.Contains(combined, "feat("):
		return "implement"
	default:
		return "general"
	}
}

// InferSessionIDFromPath extracts a session ID from a transcript file path.
func InferSessionIDFromPath(filePath string) string {
	base := filepath.Base(filePath)
	if match := SessionClaudeTranscriptPattern.FindStringSubmatch(base); len(match) > 1 {
		return match[1]
	}
	matches := SessionUUIDPattern.FindAllString(base, -1)
	if len(matches) == 0 {
		return ""
	}
	return matches[len(matches)-1]
}

// GenerateSummary creates a session summary from extracted content.
func GenerateSummary(decisions, knowledge []string, date time.Time) string {
	if len(decisions) > 0 {
		return TruncateString(decisions[0], SummaryMaxLength)
	}
	if len(knowledge) > 0 {
		return TruncateString(knowledge[0], SummaryMaxLength)
	}
	return "Session from " + date.Format("2006-01-02")
}

// CountLines quickly counts newlines in a file.
func CountLines(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer func() { _ = f.Close() }()
	buf := make([]byte, 64*1024)
	count := 0
	for {
		n, err := f.Read(buf)
		if n > 0 {
			for _, b := range buf[:n] {
				if b == '\n' {
					count++
				}
			}
		}
		if err != nil {
			break
		}
	}
	return count
}

// ExtractSnippet extracts a text snippet around a match index.
func ExtractSnippet(content string, startIdx, maxLen int) string {
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx >= len(content) {
		return ""
	}
	end := startIdx + maxLen
	if end > len(content) {
		end = len(content)
	}
	snippet := content[startIdx:end]
	if end < len(content) {
		if idx := LastSpaceIndex(snippet); idx > maxLen/2 {
			snippet = snippet[:idx]
		}
		snippet += "..."
	}
	return snippet
}

// LastSpaceIndex returns the index of the last space in s, or -1.
func LastSpaceIndex(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == ' ' {
			return i
		}
	}
	return -1
}

// ExtractIssueIDs finds issue IDs like "ol-0001" in content.
func ExtractIssueIDs(content string) []string {
	matches := IssueIDPattern.FindAllString(content, -1)
	if len(matches) == 0 {
		return nil
	}
	return matches
}

// TruncateString limits a string to maxLen characters with an ellipsis.
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Dedup removes duplicate strings preserving order.
func Dedup(items []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(items))
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

// SplitMarkdownSections splits markdown content by heading boundaries.
func SplitMarkdownSections(content string) []string {
	lines := strings.Split(content, "\n")
	var sections []string
	var current []string
	for _, line := range lines {
		if (strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "## ")) && len(current) > 0 {
			sections = append(sections, strings.Join(current, "\n"))
			current = nil
		}
		current = append(current, line)
	}
	if len(current) > 0 {
		sections = append(sections, strings.Join(current, "\n"))
	}
	if len(sections) == 0 {
		sections = []string{content}
	}
	return sections
}

// FileWithTime pairs a file path with its mod time.
type FileWithTime struct {
	Path    string
	ModTime time.Time
}

// IsTranscriptCandidate reports whether a file is a viable transcript file.
func IsTranscriptCandidate(path string, info os.FileInfo, projectsDir string) bool {
	if info.IsDir() || filepath.Ext(path) != ".jsonl" {
		return false
	}
	rel, _ := filepath.Rel(projectsDir, path)
	depth := len(filepath.SplitList(rel))
	return depth <= 3 && info.Size() > 100
}

// CollectTranscriptCandidates walks projectsDir and returns viable transcripts.
func CollectTranscriptCandidates(projectsDir string) ([]FileWithTime, error) {
	var candidates []FileWithTime
	err := filepath.Walk(projectsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && info.Name() == "subagents" {
			return filepath.SkipDir
		}
		if IsTranscriptCandidate(path, info, projectsDir) {
			candidates = append(candidates, FileWithTime{Path: path, ModTime: info.ModTime()})
		}
		return nil
	})
	return candidates, err
}
