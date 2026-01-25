package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/boshu2/agentops/plugins/olympus-kit/cli/internal/types"
)

func TestParser_Parse(t *testing.T) {
	jsonl := `{"type":"user","sessionId":"test","timestamp":"2026-01-24T10:00:00.000Z","uuid":"1","message":{"role":"user","content":"Hello"}}
{"type":"assistant","sessionId":"test","timestamp":"2026-01-24T10:00:10.000Z","uuid":"2","message":{"role":"assistant","content":"Hi there!"}}
`
	p := NewParser()
	result, err := p.Parse(strings.NewReader(jsonl))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result.TotalLines != 2 {
		t.Errorf("TotalLines = %d, want 2", result.TotalLines)
	}

	if len(result.Messages) != 2 {
		t.Fatalf("Messages count = %d, want 2", len(result.Messages))
	}

	if result.Messages[0].Role != "user" {
		t.Errorf("First message role = %q, want %q", result.Messages[0].Role, "user")
	}

	if result.Messages[1].Content != "Hi there!" {
		t.Errorf("Second message content = %q, want %q", result.Messages[1].Content, "Hi there!")
	}
}

func TestParser_SkipMalformed(t *testing.T) {
	jsonl := `{"type":"user","sessionId":"test","timestamp":"2026-01-24T10:00:00.000Z","uuid":"1","message":{"role":"user","content":"Valid"}}
{malformed json
{"type":"assistant","sessionId":"test","timestamp":"2026-01-24T10:00:10.000Z","uuid":"2","message":{"role":"assistant","content":"Also valid"}}
`
	p := NewParser()
	p.SkipMalformed = true

	result, err := p.Parse(strings.NewReader(jsonl))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result.MalformedLines != 1 {
		t.Errorf("MalformedLines = %d, want 1", result.MalformedLines)
	}

	if len(result.Messages) != 2 {
		t.Errorf("Messages count = %d, want 2", len(result.Messages))
	}
}

func TestParser_Truncation(t *testing.T) {
	longContent := strings.Repeat("x", 600)
	jsonl := `{"type":"user","sessionId":"test","timestamp":"2026-01-24T10:00:00.000Z","uuid":"1","message":{"role":"user","content":"` + longContent + `"}}`

	p := NewParser()
	p.MaxContentLength = 500

	result, err := p.Parse(strings.NewReader(jsonl))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Messages) != 1 {
		t.Fatalf("Messages count = %d, want 1", len(result.Messages))
	}

	content := result.Messages[0].Content
	if !strings.HasSuffix(content, "... [truncated]") {
		t.Errorf("Content not truncated correctly: %s", content)
	}

	// 500 chars + "... [truncated]" = ~515
	if len(content) > 520 {
		t.Errorf("Truncated content too long: %d chars", len(content))
	}
}

func TestParser_SkipNonMessageTypes(t *testing.T) {
	jsonl := `{"type":"file-history-snapshot","messageId":"123","snapshot":{}}
{"type":"progress","data":{"type":"hook_progress"}}
{"type":"user","sessionId":"test","timestamp":"2026-01-24T10:00:00.000Z","uuid":"1","message":{"role":"user","content":"Real message"}}
`
	p := NewParser()
	result, err := p.Parse(strings.NewReader(jsonl))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Messages) != 1 {
		t.Errorf("Messages count = %d, want 1 (should skip non-message types)", len(result.Messages))
	}
}

func TestParser_ParseFile_Fixtures(t *testing.T) {
	fixtures := []struct {
		name        string
		minMessages int
	}{
		{"simple-decision.jsonl", 4},
		{"multi-extract.jsonl", 5},
		{"tool-heavy.jsonl", 5},
		{"long-session.jsonl", 100},
		{"edge-cases.jsonl", 5},
	}

	p := NewParser()
	fixtureDir := "../../testdata/transcripts"

	for _, tc := range fixtures {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(fixtureDir, tc.name)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Skipf("Fixture not found: %s", path)
			}

			result, err := p.ParseFile(path)
			if err != nil {
				t.Fatalf("ParseFile failed: %v", err)
			}

			if len(result.Messages) < tc.minMessages {
				t.Errorf("Messages = %d, want at least %d", len(result.Messages), tc.minMessages)
			}
		})
	}
}

func TestParser_Unicode(t *testing.T) {
	jsonl := `{"type":"user","sessionId":"test","timestamp":"2026-01-24T10:00:00.000Z","uuid":"1","message":{"role":"user","content":"你好世界 🚀 émojis"}}`

	p := NewParser()
	result, err := p.Parse(strings.NewReader(jsonl))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Messages) != 1 {
		t.Fatalf("Messages count = %d, want 1", len(result.Messages))
	}

	if !strings.Contains(result.Messages[0].Content, "你好世界") {
		t.Error("Unicode content not preserved")
	}

	if !strings.Contains(result.Messages[0].Content, "🚀") {
		t.Error("Emoji not preserved")
	}
}

func TestParser_ParseChannel(t *testing.T) {
	jsonl := `{"type":"user","sessionId":"test","timestamp":"2026-01-24T10:00:00.000Z","uuid":"1","message":{"role":"user","content":"One"}}
{"type":"user","sessionId":"test","timestamp":"2026-01-24T10:00:10.000Z","uuid":"2","message":{"role":"user","content":"Two"}}
{"type":"user","sessionId":"test","timestamp":"2026-01-24T10:00:20.000Z","uuid":"3","message":{"role":"user","content":"Three"}}
`
	p := NewParser()
	msgCh, errCh := p.ParseChannel(strings.NewReader(jsonl))

	count := 0
	for range msgCh {
		count++
	}

	if err := <-errCh; err != nil {
		t.Fatalf("ParseChannel error: %v", err)
	}

	if count != 3 {
		t.Errorf("Message count = %d, want 3", count)
	}
}

func TestExtractor_Extract(t *testing.T) {
	e := NewExtractor()

	tests := []struct {
		name     string
		content  string
		wantType string
	}{
		{
			name:     "Decision pattern",
			content:  "**Decision:** Use context.WithCancel for graceful shutdown.",
			wantType: "decision",
		},
		{
			name:     "Solution pattern",
			content:  "**Solution:** Fixed the bug by adding null check.",
			wantType: "solution",
		},
		{
			name:     "Learning pattern",
			content:  "**Learning:** Always validate JWT expiration claims.",
			wantType: "learning",
		},
		{
			name:     "Failure pattern",
			content:  "**Failure:** Caching auth responses didn't work because of session isolation.",
			wantType: "failure",
		},
		{
			name:     "Reference with URL",
			content:  "See https://example.com/docs for more info.",
			wantType: "reference",
		},
		{
			name:     "No match",
			content:  "Just a regular message without any patterns.",
			wantType: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := createTestMessage(tc.content)
			results := e.Extract(msg)

			if tc.wantType == "" {
				if len(results) > 0 {
					t.Errorf("Expected no match, got %d", len(results))
				}
				return
			}

			found := false
			for _, r := range results {
				if string(r.Type) == tc.wantType {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected type %q not found in results", tc.wantType)
			}
		})
	}
}

func TestExtractor_ExtractBest(t *testing.T) {
	e := NewExtractor()

	// Message with multiple patterns - should return highest score
	content := "**Decision:** Use X. Also **Learning:** This teaches us Y."
	msg := createTestMessage(content)

	best := e.ExtractBest(msg)
	if best == nil {
		t.Fatal("Expected a result, got nil")
	}

	// The pattern match should give higher score than keyword
	if best.Score < 0.6 {
		t.Errorf("Score = %f, want >= 0.6", best.Score)
	}
}

// createTestMessage creates a TranscriptMessage for testing.
func createTestMessage(content string) types.TranscriptMessage {
	return types.TranscriptMessage{
		Type:    "assistant",
		Role:    "assistant",
		Content: content,
	}
}
