package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/boshu2/agentops/cli/internal/ratchet"
	"github.com/boshu2/agentops/cli/internal/storage"
	"github.com/boshu2/agentops/cli/internal/types"
)

var (
	metricsDays int
)

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Knowledge flywheel metrics",
	Long: `Track and report on knowledge flywheel metrics.

The flywheel equation:
  dK/dt = I(t) - δ·K + σ·ρ·K - B(K, K_crit)

Where:
  K     = Total knowledge artifacts
  I(t)  = New knowledge inflow
  δ     = Decay rate (default: 0.17/week)
  σ     = Retrieval effectiveness (0-1)
  ρ     = Citation rate per artifact
  B()   = Breakdown function at capacity

Escape velocity: σ × ρ > δ → Knowledge compounds

Commands:
  baseline   Capture current flywheel state
  report     Show metrics with escape velocity status`,
}

func init() {
	rootCmd.AddCommand(metricsCmd)

	// baseline subcommand
	baselineCmd := &cobra.Command{
		Use:   "baseline",
		Short: "Capture current flywheel state",
		Long: `Capture a baseline snapshot of the knowledge flywheel.

Records:
  - Total artifact counts by tier
  - Citation counts and patterns
  - Current σ, ρ estimates
  - Escape velocity status

Output is saved to .agents/ao/metrics/baseline-YYYY-MM-DD.json

Examples:
  ao metrics baseline
  ao metrics baseline --days 7
  ao metrics baseline -o json`,
		RunE: runMetricsBaseline,
	}
	baselineCmd.Flags().IntVar(&metricsDays, "days", 7, "Period in days for metrics calculation")
	metricsCmd.AddCommand(baselineCmd)

	// report subcommand
	reportCmd := &cobra.Command{
		Use:   "report",
		Short: "Show flywheel metrics report",
		Long: `Display a formatted report of knowledge flywheel metrics.

Shows:
  - Core parameters (δ, σ, ρ)
  - Derived values (σ×ρ, velocity)
  - Escape velocity status
  - Artifact counts by tier
  - Trend indicators

Examples:
  ao metrics report
  ao metrics report --days 30
  ao metrics report -o json`,
		RunE: runMetricsReport,
	}
	reportCmd.Flags().IntVar(&metricsDays, "days", 7, "Period in days for metrics calculation")
	metricsCmd.AddCommand(reportCmd)

	// cite subcommand - record a citation event
	citeCmd := &cobra.Command{
		Use:   "cite <artifact-path>",
		Short: "Record a citation event",
		Long: `Record that an artifact was cited in this session.

Citation events drive the knowledge flywheel:
  - Increases ρ (citation rate)
  - Contributes to σ×ρ calculation
  - Can trigger auto-promotion after threshold

Examples:
  ao metrics cite .agents/learnings/mutex-pattern.md
  ao metrics cite .agents/patterns/error-handling.md --type applied
  ao metrics cite .agents/research/oauth.md --session abc123`,
		Args: cobra.ExactArgs(1),
		RunE: runMetricsCite,
	}
	var citeType, citeSession, citeQuery string
	citeCmd.Flags().StringVar(&citeType, "type", "reference", "Citation type: recall, reference, applied")
	citeCmd.Flags().StringVar(&citeSession, "session", "", "Session ID (auto-detected if not provided)")
	citeCmd.Flags().StringVar(&citeQuery, "query", "", "Search query that surfaced this artifact")
	metricsCmd.AddCommand(citeCmd)
}

// runMetricsBaseline captures a baseline snapshot.
func runMetricsBaseline(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	if GetDryRun() {
		fmt.Printf("[dry-run] Would capture baseline for %d day period\n", metricsDays)
		return nil
	}

	metrics, err := computeMetrics(cwd, metricsDays)
	if err != nil {
		return fmt.Errorf("compute metrics: %w", err)
	}

	// Save baseline
	baselinePath, err := saveBaseline(cwd, metrics)
	if err != nil {
		return fmt.Errorf("save baseline: %w", err)
	}

	// Output
	switch GetOutput() {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(metrics)

	case "yaml":
		enc := yaml.NewEncoder(os.Stdout)
		return enc.Encode(metrics)

	default:
		printMetricsTable(metrics)
		fmt.Printf("\nBaseline saved: %s\n", baselinePath)
	}

	return nil
}

// runMetricsReport shows the metrics report.
func runMetricsReport(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	metrics, err := computeMetrics(cwd, metricsDays)
	if err != nil {
		return fmt.Errorf("compute metrics: %w", err)
	}

	switch GetOutput() {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(metrics)

	case "yaml":
		enc := yaml.NewEncoder(os.Stdout)
		return enc.Encode(metrics)

	default:
		printMetricsTable(metrics)
	}

	return nil
}

// runMetricsCite records a citation event.
func runMetricsCite(cmd *cobra.Command, args []string) error {
	artifactPath := args[0]

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	// Make path absolute if needed
	if !filepath.IsAbs(artifactPath) {
		artifactPath = filepath.Join(cwd, artifactPath)
	}

	// Verify artifact exists
	if _, err := os.Stat(artifactPath); os.IsNotExist(err) {
		return fmt.Errorf("artifact not found: %s", artifactPath)
	}

	// Get flags
	citeType, _ := cmd.Flags().GetString("type")
	citeSession, _ := cmd.Flags().GetString("session")
	citeQuery, _ := cmd.Flags().GetString("query")

	// Auto-detect session ID if not provided
	if citeSession == "" {
		citeSession = detectSessionID()
	}

	event := types.CitationEvent{
		ArtifactPath: artifactPath,
		SessionID:    citeSession,
		CitedAt:      time.Now(),
		CitationType: citeType,
		Query:        citeQuery,
	}

	if GetDryRun() {
		fmt.Printf("[dry-run] Would record citation:\n")
		fmt.Printf("  Artifact: %s\n", artifactPath)
		fmt.Printf("  Session: %s\n", citeSession)
		fmt.Printf("  Type: %s\n", citeType)
		return nil
	}

	if err := ratchet.RecordCitation(cwd, event); err != nil {
		return fmt.Errorf("record citation: %w", err)
	}

	fmt.Printf("Citation recorded: %s\n", filepath.Base(artifactPath))
	return nil
}

// periodCitationStats holds citation statistics for a period
type periodCitationStats struct {
	citations   []types.CitationEvent
	uniqueCited map[string]bool
}

// filterCitationsForPeriod filters citations to a time period
func filterCitationsForPeriod(citations []types.CitationEvent, start, end time.Time) periodCitationStats {
	stats := periodCitationStats{
		uniqueCited: make(map[string]bool),
	}
	for _, c := range citations {
		if c.CitedAt.After(start) && c.CitedAt.Before(end) {
			stats.citations = append(stats.citations, c)
			stats.uniqueCited[c.ArtifactPath] = true
		}
	}
	return stats
}

// computeSigmaRho calculates retrieval effectiveness (σ) and citation rate (ρ)
func computeSigmaRho(totalArtifacts, uniqueCited, citationCount, days int) (sigma, rho float64) {
	if totalArtifacts > 0 {
		sigma = float64(uniqueCited) / float64(totalArtifacts)
	}
	weeks := float64(days) / 7.0
	if weeks > 0 && uniqueCited > 0 {
		rho = float64(citationCount) / float64(uniqueCited) / weeks
	}
	return sigma, rho
}

// countLoopMetrics counts learnings created vs found for loop closure
func countLoopMetrics(baseDir string, periodStart time.Time, periodCitations []types.CitationEvent) (created, found int) {
	created, _ = countNewArtifactsInDir(filepath.Join(baseDir, ".agents", "learnings"), periodStart)
	for _, c := range periodCitations {
		if strings.Contains(c.ArtifactPath, "/learnings/") {
			found++
		}
	}
	return created, found
}

// countBypassCitations counts prior art bypass citations
func countBypassCitations(citations []types.CitationEvent) int {
	count := 0
	for _, c := range citations {
		if c.CitationType == "bypass" || strings.HasPrefix(c.ArtifactPath, "bypass:") {
			count++
		}
	}
	return count
}

// computeMetrics calculates flywheel metrics for a period.
func computeMetrics(baseDir string, days int) (*types.FlywheelMetrics, error) {
	now := time.Now()
	periodStart := now.AddDate(0, 0, -days)

	metrics := &types.FlywheelMetrics{
		Timestamp:   now,
		PeriodStart: periodStart,
		PeriodEnd:   now,
		Delta:       types.DefaultDelta,
		TierCounts:  make(map[string]int),
	}

	// Count artifacts
	totalArtifacts, tierCounts, err := countArtifacts(baseDir)
	if err != nil {
		VerbosePrintf("Warning: count artifacts: %v\n", err)
	}
	metrics.TotalArtifacts = totalArtifacts
	metrics.TierCounts = tierCounts

	// Load and filter citations
	citations, err := ratchet.LoadCitations(baseDir)
	if err != nil {
		VerbosePrintf("Warning: load citations: %v\n", err)
	}
	stats := filterCitationsForPeriod(citations, periodStart, now)
	metrics.CitationsThisPeriod = len(stats.citations)
	metrics.UniqueCitedArtifacts = len(stats.uniqueCited)

	// Calculate σ and ρ
	metrics.Sigma, metrics.Rho = computeSigmaRho(
		metrics.TotalArtifacts, metrics.UniqueCitedArtifacts, metrics.CitationsThisPeriod, days,
	)
	metrics.SigmaRho = metrics.Sigma * metrics.Rho
	metrics.Velocity = metrics.SigmaRho - metrics.Delta
	metrics.AboveEscapeVelocity = metrics.SigmaRho > metrics.Delta

	// Count new and stale artifacts
	if newCount, err := countNewArtifacts(baseDir, periodStart); err == nil {
		metrics.NewArtifacts = newCount
	}
	if staleCount, err := countStaleArtifacts(baseDir, citations, 90); err == nil {
		metrics.StaleArtifacts = staleCount
	}

	// Loop closure metrics
	metrics.LearningsCreated, metrics.LearningsFound = countLoopMetrics(baseDir, periodStart, stats.citations)
	if metrics.LearningsCreated > 0 {
		metrics.LoopClosureRatio = float64(metrics.LearningsFound) / float64(metrics.LearningsCreated)
	}
	metrics.PriorArtBypasses = countBypassCitations(stats.citations)

	// Retros
	retros, retrosWithLearnings, _ := countRetros(baseDir, periodStart)
	metrics.TotalRetros = retros
	metrics.RetrosWithLearnings = retrosWithLearnings

	// MemRL utility metrics
	utilityStats := computeUtilityMetrics(baseDir)
	metrics.MeanUtility = utilityStats.mean
	metrics.UtilityStdDev = utilityStats.stdDev
	metrics.HighUtilityCount = utilityStats.highCount
	metrics.LowUtilityCount = utilityStats.lowCount

	return metrics, nil
}

// countArtifacts counts knowledge artifacts by tier.
func countArtifacts(baseDir string) (int, map[string]int, error) {
	tierCounts := map[string]int{
		"observation": 0,
		"learning":    0,
		"pattern":     0,
		"skill":       0,
		"core":        0,
	}
	total := 0

	// Tier locations
	tierDirs := map[string]string{
		"observation": filepath.Join(baseDir, ".agents", "candidates"),
		"learning":    filepath.Join(baseDir, ".agents", "learnings"),
		"pattern":     filepath.Join(baseDir, ".agents", "patterns"),
	}

	for tier, dir := range tierDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}
		files, err := filepath.Glob(filepath.Join(dir, "*.md"))
		if err != nil {
			continue
		}
		// Also count JSONL
		jsonlFiles, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
		files = append(files, jsonlFiles...)

		tierCounts[tier] = len(files)
		total += len(files)
	}

	// Count research artifacts
	researchDir := filepath.Join(baseDir, ".agents", "research")
	if _, err := os.Stat(researchDir); err == nil {
		files, _ := filepath.Glob(filepath.Join(researchDir, "*.md"))
		tierCounts["observation"] += len(files)
		total += len(files)
	}

	// Count retros
	retrosDir := filepath.Join(baseDir, ".agents", "retros")
	if _, err := os.Stat(retrosDir); err == nil {
		files, _ := filepath.Glob(filepath.Join(retrosDir, "*.md"))
		tierCounts["learning"] += len(files)
		total += len(files)
	}

	// Count sessions
	sessionsDir := filepath.Join(baseDir, storage.DefaultBaseDir, storage.SessionsDir)
	if _, err := os.Stat(sessionsDir); err == nil {
		files, _ := filepath.Glob(filepath.Join(sessionsDir, "*.jsonl"))
		total += len(files)
	}

	return total, tierCounts, nil
}

// countNewArtifacts counts artifacts created after a time.
func countNewArtifacts(baseDir string, since time.Time) (int, error) {
	count := 0

	dirs := []string{
		filepath.Join(baseDir, ".agents", "learnings"),
		filepath.Join(baseDir, ".agents", "patterns"),
		filepath.Join(baseDir, ".agents", "candidates"),
		filepath.Join(baseDir, ".agents", "research"),
		filepath.Join(baseDir, ".agents", "retros"),
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}
		_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error { //nolint:errcheck
			if err != nil || info.IsDir() {
				return nil
			}
			if info.ModTime().After(since) {
				count++
			}
			return nil
		})
	}

	return count, nil
}

// countStaleArtifacts counts artifacts not cited in N days.
func countStaleArtifacts(baseDir string, citations []types.CitationEvent, staleDays int) (int, error) {
	staleThreshold := time.Now().AddDate(0, 0, -staleDays)

	// Build set of recently cited artifacts
	recentlyCited := make(map[string]bool)
	for _, c := range citations {
		if c.CitedAt.After(staleThreshold) {
			recentlyCited[c.ArtifactPath] = true
		}
	}

	// Count all artifacts, mark stale if not recently cited
	staleCount := 0
	dirs := []string{
		filepath.Join(baseDir, ".agents", "learnings"),
		filepath.Join(baseDir, ".agents", "patterns"),
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}
		_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error { //nolint:errcheck
			if err != nil || info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".md") && !strings.HasSuffix(path, ".jsonl") {
				return nil
			}
			// Check if recently cited
			absPath, _ := filepath.Abs(path)
			if !recentlyCited[absPath] {
				staleCount++
			}
			return nil
		})
	}

	return staleCount, nil
}

// saveBaseline saves metrics to a baseline file.
func saveBaseline(baseDir string, metrics *types.FlywheelMetrics) (string, error) {
	metricsDir := filepath.Join(baseDir, ".agents", "ao", "metrics")
	if err := os.MkdirAll(metricsDir, 0755); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("baseline-%s.json", metrics.Timestamp.Format("2006-01-02"))
	path := filepath.Join(metricsDir, filename)

	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}

	return path, nil
}

// printMetricsTable prints a formatted metrics table.
func printMetricsTable(m *types.FlywheelMetrics) {
	fmt.Println()
	fmt.Println("Knowledge Flywheel Metrics")
	fmt.Println("==========================")
	fmt.Printf("Period: %s to %s\n\n",
		m.PeriodStart.Format("2006-01-02"),
		m.PeriodEnd.Format("2006-01-02"))

	fmt.Println("PARAMETERS:")
	fmt.Printf("  δ (decay rate):     %.2f/week (literature baseline)\n", m.Delta)
	fmt.Printf("  σ (retrieval):      %.2f (%d%% relevant artifacts surfaced)\n",
		m.Sigma, int(m.Sigma*100))
	fmt.Printf("  ρ (citation rate):  %.2f refs/artifact/week\n", m.Rho)
	fmt.Println()

	fmt.Println("DERIVED:")
	fmt.Printf("  σ × ρ = %.3f\n", m.SigmaRho)
	fmt.Printf("  δ     = %.3f\n", m.Delta)
	fmt.Println("  ────────────────")

	velocitySign := "+"
	if m.Velocity < 0 {
		velocitySign = ""
	}
	status := m.EscapeVelocityStatus()
	statusIndicator := "✗"
	if m.AboveEscapeVelocity {
		statusIndicator = "✓"
	}
	fmt.Printf("  VELOCITY: %s%.3f/week (%s %s)\n", velocitySign, m.Velocity, status, statusIndicator)
	fmt.Println()

	fmt.Println("COUNTS:")
	fmt.Printf("  Knowledge items:    %d\n", m.TotalArtifacts)
	fmt.Printf("  Citation events:    %d this period\n", m.CitationsThisPeriod)
	fmt.Printf("  Unique cited:       %d\n", m.UniqueCitedArtifacts)
	fmt.Printf("  New artifacts:      %d\n", m.NewArtifacts)
	fmt.Printf("  Stale (90d+):       %d\n", m.StaleArtifacts)
	fmt.Println()

	if len(m.TierCounts) > 0 {
		fmt.Println("TIER DISTRIBUTION:")
		tiers := []string{"observation", "learning", "pattern", "skill", "core"}
		for _, tier := range tiers {
			if count, ok := m.TierCounts[tier]; ok && count > 0 {
				fmt.Printf("  %-12s: %d\n", tier, count)
			}
		}
		fmt.Println()
	}

	fmt.Printf("STATUS: %s\n", status)

	// Loop closure metrics section
	if m.LearningsCreated > 0 || m.LearningsFound > 0 || m.TotalRetros > 0 {
		fmt.Println()
		fmt.Println("LOOP CLOSURE (R1):")
		fmt.Printf("  Learnings created:  %d\n", m.LearningsCreated)
		fmt.Printf("  Learnings found:    %d\n", m.LearningsFound)
		loopStatus := "OPEN"
		if m.LoopClosureRatio >= 1.0 {
			loopStatus = "CLOSED ✓"
		} else if m.LoopClosureRatio > 0 {
			loopStatus = "PARTIAL"
		}
		fmt.Printf("  Closure ratio:      %.2f (%s)\n", m.LoopClosureRatio, loopStatus)
		if m.TotalRetros > 0 {
			fmt.Printf("  Retros:             %d (%d with learnings)\n", m.TotalRetros, m.RetrosWithLearnings)
		}
		if m.PriorArtBypasses > 0 {
			fmt.Printf("  Prior art bypasses: %d (review recommended)\n", m.PriorArtBypasses)
		}
	}

	// MemRL utility metrics section
	if m.MeanUtility > 0 || m.HighUtilityCount > 0 || m.LowUtilityCount > 0 {
		fmt.Println()
		fmt.Println("UTILITY (MemRL):")
		fmt.Printf("  Mean utility:        %.3f\n", m.MeanUtility)
		fmt.Printf("  Std deviation:       %.3f\n", m.UtilityStdDev)
		fmt.Printf("  High utility (>0.7): %d\n", m.HighUtilityCount)
		fmt.Printf("  Low utility (<0.3):  %d\n", m.LowUtilityCount)

		// Health indicator
		if m.MeanUtility >= 0.6 {
			fmt.Printf("  Status:              HEALTHY ✓ (learnings are effective)\n")
		} else if m.MeanUtility >= 0.4 {
			fmt.Printf("  Status:              NEUTRAL (need more feedback data)\n")
		} else {
			fmt.Printf("  Status:              REVIEW ✗ (learnings may need updating)\n")
		}
	}
}

// detectSessionID tries to detect the current session ID.
func detectSessionID() string {
	// Check CLAUDE_SESSION_ID env var
	if id := os.Getenv("CLAUDE_SESSION_ID"); id != "" {
		return id
	}

	// Check for session file in current dir
	// This is a fallback - real session ID should come from Claude
	return fmt.Sprintf("session-%s", time.Now().Format("20060102-150405"))
}

// countNewArtifactsInDir counts artifacts created after a time in a specific directory.
func countNewArtifactsInDir(dir string, since time.Time) (int, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return 0, nil
	}

	count := 0
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error { //nolint:errcheck
		if err != nil || info.IsDir() {
			return nil
		}
		if info.ModTime().After(since) {
			count++
		}
		return nil
	})

	return count, nil
}

// countRetros counts retro artifacts and how many have associated learnings.
func countRetros(baseDir string, since time.Time) (total int, withLearnings int, err error) {
	retrosDir := filepath.Join(baseDir, ".agents", "retros")
	if _, err := os.Stat(retrosDir); os.IsNotExist(err) {
		return 0, 0, nil
	}

	_ = filepath.Walk(retrosDir, func(path string, info os.FileInfo, err error) error { //nolint:errcheck
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}
		if info.ModTime().After(since) {
			total++
			// Check if retro has learnings section
			content, readErr := os.ReadFile(path)
			if readErr == nil {
				text := string(content)
				if strings.Contains(text, "## Learnings") ||
					strings.Contains(text, "## Key Learnings") ||
					strings.Contains(text, "### Learnings") {
					withLearnings++
				}
			}
		}
		return nil
	})

	return total, withLearnings, nil
}

// utilityStats holds computed utility statistics.
type utilityStats struct {
	mean      float64
	stdDev    float64
	highCount int // utility > 0.7
	lowCount  int // utility < 0.3
}

// computeUtilityMetrics calculates MemRL utility statistics from learnings.
func computeUtilityMetrics(baseDir string) utilityStats {
	var stats utilityStats
	var utilities []float64

	learningsDir := filepath.Join(baseDir, ".agents", "learnings")
	if _, err := os.Stat(learningsDir); os.IsNotExist(err) {
		return stats
	}

	// Scan JSONL files for utility values
	_ = filepath.Walk(learningsDir, func(path string, info os.FileInfo, err error) error { //nolint:errcheck
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".jsonl") {
			return nil
		}

		utility := parseUtilityFromFile(path)
		if utility > 0 {
			utilities = append(utilities, utility)
		}
		return nil
	})

	if len(utilities) == 0 {
		return stats
	}

	// Calculate mean
	var sum float64
	for _, u := range utilities {
		sum += u
	}
	stats.mean = sum / float64(len(utilities))

	// Calculate standard deviation
	var variance float64
	for _, u := range utilities {
		variance += (u - stats.mean) * (u - stats.mean)
	}
	stats.stdDev = math.Sqrt(variance / float64(len(utilities)))

	// Count high/low utility
	for _, u := range utilities {
		if u > 0.7 {
			stats.highCount++
		}
		if u < 0.3 {
			stats.lowCount++
		}
	}

	return stats
}

// parseUtilityFromFile extracts utility value from a JSONL file.
func parseUtilityFromFile(path string) float64 {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		var data map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &data); err == nil {
			if utility, ok := data["utility"].(float64); ok {
				return utility
			}
		}
	}
	return 0
}
