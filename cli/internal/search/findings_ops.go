package search

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SectionFindings is the directory name under .agents/ where findings live.
const SectionFindings = "findings"

// FindingStats summarizes a collection of findings.
type FindingStats struct {
	Total           int                `json:"total"`
	ByStatus        map[string]int     `json:"by_status"`
	BySeverity      map[string]int     `json:"by_severity"`
	ByDetectability map[string]int     `json:"by_detectability"`
	TotalHits       int                `json:"total_hits"`
	MostCited       []KnowledgeFinding `json:"most_cited,omitempty"`
}

// RepoFindingsDir returns the path to the local findings directory.
func RepoFindingsDir(cwd string) string {
	return filepath.Join(cwd, ".agents", SectionFindings)
}

// ResolveManagedFindingsDir interprets a user-supplied path as either a repo
// root or a findings directory and returns the canonical findings dir.
func ResolveManagedFindingsDir(path string) string {
	clean := filepath.Clean(path)
	if filepath.Base(clean) == SectionFindings {
		return clean
	}
	return filepath.Join(clean, ".agents", SectionFindings)
}

// ResolveExistingFindingsDir returns an existing findings directory for path.
func ResolveExistingFindingsDir(path string) (string, error) {
	candidate := ResolveManagedFindingsDir(path)
	if info, err := os.Stat(candidate); err == nil && info.IsDir() {
		return candidate, nil
	}
	if info, err := os.Stat(filepath.Clean(path)); err == nil && info.IsDir() && filepath.Base(filepath.Clean(path)) == SectionFindings {
		return filepath.Clean(path), nil
	}
	return "", fmt.Errorf("no findings directory found at %s", path)
}

// MatchesIDFunc matches a finding by ID against a search ID.
type MatchesIDFunc func(findingID, file, query string) bool

// SelectFindingFiles returns either every .md file in dir or those matching ids.
func SelectFindingFiles(dir string, ids []string, all bool, matches MatchesIDFunc) ([]string, error) {
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
			finding, err := ParseFindingFile(file)
			if err != nil {
				continue
			}
			if matches(finding.ID, file, id) {
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

// FindLocalFindingByID locates a finding in cwd by ID.
func FindLocalFindingByID(cwd, id string, matches MatchesIDFunc) (KnowledgeFinding, error) {
	dir := RepoFindingsDir(cwd)
	files, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return KnowledgeFinding{}, err
	}
	for _, file := range files {
		finding, err := ParseFindingFile(file)
		if err != nil {
			continue
		}
		if matches(finding.ID, file, id) {
			return finding, nil
		}
	}
	return KnowledgeFinding{}, fmt.Errorf("finding %q not found", id)
}

// CopyFindingFile copies a finding file from src to dst.
func CopyFindingFile(src, dst string, force bool) error {
	if !force {
		if _, err := os.Stat(dst); err == nil {
			return fmt.Errorf("destination already exists: %s", dst)
		}
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read %s: %w", src, err)
	}
	return WriteFindingFileAtomic(dst, data, 0o644)
}

// BuildFindingStats aggregates statistics for a slice of findings.
func BuildFindingStats(findings []KnowledgeFinding) FindingStats {
	stats := FindingStats{
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
		stats.ByStatus[NormalizeStatKey(finding.Status, "unknown")]++
		stats.BySeverity[NormalizeStatKey(finding.Severity, "unknown")]++
		stats.ByDetectability[NormalizeStatKey(finding.Detectability, "unknown")]++
		stats.TotalHits += finding.HitCount
	}
	if len(findings) > 5 {
		stats.MostCited = findings[:5]
	} else {
		stats.MostCited = findings
	}
	return stats
}

// NormalizeStatKey trims a value or returns the fallback if empty.
func NormalizeStatKey(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

// UpdateFindingFrontMatter rewrites the YAML front matter of path with updates.
func UpdateFindingFrontMatter(path string, updates map[string]string) error {
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
	return WriteFindingFileAtomic(path, []byte(out), 0o644)
}

// WriteFindingFileAtomic writes data to path via a temp file and rename.
func WriteFindingFileAtomic(path string, data []byte, mode os.FileMode) error {
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
