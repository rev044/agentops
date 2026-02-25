package main

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// formatAge
// ---------------------------------------------------------------------------

func TestInbox_formatAge(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name        string
		timestamp   time.Time
		wantSuffix  string
		wantContain string
	}{
		{"seconds ago", now.Add(-30 * time.Second), "ago", "s ago"},
		{"minutes ago", now.Add(-5 * time.Minute), "ago", "m ago"},
		{"hours ago", now.Add(-3 * time.Hour), "ago", "h ago"},
		{"days ago", now.AddDate(0, 0, -5), "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatAge(tt.timestamp)
			if got == "" {
				t.Error("formatAge returned empty string")
			}
			if tt.wantContain != "" && !strings.Contains(got, tt.wantContain) {
				t.Errorf("formatAge(%v) = %q, want to contain %q", tt.timestamp, got, tt.wantContain)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// truncateMessage
// ---------------------------------------------------------------------------

func TestInbox_truncateMessage(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		max  int
		want string
	}{
		{"short message", "Hello", 60, "Hello"},
		{"exact length", "1234567890", 10, "1234567890"},
		{"needs truncation", "This is a very long message that should be truncated", 20, "This is a very lo..."},
		{"newlines replaced", "Line 1\nLine 2", 60, "Line 1 Line 2"},
		{"whitespace trimmed", "  Hello  ", 60, "Hello"},
		{"empty string", "", 60, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateMessage(tt.msg, tt.max)
			if got != tt.want {
				t.Errorf("truncateMessage(%q, %d) = %q, want %q", tt.msg, tt.max, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// generateMessageID
// ---------------------------------------------------------------------------

func TestInbox_generateMessageID(t *testing.T) {
	id := generateMessageID()
	if !strings.HasPrefix(id, "msg-") {
		t.Errorf("generateMessageID() = %q, want prefix 'msg-'", id)
	}
	if len(id) < 10 {
		t.Errorf("generateMessageID() too short: %q", id)
	}
}

// ---------------------------------------------------------------------------
// parseSinceDuration
// ---------------------------------------------------------------------------

func TestInbox_parseSinceDuration(t *testing.T) {
	tests := []struct {
		name        string
		since       string
		wantZero    bool
		wantWarning bool
	}{
		{"empty", "", true, false},
		{"valid 5m", "5m", false, false},
		{"valid 1h", "1h", false, false},
		{"invalid abc", "abc", true, true},
		{"invalid 5x", "5x", true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cutoff, warning := parseSinceDuration(tt.since)
			if tt.wantZero && !cutoff.IsZero() {
				t.Errorf("expected zero time, got %v", cutoff)
			}
			if !tt.wantZero && cutoff.IsZero() {
				t.Errorf("expected non-zero time")
			}
			if tt.wantWarning && warning == "" {
				t.Errorf("expected warning for %q", tt.since)
			}
			if !tt.wantWarning && warning != "" {
				t.Errorf("unexpected warning for %q: %s", tt.since, warning)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// isInboxRecipient
// ---------------------------------------------------------------------------

func TestInbox_isInboxRecipient(t *testing.T) {
	tests := []struct {
		to   string
		want bool
	}{
		{"mayor", true},
		{"all", true},
		{"", true},
		{"agent-1", false},
		{"witness", false},
	}
	for _, tt := range tests {
		t.Run(tt.to, func(t *testing.T) {
			got := isInboxRecipient(tt.to)
			if got != tt.want {
				t.Errorf("isInboxRecipient(%q) = %v, want %v", tt.to, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// messageMatchesFilters
// ---------------------------------------------------------------------------

func TestInbox_messageMatchesFilters(t *testing.T) {
	now := time.Now()
	msg := Message{
		From:      "agent-1",
		To:        "mayor",
		Timestamp: now.Add(-5 * time.Minute),
		Read:      false,
	}

	tests := []struct {
		name       string
		sinceTime  time.Time
		from       string
		unreadOnly bool
		want       bool
	}{
		{"no filters", time.Time{}, "", false, true},
		{"since filter pass", now.Add(-10 * time.Minute), "", false, true},
		{"since filter fail", now, "", false, false},
		{"from filter pass", time.Time{}, "agent-1", false, true},
		{"from filter fail", time.Time{}, "witness", false, false},
		{"unread only pass", time.Time{}, "", true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := messageMatchesFilters(msg, tt.sinceTime, tt.from, tt.unreadOnly)
			if got != tt.want {
				t.Errorf("messageMatchesFilters = %v, want %v", got, tt.want)
			}
		})
	}

	// Test unread filter with a read message
	readMsg := Message{From: "agent-1", To: "mayor", Timestamp: now, Read: true}
	if messageMatchesFilters(readMsg, time.Time{}, "", true) {
		t.Error("expected false for read message with unread filter")
	}

	// Test non-inbox recipient
	agentMsg := Message{From: "agent-1", To: "agent-2", Timestamp: now}
	if messageMatchesFilters(agentMsg, time.Time{}, "", false) {
		t.Error("expected false for non-inbox recipient")
	}
}

// ---------------------------------------------------------------------------
// applyLimit
// ---------------------------------------------------------------------------

func TestInbox_applyLimit(t *testing.T) {
	msgs := []Message{
		{ID: "1"}, {ID: "2"}, {ID: "3"}, {ID: "4"}, {ID: "5"},
	}

	tests := []struct {
		name  string
		limit int
		want  int
	}{
		{"no limit (0)", 0, 5},
		{"limit 3", 3, 3},
		{"limit larger than slice", 10, 5},
		{"limit 1", 1, 1},
		{"negative limit", -1, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyLimit(msgs, tt.limit)
			if len(got) != tt.want {
				t.Errorf("applyLimit(len=%d, limit=%d) = %d, want %d", len(msgs), tt.limit, len(got), tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// scanMessagesFromReader
// ---------------------------------------------------------------------------

func TestInbox_scanMessagesFromReader(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
	}{
		{
			"valid messages",
			`{"id":"1","from":"a","to":"mayor","body":"hi","timestamp":"2024-01-01T00:00:00Z","type":"progress"}
{"id":"2","from":"b","to":"mayor","body":"bye","timestamp":"2024-01-02T00:00:00Z","type":"completion"}
`,
			2,
		},
		{
			"mixed valid and invalid",
			`{"id":"1","from":"a","to":"mayor","body":"hi","timestamp":"2024-01-01T00:00:00Z","type":"progress"}
invalid json
{"id":"2","from":"b","to":"mayor","body":"bye","timestamp":"2024-01-02T00:00:00Z","type":"completion"}
`,
			2,
		},
		{
			"all corrupted",
			"not json\nalso not json\n",
			0,
		},
		{
			"empty lines skipped",
			`{"id":"1","from":"a","to":"mayor","body":"hi","timestamp":"2024-01-01T00:00:00Z","type":"progress"}

`,
			1,
		},
		{
			"empty input",
			"",
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := bufio.NewScanner(strings.NewReader(tt.input))
			msgs := scanMessagesFromReader(scanner)
			if len(msgs) != tt.wantLen {
				t.Errorf("got %d messages, want %d", len(msgs), tt.wantLen)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// buildIDSet
// ---------------------------------------------------------------------------

func TestInbox_buildIDSet(t *testing.T) {
	msgs := []Message{
		{ID: "msg-1"},
		{ID: "msg-2"},
		{ID: "msg-3"},
	}
	idSet := buildIDSet(msgs)
	if len(idSet) != 3 {
		t.Errorf("expected 3 IDs, got %d", len(idSet))
	}
	if !idSet["msg-1"] || !idSet["msg-2"] || !idSet["msg-3"] {
		t.Error("missing expected IDs in set")
	}
	if idSet["msg-99"] {
		t.Error("unexpected ID in set")
	}
}

func TestInbox_buildIDSet_empty(t *testing.T) {
	idSet := buildIDSet(nil)
	if len(idSet) != 0 {
		t.Errorf("expected 0 IDs, got %d", len(idSet))
	}
}

// ---------------------------------------------------------------------------
// writeMessagesJSONL
// ---------------------------------------------------------------------------

func TestInbox_writeMessagesJSONL(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.jsonl")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	msgs := []Message{
		{ID: "1", From: "agent-1", To: "mayor", Body: "Test 1", Type: "progress"},
		{ID: "2", From: "agent-2", To: "mayor", Body: "Test 2", Type: "completion"},
	}

	err = writeMessagesJSONL(f, msgs)
	if err != nil {
		t.Fatalf("writeMessagesJSONL failed: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	// Read back and verify
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}

	var parsed Message
	if err := json.Unmarshal([]byte(lines[0]), &parsed); err != nil {
		t.Fatalf("unmarshal first line: %v", err)
	}
	if parsed.ID != "1" {
		t.Errorf("first message ID = %q, want '1'", parsed.ID)
	}
}

func TestInbox_writeMessagesJSONL_empty(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "empty.jsonl")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	err = writeMessagesJSONL(f, nil)
	if err != nil {
		t.Fatalf("writeMessagesJSONL(nil) failed: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	if len(data) != 0 {
		t.Errorf("expected empty file, got %d bytes", len(data))
	}
}

// ---------------------------------------------------------------------------
// appendMessage
// ---------------------------------------------------------------------------

func TestInbox_appendMessage(t *testing.T) {
	tmp := t.TempDir()
	msg := &Message{
		ID:        "msg-test-1",
		From:      "test-agent",
		To:        "mayor",
		Body:      "Hello mayor",
		Timestamp: time.Now(),
		Type:      "progress",
	}

	err := appendMessage(tmp, msg)
	if err != nil {
		t.Fatalf("appendMessage failed: %v", err)
	}

	// Verify file exists and contains the message
	messagesPath := filepath.Join(tmp, ".agents", "mail", "messages.jsonl")
	data, err := os.ReadFile(messagesPath)
	if err != nil {
		t.Fatalf("read messages file: %v", err)
	}

	var parsed Message
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(data))), &parsed); err != nil {
		t.Fatalf("unmarshal message: %v", err)
	}
	if parsed.ID != "msg-test-1" {
		t.Errorf("ID = %q, want 'msg-test-1'", parsed.ID)
	}
	if parsed.From != "test-agent" {
		t.Errorf("From = %q, want 'test-agent'", parsed.From)
	}
}

// ---------------------------------------------------------------------------
// loadAndWarnMessages
// ---------------------------------------------------------------------------

func TestInbox_loadAndWarnMessages_noFile(t *testing.T) {
	msgs, corrupted, err := loadAndWarnMessages(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msgs != nil {
		t.Errorf("expected nil messages, got %d", len(msgs))
	}
	if corrupted != 0 {
		t.Errorf("expected 0 corrupted, got %d", corrupted)
	}
}

func TestInbox_loadAndWarnMessages_withCorrupted(t *testing.T) {
	content := `{"id":"msg-1","from":"agent","to":"mayor","body":"test","timestamp":"2024-01-01T00:00:00Z","type":"progress"}
invalid json
{"id":"msg-2","from":"agent","to":"mayor","body":"test2","timestamp":"2024-01-02T00:00:00Z","type":"progress"}
`
	dir := setupInboxDir(t, content)
	msgs, corrupted, err := loadAndWarnMessages(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 2 {
		t.Errorf("expected 2 messages, got %d", len(msgs))
	}
	if corrupted != 1 {
		t.Errorf("expected 1 corrupted, got %d", corrupted)
	}
}

// ---------------------------------------------------------------------------
// renderInboxTable (smoke test)
// ---------------------------------------------------------------------------

func TestInbox_renderInboxTable(t *testing.T) {
	now := time.Now()
	messages := []Message{
		{From: "agent-1", Type: "progress", Body: "Working on it", Timestamp: now, Read: false},
		{From: "witness", Type: "blocker", Body: "Agent stuck", Timestamp: now.Add(-5 * time.Minute), Read: true},
	}
	// Should not panic. Output goes to stdout.
	renderInboxTable(messages, 5)
}

func TestInbox_renderInboxTable_truncated(t *testing.T) {
	now := time.Now()
	messages := []Message{
		{From: "agent-1", Type: "progress", Body: "Test", Timestamp: now},
	}
	// totalMatching > len(limited) should show "Showing X of Y"
	renderInboxTable(messages, 10)
}

// ---------------------------------------------------------------------------
// openLockedFile
// ---------------------------------------------------------------------------

func TestInbox_openLockedFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.jsonl")
	if err := os.WriteFile(path, []byte("test\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	file, cleanup, err := openLockedFile(path)
	if err != nil {
		t.Fatalf("openLockedFile failed: %v", err)
	}
	if file == nil {
		t.Fatal("expected non-nil file")
	}
	if cleanup == nil {
		t.Fatal("expected non-nil cleanup function")
	}

	// Read to verify it works
	data := make([]byte, 5)
	n, err := file.Read(data)
	if err != nil {
		t.Fatalf("read from locked file: %v", err)
	}
	if n != 5 {
		t.Errorf("read %d bytes, want 5", n)
	}

	if err := cleanup(); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}
}

func TestInbox_openLockedFile_missingFile(t *testing.T) {
	_, _, err := openLockedFile("/nonexistent/path")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

// ---------------------------------------------------------------------------
// Message JSON round-trip
// ---------------------------------------------------------------------------

func TestInbox_MessageJSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	msg := Message{
		ID:        "msg-12345",
		From:      "agent-1",
		To:        "mayor",
		Body:      "Test message body",
		Timestamp: now,
		Read:      false,
		Type:      "progress",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != msg.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, msg.ID)
	}
	if decoded.From != msg.From {
		t.Errorf("From = %q, want %q", decoded.From, msg.From)
	}
	if decoded.To != msg.To {
		t.Errorf("To = %q, want %q", decoded.To, msg.To)
	}
	if decoded.Body != msg.Body {
		t.Errorf("Body = %q, want %q", decoded.Body, msg.Body)
	}
	if decoded.Read != msg.Read {
		t.Errorf("Read = %v, want %v", decoded.Read, msg.Read)
	}
	if decoded.Type != msg.Type {
		t.Errorf("Type = %q, want %q", decoded.Type, msg.Type)
	}
}

// ---------------------------------------------------------------------------
// markMessagesRead
// ---------------------------------------------------------------------------

func TestInbox_markMessagesRead(t *testing.T) {
	tmp := t.TempDir()
	mailDir := filepath.Join(tmp, ".agents", "mail")
	if err := os.MkdirAll(mailDir, 0o700); err != nil {
		t.Fatal(err)
	}

	// Write initial unread messages
	msgs := []Message{
		{ID: "msg-1", From: "a", To: "mayor", Body: "hi", Timestamp: time.Now(), Read: false, Type: "progress"},
		{ID: "msg-2", From: "b", To: "mayor", Body: "bye", Timestamp: time.Now(), Read: false, Type: "progress"},
		{ID: "msg-3", From: "c", To: "mayor", Body: "done", Timestamp: time.Now(), Read: false, Type: "completion"},
	}
	messagesPath := filepath.Join(mailDir, "messages.jsonl")
	f, err := os.Create(messagesPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, msg := range msgs {
		data, _ := json.Marshal(msg)
		f.WriteString(string(data) + "\n")
	}
	f.Close()

	// Mark msg-1 and msg-3 as read
	toMark := []Message{msgs[0], msgs[2]}
	if err := markMessagesRead(tmp, toMark); err != nil {
		t.Fatalf("markMessagesRead failed: %v", err)
	}

	// Reload and verify
	loaded, _, err := loadMessages(tmp)
	if err != nil {
		t.Fatalf("loadMessages failed: %v", err)
	}
	for _, msg := range loaded {
		switch msg.ID {
		case "msg-1":
			if !msg.Read {
				t.Error("msg-1 should be read")
			}
		case "msg-2":
			if msg.Read {
				t.Error("msg-2 should remain unread")
			}
		case "msg-3":
			if !msg.Read {
				t.Error("msg-3 should be read")
			}
		}
	}
}

// ---------------------------------------------------------------------------
// filterMessages (comprehensive)
// ---------------------------------------------------------------------------

func TestInbox_filterMessages_combined(t *testing.T) {
	now := time.Now()
	messages := []Message{
		{ID: "1", From: "agent-1", To: "mayor", Timestamp: now, Read: false},
		{ID: "2", From: "witness", To: "mayor", Timestamp: now.Add(-30 * time.Minute), Read: true},
		{ID: "3", From: "agent-1", To: "all", Timestamp: now.Add(-2 * time.Hour), Read: false},
		{ID: "4", From: "agent-2", To: "agent-1", Timestamp: now, Read: false}, // not inbox recipient
	}

	// Combine from + unread
	filtered, _ := filterMessages(messages, "", "agent-1", true)
	if len(filtered) != 2 { // msg 1 and 3
		t.Errorf("expected 2 messages for agent-1 unread, got %d", len(filtered))
	}

	// Combine since + from
	filtered2, _ := filterMessages(messages, "1h", "agent-1", false)
	if len(filtered2) != 1 { // only msg 1
		t.Errorf("expected 1 message for agent-1 in last 1h, got %d", len(filtered2))
	}
}

// ---------------------------------------------------------------------------
// renderInboxJSON (smoke test)
// ---------------------------------------------------------------------------

func TestInbox_renderInboxJSON(t *testing.T) {
	msgs := []Message{
		{ID: "1", From: "agent", To: "mayor", Body: "test", Type: "progress"},
	}
	// Outputs to stdout; just ensure no error
	err := renderInboxJSON(msgs, 5, 1)
	if err != nil {
		t.Fatalf("renderInboxJSON failed: %v", err)
	}
}
