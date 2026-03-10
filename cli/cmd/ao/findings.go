package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/boshu2/agentops/cli/internal/config"
)

type findingStats struct {
	Total           int                `json:"total"`
	ByStatus        map[string]int     `json:"by_status"`
	BySeverity      map[string]int     `json:"by_severity"`
	ByDetectability map[string]int     `json:"by_detectability"`
	TotalHits       int                `json:"total_hits"`
	MostCited       []knowledgeFinding `json:"most_cited,omitempty"`
}

var (
	findingsListLimit   int
	findingsListAll     bool
	findingsExportTo    string
	findingsExportAll   bool
	findingsExportForce bool
	findingsPullFrom    string
	findingsPullAll     bool
	findingsPullForce   bool
	findingsRetireBy    string
)

var findingsCmd = &cobra.Command{
	Use:   "findings",
	Short: "Manage promoted findings",
	Long: `Manage promoted finding artifacts under .agents/findings/.

Use this command family to inspect, export, import, retire, and summarize
promoted findings. The lookup/inject runtime consumes these artifacts for
cross-session prevention, and active citations update their lifecycle fields.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var findingsListCmd = &cobra.Command{
	Use:   "list [query]",
	Short: "List active findings",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runFindingsList,
}

var findingsExportCmd = &cobra.Command{
	Use:   "export <id...>",
	Short: "Export finding artifacts to another repo or findings directory",
	Args:  cobra.ArbitraryArgs,
	RunE:  runFindingsExport,
}

var findingsPullCmd = &cobra.Command{
	Use:   "pull <id...>",
	Short: "Pull finding artifacts from another repo or findings directory",
	Args:  cobra.ArbitraryArgs,
	RunE:  runFindingsPull,
}

var findingsRetireCmd = &cobra.Command{
	Use:   "retire <id>",
	Short: "Retire a finding artifact",
	Args:  cobra.ExactArgs(1),
	RunE:  runFindingsRetire,
}

var findingsStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Summarize local finding artifact inventory",
	RunE:  runFindingsStats,
}

func init() {
	findingsCmd.GroupID = "knowledge"
	rootCmd.AddCommand(findingsCmd)
	findingsCmd.AddCommand(findingsListCmd)
	findingsCmd.AddCommand(findingsExportCmd)
	findingsCmd.AddCommand(findingsPullCmd)
	findingsCmd.AddCommand(findingsRetireCmd)
	findingsCmd.AddCommand(findingsStatsCmd)

	findingsListCmd.Flags().IntVar(&findingsListLimit, "limit", 20, "Maximum findings to return")
	findingsListCmd.Flags().BoolVar(&findingsListAll, "all", false, "Include retired and superseded findings")

	findingsExportCmd.Flags().StringVar(&findingsExportTo, "to", "", "Destination repo root or .agents/findings directory")
	findingsExportCmd.Flags().BoolVar(&findingsExportAll, "all", false, "Export every local finding")
	findingsExportCmd.Flags().BoolVar(&findingsExportForce, "force", false, "Overwrite destination files if they already exist")

	findingsPullCmd.Flags().StringVar(&findingsPullFrom, "from", "", "Source repo root or .agents/findings directory")
	findingsPullCmd.Flags().BoolVar(&findingsPullAll, "all", false, "Pull every source finding")
	findingsPullCmd.Flags().BoolVar(&findingsPullForce, "force", false, "Overwrite local files if they already exist")

	findingsRetireCmd.Flags().StringVar(&findingsRetireBy, "by", "", "Retired-by marker (defaults to current user or ao findings retire)")
}

func runFindingsList(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	query := ""
	if len(args) > 0 {
		query = args[0]
	}

	cfg, _ := config.Load(nil)
	globalFindingsDir := globalFindingsDirFromConfig(cfg)
	globalWeight := 0.8
	if cfg != nil {
		globalWeight = cfg.Paths.GlobalWeight
	}

	findings, err := collectFindingsWithOptions(cwd, query, findingsListLimit, globalFindingsDir, globalWeight, findingsListAll)
	if err != nil {
		return err
	}

	if GetOutput() == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(findings)
	}

	if len(findings) == 0 {
		fmt.Println("No findings found.")
		return nil
	}

	for _, f := range findings {
		scope := "local"
		if f.Global {
			scope = "global"
		}
		fmt.Printf("%s\t%s\t%d\t%s\t%s\n", f.ID, emptyIfMissing(f.Status), f.HitCount, scope, f.Title)
	}
	return nil
}

func runFindingsExport(cmd *cobra.Command, args []string) error {
	if findingsExportTo == "" {
		return fmt.Errorf("--to is required")
	}
	if !findingsExportAll && len(args) == 0 {
		return fmt.Errorf("provide one or more finding IDs or use --all")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	sourceDir := repoFindingsDir(cwd)
	targetDir := resolveManagedFindingsDir(findingsExportTo)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("create export dir: %w", err)
	}

	selected, err := selectFindingFiles(sourceDir, args, findingsExportAll)
	if err != nil {
		return err
	}

	var copied []string
	for _, src := range selected {
		dst := filepath.Join(targetDir, filepath.Base(src))
		if err := copyFindingFile(src, dst, findingsExportForce); err != nil {
			return err
		}
		copied = append(copied, dst)
	}

	return printFindingTransferResult("exported", copied)
}

func runFindingsPull(cmd *cobra.Command, args []string) error {
	if findingsPullFrom == "" {
		return fmt.Errorf("--from is required")
	}
	if !findingsPullAll && len(args) == 0 {
		return fmt.Errorf("provide one or more finding IDs or use --all")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	sourceDir, err := resolveExistingFindingsDir(findingsPullFrom)
	if err != nil {
		return err
	}
	targetDir := repoFindingsDir(cwd)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("create local findings dir: %w", err)
	}

	selected, err := selectFindingFiles(sourceDir, args, findingsPullAll)
	if err != nil {
		return err
	}

	var copied []string
	for _, src := range selected {
		dst := filepath.Join(targetDir, filepath.Base(src))
		if err := copyFindingFile(src, dst, findingsPullForce); err != nil {
			return err
		}
		copied = append(copied, dst)
	}
	bestEffortRefreshFindingCompiler(cwd)
	return printFindingTransferResult("pulled", copied)
}

func runFindingsRetire(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	finding, err := findLocalFindingByID(cwd, args[0])
	if err != nil {
		return err
	}

	retiredBy := findingsRetireBy
	if retiredBy == "" {
		retiredBy = strings.TrimSpace(os.Getenv("USER"))
	}
	if retiredBy == "" {
		retiredBy = "ao findings retire"
	}

	if err := updateFindingFrontMatter(finding.Source, map[string]string{
		"status":     "retired",
		"retired_by": retiredBy,
	}); err != nil {
		return err
	}
	bestEffortRefreshFindingCompiler(cwd)

	if GetOutput() == "json" {
		result := map[string]string{
			"id":         finding.ID,
			"status":     "retired",
			"retired_by": retiredBy,
			"source":     finding.Source,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	fmt.Printf("Retired finding %s (%s)\n", finding.ID, retiredBy)
	return nil
}

func runFindingsStats(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	findings, err := collectFindingsFromDir(repoFindingsDir(cwd), "", time.Now(), false, true)
	if err != nil {
		return err
	}
	stats := buildFindingStats(findings)

	if GetOutput() == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(stats)
	}

	fmt.Printf("Total findings: %d\n", stats.Total)
	fmt.Printf("Total hits: %d\n", stats.TotalHits)
	fmt.Println("By status:")
	printStringCountMap(stats.ByStatus)
	fmt.Println("By severity:")
	printStringCountMap(stats.BySeverity)
	fmt.Println("By detectability:")
	printStringCountMap(stats.ByDetectability)
	if len(stats.MostCited) > 0 {
		fmt.Println("Most cited:")
		for _, f := range stats.MostCited {
			fmt.Printf("  %s (%d hits)\n", f.ID, f.HitCount)
		}
	}
	return nil
}

func globalFindingsDirFromConfig(cfg *config.Config) string {
	if cfg == nil || cfg.Paths.GlobalLearningsDir == "" {
		return ""
	}
	return filepath.Join(filepath.Dir(cfg.Paths.GlobalLearningsDir), SectionFindings)
}

func repoFindingsDir(cwd string) string {
	return filepath.Join(cwd, ".agents", SectionFindings)
}

func resolveManagedFindingsDir(path string) string {
	clean := filepath.Clean(path)
	if filepath.Base(clean) == SectionFindings {
		return clean
	}
	return filepath.Join(clean, ".agents", SectionFindings)
}

func resolveExistingFindingsDir(path string) (string, error) {
	candidate := resolveManagedFindingsDir(path)
	if info, err := os.Stat(candidate); err == nil && info.IsDir() {
		return candidate, nil
	}
	if info, err := os.Stat(filepath.Clean(path)); err == nil && info.IsDir() && filepath.Base(filepath.Clean(path)) == SectionFindings {
		return filepath.Clean(path), nil
	}
	return "", fmt.Errorf("no findings directory found at %s", path)
}

func selectFindingFiles(dir string, ids []string, all bool) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	if all {
		return files, nil
	}
	selected := make([]string, 0, len(ids))
	for _, id := range ids {
		found := ""
		for _, file := range files {
			finding, err := parseFindingFile(file)
			if err != nil {
				continue
			}
			if matchesID(finding.ID, file, id) {
				found = file
				break
			}
		}
		if found == "" {
			return nil, fmt.Errorf("finding %q not found in %s", id, dir)
		}
		selected = append(selected, found)
	}
	return selected, nil
}

func findLocalFindingByID(cwd, id string) (knowledgeFinding, error) {
	dir := repoFindingsDir(cwd)
	files, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return knowledgeFinding{}, err
	}
	for _, file := range files {
		finding, err := parseFindingFile(file)
		if err != nil {
			continue
		}
		if matchesID(finding.ID, file, id) {
			return finding, nil
		}
	}
	return knowledgeFinding{}, fmt.Errorf("finding %q not found", id)
}

func copyFindingFile(src, dst string, force bool) error {
	if !force {
		if _, err := os.Stat(dst); err == nil {
			return fmt.Errorf("destination already exists: %s", dst)
		}
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read %s: %w", src, err)
	}
	return writeFindingFileAtomic(dst, data, 0o644)
}

func printFindingTransferResult(action string, paths []string) error {
	if GetOutput() == "json" {
		result := map[string]any{
			"action": action,
			"files":  paths,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}
	if len(paths) == 0 {
		fmt.Printf("No findings %s.\n", action)
		return nil
	}
	fmt.Printf("%s %d finding(s):\n", strings.Title(action), len(paths))
	for _, path := range paths {
		fmt.Printf("  %s\n", path)
	}
	return nil
}

func buildFindingStats(findings []knowledgeFinding) findingStats {
	stats := findingStats{
		Total:           len(findings),
		ByStatus:        make(map[string]int),
		BySeverity:      make(map[string]int),
		ByDetectability: make(map[string]int),
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].HitCount == findings[j].HitCount {
			return findings[i].ID < findings[j].ID
		}
		return findings[i].HitCount > findings[j].HitCount
	})
	for _, finding := range findings {
		stats.ByStatus[normalizeStatKey(finding.Status, "unknown")]++
		stats.BySeverity[normalizeStatKey(finding.Severity, "unknown")]++
		stats.ByDetectability[normalizeStatKey(finding.Detectability, "unknown")]++
		stats.TotalHits += finding.HitCount
	}
	if len(findings) > 5 {
		stats.MostCited = findings[:5]
	} else {
		stats.MostCited = findings
	}
	return stats
}

func normalizeStatKey(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func printStringCountMap(values map[string]int) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("  %s: %d\n", key, values[key])
	}
}

func updateFindingFrontMatter(path string, updates map[string]string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read finding: %w", err)
	}

	lines := strings.Split(strings.TrimRight(string(content), "\n"), "\n")
	frontMatterEnd := -1
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				frontMatterEnd = i
				break
			}
		}
	}

	frontMatter := []string{}
	body := lines
	if frontMatterEnd >= 0 {
		frontMatter = append(frontMatter, lines[1:frontMatterEnd]...)
		body = lines[frontMatterEnd+1:]
	}

	keys := make([]string, 0, len(updates))
	for key := range updates {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	indexByKey := make(map[string]int)
	for i, line := range frontMatter {
		trimmed := strings.TrimSpace(line)
		for _, key := range keys {
			if strings.HasPrefix(trimmed, key+":") {
				indexByKey[key] = i
			}
		}
	}

	for _, key := range keys {
		line := fmt.Sprintf("%s: %s", key, updates[key])
		if idx, ok := indexByKey[key]; ok {
			frontMatter[idx] = line
			continue
		}
		frontMatter = append(frontMatter, line)
	}

	outLines := []string{"---"}
	outLines = append(outLines, frontMatter...)
	outLines = append(outLines, "---")
	outLines = append(outLines, body...)
	out := strings.Join(outLines, "\n")
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	return writeFindingFileAtomic(path, []byte(out), 0o644)
}

func writeFindingFileAtomic(path string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp.*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func bestEffortRefreshFindingCompiler(cwd string) {
	script := filepath.Join(cwd, "hooks", "finding-compiler.sh")
	if _, err := os.Stat(script); err != nil {
		return
	}
	cmd := exec.Command("bash", script, "--quiet")
	cmd.Dir = cwd
	cmd.Stdout = nil
	cmd.Stderr = nil
	_ = cmd.Run()
}
