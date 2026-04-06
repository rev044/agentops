package rpi

import (
	"regexp"
	"strings"
)

// ComplexityLevel classifies the ceremony complexity of an RPI goal.
// It determines how many gates and council validations are required.
type ComplexityLevel string

const (
	// ComplexityFast is for trivial tasks: short goals, no complex keywords.
	// Fast mode skips the validation phase (phase 3) and goes directly to completion.
	ComplexityFast ComplexityLevel = "fast"

	// ComplexityStandard is for normal tasks: most day-to-day work.
	// Standard mode uses the full 3-phase lifecycle with default gate settings.
	ComplexityStandard ComplexityLevel = "standard"

	// ComplexityFull is for complex work: refactors, migrations, rewrites.
	// Full mode uses the 3-phase lifecycle with additional rigor (no --quick shortcuts).
	ComplexityFull ComplexityLevel = "full"
)

// ComplexityScore holds intermediate scoring data used to classify a goal.
type ComplexityScore struct {
	// DescLen is the character length of the goal description (after trimming).
	DescLen int
	// ScopeKeywords is the count of multi-file scope keywords found.
	ScopeKeywords int
	// ComplexKeywords is the count of high-complexity operation keywords found.
	ComplexKeywords int
	// SimpleKeywords is the count of low-complexity operation keywords found.
	SimpleKeywords int
}

// ComplexScopeKeywords are words that suggest the goal spans multiple files or systems.
// These are matched as whole words to avoid false positives (e.g. "globally" vs "global").
var ComplexScopeKeywords = []string{
	"all", "entire", "across", "everywhere",
	"every file", "every module",
	"system-wide", "systemwide", "global", "throughout", "codebase",
}

// ComplexOperationKeywords are verbs or nouns that indicate significant engineering work.
// These are matched as whole words to prevent substring false positives
// (e.g. "support" matching "port", "restructure" matching "structure").
var ComplexOperationKeywords = []string{
	"refactor", "migrate", "migration", "rewrite", "redesign", "rearchitect",
	"overhaul", "restructure", "reorganize", "decouple", "deprecate",
	"split", "extract module", "port",
}

// SimpleOperationKeywords are verbs that indicate small, focused changes.
var SimpleOperationKeywords = []string{
	"fix", "add", "update", "change", "rename", "tweak", "bump", "patch",
	"correct", "typo", "adjust", "enable", "disable", "toggle", "remove",
	"delete", "cleanup", "clean up",
}

// ContainsWholeWord returns true if text contains kw as a whole word (word-boundary match).
// It handles both single-word and multi-word keyword phrases.
// For multi-word phrases, we check that the entire phrase appears surrounded by non-word chars.
func ContainsWholeWord(text, kw string) bool {
	// Build a word-boundary regex for the keyword.
	// `\b` handles transitions between word and non-word characters.
	pattern := `(?i)\b` + regexp.QuoteMeta(kw) + `\b`
	re, err := regexp.Compile(pattern)
	if err != nil {
		// Fall back to simple substring match if pattern compilation fails.
		return strings.Contains(text, kw)
	}
	return re.MatchString(text)
}

// ClassifyComplexity analyzes a goal description and returns the appropriate ComplexityLevel.
//
// Scoring algorithm:
//   - Short goal (<=30 chars) with no complex/scope keywords -> fast
//   - Goal >120 chars, or with complex-operation or 2+ scope keywords -> full
//   - Everything else -> standard
//
// Fast path: skips council validation phase (phase 3). Used for small, well-understood tasks.
// Standard path: full 3-phase lifecycle, gates use --quick by default.
// Full path: full 3-phase lifecycle, gates use full council (no shortcuts).
func ClassifyComplexity(goal string) ComplexityLevel {
	score := ScoreGoal(goal)
	return LevelFromScore(score)
}

// ScoreGoal computes a ComplexityScore from the goal string using whole-word matching.
func ScoreGoal(goal string) ComplexityScore {
	lower := strings.ToLower(strings.TrimSpace(goal))
	s := ComplexityScore{
		DescLen: len(lower),
	}

	for _, kw := range ComplexScopeKeywords {
		if ContainsWholeWord(lower, kw) {
			s.ScopeKeywords++
		}
	}

	for _, kw := range ComplexOperationKeywords {
		if ContainsWholeWord(lower, kw) {
			s.ComplexKeywords++
		}
	}

	for _, kw := range SimpleOperationKeywords {
		if ContainsWholeWord(lower, kw) {
			s.SimpleKeywords++
		}
	}

	return s
}

// LevelFromScore converts a ComplexityScore into a ComplexityLevel.
//
// Thresholds (tuned for practical RPI usage):
//   - fast:     <=30 chars AND no complex/scope keywords
//   - full:     complex-operation keyword OR 2+ scope keywords OR >120 chars
//   - standard: everything else (31-120 chars, or 1 scope keyword, or ambiguous)
func LevelFromScore(s ComplexityScore) ComplexityLevel {
	// Explicit full-complexity signals always win.
	if s.ComplexKeywords > 0 || s.ScopeKeywords > 1 {
		return ComplexityFull
	}

	// Long descriptions suggest broader scope even without keywords.
	if s.DescLen > 120 {
		return ComplexityFull
	}

	// Borderline: moderate length or some scope signal -> standard.
	if s.DescLen > 30 || s.ScopeKeywords > 0 {
		return ComplexityStandard
	}

	// Trivial tasks: short goal with no complex or scope keywords.
	if s.ComplexKeywords == 0 && s.ScopeKeywords == 0 {
		return ComplexityFast
	}

	return ComplexityStandard
}
