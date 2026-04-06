package rpi

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// ProcessInfo holds parsed process metadata from the process table.
type ProcessInfo struct {
	PID     int
	PPID    int
	Command string
}

// CancelTarget describes one active RPI run to be cancelled.
type CancelTarget struct {
	Kind         string
	RunID        string
	Root         string
	StatePath    string
	LeasePath    string
	WorktreePath string
	PIDs         []int
}

// SupervisorLeaseMetadata is the JSON structure stored in supervisor.lock.
type SupervisorLeaseMetadata struct {
	RunID      string `json:"run_id"`
	PID        int    `json:"pid"`
	Host       string `json:"host"`
	CWD        string `json:"cwd"`
	AcquiredAt string `json:"acquired_at"`
	RenewedAt  string `json:"renewed_at"`
	ExpiresAt  string `json:"expires_at"`
}

// ParseCancelSignal converts a user-facing signal name to a syscall.Signal.
func ParseCancelSignal(raw string) (syscall.Signal, error) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "", "TERM", "SIGTERM":
		return syscall.SIGTERM, nil
	case "KILL", "SIGKILL":
		return syscall.SIGKILL, nil
	case "INT", "SIGINT":
		return syscall.SIGINT, nil
	default:
		return 0, fmt.Errorf("unsupported signal %q (valid: TERM|KILL|INT)", raw)
	}
}

// FilterKillablePIDs removes PIDs <= 1 and the caller's own PID, returning a sorted slice.
func FilterKillablePIDs(pids []int, selfPID int) []int {
	var out []int
	for _, pid := range DedupeInts(pids) {
		if pid <= 1 || pid == selfPID {
			continue
		}
		out = append(out, pid)
	}
	sort.Ints(out)
	return out
}

// DedupeInts returns a sorted, deduplicated copy of in.
func DedupeInts(in []int) []int {
	set := make(map[int]struct{}, len(in))
	var out []int
	for _, n := range in {
		if _, ok := set[n]; ok {
			continue
		}
		set[n] = struct{}{}
		out = append(out, n)
	}
	sort.Ints(out)
	return out
}

// DescendantPIDs returns all transitive child PIDs of parentPID via BFS.
func DescendantPIDs(parentPID int, procs []ProcessInfo) []int {
	children := make(map[int][]int)
	for _, proc := range procs {
		children[proc.PPID] = append(children[proc.PPID], proc.PID)
	}

	var out []int
	queue := []int{parentPID}
	seen := map[int]struct{}{parentPID: {}}

	for len(queue) > 0 {
		pid := queue[0]
		queue = queue[1:]
		for _, child := range children[pid] {
			if _, ok := seen[child]; ok {
				continue
			}
			seen[child] = struct{}{}
			out = append(out, child)
			queue = append(queue, child)
		}
	}
	sort.Ints(out)
	return out
}

// ProcessExists returns true if a PID appears in the process table.
func ProcessExists(pid int, procs []ProcessInfo) bool {
	for _, proc := range procs {
		if proc.PID == pid {
			return true
		}
	}
	return false
}

// CollectRunProcessPIDs gathers all PIDs associated with an RPI run:
// the orchestrator, its descendants, tmux session processes, and worktree processes.
func CollectRunProcessPIDs(orchestratorPID int, runID, worktreePath string, procs []ProcessInfo) []int {
	set := make(map[int]struct{})

	addWithDescendants := func(pid int) {
		if pid <= 1 || !ProcessExists(pid, procs) {
			return
		}
		set[pid] = struct{}{}
		for _, child := range DescendantPIDs(pid, procs) {
			set[child] = struct{}{}
		}
	}

	addWithDescendants(orchestratorPID)

	sessionNeedle := fmt.Sprintf("ao-rpi-%s-p", runID)
	for _, proc := range procs {
		cmd := proc.Command
		if strings.Contains(cmd, sessionNeedle) {
			addWithDescendants(proc.PID)
		}
		if worktreePath != "" && strings.Contains(cmd, worktreePath) {
			addWithDescendants(proc.PID)
		}
	}

	var pids []int
	for pid := range set {
		pids = append(pids, pid)
	}
	sort.Ints(pids)
	return pids
}

// SupervisorLeaseExpired returns true when the lease expiry has passed.
// Backward-compatible: leases without an ExpiresAt field are never expired.
// Corrupted expiry metadata is treated as stale (expired).
func SupervisorLeaseExpired(meta SupervisorLeaseMetadata, now time.Time) bool {
	expiryRaw := strings.TrimSpace(meta.ExpiresAt)
	if expiryRaw == "" {
		return false
	}
	expiry, err := time.Parse(time.RFC3339, expiryRaw)
	if err != nil {
		return true
	}
	return now.After(expiry)
}

// ListProcesses returns a snapshot of the system process table via `ps`.
func ListProcesses() ([]ProcessInfo, error) {
	cmd := exec.Command("ps", "-axo", "pid=,ppid=,command=")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list processes: %w", err)
	}
	return ParseProcessList(out)
}

// ParseProcessList parses raw `ps -axo pid=,ppid=,command=` output into ProcessInfo entries.
func ParseProcessList(data []byte) ([]ProcessInfo, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var procs []ProcessInfo
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		ppid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		procs = append(procs, ProcessInfo{
			PID:     pid,
			PPID:    ppid,
			Command: strings.Join(fields[2:], " "),
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("parse process list: %w", err)
	}
	return procs, nil
}
