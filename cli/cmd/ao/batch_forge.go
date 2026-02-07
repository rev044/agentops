package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/boshu2/agentops/cli/internal/formatter"
	"github.com/boshu2/agentops/cli/internal/parser"
	"github.com/boshu2/agentops/cli/internal/storage"
)

var forgeBatchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Process multiple transcripts at once",
	Long: `Find and process pending transcripts in bulk.

Scans standard Claude Code transcript locations, processes each through
the forge extraction pipeline, and deduplicates similar learnings.

Examples:
  ao forge batch                    # Process all pending transcripts
  ao forge batch --dry-run          # List what would be processed
  ao forge batch --dir ~/.claude/projects/my-project`,
	RunE: runForgeBatch,
}

var batchDir string

func init() {
	forgeCmd.AddCommand(forgeBatchCmd)
	forgeBatchCmd.Flags().StringVar(&batchDir, "dir", "", "Specific directory to scan (default: all Claude project dirs)")
}

func runForgeBatch(cmd *cobra.Command, args []string) error {
	transcripts, err := findPendingTranscripts(batchDir)
	if err != nil {
		return fmt.Errorf("find transcripts: %w", err)
	}

	if len(transcripts) == 0 {
		fmt.Println("No pending transcripts found.")
		return nil
	}

	if GetDryRun() {
		fmt.Printf("[dry-run] Would process %d transcript(s):\n", len(transcripts))
		for _, t := range transcripts {
			fmt.Printf("  - %s (%s)\n", t.path, humanSize(t.size))
		}
		return nil
	}

	fmt.Printf("Found %d transcript(s) to process.\n", len(transcripts))

	// Initialize storage
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	baseDir := filepath.Join(cwd, storage.DefaultBaseDir)
	fs := storage.NewFileStorage(
		storage.WithBaseDir(baseDir),
		storage.WithFormatters(
			formatter.NewMarkdownFormatter(),
			formatter.NewJSONLFormatter(),
		),
	)

	if err := fs.Init(); err != nil {
		return fmt.Errorf("initialize storage: %w", err)
	}

	p := parser.NewParser()
	p.MaxContentLength = 0
	extractor := parser.NewExtractor()

	var (
		totalProcessed  int
		totalDecisions  int
		totalKnowledge  int
		totalDupsRemoved int
		allKnowledge    []string
		allDecisions    []string
	)

	for i, t := range transcripts {
		fmt.Printf("[%d/%d] Processing %s...\n", i+1, len(transcripts), filepath.Base(t.path))

		session, err := processTranscript(t.path, p, extractor, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: skipping %s: %v\n", t.path, err)
			continue
		}

		// Write session
		sessionPath, err := fs.WriteSession(session)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: failed to write session for %s: %v\n", t.path, err)
			continue
		}

		// Write index entry
		indexEntry := &storage.IndexEntry{
			SessionID:   session.ID,
			Date:        session.Date,
			SessionPath: sessionPath,
			Summary:     session.Summary,
		}
		if err := fs.WriteIndex(indexEntry); err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: failed to index session: %v\n", err)
		}

		// Write provenance
		provRecord := &storage.ProvenanceRecord{
			ID:           fmt.Sprintf("prov-%s", session.ID[:7]),
			ArtifactPath: sessionPath,
			ArtifactType: "session",
			SourcePath:   t.path,
			SourceType:   "transcript",
			SessionID:    session.ID,
			CreatedAt:    time.Now(),
		}
		if err := fs.WriteProvenance(provRecord); err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: failed to write provenance: %v\n", err)
		}

		totalProcessed++
		totalDecisions += len(session.Decisions)
		totalKnowledge += len(session.Knowledge)
		allKnowledge = append(allKnowledge, session.Knowledge...)
		allDecisions = append(allDecisions, session.Decisions...)

		VerbosePrintf("  -> %d decisions, %d learnings\n", len(session.Decisions), len(session.Knowledge))
	}

	// Deduplicate across all sessions
	dedupedKnowledge := dedupSimilar(allKnowledge)
	dedupedDecisions := dedupSimilar(allDecisions)
	knowledgeDups := len(allKnowledge) - len(dedupedKnowledge)
	decisionDups := len(allDecisions) - len(dedupedDecisions)
	totalDupsRemoved = knowledgeDups + decisionDups

	fmt.Printf("\n--- Batch Forge Summary ---\n")
	fmt.Printf("Transcripts processed: %d\n", totalProcessed)
	fmt.Printf("Decisions extracted:   %d\n", totalDecisions)
	fmt.Printf("Learnings extracted:   %d\n", totalKnowledge)
	fmt.Printf("Duplicates removed:    %d\n", totalDupsRemoved)
	fmt.Printf("Unique decisions:      %d\n", len(dedupedDecisions))
	fmt.Printf("Unique learnings:      %d\n", len(dedupedKnowledge))
	fmt.Printf("Output:                %s\n", baseDir)

	return nil
}

// transcriptCandidate represents a discovered transcript file.
type transcriptCandidate struct {
	path    string
	modTime time.Time
	size    int64
}

// findPendingTranscripts discovers JSONL transcript files in Claude project directories.
func findPendingTranscripts(specificDir string) ([]transcriptCandidate, error) {
	var searchDirs []string

	if specificDir != "" {
		searchDirs = []string{specificDir}
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home directory: %w", err)
		}
		projectsDir := filepath.Join(homeDir, ".claude", "projects")
		if _, err := os.Stat(projectsDir); os.IsNotExist(err) {
			return nil, nil // No projects dir, nothing to process
		}
		searchDirs = []string{projectsDir}
	}

	var candidates []transcriptCandidate

	for _, dir := range searchDirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			// Skip subagent directories
			if info.IsDir() && info.Name() == "subagents" {
				return filepath.SkipDir
			}
			if !info.IsDir() && filepath.Ext(path) == ".jsonl" && info.Size() > 100 {
				candidates = append(candidates, transcriptCandidate{
					path:    path,
					modTime: info.ModTime(),
					size:    info.Size(),
				})
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walk %s: %w", dir, err)
		}
	}

	// Sort by modification time, oldest first (process in chronological order)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].modTime.Before(candidates[j].modTime)
	})

	return candidates, nil
}

// dedupSimilar removes exact duplicates and near-duplicates from a string slice.
// Near-duplicates are detected by comparing normalized prefixes.
func dedupSimilar(items []string) []string {
	if len(items) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	result := make([]string, 0, len(items))

	for _, item := range items {
		key := normalizeForDedup(item)
		if !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}

	return result
}

// normalizeForDedup creates a normalized key for deduplication.
// Lowercases, trims whitespace, and truncates to first 80 chars
// to catch near-duplicate snippets that differ only in trailing content.
func normalizeForDedup(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	// Remove trailing ellipsis from snippets
	s = strings.TrimSuffix(s, "...")
	s = strings.TrimSpace(s)
	if len(s) > 80 {
		s = s[:80]
	}
	return s
}

// humanSize returns a human-readable file size string.
func humanSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMG"[exp])
}
