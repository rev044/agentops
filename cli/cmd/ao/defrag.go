package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	defragPrune            bool
	defragDedup            bool
	defragOscillationSweep bool
	defragStaleDays        int
	defragOutputDir        string
	defragQuiet            bool
)

var defragCmd = &cobra.Command{
	Use:   "defrag",
	Short: "Prune, deduplicate, and sweep oscillating goals from .agents/",
	Long: `Defrag performs mechanical cleanup of the knowledge base:

  --prune             Find orphaned learnings (no references, >N days old)
  --dedup             Flag learnings with >80% content similarity
  --oscillation-sweep Read evolve cycle history; flag goals alternating >=3 cycles

By default, defrag applies prune/dedup changes unless you pass the global
--dry-run flag. Use ao --dry-run defrag ... to inspect the report without
deleting orphaned or duplicate learnings.

Output: .agents/defrag/YYYY-MM-DD.json with full delta report.

Examples:
  ao --dry-run defrag --prune --dedup --oscillation-sweep   # full report only
  ao defrag --prune --stale-days 14                         # apply prune/delete rules`,
	RunE: runDefrag,
}

func init() {
	defragCmd.GroupID = "knowledge"
	rootCmd.AddCommand(defragCmd)
	defragCmd.Flags().BoolVar(&defragPrune, "prune", false,
		"Find orphaned learnings not referenced in patterns or research")
	defragCmd.Flags().BoolVar(&defragDedup, "dedup", false,
		"Flag learnings with >80% content similarity")
	defragCmd.Flags().BoolVar(&defragOscillationSweep, "oscillation-sweep", false,
		"Flag evolve goals alternating improved/fail >=3 consecutive cycles")
	defragCmd.Flags().IntVar(&defragStaleDays, "stale-days", 30,
		"Days after which an unreferenced learning is considered stale")
	defragCmd.Flags().StringVar(&defragOutputDir, "output-dir", ".agents/defrag",
		"Directory for defrag report JSON")
	defragCmd.Flags().BoolVar(&defragQuiet, "quiet", false, "Suppress progress output")
}

// DefragReport is the top-level output of a defrag run.
type DefragReport struct {
	Timestamp   time.Time          `json:"timestamp"`
	DryRun      bool               `json:"dry_run"`
	Prune       *PruneResult       `json:"prune,omitempty"`
	Dedup       *DefragDedupResult `json:"dedup,omitempty"`
	Oscillation *OscillationResult `json:"oscillation,omitempty"`
}

// PruneResult holds orphan-detection results.
type PruneResult struct {
	TotalLearnings int      `json:"total_learnings"`
	StaleCount     int      `json:"stale_count"`
	Orphans        []string `json:"orphans,omitempty"`
	Deleted        []string `json:"deleted,omitempty"`
}

// DefragDedupResult holds near-duplicate detection results for defrag.
type DefragDedupResult struct {
	Checked        int         `json:"checked"`
	DuplicatePairs [][2]string `json:"duplicate_pairs,omitempty"`
	Deleted        []string    `json:"deleted,omitempty"`
}

// OscillationResult holds oscillating-goal sweep results.
type OscillationResult struct {
	OscillatingGoals []OscillatingGoal `json:"oscillating_goals,omitempty"`
}

// OscillatingGoal describes a goal that alternates improved/fail.
type OscillatingGoal struct {
	Target           string `json:"target"`
	AlternationCount int    `json:"alternation_count"`
	LastCycle        int    `json:"last_cycle"`
}

func runDefrag(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	isDryRun := GetDryRun()
	defragDefaultModes()

	report := &DefragReport{
		Timestamp: time.Now().UTC(),
		DryRun:    isDryRun,
	}

	if err := runDefragPhases(cwd, isDryRun, report); err != nil {
		return err
	}

	return writeDefragReport(defragOutputDir, report, cmd.OutOrStdout())
}

// defragDefaultModes enables all mode flags when none are explicitly set.
func defragDefaultModes() {
	if !defragPrune && !defragDedup && !defragOscillationSweep {
		defragPrune = true
		defragDedup = true
		defragOscillationSweep = true
	}
}

// runDefragPhases executes the selected defrag operations and populates the report.
func runDefragPhases(cwd string, isDryRun bool, report *DefragReport) error {
	if defragPrune {
		result, err := executePrune(cwd, isDryRun, defragStaleDays)
		if err != nil {
			return err
		}
		report.Prune = result
	}

	if defragDedup {
		result, err := executeDedup(cwd, isDryRun)
		if err != nil {
			return err
		}
		report.Dedup = result
	}

	if defragOscillationSweep {
		result, err := sweepOscillatingGoals(cwd)
		if err != nil {
			return fmt.Errorf("oscillation sweep: %w", err)
		}
		report.Oscillation = result
	}

	return nil
}

// executePrune finds orphan learnings and optionally deletes them.
func executePrune(cwd string, isDryRun bool, staleDays int) (*PruneResult, error) {
	result, err := findOrphanLearnings(cwd, staleDays)
	if err != nil {
		return nil, fmt.Errorf("prune: %w", err)
	}
	if !isDryRun && len(result.Orphans) > 0 {
		for _, orphan := range result.Orphans {
			p := filepath.Join(cwd, orphan)
			if err := os.Remove(p); err != nil {
				return nil, fmt.Errorf("delete orphan %s: %w", orphan, err)
			}
			result.Deleted = append(result.Deleted, orphan)
		}
	}
	return result, nil
}

// executeDedup finds duplicate learnings and optionally removes them.
func executeDedup(cwd string, isDryRun bool) (*DefragDedupResult, error) {
	result, err := findDuplicateLearnings(cwd)
	if err != nil {
		return nil, fmt.Errorf("dedup: %w", err)
	}
	if !isDryRun {
		for _, pair := range result.DuplicatePairs {
			// Keep pair[0], delete pair[1]. Prefer named files over hash-named ones.
			keep, del := pair[0], pair[1]
			if isHashNamed(pair[0]) && !isHashNamed(pair[1]) {
				keep, del = pair[1], pair[0]
			}
			_ = keep
			p := filepath.Join(cwd, ".agents", "learnings", del)
			if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
				return nil, fmt.Errorf("dedup remove %s: %w", del, err)
			}
			result.Deleted = append(result.Deleted, del)
		}
		result.DuplicatePairs = nil // pairs resolved
	}
	return result, nil
}

// findOrphanLearnings scans .agents/learnings/ for files older than staleDays
// that are not referenced in any .agents/patterns/ or .agents/research/ file.
func findOrphanLearnings(cwd string, staleDays int) (*PruneResult, error) {
	learningsDir := filepath.Join(cwd, ".agents", "learnings")
	entries, err := os.ReadDir(learningsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return &PruneResult{}, nil
		}
		return nil, fmt.Errorf("read learnings dir: %w", err)
	}

	cutoff := time.Now().AddDate(0, 0, -staleDays)

	// Collect reference content from patterns and research dirs
	refContent, err := collectReferenceContent(cwd)
	if err != nil {
		return nil, err
	}

	result := &PruneResult{}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		result.TotalLearnings++

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().After(cutoff) {
			continue
		}
		result.StaleCount++

		// Check if filename appears in any reference file
		if !strings.Contains(refContent, entry.Name()) {
			relPath := filepath.Join(".agents", "learnings", entry.Name())
			result.Orphans = append(result.Orphans, relPath)
		}
	}

	sort.Strings(result.Orphans)
	return result, nil
}

// collectReferenceContent reads all .md files from .agents/patterns/ and
// .agents/research/ and returns their concatenated content for link checking.
func collectReferenceContent(cwd string) (string, error) {
	var buf strings.Builder
	for _, sub := range []string{SectionPatterns, SectionResearch} {
		dir := filepath.Join(cwd, ".agents", sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", fmt.Errorf("read %s dir: %w", sub, err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				continue
			}
			buf.Write(data)
			buf.WriteByte('\n')
		}
	}
	return buf.String(), nil
}

// isHashNamed returns true if the filename looks like an auto-generated hash name
// (8 hex chars preceded by a date prefix, e.g. "2026-02-23-4556c2b4.md").
func isHashNamed(name string) bool {
	// Strip path components — operate on basename only.
	base := filepath.Base(name)
	// Remove .md extension.
	stem := strings.TrimSuffix(base, ".md")
	// Pattern: YYYY-MM-DD-<8hexchars>
	parts := strings.Split(stem, "-")
	if len(parts) < 4 {
		return false
	}
	last := parts[len(parts)-1]
	if len(last) != 8 {
		return false
	}
	for _, c := range last {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

// findDuplicateLearnings reads all .agents/learnings/*.md files and flags
// pairs with >80% trigram overlap as near-duplicates.
func findDuplicateLearnings(cwd string) (*DefragDedupResult, error) {
	learningsDir := filepath.Join(cwd, ".agents", "learnings")
	entries, err := os.ReadDir(learningsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return &DefragDedupResult{}, nil
		}
		return nil, fmt.Errorf("read learnings dir: %w", err)
	}

	type learningFile struct {
		name     string
		trigrams map[string]bool
	}

	var files []learningFile

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(learningsDir, entry.Name()))
		if err != nil {
			continue
		}
		text := strings.ToLower(string(data))
		tg := buildTrigrams(text)
		files = append(files, learningFile{name: entry.Name(), trigrams: tg})
	}

	result := &DefragDedupResult{Checked: len(files)}

	// O(n^2) pairwise comparison — fine for ~65 files
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			overlap := trigramOverlap(files[i].trigrams, files[j].trigrams)
			if overlap > 0.80 {
				result.DuplicatePairs = append(result.DuplicatePairs, [2]string{
					files[i].name, files[j].name,
				})
			}
		}
	}

	return result, nil
}

// buildTrigrams returns the set of character trigrams from text.
func buildTrigrams(text string) map[string]bool {
	tg := make(map[string]bool)
	runes := []rune(text)
	for i := 0; i+2 < len(runes); i++ {
		tg[string(runes[i:i+3])] = true
	}
	return tg
}

// trigramOverlap returns the Jaccard similarity of two trigram sets.
func trigramOverlap(a, b map[string]bool) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 0
	}

	intersect := 0
	for k := range a {
		if b[k] {
			intersect++
		}
	}

	union := len(a) + len(b) - intersect
	if union == 0 {
		return 0
	}
	return float64(intersect) / float64(union)
}

// cycleRecord represents one line in cycle-history.jsonl.
type cycleRecord struct {
	Cycle  int    `json:"cycle"`
	Target string `json:"target"`
	Result string `json:"result"`
}

// sweepOscillatingGoals parses .agents/evolve/cycle-history.jsonl and finds
// goals whose result alternates between "improved" and non-"improved" >=3 times.
func sweepOscillatingGoals(cwd string) (*OscillationResult, error) {
	histPath := filepath.Join(cwd, ".agents", "evolve", "cycle-history.jsonl")
	f, err := os.Open(histPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &OscillationResult{}, nil
		}
		return nil, fmt.Errorf("open cycle history: %w", err)
	}
	defer f.Close()

	// Group records by target, preserving order
	targetRecords := make(map[string][]cycleRecord)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var rec cycleRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			continue // skip malformed lines
		}
		if rec.Target == "" {
			continue
		}
		targetRecords[rec.Target] = append(targetRecords[rec.Target], rec)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan cycle history: %w", err)
	}

	result := &OscillationResult{}

	// Sort targets for deterministic output
	targets := make([]string, 0, len(targetRecords))
	for t := range targetRecords {
		targets = append(targets, t)
	}
	sort.Strings(targets)

	for _, target := range targets {
		records := targetRecords[target]
		alternations := countAlternations(records)
		if alternations >= 3 {
			lastCycle := records[len(records)-1].Cycle
			result.OscillatingGoals = append(result.OscillatingGoals, OscillatingGoal{
				Target:           target,
				AlternationCount: alternations,
				LastCycle:        lastCycle,
			})
		}
	}

	return result, nil
}

// countAlternations counts how many times the result alternates between
// "improved" and non-"improved" in a sequence of records.
func countAlternations(records []cycleRecord) int {
	if len(records) < 2 {
		return 0
	}
	count := 0
	for i := 1; i < len(records); i++ {
		prevImproved := records[i-1].Result == "improved"
		currImproved := records[i].Result == "improved"
		if prevImproved != currImproved {
			count++
		}
	}
	return count
}

// writeDefragReport writes the report as dated JSON and latest.json.
// The w parameter receives human-readable or JSON output that would previously
// have been written directly to os.Stdout.
func writeDefragReport(dir string, r *DefragReport, w io.Writer) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	data = append(data, '\n')

	dateStr := r.Timestamp.Format("2006-01-02")
	datedPath := filepath.Join(dir, dateStr+".json")
	latestPath := filepath.Join(dir, "latest.json")

	// Write the SHA-256 to prevent identical rewrites
	hash := fmt.Sprintf("%x", sha256.Sum256(data))
	_ = hash // reserved for future dedup

	if err := os.WriteFile(datedPath, data, 0o644); err != nil {
		return fmt.Errorf("write dated report: %w", err)
	}
	if err := os.WriteFile(latestPath, data, 0o644); err != nil {
		return fmt.Errorf("write latest report: %w", err)
	}

	// Handle output format
	if GetOutput() == "json" {
		return json.NewEncoder(w).Encode(r)
	}
	if !defragQuiet {
		printDefragSummary(w, r)
	}

	return nil
}

// printDefragSummary prints a human-readable summary to w.
func printDefragSummary(w io.Writer, r *DefragReport) {
	fmt.Fprintf(w, "Defrag report: %s\n", r.Timestamp.Format(time.RFC3339))
	if r.DryRun {
		fmt.Fprintln(w, "  Mode: dry-run (no changes applied)")
	}
	if r.Prune != nil {
		fmt.Fprintf(w, "  Prune: %d total learnings, %d stale, %d orphans\n",
			r.Prune.TotalLearnings, r.Prune.StaleCount, len(r.Prune.Orphans))
		if len(r.Prune.Deleted) > 0 {
			fmt.Fprintf(w, "  Deleted: %d files\n", len(r.Prune.Deleted))
		}
	}
	if r.Dedup != nil {
		fmt.Fprintf(w, "  Dedup: %d checked, %d duplicate pairs\n",
			r.Dedup.Checked, len(r.Dedup.DuplicatePairs))
	}
	if r.Oscillation != nil {
		fmt.Fprintf(w, "  Oscillation: %d oscillating goals\n",
			len(r.Oscillation.OscillatingGoals))
	}
}
