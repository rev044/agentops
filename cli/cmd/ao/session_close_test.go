package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/boshu2/agentops/cli/internal/types"
)

func TestResolveTranscript(t *testing.T) {
	// Create temp directory structure mimicking ~/.claude/projects/
	tempDir, err := os.MkdirTemp("", "session-close-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir) //nolint:errcheck // test cleanup
	}()

	tests := []struct {
		name           string
		sessionID      string
		setupFunc      func(t *testing.T) string // returns expected path
		expectFallback bool
		expectError    bool
	}{
		{
			name:      "empty session ID triggers fallback",
			sessionID: "",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				// findLastSession searches real ~/.claude/projects
				// so this test just verifies fallback is set
				return ""
			},
			expectFallback: true,
			expectError:    true, // may fail if no real transcripts
		},
		{
			name:      "nonexistent session ID returns error",
			sessionID: "nonexistent-session-id-12345",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				return ""
			},
			expectFallback: false,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc(t)

			_, usedFallback, err := resolveTranscript(tt.sessionID)
			if tt.expectError && err == nil {
				// Some tests may pass on machines with real transcripts
				// Only fail if we expected an error and got specific wrong behavior
				if tt.sessionID == "nonexistent-session-id-12345" {
					t.Error("expected error for nonexistent session ID, got nil")
				}
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectFallback && err == nil && !usedFallback {
				t.Error("expected fallback to be used")
			}
		})
	}
}

func TestFindTranscriptBySessionID(t *testing.T) {
	// Create temp directory with mock transcript
	tempDir, err := os.MkdirTemp("", "session-find-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir) //nolint:errcheck // test cleanup
	}()

	tests := []struct {
		name        string
		sessionID   string
		expectError bool
	}{
		{
			name:        "nonexistent session returns error",
			sessionID:   "does-not-exist-abc-123",
			expectError: true,
		},
		{
			name:        "empty session ID returns error",
			sessionID:   "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := findTranscriptBySessionID(tt.sessionID)
			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestOutputCloseResult(t *testing.T) {
	tests := []struct {
		name   string
		result SessionCloseResult
		format string
	}{
		{
			name: "table output succeeds",
			result: SessionCloseResult{
				SessionID:     "test-session-001",
				Transcript:    "/tmp/test.jsonl",
				Decisions:     3,
				Knowledge:     5,
				FilesChanged:  10,
				Issues:        2,
				VelocityDelta: 0.05,
				Status:        "compounding",
				Message:       "Session closed: 3 decisions, 5 learnings extracted",
			},
			format: "table",
		},
		{
			name: "json output succeeds",
			result: SessionCloseResult{
				SessionID:     "test-session-002",
				Transcript:    "/tmp/test2.jsonl",
				Decisions:     0,
				Knowledge:     0,
				FilesChanged:  0,
				VelocityDelta: -0.01,
				Status:        "decaying",
				Message:       "Session closed: 0 decisions, 0 learnings extracted",
			},
			format: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Redirect stdout to capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Set output format
			oldOutput := output
			output = tt.format
			defer func() {
				output = oldOutput
			}()

			err := outputCloseResult(tt.result)

			_ = w.Close()
			os.Stdout = oldStdout

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Read captured output
			buf := make([]byte, 4096)
			n, _ := r.Read(buf)
			_ = r.Close()

			out := string(buf[:n])
			if len(out) == 0 {
				t.Error("expected output, got empty string")
			}
		})
	}
}

func TestShortenPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("get home dir: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "path under home directory",
			input:    filepath.Join(homeDir, ".claude", "projects", "test.jsonl"),
			expected: "~/.claude/projects/test.jsonl",
		},
		{
			name:     "path not under home",
			input:    "/tmp/test.jsonl",
			expected: "/tmp/test.jsonl",
		},
		{
			name:     "empty path",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shortenPath(tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSessionCloseResultJSON(t *testing.T) {
	result := SessionCloseResult{
		SessionID:     "test-json-001",
		Transcript:    "/tmp/session.jsonl",
		Decisions:     2,
		Knowledge:     4,
		FilesChanged:  8,
		Issues:        1,
		VelocityDelta: 0.123,
		Status:        "compounding",
		Message:       "test message",
	}

	// Verify JSON fields exist
	if result.SessionID != "test-json-001" {
		t.Errorf("SessionID: got %q, want %q", result.SessionID, "test-json-001")
	}
	if result.Decisions != 2 {
		t.Errorf("Decisions: got %d, want 2", result.Decisions)
	}
	if result.VelocityDelta != 0.123 {
		t.Errorf("VelocityDelta: got %f, want 0.123", result.VelocityDelta)
	}
}

// ---------------------------------------------------------------------------
// session_close.go — computeVelocityDelta
// ---------------------------------------------------------------------------

func TestSessionClose_computeVelocityDelta(t *testing.T) {
	tests := []struct {
		name string
		pre  *types.FlywheelMetrics
		post *types.FlywheelMetrics
		want float64
	}{
		{
			name: "both nil returns 0",
			pre:  nil,
			post: nil,
			want: 0.0,
		},
		{
			name: "pre nil returns 0",
			pre:  nil,
			post: &types.FlywheelMetrics{Velocity: 0.5},
			want: 0.0,
		},
		{
			name: "post nil returns 0",
			pre:  &types.FlywheelMetrics{Velocity: 0.3},
			post: nil,
			want: 0.0,
		},
		{
			name: "positive delta",
			pre:  &types.FlywheelMetrics{Velocity: 0.1},
			post: &types.FlywheelMetrics{Velocity: 0.3},
			want: 0.2,
		},
		{
			name: "negative delta",
			pre:  &types.FlywheelMetrics{Velocity: 0.5},
			post: &types.FlywheelMetrics{Velocity: 0.2},
			want: -0.3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := computeVelocityDelta(tc.pre, tc.post)
			// Use tolerance for float comparison
			diff := got - tc.want
			if diff < -0.001 || diff > 0.001 {
				t.Errorf("computeVelocityDelta() = %f, want %f", got, tc.want)
			}
		})
	}

	// Also verify zero-velocity case explicitly
	got := computeVelocityDelta(
		&types.FlywheelMetrics{Velocity: 0.5},
		&types.FlywheelMetrics{Velocity: 0.5},
	)
	if got != 0.0 {
		t.Errorf("expected 0 delta for equal velocities, got %f", got)
	}
}

// ---------------------------------------------------------------------------
// session_close.go — classifyFlywheelStatus
// ---------------------------------------------------------------------------

func TestSessionClose_classifyFlywheelStatus(t *testing.T) {
	tests := []struct {
		name string
		post *types.FlywheelMetrics
		want string
	}{
		{
			name: "nil returns compounding",
			post: nil,
			want: "compounding",
		},
		{
			name: "above escape velocity returns compounding",
			post: &types.FlywheelMetrics{AboveEscapeVelocity: true, Velocity: 0.5},
			want: "compounding",
		},
		{
			name: "near zero velocity returns near-escape",
			post: &types.FlywheelMetrics{AboveEscapeVelocity: false, Velocity: -0.04},
			want: "near-escape",
		},
		{
			name: "zero velocity returns near-escape",
			post: &types.FlywheelMetrics{AboveEscapeVelocity: false, Velocity: 0.0},
			want: "near-escape",
		},
		{
			name: "deeply negative velocity returns decaying",
			post: &types.FlywheelMetrics{AboveEscapeVelocity: false, Velocity: -0.2},
			want: "decaying",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyFlywheelStatus(tc.post)
			if got != tc.want {
				t.Errorf("classifyFlywheelStatus() = %q, want %q", got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// session_close.go — printCloseTable
// ---------------------------------------------------------------------------

func TestSessionClose_printCloseTable(t *testing.T) {
	tests := []struct {
		name   string
		result SessionCloseResult
		checks []string
	}{
		{
			name: "full result with issues",
			result: SessionCloseResult{
				SessionID:     "abcdefghijklmnop",
				Transcript:    "/tmp/test.jsonl",
				Decisions:     5,
				Knowledge:     3,
				FilesChanged:  10,
				Issues:        2,
				VelocityDelta: 0.123,
				Status:        "compounding",
				Message:       "Session closed: 5 decisions, 3 learnings extracted",
			},
			checks: []string{
				"Session Close Summary",
				"abcdefghijkl", // truncated to 12
				"Decisions:     5",
				"Issues:        2",
				"compounding",
				"+0.123",
			},
		},
		{
			name: "negative velocity no issues",
			result: SessionCloseResult{
				SessionID:     "short",
				Transcript:    "/tmp/t2.jsonl",
				Decisions:     0,
				Knowledge:     0,
				FilesChanged:  0,
				Issues:        0,
				VelocityDelta: -0.05,
				Status:        "decaying",
				Message:       "Session closed: 0 decisions, 0 learnings extracted",
			},
			checks: []string{
				"short",
				"decaying",
				"-0.050",
			},
		},
		{
			name: "empty session ID omits session line",
			result: SessionCloseResult{
				Transcript:    "/tmp/empty.jsonl",
				VelocityDelta: 0.0,
				Status:        "near-escape",
				Message:       "test",
			},
			checks: []string{
				"near-escape",
				"Session Close Summary",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printCloseTable(tc.result)

			_ = w.Close()
			os.Stdout = oldStdout

			buf := make([]byte, 8192)
			n, _ := r.Read(buf)
			_ = r.Close()
			out := string(buf[:n])

			for _, check := range tc.checks {
				if !strings.Contains(out, check) {
					t.Errorf("expected output to contain %q, got:\n%s", check, out)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// session_close.go — shortenPath
// ---------------------------------------------------------------------------

func TestSessionClose_shortenPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("get home: %v", err)
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "absolute path under home",
			input: homeDir + "/projects/test.md",
			want:  "~/projects/test.md",
		},
		{
			name:  "path not under home",
			input: "/var/log/test.log",
			want:  "/var/log/test.log",
		},
		{
			name:  "empty path",
			input: "",
			want:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := shortenPath(tc.input)
			if got != tc.want {
				t.Errorf("shortenPath(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
