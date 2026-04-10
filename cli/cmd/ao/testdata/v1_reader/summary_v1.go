// Package v1reader holds a FROZEN snapshot of the overnightSummary
// struct as it existed on 2026-04-09 (schema v1). Used by the
// backward-compatibility test in overnight_test.go to verify that
// v2 JSON output remains parseable by a strict v1 reader.
//
// DO NOT EDIT this struct to match the current overnightSummary —
// it is intentionally frozen. When schema_version advances again
// to v3+, add a new frozen snapshot alongside this one; do not
// replace this one. This package has NO methods — only struct
// definitions — so it cannot accidentally drift from the snapshot.
package v1reader

// OvernightSummaryV1 mirrors overnightSummary from cli/cmd/ao/overnight.go
// at the 2026-04-09 HEAD before Wave 1 of the Dream nightly compounder
// landed. Every v1 field is required; unknown fields in the incoming
// JSON are ignored by encoding/json by default.
//
// The shape intentionally omits the schema v2 additive fields
// (Iterations, FitnessDelta, PlateauReason, RegressionReason) because
// those are the additive bump a strict v1 reader must tolerate.
type OvernightSummaryV1 struct {
	SchemaVersion int                `json:"schema_version"`
	Mode          string             `json:"mode"`
	RunID         string             `json:"run_id"`
	Goal          string             `json:"goal,omitempty"`
	RepoRoot      string             `json:"repo_root"`
	OutputDir     string             `json:"output_dir"`
	Status        string             `json:"status"`
	DryRun        bool               `json:"dry_run"`
	StartedAt     string             `json:"started_at"`
	FinishedAt    string             `json:"finished_at,omitempty"`
	Duration      string             `json:"duration,omitempty"`
	Runtime       OvernightRuntimeV1 `json:"runtime"`
	Steps         []OvernightStepV1  `json:"steps"`
	Artifacts     map[string]string  `json:"artifacts,omitempty"`
	MetricsHealth map[string]any     `json:"metrics_health,omitempty"`
	RetrievalLive map[string]any     `json:"retrieval_live,omitempty"`
	CloseLoop     map[string]any     `json:"close_loop,omitempty"`
	Briefing      map[string]any     `json:"briefing,omitempty"`
	Degraded      []string           `json:"degraded,omitempty"`
	Recommended   []string           `json:"recommended,omitempty"`
	NextAction    string             `json:"next_action,omitempty"`
}

// OvernightRuntimeV1 is the frozen v1 shape of overnightRuntimeSummary.
type OvernightRuntimeV1 struct {
	KeepAwake          bool   `json:"keep_awake"`
	KeepAwakeMode      string `json:"keep_awake_mode"`
	KeepAwakeNote      string `json:"keep_awake_note,omitempty"`
	RequestedTimeout   string `json:"requested_timeout"`
	EffectiveTimeout   string `json:"effective_timeout"`
	LockPath           string `json:"lock_path"`
	LogPath            string `json:"log_path"`
	ProcessContractDoc string `json:"process_contract_doc"`
	ReportContractDoc  string `json:"report_contract_doc"`
}

// OvernightStepV1 is the frozen v1 shape of overnightStepSummary.
type OvernightStepV1 struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Command  string `json:"command,omitempty"`
	Artifact string `json:"artifact,omitempty"`
	Note     string `json:"note,omitempty"`
}
