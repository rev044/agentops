package main

import (
	"cmp"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/boshu2/agentops/cli/internal/types"
)

func collectFindings(cwd, query string, limit int, globalDir string, globalWeight float64) ([]finding, error) {
	findingsDir := filepath.Join(cwd, ".agents", SectionFindings)
	if _, err := os.Stat(findingsDir); os.IsNotExist(err) {
		findingsDir = findAgentsSubdir(cwd, SectionFindings)
	}

	queryLower := strings.ToLower(query)
	now := time.Now()

	local, err := collectFindingsFromDir(findingsDir, queryLower, now, false)
	if err != nil {
		return nil, err
	}

	localPaths := make(map[string]bool)
	localFiles, _ := filepath.Glob(filepath.Join(findingsDir, "*.md"))
	for _, f := range localFiles {
		if abs, err := filepath.Abs(f); err == nil {
			localPaths[abs] = true
		}
	}

	findings := append([]finding{}, local...)
	if globalDir != "" {
		globalFiles, _ := filepath.Glob(filepath.Join(globalDir, "*.md"))
		for _, file := range globalFiles {
			if abs, err := filepath.Abs(file); err == nil && localPaths[abs] {
				continue
			}
			f, err := parseFindingFile(file)
			if err != nil {
				continue
			}
			applyFindingFreshness(&f, file, now)
			if !findingMatchesQuery(f, queryLower) {
				continue
			}
			f.Global = true
			findings = append(findings, f)
		}
	}

	if len(findings) == 0 {
		return nil, nil
	}

	items := make([]scorable, len(findings))
	for i := range findings {
		items[i] = &findings[i]
	}
	applyCompositeScoringTo(items, types.DefaultLambda)

	if globalWeight > 0 && globalWeight < 1.0 {
		for i := range findings {
			if findings[i].Global {
				findings[i].CompositeScore *= globalWeight
			}
		}
	}

	slices.SortFunc(findings, func(a, b finding) int {
		return cmp.Compare(b.CompositeScore, a.CompositeScore)
	})
	if len(findings) > limit {
		findings = findings[:limit]
	}
	return findings, nil
}

func collectFindingsFromDir(dir, queryLower string, now time.Time, isGlobal bool) ([]finding, error) {
	if dir == "" {
		return nil, nil
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return nil, err
	}
	var result []finding
	for _, file := range files {
		f, err := parseFindingFile(file)
		if err != nil {
			continue
		}
		applyFindingFreshness(&f, file, now)
		if !findingMatchesQuery(f, queryLower) {
			continue
		}
		f.Global = isGlobal
		result = append(result, f)
	}
	return result, nil
}

func applyFindingFreshness(f *finding, file string, now time.Time) {
	info, err := os.Stat(file)
	if err != nil || info == nil {
		f.FreshnessScore = 0.5
		if f.Utility == 0 {
			f.Utility = types.InitialUtility
		}
		return
	}
	f.AgeWeeks = now.Sub(info.ModTime()).Hours() / (24 * 7)
	f.FreshnessScore = freshnessScore(f.AgeWeeks)
	if f.Utility == 0 {
		f.Utility = types.InitialUtility
	}
}

func findingMatchesQuery(f finding, queryLower string) bool {
	if queryLower == "" {
		return true
	}
	haystack := strings.ToLower(strings.Join([]string{
		f.ID,
		f.Title,
		f.Summary,
		f.SourceSkill,
		f.Severity,
		f.Detectability,
		strings.Join(f.ScopeTags, " "),
		strings.Join(f.CompilerTargets, " "),
	}, " "))
	return strings.Contains(haystack, queryLower)
}

func parseFindingFile(path string) (finding, error) {
	f := finding{
		ID:     strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
		Source: path,
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return f, err
	}
	lines := strings.Split(string(content), "\n")
	fm, contentStart := parseFrontMatter(lines)
	if fm.HasUtility {
		f.Utility = fm.Utility
	}
	for i := 1; i < len(lines) && i < contentStart; i++ {
		line := strings.TrimSpace(lines[i])
		switch {
		case strings.HasPrefix(line, "id:"):
			f.ID = trimField(line)
		case strings.HasPrefix(line, "title:"):
			f.Title = trimField(line)
		case strings.HasPrefix(line, "source_skill:"), strings.HasPrefix(line, "source-skill:"):
			f.SourceSkill = trimField(line)
		case strings.HasPrefix(line, "severity:"):
			f.Severity = trimField(line)
		case strings.HasPrefix(line, "detectability:"):
			f.Detectability = trimField(line)
		case strings.HasPrefix(line, "status:"):
			f.Status = trimField(line)
		case strings.HasPrefix(line, "compiler_targets:"), strings.HasPrefix(line, "compiler-targets:"):
			f.CompilerTargets = parseListField(trimField(line))
		case strings.HasPrefix(line, "scope_tags:"), strings.HasPrefix(line, "scope-tags:"):
			f.ScopeTags = parseListField(trimField(line))
		}
	}
	for i := contentStart; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "# ") && f.Title == "" {
			f.Title = strings.TrimPrefix(line, "# ")
		}
	}
	if f.Title == "" {
		f.Title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	f.Summary = extractSummary(lines, contentStart)
	return f, nil
}

func trimField(line string) string {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return ""
	}
	return strings.Trim(strings.TrimSpace(parts[1]), "\"'")
}

func parseListField(raw string) []string {
	raw = strings.TrimSpace(strings.Trim(raw, "[]"))
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.Trim(strings.TrimSpace(part), "\"'")
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
