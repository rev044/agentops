package quality

// MetricsOptions configures a flywheel metrics computation pass.
//
// It is used by the cmd/ao layer to thread CLI flags through to the
// internal/quality helpers without leaking cobra/flag types into the package.
type MetricsOptions struct {
	// BaseDir is the repository root used to locate .agents/ artifacts.
	BaseDir string
	// Days is the period (in days) for citation/artifact filtering.
	Days int
	// Namespace is the metric namespace to filter citations against.
	Namespace string
	// JSON indicates the caller wants machine-readable output.
	JSON bool
}
