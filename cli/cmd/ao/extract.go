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

	"github.com/boshu2/agentops/cli/internal/storage"
)

// PendingExtraction represents a session queued for learning extraction.
type PendingExtraction struct {
	SessionID      string    `json:"session_id"`
	SessionPath    string    `json:"session_path"`
	TranscriptPath string    `json:"transcript_path"`
	Summary        string    `json:"summary"`
	Decisions      []string  `json:"decisions,omitempty"`
	Knowledge      []string  `json:"knowledge,omitempty"`
	QueuedAt       time.Time `json:"queued_at"`
}

var (
	extractMaxContent int
	extractClear      bool
)

var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Process pending learning extractions",
	Long: `Check for pending session extractions and output a prompt for Claude to process.

This command is designed to be called from a SessionStart hook. If there are
pending sessions (queued by 'ao forge --queue'), it outputs a structured prompt
that asks Claude to extract learnings and write them to .agents/learnings/.

The prompt includes:
  - Session summary and context
  - Key decisions and knowledge snippets
  - Clear instructions for Claude to extract 1-3 learnings
  - File path where learnings should be written

If no pending extractions exist, outputs nothing (silent).

Examples:
  ao extract                    # Check and output extraction prompt
  ao extract --clear            # Clear pending queue without processing
  ao extract --max-content 4000 # Limit content size`,
	RunE: runExtract,
}

func init() {
	rootCmd.AddCommand(extractCmd)
	extractCmd.Flags().IntVar(&extractMaxContent, "max-content", 3000, "Maximum characters of session content to include")
	extractCmd.Flags().BoolVar(&extractClear, "clear", false, "Clear pending queue without processing")
}

func runExtract(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	pendingPath := filepath.Join(cwd, storage.DefaultBaseDir, "pending.jsonl")

	// Check if pending file exists
	if _, err := os.Stat(pendingPath); os.IsNotExist(err) {
		return nil // No pending extractions, silent exit
	}

	// Read pending extractions
	pending, err := readPendingExtractions(pendingPath)
	if err != nil {
		return fmt.Errorf("read pending: %w", err)
	}

	if len(pending) == 0 {
		return nil // No pending extractions
	}

	// If --clear flag, just remove the file
	if extractClear {
		if err := os.Remove(pendingPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("clear pending: %w", err)
		}
		fmt.Printf("Cleared %d pending extraction(s)\n", len(pending))
		return nil
	}

	// Process the most recent pending extraction
	extraction := pending[len(pending)-1]

	// Output the extraction prompt for Claude
	outputExtractionPrompt(extraction, cwd, extractMaxContent)

	// Clear the pending file after outputting
	if err := os.Remove(pendingPath); err != nil && !os.IsNotExist(err) {
		// Non-fatal, just log
		VerbosePrintf("Warning: failed to clear pending file: %v\n", err)
	}

	return nil
}

func readPendingExtractions(path string) ([]PendingExtraction, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var pending []PendingExtraction
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var p PendingExtraction
		if err := json.Unmarshal([]byte(line), &p); err != nil {
			continue // Skip malformed lines
		}
		pending = append(pending, p)
	}

	return pending, scanner.Err()
}

func outputExtractionPrompt(extraction PendingExtraction, cwd string, maxContent int) {
	// Generate output file path
	date := extraction.QueuedAt.Format("2006-01-02")
	shortID := extraction.SessionID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	outputPath := filepath.Join(cwd, ".agents", "learnings", fmt.Sprintf("%s-%s.md", date, shortID))

	fmt.Println("---")
	fmt.Println("# Knowledge Extraction Request")
	fmt.Println()
	fmt.Println("A previous session has been queued for learning extraction. Please process it.")
	fmt.Println()
	fmt.Println("## Session Context")
	fmt.Println()
	fmt.Printf("- **Session ID**: %s\n", extraction.SessionID)
	fmt.Printf("- **Date**: %s\n", extraction.QueuedAt.Format("2006-01-02 15:04"))
	fmt.Printf("- **Summary**: %s\n", extraction.Summary)
	fmt.Println()

	// Include decisions if present
	if len(extraction.Decisions) > 0 {
		fmt.Println("## Key Decisions")
		fmt.Println()
		charCount := 0
		for i, d := range extraction.Decisions {
			if charCount > maxContent/2 {
				fmt.Printf("- ... and %d more\n", len(extraction.Decisions)-i)
				break
			}
			fmt.Printf("- %s\n", truncateForPrompt(d, 200))
			charCount += len(d)
		}
		fmt.Println()
	}

	// Include knowledge snippets if present
	if len(extraction.Knowledge) > 0 {
		fmt.Println("## Knowledge Snippets")
		fmt.Println()
		charCount := 0
		for i, k := range extraction.Knowledge {
			if charCount > maxContent/2 {
				fmt.Printf("- ... and %d more\n", len(extraction.Knowledge)-i)
				break
			}
			fmt.Printf("- %s\n", truncateForPrompt(k, 200))
			charCount += len(k)
		}
		fmt.Println()
	}

	fmt.Println("## Your Task")
	fmt.Println()
	fmt.Println("Extract **1-3 actionable learnings** from this session and write them to:")
	fmt.Println()
	fmt.Printf("```\n%s\n```\n", outputPath)
	fmt.Println()
	fmt.Println("### Learning Format")
	fmt.Println()
	fmt.Println("Use this markdown format for each learning:")
	fmt.Println()
	fmt.Println("```markdown")
	fmt.Println("# Learning: [Short Title]")
	fmt.Println()
	fmt.Println("**ID**: L[N]")
	fmt.Println("**Category**: [architecture|debugging|process|testing|security]")
	fmt.Println("**Confidence**: [high|medium|low]")
	fmt.Println()
	fmt.Println("## What We Learned")
	fmt.Println()
	fmt.Println("[1-2 sentences describing the insight]")
	fmt.Println()
	fmt.Println("## Why It Matters")
	fmt.Println()
	fmt.Println("[1 sentence on impact/value]")
	fmt.Println()
	fmt.Println("## Source")
	fmt.Println()
	fmt.Printf("Session: %s\n", extraction.SessionID)
	fmt.Println("```")
	fmt.Println()
	fmt.Println("### Guidelines")
	fmt.Println()
	fmt.Println("- Only extract learnings that would help **future sessions**")
	fmt.Println("- Skip trivial or context-specific details")
	fmt.Println("- Focus on: debugging insights, architectural decisions, process improvements")
	fmt.Println("- If nothing worth extracting, create the file with a note: \"No significant learnings from this session.\"")
	fmt.Println()
	fmt.Println("**After writing the file, continue with your normal work.**")
	fmt.Println("---")
}

func truncateForPrompt(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.Join(strings.Fields(s), " ") // Normalize whitespace
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
