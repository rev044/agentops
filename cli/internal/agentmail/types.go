// Package agentmail provides parsing and types for Agent Mail messages
// used in distributed mode coordination between crank, swarm, and demigods.
package agentmail

import (
	"time"
)

// MessageType represents the type of Agent Mail message.
type MessageType string

const (
	// MessageTypeBeadAccepted indicates a demigod has accepted a bead to work on.
	MessageTypeBeadAccepted MessageType = "BEAD_ACCEPTED"

	// MessageTypeProgress is a periodic progress update during implementation.
	MessageTypeProgress MessageType = "PROGRESS"

	// MessageTypeHelpRequest indicates a demigod is stuck and needs guidance.
	MessageTypeHelpRequest MessageType = "HELP_REQUEST"

	// MessageTypeHelpResponse is a response to a HELP_REQUEST.
	MessageTypeHelpResponse MessageType = "HELP_RESPONSE"

	// MessageTypeOfferingReady indicates work is complete and ready for review.
	MessageTypeOfferingReady MessageType = "OFFERING_READY"

	// MessageTypeDone is an alternate completion message.
	MessageTypeDone MessageType = "DONE"

	// MessageTypeFailed indicates implementation failed.
	MessageTypeFailed MessageType = "FAILED"

	// MessageTypeCheckpoint indicates context exhaustion with partial progress.
	MessageTypeCheckpoint MessageType = "CHECKPOINT"

	// MessageTypeSpawnRequest requests spawning a new worker.
	MessageTypeSpawnRequest MessageType = "SPAWN_REQUEST"

	// MessageTypeSpawnAck acknowledges a spawn request.
	MessageTypeSpawnAck MessageType = "SPAWN_ACK"

	// MessageTypeUnknown is used when the message type cannot be determined.
	MessageTypeUnknown MessageType = "UNKNOWN"
)

// IsCompletionType returns true if this message type signals task completion.
func (t MessageType) IsCompletionType() bool {
	switch t {
	case MessageTypeOfferingReady, MessageTypeDone, MessageTypeFailed, MessageTypeCheckpoint:
		return true
	default:
		return false
	}
}

// IsSuccessType returns true if this message type signals successful completion.
func (t MessageType) IsSuccessType() bool {
	return t == MessageTypeOfferingReady || t == MessageTypeDone
}

// HelpRequestIssueType categorizes the type of help needed.
type HelpRequestIssueType string

const (
	HelpRequestIssueTypeStuck       HelpRequestIssueType = "STUCK"
	HelpRequestIssueTypeSpecUnclear HelpRequestIssueType = "SPEC_UNCLEAR"
	HelpRequestIssueTypeBlocked     HelpRequestIssueType = "BLOCKED"
	HelpRequestIssueTypeTechnical   HelpRequestIssueType = "TECHNICAL"
)

// FailureType categorizes the type of failure.
type FailureType string

const (
	FailureTypeTestsFail      FailureType = "TESTS_FAIL"
	FailureTypeBuildFail      FailureType = "BUILD_FAIL"
	FailureTypeSpecImpossible FailureType = "SPEC_IMPOSSIBLE"
	FailureTypeContextHigh    FailureType = "CONTEXT_HIGH"
	FailureTypeError          FailureType = "ERROR"
)

// CheckpointReason categorizes why a checkpoint was created.
type CheckpointReason string

const (
	CheckpointReasonContextHigh CheckpointReason = "CONTEXT_HIGH"
	CheckpointReasonManual      CheckpointReason = "MANUAL"
	CheckpointReasonTimeout     CheckpointReason = "TIMEOUT"
)

// Message represents a parsed Agent Mail message with structured content.
type Message struct {
	// ID is the unique identifier for this message.
	ID string `json:"id,omitempty"`

	// Type is the parsed message type.
	Type MessageType `json:"type"`

	// Subject is the original message subject line.
	Subject string `json:"subject"`

	// From is the sender's agent identifier.
	From string `json:"from"`

	// To is the recipient's agent identifier.
	To string `json:"to"`

	// Body is the raw message body (markdown).
	Body string `json:"body,omitempty"`

	// ThreadID groups related messages.
	ThreadID string `json:"thread_id,omitempty"`

	// Timestamp is when the message was sent/received.
	Timestamp time.Time `json:"timestamp,omitempty"`

	// AckRequired indicates if acknowledgement is needed.
	AckRequired bool `json:"ack_required,omitempty"`

	// Acknowledged indicates if the message has been acknowledged.
	Acknowledged bool `json:"acknowledged,omitempty"`

	// Parsed contains type-specific parsed content.
	Parsed ParsedContent `json:"parsed,omitempty"`
}

// ParsedContent contains type-specific parsed fields.
type ParsedContent struct {
	// BeadID is the issue/bead identifier (used by most message types).
	BeadID string `json:"bead_id,omitempty"`

	// Title is the bead/issue title.
	Title string `json:"title,omitempty"`

	// Status indicates completion status (DONE, FAILED, etc.).
	Status string `json:"status,omitempty"`

	// -- Progress fields --

	// Step indicates current step in progress.
	Step string `json:"step,omitempty"`

	// ContextUsage is approximate context window usage percentage.
	ContextUsage int `json:"context_usage,omitempty"`

	// FilesTouched lists files modified.
	FilesTouched []string `json:"files_touched,omitempty"`

	// -- Help Request fields --

	// IssueType categorizes the help request.
	IssueType HelpRequestIssueType `json:"issue_type,omitempty"`

	// Problem describes the issue.
	Problem string `json:"problem,omitempty"`

	// WhatTried describes approaches attempted.
	WhatTried string `json:"what_tried,omitempty"`

	// Question is the specific question needing an answer.
	Question string `json:"question,omitempty"`

	// -- Completion fields (OFFERING_READY, DONE) --

	// CommitSHA is the commit hash for the changes.
	CommitSHA string `json:"commit_sha,omitempty"`

	// Files lists changed files.
	Files []string `json:"files,omitempty"`

	// TestsPass indicates test status.
	TestsPass bool `json:"tests_pass,omitempty"`

	// LintPass indicates lint status.
	LintPass bool `json:"lint_pass,omitempty"`

	// BuildPass indicates build status.
	BuildPass bool `json:"build_pass,omitempty"`

	// Summary is a brief description of what was done.
	Summary string `json:"summary,omitempty"`

	// -- Failure fields --

	// FailureType categorizes the failure.
	FailureType FailureType `json:"failure_type,omitempty"`

	// Reason describes why it failed.
	Reason string `json:"reason,omitempty"`

	// InternalAttempts is how many retries were attempted.
	InternalAttempts int `json:"internal_attempts,omitempty"`

	// Recommendation suggests what to do next.
	Recommendation string `json:"recommendation,omitempty"`

	// -- Checkpoint fields --

	// CheckpointReason explains why checkpoint was created.
	CheckpointReason CheckpointReason `json:"checkpoint_reason,omitempty"`

	// PartialCommitSHA is the commit with partial work.
	PartialCommitSHA string `json:"partial_commit_sha,omitempty"`

	// Progress describes what's done and what remains.
	Progress string `json:"progress,omitempty"`

	// NextSteps is guidance for the successor demigod.
	NextSteps string `json:"next_steps,omitempty"`

	// -- Spawn fields --

	// IssueID is the issue to spawn for.
	IssueID string `json:"issue_id,omitempty"`

	// Resume indicates if resuming from checkpoint.
	Resume bool `json:"resume,omitempty"`

	// Orchestrator is the requesting orchestrator.
	Orchestrator string `json:"orchestrator,omitempty"`
}

// RawMessage represents the raw Agent Mail message structure from MCP.
type RawMessage struct {
	ID          string `json:"id"`
	ProjectKey  string `json:"project_key"`
	SenderName  string `json:"sender_name"`
	To          string `json:"to"`
	Subject     string `json:"subject"`
	BodyMD      string `json:"body_md"`
	ThreadID    string `json:"thread_id,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
	AckRequired bool   `json:"ack_required,omitempty"`
	Acked       bool   `json:"acked,omitempty"`
}
