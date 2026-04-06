package lifecycle

import (
	"path/filepath"
	"strings"
)

// IsContainedPath checks if path is contained within baseDir.
func IsContainedPath(baseDir, path string) bool {
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return false
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	cleanBase := filepath.Clean(absBase)
	cleanPath := filepath.Clean(absPath)
	if !strings.HasSuffix(cleanBase, string(filepath.Separator)) {
		cleanBase += string(filepath.Separator)
	}
	return strings.HasPrefix(cleanPath+string(filepath.Separator), cleanBase) || cleanPath == filepath.Clean(absBase)
}

// IsArtifactFile checks if a filename is a valid artifact file type.
func IsArtifactFile(name string) bool {
	return strings.HasSuffix(name, ".md") || strings.HasSuffix(name, ".jsonl")
}

// ParseMarkdownField extracts a value for a field from a markdown line.
func ParseMarkdownField(line, field string) (string, bool) {
	prefixes := []string{
		"**" + field + "**:",
		"**" + field + ":**",
		"- **" + field + "**:",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix)), true
		}
	}
	return "", false
}
