package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ContextDeclaration represents a skill's context window policy.
type ContextDeclaration struct {
	Window     string         `yaml:"window"`      // isolated, fork, inherit
	Sections   *SectionFilter `yaml:"sections"`
	Intent     *IntentConfig  `yaml:"intent"`
	IntelScope string         `yaml:"intel_scope"`
}

// SectionFilter controls which knowledge sections to include or exclude.
type SectionFilter struct {
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
}

// IntentConfig declares the skill's intent mode.
type IntentConfig struct {
	Mode string `yaml:"mode"` // questions, task, none
}

// parseContextDeclaration reads a skill's SKILL.md frontmatter and parses the context field.
// Returns nil, nil if no context field exists (meaning: use defaults).
func parseContextDeclaration(skillName string) (*ContextDeclaration, error) {
	skillPath, err := resolveSkillPath(skillName)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(skillPath)
	if err != nil {
		return nil, fmt.Errorf("read SKILL.md: %w", err)
	}

	frontmatter, err := extractFrontmatter(string(data))
	if err != nil {
		return nil, err
	}
	if frontmatter == "" {
		return nil, nil
	}

	return parseContextFromFrontmatter([]byte(frontmatter))
}

// extractFrontmatter pulls YAML content between --- markers from a markdown file.
func extractFrontmatter(content string) (string, error) {
	const marker = "---"
	if !strings.HasPrefix(strings.TrimSpace(content), marker) {
		return "", nil
	}

	trimmed := strings.TrimSpace(content)
	// Find the opening marker
	startIdx := strings.Index(trimmed, marker)
	if startIdx < 0 {
		return "", nil
	}
	afterStart := startIdx + len(marker)

	// Find the closing marker
	endIdx := strings.Index(trimmed[afterStart:], marker)
	if endIdx < 0 {
		return "", nil
	}

	return trimmed[afterStart : afterStart+endIdx], nil
}

// parseContextFromFrontmatter parses the context field from YAML frontmatter bytes.
// Handles both string form (context: fork) and object form (full struct).
func parseContextFromFrontmatter(frontmatter []byte) (*ContextDeclaration, error) {
	// Parse into a generic node tree so we can detect whether context is scalar vs mapping
	var doc yaml.Node
	if err := yaml.Unmarshal(frontmatter, &doc); err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}

	// doc is a Document node; its first child is the mapping
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return nil, nil
	}
	mapping := doc.Content[0]
	if mapping.Kind != yaml.MappingNode {
		return nil, nil
	}

	// Walk the mapping key/value pairs to find "context"
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		keyNode := mapping.Content[i]
		valNode := mapping.Content[i+1]

		if keyNode.Value != "context" {
			continue
		}

		// String form: context: fork
		if valNode.Kind == yaml.ScalarNode {
			return &ContextDeclaration{Window: valNode.Value}, nil
		}

		// Object form: context: { window: ..., sections: ..., ... }
		if valNode.Kind == yaml.MappingNode {
			var decl ContextDeclaration
			if err := valNode.Decode(&decl); err != nil {
				return nil, fmt.Errorf("decode context declaration: %w", err)
			}
			return &decl, nil
		}

		return nil, fmt.Errorf("unexpected context node kind: %d", valNode.Kind)
	}

	// No context field found
	return nil, nil
}

// resolveSkillPath finds the SKILL.md for a given skill name.
// Search order: local repo, installed skills, plugin cache.
func resolveSkillPath(skillName string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	// 1. Local repo: skills/<name>/SKILL.md
	local := filepath.Join(cwd, "skills", skillName, "SKILL.md")
	if _, err := os.Stat(local); err == nil {
		return local, nil
	}

	// 2. Installed: ~/.claude/skills/<name>/SKILL.md
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}
	installed := filepath.Join(home, ".claude", "skills", skillName, "SKILL.md")
	if _, err := os.Stat(installed); err == nil {
		return installed, nil
	}

	// 3. Plugin cache: ~/.claude/plugins/cache/agentops-marketplace/agentops/*/skills/<name>/SKILL.md
	cachePattern := filepath.Join(home, ".claude", "plugins", "cache", "agentops-marketplace", "agentops", "*", "skills", skillName, "SKILL.md")
	matches, _ := filepath.Glob(cachePattern)
	if len(matches) > 0 {
		return matches[0], nil
	}

	return "", fmt.Errorf("skill %q not found", skillName)
}

// applyContextFilter filters knowledge based on the context declaration.
// Mutates the knowledge struct in place and returns it.
func applyContextFilter(knowledge *injectedKnowledge, decl *ContextDeclaration) *injectedKnowledge {
	if decl == nil {
		return knowledge
	}

	// Apply section excludes
	if decl.Sections != nil {
		for _, section := range decl.Sections.Exclude {
			switch section {
			case "HISTORY":
				knowledge.Sessions = nil
			case "INTEL":
				knowledge.Learnings = nil
				knowledge.Patterns = nil
			case "TASK":
				knowledge.BeadID = ""
				knowledge.Predecessor = nil
			}
		}
	}

	// Apply intel_scope
	if decl.IntelScope == "none" {
		knowledge.Learnings = nil
		knowledge.Patterns = nil
	}

	// intent.mode is declaration-only in v1 — do NOT filter on it.
	// Skills declare their intent mode so upstream orchestrators (e.g. hooks)
	// can adapt behavior, but the inject pipeline does not act on it yet.

	return knowledge
}
