package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	minePkg "github.com/boshu2/agentops/cli/internal/mine"
)

var (
	mineSourcesFlag   string
	mineSince         string
	mineOutputDir     string
	mineQuiet         bool
	mineEmitWorkItems bool
)

var mineCmd = &cobra.Command{
	Use:   "mine",
	Short: "Extract knowledge signal from git, .agents/, and code",
	Long: `Mine scans all reachable data sources for patterns and insights
that were never explicitly extracted into learnings or patterns.

Sources (--sources flag, comma-separated):
  git     Git log + diffs: recurring fix patterns, co-change clusters
  agents  .agents/research/ files not yet referenced in learnings
  code    gocyclo hotspots: functions edited repeatedly or high CC
  events  RPI C2 event streams: error patterns, gate verdicts (opt-in)

Output goes to .agents/mine/YYYY-MM-DD-HH.json (structured JSON).
Mine is non-destructive: it only reads and appends.

Examples:
  ao mine                           # all sources, last 26h
  ao mine --since 7d --sources git  # git only, last week
  ao mine --dry-run                 # show what would be extracted`,
	RunE: runMine,
}

func init() {
	mineCmd.GroupID = "knowledge"
	rootCmd.AddCommand(mineCmd)
	mineCmd.Flags().StringVar(&mineSourcesFlag, "sources", "git,agents,code",
		"Comma-separated sources to mine (git, agents, code, events)")
	mineCmd.Flags().StringVar(&mineSince, "since", "26h",
		"How far back to look (e.g. 26h, 7d)")
	mineCmd.Flags().StringVar(&mineOutputDir, "output-dir", ".agents/mine",
		"Directory for mine output JSON")
	mineCmd.Flags().BoolVar(&mineQuiet, "quiet", false, "Suppress progress output")
	mineCmd.Flags().BoolVar(&mineEmitWorkItems, "emit-work-items", false,
		"Append actionable mine findings to .agents/rpi/next-work.jsonl for evolve to pick up")
}

// MineReport is the top-level output of ao mine.
type MineReport struct {
	Timestamp    time.Time       `json:"timestamp"`
	SinceSeconds int64           `json:"since_seconds"`
	Sources      []string        `json:"sources"`
	Git          *GitFindings    `json:"git,omitempty"`
	Agents       *AgentsFindings `json:"agents,omitempty"`
	Code         *CodeFindings   `json:"code,omitempty"`
	Events       *EventsFindings `json:"events,omitempty"`
}

// GitFindings holds signal extracted from git log.
type GitFindings struct {
	CommitCount      int      `json:"commit_count"`
	TopCoChangeFiles []string `json:"top_co_change_files,omitempty"`
	RecurringFixes   []string `json:"recurring_fixes,omitempty"`
}

// AgentsFindings holds signal from .agents/ directory scanning.
type AgentsFindings struct {
	TotalResearch    int      `json:"total_research"`
	OrphanedResearch []string `json:"orphaned_research,omitempty"`
}

// CodeFindings holds signal from code complexity analysis.
type CodeFindings struct {
	Hotspots []ComplexityHotspot `json:"hotspots,omitempty"`
	Skipped  bool                `json:"skipped,omitempty"`
}

// ComplexityHotspot represents a high-complexity function with recent edits.
type ComplexityHotspot struct {
	File        string `json:"file"`
	Func        string `json:"func"`
	Complexity  int    `json:"complexity"`
	RecentEdits int    `json:"recent_edits"`
}

// EventsFindings holds signal extracted from RPI C2 event streams.
type EventsFindings struct {
	RunsScanned     int                  `json:"runs_scanned"`
	TotalEvents     int                  `json:"total_events"`
	EventTypeCounts map[string]int       `json:"event_type_counts,omitempty"`
	ErrorEvents     []EventErrorSummary  `json:"error_events,omitempty"`
	GateVerdicts    []GateVerdictSummary `json:"gate_verdicts,omitempty"`
}

// EventErrorSummary captures an error event from a run.
type EventErrorSummary struct {
	RunID     string `json:"run_id"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// GateVerdictSummary captures a gate verdict event from a run.
type GateVerdictSummary struct {
	RunID   string `json:"run_id"`
	Phase   int    `json:"phase"`
	Type    string `json:"type"`
	Verdict string `json:"verdict"`
}

// validMineSources enumerates the allowed source names.
var validMineSources = minePkg.ValidSources

func runMine(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	window, err := parseMineWindow(mineSince)
	if err != nil {
		return fmt.Errorf("parse --since: %w", err)
	}

	sources, err := splitSources(mineSourcesFlag)
	if err != nil {
		return err
	}

	if GetDryRun() {
		return printMineDryRun(cmd.OutOrStdout(), sources, window)
	}

	report := &MineReport{
		Timestamp:    time.Now().UTC(),
		SinceSeconds: int64(window.Seconds()),
		Sources:      sources,
	}

	runMineSources(cwd, sources, window, report, cmd.ErrOrStderr())

	return finalizeMineReport(cmd, cwd, report)
}

func runMineSources(cwd string, sources []string, window time.Duration, report *MineReport, errW io.Writer) {
	for _, src := range sources {
		switch src {
		case "git":
			findings, gitErr := mineGitLog(cwd, window)
			if gitErr != nil {
				if !mineQuiet {
					fmt.Fprintf(errW, "warning: git source: %v\n", gitErr)
				}
				continue
			}
			report.Git = findings
		case "agents":
			findings, agErr := mineAgentsDir(cwd)
			if agErr != nil {
				if !mineQuiet {
					fmt.Fprintf(errW, "warning: agents source: %v\n", agErr)
				}
				continue
			}
			report.Agents = findings
		case "code":
			findings, codeErr := mineCodeComplexity(cwd, window)
			if codeErr != nil {
				if !mineQuiet {
					fmt.Fprintf(errW, "warning: code source: %v\n", codeErr)
				}
				continue
			}
			report.Code = findings
		case "events":
			findings, evErr := mineEvents(cwd, window)
			if evErr != nil {
				if !mineQuiet {
					fmt.Fprintf(errW, "warning: events source: %v\n", evErr)
				}
				continue
			}
			report.Events = findings
		}
	}
}

func finalizeMineReport(cmd *cobra.Command, cwd string, report *MineReport) error {
	if err := writeMineReport(mineOutputDir, report); err != nil {
		return err
	}

	if mineEmitWorkItems {
		if err := emitMineWorkItems(cwd, report); err != nil && !mineQuiet {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: emit-work-items: %v\n", err)
		}
	}

	if GetOutput() == "json" {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(report)
	}

	if !mineQuiet {
		printMineSummary(cmd.OutOrStdout(), report)
	}

	return nil
}

// parseMineWindow parses a duration string with support for "h", "m", and "d" suffixes.
func parseMineWindow(s string) (time.Duration, error) { return minePkg.ParseWindow(s) }

// splitSources splits and validates a comma-separated source list.
func splitSources(s string) ([]string, error) { return minePkg.SplitSources(s) }

// mineGitLog extracts signal from git log within the given time window.
func mineGitLog(cwd string, window time.Duration) (*GitFindings, error) {
	sinceArg := fmt.Sprintf("--since=%d seconds ago", int64(window.Seconds()))
	cmd := exec.Command("git", "log", sinceArg, "--name-only", "--pretty=format:%H %s")
	cmd.Dir = cwd

	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			return nil, fmt.Errorf("git log: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("git log: %w", err)
	}

	findings := &GitFindings{}
	fileFreq := make(map[string]int)
	var fixPatterns []string
	fixRe := regexp.MustCompile(`(?i)^[0-9a-f]+ (fix|bugfix|hotfix)`)

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Commit line: starts with a hash
		if len(line) >= 41 && line[40] == ' ' {
			findings.CommitCount++
			if fixRe.MatchString(line) {
				// Extract the subject after the hash
				subject := line[41:]
				fixPatterns = append(fixPatterns, subject)
			}
		} else {
			// File line
			fileFreq[line]++
		}
	}
	if scanErr := scanner.Err(); scanErr != nil {
		return findings, fmt.Errorf("scan git log output: %w", scanErr)
	}

	// Build co-change clusters: files appearing in >=3 distinct commits
	type fileCount struct {
		file  string
		count int
	}
	var frequent []fileCount
	for f, c := range fileFreq {
		if c >= 3 {
			frequent = append(frequent, fileCount{f, c})
		}
	}
	sort.Slice(frequent, func(i, j int) bool {
		return frequent[i].count > frequent[j].count
	})

	if len(frequent) > 0 {
		topFiles := make([]string, len(frequent))
		for i, fc := range frequent {
			topFiles[i] = fc.file
		}
		findings.TopCoChangeFiles = topFiles
	}

	findings.RecurringFixes = fixPatterns

	return findings, nil
}

// mineAgentsDir scans .agents/research/ for files not referenced in learnings.
func mineAgentsDir(cwd string) (*AgentsFindings, error) {
	researchDir := filepath.Join(cwd, ".agents", "research")
	learningsDir := filepath.Join(cwd, ".agents", "learnings")

	findings := &AgentsFindings{}

	researchFiles, err := minePkg.ListMarkdownFiles(researchDir)
	if err != nil {
		if os.IsNotExist(err) {
			return findings, nil
		}
		return findings, fmt.Errorf("read research dir: %w", err)
	}
	findings.TotalResearch = len(researchFiles)
	if len(researchFiles) == 0 {
		return findings, nil
	}

	learningsContent, err := readDirContent(learningsDir)
	if err != nil && !os.IsNotExist(err) {
		return findings, fmt.Errorf("read learnings dir: %w", err)
	}

	findings.OrphanedResearch = minePkg.FindOrphanedResearch(researchFiles, learningsContent)
	return findings, nil
}

// readDirContent reads all .md file contents from a directory.
func readDirContent(dir string) (map[string]string, error) { return minePkg.ReadDirContent(dir) }

// gocycloLineRe matches gocyclo output lines like "15 main runMine cli/cmd/ao/mine.go:100:1"
var gocycloLineRe = regexp.MustCompile(`^(\d+)\s+(\S+)\s+(\S+)\s+(\S+):(\d+):\d+$`)

// mineCodeComplexity runs gocyclo and correlates with recent git edits.
func mineCodeComplexity(cwd string, window time.Duration) (*CodeFindings, error) {
	findings := &CodeFindings{}

	gocycloPath, err := exec.LookPath("gocyclo")
	if err != nil {
		findings.Skipped = true
		return findings, nil
	}

	cmd := exec.Command(gocycloPath, "-top", "10", "cli/")
	cmd.Dir = cwd

	out, err := cmd.Output()
	if err != nil {
		// gocyclo may fail if cli/ doesn't exist
		findings.Skipped = true
		return findings, nil
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		matches := gocycloLineRe.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		cc, _ := strconv.Atoi(matches[1])
		funcName := matches[3]
		file := matches[4]

		recentEdits := countRecentEdits(cwd, file, window)

		findings.Hotspots = append(findings.Hotspots, ComplexityHotspot{
			File:        file,
			Func:        funcName,
			Complexity:  cc,
			RecentEdits: recentEdits,
		})
	}
	if scanErr := scanner.Err(); scanErr != nil {
		return findings, fmt.Errorf("scan gocyclo output: %w", scanErr)
	}

	return findings, nil
}

// countRecentEdits counts how many commits touched a file within the given window.
func countRecentEdits(cwd, file string, window time.Duration) int {
	sinceArg := fmt.Sprintf("--since=%d seconds ago", int64(window.Seconds()))
	cmd := exec.Command("git", "log", sinceArg, "--oneline", "--", file)
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0
	}
	return len(lines)
}

// writeMineReport writes the mine report as JSON to the output directory.
func writeMineReport(dir string, r *MineReport) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal mine report: %w", err)
	}
	dateStr := r.Timestamp.Format("2006-01-02-15")
	return minePkg.WriteMineReportJSON(dir, data, dateStr)
}

// printMineDryRun prints what would be extracted without actually doing it.
func printMineDryRun(w io.Writer, sources []string, window time.Duration) error {
	fmt.Fprintf(w, "[dry-run] ao mine\n")
	fmt.Fprintf(w, "  sources: %s\n", strings.Join(sources, ", "))
	fmt.Fprintf(w, "  window:  %s\n", window)
	fmt.Fprintf(w, "  output:  %s\n", mineOutputDir)
	fmt.Fprintln(w, "\nNo files will be written.")
	return nil
}

// collectMineWorkItems builds work items from a mine report.
func collectMineWorkItems(r *MineReport) []mineWorkItemEmit {
	var items []mineWorkItemEmit
	if r.Code != nil {
		hotspots := make([]minePkg.ComplexityHotspot, len(r.Code.Hotspots))
		for i, h := range r.Code.Hotspots {
			hotspots[i] = minePkg.ComplexityHotspot{File: h.File, Func: h.Func, Complexity: h.Complexity, RecentEdits: h.RecentEdits}
		}
		items = append(items, minePkg.CollectWorkItemsFromHotspots(hotspots)...)
	}
	if r.Agents != nil {
		items = append(items, minePkg.CollectWorkItemsFromOrphans(r.Agents.OrphanedResearch)...)
	}
	return items
}

// loadExistingMineIDs scans a JSONL file for unconsumed compile-mine item IDs.
func loadExistingMineIDs(path string) (map[string]bool, error) {
	return minePkg.LoadExistingMineIDs(path)
}

// writeMineWorkItems appends one JSONL line per work item to the given path.
func writeMineWorkItems(path string, items []mineWorkItemEmit, ts string) error {
	return minePkg.WriteWorkItems(path, items, ts)
}

// emitMineWorkItems translates mine findings into next-work.jsonl entries for evolve.
// Orphaned research files map to severity:medium; code hotspots map to severity:high.
// Dedup: item-level — each item gets a stable ID; only new items are emitted.
func emitMineWorkItems(cwd string, r *MineReport) error {
	items := collectMineWorkItems(r)
	if len(items) == 0 {
		return nil // nothing to emit
	}

	nextWorkPath := filepath.Join(cwd, ".agents", "rpi", "next-work.jsonl")
	if err := os.MkdirAll(filepath.Dir(nextWorkPath), 0o750); err != nil {
		return fmt.Errorf("ensure next-work dir: %w", err)
	}

	existingIDs, err := loadExistingMineIDs(nextWorkPath)
	if err != nil {
		return fmt.Errorf("load existing mine IDs: %w", err)
	}
	var newItems []mineWorkItemEmit
	for _, item := range items {
		if !existingIDs[item.ID] {
			newItems = append(newItems, item)
		}
	}
	if len(newItems) == 0 {
		return nil // all items already present
	}

	return writeMineWorkItems(nextWorkPath, newItems, r.Timestamp.UTC().Format(time.RFC3339))
}

// mineWorkItemEmit is a single work item within a next-work.jsonl entry.
type mineWorkItemEmit = minePkg.WorkItemEmit

// mineWorkItemID generates a stable ID from the item's identifying fields.
func mineWorkItemID(item mineWorkItemEmit) string { return minePkg.WorkItemID(item) }

// printMineSummary prints a human-readable summary of the mine report.
func printMineSummary(w io.Writer, r *MineReport) {
	fmt.Fprintln(w, "Mine complete.")
	if r.Git != nil {
		fmt.Fprintf(w, "  git: %d commits", r.Git.CommitCount)
		if len(r.Git.TopCoChangeFiles) > 0 {
			fmt.Fprintf(w, ", %d co-change files", len(r.Git.TopCoChangeFiles))
		}
		if len(r.Git.RecurringFixes) > 0 {
			fmt.Fprintf(w, ", %d fix patterns", len(r.Git.RecurringFixes))
		}
		fmt.Fprintln(w)
	}
	if r.Agents != nil {
		fmt.Fprintf(w, "  agents: %d research files, %d orphaned\n",
			r.Agents.TotalResearch, len(r.Agents.OrphanedResearch))
	}
	if r.Code != nil {
		if r.Code.Skipped {
			fmt.Fprintln(w, "  code: skipped (gocyclo not installed)")
		} else {
			fmt.Fprintf(w, "  code: %d hotspots\n", len(r.Code.Hotspots))
		}
	}
	if r.Events != nil {
		fmt.Fprintf(w, "  events: %d runs scanned, %d total events", r.Events.RunsScanned, r.Events.TotalEvents)
		if len(r.Events.ErrorEvents) > 0 {
			fmt.Fprintf(w, ", %d errors", len(r.Events.ErrorEvents))
		}
		fmt.Fprintln(w)
	}
}

// mineEvents scans RPI C2 event streams for patterns.
func mineEvents(cwd string, window time.Duration) (*EventsFindings, error) {
	runs := scanRegistryRuns(cwd)
	if len(runs) == 0 {
		return &EventsFindings{}, nil
	}

	cutoff := time.Now().Add(-window)
	findings := &EventsFindings{
		EventTypeCounts: make(map[string]int),
	}

	for _, run := range runs {
		if run.StartedAt != "" {
			t, err := time.Parse(time.RFC3339, run.StartedAt)
			if err == nil && t.Before(cutoff) {
				continue
			}
		}

		events, err := loadRPIC2Events(cwd, run.RunID)
		if err != nil || len(events) == 0 {
			continue
		}

		findings.RunsScanned++
		for _, ev := range events {
			findings.TotalEvents++
			findings.EventTypeCounts[ev.Type]++

			if ev.Type == "error" {
				findings.ErrorEvents = append(findings.ErrorEvents, EventErrorSummary{
					RunID:     ev.RunID,
					Message:   ev.Message,
					Timestamp: ev.Timestamp,
				})
			}

			if strings.HasPrefix(ev.Type, "gate.") && strings.HasSuffix(ev.Type, ".verdict") {
				verdict := ""
				if ev.Details != nil {
					var d map[string]interface{}
					if json.Unmarshal(ev.Details, &d) == nil {
						if v, ok := d["verdict"].(string); ok {
							verdict = v
						}
					}
				}
				findings.GateVerdicts = append(findings.GateVerdicts, GateVerdictSummary{
					RunID:   ev.RunID,
					Phase:   ev.Phase,
					Type:    ev.Type,
					Verdict: verdict,
				})
			}
		}
	}

	return findings, nil
}
