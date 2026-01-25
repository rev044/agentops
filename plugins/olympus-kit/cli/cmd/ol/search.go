package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/boshu2/agentops/plugins/olympus-kit/cli/internal/storage"
	"github.com/boshu2/agentops/plugins/olympus-kit/cli/pkg/vault"
)

const (
	// ContextLineMaxLength is the maximum length for context lines in search results.
	ContextLineMaxLength = 100

	// MaxContextLines is the maximum number of context lines to show per result.
	MaxContextLines = 3
)

var (
	searchLimit int
	searchType  string
	searchUseSC bool
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search knowledge base",
	Long: `Search olympus knowledge using file-based search.

By default, searches markdown and JSONL files in .agents/olympus/sessions/.
Optionally use Smart Connections for semantic search if Obsidian is running.

Examples:
  ol search "mutex pattern"
  ol search "authentication" --limit 20
  ol search "database migration" --type decisions
  ol search "config" --use-sc  # Enable Smart Connections semantic search`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().IntVar(&searchLimit, "limit", 10, "Maximum results to return")
	searchCmd.Flags().StringVar(&searchType, "type", "", "Filter by type: decisions, knowledge, sessions")
	searchCmd.Flags().BoolVar(&searchUseSC, "use-sc", false, "Enable Smart Connections semantic search (requires Obsidian)")
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	if GetDryRun() {
		fmt.Printf("[dry-run] Would search for: %s\n", query)
		return nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	baseDir := filepath.Join(cwd, storage.DefaultBaseDir)
	sessionsDir := filepath.Join(baseDir, storage.SessionsDir)

	// Check if sessions directory exists
	if _, err := os.Stat(sessionsDir); os.IsNotExist(err) {
		fmt.Println("No olympus data found.")
		fmt.Println("Run 'ol init' and 'ol forge transcript <path>' first.")
		return nil
	}

	results, err := selectAndSearch(query, sessionsDir, searchLimit)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		fmt.Printf("No results found for: %s\n", query)
		return nil
	}

	// Filter by type if specified
	if searchType != "" {
		results = filterByType(results, searchType)
	}

	// Limit results
	if len(results) > searchLimit {
		results = results[:searchLimit]
	}

	// Output results
	if GetOutput() == "json" {
		data, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	displaySearchResults(query, results)
	return nil
}

// selectAndSearch chooses the search backend and executes the search.
// Default: file-based search. Optional: Smart Connections with --use-sc flag.
func selectAndSearch(query, sessionsDir string, limit int) ([]searchResult, error) {
	// Only use Smart Connections if explicitly requested with --use-sc
	if searchUseSC {
		vaultPath := vault.DetectVault("")
		if vaultPath != "" && vault.HasSmartConnections(vaultPath) {
			VerbosePrintf("Using Smart Connections for semantic search...\n")
			results, err := searchSmartConnections(query, sessionsDir, limit)
			if err != nil {
				// Fall back to file search
				VerbosePrintf("Smart Connections failed, falling back to file search: %v\n", err)
				return searchFiles(query, sessionsDir, limit)
			}
			return results, nil
		}
		VerbosePrintf("Smart Connections not available, using file-based search...\n")
	}

	VerbosePrintf("Using file-based search...\n")
	return searchFiles(query, sessionsDir, limit)
}

// displaySearchResults formats and prints search results to stdout.
func displaySearchResults(query string, results []searchResult) {
	fmt.Printf("Found %d result(s) for: %s\n\n", len(results), query)

	for i, r := range results {
		fmt.Printf("%d. %s\n", i+1, r.Path)
		if r.Context != "" {
			lines := strings.Split(r.Context, "\n")
			for _, line := range lines {
				if line != "" {
					fmt.Printf("   %s\n", line)
				}
			}
		}
		fmt.Println()
	}
}

type searchResult struct {
	Path    string  `json:"path"`
	Score   float64 `json:"score,omitempty"`
	Context string  `json:"context,omitempty"`
	Type    string  `json:"type,omitempty"`
}

// searchFiles performs grep-based search on markdown and JSONL files.
func searchFiles(query string, dir string, limit int) ([]searchResult, error) {
	var results []searchResult

	// Search markdown files
	mdResults, err := grepFiles(query, dir, "*.md", limit)
	if err != nil {
		return nil, err
	}
	results = append(results, mdResults...)

	// Search JSONL files
	jsonlResults, err := searchJSONL(query, dir, limit)
	if err != nil {
		return nil, err
	}
	results = append(results, jsonlResults...)

	// Dedupe by path
	seen := make(map[string]bool)
	unique := make([]searchResult, 0)
	for _, r := range results {
		if !seen[r.Path] {
			seen[r.Path] = true
			unique = append(unique, r)
		}
	}

	return unique, nil
}

// grepFiles uses grep to search files.
func grepFiles(query, dir, pattern string, limit int) ([]searchResult, error) {
	cmd, useRipgrep := buildGrepCommand(query, dir, pattern)

	output, err := executeGrepWithFallback(cmd, useRipgrep, query, dir)
	if err != nil {
		return nil, err
	}
	if output == nil {
		return nil, nil
	}

	return parseGrepResults(output, pattern, query, useRipgrep), nil
}

// buildGrepCommand creates the grep/ripgrep command.
// Prefers ripgrep (rg) if available, falls back to grep.
func buildGrepCommand(query, dir, pattern string) (*exec.Cmd, bool) {
	if _, err := exec.LookPath("rg"); err == nil {
		// ripgrep with glob pattern
		return exec.Command("rg", "-l", "-i", "--max-count", "1", "--glob", pattern, query, dir), true
	}
	// grep recursive search
	return exec.Command("grep", "-l", "-i", "-r", query, dir), false
}

// executeGrepWithFallback runs the grep command with retry logic.
// If ripgrep glob pattern fails, retries without glob filter (recursive mode).
func executeGrepWithFallback(cmd *exec.Cmd, useRipgrep bool, query, dir string) ([]byte, error) {
	output, err := cmd.Output()
	if err == nil {
		return output, nil
	}

	// Both grep and rg return exit code 1 if no matches - this is normal
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
		return nil, nil
	}

	// If ripgrep failed with glob, retry without glob filter (recursive search)
	if useRipgrep {
		VerbosePrintf("ripgrep glob failed, trying recursive search: %v\n", err)
		fallbackCmd := exec.Command("rg", "-l", "-i", "--max-count", "1", query, dir)
		output, err = fallbackCmd.Output()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				return nil, nil
			}
			return nil, fmt.Errorf("search failed: %w", err)
		}
		return output, nil
	}

	return nil, fmt.Errorf("grep failed: %w", err)
}

// parseGrepResults converts grep output lines into search results.
func parseGrepResults(output []byte, pattern, query string, useRipgrep bool) []searchResult {
	var results []searchResult
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		// Filter by pattern if using grep (which doesn't filter by extension)
		if !useRipgrep && pattern != "" {
			matched, _ := filepath.Match(pattern, filepath.Base(line))
			if !matched {
				continue
			}
		}
		context := getFileContext(line, query)
		results = append(results, searchResult{
			Path:    line,
			Context: context,
			Type:    "session",
		})
	}

	return results
}

// getFileContext gets context around a match in a file.
func getFileContext(path, query string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	queryLower := strings.ToLower(query)
	var context []string

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(strings.ToLower(line), queryLower) {
			// Clean up the line
			line = strings.TrimSpace(line)
			if len(line) > ContextLineMaxLength {
				line = line[:ContextLineMaxLength] + "..."
			}
			context = append(context, line)
			if len(context) >= MaxContextLines {
				break
			}
		}
	}

	return strings.Join(context, "\n")
}

// searchJSONL searches JSONL files using jq-like parsing.
func searchJSONL(query string, dir string, limit int) ([]searchResult, error) {
	var results []searchResult

	files, err := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	if err != nil {
		return nil, err
	}

	queryLower := strings.ToLower(query)

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)

		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(strings.ToLower(line), queryLower) {
				// Parse JSON to get meaningful context
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(line), &data); err == nil {
					context := ""
					if summary, ok := data["summary"].(string); ok {
						context = summary
						if len(context) > ContextLineMaxLength {
							context = context[:ContextLineMaxLength] + "..."
						}
					}
					results = append(results, searchResult{
						Path:    file,
						Context: context,
						Type:    "session",
					})
					break // One match per file
				}
			}
		}
		f.Close()

		if len(results) >= limit {
			break
		}
	}

	return results, nil
}

// searchSmartConnections uses Smart Connections HTTP API for semantic search.
// Smart Connections exposes an HTTP API at localhost:37042 when Obsidian is running.
// Falls back to file-based search if not available.
func searchSmartConnections(query, dir string, limit int) ([]searchResult, error) {
	// Smart Connections HTTP API endpoint
	const scAPIBase = "http://localhost:37042"

	// Try to connect to Smart Connections API
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Build search request
	searchURL := fmt.Sprintf("%s/search?query=%s&limit=%d",
		scAPIBase, url.QueryEscape(query), limit)

	resp, err := client.Get(searchURL)
	if err != nil {
		// API not available - fall back to file search
		VerbosePrintf("Smart Connections API not available: %v\n", err)
		return nil, fmt.Errorf("smart connections not running: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("smart connections API error: %s", resp.Status)
	}

	// Parse response
	var scResponse struct {
		Results []struct {
			Path    string  `json:"path"`
			Score   float64 `json:"score"`
			Content string  `json:"content,omitempty"`
			Title   string  `json:"title,omitempty"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&scResponse); err != nil {
		return nil, fmt.Errorf("parse Smart Connections response: %w", err)
	}

	// Convert to searchResult format
	var results []searchResult
	for _, r := range scResponse.Results {
		context := r.Content
		if context == "" && r.Title != "" {
			context = r.Title
		}
		if len(context) > ContextLineMaxLength {
			context = context[:ContextLineMaxLength] + "..."
		}

		results = append(results, searchResult{
			Path:    r.Path,
			Score:   r.Score,
			Context: context,
			Type:    classifyResultType(r.Path),
		})
	}

	return results, nil
}

// classifyResultType determines the knowledge type based on file path.
func classifyResultType(path string) string {
	pathLower := strings.ToLower(path)

	if strings.Contains(pathLower, "/learnings/") {
		return "learning"
	}
	if strings.Contains(pathLower, "/patterns/") {
		return "pattern"
	}
	if strings.Contains(pathLower, "/retros/") {
		return "retro"
	}
	if strings.Contains(pathLower, "/research/") {
		return "research"
	}
	if strings.Contains(pathLower, "/sessions/") {
		return "session"
	}
	if strings.Contains(pathLower, "/decisions/") {
		return "decision"
	}

	return "knowledge"
}

// filterByType filters results by knowledge type.
func filterByType(results []searchResult, filterType string) []searchResult {
	var filtered []searchResult
	for _, r := range results {
		if r.Type == filterType || filterType == "" {
			filtered = append(filtered, r)
		}
	}
	return filtered
}
