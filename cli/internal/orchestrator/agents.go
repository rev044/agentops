// Package orchestrator provides hierarchical agent dispatch and synthesis.
package orchestrator

// AgentMapping maps pod categories to Claude Code Task tool subagent_types.
// These are the agents registered in the claude-plugin marketplace.
var AgentMapping = map[string]AgentConfig{
	"security": {
		SubagentType: "agentops:security-expert",
		FallbackType: "agentops:security-reviewer",
		Description:  "Security vulnerability assessment and OWASP Top 10 analysis",
	},
	"quality": {
		SubagentType: "agentops:code-quality-expert",
		FallbackType: "agentops:code-reviewer",
		Description:  "Code quality, complexity analysis, and maintainability review",
	},
	"architecture": {
		SubagentType: "agentops:architecture-expert",
		FallbackType: "agentops:code-reviewer",
		Description:  "Architecture review, design patterns, and coupling analysis",
	},
	"ux": {
		SubagentType: "agentops:ux-expert",
		FallbackType: "agentops:code-reviewer",
		Description:  "Accessibility (WCAG), UX patterns, and user experience review",
	},
	"data": {
		SubagentType: "agentops:code-reviewer",
		FallbackType: "agentops:code-reviewer",
		Description:  "Data validation, integrity, and storage patterns",
	},
	"ops": {
		SubagentType: "agentops:code-reviewer",
		FallbackType: "agentops:code-reviewer",
		Description:  "Operations, logging, monitoring, and deployment review",
	},
}

// AgentConfig describes how to invoke an agent for a category.
type AgentConfig struct {
	// SubagentType is the Task tool subagent_type parameter.
	SubagentType string `json:"subagent_type"`

	// FallbackType is used if the primary agent is unavailable.
	FallbackType string `json:"fallback_type"`

	// Description explains the agent's focus area.
	Description string `json:"description"`
}

// AgentDispatch describes a single agent invocation for the skill to execute.
type AgentDispatch struct {
	// ID uniquely identifies this dispatch.
	ID string `json:"id"`

	// Category is the validation focus (security, quality, etc).
	Category string `json:"category"`

	// SubagentType for the Task tool.
	SubagentType string `json:"subagent_type"`

	// Prompt to send to the agent.
	Prompt string `json:"prompt"`

	// Files to analyze.
	Files []string `json:"files"`

	// OutputPath where the agent should write findings.
	OutputPath string `json:"output_path"`

	// Model preference (haiku for analysis, sonnet for synthesis).
	Model string `json:"model,omitempty"`
}

// DispatchPlan describes the full dispatch plan for a validation run.
type DispatchPlan struct {
	// PlanID uniquely identifies this plan.
	PlanID string `json:"plan_id"`

	// Wave1 dispatches for parallel execution.
	Wave1 []AgentDispatch `json:"wave1"`

	// FindingsDir where agents write findings.
	FindingsDir string `json:"findings_dir"`

	// Created timestamp.
	Created string `json:"created"`

	// Config used to generate this plan.
	Config DispatchConfig `json:"config"`
}

// GetAgentForCategory returns the agent config for a category.
func GetAgentForCategory(category string) AgentConfig {
	if config, ok := AgentMapping[category]; ok {
		return config
	}
	// Default to code-reviewer for unknown categories
	return AgentConfig{
		SubagentType: "agentops:code-reviewer",
		FallbackType: "agentops:code-reviewer",
		Description:  "General code review",
	}
}

// BuildPromptForCategory creates a focused prompt for an agent.
func BuildPromptForCategory(category string, files []string) string {
	var focus string
	switch category {
	case "security":
		focus = `Focus your analysis on security vulnerabilities:
- Authentication and authorization flaws
- Injection vulnerabilities (SQL, command, XSS)
- Cryptographic issues (weak algorithms, key management)
- Secrets exposure (API keys, credentials in code)
- OWASP Top 10 vulnerabilities
- Input validation and sanitization
- File permission issues (0644 should be 0600 for sensitive files)`

	case "architecture":
		focus = `Focus your analysis on architecture and design:
- Design patterns (appropriate use, anti-patterns)
- Coupling and cohesion (tight coupling, god classes)
- Scalability concerns (bottlenecks, resource leaks)
- Error handling patterns (consistent, informative)
- API design (REST conventions, versioning)
- Separation of concerns (layers, boundaries)`

	case "quality":
		focus = `Focus your analysis on code quality:
- Cyclomatic complexity (functions over 10, files over 50)
- Code duplication (copy-paste patterns)
- Test coverage gaps (untested branches, edge cases)
- Maintainability issues (dead code, unclear naming)
- Documentation quality (missing docstrings, outdated comments)
- Type safety and error handling`

	case "ux":
		focus = `Focus your analysis on user experience:
- Accessibility (WCAG compliance, screen reader support)
- Error messages (helpful, actionable, non-technical)
- Loading states and feedback (spinners, progress indicators)
- Responsive design (mobile, tablet, desktop)
- Performance (bundle size, render blocking)
- Emoji accessibility (prefer ASCII alternatives)`

	case "data":
		focus = `Focus your analysis on data handling:
- Data validation (input sanitization, type checking)
- Database migrations (rollback safety, data integrity)
- Data integrity constraints (foreign keys, unique constraints)
- Backup and recovery (disaster recovery, data loss prevention)
- Privacy (PII handling, GDPR compliance)
- Atomic operations (race conditions, concurrent access)`

	case "ops":
		focus = `Focus your analysis on operations:
- Logging (structured, appropriate levels, no secrets)
- Monitoring (metrics, health checks, alerting)
- Configuration management (env vars, secrets handling)
- Deployment (rollback safety, blue-green, canary)
- Error recovery (retry logic, circuit breakers)
- Resource management (connection pooling, cleanup)`

	default:
		focus = `Perform a general code review focusing on:
- Code quality and maintainability
- Security vulnerabilities
- Performance concerns
- Best practices`
	}

	return focus
}
