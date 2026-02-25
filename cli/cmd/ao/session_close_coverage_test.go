package main

import (
	"os"
	"strings"
	"testing"

	"github.com/boshu2/agentops/cli/internal/types"
)

// ---------------------------------------------------------------------------
// session_close.go — computeVelocityDelta
// ---------------------------------------------------------------------------

func TestCov3_sessionClose_computeVelocityDelta(t *testing.T) {
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
}

// ---------------------------------------------------------------------------
// session_close.go — classifyFlywheelStatus
// ---------------------------------------------------------------------------

func TestCov3_sessionClose_classifyFlywheelStatus(t *testing.T) {
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

func TestCov3_sessionClose_printCloseTable(t *testing.T) {
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

func TestCov3_sessionClose_shortenPath(t *testing.T) {
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
