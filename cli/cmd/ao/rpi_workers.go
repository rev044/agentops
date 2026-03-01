package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/boshu2/agentops/cli/internal/formatter"
	"github.com/spf13/cobra"
)

var (
	rpiWorkersRunID string
	rpiWorkersJSON  bool
)

const workerEventStaleThreshold = 10 * time.Minute

type rpiWorkerStatus struct {
	WorkerID      string `json:"worker_id"`
	Health        string `json:"health"`
	Reason        string `json:"reason"`
	LastEventType string `json:"last_event_type,omitempty"`
	LastEventAt   string `json:"last_event_at,omitempty"`
	Backend       string `json:"backend,omitempty"`
	Phase         int    `json:"phase,omitempty"`
}

type rpiWorkersOutput struct {
	RunID       string            `json:"run_id"`
	GeneratedAt string            `json:"generated_at"`
	Workers     []rpiWorkerStatus `json:"workers"`
}

func init() {
	workersCmd := &cobra.Command{
		Use:   "workers",
		Short: "Show per-worker health derived from normalized RPI events",
		RunE:  runRPIWorkers,
	}
	workersCmd.Flags().StringVar(&rpiWorkersRunID, "run-id", "", "Run ID to inspect (defaults to latest phased state)")
	workersCmd.Flags().BoolVar(&rpiWorkersJSON, "json", false, "Render workers output as JSON")
	addRPISubcommand(workersCmd)
}

func runRPIWorkers(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	runID, _, root, err := resolveNudgeRun(cwd, strings.TrimSpace(rpiWorkersRunID))
	if err != nil {
		return err
	}

	events, err := loadRPIC2Events(root, runID)
	if err != nil {
		return err
	}
	heartbeat := readRunHeartbeat(root, runID)
	workers := projectWorkerHealth(events, heartbeat)

	output := rpiWorkersOutput{
		RunID:       runID,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339Nano),
		Workers:     workers,
	}
	if rpiWorkersJSON || GetOutput() == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	}

	if len(workers) == 0 {
		fmt.Println("No worker events found.")
		return nil
	}
	fmt.Printf("RUN-ID: %s\n", runID)
	tbl := formatter.NewTable(os.Stdout, "WORKER_ID", "HEALTH", "REASON", "LAST_EVENT", "LAST_EVENT_AT")
	for _, worker := range workers {
		tbl.AddRow(worker.WorkerID, worker.Health, worker.Reason, worker.LastEventType, worker.LastEventAt)
	}
	return tbl.Render()
}

func projectWorkerHealth(events []RPIC2Event, heartbeat time.Time) []rpiWorkerStatus {
	type snapshot struct {
		status rpiWorkerStatus
		time   time.Time
	}
	byWorker := make(map[string]snapshot)

	for _, ev := range events {
		workerID := strings.TrimSpace(ev.WorkerID)
		if workerID == "" {
			continue
		}
		timestamp, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(ev.Timestamp))
		if err != nil {
			continue
		}
		current, ok := byWorker[workerID]
		if ok && !timestamp.After(current.time) {
			continue
		}
		status := rpiWorkerStatus{
			WorkerID:      workerID,
			LastEventType: ev.Type,
			LastEventAt:   ev.Timestamp,
			Backend:       ev.Backend,
			Phase:         ev.Phase,
		}
		applyWorkerHealth(&status, ev, timestamp, heartbeat)
		byWorker[workerID] = snapshot{status: status, time: timestamp}
	}

	workers := make([]rpiWorkerStatus, 0, len(byWorker))
	for _, entry := range byWorker {
		workers = append(workers, entry.status)
	}
	sort.Slice(workers, func(i, j int) bool {
		return workers[i].WorkerID < workers[j].WorkerID
	})
	return workers
}

func applyWorkerHealth(status *rpiWorkerStatus, event RPIC2Event, eventAt, heartbeat time.Time) {
	typeLower := strings.ToLower(strings.TrimSpace(event.Type))
	if strings.Contains(typeLower, "failed") {
		status.Health = "failed"
		status.Reason = "worker_failed_event"
		return
	}
	if strings.HasSuffix(typeLower, "rpi_worker_end") {
		status.Health = "healthy"
		status.Reason = "worker_completed"
		return
	}
	now := time.Now().UTC()
	heartbeatFresh := !heartbeat.IsZero() && now.Sub(heartbeat) <= heartbeatLiveThreshold
	eventFresh := now.Sub(eventAt) <= workerEventStaleThreshold

	switch {
	case eventFresh && heartbeatFresh:
		status.Health = "healthy"
		status.Reason = "worker_active"
	case !eventFresh:
		status.Health = "stale"
		status.Reason = "stale_worker_event"
	case !heartbeatFresh:
		status.Health = "unknown"
		status.Reason = "run_heartbeat_stale"
	default:
		status.Health = "unknown"
		status.Reason = "insufficient_signal"
	}
}
