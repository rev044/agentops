package agentmail

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Parser provides message parsing for Agent Mail messages.
type Parser struct {
	// Strict mode returns errors for malformed messages instead of partial parses.
	Strict bool
}

// NewParser creates a parser with default settings.
func NewParser() *Parser {
	return &Parser{
		Strict: false,
	}
}

// ParseError represents an error during message parsing.
type ParseError struct {
	Field   string
	Message string
	Raw     string
}

func (e *ParseError) Error() string {
	if e.Field != "" {
		return "parse error in " + e.Field + ": " + e.Message
	}
	return "parse error: " + e.Message
}

// Parse parses a raw Agent Mail message into a structured Message.
func (p *Parser) Parse(raw *RawMessage) (*Message, error) {
	if raw == nil {
		return nil, &ParseError{Message: "nil message"}
	}

	msg := &Message{
		ID:           raw.ID,
		Subject:      raw.Subject,
		From:         raw.SenderName,
		To:           raw.To,
		Body:         raw.BodyMD,
		ThreadID:     raw.ThreadID,
		AckRequired:  raw.AckRequired,
		Acknowledged: raw.Acked,
	}

	// Parse timestamp
	if raw.Timestamp != "" {
		if ts, err := time.Parse(time.RFC3339, raw.Timestamp); err == nil {
			msg.Timestamp = ts
		}
	}

	// Determine message type from subject
	msg.Type = p.parseMessageType(raw.Subject)

	// Parse type-specific content from body
	msg.Parsed = p.parseBody(msg.Type, raw.BodyMD, raw.Subject)

	// Extract bead ID from subject if not found in body
	if msg.Parsed.BeadID == "" {
		msg.Parsed.BeadID = p.extractBeadIDFromSubject(raw.Subject)
	}

	return msg, nil
}

// ParseBatch parses multiple raw messages.
func (p *Parser) ParseBatch(raws []*RawMessage) ([]*Message, []error) {
	messages := make([]*Message, 0, len(raws))
	var errs []error

	for _, raw := range raws {
		msg, err := p.Parse(raw)
		if err != nil {
			errs = append(errs, err)
			if p.Strict {
				continue
			}
		}
		if msg != nil {
			messages = append(messages, msg)
		}
	}

	return messages, errs
}

// parseMessageType determines the message type from the subject line.
func (p *Parser) parseMessageType(subject string) MessageType {
	upper := strings.ToUpper(subject)

	// Check for exact matches first
	switch {
	case strings.Contains(upper, "BEAD_ACCEPTED"):
		return MessageTypeBeadAccepted
	case strings.Contains(upper, "OFFERING_READY"):
		return MessageTypeOfferingReady
	case strings.Contains(upper, "HELP_REQUEST"):
		return MessageTypeHelpRequest
	case strings.Contains(upper, "HELP_RESPONSE"):
		return MessageTypeHelpResponse
	case strings.Contains(upper, "SPAWN_REQUEST"):
		return MessageTypeSpawnRequest
	case strings.Contains(upper, "SPAWN_ACK"):
		return MessageTypeSpawnAck
	case strings.Contains(upper, "CHECKPOINT"):
		return MessageTypeCheckpoint
	case strings.Contains(upper, "PROGRESS"):
		return MessageTypeProgress
	case strings.Contains(upper, "FAILED"):
		return MessageTypeFailed
	case strings.Contains(upper, "DONE"):
		return MessageTypeDone
	default:
		return MessageTypeUnknown
	}
}

// extractBeadIDFromSubject extracts bead ID from subject like "[ol-527.1] PROGRESS"
func (p *Parser) extractBeadIDFromSubject(subject string) string {
	// Pattern: [bead-id] or bead-id: at start
	bracketRe := regexp.MustCompile(`^\[([^\]]+)\]`)
	if m := bracketRe.FindStringSubmatch(subject); len(m) > 1 {
		return m[1]
	}

	colonRe := regexp.MustCompile(`^([a-zA-Z0-9]+-[a-zA-Z0-9.]+):`)
	if m := colonRe.FindStringSubmatch(subject); len(m) > 1 {
		return m[1]
	}

	return ""
}

// parseBody extracts type-specific fields from the message body.
func (p *Parser) parseBody(msgType MessageType, body string, subject string) ParsedContent {
	content := ParsedContent{}

	switch msgType {
	case MessageTypeBeadAccepted:
		p.parseBeadAccepted(&content, body)
	case MessageTypeProgress:
		p.parseProgress(&content, body)
	case MessageTypeHelpRequest:
		p.parseHelpRequest(&content, body)
	case MessageTypeOfferingReady, MessageTypeDone:
		p.parseCompletion(&content, body)
	case MessageTypeFailed:
		p.parseFailed(&content, body)
	case MessageTypeCheckpoint:
		p.parseCheckpoint(&content, body)
	case MessageTypeSpawnRequest:
		p.parseSpawnRequest(&content, body)
	case MessageTypeSpawnAck:
		p.parseSpawnAck(&content, body)
	}

	return content
}

// parseBeadAccepted extracts fields from BEAD_ACCEPTED messages.
// Expected format:
//
//	Accepted bead: <bead-id>
//	Title: <title>
//	Starting implementation at: <timestamp>
func (p *Parser) parseBeadAccepted(content *ParsedContent, body string) {
	content.BeadID = p.extractField(body, "Accepted bead:")
	content.Title = p.extractField(body, "Title:")
}

// parseProgress extracts fields from PROGRESS messages.
// Expected format:
//
//	Bead: <bead-id>
//	Step: <step description>
//	Status: <what's happening>
//	Context usage: <N%>
//	Files touched: <list>
func (p *Parser) parseProgress(content *ParsedContent, body string) {
	content.BeadID = p.extractField(body, "Bead:")
	content.Step = p.extractField(body, "Step:")
	content.Status = p.extractField(body, "Status:")

	// Parse context usage percentage
	if usage := p.extractField(body, "Context usage:"); usage != "" {
		usage = strings.TrimSuffix(strings.TrimSpace(usage), "%")
		if n, err := strconv.Atoi(usage); err == nil {
			content.ContextUsage = n
		}
	}

	// Parse files touched
	if files := p.extractField(body, "Files touched:"); files != "" {
		content.FilesTouched = p.parseFileList(files)
	}
}

// parseHelpRequest extracts fields from HELP_REQUEST messages.
// Expected format:
//
//	Bead: <bead-id>
//	Issue Type: STUCK | SPEC_UNCLEAR | BLOCKED | TECHNICAL
//	## Problem
//	<description>
//	## What I Tried
//	<approaches>
//	## Files Touched
//	- path/to/file
//	## Question
//	<specific question>
func (p *Parser) parseHelpRequest(content *ParsedContent, body string) {
	content.BeadID = p.extractField(body, "Bead:")

	// Parse issue type
	if issueType := p.extractField(body, "Issue Type:"); issueType != "" {
		content.IssueType = HelpRequestIssueType(strings.ToUpper(strings.TrimSpace(issueType)))
	}

	// Extract sections
	content.Problem = p.extractSection(body, "## Problem", "##")
	content.WhatTried = p.extractSection(body, "## What I Tried", "##")
	content.Question = p.extractSection(body, "## Question", "##")

	// Parse files touched
	if files := p.extractSection(body, "## Files Touched", "##"); files != "" {
		content.FilesTouched = p.parseMarkdownList(files)
	}
}

// parseCompletion extracts fields from OFFERING_READY/DONE messages.
// Expected format:
//
//	Bead: <bead-id>
//	Status: DONE
//	## Changes
//	- Commit: <sha>
//	- Files: <list>
//	## Self-Validation
//	- Tests: PASS/FAIL
//	- Lint: PASS/FAIL
//	- Build: PASS/FAIL
//	## Summary
//	<description>
func (p *Parser) parseCompletion(content *ParsedContent, body string) {
	content.BeadID = p.extractField(body, "Bead:")
	content.Status = p.extractField(body, "Status:")

	// Extract changes section
	changes := p.extractSection(body, "## Changes", "##")
	if commit := p.extractField(changes, "Commit:"); commit != "" {
		content.CommitSHA = strings.TrimSpace(commit)
	}
	if files := p.extractField(changes, "Files:"); files != "" {
		content.Files = p.parseFileList(files)
	}

	// Extract validation section
	validation := p.extractSection(body, "## Self-Validation", "##")
	content.TestsPass = p.checkPassFail(p.extractField(validation, "Tests:"))
	content.LintPass = p.checkPassFail(p.extractField(validation, "Lint:"))
	content.BuildPass = p.checkPassFail(p.extractField(validation, "Build:"))

	// Extract summary
	content.Summary = strings.TrimSpace(p.extractSection(body, "## Summary", "##"))
}

// parseFailed extracts fields from FAILED messages.
// Expected format:
//
//	Bead: <bead-id>
//	Status: FAILED
//	## Failure
//	Type: TESTS_FAIL | BUILD_FAIL | SPEC_IMPOSSIBLE | ERROR
//	Reason: <description>
//	Internal Attempts: <count>
//	## Partial Progress
//	- Commit: <sha>
//	- Files: <list>
//	## Recommendation
//	<what to do>
func (p *Parser) parseFailed(content *ParsedContent, body string) {
	content.BeadID = p.extractField(body, "Bead:")
	content.Status = "FAILED"

	// Extract failure section
	failure := p.extractSection(body, "## Failure", "##")
	if failType := p.extractField(failure, "Type:"); failType != "" {
		content.FailureType = FailureType(strings.ToUpper(strings.TrimSpace(failType)))
	}
	content.Reason = p.extractField(failure, "Reason:")

	if attempts := p.extractField(failure, "Internal Attempts:"); attempts != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(attempts)); err == nil {
			content.InternalAttempts = n
		}
	}

	// Extract partial progress
	partial := p.extractSection(body, "## Partial Progress", "##")
	if commit := p.extractField(partial, "Commit:"); commit != "" {
		content.PartialCommitSHA = strings.TrimSpace(commit)
	}
	if files := p.extractField(partial, "Files:"); files != "" {
		content.Files = p.parseFileList(files)
	}

	// Extract recommendation
	content.Recommendation = strings.TrimSpace(p.extractSection(body, "## Recommendation", "##"))
}

// parseCheckpoint extracts fields from CHECKPOINT messages.
// Expected format:
//
//	Bead: <bead-id>
//	Reason: CONTEXT_HIGH
//	## Progress
//	- Commit: <sha>
//	- Description: <what's done>
//	- Context usage: <pct>%
//	## Next Steps for Successor
//	<guidance>
func (p *Parser) parseCheckpoint(content *ParsedContent, body string) {
	content.BeadID = p.extractField(body, "Bead:")

	if reason := p.extractField(body, "Reason:"); reason != "" {
		content.CheckpointReason = CheckpointReason(strings.ToUpper(strings.TrimSpace(reason)))
	}

	// Extract progress section
	progress := p.extractSection(body, "## Progress", "##")
	if commit := p.extractField(progress, "Commit:"); commit != "" {
		content.PartialCommitSHA = strings.TrimSpace(commit)
	}
	content.Progress = p.extractField(progress, "Description:")

	if usage := p.extractField(progress, "Context usage:"); usage != "" {
		usage = strings.TrimSuffix(strings.TrimSpace(usage), "%")
		if n, err := strconv.Atoi(usage); err == nil {
			content.ContextUsage = n
		}
	}

	// Extract next steps
	content.NextSteps = strings.TrimSpace(p.extractSection(body, "## Next Steps for Successor", "##"))
}

// parseSpawnRequest extracts fields from SPAWN_REQUEST messages.
func (p *Parser) parseSpawnRequest(content *ParsedContent, body string) {
	content.IssueID = p.extractField(body, "Issue:")
	if content.IssueID == "" {
		content.IssueID = p.extractField(body, "Bead:")
	}

	resume := strings.ToLower(p.extractField(body, "Resume:"))
	content.Resume = resume == "true" || resume == "yes"

	if checkpoint := p.extractField(body, "Checkpoint:"); checkpoint != "" {
		content.PartialCommitSHA = strings.TrimSpace(checkpoint)
	}

	content.Orchestrator = p.extractField(body, "Orchestrator:")
}

// parseSpawnAck extracts fields from SPAWN_ACK messages.
func (p *Parser) parseSpawnAck(content *ParsedContent, body string) {
	content.IssueID = p.extractField(body, "Issue:")
	if content.IssueID == "" {
		content.IssueID = p.extractField(body, "Bead:")
	}
	content.Status = p.extractField(body, "Status:")
}

// extractField extracts a single-line field value after a label.
// For "Label: value", returns "value".
func (p *Parser) extractField(text, label string) string {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if idx := strings.Index(line, label); idx != -1 {
			value := strings.TrimSpace(line[idx+len(label):])
			return value
		}
	}
	return ""
}

// extractSection extracts content between a header and the next header.
func (p *Parser) extractSection(text, startHeader, endMarker string) string {
	lower := strings.ToLower(text)
	startLower := strings.ToLower(startHeader)

	start := strings.Index(lower, startLower)
	if start == -1 {
		return ""
	}

	// Move past the header line
	start += len(startHeader)
	if idx := strings.Index(text[start:], "\n"); idx != -1 {
		start += idx + 1
	}

	// Find the next section or end
	remaining := text[start:]
	end := len(remaining)

	// Look for next header
	if endMarker != "" {
		if idx := strings.Index(remaining, endMarker); idx != -1 {
			end = idx
		}
	}

	return strings.TrimSpace(remaining[:end])
}

// parseFileList parses a comma-separated or newline-separated file list.
func (p *Parser) parseFileList(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	var files []string

	// Try comma separation first
	if strings.Contains(s, ",") {
		parts := strings.Split(s, ",")
		for _, part := range parts {
			if f := strings.TrimSpace(part); f != "" {
				files = append(files, f)
			}
		}
		return files
	}

	// Try newline separation
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		if f := strings.TrimSpace(line); f != "" && !strings.HasPrefix(f, "-") {
			files = append(files, f)
		} else if strings.HasPrefix(strings.TrimSpace(line), "-") {
			// Handle markdown list items
			f := strings.TrimPrefix(strings.TrimSpace(line), "-")
			f = strings.TrimSpace(f)
			if f != "" {
				files = append(files, f)
			}
		}
	}

	return files
}

// parseMarkdownList parses a markdown bullet list.
func (p *Parser) parseMarkdownList(s string) []string {
	var items []string
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
			item := strings.TrimPrefix(line, "-")
			item = strings.TrimPrefix(item, "*")
			item = strings.TrimSpace(item)
			if item != "" {
				items = append(items, item)
			}
		}
	}
	return items
}

// checkPassFail returns true if the value indicates PASS.
func (p *Parser) checkPassFail(s string) bool {
	upper := strings.ToUpper(strings.TrimSpace(s))
	return upper == "PASS" || upper == "TRUE" || upper == "YES" || upper == "OK"
}

// FilterByType returns messages of a specific type.
func FilterByType(messages []*Message, msgType MessageType) []*Message {
	var filtered []*Message
	for _, msg := range messages {
		if msg.Type == msgType {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

// FilterByBeadID returns messages for a specific bead.
func FilterByBeadID(messages []*Message, beadID string) []*Message {
	var filtered []*Message
	for _, msg := range messages {
		if msg.Parsed.BeadID == beadID {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

// FilterCompletions returns all completion messages (OFFERING_READY, DONE, FAILED, CHECKPOINT).
func FilterCompletions(messages []*Message) []*Message {
	var filtered []*Message
	for _, msg := range messages {
		if msg.Type.IsCompletionType() {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

// FilterSuccesses returns successful completion messages.
func FilterSuccesses(messages []*Message) []*Message {
	var filtered []*Message
	for _, msg := range messages {
		if msg.Type.IsSuccessType() {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

// FilterPending returns messages that need acknowledgement.
func FilterPending(messages []*Message) []*Message {
	var filtered []*Message
	for _, msg := range messages {
		if msg.AckRequired && !msg.Acknowledged {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}
