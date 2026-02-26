package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	notebookMemoryFile string
	notebookQuiet      bool
	notebookMaxLines   int
	notebookSource     string
	notebookSessionID  string
)

// notebookCmd is the parent for notebook-related subcommands.
var notebookCmd = &cobra.Command{
	Use:   "notebook",
	Short: "Manage the session notebook (MEMORY.md)",
}

// notebookUpdateCmd distills session insights into MEMORY.md.
var notebookUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update MEMORY.md with latest session insights",
	Long: `Reads the most recent session data and updates MEMORY.md with a "Last Session"
section containing summary, decisions, and next steps.

Source priority (--source auto): session artifacts (.agents/ao/sessions/*.jsonl)
first, then pending.jsonl as fallback. Use --source sessions or --source pending
to force a specific source.

Stays within the configured line budget (default 190) by pruning stale entries
from the longest sections first.`,
	RunE: runNotebookUpdate,
}

func init() {
	notebookUpdateCmd.Flags().StringVar(&notebookMemoryFile, "memory-file", "", "Path to MEMORY.md (auto-detected if omitted)")
	notebookUpdateCmd.Flags().BoolVar(&notebookQuiet, "quiet", false, "Suppress output (for hooks)")
	notebookUpdateCmd.Flags().IntVar(&notebookMaxLines, "max-lines", 190, "Maximum lines in MEMORY.md")
	notebookUpdateCmd.Flags().StringVar(&notebookSource, "source", "auto", "Source: auto|sessions|pending")
	notebookUpdateCmd.Flags().StringVar(&notebookSessionID, "session", "", "Specific session ID to update from")

	notebookCmd.AddCommand(notebookUpdateCmd)
}

// pendingEntry represents one entry from pending.jsonl (forge output).
type pendingEntry struct {
	SessionID      string   `json:"session_id"`
	Summary        string   `json:"summary"`
	Decisions      []string `json:"decisions,omitempty"`
	Knowledge      []string `json:"knowledge,omitempty"`
	QueuedAt       time.Time `json:"queued_at"`
}

// notebookSection represents a parsed section of MEMORY.md.
type notebookSection struct {
	Heading string
	Lines   []string // lines of content (not including the heading)
}

func runNotebookUpdate(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	// Step 1: Find MEMORY.md
	memoryFile := notebookMemoryFile
	if memoryFile == "" {
		memoryFile, err = findMemoryFile(cwd)
		if err != nil {
			if !notebookQuiet {
				fmt.Println("No MEMORY.md found — skipping notebook update.")
			}
			return nil // graceful skip, not an error
		}
	}

	// Step 2: Read latest session entry (source priority: --session > --source)
	var entry *pendingEntry
	if notebookSessionID != "" {
		entry, err = readSessionByID(cwd, notebookSessionID)
		if err != nil || entry == nil {
			if !notebookQuiet {
				VerbosePrintf("Session %s not found.\n", notebookSessionID)
			}
			return nil
		}
	} else {
		entry, err = resolveNotebookSource(cwd, notebookSource)
		if err != nil || entry == nil {
			if !notebookQuiet {
				VerbosePrintf("No session data — nothing to update.\n")
			}
			return nil
		}
	}

	// Step 2.5: Skip if this entry was already processed (prevent replay)
	cursorPath := filepath.Join(cwd, ".agents", "ao", "notebook-cursor.json")
	if lastID, _ := readNotebookCursor(cursorPath); lastID == entry.SessionID && entry.SessionID != "" {
		VerbosePrintf("Session %s already processed — skipping.\n", entry.SessionID)
		return nil
	}

	// Step 3: Read current MEMORY.md
	sections, err := parseNotebookSections(memoryFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("parse MEMORY.md: %w", err)
	}

	// Step 4: Build "Last Session" section
	lastSession := buildLastSessionSection(entry)

	// Step 5: Replace or insert "Last Session" at the top (after title)
	sections = upsertLastSession(sections, lastSession)

	// Step 6: Prune to stay under budget
	sections = pruneNotebook(sections, notebookMaxLines)

	// Step 7: Write back atomically
	content := renderNotebook(sections)
	if err := atomicWriteFile(memoryFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("write MEMORY.md: %w", err)
	}

	// Step 8: Record cursor so we don't replay this entry
	if entry.SessionID != "" {
		_ = writeNotebookCursor(cursorPath, entry.SessionID)
	}

	if !notebookQuiet {
		lineCount := strings.Count(content, "\n")
		fmt.Printf("Updated %s (%d lines)\n", memoryFile, lineCount)
	}

	return nil
}

// findMemoryFile locates the Claude Code project MEMORY.md for the given cwd.
func findMemoryFile(cwd string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}

	// Convention: ~/.claude/projects/-<path-with-dashes>/memory/MEMORY.md
	normalizedPath := strings.ReplaceAll(cwd, "/", "-")
	memoryFile := filepath.Join(homeDir, ".claude", "projects", normalizedPath, "memory", "MEMORY.md")
	if _, err := os.Stat(memoryFile); err == nil {
		return memoryFile, nil
	}

	// Fallback: search for any matching project memory directory
	projectsDir := filepath.Join(homeDir, ".claude", "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return "", fmt.Errorf("no MEMORY.md found")
	}

	// Prefer exact suffix match first (avoids matching "cli" in unrelated projects)
	lastComponent := filepath.Base(cwd)
	suffix := "-" + lastComponent
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), suffix) {
			candidate := filepath.Join(projectsDir, e.Name(), "memory", "MEMORY.md")
			if _, err := os.Stat(candidate); err == nil {
				return candidate, nil
			}
		}
	}
	return "", fmt.Errorf("no MEMORY.md found for %s", cwd)
}

// readLatestPendingEntry reads the most recent entry from pending.jsonl.
func readLatestPendingEntry(cwd string) (*pendingEntry, error) {
	pendingPath := filepath.Join(cwd, ".agents", "ao", "pending.jsonl")
	f, err := os.Open(pendingPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var latest *pendingEntry
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 256*1024), 256*1024) // 256KB line buffer
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry pendingEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue // skip malformed lines
		}
		latest = &entry // last valid entry wins (most recent)
	}

	return latest, scanner.Err()
}

// sessionEntry is used to unmarshal session JSONL files from .agents/ao/sessions/.
type sessionEntry struct {
	SessionID string    `json:"session_id"`
	Date      time.Time `json:"date"`
	Summary   string    `json:"summary,omitempty"`
	Decisions []string  `json:"decisions,omitempty"`
	Knowledge []string  `json:"knowledge,omitempty"`
}

func (s *sessionEntry) toPendingEntry() *pendingEntry {
	return &pendingEntry{
		SessionID: s.SessionID,
		Summary:   s.Summary,
		Decisions: s.Decisions,
		Knowledge: s.Knowledge,
		QueuedAt:  s.Date,
	}
}

// resolveNotebookSource reads session data using the specified source strategy.
func resolveNotebookSource(cwd string, source string) (*pendingEntry, error) {
	switch source {
	case "sessions":
		return readLatestSessionEntry(cwd)
	case "pending":
		return readLatestPendingEntry(cwd)
	case "auto":
		entry, err := readLatestSessionEntry(cwd)
		if entry != nil && err == nil {
			return entry, nil
		}
		return readLatestPendingEntry(cwd)
	default:
		return nil, fmt.Errorf("unknown source: %s (use auto, sessions, or pending)", source)
	}
}

// readLatestSessionEntry reads the most recent session from .agents/ao/sessions/*.jsonl.
func readLatestSessionEntry(cwd string) (*pendingEntry, error) {
	sessionsDir := filepath.Join(cwd, ".agents", "ao", "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return nil, err
	}

	// Filter for .jsonl files and find the latest (date-prefixed names sort chronologically)
	var latest string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) != ".jsonl" {
			continue
		}
		if e.Name() > latest {
			latest = e.Name()
		}
	}
	if latest == "" {
		return nil, fmt.Errorf("no session files found in %s", sessionsDir)
	}

	return readSessionFile(filepath.Join(sessionsDir, latest))
}

// readSessionByID finds and reads a specific session by ID prefix.
// Detects ambiguous matches and returns an error if multiple files match.
func readSessionByID(cwd string, id string) (*pendingEntry, error) {
	sessionsDir := filepath.Join(cwd, ".agents", "ao", "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return nil, err
	}

	var matches []string
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".jsonl" {
			continue
		}
		if strings.Contains(e.Name(), id) {
			matches = append(matches, e.Name())
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("session %s not found", id)
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("ambiguous session ID %s matches %d files", id, len(matches))
	}
	return readSessionFile(filepath.Join(sessionsDir, matches[0]))
}

// readSessionFile reads a single session JSONL file and maps it to pendingEntry.
// Session JSONL files are single-record (one JSON object per file).
func readSessionFile(path string) (*pendingEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 256*1024), 256*1024)
	if scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			return nil, fmt.Errorf("empty session file: %s", path)
		}
		var entry sessionEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, fmt.Errorf("parse session file %s: %w", path, err)
		}
		return entry.toPendingEntry(), nil
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read session file %s: %w", path, err)
	}
	return nil, fmt.Errorf("empty session file: %s", path)
}

// parseNotebookSections reads MEMORY.md and splits it into sections by ## headings.
func parseNotebookSections(path string) ([]notebookSection, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return parseSectionsFromString(string(data)), nil
}

// parseSectionsFromString splits markdown content into sections by ## headings.
func parseSectionsFromString(content string) []notebookSection {
	trimmed := strings.TrimRight(content, "\n")
	if trimmed == "" {
		return nil
	}
	lines := strings.Split(trimmed, "\n")
	var sections []notebookSection
	var current *notebookSection

	for i, line := range lines {
		// Only treat lines as headings if they start with # or ## AND
		// are preceded by a blank line (or are the first line)
		prevBlank := i == 0 || strings.TrimSpace(lines[i-1]) == ""
		if (strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "# ")) && prevBlank {
			if current != nil {
				sections = append(sections, *current)
			}
			current = &notebookSection{
				Heading: line,
				Lines:   nil,
			}
		} else if current != nil {
			current.Lines = append(current.Lines, line)
		} else {
			// Lines before the first heading go into a preamble section
			if len(sections) == 0 {
				current = &notebookSection{Heading: "", Lines: nil}
			}
			if current != nil {
				current.Lines = append(current.Lines, line)
			}
		}
	}
	if current != nil {
		sections = append(sections, *current)
	}

	return sections
}

// buildLastSessionSection creates the "Last Session" content from a pending entry.
func buildLastSessionSection(entry *pendingEntry) notebookSection {
	var lines []string

	date := entry.QueuedAt.Format("2006-01-02")
	if entry.QueuedAt.IsZero() {
		date = time.Now().Format("2006-01-02")
	}

	lines = append(lines, fmt.Sprintf("- **Date:** %s", date))

	if entry.Summary != "" {
		summary := truncateText(entry.Summary, 200)
		lines = append(lines, fmt.Sprintf("- **Summary:** %s", summary))
	}

	if len(entry.Decisions) > 0 {
		lines = append(lines, "- **Key decisions:**")
		for _, d := range entry.Decisions {
			lines = append(lines, fmt.Sprintf("  - %s", truncateText(d, 150)))
		}
	}

	// Categorize knowledge items
	var worked, next, other []string
	for _, k := range entry.Knowledge {
		kLower := strings.ToLower(k)
		switch {
		case strings.HasPrefix(kLower, "next:") || strings.HasPrefix(kLower, "next ") ||
			strings.HasPrefix(kLower, "todo:") || strings.HasPrefix(kLower, "todo ") ||
			strings.HasPrefix(kLower, "follow-up:") || strings.HasPrefix(kLower, "follow-up "):
			next = append(next, k)
		case strings.HasPrefix(kLower, "worked:") || strings.HasPrefix(kLower, "worked ") ||
			strings.HasPrefix(kLower, "success:") || strings.HasPrefix(kLower, "success ") ||
			strings.HasPrefix(kLower, "resolved:") || strings.HasPrefix(kLower, "resolved "):
			worked = append(worked, k)
		default:
			other = append(other, k)
		}
	}

	if len(worked) > 0 {
		lines = append(lines, "- **What worked:**")
		for _, w := range worked {
			lines = append(lines, fmt.Sprintf("  - %s", truncateText(w, 150)))
		}
	}

	if len(next) > 0 {
		lines = append(lines, "- **Next:**")
		for _, n := range next {
			lines = append(lines, fmt.Sprintf("  - %s", truncateText(n, 150)))
		}
	}

	if len(other) > 0 {
		lines = append(lines, "- **Insights:**")
		for _, o := range other {
			lines = append(lines, fmt.Sprintf("  - %s", truncateText(o, 150)))
		}
	}

	lines = append(lines, "")

	return notebookSection{
		Heading: "## Last Session",
		Lines:   lines,
	}
}

// upsertLastSession replaces an existing "Last Session" section or inserts one
// right after the title (first # heading or preamble).
func upsertLastSession(sections []notebookSection, lastSession notebookSection) []notebookSection {
	// Find and replace existing Last Session, removing any duplicates
	for i, s := range sections {
		if s.Heading == "## Last Session" {
			sections[i] = lastSession
			// Remove any further duplicates
			result := sections[:i+1]
			for _, rem := range sections[i+1:] {
				if rem.Heading != "## Last Session" {
					result = append(result, rem)
				}
			}
			return result
		}
	}

	// Insert after the first section (title/preamble)
	if len(sections) == 0 {
		return []notebookSection{lastSession}
	}

	result := make([]notebookSection, 0, len(sections)+1)
	result = append(result, sections[0])
	result = append(result, lastSession)
	result = append(result, sections[1:]...)
	return result
}

// pruneNotebook trims the longest sections to stay under the line budget.
// Iteration is capped to prevent runaway loops when sections resist pruning.
func pruneNotebook(sections []notebookSection, maxLines int) []notebookSection {
	const maxIterations = 100
	iteration := 0
	for totalLines(sections) > maxLines && iteration < maxIterations {
		iteration++
		// Find the longest content section (skip preamble and Last Session)
		longestIdx := -1
		longestLen := 0
		for i, s := range sections {
			if s.Heading == "" || s.Heading == "## Last Session" {
				continue // don't prune preamble or Last Session
			}
			if strings.HasPrefix(s.Heading, "# ") && !strings.HasPrefix(s.Heading, "## ") {
				continue // don't prune the title
			}
			contentLen := len(s.Lines)
			if contentLen > longestLen {
				longestLen = contentLen
				longestIdx = i
			}
		}

		if longestIdx < 0 || longestLen == 0 {
			break // truly nothing left to prune
		}

		// Remove the last non-empty line from the longest section
		lines := sections[longestIdx].Lines
		for j := len(lines) - 1; j >= 0; j-- {
			if strings.TrimSpace(lines[j]) != "" {
				sections[longestIdx].Lines = append(lines[:j], lines[j+1:]...)
				break
			}
		}
	}

	return sections
}

// totalLines counts the total lines across all sections.
func totalLines(sections []notebookSection) int {
	count := 0
	for _, s := range sections {
		if s.Heading != "" {
			count++ // heading line
		}
		count += len(s.Lines)
	}
	return count
}

// renderNotebook produces the final MEMORY.md content from sections.
func renderNotebook(sections []notebookSection) string {
	var b strings.Builder
	for _, s := range sections {
		if s.Heading != "" {
			b.WriteString(s.Heading)
			b.WriteString("\n")
		}
		for _, line := range s.Lines {
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
	// Trim trailing whitespace to ensure idempotency on re-parse
	return strings.TrimRight(b.String(), "\n") + "\n"
}

// readNotebookCursor reads the last-processed session ID from the cursor file.
func readNotebookCursor(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var cursor struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(data, &cursor); err != nil {
		return "", err
	}
	return cursor.SessionID, nil
}

// writeNotebookCursor writes the last-processed session ID to the cursor file.
func writeNotebookCursor(path string, sessionID string) error {
	data, err := json.Marshal(struct {
		SessionID string `json:"session_id"`
		UpdatedAt string `json:"updated_at"`
	}{
		SessionID: sessionID,
		UpdatedAt: time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}
