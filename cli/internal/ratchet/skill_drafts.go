package ratchet

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const skillDraftSessionThreshold = 3

var datedPatternPrefix = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-`)

type SkillDraftResult struct {
	Evaluated int
	Generated int
	Paths     []string
}

type skillDraftEvidence struct {
	PatternPath       string    `json:"pattern_path"`
	SessionRefs       int       `json:"session_refs"`
	SuggestedTier     string    `json:"suggested_tier"`
	GeneratedAt       time.Time `json:"generated_at"`
	DraftPath         string    `json:"draft_path"`
	EvidenceThreshold int       `json:"evidence_threshold"`
}

func GenerateSkillDrafts(baseDir string) (SkillDraftResult, error) {
	result := SkillDraftResult{}

	patternPaths, err := filepath.Glob(filepath.Join(baseDir, ".agents", "patterns", "*.md"))
	if err != nil {
		return result, fmt.Errorf("glob patterns: %w", err)
	}
	sort.Strings(patternPaths)

	if len(patternPaths) == 0 {
		return result, nil
	}

	validator, err := NewValidator(baseDir)
	if err != nil {
		return result, fmt.Errorf("create validator: %w", err)
	}

	for _, patternPath := range patternPaths {
		result.Evaluated++

		sessionRefs := validator.countSessionRefs(patternPath)
		if sessionRefs < skillDraftSessionThreshold {
			continue
		}

		draftPath, err := writeSkillDraft(baseDir, patternPath, sessionRefs)
		if err != nil {
			return result, err
		}
		result.Generated++
		result.Paths = append(result.Paths, draftPath)
	}

	return result, nil
}

func writeSkillDraft(baseDir, patternPath string, sessionRefs int) (string, error) {
	slug := draftSlug(patternPath)
	draftDir := filepath.Join(baseDir, ".agents", "skill-drafts", slug)
	if err := os.MkdirAll(draftDir, 0o755); err != nil {
		return "", fmt.Errorf("create draft dir %s: %w", draftDir, err)
	}

	suggestedTier := suggestDraftTier(slug)
	skillPath := filepath.Join(draftDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(renderSkillDraft(patternPath, slug, suggestedTier)), 0o644); err != nil {
		return "", fmt.Errorf("write skill draft %s: %w", skillPath, err)
	}

	evidencePath := filepath.Join(draftDir, "evidence.json")
	evidence := skillDraftEvidence{
		PatternPath:       patternPath,
		SessionRefs:       sessionRefs,
		SuggestedTier:     suggestedTier,
		GeneratedAt:       time.Now().UTC(),
		DraftPath:         skillPath,
		EvidenceThreshold: skillDraftSessionThreshold,
	}
	data, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal skill draft evidence: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(evidencePath, data, 0o644); err != nil {
		return "", fmt.Errorf("write skill draft evidence %s: %w", evidencePath, err)
	}

	return skillPath, nil
}

func draftSlug(patternPath string) string {
	base := strings.TrimSuffix(filepath.Base(patternPath), filepath.Ext(patternPath))
	base = datedPatternPrefix.ReplaceAllString(base, "")
	base = strings.TrimSuffix(base, "-pattern")
	base = strings.ReplaceAll(base, "_", "-")
	base = strings.ToLower(base)
	base = strings.Trim(base, "-")
	if base == "" {
		return "generated-skill-draft"
	}
	return base
}

func suggestDraftTier(slug string) string {
	switch {
	case strings.Contains(slug, "research"), strings.Contains(slug, "knowledge"), strings.Contains(slug, "trace"):
		return "knowledge"
	case strings.Contains(slug, "status"), strings.Contains(slug, "handoff"), strings.Contains(slug, "recover"):
		return "session"
	case strings.Contains(slug, "release"), strings.Contains(slug, "readme"), strings.Contains(slug, "product"):
		return "product"
	default:
		return "execution"
	}
}

func renderSkillDraft(patternPath, slug, suggestedTier string) string {
	return fmt.Sprintf(`---
name: %s
description: 'Draft generated from recurring pattern evidence. Triggers: "%s", "refine %s".'
skill_api_version: 1
context:
  window: fork
  intent:
    mode: task
  sections:
    exclude: [HISTORY]
  intel_scope: topic
metadata:
  tier: %s
---

# %s

## Purpose

Draft generated from recurring pattern evidence in %q.

## When to Use

- When this repeated operator behavior deserves a dedicated skill boundary
- When the same pattern keeps showing up across sessions and should be refined into a reusable surface

## Inputs

- The user intent this pattern should handle
- Repo context relevant to the recurring pattern
- The source pattern and supporting session evidence

## Instructions

1. Review the source pattern at %q.
2. Turn the repeated behavior into a deterministic operator flow with a clear start and finish.
3. Add references, scripts, and validations before promoting this draft into skills/.

## Output

- A reviewable skill draft ready for human refinement and promotion

## Examples

~~~text
Refine this generated draft into a production-ready skill and validate it against the source pattern.
~~~

## Troubleshooting

- Symptom: The pattern is still too broad.
  Fix: Split it into a narrower skill boundary before promotion.
- Symptom: The pattern is runtime-specific.
  Fix: Define the shared contract first, then add runtime-specific tailoring only where it is actually needed.
`, slug, slug, slug, suggestedTier, slug, patternPath, patternPath)
}
