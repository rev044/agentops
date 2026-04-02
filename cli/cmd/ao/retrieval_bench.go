package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	benchCorpus  string
	benchJSON    bool
	benchK       int
	benchLive    bool
	benchGlobal  bool
)

// benchQuery defines a single benchmark query with expected results.
type benchQuery struct {
	Query    string   `json:"query"`
	Expected []string `json:"expected"` // IDs expected in top K
	BestID   string   `json:"best_id"`  // single best expected result for MRR
}

// benchResult holds the result of a single query benchmark.
type benchResult struct {
	Query     string   `json:"query"`
	PAtK      float64  `json:"precision_at_k"`
	MRR       float64  `json:"mrr"`
	Pass      bool     `json:"pass"`
	ResultIDs []string `json:"result_ids"`
}

// benchReport holds the overall benchmark report.
type benchReport struct {
	Queries    int            `json:"queries"`
	K          int            `json:"k"`
	AvgPAtK    float64        `json:"avg_precision_at_k"`
	AvgMRR     float64        `json:"avg_mrr"`
	TargetPAtK float64        `json:"target_precision_at_k"`
	TargetMRR  float64        `json:"target_mrr"`
	Results    []benchResult  `json:"results"`
}

// liveQueryResult holds the result of a single query against the live corpus.
type liveQueryResult struct {
	Query      string   `json:"query"`
	Count      int      `json:"count"`
	TopIDs     []string `json:"top_ids"`
	MinScore   float64  `json:"min_score"`
	MaxScore   float64  `json:"max_score"`
	MeanScore  float64  `json:"mean_score"`
}

// liveReport holds results from benchmarking against the real .agents/learnings/ directory.
type liveReport struct {
	Mode           string            `json:"mode"`
	TotalLearnings int               `json:"total_learnings"`
	Queries        int               `json:"queries"`
	K              int               `json:"k"`
	Results        []liveQueryResult `json:"results"`
}

// liveQueries are broad queries that exercise real-world retrieval patterns.
var liveQueries = []string{
	"CI pipeline",
	"session intelligence",
	"hook authoring",
	"flywheel",
	"testing",
	"refactor",
	"security",
	"performance",
	"debugging",
	"architecture",
}

// runLiveBench benchmarks against the actual .agents/learnings/ directory.
// When global is true, benchmarks against ~/.agents/learnings/ (cross-rig aggregated store).
func runLiveBench(k int, asJSON, global bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	globalDir := ""
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("resolving home directory: %w", err)
		}
		globalDir = filepath.Join(home, ".agents", "learnings")
		if _, err := os.Stat(globalDir); err != nil {
			return fmt.Errorf("global knowledge store not found at %s — run 'ao harvest' first", globalDir)
		}
	}

	// Count total learnings available
	allLearnings, err := collectLearnings(cwd, "", 1000, globalDir, 1.0)
	if err != nil {
		return fmt.Errorf("collecting all learnings: %w", err)
	}

	mode := "live-local"
	if global {
		mode = "live-global"
	}
	report := liveReport{
		Mode:           mode,
		TotalLearnings: len(allLearnings),
		Queries:        len(liveQueries),
		K:              k,
	}

	for _, q := range liveQueries {
		results, err := collectLearnings(cwd, q, k*3, globalDir, 1.0)
		if err != nil {
			return fmt.Errorf("collectLearnings(%q): %w", q, err)
		}

		lr := liveQueryResult{
			Query: q,
			Count: len(results),
		}

		if len(results) > 0 {
			topN := k
			if topN > len(results) {
				topN = len(results)
			}
			lr.TopIDs = make([]string, 0, topN)
			for _, r := range results[:topN] {
				lr.TopIDs = append(lr.TopIDs, r.ID)
			}

			lr.MinScore = results[len(results)-1].CompositeScore
			lr.MaxScore = results[0].CompositeScore
			var sum float64
			for _, r := range results {
				sum += r.CompositeScore
			}
			lr.MeanScore = sum / float64(len(results))
		}

		report.Results = append(report.Results, lr)
	}

	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(report)
	}

	modeLabel := "Live (local)"
	storeLabel := ".agents/learnings/"
	if global {
		modeLabel = "Live (global)"
		storeLabel = "~/.agents/learnings/ (cross-rig)"
	}
	fmt.Printf("Retrieval Quality Report (%s)\n", modeLabel)
	fmt.Println("================================")
	fmt.Printf("Corpus:      %d learnings in %s\n", report.TotalLearnings, storeLabel)
	fmt.Printf("Queries:     %d\n", report.Queries)
	fmt.Printf("K:           %d\n", k)
	fmt.Println()

	if report.TotalLearnings == 0 {
		fmt.Println("No learnings found. Run /retro or /post-mortem to populate the knowledge base.")
		return nil
	}

	fmt.Println("Per-query breakdown:")
	for _, r := range report.Results {
		if r.Count == 0 {
			fmt.Printf("  %-25s  hits=0  (no matches)\n", fmt.Sprintf("%q", r.Query))
		} else {
			fmt.Printf("  %-25s  hits=%-3d  score=[%.2f, %.2f]  mean=%.2f  top=%v\n",
				fmt.Sprintf("%q", r.Query), r.Count, r.MinScore, r.MaxScore, r.MeanScore, r.TopIDs)
		}
	}
	return nil
}

// defaultBenchQueries returns the built-in benchmark query set.
func defaultBenchQueries() []benchQuery {
	return []benchQuery{
		{Query: "CI pipeline", Expected: []string{"ci-1.md", "ci-2.md", "ci-3.md"}, BestID: "ci-1.md"},
		{Query: "session intelligence", Expected: []string{"si-1.md", "si-2.md", "si-3.md"}, BestID: "si-1.md"},
		{Query: "hook authoring", Expected: []string{"hook-1.md", "hook-2.md", "hook-3.md"}, BestID: "hook-1.md"},
		{Query: "database", Expected: []string{"db-1.md", "db-2.md", "db-3.md"}, BestID: "db-1.md"},
		{Query: "swarm", Expected: []string{"swarm-1.md", "swarm-2.md"}, BestID: "swarm-1.md"},
	}
}

var retrievalBenchCmd = &cobra.Command{
	Use:   "retrieval-bench",
	Short: "Run retrieval quality benchmarks",
	Long:  "Measure Precision@K and MRR against a curated corpus of learning artifacts.",
	GroupID: "knowledge",
	RunE: func(cmd *cobra.Command, args []string) error {
		if benchLive {
			return runLiveBench(benchK, benchJSON, benchGlobal)
		}

		corpusDir := benchCorpus
		if corpusDir == "" {
			// Default: use embedded testdata
			exe, err := os.Executable()
			if err == nil {
				corpusDir = filepath.Join(filepath.Dir(exe), "..", "cmd", "ao", "testdata", "retrieval-bench")
			}
			// Fallback: try relative to working directory
			if _, err := os.Stat(corpusDir); err != nil {
				corpusDir = filepath.Join("cli", "cmd", "ao", "testdata", "retrieval-bench")
			}
			if _, err := os.Stat(corpusDir); err != nil {
				corpusDir = filepath.Join("cmd", "ao", "testdata", "retrieval-bench")
			}
			if _, err := os.Stat(corpusDir); err != nil {
				return fmt.Errorf("benchmark corpus not found; specify --corpus path")
			}
		}

		// Set up temp dir with corpus as .agents/learnings/
		tmpDir, err := os.MkdirTemp("", "retrieval-bench-*")
		if err != nil {
			return fmt.Errorf("creating temp dir: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		learningsDir := filepath.Join(tmpDir, ".agents", "learnings")
		if err := os.MkdirAll(learningsDir, 0o755); err != nil {
			return fmt.Errorf("creating learnings dir: %w", err)
		}

		entries, err := os.ReadDir(corpusDir)
		if err != nil {
			return fmt.Errorf("reading corpus: %w", err)
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(corpusDir, e.Name()))
			if err != nil {
				return fmt.Errorf("reading %s: %w", e.Name(), err)
			}
			if err := os.WriteFile(filepath.Join(learningsDir, e.Name()), data, 0o644); err != nil {
				return fmt.Errorf("writing %s: %w", e.Name(), err)
			}
		}

		queries := defaultBenchQueries()
		k := benchK
		targetPAtK := 0.67
		targetMRR := 0.50

		report := benchReport{
			Queries:    len(queries),
			K:          k,
			TargetPAtK: targetPAtK,
			TargetMRR:  targetMRR,
		}

		var sumPAtK, sumMRR float64
		for _, q := range queries {
			results, err := collectLearnings(tmpDir, q.Query, 10, "", 0)
			if err != nil {
				return fmt.Errorf("collectLearnings(%q): %w", q.Query, err)
			}

			expectedSet := make(map[string]bool, len(q.Expected))
			for _, id := range q.Expected {
				expectedSet[id] = true
			}

			qK := k
			if len(q.Expected) < qK {
				qK = len(q.Expected)
			}

			// Calculate P@K
			n := qK
			if n > len(results) {
				n = len(results)
			}
			hits := 0
			for _, r := range results[:n] {
				if expectedSet[r.ID] {
					hits++
				}
			}
			pAtK := float64(hits) / float64(qK)

			// Calculate MRR
			mrr := 0.0
			for i, r := range results {
				if r.ID == q.BestID {
					mrr = 1.0 / float64(i+1)
					break
				}
			}

			ids := make([]string, 0, len(results))
			for _, r := range results {
				ids = append(ids, r.ID)
			}

			report.Results = append(report.Results, benchResult{
				Query:     q.Query,
				PAtK:      pAtK,
				MRR:       mrr,
				Pass:      pAtK >= targetPAtK && mrr >= targetMRR,
				ResultIDs: ids,
			})

			sumPAtK += pAtK
			sumMRR += mrr
		}

		report.AvgPAtK = sumPAtK / float64(len(queries))
		report.AvgMRR = sumMRR / float64(len(queries))

		if benchJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(report)
		}

		// Human-readable output
		fmt.Println("Retrieval Quality Report")
		fmt.Println("========================")
		fmt.Printf("Queries:     %d\n", report.Queries)
		fmt.Printf("Precision@%d: %.2f (target: %.2f)\n", k, report.AvgPAtK, targetPAtK)
		fmt.Printf("MRR:         %.2f (target: %.2f)\n", report.AvgMRR, targetMRR)
		fmt.Println()
		fmt.Println("Per-query breakdown:")
		for _, r := range report.Results {
			status := "PASS"
			if !r.Pass {
				status = "WARN (below target)"
			}
			fmt.Printf("  %-30s P@%d=%.2f  MRR=%.2f  %s\n", fmt.Sprintf("%q", r.Query), k, r.PAtK, r.MRR, status)
		}
		return nil
	},
}

func init() {
	retrievalBenchCmd.GroupID = "knowledge"
	rootCmd.AddCommand(retrievalBenchCmd)
	retrievalBenchCmd.Flags().StringVar(&benchCorpus, "corpus", "", "Path to benchmark corpus directory")
	retrievalBenchCmd.Flags().BoolVar(&benchJSON, "json", false, "JSON output")
	retrievalBenchCmd.Flags().IntVar(&benchK, "k", 3, "K for Precision@K")
	retrievalBenchCmd.Flags().BoolVar(&benchLive, "live", false, "Benchmark against real .agents/learnings/ instead of synthetic corpus")
	retrievalBenchCmd.Flags().BoolVar(&benchGlobal, "global", false, "Include ~/.agents/learnings/ (cross-rig aggregated store, requires --live)")
}
