// Package lifecycle contains pure helpers for learning maturity, expiry,
// curation, eviction, and seeding lifecycle stages. These functions are
// extracted from cmd/ao for testability and reuse.
package lifecycle

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/boshu2/agentops/cli/internal/types"
)

// DefaultLearningMetadataFields returns the default frontmatter values used
// when normalizing learning files that lack metadata.
func DefaultLearningMetadataFields() map[string]string {
	return map[string]string{
		"utility":       fmt.Sprintf("%.4f", types.InitialUtility),
		"maturity":      "provisional",
		"confidence":    "0.0000",
		"reward_count":  "0",
		"helpful_count": "0",
		"harmful_count": "0",
	}
}

// ParseFrontmatterFields extracts specific fields from YAML frontmatter in a
// markdown file. Returns a map of field name to value for the requested fields.
func ParseFrontmatterFields(path string, fields ...string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(f)
	inFrontmatter := false
	dashCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "---" {
			dashCount++
			if dashCount == 1 {
				inFrontmatter = true
				continue
			}
			if dashCount == 2 {
				break
			}
		}

		if inFrontmatter {
			for _, field := range fields {
				prefix := field + ":"
				if strings.HasPrefix(trimmed, prefix) {
					val := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
					val = strings.Trim(val, "\"'")
					result[field] = val
				}
			}
		}
	}

	return result, scanner.Err()
}

// IsLowSignalLearningBody returns true if a learning body is too short or
// otherwise lacks meaningful signal.
func IsLowSignalLearningBody(body string) bool {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return true
	}
	if len(trimmed) < 50 {
		return true
	}
	if len(strings.Fields(trimmed)) < 12 {
		return true
	}
	lower := strings.ToLower(trimmed)
	for _, prefix := range []string{
		"till ", "still ", "will ", "let me ",
		"and ", "but ", "or ", "however ", "therefore ",
		"additionally ", "furthermore ",
		"- ", "* ",
	} {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	sentenceEnders := 0
	runes := []rune(trimmed)
	for i, ch := range runes {
		if ch == '.' || ch == '!' || ch == '?' {
			if i == len(runes)-1 || runes[i+1] == ' ' || runes[i+1] == '\n' {
				sentenceEnders++
			}
		}
	}
	return sentenceEnders == 0
}

// StripLearningHeading removes the leading "# Title" from a learning body.
func StripLearningHeading(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) >= 3 && strings.HasPrefix(strings.TrimSpace(lines[0]), "# ") && strings.TrimSpace(lines[1]) == "" {
		return strings.TrimSpace(strings.Join(lines[2:], "\n"))
	}
	return strings.TrimSpace(content)
}

// IsEvictionEligible reports whether a learning's metrics make it a candidate
// for eviction.
func IsEvictionEligible(utility, confidence float64, maturity string) bool {
	if maturity == "established" {
		return false
	}
	if utility >= 0.3 {
		return false
	}
	return confidence < 0.3
}

// FloatValueFromData extracts a float64 from a generic map, returning the
// default if the key is missing or the wrong type.
func FloatValueFromData(data map[string]any, key string, defaultValue float64) float64 {
	value, ok := data[key].(float64)
	if !ok {
		return defaultValue
	}
	return value
}

// NonEmptyStringFromData extracts a non-empty string from a generic map,
// returning the default if missing or empty.
func NonEmptyStringFromData(data map[string]any, key, defaultValue string) string {
	value, ok := data[key].(string)
	if !ok || value == "" {
		return defaultValue
	}
	return value
}

// ReadLearningJSONLData reads the first JSON line from a .jsonl learning file.
func ReadLearningJSONLData(file string) (map[string]any, bool) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, false
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		return nil, false
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &data); err != nil {
		return nil, false
	}

	return data, true
}
