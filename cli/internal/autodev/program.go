// Package autodev manages PROGRAM.md or AUTODEV.md operational contracts for
// bounded autonomous development loops.
package autodev

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Program describes the operational contract for autonomous development.
type Program struct {
	Format             string   `json:"format"`
	Objective          string   `json:"objective"`
	MutableScope       []string `json:"mutable_scope"`
	ImmutableScope     []string `json:"immutable_scope"`
	ExperimentUnit     string   `json:"experiment_unit"`
	ValidationCommands []string `json:"validation_commands"`
	DecisionPolicy     []string `json:"decision_policy"`
	EscalationRules    string   `json:"escalation_rules"`
	StopConditions     []string `json:"stop_conditions"`
}

// ValidationError describes a specific program validation problem.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ResolveProgramPath auto-detects the canonical program file in cwd.
// PROGRAM.md takes precedence over AUTODEV.md.
func ResolveProgramPath(cwd string) string {
	for _, candidate := range []string{"PROGRAM.md", "AUTODEV.md"} {
		full := filepath.Join(cwd, candidate)
		if info, err := os.Stat(full); err == nil && !info.IsDir() {
			return candidate
		}
	}
	return ""
}

// LoadProgram reads and parses a program file.
func LoadProgram(path string) (*Program, string, error) {
	actual := strings.TrimSpace(path)
	if actual == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, "", err
		}
		rel := ResolveProgramPath(cwd)
		if rel == "" {
			return nil, "", os.ErrNotExist
		}
		actual = filepath.Join(cwd, rel)
	}

	data, err := os.ReadFile(actual)
	if err != nil {
		return nil, "", err
	}
	prog, err := ParseMarkdownProgram(data)
	if err != nil {
		return nil, "", fmt.Errorf("parsing %s: %w", actual, err)
	}
	return prog, actual, nil
}

// ParseMarkdownProgram parses PROGRAM.md or AUTODEV.md.
func ParseMarkdownProgram(data []byte) (*Program, error) {
	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("empty program file")
	}

	lines := strings.Split(content, "\n")
	return &Program{
		Format:             "md",
		Objective:          parseTextSection(lines, "Objective"),
		MutableScope:       parseListSection(lines, "Mutable Scope"),
		ImmutableScope:     parseListSection(lines, "Immutable Scope"),
		ExperimentUnit:     parseTextSection(lines, "Experiment Unit"),
		ValidationCommands: parseListSection(lines, "Validation Commands"),
		DecisionPolicy:     parseListSection(lines, "Decision Policy"),
		EscalationRules:    parseTextSection(lines, "Escalation Rules"),
		StopConditions:     parseListSection(lines, "Stop Conditions"),
	}, nil
}

// ValidateProgram returns structural validation errors for a Program.
func ValidateProgram(prog *Program) []error {
	var errs []error
	if prog == nil {
		return []error{ValidationError{Field: "program", Message: "required"}}
	}
	requireText := func(field, value string) {
		if strings.TrimSpace(value) == "" {
			errs = append(errs, ValidationError{Field: field, Message: "required"})
		}
	}
	requireList := func(field string, values []string) {
		if len(values) == 0 {
			errs = append(errs, ValidationError{Field: field, Message: "requires at least one item"})
		}
	}

	requireText("objective", prog.Objective)
	requireList("mutable scope", prog.MutableScope)
	requireList("immutable scope", prog.ImmutableScope)
	requireText("experiment unit", prog.ExperimentUnit)
	requireList("validation commands", prog.ValidationCommands)
	requireList("decision policy", prog.DecisionPolicy)
	requireText("escalation rules", prog.EscalationRules)
	requireList("stop conditions", prog.StopConditions)

	return errs
}

func parseTextSection(lines []string, heading string) string {
	var body []string
	inSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			title := strings.TrimSpace(strings.TrimPrefix(trimmed, "## "))
			if strings.EqualFold(title, heading) {
				inSection = true
				continue
			}
			if inSection {
				break
			}
			continue
		}
		if inSection {
			if strings.HasPrefix(trimmed, "#") {
				break
			}
			body = append(body, line)
		}
	}

	return strings.TrimSpace(strings.Join(body, "\n"))
}

func parseListSection(lines []string, heading string) []string {
	raw := parseTextSection(lines, heading)
	if raw == "" {
		return nil
	}

	var items []string
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "- "):
			items = append(items, trimListItem(trimmed[2:]))
		case strings.HasPrefix(trimmed, "* "):
			items = append(items, trimListItem(trimmed[2:]))
		}
	}
	return items
}

func trimListItem(value string) string {
	item := strings.TrimSpace(value)
	if len(item) >= 2 && item[0] == '`' && item[len(item)-1] == '`' {
		item = item[1 : len(item)-1]
	}
	return strings.TrimSpace(item)
}
