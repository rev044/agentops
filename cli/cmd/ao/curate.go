package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/boshu2/agentops/cli/internal/goals"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// curateOutWriter returns cmd.OutOrStdout() when cmd is non-nil, otherwise os.Stdout.
func curateOutWriter(cmd *cobra.Command) io.Writer {
	if cmd != nil {
		return cmd.OutOrStdout()
	}
	return os.Stdout
}

// validArtifactTypes enumerates the allowed artifact type values.
var validArtifactTypes = map[string]bool{
	"learning": true,
	"decision": true,
	"failure":  true,
	"pattern":  true,
}

// curateArtifact represents a cataloged knowledge artifact.
type curateArtifact struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	Content       string `json:"content"`
	Date          string `json:"date"`
	SchemaVersion int    `json:"schema_version"`
	CuratedAt     string `json:"curated_at"`
	Path          string `json:"path"`
}

// curateVerifyResult holds the output of a verify operation.
type curateVerifyResult struct {
	Verified    bool     `json:"verified"`
	GatesPassed int      `json:"gates_passed"`
	GatesFailed int      `json:"gates_failed"`
	Regressions []string `json:"regressions"`
}

// curateStatusResult holds the output of a status query.
type curateStatusResult struct {
	Learnings     int    `json:"learnings"`
	Decisions     int    `json:"decisions"`
	Failures      int    `json:"failures"`
	Patterns      int    `json:"patterns"`
	Total         int    `json:"total"`
	LastCatalogAt string `json:"last_catalog_at,omitempty"`
	LastVerifyAt  string `json:"last_verify_at,omitempty"`
	PendingVerify int    `json:"pending_verify"`
}

var curateVerifySince string

var curateCmd = &cobra.Command{
	Use:   "curate",
	Short: "Curation pipeline operations",
	Long: `Curate manages the knowledge curation pipeline: catalog artifacts,
verify gate health, and view status.

Commands:
  catalog <path>   Catalog a knowledge artifact
  verify           Verify gate health against baselines
  status           Show curation pipeline status`,
}

var curateCatalogCmd = &cobra.Command{
	Use:   "catalog <path>",
	Short: "Catalog a knowledge artifact",
	Args:  cobra.ExactArgs(1),
	RunE:  runCurateCatalog,
}

var curateVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify gate health against baselines",
	RunE:  runCurateVerify,
}

var curateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show curation pipeline status",
	RunE:  runCurateStatus,
}

func init() {
	curateCmd.GroupID = "knowledge"
	rootCmd.AddCommand(curateCmd)

	curateCmd.AddCommand(curateCatalogCmd)
	curateCmd.AddCommand(curateVerifyCmd)
	curateCmd.AddCommand(curateStatusCmd)

	curateVerifyCmd.Flags().StringVar(&curateVerifySince, "since", "", "Filter to changes within duration (e.g. 24h, 7d)")
}

// curateParseFrontmatter extracts YAML frontmatter key-value pairs from a
// markdown document delimited by --- lines. Returns the frontmatter map and
// the body content below the closing delimiter.
func curateParseFrontmatter(data string) (map[string]any, string) {
	fm := make(map[string]any)

	lines := strings.Split(data, "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		// No frontmatter — treat entire file as content
		return fm, data
	}

	closeIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			closeIdx = i
			break
		}
	}
	if closeIdx < 0 {
		return fm, data
	}

	fmText := strings.Join(lines[1:closeIdx], "\n")
	if err := yaml.Unmarshal([]byte(fmText), &fm); err != nil {
		// Frontmatter delimiter exists but YAML is malformed; fall back to body-only.
		return make(map[string]any), strings.TrimSpace(strings.Join(lines[closeIdx+1:], "\n"))
	}

	body := strings.Join(lines[closeIdx+1:], "\n")
	return fm, strings.TrimSpace(body)
}

func curateFrontmatterString(fm map[string]any, key string) string {
	v, ok := fm[key]
	if !ok || v == nil {
		return ""
	}
	switch typed := v.(type) {
	case string:
		return strings.TrimSpace(typed)
	case time.Time:
		return typed.UTC().Format("2006-01-02")
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", typed))
	}
}

func resolveCurateGoalsFile() (string, error) {
	candidates := []string{"GOALS.md", "GOALS.yaml", "GOALS.yml"}
	for _, path := range candidates {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, nil
		}
	}
	return "", os.ErrNotExist
}

// generateArtifactID creates a unique ID based on artifact type, date, and content hash.
func generateArtifactID(artifactType, date, content string) string {
	var prefix string
	if artifactType == "learning" {
		prefix = "learn"
	} else if artifactType == "decision" {
		prefix = "decis"
	} else if artifactType == "failure" {
		prefix = "fail"
	} else if artifactType == "pattern" {
		prefix = "patt"
	}

	h := sha256.Sum256([]byte(content))
	shortHash := fmt.Sprintf("%x", h[:4])

	return fmt.Sprintf("%s-%s-%s", prefix, date, shortHash)
}

// artifactDir returns the target directory for the given artifact type.
func curateArtifactDir(artifactType string) string {
	if artifactType == "pattern" {
		return ".agents/patterns"
	}
	return ".agents/learnings"
}

func runCurateCatalog(cmd *cobra.Command, args []string) error {
	inputPath := args[0]

	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("reading artifact: %w", err)
	}

	fm, body := curateParseFrontmatter(string(data))

	// Validate required fields
	artifactType := curateFrontmatterString(fm, "type")
	if artifactType == "" {
		return fmt.Errorf("artifact missing required 'type' field in frontmatter")
	}

	if !validArtifactTypes[artifactType] {
		return fmt.Errorf("unknown artifact type %q: must be one of learning, decision, failure, pattern", artifactType)
	}

	content := body
	if content == "" {
		content = curateFrontmatterString(fm, "content")
	}
	if content == "" {
		return fmt.Errorf("artifact has no content (empty body and no 'content' frontmatter field)")
	}

	date := curateFrontmatterString(fm, "date")
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}

	// Assign ID if missing
	id := curateFrontmatterString(fm, "id")
	if id == "" {
		id = generateArtifactID(artifactType, date, content)
	}

	now := time.Now().UTC()

	artifact := curateArtifact{
		ID:            id,
		Type:          artifactType,
		Content:       content,
		Date:          date,
		SchemaVersion: 1,
		CuratedAt:     now.Format(time.RFC3339),
	}

	// Write to target directory
	targetDir := curateArtifactDir(artifactType)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("creating artifact dir: %w", err)
	}

	filename := fmt.Sprintf("%s.json", id)
	targetPath := filepath.Join(targetDir, filename)
	artifact.Path = targetPath

	artifactData, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling artifact: %w", err)
	}

	if err := os.WriteFile(targetPath, artifactData, 0o600); err != nil {
		return fmt.Errorf("writing artifact: %w", err)
	}

	// Output
	if GetOutput() == "json" {
		enc := json.NewEncoder(curateOutWriter(cmd))
		enc.SetIndent("", "  ")
		return enc.Encode(artifact)
	}

	fmt.Printf("Cataloged %s artifact: %s\n", artifactType, id)
	fmt.Printf("  Path: %s\n", targetPath)
	fmt.Printf("  Date: %s\n", date)
	return nil
}

func runCurateVerify(cmd *cobra.Command, args []string) error {
	result := curateVerifyResult{
		Regressions: []string{},
	}

	// Load goals and measure
	goalsPath, resolveErr := resolveCurateGoalsFile()
	if resolveErr != nil {
		// If no goals file, report zero gates
		if GetOutput() == "json" {
			result.Verified = true
			enc := json.NewEncoder(curateOutWriter(cmd))
			enc.SetIndent("", "  ")
			return enc.Encode(result)
		}
		fmt.Println("No GOALS file found — nothing to verify")
		return nil
	}

	gf, err := goals.LoadGoals(goalsPath)
	if err != nil {
		if GetOutput() == "json" {
			result.Verified = false
			enc := json.NewEncoder(curateOutWriter(cmd))
			enc.SetIndent("", "  ")
			return enc.Encode(result)
		}
		fmt.Printf("Could not load goals from %s — nothing to verify\n", goalsPath)
		return nil
	}

	timeout := 120 * time.Second
	snap := goals.Measure(gf, timeout)

	// Count pass/fail
	for _, m := range snap.Goals {
		if m.Result == "pass" {
			result.GatesPassed++
		} else if m.Result == "fail" {
			result.GatesFailed++
		}
	}

	// Check for uncommitted changes
	gitCmd := exec.Command("git", "status", "--porcelain")
	gitOut, gitErr := gitCmd.Output()
	if gitErr == nil && len(strings.TrimSpace(string(gitOut))) > 0 {
		VerbosePrintf("Uncommitted changes detected\n")
	}

	// Load baseline and compare for regressions
	baselineDir := ".agents/ao/baselines"
	regressions, regErr := detectVerifyRegressions(baselineDir, snap)
	if regErr != nil {
		return regErr
	}
	result.Regressions = append(result.Regressions, regressions...)

	result.Verified = len(result.Regressions) == 0 && result.GatesFailed == 0

	// Save current snapshot as new baseline
	if _, saveErr := goals.SaveSnapshot(snap, baselineDir); saveErr != nil {
		VerbosePrintf("Warning: could not save baseline: %v\n", saveErr)
	}

	// Output
	if GetOutput() == "json" {
		enc := json.NewEncoder(curateOutWriter(cmd))
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	if result.Verified {
		fmt.Printf("VERIFIED: %d gates passed, 0 regressions\n", result.GatesPassed)
	} else {
		fmt.Printf("NOT VERIFIED: %d passed, %d failed\n", result.GatesPassed, result.GatesFailed)
		if len(result.Regressions) > 0 {
			fmt.Printf("  Regressions: %s\n", strings.Join(result.Regressions, ", "))
		}
	}
	return nil
}

// detectVerifyRegressions loads the latest baseline and compares against a snapshot,
// respecting the --since filter. Returns regressed goal IDs.
func detectVerifyRegressions(baselineDir string, snap *goals.Snapshot) ([]string, error) {
	baseline, baseErr := goals.LoadLatestSnapshot(baselineDir)
	if baseErr != nil {
		return nil, nil
	}

	applyFilter := true
	if curateVerifySince != "" {
		dur, parseErr := parseDuration(curateVerifySince)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid --since value %q: %w", curateVerifySince, parseErr)
		}
		if dur < 0 {
			return nil, fmt.Errorf("--since value must be positive: %q", curateVerifySince)
		}
		cutoff := time.Now().Add(-dur)
		if ts, tsErr := time.Parse(time.RFC3339, baseline.Timestamp); tsErr == nil {
			if ts.Before(cutoff) {
				applyFilter = false
			}
		} else if ts, tsErr := time.Parse("2006-01-02T15:04:05.000", baseline.Timestamp); tsErr == nil {
			if ts.Before(cutoff) {
				applyFilter = false
			}
		}
	}

	if !applyFilter {
		return nil, nil
	}

	var regressions []string
	drifts := goals.ComputeDrift(baseline, snap)
	for _, d := range drifts {
		if d.Delta == "regressed" {
			regressions = append(regressions, d.GoalID)
		}
	}
	return regressions, nil
}

// countArtifactsInDir reads JSON artifacts from a directory and returns counts by type and the latest CuratedAt time.
func countArtifactsInDir(dir string) (counts map[string]int, latest time.Time) {
	counts = make(map[string]int)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return counts, latest
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, readErr := os.ReadFile(filepath.Join(dir, e.Name()))
		if readErr != nil {
			continue
		}
		var a curateArtifact
		if json.Unmarshal(data, &a) != nil {
			continue
		}
		counts[a.Type]++
		if t, err := time.Parse(time.RFC3339, a.CuratedAt); err == nil {
			if t.After(latest) {
				latest = t
			}
		}
	}
	return counts, latest
}

func runCurateStatus(cmd *cobra.Command, args []string) error {
	result := curateStatusResult{}

	// Count artifacts by type in .agents/learnings/
	learningsDir := ".agents/learnings"
	patternsDir := ".agents/patterns"

	var latestCatalog time.Time

	learningsCounts, learningsLatest := countArtifactsInDir(learningsDir)
	result.Learnings = learningsCounts["learning"]
	result.Decisions = learningsCounts["decision"]
	result.Failures = learningsCounts["failure"]
	if learningsLatest.After(latestCatalog) {
		latestCatalog = learningsLatest
	}

	patternsCounts, patternsLatest := countArtifactsInDir(patternsDir)
	result.Patterns = patternsCounts["pattern"]
	if patternsLatest.After(latestCatalog) {
		latestCatalog = patternsLatest
	}

	result.Total = result.Learnings + result.Decisions + result.Failures + result.Patterns

	if !latestCatalog.IsZero() {
		result.LastCatalogAt = latestCatalog.Format(time.RFC3339)
	}

	// Check last verify timestamp from baselines
	baselineDir := ".agents/ao/baselines"
	if entries, err := os.ReadDir(baselineDir); err == nil {
		var latest time.Time
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
				info, infoErr := e.Info()
				if infoErr == nil {
					if info.ModTime().After(latest) {
						latest = info.ModTime()
					}
				}
			}
		}
		if !latest.IsZero() {
			result.LastVerifyAt = latest.Format(time.RFC3339)
		}
	}

	// Pending: artifacts cataloged after last verify
	if result.LastVerifyAt != "" && result.LastCatalogAt != "" {
		verifyTime, _ := time.Parse(time.RFC3339, result.LastVerifyAt)
		catalogTime, _ := time.Parse(time.RFC3339, result.LastCatalogAt)
		if catalogTime.After(verifyTime) {
			result.PendingVerify = countArtifactsSince(learningsDir, patternsDir, verifyTime)
		}
	} else if result.Total > 0 && result.LastVerifyAt == "" {
		// Never verified — all are pending
		result.PendingVerify = result.Total
	}

	// Output
	if GetOutput() == "json" {
		enc := json.NewEncoder(curateOutWriter(cmd))
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	fmt.Println("Curation Pipeline Status")
	fmt.Println("========================")
	fmt.Printf("  Learnings:  %d\n", result.Learnings)
	fmt.Printf("  Decisions:  %d\n", result.Decisions)
	fmt.Printf("  Failures:   %d\n", result.Failures)
	fmt.Printf("  Patterns:   %d\n", result.Patterns)
	fmt.Printf("  Total:      %d\n", result.Total)
	fmt.Println()
	if result.LastCatalogAt != "" {
		fmt.Printf("  Last catalog: %s\n", result.LastCatalogAt)
	}
	if result.LastVerifyAt != "" {
		fmt.Printf("  Last verify:  %s\n", result.LastVerifyAt)
	}
	if result.PendingVerify > 0 {
		fmt.Printf("  Pending:      %d artifact(s) not yet verified\n", result.PendingVerify)
	}

	return nil
}

// countArtifactsSince counts artifacts in the given dirs with CuratedAt after the given time.
func countArtifactsSince(learningsDir, patternsDir string, since time.Time) int {
	count := 0
	for _, dir := range []string{learningsDir, patternsDir} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			data, readErr := os.ReadFile(filepath.Join(dir, e.Name()))
			if readErr != nil {
				continue
			}
			var a curateArtifact
			if json.Unmarshal(data, &a) != nil {
				continue
			}
			if t, err := time.Parse(time.RFC3339, a.CuratedAt); err == nil {
				if t.After(since) {
					count++
				}
			}
		}
	}
	return count
}

// parseDuration is defined in vibe_check.go — reuse it for --since flag parsing.
