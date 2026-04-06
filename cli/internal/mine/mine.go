// Package mine provides pure helpers for knowledge mining operations.
package mine

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ValidSources enumerates the allowed source names for ao mine.
var ValidSources = map[string]bool{
	"git":    true,
	"agents": true,
	"code":   true,
	"events": true,
}

// ParseWindow parses a duration string with support for "h", "m", and "d" suffixes.
func ParseWindow(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration string")
	}
	suffix := s[len(s)-1:]
	numStr := s[:len(s)-1]
	switch suffix {
	case "d":
		days, err := strconv.Atoi(numStr)
		if err != nil {
			return 0, fmt.Errorf("invalid day count %q: %w", numStr, err)
		}
		if days <= 0 {
			return 0, fmt.Errorf("duration must be positive, got %q", s)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	case "h":
		hours, err := strconv.Atoi(numStr)
		if err != nil {
			return 0, fmt.Errorf("invalid hour count %q: %w", numStr, err)
		}
		if hours <= 0 {
			return 0, fmt.Errorf("duration must be positive, got %q", s)
		}
		return time.Duration(hours) * time.Hour, nil
	case "m":
		mins, err := strconv.Atoi(numStr)
		if err != nil {
			return 0, fmt.Errorf("invalid minute count %q: %w", numStr, err)
		}
		if mins <= 0 {
			return 0, fmt.Errorf("duration must be positive, got %q", s)
		}
		return time.Duration(mins) * time.Minute, nil
	default:
		return 0, fmt.Errorf("unsupported duration suffix %q (use h, d, or m)", suffix)
	}
}

// SplitSources splits and validates a comma-separated source list.
func SplitSources(s string) ([]string, error) {
	parts := strings.Split(s, ",")
	var sources []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !ValidSources[p] {
			return nil, fmt.Errorf("unknown source %q (valid: git, agents, code, events)", p)
		}
		sources = append(sources, p)
	}
	if len(sources) == 0 {
		return nil, fmt.Errorf("no valid sources specified")
	}
	return sources, nil
}

// ReadDirContent reads all .md file contents from a directory.
func ReadDirContent(dir string) (map[string]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	contents := make(map[string]string)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		contents[e.Name()] = string(data)
	}
	return contents, nil
}
