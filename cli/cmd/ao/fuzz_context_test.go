package main

import (
	"testing"
)

// FuzzParseContextFromFrontmatter fuzzes the context declaration parser
// that handles both string form (context: fork) and object form.
func FuzzParseContextFromFrontmatter(f *testing.F) {
	// Seed corpus: string form
	f.Add([]byte("context: fork\n"))
	f.Add([]byte("context: isolated\n"))
	f.Add([]byte("context: inherit\n"))

	// Seed corpus: object form
	f.Add([]byte("context:\n  window: fork\n"))
	f.Add([]byte("context:\n  window: inherit\n  intel_scope: full\n"))
	f.Add([]byte("context:\n  window: isolated\n  sections:\n    include:\n      - learnings\n"))
	f.Add([]byte("context:\n  window: fork\n  sections:\n    exclude:\n      - sessions\n  intent:\n    mode: questions\n"))

	// Seed corpus: edge cases
	f.Add([]byte(""))
	f.Add([]byte("no_context_field: true\n"))
	f.Add([]byte("context: invalid_window\n"))
	f.Add([]byte("context:\n"))
	f.Add([]byte("not yaml at all: [[["))
	f.Add([]byte("context: 123\n"))
	f.Add([]byte("context:\n  window: fork\n  intel_scope: invalid\n"))
	f.Add([]byte("---\ncontext: fork\n---\n"))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Must never panic — errors are acceptable
		_, _ = parseContextFromFrontmatter(data)
	})
}

// FuzzExtractFrontmatter fuzzes the frontmatter extraction from markdown content.
func FuzzExtractFrontmatter(f *testing.F) {
	// Seed corpus with realistic markdown frontmatter
	f.Add("---\ncontext: fork\n---\n# Skill Title\n")
	f.Add("---\ntitle: test\ncontext:\n  window: inherit\n---\n")
	f.Add("")
	f.Add("no frontmatter here")
	f.Add("---\n---\n")
	f.Add("---\nunclosed frontmatter\n")
	f.Add("---\nkey: value\n---\n---\nmore: stuff\n---\n")
	f.Add("   ---\nkey: value\n---\n")

	f.Fuzz(func(t *testing.T, data string) {
		// Must never panic — errors are acceptable
		_, _ = extractFrontmatter(data)
	})
}
