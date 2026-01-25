package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/boshu2/agentops/cli/internal/provenance"
	"github.com/boshu2/agentops/cli/internal/storage"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show AgentOps status",
	Long: `Display the current state of AgentOps knowledge base.

Shows:
  - Number of sessions indexed
  - Recent sessions
  - Provenance statistics
  - Storage locations

Examples:
  ao status
  ao status -o json`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

type statusOutput struct {
	Initialized     bool          `json:"initialized"`
	BaseDir         string        `json:"base_dir"`
	SessionCount    int           `json:"session_count"`
	RecentSessions  []sessionInfo `json:"recent_sessions,omitempty"`
	ProvenanceStats *provStats    `json:"provenance_stats,omitempty"`
}

type sessionInfo struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Summary string `json:"summary,omitempty"`
	Path    string `json:"path"`
}

type provStats struct {
	TotalRecords   int `json:"total_records"`
	UniqueSessions int `json:"unique_sessions"`
}

func runStatus(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	baseDir := filepath.Join(cwd, storage.DefaultBaseDir)
	status := &statusOutput{
		BaseDir: baseDir,
	}

	// Check if initialized
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		status.Initialized = false
		return outputStatus(status)
	}
	status.Initialized = true

	// Load sessions from index
	fs := storage.NewFileStorage(storage.WithBaseDir(baseDir))
	sessions, err := fs.ListSessions()
	if err == nil {
		status.SessionCount = len(sessions)

		// Get recent sessions (up to 5)
		if len(sessions) > 0 {
			// Sort by date descending
			sort.Slice(sessions, func(i, j int) bool {
				return sessions[i].Date.After(sessions[j].Date)
			})

			limit := 5
			if len(sessions) < limit {
				limit = len(sessions)
			}

			for _, s := range sessions[:limit] {
				status.RecentSessions = append(status.RecentSessions, sessionInfo{
					ID:      s.SessionID,
					Date:    s.Date.Format("2006-01-02"),
					Summary: truncateStatus(s.Summary, 60),
					Path:    filepath.Base(s.SessionPath),
				})
			}
		}
	}

	// Load provenance stats
	provPath := filepath.Join(baseDir, storage.ProvenanceDir, storage.ProvenanceFile)
	graph, err := provenance.NewGraph(provPath)
	if err == nil {
		stats := graph.GetStats()
		status.ProvenanceStats = &provStats{
			TotalRecords:   stats.TotalRecords,
			UniqueSessions: stats.UniqueSessions,
		}
	}

	return outputStatus(status)
}

func outputStatus(status *statusOutput) error {
	if GetOutput() == "json" {
		data, _ := json.MarshalIndent(status, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	// Table output
	fmt.Println("AgentOps Status")
	fmt.Println("==============")
	fmt.Println()

	if !status.Initialized {
		fmt.Println("Status: Not initialized")
		fmt.Println()
		fmt.Println("Run 'ao init' to initialize AgentOps in this directory.")
		return nil
	}

	fmt.Println("Status: Initialized ✓")
	fmt.Printf("Base Directory: %s\n", status.BaseDir)
	fmt.Println()

	fmt.Printf("Sessions: %d\n", status.SessionCount)

	if len(status.RecentSessions) > 0 {
		fmt.Println("\nRecent Sessions:")
		for _, s := range status.RecentSessions {
			fmt.Printf("  %s  %s\n", s.Date, s.Summary)
		}
	}

	if status.ProvenanceStats != nil {
		fmt.Println("\nProvenance:")
		fmt.Printf("  Records: %d\n", status.ProvenanceStats.TotalRecords)
		fmt.Printf("  Sessions: %d\n", status.ProvenanceStats.UniqueSessions)
	}

	fmt.Println("\nCommands:")
	fmt.Println("  ao forge transcript <path>  - Extract knowledge from transcript")
	fmt.Println("  ao search <query>           - Search knowledge base")
	fmt.Println("  ao trace <artifact>         - Trace provenance")

	return nil
}

func truncateStatus(s string, maxLen int) string {
	// Remove newlines
	s = firstLine(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func firstLine(s string) string {
	for i, r := range s {
		if r == '\n' {
			return s[:i]
		}
	}
	return s
}
