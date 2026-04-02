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
}
