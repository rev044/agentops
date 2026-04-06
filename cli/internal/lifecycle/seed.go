package lifecycle

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/boshu2/agentops/cli/internal/goals"
)

// ValidTemplates enumerates the allowed seed template names.
var ValidTemplates = map[string]bool{
	"go-cli":     true,
	"python-lib": true,
	"web-app":    true,
	"rust-cli":   true,
	"generic":    true,
}

// ValidateTemplateMapEntries verifies that every enabled template name has a
// corresponding embedded YAML file in the provided filesystem.
func ValidateTemplateMapEntries(templates map[string]bool, templatesFS fs.FS) error {
	names := make([]string, 0, len(templates))
	for name, enabled := range templates {
		if enabled {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	for _, name := range names {
		path := filepath.Join("templates", name+".yaml")
		if _, err := fs.Stat(templatesFS, path); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("template %q missing embedded file %s", name, path)
			}
			return fmt.Errorf("validate embedded template %q (%s): %w", name, path, err)
		}
	}

	return nil
}

// TemplateConfig holds the per-template data used by BuildSeedGoalFile.
type TemplateConfig struct {
	MissionSuffix string
	NorthStars    []string
	AntiStars     []string
	Directives    []goals.Directive
}

// TemplateConfigs maps template names to their goal-file configuration.
// The "generic" entry is the fallback for unknown template names.
var TemplateConfigs = map[string]TemplateConfig{
	"go-cli": {
		MissionSuffix: " (Go CLI)",
		NorthStars:    []string{"All checks pass on every commit", "Clean go vet and golangci-lint"},
		AntiStars:     []string{"Untested changes reaching main", "Cyclomatic complexity > 25"},
		Directives: []goals.Directive{
			{Number: 1, Title: "Establish baseline", Description: "Get all gates passing and maintain a green baseline.", Steer: "increase"},
			{Number: 2, Title: "Test coverage", Description: "Maintain and increase test coverage across all packages.", Steer: "increase"},
		},
	},
	"python-lib": {
		MissionSuffix: " (Python library)",
		NorthStars:    []string{"All tests pass on every commit", "Type hints on all public APIs"},
		AntiStars:     []string{"Untested changes reaching main", "Undocumented public functions"},
		Directives: []goals.Directive{
			{Number: 1, Title: "Establish baseline", Description: "Get all gates passing and maintain a green baseline.", Steer: "increase"},
			{Number: 2, Title: "Documentation", Description: "All public APIs are documented with docstrings.", Steer: "increase"},
		},
	},
	"web-app": {
		MissionSuffix: " (web application)",
		NorthStars:    []string{"All checks pass on every commit", "No critical accessibility violations"},
		AntiStars:     []string{"Untested changes reaching main", "Unhandled runtime errors in production"},
		Directives: []goals.Directive{
			{Number: 1, Title: "Establish baseline", Description: "Get all gates passing and maintain a green baseline.", Steer: "increase"},
			{Number: 2, Title: "Test coverage", Description: "Component and integration tests for critical paths.", Steer: "increase"},
		},
	},
	"rust-cli": {
		MissionSuffix: " (Rust CLI)",
		NorthStars:    []string{"All checks pass on every commit", "Clean clippy with no warnings"},
		AntiStars:     []string{"Untested changes reaching main", "Unsafe code without justification"},
		Directives: []goals.Directive{
			{Number: 1, Title: "Establish baseline", Description: "Get all gates passing and maintain a green baseline.", Steer: "increase"},
			{Number: 2, Title: "Test coverage", Description: "Maintain and increase test coverage.", Steer: "increase"},
		},
	},
	"generic": {
		MissionSuffix: "",
		NorthStars:    []string{"All checks pass on every commit"},
		AntiStars:     []string{"Untested changes reaching main"},
		Directives: []goals.Directive{
			{Number: 1, Title: "Establish baseline", Description: "Get all gates passing and maintain a green baseline.", Steer: "increase"},
		},
	},
}

// BuildSeedGoalFile creates a GoalFile tailored to the template.
func BuildSeedGoalFile(root string, template string) *goals.GoalFile {
	cfg, ok := TemplateConfigs[template]
	if !ok {
		cfg = TemplateConfigs["generic"]
	}

	return &goals.GoalFile{
		Version:    4,
		Format:     "md",
		Mission:    fmt.Sprintf("Fitness goals for %s%s", filepath.Base(root), cfg.MissionSuffix),
		NorthStars: cfg.NorthStars,
		AntiStars:  cfg.AntiStars,
		Directives: cfg.Directives,
	}
}

// ClaudeMDSeedSection is the section appended to CLAUDE.md by ao seed.
const ClaudeMDSeedSection = `
## AgentOps Knowledge Flywheel

Knowledge compounds automatically across sessions:

- **MEMORY.md** is auto-loaded by your AI coding tool every session
- **Session hooks** extract learnings, update MEMORY.md, and prune stale knowledge
- **Skills** invoke flywheel commands at the right moments (no manual ao commands needed)

Verify the flywheel any time:

` + "```bash" + `
ao flywheel status    # escape velocity check
ao status             # current knowledge inventory
` + "```" + `
`

// ClaudeMDSeedMarker is used to detect if the seed section was already added.
const ClaudeMDSeedMarker = "## AgentOps Knowledge Flywheel"

// ClaudeMDSeedMarkerLegacy is the legacy marker for backward compatibility.
const ClaudeMDSeedMarkerLegacy = "## AgentOps Session Protocol"

// HasSeedMarker returns true if content contains the current or legacy seed marker.
func HasSeedMarker(content string) bool {
	return strings.Contains(content, ClaudeMDSeedMarker) || strings.Contains(content, ClaudeMDSeedMarkerLegacy)
}

// FindSeedMarker returns the marker string found in content (current or
// legacy), or empty string if neither is present.
func FindSeedMarker(content string) string {
	if strings.Contains(content, ClaudeMDSeedMarker) {
		return ClaudeMDSeedMarker
	}
	if strings.Contains(content, ClaudeMDSeedMarkerLegacy) {
		return ClaudeMDSeedMarkerLegacy
	}
	return ""
}
