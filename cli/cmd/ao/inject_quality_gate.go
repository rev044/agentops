package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// qualityReport holds detection results for the learning pool.
type qualityReport struct {
	TotalLearnings  int      `json:"total_learnings"`
	WithSource      int      `json:"with_source"`
	WithoutSource   int      `json:"without_source"`
	StaleCount      int      `json:"stale_count"`
	DuplicateGroups int      `json:"duplicate_groups"`
	FlaggedPaths    []string `json:"flagged_paths"`
	Score           float64  `json:"score"`
}

// scanLearningQuality scans a learnings directory and produces a quality report.
func scanLearningQuality(learningsDir string) (*qualityReport, error) {
	report := &qualityReport{
		FlaggedPaths: []string{},
	}

	entries, err := os.ReadDir(learningsDir)
	if err != nil {
		return report, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(learningsDir, entry.Name())
		hasSource, isStale, assessErr := assessLearningFile(path)
		if assessErr != nil {
			continue
		}

		report.TotalLearnings++

		if hasSource {
			report.WithSource++
		} else {
			report.WithoutSource++
		}

		if isStale {
			report.StaleCount++
		}

		if !hasSource && isStale {
			report.FlaggedPaths = append(report.FlaggedPaths, path)
		}
	}

	if report.TotalLearnings == 0 {
		report.Score = 0
		return report, nil
	}

	sourceRatio := float64(report.WithSource) / float64(report.TotalLearnings)
	staleRatio := float64(report.StaleCount) / float64(report.TotalLearnings)
	report.Score = sourceRatio * (1 - staleRatio)

	return report, nil
}

// assessLearningFile checks a single learning file for quality indicators.
// hasSource is true when the file contains a source_bead field in its frontmatter.
// isStale is true when the file's mtime is older than 90 days and no last_reward_at
// appears in the frontmatter within the last 90 days.
func assessLearningFile(path string) (hasSource bool, isStale bool, err error) {
	f, err := os.Open(path)
	if err != nil {
		return false, false, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return false, false, err
	}

	const staleDays = 90
	staleThreshold := time.Now().Add(-staleDays * 24 * time.Hour)
	mtimeStale := info.ModTime().Before(staleThreshold)

	scanner := bufio.NewScanner(f)
	inFrontMatter := false
	var lastRewardAt time.Time

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "---" {
			if !inFrontMatter {
				inFrontMatter = true
				continue
			}
			// End of frontmatter
			break
		}

		if !inFrontMatter {
			continue
		}

		if strings.HasPrefix(line, "source_bead:") || strings.HasPrefix(line, "source-bead:") {
			val := strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			if val != "" && val != "null" && val != "~" {
				hasSource = true
			}
		}

		if strings.HasPrefix(line, "last_reward_at:") || strings.HasPrefix(line, "last-reward-at:") {
			val := strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			if t, parseErr := time.Parse(time.RFC3339, val); parseErr == nil {
				lastRewardAt = t
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return hasSource, false, err
	}

	// Stale = mtime old AND no recent reward
	if mtimeStale {
		if lastRewardAt.IsZero() || lastRewardAt.Before(staleThreshold) {
			isStale = true
		}
	}

	return hasSource, isStale, nil
}
