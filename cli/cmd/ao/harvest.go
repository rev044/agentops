package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/boshu2/agentops/cli/internal/harvest"
	"github.com/spf13/cobra"
)

var (
	harvestRootsFlag     string
	harvestOutputDir     string
	harvestPromoteTo     string
	harvestMinConfidence float64
	harvestInclude       string
	harvestQuiet         bool
	harvestMaxFileSize   int64
)

var harvestCmd = &cobra.Command{
	Use:   "harvest",
	Short: "Sweep all rigs, extract and deduplicate cross-rig knowledge",
	Long: `Walks all .agents/ directories across the workspace, extracts learnings,
patterns, and research, deduplicates across rigs, and promotes high-value
items to the global knowledge hub (~/.agents/learnings/).`,
	RunE: runHarvest,
}

func init() {
	harvestCmd.GroupID = "knowledge"
	rootCmd.AddCommand(harvestCmd)

	harvestCmd.Flags().StringVar(&harvestRootsFlag, "roots", "",
		"Base directories to scan (comma-separated) (default ~/gt)")
	harvestCmd.Flags().StringVar(&harvestOutputDir, "output-dir", ".agents/harvest",
		"Directory for harvest catalog output")
	harvestCmd.Flags().StringVar(&harvestPromoteTo, "promote-to", "",
		"Promotion destination for high-value artifacts (default ~/.agents/learnings)")
	harvestCmd.Flags().Float64Var(&harvestMinConfidence, "min-confidence", 0.5,
		"Minimum confidence for promotion")
	harvestCmd.Flags().StringVar(&harvestInclude, "include", "learnings,patterns,research",
		"Artifact types to include (comma-separated)")
	harvestCmd.Flags().BoolVar(&harvestQuiet, "quiet", false, "Suppress progress output")
	harvestCmd.Flags().Int64Var(&harvestMaxFileSize, "max-file-size", 1048576,
		"Skip files larger than this (bytes)")
}

func runHarvest(cmd *cobra.Command, args []string) error {
	// Resolve home-relative defaults at runtime so generated docs don't embed absolute paths.
	if harvestRootsFlag == "" {
		home, _ := os.UserHomeDir()
		defaultRoots := []string{filepath.Join(home, "gt")}
		// Include Claude project directories if they exist
		claudeProjects := filepath.Join(home, ".claude", "projects")
		if info, err := os.Stat(claudeProjects); err == nil && info.IsDir() {
			defaultRoots = append(defaultRoots, claudeProjects)
		}
		harvestRootsFlag = strings.Join(defaultRoots, ",")
	}
	if harvestPromoteTo == "" {
		home, _ := os.UserHomeDir()
		harvestPromoteTo = filepath.Join(home, ".agents", "learnings")
	}

	roots := strings.Split(harvestRootsFlag, ",")
	for i := range roots {
		roots[i] = strings.TrimSpace(roots[i])
	}

	includeDirs := strings.Split(harvestInclude, ",")
	for i := range includeDirs {
		includeDirs[i] = strings.TrimSpace(includeDirs[i])
	}

	opts := harvest.WalkOptions{
		Roots:       roots,
		MaxFileSize: harvestMaxFileSize,
		SkipDirs:    harvest.DefaultWalkOptions().SkipDirs,
		IncludeDirs: includeDirs,
	}

	rigs, discoveryWarnings, err := harvest.DiscoverRigsWithWarnings(opts)
	if err != nil {
		return fmt.Errorf("discovering rigs: %w", err)
	}

	if !harvestQuiet {
		fmt.Printf("Discovered %d rigs\n", len(rigs))
	}

	var allArtifacts []harvest.Artifact
	warnings := append([]harvest.HarvestWarning{}, discoveryWarnings...)
	for _, warning := range discoveryWarnings {
		printHarvestWarning(warning)
	}

	totalCandidateFiles := 0
	for _, rig := range rigs {
		result := harvest.ExtractArtifactsWithStats(rig, opts)
		totalCandidateFiles += result.CandidateFiles
		if len(result.Warnings) > 0 {
			for _, warning := range result.Warnings {
				printHarvestWarning(warning)
			}
			warnings = append(warnings, result.Warnings...)
		}
		allArtifacts = append(allArtifacts, result.Artifacts...)
	}

	catalog := harvest.BuildCatalog(allArtifacts, harvestMinConfidence)
	catalog.Roots = append([]string{}, roots...)
	catalog.IncludeDirs = append([]string{}, includeDirs...)
	catalog.PromoteTo = harvestPromoteTo
	catalog.MinConfidence = harvestMinConfidence
	catalog.DryRun = GetDryRun()
	catalog.Rigs = append([]harvest.RigInfo{}, rigs...)
	catalog.Warnings = append([]harvest.HarvestWarning{}, warnings...)
	catalog.RigsScanned = len(rigs)
	catalog.TotalFiles = totalCandidateFiles
	catalog.Timestamp = time.Now().UTC()
	catalog.Summary.WarningCount = len(catalog.Warnings)

	if !harvestQuiet {
		fmt.Printf("Extracted %d artifacts (%d unique, %d duplicate excess, %d promotion candidates, %d warnings)\n",
			catalog.Summary.ArtifactsExtracted,
			catalog.Summary.UniqueArtifacts,
			catalog.Summary.DuplicateExcess,
			catalog.Summary.PromotionCandidates,
			catalog.Summary.WarningCount,
		)
	}

	promoted := 0
	if !GetDryRun() {
		var promoteErr error
		promoted, promoteErr = harvest.Promote(catalog, harvestPromoteTo, false)
		if promoteErr != nil {
			return fmt.Errorf("promoting artifacts: %w", promoteErr)
		}
	}
	catalog.PromotionCount = promoted

	outputDir := harvestOutputDir
	if err := harvest.WriteCatalog(outputDir, catalog); err != nil {
		return fmt.Errorf("writing catalog: %w", err)
	}

	if GetOutput() == "json" {
		data, marshalErr := json.MarshalIndent(catalog, "", "  ")
		if marshalErr != nil {
			return fmt.Errorf("marshaling catalog: %w", marshalErr)
		}
		fmt.Println(string(data))
		return nil
	}

	if !harvestQuiet {
		fmt.Printf("Catalog written to %s\n", outputDir)
		if !GetDryRun() {
			fmt.Printf("Promoted %d artifacts to %s\n", promoted, harvestPromoteTo)
		} else {
			fmt.Println("Dry run: no artifacts promoted")
		}
	}

	VerbosePrintf("Rigs scanned: %d, Total files: %d\n", catalog.RigsScanned, catalog.TotalFiles)

	return nil
}

func printHarvestWarning(warning harvest.HarvestWarning) {
	fmt.Fprintf(os.Stderr, "harvest: warning: %s: %s\n", warning.Stage, warning.Message)
}

// duplicateArtifactCount returns the total number of duplicate artifacts
// (excluding the kept winner from each group).
func duplicateArtifactCount(cat *harvest.Catalog) int {
	count := 0
	for _, dg := range cat.Duplicates {
		count += dg.Count - 1
	}
	return count
}
