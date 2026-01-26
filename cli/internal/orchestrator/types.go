// Package orchestrator provides hierarchical agent dispatch and synthesis.
package orchestrator

import (
	"time"
)

// Severity levels for findings.
type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityHigh     Severity = "HIGH"
	SeverityMedium   Severity = "MEDIUM"
	SeverityLow      Severity = "LOW"
	SeverityPass     Severity = "PASS"
)

// Finding represents a single validation finding.
type Finding struct {
	// ID is a unique identifier for this finding.
	ID string `json:"id"`

	// Severity indicates the importance level.
	Severity Severity `json:"severity"`

	// Category groups related findings (security, architecture, etc).
	Category string `json:"category"`

	// Title is a brief description.
	Title string `json:"title"`

	// Description provides details.
	Description string `json:"description"`

	// Files affected by this finding.
	Files []string `json:"files,omitempty"`

	// Line numbers if applicable.
	Lines []int `json:"lines,omitempty"`

	// Recommendation for fixing.
	Recommendation string `json:"recommendation,omitempty"`

	// Source identifies which agent/pod found this.
	Source string `json:"source,omitempty"`

	// Confidence score (0-1).
	Confidence float64 `json:"confidence,omitempty"`

	// Timestamp when found.
	FoundAt time.Time `json:"found_at"`
}

// PodConfig configures a validation pod.
type PodConfig struct {
	// Name identifies this pod.
	Name string `json:"name"`

	// Category is the validation focus (security, quality, etc).
	Category string `json:"category"`

	// AgentCount is the number of agents in this pod.
	AgentCount int `json:"agent_count"`

	// Files to analyze.
	Files []string `json:"files"`

	// Prompt template for agents.
	Prompt string `json:"prompt,omitempty"`

	// ContextBudget is the max context percentage (0-1).
	ContextBudget float64 `json:"context_budget"`
}

// PodResult holds findings from a single pod.
type PodResult struct {
	// Pod configuration.
	Config PodConfig `json:"config"`

	// Findings from all agents in the pod.
	Findings []Finding `json:"findings"`

	// Summary is a brief overview.
	Summary string `json:"summary,omitempty"`

	// Duration of pod execution.
	Duration time.Duration `json:"duration"`

	// AgentResults from individual agents.
	AgentResults []AgentResult `json:"agent_results,omitempty"`

	// Error if pod failed.
	Error string `json:"error,omitempty"`

	// ContextUsage is the percentage of context consumed.
	ContextUsage float64 `json:"context_usage"`
}

// AgentResult holds findings from a single agent.
type AgentResult struct {
	// AgentID identifies this agent.
	AgentID string `json:"agent_id"`

	// Findings from this agent.
	Findings []Finding `json:"findings"`

	// Duration of agent execution.
	Duration time.Duration `json:"duration"`

	// ContextUsage percentage.
	ContextUsage float64 `json:"context_usage"`

	// Error if agent failed.
	Error string `json:"error,omitempty"`
}

// ClusterResult holds synthesized findings from multiple pods.
type ClusterResult struct {
	// ClusterID identifies this cluster.
	ClusterID string `json:"cluster_id"`

	// Pods included in this cluster.
	PodNames []string `json:"pod_names"`

	// MergedFindings after deduplication.
	MergedFindings []Finding `json:"merged_findings"`

	// HighestSeverity across all pods (single-veto applied).
	HighestSeverity Severity `json:"highest_severity"`

	// Summary of cluster findings.
	Summary string `json:"summary,omitempty"`

	// Duration of synthesis.
	Duration time.Duration `json:"duration"`
}

// FinalResult holds the final synthesized validation result.
type FinalResult struct {
	// Verdict is the overall result.
	Verdict Severity `json:"verdict"`

	// Findings after final synthesis.
	Findings []Finding `json:"findings"`

	// CriticalCount of CRITICAL findings.
	CriticalCount int `json:"critical_count"`

	// HighCount of HIGH findings.
	HighCount int `json:"high_count"`

	// Summary of the validation.
	Summary string `json:"summary"`

	// Recommendations for next steps.
	Recommendations []string `json:"recommendations,omitempty"`

	// Duration of entire validation.
	Duration time.Duration `json:"duration"`

	// PodResults from Wave 1.
	PodResults []PodResult `json:"pod_results,omitempty"`

	// ClusterResults from Wave 2.
	ClusterResults []ClusterResult `json:"cluster_results,omitempty"`

	// Timestamp of completion.
	CompletedAt time.Time `json:"completed_at"`
}

// DispatchConfig configures the wave dispatcher.
type DispatchConfig struct {
	// PodSize is the number of agents per pod (default: 6-8).
	PodSize int `json:"pod_size"`

	// MaxPods is the maximum number of pods (default: 8).
	MaxPods int `json:"max_pods"`

	// ContextBudget is the threshold before summarization (default: 0.6).
	ContextBudget float64 `json:"context_budget"`

	// QuorumThreshold for cross-pod agreement (default: 0.7).
	QuorumThreshold float64 `json:"quorum_threshold"`

	// StreamResults enables streaming partial results.
	StreamResults bool `json:"stream_results"`

	// EarlyTermination enables stopping on unanimous CRITICAL.
	EarlyTermination bool `json:"early_termination"`

	// Model to use for agents (default: haiku for analysis).
	Model string `json:"model"`

	// SynthModel to use for synthesis (default: sonnet).
	SynthModel string `json:"synth_model"`
}

// DefaultDispatchConfig returns sensible defaults.
func DefaultDispatchConfig() DispatchConfig {
	return DispatchConfig{
		PodSize:          6,
		MaxPods:          8,
		ContextBudget:    0.6,
		QuorumThreshold:  0.7,
		StreamResults:    true,
		EarlyTermination: true,
		Model:            "haiku",
		SynthModel:       "sonnet",
	}
}

// PodCategories defines the standard validation pods.
var PodCategories = []PodConfig{
	{
		Name:          "security",
		Category:      "security",
		AgentCount:    6,
		ContextBudget: 0.6,
		Prompt:        "Focus on: authentication, authorization, injection vulnerabilities, cryptography, secrets handling, OWASP Top 10",
	},
	{
		Name:          "architecture",
		Category:      "architecture",
		AgentCount:    6,
		ContextBudget: 0.6,
		Prompt:        "Focus on: design patterns, coupling, cohesion, scalability, error handling, API design",
	},
	{
		Name:          "quality",
		Category:      "quality",
		AgentCount:    6,
		ContextBudget: 0.6,
		Prompt:        "Focus on: code complexity, test coverage, maintainability, documentation, naming conventions",
	},
	{
		Name:          "ux",
		Category:      "ux",
		AgentCount:    4,
		ContextBudget: 0.6,
		Prompt:        "Focus on: accessibility (WCAG), performance, error messages, user feedback, loading states",
	},
	{
		Name:          "data",
		Category:      "data",
		AgentCount:    4,
		ContextBudget: 0.6,
		Prompt:        "Focus on: data validation, migrations, integrity constraints, backup/recovery, privacy",
	},
	{
		Name:          "ops",
		Category:      "ops",
		AgentCount:    4,
		ContextBudget: 0.6,
		Prompt:        "Focus on: logging, monitoring, alerting, deployment, configuration management",
	},
}

// SeverityOrder for comparison.
var SeverityOrder = map[Severity]int{
	SeverityCritical: 4,
	SeverityHigh:     3,
	SeverityMedium:   2,
	SeverityLow:      1,
	SeverityPass:     0,
}

// HigherSeverity returns the more severe of two severities.
func HigherSeverity(a, b Severity) Severity {
	if SeverityOrder[a] >= SeverityOrder[b] {
		return a
	}
	return b
}
