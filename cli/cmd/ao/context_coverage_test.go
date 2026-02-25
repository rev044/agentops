package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	contextbudget "github.com/boshu2/agentops/cli/internal/context"
)

// --- resolveGuardSessionID ---

func TestContextCov_ResolveGuardSessionID(t *testing.T) {
	// Save and restore package-level var
	oldVal := contextSessionID
	defer func() { contextSessionID = oldVal }()

	t.Run("returns flag value when set", func(t *testing.T) {
		contextSessionID = "flag-session-123"
		t.Setenv("CLAUDE_SESSION_ID", "env-session-456")
		got, err := resolveGuardSessionID()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "flag-session-123" {
			t.Errorf("got %q, want %q", got, "flag-session-123")
		}
	})

	t.Run("falls back to env var when flag empty", func(t *testing.T) {
		contextSessionID = ""
		t.Setenv("CLAUDE_SESSION_ID", "env-session-456")
		got, err := resolveGuardSessionID()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "env-session-456" {
			t.Errorf("got %q, want %q", got, "env-session-456")
		}
	})

	t.Run("returns error when both empty", func(t *testing.T) {
		contextSessionID = ""
		t.Setenv("CLAUDE_SESSION_ID", "")
		_, err := resolveGuardSessionID()
		if err == nil {
			t.Fatal("expected error when both flag and env are empty")
		}
		if !strings.Contains(err.Error(), "session id missing") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("trims whitespace from flag", func(t *testing.T) {
		contextSessionID = "  trimmed-session  "
		got, err := resolveGuardSessionID()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "trimmed-session" {
			t.Errorf("got %q, want %q", got, "trimmed-session")
		}
	})
}

// --- resolveGuardOptions ---

func TestContextCov_ResolveGuardOptions(t *testing.T) {
	oldMaxTokens := contextMaxTokens
	oldWatchdog := contextWatchdogMinute
	oldAgentName := contextAgentName
	defer func() {
		contextMaxTokens = oldMaxTokens
		contextWatchdogMinute = oldWatchdog
		contextAgentName = oldAgentName
	}()

	t.Run("uses flag values when set", func(t *testing.T) {
		contextMaxTokens = 300000
		contextWatchdogMinute = 30
		contextAgentName = "my-agent"
		t.Setenv("CLAUDE_AGENT_NAME", "env-agent")

		maxTok, watchdog, name := resolveGuardOptions()
		if maxTok != 300000 {
			t.Errorf("maxTokens = %d, want 300000", maxTok)
		}
		if watchdog != 30*time.Minute {
			t.Errorf("watchdog = %v, want 30m", watchdog)
		}
		if name != "my-agent" {
			t.Errorf("agentName = %q, want my-agent", name)
		}
	})

	t.Run("falls back to defaults for zero values", func(t *testing.T) {
		contextMaxTokens = 0
		contextWatchdogMinute = 0
		contextAgentName = ""
		t.Setenv("CLAUDE_AGENT_NAME", "env-agent")

		maxTok, watchdog, name := resolveGuardOptions()
		if maxTok != contextbudget.DefaultMaxTokens {
			t.Errorf("maxTokens = %d, want %d", maxTok, contextbudget.DefaultMaxTokens)
		}
		if watchdog != defaultWatchdogMinutes*time.Minute {
			t.Errorf("watchdog = %v, want %v", watchdog, defaultWatchdogMinutes*time.Minute)
		}
		if name != "env-agent" {
			t.Errorf("agentName = %q, want env-agent", name)
		}
	})

	t.Run("negative maxTokens uses default", func(t *testing.T) {
		contextMaxTokens = -100
		contextWatchdogMinute = -5
		contextAgentName = ""
		t.Setenv("CLAUDE_AGENT_NAME", "")

		maxTok, watchdog, name := resolveGuardOptions()
		if maxTok != contextbudget.DefaultMaxTokens {
			t.Errorf("maxTokens = %d, want default", maxTok)
		}
		if watchdog != defaultWatchdogMinutes*time.Minute {
			t.Errorf("watchdog = %v, want default", watchdog)
		}
		if name != "" {
			t.Errorf("agentName = %q, want empty", name)
		}
	})
}

// --- outputGuardResult ---

func TestContextCov_OutputGuardResult(t *testing.T) {
	oldOutput := output
	defer func() { output = oldOutput }()

	t.Run("json mode encodes to stdout", func(t *testing.T) {
		output = "json"
		result := contextGuardResult{
			Session: contextSessionStatus{
				SessionID: "test-json-out",
				Status:    "OPTIMAL",
			},
		}
		// Just verify it doesn't error — stdout capture is complex in tests
		err := outputGuardResult(result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("table mode prints text", func(t *testing.T) {
		output = "table"
		result := contextGuardResult{
			Session: contextSessionStatus{
				SessionID:    "test-table-out",
				Status:       "WARNING",
				UsagePercent: 0.65,
				Action:       "checkpoint_and_prepare_handoff",
			},
			HandoffFile: "some/handoff.md",
			HookMessage: "Context is WARNING",
		}
		err := outputGuardResult(result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// --- applyHandoffIfCritical ---

func TestContextCov_ApplyHandoffIfCritical(t *testing.T) {
	oldWriteHandoff := contextWriteHandoff
	defer func() { contextWriteHandoff = oldWriteHandoff }()

	t.Run("no-op when write-handoff flag is false", func(t *testing.T) {
		contextWriteHandoff = false
		cwd := t.TempDir()
		status := contextSessionStatus{
			Status: string(contextbudget.StatusCritical),
		}
		usage := transcriptUsage{}
		result := &contextGuardResult{}

		err := applyHandoffIfCritical(cwd, status, usage, result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.HandoffFile != "" {
			t.Errorf("expected empty handoff file, got %q", result.HandoffFile)
		}
	})

	t.Run("no-op when status is not critical", func(t *testing.T) {
		contextWriteHandoff = true
		cwd := t.TempDir()
		status := contextSessionStatus{
			Status: string(contextbudget.StatusWarning),
		}
		usage := transcriptUsage{}
		result := &contextGuardResult{}

		err := applyHandoffIfCritical(cwd, status, usage, result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.HandoffFile != "" {
			t.Errorf("expected empty handoff file, got %q", result.HandoffFile)
		}
	})

	t.Run("writes handoff when critical and flag set", func(t *testing.T) {
		contextWriteHandoff = true
		cwd := t.TempDir()
		status := contextSessionStatus{
			SessionID:      "critical-handoff-session",
			Status:         string(contextbudget.StatusCritical),
			UsagePercent:   0.92,
			EstimatedUsage: 184000,
			MaxTokens:      200000,
			Action:         "handoff_now",
			LastTask:       "test task",
		}
		usage := transcriptUsage{
			InputTokens:             1000,
			CacheCreationInputToken: 83000,
			CacheReadInputToken:     100000,
			Model:                   "claude-opus",
			Timestamp:               time.Now().UTC(),
		}
		result := &contextGuardResult{
			HookMessage: "initial message",
		}

		err := applyHandoffIfCritical(cwd, status, usage, result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.HandoffFile == "" {
			t.Fatal("expected non-empty handoff file")
		}
		if result.PendingMarker == "" {
			t.Fatal("expected non-empty pending marker")
		}
		if !strings.Contains(result.HookMessage, "Handoff saved to") {
			t.Errorf("hook message should contain handoff path, got %q", result.HookMessage)
		}
	})
}

// --- inferAgentRole ---

func TestContextCov_InferAgentRole(t *testing.T) {
	tests := []struct {
		name         string
		agentName    string
		explicitRole string
		wantRole     string
	}{
		{name: "explicit role wins", agentName: "worker-1", explicitRole: "custom-role", wantRole: "custom-role"},
		{name: "empty name returns empty", agentName: "", explicitRole: "", wantRole: ""},
		{name: "admiral is team-lead", agentName: "fleet-admiral-01", explicitRole: "", wantRole: "team-lead"},
		{name: "captain is team-lead", agentName: "captain-kirk", explicitRole: "", wantRole: "team-lead"},
		{name: "coordinator is team-lead", agentName: "build-coordinator", explicitRole: "", wantRole: "team-lead"},
		{name: "orchestrator is team-lead", agentName: "test-orchestrator", explicitRole: "", wantRole: "team-lead"},
		{name: "quarterback is team-lead", agentName: "quarterback-7", explicitRole: "", wantRole: "team-lead"},
		{name: "mayor is team-lead", agentName: "mayor-of-town", explicitRole: "", wantRole: "team-lead"},
		{name: "leader is team-lead", agentName: "team-leader", explicitRole: "", wantRole: "team-lead"},
		{name: "lead is team-lead", agentName: "tech-lead", explicitRole: "", wantRole: "team-lead"},
		{name: "red-cell is review", agentName: "red-cell-alpha", explicitRole: "", wantRole: "review"},
		{name: "navigator is review", agentName: "navigator-01", explicitRole: "", wantRole: "review"},
		{name: "judge is review", agentName: "judge-dredd", explicitRole: "", wantRole: "review"},
		{name: "reviewer is review", agentName: "code-reviewer", explicitRole: "", wantRole: "review"},
		{name: "worker is worker", agentName: "worker-42", explicitRole: "", wantRole: "worker"},
		{name: "crew is worker", agentName: "crew-member", explicitRole: "", wantRole: "worker"},
		{name: "mate is worker", agentName: "first-mate", explicitRole: "", wantRole: "worker"},
		{name: "unknown is agent", agentName: "random-name", explicitRole: "", wantRole: "agent"},
		{name: "whitespace-only explicit returns empty for empty name", agentName: "", explicitRole: "  ", wantRole: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferAgentRole(tt.agentName, tt.explicitRole)
			if got != tt.wantRole {
				t.Errorf("inferAgentRole(%q, %q) = %q, want %q", tt.agentName, tt.explicitRole, got, tt.wantRole)
			}
		})
	}
}

// --- remainingPercent ---

func TestContextCov_RemainingPercent(t *testing.T) {
	tests := []struct {
		name    string
		usage   float64
		want    float64
		wantMin float64
		wantMax float64
	}{
		{name: "zero usage", usage: 0, wantMin: 1.0, wantMax: 1.0},
		{name: "half usage", usage: 0.5, wantMin: 0.5, wantMax: 0.5},
		{name: "full usage", usage: 1.0, wantMin: 0.0, wantMax: 0.0},
		{name: "over 100% usage clamps to 0", usage: 1.5, wantMin: 0.0, wantMax: 0.0},
		{name: "negative usage clamps to 1", usage: -0.5, wantMin: 1.0, wantMax: 1.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := remainingPercent(tt.usage)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("remainingPercent(%f) = %f, want in [%f, %f]", tt.usage, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// --- readinessRank ---

func TestContextCov_ReadinessRank(t *testing.T) {
	tests := []struct {
		readiness string
		wantRank  int
	}{
		{contextReadinessCritical, 0},
		{contextReadinessRed, 1},
		{contextReadinessAmber, 2},
		{contextReadinessGreen, 3},
		{"UNKNOWN", 4},
		{"", 4},
		{"  CRITICAL  ", 0},
	}
	for _, tt := range tests {
		t.Run(tt.readiness, func(t *testing.T) {
			got := readinessRank(tt.readiness)
			if got != tt.wantRank {
				t.Errorf("readinessRank(%q) = %d, want %d", tt.readiness, got, tt.wantRank)
			}
		})
	}
}

// --- tmuxTargetFromPaneID ---

func TestContextCov_TmuxTargetFromPaneID(t *testing.T) {
	tests := []struct {
		name   string
		paneID string
		want   string
	}{
		{name: "empty returns empty", paneID: "", want: ""},
		{name: "in-process returns empty", paneID: "in-process", want: ""},
		{name: "session:window.pane strips pane", paneID: "my-session:0.2", want: "my-session:0"},
		{name: "no dot returns as-is", paneID: "my-session:0", want: "my-session:0"},
		{name: "whitespace only returns empty", paneID: "   ", want: ""},
		{name: "multiple dots uses last", paneID: "a.b.c", want: "a.b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tmuxTargetFromPaneID(tt.paneID)
			if got != tt.want {
				t.Errorf("tmuxTargetFromPaneID(%q) = %q, want %q", tt.paneID, got, tt.want)
			}
		})
	}
}

// --- tmuxSessionFromTarget ---

func TestContextCov_TmuxSessionFromTarget(t *testing.T) {
	tests := []struct {
		name   string
		target string
		want   string
	}{
		{name: "empty returns empty", target: "", want: ""},
		{name: "session:window extracts session", target: "my-session:0", want: "my-session"},
		{name: "no colon returns as-is", target: "my-session", want: "my-session"},
		{name: "whitespace trimmed", target: "  my-session:0  ", want: "my-session"},
		{name: "multiple colons uses first", target: "a:b:c", want: "a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tmuxSessionFromTarget(tt.target)
			if got != tt.want {
				t.Errorf("tmuxSessionFromTarget(%q) = %q, want %q", tt.target, got, tt.want)
			}
		})
	}
}

// --- toRepoRelative ---

func TestContextCov_ToRepoRelative(t *testing.T) {
	tests := []struct {
		name     string
		cwd      string
		fullPath string
		want     string
	}{
		{name: "empty path returns empty", cwd: "/a/b", fullPath: "", want: ""},
		{name: "child path returns relative", cwd: "/a/b", fullPath: "/a/b/c/d.txt", want: "c/d.txt"},
		{name: "same dir returns dot file", cwd: "/a/b", fullPath: "/a/b/file.txt", want: "file.txt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toRepoRelative(tt.cwd, tt.fullPath)
			if got != tt.want {
				t.Errorf("toRepoRelative(%q, %q) = %q, want %q", tt.cwd, tt.fullPath, got, tt.want)
			}
		})
	}
}

// --- contextWithTimeout ---

func TestContextCov_ContextWithTimeout(t *testing.T) {
	t.Run("positive timeout creates context with deadline", func(t *testing.T) {
		ctx, cancel := contextWithTimeout(5 * time.Second)
		defer cancel()
		deadline, ok := ctx.Deadline()
		if !ok {
			t.Fatal("expected deadline to be set")
		}
		if time.Until(deadline) <= 0 {
			t.Fatal("expected future deadline")
		}
	})

	t.Run("zero timeout creates cancel-only context", func(t *testing.T) {
		ctx, cancel := contextWithTimeout(0)
		defer cancel()
		_, ok := ctx.Deadline()
		if ok {
			t.Fatal("expected no deadline for zero timeout")
		}
	})

	t.Run("negative timeout creates cancel-only context", func(t *testing.T) {
		ctx, cancel := contextWithTimeout(-1 * time.Second)
		defer cancel()
		_, ok := ctx.Deadline()
		if ok {
			t.Fatal("expected no deadline for negative timeout")
		}
	})
}

// --- truncateDisplay ---

func TestContextCov_TruncateDisplay(t *testing.T) {
	tests := []struct {
		name string
		s    string
		max  int
		want string
	}{
		{name: "short string unchanged", s: "abc", max: 10, want: "abc"},
		{name: "exact length unchanged", s: "abcde", max: 5, want: "abcde"},
		{name: "long string truncated with ellipsis", s: "abcdefghij", max: 7, want: "abcd..."},
		{name: "max 3 no ellipsis", s: "abcdefgh", max: 3, want: "abc"},
		{name: "max 2 no ellipsis", s: "abcdefgh", max: 2, want: "ab"},
		{name: "empty string", s: "", max: 5, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateDisplay(tt.s, tt.max)
			if got != tt.want {
				t.Errorf("truncateDisplay(%q, %d) = %q, want %q", tt.s, tt.max, got, tt.want)
			}
		})
	}
}

// --- estimateTokens ---

func TestContextCov_EstimateTokens(t *testing.T) {
	tests := []struct {
		name string
		text string
		want int
	}{
		{name: "empty string", text: "", want: 0},
		{name: "whitespace only", text: "   ", want: 0},
		{name: "short text returns 1", text: "ab", want: 1},
		{name: "exactly 4 chars", text: "abcd", want: 1},
		{name: "longer text", text: "this is a longer text for token estimation test", want: len("this is a longer text for token estimation test") / 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := estimateTokens(tt.text)
			if got != tt.want {
				t.Errorf("estimateTokens(%q) = %d, want %d", tt.text, got, tt.want)
			}
		})
	}
}

// --- actionForStatus ---

func TestContextCov_ActionForStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		stale  bool
		want   string
	}{
		{name: "critical", status: string(contextbudget.StatusCritical), stale: false, want: "handoff_now"},
		{name: "warning", status: string(contextbudget.StatusWarning), stale: false, want: "checkpoint_and_prepare_handoff"},
		{name: "optimal", status: string(contextbudget.StatusOptimal), stale: false, want: "continue"},
		{name: "stale critical recovers", status: string(contextbudget.StatusCritical), stale: true, want: "recover_dead_session"},
		{name: "stale warning recovers", status: string(contextbudget.StatusWarning), stale: true, want: "recover_dead_session"},
		{name: "stale optimal investigates", status: string(contextbudget.StatusOptimal), stale: true, want: "investigate_stale_session"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := actionForStatus(tt.status, tt.stale)
			if got != tt.want {
				t.Errorf("actionForStatus(%q, %v) = %q, want %q", tt.status, tt.stale, got, tt.want)
			}
		})
	}
}

// --- hookMessageForStatus ---

func TestContextCov_HookMessageForStatus(t *testing.T) {
	tests := []struct {
		name   string
		status contextSessionStatus
		want   string // substring to find in result
		empty  bool   // if true, expect empty string
	}{
		{
			name: "handoff_now action",
			status: contextSessionStatus{
				Action:           "handoff_now",
				UsagePercent:     0.92,
				Readiness:        contextReadinessCritical,
				RemainingPercent: 0.08,
			},
			want: "CRITICAL",
		},
		{
			name: "checkpoint action",
			status: contextSessionStatus{
				Action:           "checkpoint_and_prepare_handoff",
				UsagePercent:     0.65,
				Readiness:        contextReadinessAmber,
				RemainingPercent: 0.35,
			},
			want: "WARNING",
		},
		{
			name: "recover with restart attempt success",
			status: contextSessionStatus{
				Action:         "recover_dead_session",
				RestartAttempt: true,
				RestartSuccess: true,
				TmuxSession:   "my-session",
			},
			want: "auto-restarted",
		},
		{
			name: "recover with restart attempt failure",
			status: contextSessionStatus{
				Action:         "recover_dead_session",
				RestartAttempt: true,
				RestartSuccess: false,
				RestartMessage: "tmux unavailable",
			},
			want: "auto-restart failed",
		},
		{
			name: "recover without restart attempt but with message",
			status: contextSessionStatus{
				Action:         "recover_dead_session",
				RestartAttempt: false,
				RestartMessage: "missing tmux target",
			},
			want: "stale with unfinished work",
		},
		{
			name: "recover without restart attempt and no message",
			status: contextSessionStatus{
				Action: "recover_dead_session",
			},
			want: "stale with unfinished work",
		},
		{
			name: "continue with RED readiness",
			status: contextSessionStatus{
				Action:           "continue",
				Readiness:        contextReadinessRed,
				RemainingPercent: 0.45,
			},
			want: "Hull is RED",
		},
		{
			name:  "continue with GREEN readiness returns empty",
			status: contextSessionStatus{Action: "continue", Readiness: contextReadinessGreen},
			empty: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hookMessageForStatus(tt.status)
			if tt.empty {
				if got != "" {
					t.Errorf("expected empty, got %q", got)
				}
				return
			}
			if !strings.Contains(got, tt.want) {
				t.Errorf("hookMessageForStatus() = %q, want substring %q", got, tt.want)
			}
		})
	}
}

// --- readPersistedAssignment ---

func TestContextCov_ReadPersistedAssignment(t *testing.T) {
	t.Run("returns false for missing file", func(t *testing.T) {
		dir := t.TempDir()
		_, ok := readPersistedAssignment(dir, "nonexistent-session")
		if ok {
			t.Error("expected false for missing file")
		}
	})

	t.Run("returns false for invalid JSON", func(t *testing.T) {
		dir := t.TempDir()
		contextDir := filepath.Join(dir, ".agents", "ao", "context")
		if err := os.MkdirAll(contextDir, 0755); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(contextDir, "assignment-bad-json.json")
		if err := os.WriteFile(path, []byte("{invalid}"), 0644); err != nil {
			t.Fatal(err)
		}
		_, ok := readPersistedAssignment(dir, "bad-json")
		if ok {
			t.Error("expected false for invalid JSON")
		}
	})

	t.Run("returns false for empty assignment", func(t *testing.T) {
		dir := t.TempDir()
		contextDir := filepath.Join(dir, ".agents", "ao", "context")
		if err := os.MkdirAll(contextDir, 0755); err != nil {
			t.Fatal(err)
		}
		snapshot := contextAssignmentSnapshot{SessionID: "empty-assign"}
		data, _ := json.MarshalIndent(snapshot, "", "  ")
		path := filepath.Join(contextDir, "assignment-empty-assign.json")
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
		_, ok := readPersistedAssignment(dir, "empty-assign")
		if ok {
			t.Error("expected false for empty assignment fields")
		}
	})

	t.Run("returns assignment for valid file", func(t *testing.T) {
		dir := t.TempDir()
		contextDir := filepath.Join(dir, ".agents", "ao", "context")
		if err := os.MkdirAll(contextDir, 0755); err != nil {
			t.Fatal(err)
		}
		snapshot := contextAssignmentSnapshot{
			SessionID: "valid-session",
			AgentName: "worker-7",
			IssueID:   "ag-123",
			TeamName:  "alpha",
		}
		data, _ := json.MarshalIndent(snapshot, "", "  ")
		path := filepath.Join(contextDir, "assignment-valid-session.json")
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
		assignment, ok := readPersistedAssignment(dir, "valid-session")
		if !ok {
			t.Fatal("expected true for valid file")
		}
		if assignment.AgentName != "worker-7" {
			t.Errorf("AgentName = %q, want worker-7", assignment.AgentName)
		}
		if assignment.IssueID != "ag-123" {
			t.Errorf("IssueID = %q, want ag-123", assignment.IssueID)
		}
		if assignment.TeamName != "alpha" {
			t.Errorf("TeamName = %q, want alpha", assignment.TeamName)
		}
	})
}

// --- mergePersistedAssignment ---

func TestContextCov_MergePersistedAssignment(t *testing.T) {
	t.Run("nil status is safe", func(t *testing.T) {
		dir := t.TempDir()
		mergePersistedAssignment(dir, nil) // should not panic
	})

	t.Run("empty session ID is safe", func(t *testing.T) {
		dir := t.TempDir()
		status := &contextSessionStatus{}
		mergePersistedAssignment(dir, status) // should not panic
	})

	t.Run("merges persisted fields into status", func(t *testing.T) {
		dir := t.TempDir()
		contextDir := filepath.Join(dir, ".agents", "ao", "context")
		if err := os.MkdirAll(contextDir, 0755); err != nil {
			t.Fatal(err)
		}
		snapshot := contextAssignmentSnapshot{
			SessionID: "merge-test",
			AgentName: "persisted-agent",
			TeamName:  "persisted-team",
		}
		data, _ := json.MarshalIndent(snapshot, "", "  ")
		if err := os.WriteFile(filepath.Join(contextDir, "assignment-merge-test.json"), data, 0644); err != nil {
			t.Fatal(err)
		}

		status := &contextSessionStatus{SessionID: "merge-test"}
		mergePersistedAssignment(dir, status)
		if status.AgentName != "persisted-agent" {
			t.Errorf("AgentName = %q, want persisted-agent", status.AgentName)
		}
		if status.TeamName != "persisted-team" {
			t.Errorf("TeamName = %q, want persisted-team", status.TeamName)
		}
	})
}

// --- findTeamMemberByName ---

func TestContextCov_FindTeamMemberByName(t *testing.T) {
	t.Run("empty name returns false", func(t *testing.T) {
		_, _, ok := findTeamMemberByName("")
		if ok {
			t.Error("expected false for empty name")
		}
	})

	t.Run("empty HOME returns false", func(t *testing.T) {
		t.Setenv("HOME", "")
		_, _, ok := findTeamMemberByName("worker-1")
		if ok {
			t.Error("expected false for empty HOME")
		}
	})

	t.Run("no teams dir returns false", func(t *testing.T) {
		tmpHome := t.TempDir()
		t.Setenv("HOME", tmpHome)
		_, _, ok := findTeamMemberByName("worker-1")
		if ok {
			t.Error("expected false when no teams dir exists")
		}
	})

	t.Run("finds member across multiple teams", func(t *testing.T) {
		tmpHome := t.TempDir()
		t.Setenv("HOME", tmpHome)

		// Create two team dirs
		team1Dir := filepath.Join(tmpHome, ".claude", "teams", "alpha-team")
		team2Dir := filepath.Join(tmpHome, ".claude", "teams", "beta-team")
		if err := os.MkdirAll(team1Dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(team2Dir, 0755); err != nil {
			t.Fatal(err)
		}

		// Write configs
		cfg1 := `{"members":[{"name":"alice","agentType":"lead","tmuxPaneId":"sess:0.1"}]}`
		cfg2 := `{"members":[{"name":"bob","agentType":"worker","tmuxPaneId":"sess:0.2"}]}`
		if err := os.WriteFile(filepath.Join(team1Dir, "config.json"), []byte(cfg1), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(team2Dir, "config.json"), []byte(cfg2), 0644); err != nil {
			t.Fatal(err)
		}

		teamName, member, ok := findTeamMemberByName("bob")
		if !ok {
			t.Fatal("expected to find bob")
		}
		if teamName != "beta-team" {
			t.Errorf("teamName = %q, want beta-team", teamName)
		}
		if member.AgentType != "worker" {
			t.Errorf("agentType = %q, want worker", member.AgentType)
		}
	})

	t.Run("skips non-directory entries in teams dir", func(t *testing.T) {
		tmpHome := t.TempDir()
		t.Setenv("HOME", tmpHome)

		teamsDir := filepath.Join(tmpHome, ".claude", "teams")
		if err := os.MkdirAll(teamsDir, 0755); err != nil {
			t.Fatal(err)
		}
		// Create a regular file (not a directory)
		if err := os.WriteFile(filepath.Join(teamsDir, "not-a-team.txt"), []byte("junk"), 0644); err != nil {
			t.Fatal(err)
		}

		_, _, ok := findTeamMemberByName("anyone")
		if ok {
			t.Error("expected false when only non-dir entries exist")
		}
	})
}

// --- maybeAutoRestartStaleSession additional cases ---

func TestContextCov_MaybeAutoRestartStaleSession_EdgeCases(t *testing.T) {
	t.Run("non-recover action returns unchanged", func(t *testing.T) {
		status := contextSessionStatus{
			Action:     "continue",
			TmuxTarget: "some:target",
		}
		result := maybeAutoRestartStaleSession(status)
		if result.RestartAttempt {
			t.Error("should not attempt restart for non-recover action")
		}
	})

	t.Run("empty tmux target sets message", func(t *testing.T) {
		status := contextSessionStatus{
			Action:     "recover_dead_session",
			TmuxTarget: "",
		}
		result := maybeAutoRestartStaleSession(status)
		if result.RestartMessage != "missing tmux target mapping" {
			t.Errorf("message = %q, want 'missing tmux target mapping'", result.RestartMessage)
		}
	})

	t.Run("empty session name from target sets message", func(t *testing.T) {
		// tmuxSessionFromTarget("") returns "" and tmuxSessionFromTarget(":0") returns ""
		// because idx=0 is not > 0
		status := contextSessionStatus{
			Action:     "recover_dead_session",
			TmuxTarget: ":0",
		}
		// tmux may not be available, so check what branch we land in
		result := maybeAutoRestartStaleSession(status)
		// Either tmux unavailable or invalid target
		if result.RestartMessage == "" {
			t.Error("expected non-empty restart message for edge case target")
		}
	})
}

// --- parseTimestamp additional coverage ---

func TestContextCov_ParseTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantZero bool
	}{
		{name: "empty string", input: "", wantZero: true},
		{name: "whitespace only", input: "   ", wantZero: true},
		{name: "RFC3339Nano", input: "2026-02-20T10:00:00.123456789Z", wantZero: false},
		{name: "RFC3339", input: "2026-02-20T10:00:00Z", wantZero: false},
		{name: "invalid format", input: "not-a-timestamp", wantZero: true},
		{name: "date only", input: "2026-02-20", wantZero: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTimestamp(tt.input)
			if tt.wantZero && !got.IsZero() {
				t.Errorf("expected zero time for %q, got %v", tt.input, got)
			}
			if !tt.wantZero && got.IsZero() {
				t.Errorf("expected non-zero time for %q", tt.input)
			}
		})
	}
}

// --- readFileTail ---

func TestContextCov_ReadFileTail(t *testing.T) {
	t.Run("reads empty file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "empty.jsonl")
		if err := os.WriteFile(path, []byte{}, 0644); err != nil {
			t.Fatal(err)
		}
		data, err := readFileTail(path, 1024)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(data) != 0 {
			t.Errorf("expected empty data, got %d bytes", len(data))
		}
	})

	t.Run("returns error for nonexistent file", func(t *testing.T) {
		_, err := readFileTail("/nonexistent/path/file.jsonl", 1024)
		if err == nil {
			t.Fatal("expected error for nonexistent file")
		}
	})
}

// --- collectTrackedSessionStatuses ---

func TestContextCov_CollectTrackedSessionStatuses(t *testing.T) {
	t.Run("returns nil when no budget files", func(t *testing.T) {
		dir := t.TempDir()
		statuses, err := collectTrackedSessionStatuses(dir, 20*time.Minute)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if statuses != nil {
			t.Errorf("expected nil, got %d statuses", len(statuses))
		}
	})

	t.Run("loads budget files and returns statuses", func(t *testing.T) {
		dir := t.TempDir()
		tmpHome := t.TempDir()
		t.Setenv("HOME", tmpHome)

		// Create a budget file
		contextDir := filepath.Join(dir, ".agents", "ao", "context")
		if err := os.MkdirAll(contextDir, 0755); err != nil {
			t.Fatal(err)
		}

		sessionID := "tracked-session-01"
		// Create transcript so collectSessionStatus can find it
		transcriptDir := filepath.Join(tmpHome, ".claude", "projects", "proj", "conversations")
		if err := os.MkdirAll(transcriptDir, 0755); err != nil {
			t.Fatal(err)
		}
		transcriptLines := []map[string]any{
			{
				"type":      "user",
				"timestamp": time.Now().Add(-1 * time.Minute).UTC().Format(time.RFC3339),
				"message": map[string]any{
					"role":    "user",
					"content": "test task",
				},
			},
			{
				"type":      "assistant",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"message": map[string]any{
					"role":  "assistant",
					"model": "claude-sonnet",
					"usage": map[string]any{
						"input_tokens":                1000,
						"cache_creation_input_tokens": 2000,
						"cache_read_input_tokens":     3000,
					},
				},
			},
		}
		var b strings.Builder
		for _, line := range transcriptLines {
			data, _ := json.Marshal(line)
			b.Write(data)
			b.WriteByte('\n')
		}
		if err := os.WriteFile(filepath.Join(transcriptDir, sessionID+".jsonl"), []byte(b.String()), 0644); err != nil {
			t.Fatal(err)
		}

		tracker := contextbudget.NewBudgetTracker(sessionID)
		tracker.MaxTokens = contextbudget.DefaultMaxTokens
		tracker.UpdateUsage(6000)
		data, _ := json.MarshalIndent(tracker, "", "  ")
		if err := os.WriteFile(filepath.Join(contextDir, "budget-"+sessionID+".json"), data, 0644); err != nil {
			t.Fatal(err)
		}

		statuses, err := collectTrackedSessionStatuses(dir, 20*time.Minute)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(statuses) == 0 {
			t.Fatal("expected at least one status")
		}
		if statuses[0].SessionID != sessionID {
			t.Errorf("SessionID = %q, want %q", statuses[0].SessionID, sessionID)
		}
	})
}

// --- renderHandoffMarkdown ---

func TestContextCov_RenderHandoffMarkdown(t *testing.T) {
	now := time.Date(2026, 2, 25, 12, 0, 0, 0, time.UTC)
	status := contextSessionStatus{
		SessionID:        "render-test",
		Status:           "CRITICAL",
		UsagePercent:     0.92,
		RemainingPercent: 0.08,
		Readiness:        contextReadinessCritical,
		Action:           "handoff_now",
		LastTask:         "test rendering",
		AgentName:        "worker-1",
		AgentRole:        "worker",
		TeamName:         "alpha",
		IssueID:          "ag-xyz",
		TmuxTarget:       "sess:0",
		Model:            "claude-opus",
		EstimatedUsage:   184000,
		MaxTokens:        200000,
		Recommendation:   "CRITICAL: Context nearly full.",
		IsStale:          true,
	}
	usage := transcriptUsage{
		InputTokens:             1000,
		CacheCreationInputToken: 83000,
		CacheReadInputToken:     100000,
	}

	md := renderHandoffMarkdown(now, status, usage, "ag-xyz", []string{"file1.go", "file2.go"})

	checks := []string{
		"# Auto-Handoff",
		"render-test",
		"CRITICAL",
		"handoff_now",
		"test rendering",
		"ag-xyz",
		"worker-1",
		"worker",
		"alpha",
		"file1.go",
		"file2.go",
		"stale",
		"claude-opus",
		"184000",
	}
	for _, check := range checks {
		if !strings.Contains(md, check) {
			t.Errorf("handoff markdown missing %q", check)
		}
	}

	// Test with no changed files
	md2 := renderHandoffMarkdown(now, status, usage, "none", nil)
	if !strings.Contains(md2, "none\n") {
		t.Error("expected 'none' for no changed files")
	}

	// Test with non-stale status
	status.IsStale = false
	md3 := renderHandoffMarkdown(now, status, usage, "none", nil)
	if !strings.Contains(md3, "none detected") {
		t.Error("expected 'none detected' for non-stale blockers")
	}
}
