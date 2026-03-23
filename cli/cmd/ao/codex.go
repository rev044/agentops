package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/boshu2/agentops/cli/internal/pool"
	"github.com/boshu2/agentops/cli/internal/ratchet"
	"github.com/boshu2/agentops/cli/internal/storage"
	"github.com/boshu2/agentops/cli/internal/types"
)

var (
	codexStartLimit            int
	codexStartQuery            string
	codexStartNoMaintenance    bool
	codexStopSessionID         string
	codexStopTranscriptPath    string
	codexStopAutoExtract       bool
	codexStopNoHistoryFallback bool
	codexStopNoCloseLoop       bool
	codexStatusDays            int
)

type codexArtifactRef struct {
	Title      string `json:"title"`
	Path       string `json:"path"`
	ModifiedAt string `json:"modified_at"`
}

type codexLifecycleEvent struct {
	SessionID           string `json:"session_id,omitempty"`
	ThreadName          string `json:"thread_name,omitempty"`
	Query               string `json:"query,omitempty"`
	Timestamp           string `json:"timestamp"`
	TranscriptPath      string `json:"transcript_path,omitempty"`
	TranscriptSource    string `json:"transcript_source,omitempty"`
	SyntheticTranscript bool   `json:"synthetic_transcript,omitempty"`
	StartupContextPath  string `json:"startup_context_path,omitempty"`
	MemoryPath          string `json:"memory_path,omitempty"`
	Status              string `json:"status,omitempty"`
	Summary             string `json:"summary,omitempty"`
	HandoffPath         string `json:"handoff_path,omitempty"`
}

type codexLifecycleState struct {
	SchemaVersion int                     `json:"schema_version"`
	Runtime       lifecycleRuntimeProfile `json:"runtime"`
	LastStart     *codexLifecycleEvent    `json:"last_start,omitempty"`
	LastStop      *codexLifecycleEvent    `json:"last_stop,omitempty"`
	UpdatedAt     string                  `json:"updated_at"`
}

type codexStartResult struct {
	Runtime            lifecycleRuntimeProfile  `json:"runtime"`
	ContextQuery       string                   `json:"context_query,omitempty"`
	StartupContextPath string                   `json:"startup_context_path"`
	MemoryPath         string                   `json:"memory_path,omitempty"`
	CloseLoop          *flywheelCloseLoopResult `json:"close_loop,omitempty"`
	Flywheel           *flywheelBrief           `json:"flywheel,omitempty"`
	Learnings          []learning               `json:"learnings,omitempty"`
	Patterns           []pattern                `json:"patterns,omitempty"`
	Findings           []knowledgeFinding       `json:"findings,omitempty"`
	RecentSessions     []session                `json:"recent_sessions,omitempty"`
	NextWork           []nextWorkItem           `json:"next_work,omitempty"`
	Research           []codexArtifactRef       `json:"research,omitempty"`
	StatePath          string                   `json:"state_path"`
}

type codexStopResult struct {
	Runtime             lifecycleRuntimeProfile  `json:"runtime"`
	TranscriptPath      string                   `json:"transcript_path"`
	TranscriptSource    string                   `json:"transcript_source"`
	SyntheticTranscript bool                     `json:"synthetic_transcript,omitempty"`
	Session             SessionCloseResult       `json:"session"`
	CloseLoop           *flywheelCloseLoopResult `json:"close_loop,omitempty"`
	MemoryPath          string                   `json:"memory_path,omitempty"`
	StatePath           string                   `json:"state_path"`
}

type codexCaptureHealth struct {
	SessionsIndexed   int    `json:"sessions_indexed"`
	LastForgeTime     string `json:"last_forge_time,omitempty"`
	LastForgeAge      string `json:"last_forge_age,omitempty"`
	PendingKnowledge  int    `json:"pending_knowledge"`
	PendingQuarantine int    `json:"pending_quarantine"`
}

type codexRetrievalHealth struct {
	Learnings int `json:"learnings"`
	Patterns  int `json:"patterns"`
	Findings  int `json:"findings"`
	NextWork  int `json:"next_work"`
	Research  int `json:"research"`
}

type codexPromotionHealth struct {
	PendingPool  int `json:"pending_pool"`
	StagedPool   int `json:"staged_pool"`
	RejectedPool int `json:"rejected_pool"`
}

type codexCitationHealth struct {
	WindowDays      int `json:"window_days"`
	Total           int `json:"total"`
	UniqueArtifacts int `json:"unique_artifacts"`
	Retrieved       int `json:"retrieved"`
	Reference       int `json:"reference"`
	Applied         int `json:"applied"`
}

type codexStatusResult struct {
	Runtime   lifecycleRuntimeProfile `json:"runtime"`
	State     *codexLifecycleState    `json:"state,omitempty"`
	Flywheel  *flywheelBrief          `json:"flywheel,omitempty"`
	Capture   codexCaptureHealth      `json:"capture"`
	Retrieval codexRetrievalHealth    `json:"retrieval"`
	Promotion codexPromotionHealth    `json:"promotion"`
	Citations codexCitationHealth     `json:"citations"`
}

var codexCmd = &cobra.Command{
	Use:   "codex",
	Short: "Codex-native lifecycle commands for hookless sessions",
	Long: `Codex-native lifecycle commands for runtimes without Claude/OpenCode lifecycle hooks.

Use these commands to make the knowledge flywheel explicit in Codex:
  ao codex start   Surface prior context and run safe maintenance
  ao codex stop    Forge the current session, queue learnings, and close the loop
  ao codex status  Show hookless lifecycle health and flywheel status`,
}

var codexStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a Codex session with explicit flywheel maintenance",
	Args:  cobra.NoArgs,
	RunE:  runCodexStart,
}

var codexStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Close a Codex session without relying on runtime hooks",
	Args:  cobra.NoArgs,
	RunE:  runCodexStop,
}

var codexStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show Codex hookless lifecycle health",
	Args:  cobra.NoArgs,
	RunE:  runCodexStatus,
}

func init() {
	codexCmd.GroupID = "workflow"
	rootCmd.AddCommand(codexCmd)
	codexCmd.AddCommand(codexStartCmd, codexStopCmd, codexStatusCmd)

	codexStartCmd.Flags().IntVar(&codexStartLimit, "limit", 3, "Maximum artifacts to surface per category")
	codexStartCmd.Flags().StringVar(&codexStartQuery, "query", "", "Optional startup query (defaults to the current Codex thread name)")
	codexStartCmd.Flags().BoolVar(&codexStartNoMaintenance, "no-maintenance", false, "Skip safe close-loop maintenance on start")

	codexStopCmd.Flags().StringVar(&codexStopSessionID, "session", "", "Codex session ID to close (defaults to the active thread)")
	codexStopCmd.Flags().StringVar(&codexStopTranscriptPath, "transcript", "", "Explicit transcript path to forge instead of runtime discovery")
	codexStopCmd.Flags().BoolVar(&codexStopAutoExtract, "auto-extract", true, "Write lightweight learnings and handoff artifacts during closeout")
	codexStopCmd.Flags().BoolVar(&codexStopNoHistoryFallback, "no-history-fallback", false, "Disable history.jsonl fallback when no archived Codex transcript exists")
	codexStopCmd.Flags().BoolVar(&codexStopNoCloseLoop, "no-close-loop", false, "Skip flywheel close-loop maintenance after forging")

	codexStatusCmd.Flags().IntVar(&codexStatusDays, "days", 7, "Citation window in days for Codex lifecycle health")
}

func runCodexStart(cmd *cobra.Command, args []string) error {
	cwd, err := resolveProjectDir()
	if err != nil {
		return err
	}
	if err := ensureCodexLifecycleDirs(cwd); err != nil {
		return err
	}

	profile := detectCodexLifecycleProfile()
	sessionID := profile.SessionID
	if strings.TrimSpace(sessionID) == "" {
		sessionID = resolveSessionID("")
	}

	query := strings.TrimSpace(codexStartQuery)
	if query == "" {
		query = strings.TrimSpace(profile.ThreadName)
	}
	if query == "" {
		query = "codex startup"
	}

	var closeLoop *flywheelCloseLoopResult
	if !codexStartNoMaintenance {
		threshold, err := time.ParseDuration(defaultAutoPromoteThreshold)
		if err != nil {
			return fmt.Errorf("parse default close-loop threshold: %w", err)
		}
		result, err := performFlywheelCloseLoop(cwd, filepath.Join(".agents", "knowledge", "pending"), threshold, true)
		if err != nil {
			return fmt.Errorf("run codex startup maintenance: %w", err)
		}
		closeLoop = &result
	}

	learnings, patterns, findings, recentSessions, nextWork, research := collectCodexStartupArtifacts(cwd, query, codexStartLimit)
	recordLookupCitations(cwd, learnings, patterns, findings, sessionID, query, "retrieved")

	memoryPath, err := syncCodexMemory(cwd)
	if err != nil {
		VerbosePrintf("Warning: codex memory sync: %v\n", err)
	}

	startupContextPath, err := writeCodexStartupContext(cwd, profile, query, learnings, patterns, findings, recentSessions, nextWork, research)
	if err != nil {
		return fmt.Errorf("write codex startup context: %w", err)
	}

	state, statePath, err := loadOrInitCodexLifecycleState(cwd)
	if err != nil {
		return err
	}
	state.Runtime = profile
	state.LastStart = &codexLifecycleEvent{
		SessionID:          sessionID,
		ThreadName:         profile.ThreadName,
		Query:              query,
		Timestamp:          time.Now().UTC().Format(time.RFC3339),
		StartupContextPath: startupContextPath,
		MemoryPath:         memoryPath,
		Status:             lifecycleModeCodexHookless,
		Summary:            fmt.Sprintf("surfaced %d learnings, %d patterns, %d findings", len(learnings), len(patterns), len(findings)),
	}
	state.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := saveCodexLifecycleState(statePath, state); err != nil {
		return err
	}

	result := codexStartResult{
		Runtime:            profile,
		ContextQuery:       query,
		StartupContextPath: startupContextPath,
		MemoryPath:         memoryPath,
		CloseLoop:          closeLoop,
		Flywheel:           loadFlywheelBrief(cwd),
		Learnings:          learnings,
		Patterns:           patterns,
		Findings:           findings,
		RecentSessions:     recentSessions,
		NextWork:           nextWork,
		Research:           research,
		StatePath:          statePath,
	}
	return outputCodexStartResult(result)
}

func runCodexStop(cmd *cobra.Command, args []string) error {
	cwd, err := resolveProjectDir()
	if err != nil {
		return err
	}
	if err := ensureCodexLifecycleDirs(cwd); err != nil {
		return err
	}

	profile := detectCodexLifecycleProfile()
	sessionID := strings.TrimSpace(codexStopSessionID)
	if sessionID == "" {
		sessionID = strings.TrimSpace(profile.SessionID)
	}

	transcriptPath := strings.TrimSpace(codexStopTranscriptPath)
	transcriptSource := "explicit"
	syntheticTranscript := false

	if transcriptPath == "" {
		transcriptPath, transcriptSource, syntheticTranscript, sessionID, err = resolveCodexStopTranscript(cwd, sessionID, codexStopNoHistoryFallback)
		if err != nil {
			return err
		}
	}

	closeResult, err := forgeExtractReportWithOptions(transcriptPath, cwd, codexStopAutoExtract, false)
	if err != nil {
		return err
	}

	var closeLoop *flywheelCloseLoopResult
	if !codexStopNoCloseLoop {
		threshold, err := time.ParseDuration(defaultAutoPromoteThreshold)
		if err != nil {
			return fmt.Errorf("parse default close-loop threshold: %w", err)
		}
		result, err := performFlywheelCloseLoop(cwd, filepath.Join(".agents", "knowledge", "pending"), threshold, true)
		if err != nil {
			return fmt.Errorf("run codex close-loop maintenance: %w", err)
		}
		closeLoop = &result
	}

	memoryPath, err := syncCodexMemory(cwd)
	if err != nil {
		VerbosePrintf("Warning: codex memory sync: %v\n", err)
	}

	state, statePath, err := loadOrInitCodexLifecycleState(cwd)
	if err != nil {
		return err
	}
	state.Runtime = profile
	state.LastStop = &codexLifecycleEvent{
		SessionID:           closeResult.SessionID,
		ThreadName:          profile.ThreadName,
		Timestamp:           time.Now().UTC().Format(time.RFC3339),
		TranscriptPath:      transcriptPath,
		TranscriptSource:    transcriptSource,
		SyntheticTranscript: syntheticTranscript,
		MemoryPath:          memoryPath,
		Status:              closeResult.Status,
		Summary:             closeResult.Message,
		HandoffPath:         closeResult.HandoffWritten,
	}
	state.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := saveCodexLifecycleState(statePath, state); err != nil {
		return err
	}

	result := codexStopResult{
		Runtime:             profile,
		TranscriptPath:      transcriptPath,
		TranscriptSource:    transcriptSource,
		SyntheticTranscript: syntheticTranscript,
		Session:             closeResult,
		CloseLoop:           closeLoop,
		MemoryPath:          memoryPath,
		StatePath:           statePath,
	}
	return outputCodexStopResult(result)
}

func runCodexStatus(cmd *cobra.Command, args []string) error {
	cwd, err := resolveProjectDir()
	if err != nil {
		return err
	}

	profile := detectCodexLifecycleProfile()
	state, _, err := loadOrInitCodexLifecycleState(cwd)
	if err != nil {
		return err
	}

	result := codexStatusResult{
		Runtime:   profile,
		State:     state,
		Flywheel:  loadFlywheelBrief(cwd),
		Capture:   collectCodexCaptureHealth(cwd),
		Retrieval: collectCodexRetrievalHealth(cwd),
		Promotion: collectCodexPromotionHealth(cwd),
		Citations: collectCodexCitationHealth(cwd, codexStatusDays),
	}
	return outputCodexStatusResult(result)
}

func collectCodexStartupArtifacts(cwd, query string, limit int) ([]learning, []pattern, []knowledgeFinding, []session, []nextWorkItem, []codexArtifactRef) {
	if limit <= 0 {
		limit = 3
	}

	learnings, _ := collectLearnings(cwd, query, limit, "", 0)
	patterns, _ := collectPatterns(cwd, query, limit, "", 0)
	findings, _ := collectFindings(cwd, query, limit, "", 0)
	recentSessions, _ := collectRecentSessions(cwd, query, minInt(limit, MaxSessionsToInject))

	repoFilter := filepath.Base(cwd)
	if root := findGitRoot(cwd); root != "" {
		repoFilter = filepath.Base(root)
	}
	nextWork, _ := readUnconsumedItems(filepath.Join(cwd, ".agents", "rpi", "next-work.jsonl"), repoFilter)
	if len(nextWork) > limit {
		nextWork = nextWork[:limit]
	}

	research := collectRecentResearchArtifacts(cwd, query, limit)
	return learnings, patterns, findings, recentSessions, nextWork, research
}

func collectRecentResearchArtifacts(cwd, query string, limit int) []codexArtifactRef {
	researchDir := filepath.Join(cwd, ".agents", SectionResearch)
	entries, err := os.ReadDir(researchDir)
	if err != nil {
		return nil
	}

	type researchFile struct {
		path    string
		modTime time.Time
	}
	var files []researchFile
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, researchFile{
			path:    filepath.Join(researchDir, entry.Name()),
			modTime: info.ModTime(),
		})
	}
	slices.SortFunc(files, func(a, b researchFile) int {
		return b.modTime.Compare(a.modTime)
	})

	queryLower := strings.ToLower(strings.TrimSpace(query))
	var artifacts []codexArtifactRef
	for _, file := range files {
		if queryLower != "" {
			baseLower := strings.ToLower(filepath.Base(file.path))
			if !strings.Contains(baseLower, queryLower) {
				data, err := os.ReadFile(file.path)
				if err != nil || !strings.Contains(strings.ToLower(string(data)), queryLower) {
					continue
				}
			}
		}
		artifacts = append(artifacts, codexArtifactRef{
			Title:      strings.TrimSuffix(filepath.Base(file.path), filepath.Ext(file.path)),
			Path:       file.path,
			ModifiedAt: file.modTime.UTC().Format(time.RFC3339),
		})
		if len(artifacts) >= limit {
			break
		}
	}
	return artifacts
}

func resolveCodexStopTranscript(cwd, sessionID string, noHistoryFallback bool) (string, string, bool, string, error) {
	if sessionID != "" {
		if path, err := findTranscriptBySessionID(sessionID); err == nil {
			return path, "archived", false, sessionID, nil
		}
		if !noHistoryFallback {
			path, err := synthesizeCodexHistoryTranscript(cwd, sessionID)
			if err == nil {
				return path, "history-fallback", true, sessionID, nil
			}
		}
	}

	if path, err := findLastCodexArchivedTranscript(); err == nil {
		return path, "archived", false, extractSessionIDFromCodexArchivedPath(path), nil
	}

	if noHistoryFallback {
		return "", "", false, sessionID, fmt.Errorf("no Codex transcript found and history fallback is disabled")
	}

	fallbackSessionID := sessionID
	if fallbackSessionID == "" {
		fallbackSessionID = resolveCodexSessionIDFromHome()
	}
	if fallbackSessionID == "" {
		return "", "", false, "", fmt.Errorf("no Codex transcript or active history session found")
	}
	path, err := synthesizeCodexHistoryTranscript(cwd, fallbackSessionID)
	if err != nil {
		return "", "", false, fallbackSessionID, err
	}
	return path, "history-fallback", true, fallbackSessionID, nil
}

func resolveCodexSessionIDFromHome() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return resolveCodexSessionID(homeDir)
}

func extractSessionIDFromCodexArchivedPath(path string) string {
	match := codexArchivedSessionPattern.FindStringSubmatch(filepath.Base(path))
	if len(match) == 2 {
		return match[1]
	}
	return ""
}

func syncCodexMemory(cwd string) (string, error) {
	root := findGitRoot(cwd)
	if root == "" {
		root = cwd
	}
	path := filepath.Join(root, "MEMORY.md")
	if err := syncMemory(cwd, path, 10, true); err != nil {
		return path, err
	}
	return path, nil
}

func codexLifecycleStatePath(cwd string) string {
	return filepath.Join(cwd, ".agents", "ao", "codex", "state.json")
}

func ensureCodexLifecycleDirs(cwd string) error {
	for _, dir := range []string{
		filepath.Join(cwd, ".agents", "ao", "codex"),
		filepath.Join(cwd, ".agents", "ao", "codex", "transcripts"),
	} {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("create codex lifecycle dir %s: %w", dir, err)
		}
	}
	return nil
}

func loadOrInitCodexLifecycleState(cwd string) (*codexLifecycleState, string, error) {
	if err := ensureCodexLifecycleDirs(cwd); err != nil {
		return nil, "", err
	}
	path := codexLifecycleStatePath(cwd)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &codexLifecycleState{SchemaVersion: 1}, path, nil
		}
		return nil, "", fmt.Errorf("read codex lifecycle state: %w", err)
	}

	var state codexLifecycleState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, "", fmt.Errorf("parse codex lifecycle state: %w", err)
	}
	if state.SchemaVersion == 0 {
		state.SchemaVersion = 1
	}
	return &state, path, nil
}

func saveCodexLifecycleState(path string, state *codexLifecycleState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal codex lifecycle state: %w", err)
	}
	if err := atomicWriteFile(path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write codex lifecycle state: %w", err)
	}
	return nil
}

func writeCodexStartupContext(cwd string, profile lifecycleRuntimeProfile, query string, learnings []learning, patterns []pattern, findings []knowledgeFinding, recentSessions []session, nextWork []nextWorkItem, research []codexArtifactRef) (string, error) {
	var sb strings.Builder
	sb.WriteString("# Codex Startup Context\n\n")
	sb.WriteString(fmt.Sprintf("- Runtime: %s\n", profile.Runtime))
	sb.WriteString(fmt.Sprintf("- Lifecycle mode: %s\n", profile.Mode))
	if profile.ThreadName != "" {
		sb.WriteString(fmt.Sprintf("- Thread: %s\n", profile.ThreadName))
	}
	if query != "" {
		sb.WriteString(fmt.Sprintf("- Query: %s\n", query))
	}
	sb.WriteString("\n## Learnings\n")
	if len(learnings) == 0 {
		sb.WriteString("- None surfaced\n")
	} else {
		for _, item := range learnings {
			sb.WriteString(fmt.Sprintf("- %s\n", firstLine(item.Title)))
		}
	}
	sb.WriteString("\n## Patterns\n")
	if len(patterns) == 0 {
		sb.WriteString("- None surfaced\n")
	} else {
		for _, item := range patterns {
			sb.WriteString(fmt.Sprintf("- %s\n", firstLine(item.Name)))
		}
	}
	sb.WriteString("\n## Findings\n")
	if len(findings) == 0 {
		sb.WriteString("- None surfaced\n")
	} else {
		for _, item := range findings {
			sb.WriteString(fmt.Sprintf("- %s\n", firstLine(item.Title)))
		}
	}
	sb.WriteString("\n## Next Work\n")
	if len(nextWork) == 0 {
		sb.WriteString("- No queued next work\n")
	} else {
		for _, item := range nextWork {
			sb.WriteString(fmt.Sprintf("- %s\n", firstLine(item.Title)))
		}
	}
	sb.WriteString("\n## Recent Sessions\n")
	if len(recentSessions) == 0 {
		sb.WriteString("- No recent session summaries\n")
	} else {
		for _, item := range recentSessions {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", item.Date, firstLine(item.Summary)))
		}
	}
	sb.WriteString("\n## Research\n")
	if len(research) == 0 {
		sb.WriteString("- No recent research surfaced\n")
	} else {
		for _, item := range research {
			sb.WriteString(fmt.Sprintf("- %s\n", item.Title))
		}
	}

	path := filepath.Join(cwd, ".agents", "ao", "codex", "startup-context.md")
	if err := atomicWriteFile(path, []byte(sb.String()), 0o600); err != nil {
		return "", err
	}
	return path, nil
}

func countGlobMatches(pattern string) int {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return 0
	}
	return len(matches)
}

func collectCodexCaptureHealth(cwd string) codexCaptureHealth {
	result := codexCaptureHealth{
		SessionsIndexed:   countGlobMatches(filepath.Join(cwd, storage.DefaultBaseDir, storage.SessionsDir, "*.jsonl")),
		PendingKnowledge:  countGlobMatches(filepath.Join(cwd, ".agents", "knowledge", "pending", "*.md")),
		PendingQuarantine: countGlobMatches(filepath.Join(cwd, ".agents", "knowledge", "pending", ".quarantine", "*.md")),
	}
	if lastForge := findLastForgeTime(cwd); !lastForge.IsZero() {
		result.LastForgeTime = lastForge.UTC().Format(time.RFC3339)
		result.LastForgeAge = formatDurationBrief(time.Since(lastForge))
	}
	return result
}

func collectCodexRetrievalHealth(cwd string) codexRetrievalHealth {
	repoFilter := filepath.Base(cwd)
	if root := findGitRoot(cwd); root != "" {
		repoFilter = filepath.Base(root)
	}
	nextWork, _ := readUnconsumedItems(filepath.Join(cwd, ".agents", "rpi", "next-work.jsonl"), repoFilter)
	return codexRetrievalHealth{
		Learnings: countGlobMatches(filepath.Join(cwd, ".agents", "learnings", "*.md")),
		Patterns:  countGlobMatches(filepath.Join(cwd, ".agents", "patterns", "*.md")),
		Findings:  countGlobMatches(filepath.Join(cwd, ".agents", SectionFindings, "*.md")),
		NextWork:  len(nextWork),
		Research:  countGlobMatches(filepath.Join(cwd, ".agents", SectionResearch, "*.md")),
	}
}

func collectCodexPromotionHealth(cwd string) codexPromotionHealth {
	p := pool.NewPool(cwd)
	pending, _ := p.List(pool.ListOptions{Status: types.PoolStatusPending})
	staged, _ := p.List(pool.ListOptions{Status: types.PoolStatusStaged})
	rejected, _ := p.List(pool.ListOptions{Status: types.PoolStatusRejected})
	return codexPromotionHealth{
		PendingPool:  len(pending),
		StagedPool:   len(staged),
		RejectedPool: len(rejected),
	}
}

func collectCodexCitationHealth(cwd string, days int) codexCitationHealth {
	result := codexCitationHealth{WindowDays: days}
	citations, err := ratchet.LoadCitations(cwd)
	if err != nil {
		return result
	}
	end := time.Now()
	start := end.AddDate(0, 0, -days)
	unique := make(map[string]bool)
	for _, citation := range citations {
		if citation.CitedAt.Before(start) || citation.CitedAt.After(end) {
			continue
		}
		if !isRetrievableArtifactPath(cwd, citation.ArtifactPath) {
			continue
		}
		result.Total++
		unique[canonicalArtifactKey(cwd, citation.ArtifactPath)] = true
		switch canonicalCitationType(citation.CitationType) {
		case "applied":
			result.Applied++
		case "reference":
			result.Reference++
		default:
			result.Retrieved++
		}
	}
	result.UniqueArtifacts = len(unique)
	return result
}

func outputCodexStartResult(result codexStartResult) error {
	if GetOutput() == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	fmt.Println("Codex Start")
	fmt.Println("===========")
	fmt.Printf("Mode: %s (%s)\n", result.Runtime.Mode, result.Runtime.Runtime)
	if result.Runtime.ThreadName != "" {
		fmt.Printf("Thread: %s\n", result.Runtime.ThreadName)
	}
	fmt.Printf("Startup context: %s\n", result.StartupContextPath)
	if result.MemoryPath != "" {
		fmt.Printf("Memory: %s\n", result.MemoryPath)
	}
	if result.CloseLoop != nil {
		fmt.Printf("Maintenance: ingest=%d promote=%d reward=%d\n",
			result.CloseLoop.Ingest.Added, result.CloseLoop.AutoPromote.Promoted, result.CloseLoop.CitationFeedback.Rewarded)
	}
	fmt.Println()
	printNamedItems("Learnings", result.Learnings, func(item learning) string { return firstLine(item.Title) })
	printNamedItems("Patterns", result.Patterns, func(item pattern) string { return firstLine(item.Name) })
	printNamedItems("Findings", result.Findings, func(item knowledgeFinding) string { return firstLine(item.Title) })
	printNamedItems("Next Work", result.NextWork, func(item nextWorkItem) string { return firstLine(item.Title) })
	return nil
}

func outputCodexStopResult(result codexStopResult) error {
	if GetOutput() == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	fmt.Println("Codex Stop")
	fmt.Println("==========")
	fmt.Printf("Mode: %s (%s)\n", result.Runtime.Mode, result.Runtime.Runtime)
	fmt.Printf("Transcript: %s\n", shortenPath(result.TranscriptPath))
	fmt.Printf("Source: %s\n", result.TranscriptSource)
	if result.SyntheticTranscript {
		fmt.Println("Transcript mode: synthesized from Codex history.jsonl")
	}
	fmt.Printf("Session: %s\n", result.Session.SessionID)
	fmt.Printf("Learnings: %d extracted, %d rejected\n", result.Session.LearningsExtracted, result.Session.LearningsRejected)
	if result.Session.HandoffWritten != "" {
		fmt.Printf("Handoff: %s\n", shortenPath(result.Session.HandoffWritten))
	}
	if result.CloseLoop != nil {
		fmt.Printf("Close-loop: ingest=%d promote=%d reward=%d\n",
			result.CloseLoop.Ingest.Added, result.CloseLoop.AutoPromote.Promoted, result.CloseLoop.CitationFeedback.Rewarded)
	}
	return nil
}

func outputCodexStatusResult(result codexStatusResult) error {
	if GetOutput() == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	fmt.Println("Codex Lifecycle Status")
	fmt.Println("======================")
	fmt.Printf("Mode: %s (%s)\n", result.Runtime.Mode, result.Runtime.Runtime)
	if result.Runtime.ThreadName != "" {
		fmt.Printf("Thread: %s\n", result.Runtime.ThreadName)
	}
	fmt.Println()
	fmt.Printf("Capture: sessions=%d pending=%d quarantine=%d\n",
		result.Capture.SessionsIndexed, result.Capture.PendingKnowledge, result.Capture.PendingQuarantine)
	if result.Capture.LastForgeAge != "" {
		fmt.Printf("Last forge: %s ago\n", result.Capture.LastForgeAge)
	}
	fmt.Printf("Retrieval: learnings=%d patterns=%d findings=%d next-work=%d research=%d\n",
		result.Retrieval.Learnings, result.Retrieval.Patterns, result.Retrieval.Findings, result.Retrieval.NextWork, result.Retrieval.Research)
	fmt.Printf("Promotion: pending=%d staged=%d rejected=%d\n",
		result.Promotion.PendingPool, result.Promotion.StagedPool, result.Promotion.RejectedPool)
	fmt.Printf("Citations (%dd): total=%d unique=%d retrieved=%d reference=%d applied=%d\n",
		result.Citations.WindowDays, result.Citations.Total, result.Citations.UniqueArtifacts,
		result.Citations.Retrieved, result.Citations.Reference, result.Citations.Applied)
	if result.Flywheel != nil {
		sign := "+"
		if result.Flywheel.Velocity < 0 {
			sign = ""
		}
		fmt.Printf("Flywheel: %s (%s%.3f/week)\n", result.Flywheel.Status, sign, result.Flywheel.Velocity)
	}
	if result.State != nil {
		if result.State.LastStart != nil {
			fmt.Printf("Last start: %s\n", result.State.LastStart.Timestamp)
		}
		if result.State.LastStop != nil {
			fmt.Printf("Last stop: %s\n", result.State.LastStop.Timestamp)
		}
	}
	return nil
}

func printNamedItems[T any](heading string, items []T, label func(T) string) {
	fmt.Printf("%s:\n", heading)
	if len(items) == 0 {
		fmt.Println("  - none")
		return
	}
	for _, item := range items {
		fmt.Printf("  - %s\n", label(item))
	}
}
