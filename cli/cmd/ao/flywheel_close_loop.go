package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/boshu2/agentops/cli/internal/pool"
	"github.com/boshu2/agentops/cli/internal/ratchet"
	"github.com/boshu2/agentops/cli/internal/types"
)

var (
	flywheelCloseLoopPendingDir string
	flywheelCloseLoopThreshold  string
	flywheelCloseLoopQuiet      bool
)

type flywheelCloseLoopResult struct {
	Ingest      poolIngestResult             `json:"ingest"`
	AutoPromote poolAutoPromotePromoteResult `json:"auto_promote"`
	AntiPattern struct {
		Eligible int      `json:"eligible"`
		Promoted int      `json:"promoted"`
		Paths    []string `json:"paths,omitempty"`
	} `json:"anti_pattern"`
	Store struct {
		Categorize bool   `json:"categorize"`
		Indexed    int    `json:"indexed"`
		IndexPath  string `json:"index_path,omitempty"`
	} `json:"store"`
	CitationFeedback struct {
		Processed int `json:"processed"`
		Rewarded  int `json:"rewarded"`
		Skipped   int `json:"skipped"`
	} `json:"citation_feedback"`
}

var flywheelCloseLoopCmd = &cobra.Command{
	Use:   "close-loop",
	Short: "Close the knowledge flywheel loop",
	Long: `Close the knowledge flywheel loop by chaining:

  pool ingest → pool auto-promote (promote) → maturity transitions → store (categorize)

Designed to be safe for hooks with --quiet.

Examples:
  ao flywheel close-loop
  ao flywheel close-loop --threshold 24h --pending-dir .agents/knowledge/pending
  ao flywheel close-loop --json
  ao flywheel close-loop --dry-run`,
	RunE: runFlywheelCloseLoop,
}

func init() {
	flywheelCmd.AddCommand(flywheelCloseLoopCmd)
	flywheelCloseLoopCmd.Flags().StringVar(&flywheelCloseLoopPendingDir, "pending-dir", filepath.Join(".agents", "knowledge", "pending"), "Pending directory to ingest from")
	flywheelCloseLoopCmd.Flags().StringVar(&flywheelCloseLoopThreshold, "threshold", defaultAutoPromoteThreshold, "Minimum age for auto-promotion (default: 24h)")
	flywheelCloseLoopCmd.Flags().BoolVar(&flywheelCloseLoopQuiet, "quiet", false, "Suppress non-essential output (hook-friendly)")
}

func runFlywheelCloseLoop(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	threshold, _, err := resolveAutoPromoteThreshold(cmd, "threshold", flywheelCloseLoopThreshold)
	if err != nil {
		return err
	}

	result := flywheelCloseLoopResult{}

	// 1) pool ingest (pending markdown → pool candidates)
	ingestFiles, err := resolveIngestFiles(cwd, flywheelCloseLoopPendingDir, nil)
	if err != nil {
		return err
	}
	result.Ingest, err = ingestPendingFilesToPool(cwd, ingestFiles)
	if err != nil {
		return err
	}

	// 2) auto-promote + promote
	p := pool.NewPool(cwd)
	result.AutoPromote, err = autoPromoteAndPromoteToArtifacts(p, threshold, true)
	if err != nil {
		return err
	}

	// 3) apply ALL maturity transitions (not just anti-patterns)
	maturityResult, err := applyAllMaturityTransitions(cwd)
	if err != nil {
		return err
	}
	result.AntiPattern.Eligible = maturityResult.Total
	result.AntiPattern.Promoted = maturityResult.Applied
	result.AntiPattern.Paths = maturityResult.ChangedPaths

	// 4) store index (categorize) for newly created/changed artifacts
	pathsToIndex := append([]string{}, result.AutoPromote.Artifacts...)
	pathsToIndex = append(pathsToIndex, maturityResult.ChangedPaths...)
	result.Store.Categorize = true
	indexed, indexPath, err := storeIndexUpsert(cwd, pathsToIndex, true)
	if err != nil {
		return err
	}
	result.Store.Indexed = indexed
	result.Store.IndexPath = indexPath

	// 5) citation-to-utility feedback: process unprocessed citations
	processed, rewarded, skipped := processCitationFeedback(cwd)
	result.CitationFeedback.Processed = processed
	result.CitationFeedback.Rewarded = rewarded
	result.CitationFeedback.Skipped = skipped

	// 6) auto-promote learnings whose utility was bumped by citation feedback
	promoteCitedLearnings(cwd, flywheelCloseLoopQuiet)

	return outputFlywheelCloseLoopResult(result)
}

func outputFlywheelCloseLoopResult(res flywheelCloseLoopResult) error {
	switch GetOutput() {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(res)
	default:
		if flywheelCloseLoopQuiet {
			return nil
		}
		fmt.Println()
		fmt.Println("Flywheel Close-Loop Summary")
		fmt.Println("===========================")
		fmt.Printf("Pool ingest: added=%d (files=%d, skipped_existing=%d, skipped_malformed=%d)\n",
			res.Ingest.Added, res.Ingest.FilesScanned, res.Ingest.SkippedExisting, res.Ingest.SkippedMalformed)
		fmt.Printf("Auto-promote: promoted=%d (threshold=%s)\n", res.AutoPromote.Promoted, res.AutoPromote.Threshold)
		fmt.Printf("Anti-patterns: promoted=%d (eligible=%d)\n", res.AntiPattern.Promoted, res.AntiPattern.Eligible)
		fmt.Printf("Store: indexed=%d (categorize=%v)\n", res.Store.Indexed, res.Store.Categorize)
		fmt.Printf("Citation feedback: processed=%d (rewarded=%d, skipped=%d)\n",
			res.CitationFeedback.Processed, res.CitationFeedback.Rewarded, res.CitationFeedback.Skipped)
		fmt.Println()
		return nil
	}
}

func ingestPendingFilesToPool(cwd string, files []string) (poolIngestResult, error) {
	p := pool.NewPool(cwd)
	res := poolIngestResult{FilesScanned: len(files)}
	if len(files) == 0 {
		return res, nil
	}

	for _, f := range files {
		data, rerr := os.ReadFile(f)
		if rerr != nil {
			res.Errors++
			continue
		}
		fileDate, sessionHint := parsePendingFileHeader(string(data), f)
		blocks := parseLearningBlocks(string(data))
		res.CandidatesFound += len(blocks)
		for _, b := range blocks {
			cand, scoring, ok := buildCandidateFromLearningBlock(b, f, fileDate, sessionHint)
			if !ok {
				res.SkippedMalformed++
				continue
			}
			if _, gerr := p.Get(cand.ID); gerr == nil {
				res.SkippedExisting++
				continue
			}
			if GetDryRun() {
				res.Added++
				res.AddedIDs = append(res.AddedIDs, cand.ID)
				continue
			}
			if err := p.AddAt(cand, scoring, cand.ExtractedAt); err != nil {
				res.Errors++
				continue
			}
			res.Added++
			res.AddedIDs = append(res.AddedIDs, cand.ID)
		}
	}

	return res, nil
}

// promotionContext holds shared state for candidate promotion across flywheel and pool commands.
type promotionContext struct {
	pool            *pool.Pool
	threshold       time.Duration
	includeGold     bool
	citationCounts  map[string]int
	promotedContent map[string]bool
}

func (c *promotionContext) isEligibleTier(tier types.Tier) bool {
	return tier == types.TierSilver || (c.includeGold && tier == types.TierGold)
}

func (c *promotionContext) processCandidate(e pool.PoolEntry, result *poolAutoPromotePromoteResult) {
	if !c.isEligibleTier(e.Candidate.Tier) {
		return
	}
	if e.ScoringResult.GateRequired || e.Age < c.threshold {
		if e.ScoringResult.GateRequired {
			result.Skipped++
			result.SkippedIDs = append(result.SkippedIDs, e.Candidate.ID)
		}
		return
	}
	if reason := checkPromotionCriteria(c.pool.BaseDir, e, c.threshold, c.citationCounts, c.promotedContent); reason != "" {
		result.Skipped++
		result.SkippedIDs = append(result.SkippedIDs, e.Candidate.ID)
		VerbosePrintf("Skipping %s: %s\n", e.Candidate.ID, reason)
		return
	}
	result.Considered++
	if GetDryRun() {
		result.Promoted++
		return
	}
	stageAndPromoteEntry(c.pool, e, result, c.promotedContent)
}

func stageAndPromoteEntry(p *pool.Pool, e pool.PoolEntry, result *poolAutoPromotePromoteResult, promotedContent map[string]bool) {
	if err := p.Stage(e.Candidate.ID, types.TierSilver); err != nil {
		result.Skipped++
		result.SkippedIDs = append(result.SkippedIDs, e.Candidate.ID)
		return
	}
	artifactPath, err := p.Promote(e.Candidate.ID)
	if err != nil {
		result.Skipped++
		result.SkippedIDs = append(result.SkippedIDs, e.Candidate.ID)
		return
	}
	result.Promoted++
	result.Artifacts = append(result.Artifacts, artifactPath)
	promotedContent[normalizeContent(e.Candidate.Content)] = true
}

func autoPromoteAndPromoteToArtifacts(p *pool.Pool, threshold time.Duration, includeGold bool) (poolAutoPromotePromoteResult, error) {
	entries, err := p.List(pool.ListOptions{
		Status: types.PoolStatusPending,
	})
	if err != nil {
		return poolAutoPromotePromoteResult{}, fmt.Errorf("list pending: %w", err)
	}

	result := poolAutoPromotePromoteResult{
		Threshold: threshold.String(),
	}
	ctx := &promotionContext{
		pool:        p,
		threshold:   threshold,
		includeGold: includeGold,
	}
	ctx.citationCounts, ctx.promotedContent = loadPromotionGateContext(p.BaseDir)

	for _, e := range entries {
		ctx.processCandidate(e, &result)
	}

	return result, nil
}

// maturityTransitionSummary holds the results of applying all maturity transitions.
type maturityTransitionSummary struct {
	Total        int      `json:"total"`
	Applied      int      `json:"applied"`
	ChangedPaths []string `json:"changed_paths,omitempty"`
}

// applyAllMaturityTransitions scans learnings and applies all pending maturity transitions.
func applyAllMaturityTransitions(cwd string) (maturityTransitionSummary, error) {
	learningsDir := filepath.Join(cwd, ".agents", "learnings")
	if _, err := os.Stat(learningsDir); os.IsNotExist(err) {
		return maturityTransitionSummary{}, nil
	}

	results, err := ratchet.ScanForMaturityTransitions(learningsDir)
	if err != nil {
		return maturityTransitionSummary{}, fmt.Errorf("scan transitions: %w", err)
	}

	summary := maturityTransitionSummary{Total: len(results)}
	if len(results) == 0 || GetDryRun() {
		return summary, nil
	}

	for _, r := range results {
		learningPath, ferr := findLearningFile(filepath.Dir(learningsDir), r.LearningID)
		if ferr != nil {
			continue
		}
		applied, aerr := ratchet.ApplyMaturityTransition(learningPath)
		if aerr != nil {
			continue
		}
		if applied.Transitioned {
			summary.Applied++
			summary.ChangedPaths = append(summary.ChangedPaths, learningPath)
		}
	}

	return summary, nil
}

// loadExistingIndexEntries reads existing entries from a JSONL index file (best-effort).
func loadExistingIndexEntries(indexPath string) map[string]IndexEntry {
	existing := make(map[string]IndexEntry)
	f, err := os.Open(indexPath)
	if err != nil {
		return existing
	}
	defer func() {
		_ = f.Close() //nolint:errcheck // best-effort
	}()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	for scanner.Scan() {
		var e IndexEntry
		if err := json.Unmarshal(scanner.Bytes(), &e); err == nil && e.Path != "" {
			existing[e.Path] = e
		}
	}
	return existing
}

// upsertIndexPaths creates/updates index entries for the given paths. Returns count of indexed paths.
func upsertIndexPaths(existing map[string]IndexEntry, paths []string, categorize bool) int {
	indexed := 0
	for _, p := range paths {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err != nil {
			continue
		}
		entry, err := createIndexEntry(p, categorize)
		if err != nil {
			continue
		}
		existing[p] = *entry
		indexed++
	}
	return indexed
}

// writeIndexFile writes the entries map as sorted JSONL to the given path.
func writeIndexFile(indexPath string, existing map[string]IndexEntry) error {
	if err := os.MkdirAll(filepath.Dir(indexPath), 0750); err != nil {
		return err
	}
	out, err := os.OpenFile(indexPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close() //nolint:errcheck // write completed
	}()

	pathsSorted := make([]string, 0, len(existing))
	for p := range existing {
		pathsSorted = append(pathsSorted, p)
	}
	sort.Strings(pathsSorted)

	enc := json.NewEncoder(out)
	for _, p := range pathsSorted {
		if err := enc.Encode(existing[p]); err != nil {
			return err
		}
	}
	return nil
}

// storeIndexUpsert updates the store index for the provided paths, de-duplicating by path.
// It returns how many paths were (re)indexed and the index path.
func storeIndexUpsert(baseDir string, paths []string, categorize bool) (int, string, error) {
	indexPath := filepath.Join(baseDir, IndexDir, IndexFileName)
	if len(paths) == 0 {
		return 0, indexPath, nil
	}

	existing := loadExistingIndexEntries(indexPath)
	indexed := upsertIndexPaths(existing, paths, categorize)

	if GetDryRun() {
		return indexed, indexPath, nil
	}

	if err := writeIndexFile(indexPath, existing); err != nil {
		return indexed, indexPath, err
	}

	return indexed, indexPath, nil
}
