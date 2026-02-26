package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/boshu2/agentops/cli/internal/storage"
)

var doctorJSON bool

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check AgentOps health",
	Long: `Run health checks on your AgentOps installation.

Validates that all required components are present and configured.
Optional components are reported as warnings but do not cause failure.

Examples:
  ao doctor
  ao doctor --json`,
	RunE: runDoctor,
}

func init() {
	doctorCmd.GroupID = "core"
	doctorCmd.Flags().BoolVar(&doctorJSON, "json", false, "Output results as JSON")
	rootCmd.AddCommand(doctorCmd)
}

type doctorCheck struct {
	Name     string `json:"name"`
	Status   string `json:"status"` // "pass", "warn", "fail"
	Detail   string `json:"detail"`
	Required bool   `json:"required"`
}

type doctorOutput struct {
	Checks  []doctorCheck `json:"checks"`
	Result  string        `json:"result"` // "HEALTHY", "DEGRADED", "UNHEALTHY"
	Summary string        `json:"summary"`
}

// gatherDoctorChecks runs all doctor checks and returns the results.
func gatherDoctorChecks() []doctorCheck {
	return []doctorCheck{
		{Name: "ao CLI", Status: "pass", Detail: formatVersion(version), Required: true},
		checkCLIDependencies(),
		checkHookCoverage(),
		checkKnowledgeBase(),
		checkKnowledgeFreshness(),
		checkSearchIndex(),
		checkFlywheelHealth(),
		checkSkills(),
		checkSkillIntegrity(),
		checkStaleReferences(),
		checkOptionalCLI("codex", "needed for --mixed council"),
	}
}

// doctorStatusIcon returns the display icon for a check status.
func doctorStatusIcon(status string) string {
	switch status {
	case "pass":
		return "\u2713"
	case "warn":
		return "!"
	case "fail":
		return "\u2717"
	}
	return "?"
}

// renderDoctorTable writes the formatted doctor output table.
func renderDoctorTable(w io.Writer, output doctorOutput) {
	_, _ = fmt.Fprintln(w, "ao doctor")
	_, _ = fmt.Fprintln(w, "\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500")

	maxName := 0
	for _, c := range output.Checks {
		if len(c.Name) > maxName {
			maxName = len(c.Name)
		}
	}

	for _, c := range output.Checks {
		padding := strings.Repeat(" ", maxName-len(c.Name))
		_, _ = fmt.Fprintf(w, "%s %s%s  %s\n", doctorStatusIcon(c.Status), c.Name, padding, c.Detail)
	}
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintf(w, "%s\n", output.Summary)
}

// hasRequiredFailure returns true if any required check has failed.
func hasRequiredFailure(checks []doctorCheck) bool {
	for _, c := range checks {
		if c.Required && c.Status == "fail" {
			return true
		}
	}
	return false
}

func runDoctor(cmd *cobra.Command, args []string) error {
	output := computeResult(gatherDoctorChecks())
	w := cmd.OutOrStdout()

	if doctorJSON {
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal doctor output: %w", err)
		}
		fmt.Fprintln(w, string(data))
		return nil
	}

	renderDoctorTable(w, output)

	if hasRequiredFailure(output.Checks) {
		return fmt.Errorf("doctor failed: one or more required checks did not pass")
	}

	return nil
}

// checkCLIDependencies verifies gt and bd are available in PATH.
func checkCLIDependencies() doctorCheck {
	gtOk := false
	bdOk := false

	if _, err := exec.LookPath("gt"); err == nil {
		gtOk = true
	}
	if _, err := exec.LookPath("bd"); err == nil {
		bdOk = true
	}

	if gtOk && bdOk {
		return doctorCheck{
			Name:     "CLI Dependencies",
			Status:   "pass",
			Detail:   "gt and bd available",
			Required: false,
		}
	}

	var missing []string
	var hints []string
	if !gtOk {
		missing = append(missing, "gt")
		hints = append(hints, "install with 'brew install gastown'")
	}
	if !bdOk {
		missing = append(missing, "bd")
		hints = append(hints, "install with 'brew install beads'")
	}

	return doctorCheck{
		Name:     "CLI Dependencies",
		Status:   "warn",
		Detail:   fmt.Sprintf("%s not found \u2014 %s", strings.Join(missing, ", "), strings.Join(hints, "; ")),
		Required: false,
	}
}

// checkHookCoverage checks if Claude hooks are installed with event coverage.
func checkHookCoverage() doctorCheck {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return doctorCheck{Name: "Hook Coverage", Status: "fail", Detail: "cannot determine home directory", Required: true}
	}
	contract := resolveHookCoverageContract()

	// Prefer settings.json (active Claude configuration).
	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")
	if data, err := os.ReadFile(settingsPath); err == nil {
		if hooksMap, ok := extractHooksMap(data); ok {
			return evaluateHookCoverageWithContract(hooksMap, contract)
		}
	}

	// Fallback: standalone hooks.json format.
	hooksPath := filepath.Join(homeDir, ".claude", "hooks.json")
	if data, err := os.ReadFile(hooksPath); err == nil {
		if hooksMap, ok := extractHooksMap(data); ok {
			return evaluateHookCoverageWithContract(hooksMap, contract)
		}
	}

	return doctorCheck{
		Name:     "Hook Coverage",
		Status:   "warn",
		Detail:   "No hooks found \u2014 run 'ao hooks install --force'" + hookCoverageFallbackDetail(contract.FallbackReason),
		Required: false,
	}
}

func evaluateHookCoverage(hooksMap map[string]any) doctorCheck {
	return evaluateHookCoverageWithContract(hooksMap, resolveHookCoverageContract())
}

func hookCoverageFallbackDetail(reason string) string {
	if reason == "" {
		return ""
	}
	return fmt.Sprintf(" (coverage contract fallback: %s)", reason)
}

func evaluateHookCoverageWithContract(hooksMap map[string]any, contract hookCoverageContract) doctorCheck {
	activeEvents := contract.ActiveEvents
	if len(activeEvents) == 0 {
		activeEvents = AllEventNames()
	}
	installedEvents := countInstalledEventsForList(hooksMap, activeEvents)
	fallbackSuffix := hookCoverageFallbackDetail(contract.FallbackReason)

	if installedEvents == 0 {
		return doctorCheck{
			Name:     "Hook Coverage",
			Status:   "warn",
			Detail:   "No hooks found \u2014 run 'ao hooks install --force'" + fallbackSuffix,
			Required: false,
		}
	}

	if !hookGroupContainsAo(hooksMap, "SessionStart") {
		return doctorCheck{
			Name:     "Hook Coverage",
			Status:   "warn",
			Detail:   "Non-ao hooks detected \u2014 run 'ao hooks install --force'" + fallbackSuffix,
			Required: false,
		}
	}

	if installedEvents < len(activeEvents) {
		return doctorCheck{
			Name:     "Hook Coverage",
			Status:   "warn",
			Detail:   fmt.Sprintf("Partial coverage: %d/%d events \u2014 run 'ao hooks install --force'%s", installedEvents, len(activeEvents), fallbackSuffix),
			Required: false,
		}
	}

	return doctorCheck{
		Name:     "Hook Coverage",
		Status:   "pass",
		Detail:   fmt.Sprintf("Full coverage: %d/%d events%s", installedEvents, len(activeEvents), fallbackSuffix),
		Required: false,
	}
}

func extractHooksMap(data []byte) (map[string]any, bool) {
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, false
	}

	// settings.json shape
	if hooksRaw, ok := parsed["hooks"]; ok {
		if hooksMap, ok := hooksRaw.(map[string]any); ok {
			return hooksMap, true
		}
	}

	// hooks.json shape with top-level events
	for _, event := range AllEventNames() {
		if _, ok := parsed[event]; ok {
			return parsed, true
		}
	}

	return nil, false
}

func countHooksInMap(raw any) int {
	count := 0
	switch v := raw.(type) {
	case map[string]any:
		for _, val := range v {
			if arr, ok := val.([]any); ok {
				count += len(arr)
			} else {
				// Recurse into nested maps
				count += countHooksInMap(val)
			}
		}
	case []any:
		count += len(v)
	}
	return count
}

func countInstalledEvents(hooksMap map[string]any) int {
	installed := 0
	for _, event := range AllEventNames() {
		if groups, ok := hooksMap[event].([]any); ok && len(groups) > 0 {
			installed++
		}
	}
	return installed
}

// checkKnowledgeBase checks that the .agents/ao directory exists.
func checkKnowledgeBase() doctorCheck {
	cwd, err := os.Getwd()
	if err != nil {
		return doctorCheck{Name: "Knowledge Base", Status: "fail", Detail: "cannot determine working directory", Required: true}
	}

	baseDir := filepath.Join(cwd, storage.DefaultBaseDir)
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return doctorCheck{Name: "Knowledge Base", Status: "fail", Detail: ".agents/ao not initialized", Required: true}
	}

	return doctorCheck{Name: "Knowledge Base", Status: "pass", Detail: ".agents/ao initialized", Required: true}
}

// newestFileModTime returns the most recent modification time among regular files in entries.
// Returns zero time if no regular files are found.
func newestFileModTime(entries []os.DirEntry) time.Time {
	var newest time.Time
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(newest) {
			newest = info.ModTime()
		}
	}
	return newest
}

// checkKnowledgeFreshness checks the most recent file in .agents/ao/sessions/.
func checkKnowledgeFreshness() doctorCheck {
	cwd, err := os.Getwd()
	if err != nil {
		return doctorCheck{Name: "Knowledge Freshness", Status: "warn", Detail: "cannot determine working directory", Required: false}
	}

	noSessionsCheck := doctorCheck{
		Name:     "Knowledge Freshness",
		Status:   "warn",
		Detail:   "No sessions found \u2014 run 'ao forge transcript' after your next session",
		Required: false,
	}

	sessionsDir := filepath.Join(cwd, storage.DefaultBaseDir, "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil || len(entries) == 0 {
		return noSessionsCheck
	}

	newest := newestFileModTime(entries)
	if newest.IsZero() {
		return noSessionsCheck
	}

	age := time.Since(newest)
	ageStr := formatDuration(age)

	if age > 14*24*time.Hour {
		return doctorCheck{
			Name:     "Knowledge Freshness",
			Status:   "warn",
			Detail:   fmt.Sprintf("Last session: %s ago \u2014 knowledge may be stale", ageStr),
			Required: false,
		}
	}

	return doctorCheck{
		Name:     "Knowledge Freshness",
		Status:   "pass",
		Detail:   fmt.Sprintf("Last session: %s ago", ageStr),
		Required: false,
	}
}

// formatVersion ensures the version string has exactly one "v" prefix.
func formatVersion(v string) string {
	if strings.HasPrefix(v, "v") {
		return v
	}
	return "v" + v
}

// formatDuration produces a human-readable duration string like "2h", "5d", "3m".
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	days := int(d.Hours() / 24)
	return fmt.Sprintf("%dd", days)
}

// checkSearchIndex checks if the search index exists and counts terms.
func checkSearchIndex() doctorCheck {
	cwd, err := os.Getwd()
	if err != nil {
		return doctorCheck{Name: "Search Index", Status: "warn", Detail: "cannot determine working directory", Required: false}
	}

	indexPath := filepath.Join(cwd, IndexDir, IndexFileName)
	info, err := os.Stat(indexPath)
	if err != nil {
		return doctorCheck{
			Name:     "Search Index",
			Status:   "warn",
			Detail:   "No search index \u2014 run 'ao store rebuild' for faster searches",
			Required: false,
		}
	}

	if info.Size() == 0 {
		return doctorCheck{
			Name:     "Search Index",
			Status:   "warn",
			Detail:   "Search index is empty \u2014 run 'ao store rebuild'",
			Required: false,
		}
	}

	// Count lines (each line is a term/entry)
	lines := countFileLines(indexPath)

	return doctorCheck{
		Name:     "Search Index",
		Status:   "pass",
		Detail:   fmt.Sprintf("Index exists (%s terms)", formatNumber(lines)),
		Required: false,
	}
}

// countFileLines counts non-empty lines in a file.
func countFileLines(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close() //nolint:errcheck // best-effort close

	count := 0
	scanner := bufio.NewScanner(f)
	// Increase buffer for potentially long JSONL lines
	scanner.Buffer(make([]byte, 256*1024), 1024*1024)
	for scanner.Scan() {
		if len(strings.TrimSpace(scanner.Text())) > 0 {
			count++
		}
	}
	return count
}

// formatNumber adds comma separators to an integer (e.g., 1247 -> "1,247").
func formatNumber(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}

	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}

// checkFlywheelHealth checks if .agents/ao/learnings/ has files.
// Counts .md and .jsonl files only, matching the metrics/badge counting method.
func checkFlywheelHealth() doctorCheck {
	cwd, err := os.Getwd()
	if err != nil {
		return doctorCheck{Name: "Flywheel Health", Status: "warn", Detail: "cannot determine working directory", Required: false}
	}

	learningsDir := filepath.Join(cwd, storage.DefaultBaseDir, "learnings")
	total := countLearningFiles(learningsDir)

	if total == 0 {
		// Also check the older path
		altDir := filepath.Join(cwd, ".agents", "learnings")
		total = countLearningFiles(altDir)
	}

	if total == 0 {
		return doctorCheck{
			Name:     "Flywheel Health",
			Status:   "warn",
			Detail:   "No learnings found \u2014 the flywheel hasn't started",
			Required: false,
		}
	}

	// Count established learnings (those with "established" or "promoted" in filename or content)
	established := countEstablished(filepath.Join(cwd, storage.DefaultBaseDir, "learnings"))
	if established == 0 {
		// Check alt path too
		established = countEstablished(filepath.Join(cwd, ".agents", "learnings"))
	}

	detail := fmt.Sprintf("%d learnings in flywheel", total)
	if established > 0 {
		detail = fmt.Sprintf("%d learnings (%d established)", total, established)
	}

	return doctorCheck{
		Name:     "Flywheel Health",
		Status:   "pass",
		Detail:   detail,
		Required: false,
	}
}

// countEstablished counts files in a directory whose name contains "established" or "promoted".
func countEstablished(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		lower := strings.ToLower(e.Name())
		if strings.Contains(lower, "established") || strings.Contains(lower, "promoted") {
			count++
		}
	}
	return count
}

func checkSkills() doctorCheck {
	// Skills are installed globally at ~/.claude/skills/, not in the local repo.
	// They may be symlinks pointing to ~/.agents/skills/.
	home, err := os.UserHomeDir()
	if err != nil {
		return doctorCheck{Name: "Plugin", Status: "warn", Detail: "cannot determine home directory", Required: false}
	}

	skillsDirs := []string{
		filepath.Join(home, ".codex", "skills"),
		filepath.Join(home, ".claude", "skills"),
		filepath.Join(home, ".agents", "skills"),
	}

	count := 0
	for _, skillsDir := range skillsDirs {
		entries, err := os.ReadDir(skillsDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			// Use os.Stat to follow symlinks (e.IsDir() doesn't follow symlinks)
			info, err := os.Stat(filepath.Join(skillsDir, e.Name()))
			if err != nil || !info.IsDir() {
				continue
			}
			skillFile := filepath.Join(skillsDir, e.Name(), "SKILL.md")
			if _, err := os.Stat(skillFile); err == nil {
				count++
			}
		}
		if count > 0 {
			break // Found skills in this directory, don't double-count
		}
	}

	if count == 0 {
		return doctorCheck{Name: "Plugin", Status: "warn", Detail: "no skills found — run 'npx skills@latest add <package> --all -g'", Required: false}
	}

	return doctorCheck{
		Name:     "Plugin",
		Status:   "pass",
		Detail:   fmt.Sprintf("%d skills found", count),
		Required: false,
	}
}

// findHealScript searches for heal.sh in known locations and returns the path if found.
func findHealScript() string {
	// 1. In-repo (when running from agentops checkout)
	if p := "skills/heal-skill/scripts/heal.sh"; fileExists(p) {
		abs, err := filepath.Abs(p)
		if err == nil {
			return abs
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// 2. Installed via npx skills (Codex-native location)
	if p := filepath.Join(home, ".codex", "skills", "heal-skill", "scripts", "heal.sh"); fileExists(p) {
		return p
	}

	// 3. Installed via npx skills (Claude location)
	if p := filepath.Join(home, ".claude", "skills", "heal-skill", "scripts", "heal.sh"); fileExists(p) {
		return p
	}

	// 4. Alt install location
	if p := filepath.Join(home, ".agents", "skills", "heal-skill", "scripts", "heal.sh"); fileExists(p) {
		return p
	}

	return ""
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// checkSkillIntegrity runs heal.sh --strict to validate skill hygiene.
func checkSkillIntegrity() doctorCheck {
	healPath := findHealScript()
	if healPath == "" {
		return doctorCheck{
			Name:     "Skill Integrity",
			Status:   "warn",
			Detail:   "heal-skill not installed, skipping integrity check",
			Required: false,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", healPath, "--strict")
	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return doctorCheck{
			Name:     "Skill Integrity",
			Status:   "warn",
			Detail:   "heal.sh timed out after 30s",
			Required: false,
		}
	}

	if err == nil {
		return doctorCheck{
			Name:     "Skill Integrity",
			Status:   "pass",
			Detail:   "All skill integrity checks passed",
			Required: false,
		}
	}

	// --strict exits 1 when findings exist — count them
	findings := countHealFindings(string(output))
	return doctorCheck{
		Name:     "Skill Integrity",
		Status:   "warn",
		Detail:   fmt.Sprintf("%d skill hygiene finding(s) \u2014 run 'heal.sh --check' for details", findings),
		Required: false,
	}
}

// countHealFindings counts lines matching the heal.sh report format: [CODE] path: message
func countHealFindings(output string) int {
	count := 0
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 0 && strings.HasPrefix(trimmed, "[") && strings.Contains(trimmed, "]") {
			count++
		}
	}
	if count == 0 {
		// Fallback: count from the summary line "N finding(s) detected."
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, "finding(s) detected") {
				_, _ = fmt.Sscanf(strings.TrimSpace(line), "%d", &count)
				break
			}
		}
	}
	return count
}

// deprecatedCommands maps old namespace-qualified command references to their
// new flat replacements. Used by checkStaleReferences to detect lingering
// namespace references in hooks and skill files.
var deprecatedCommands = map[string]string{
	"ao know forge":              "ao forge",
	"ao know inject":             "ao inject",
	"ao know search":             "ao search",
	"ao know lookup":             "ao lookup",
	"ao know trace":              "ao trace",
	"ao know store":              "ao store",
	"ao know index":              "ao index",
	"ao know temper":             "ao temper",
	"ao know feedback":           "ao feedback",
	"ao know migrate":            "ao migrate",
	"ao know batch-feedback":     "ao batch-feedback",
	"ao know session-outcome":    "ao session-outcome",
	"ao work rpi":                "ao rpi",
	"ao work ratchet":            "ao ratchet",
	"ao work goals":              "ao goals",
	"ao work session":            "ao session",
	"ao work feedback-loop":      "ao feedback-loop",
	"ao work context":            "ao context",
	"ao work task-sync":          "ao task-sync",
	"ao work task-feedback":      "ao task-feedback",
	"ao work task-status":        "ao task-status",
	"ao quality flywheel":        "ao flywheel",
	"ao quality pool":            "ao pool",
	"ao quality metrics":         "ao metrics",
	"ao quality gate":            "ao gate",
	"ao quality maturity":        "ao maturity",
	"ao quality constraint":      "ao constraint",
	"ao quality vibe-check":      "ao vibe-check",
	"ao quality badge":           "ao badge",
	"ao quality contradict":      "ao contradict",
	"ao quality dedup":           "ao dedup",
	"ao quality anti-patterns":   "ao anti-patterns",
	"ao quality curate":          "ao curate",
	"ao quality promote-anti-patterns": "ao promote-anti-patterns",
	"ao settings config":         "ao config",
	"ao settings plans":          "ao plans",
	"ao settings hooks":          "ao hooks",
	"ao settings memory":         "ao memory",
	"ao settings notebook":       "ao notebook",
	"ao settings worktree":       "ao worktree",
	"ao start demo":              "ao demo",
	"ao start init":              "ao init",
	"ao start seed":              "ao seed",
	"ao start quick-start":       "ao quick-start",
}

// staleReference records a single deprecated command reference found in a file.
type staleReference struct {
	File       string `json:"file"`
	OldCommand string `json:"old_command"`
	NewCommand string `json:"new_command"`
}

// checkStaleReferences scans hooks/*.sh, hooks/examples/*.sh,
// cli/embedded/hooks/*.sh, skills/*/SKILL.md, docs/*.md, docs/contracts/*.md,
// docs/plans/*.md, and scripts/*.sh for deprecated command references and
// reports them as warnings.
func checkStaleReferences() doctorCheck {
	var refs []staleReference

	// Scan hooks/*.sh
	hookFiles, _ := filepath.Glob("hooks/*.sh")
	for _, f := range hookFiles {
		found := scanFileForDeprecatedCommands(f)
		refs = append(refs, found...)
	}

	// Scan skills/*/SKILL.md
	skillFiles, _ := filepath.Glob("skills/*/SKILL.md")
	for _, f := range skillFiles {
		found := scanFileForDeprecatedCommands(f)
		refs = append(refs, found...)
	}

	// Scan docs/*.md
	docFiles, _ := filepath.Glob("docs/*.md")
	for _, f := range docFiles {
		found := scanFileForDeprecatedCommands(f)
		refs = append(refs, found...)
	}

	// Scan scripts/*.sh
	scriptFiles, _ := filepath.Glob("scripts/*.sh")
	for _, f := range scriptFiles {
		found := scanFileForDeprecatedCommands(f)
		refs = append(refs, found...)
	}

	// Scan hooks/examples/*.sh
	exampleHookFiles, _ := filepath.Glob("hooks/examples/*.sh")
	for _, f := range exampleHookFiles {
		found := scanFileForDeprecatedCommands(f)
		refs = append(refs, found...)
	}

	// Scan cli/embedded/hooks/*.sh
	embeddedHookFiles, _ := filepath.Glob("cli/embedded/hooks/*.sh")
	for _, f := range embeddedHookFiles {
		found := scanFileForDeprecatedCommands(f)
		refs = append(refs, found...)
	}

	// Scan docs/contracts/*.md
	contractDocFiles, _ := filepath.Glob("docs/contracts/*.md")
	for _, f := range contractDocFiles {
		found := scanFileForDeprecatedCommands(f)
		refs = append(refs, found...)
	}

	// Scan docs/plans/*.md
	planDocFiles, _ := filepath.Glob("docs/plans/*.md")
	for _, f := range planDocFiles {
		found := scanFileForDeprecatedCommands(f)
		refs = append(refs, found...)
	}

	if len(refs) == 0 {
		return doctorCheck{
			Name:     "Stale References",
			Status:   "pass",
			Detail:   "No deprecated command references found",
			Required: false,
		}
	}

	// Build a summary of unique old commands found
	seen := make(map[string]bool)
	for _, r := range refs {
		seen[r.OldCommand] = true
	}
	cmds := make([]string, 0, len(seen))
	for cmd := range seen {
		cmds = append(cmds, cmd)
	}

	detail := fmt.Sprintf("%d stale reference(s) in %d file(s)", len(refs), countUniqueFiles(refs))
	if len(cmds) <= 3 {
		detail += fmt.Sprintf(" — update: %s", strings.Join(cmds, ", "))
	}

	return doctorCheck{
		Name:     "Stale References",
		Status:   "warn",
		Detail:   detail,
		Required: false,
	}
}

// scanFileForDeprecatedCommands reads a file and checks each line for
// deprecated command patterns (old namespace-qualified commands like
// "ao work rpi" that should be replaced with flat "ao rpi").
func scanFileForDeprecatedCommands(path string) []staleReference {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var refs []staleReference
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		for oldCmd, newCmd := range deprecatedCommands {
			idx := strings.Index(line, oldCmd)
			if idx < 0 {
				continue
			}
			// Check the character after the match to avoid false positives.
			// e.g., "ao work rpi" should not match "ao work rpi-something"
			afterIdx := idx + len(oldCmd)
			if afterIdx < len(line) {
				ch := line[afterIdx]
				if ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch == '-' {
					continue
				}
			}

			refs = append(refs, staleReference{
				File:       path,
				OldCommand: oldCmd,
				NewCommand: newCmd,
			})
			// Only report each deprecated command once per file
			break
		}
	}

	return refs
}

// countUniqueFiles counts the number of distinct files in a slice of staleReferences.
func countUniqueFiles(refs []staleReference) int {
	seen := make(map[string]bool)
	for _, r := range refs {
		seen[r.File] = true
	}
	return len(seen)
}

func checkOptionalCLI(name string, reason string) doctorCheck {
	_, err := exec.LookPath(name)
	if err != nil {
		return doctorCheck{
			Name:     strings.Title(name) + " CLI", //nolint:staticcheck
			Status:   "warn",
			Detail:   fmt.Sprintf("not found (optional \u2014 %s)", reason),
			Required: false,
		}
	}

	return doctorCheck{
		Name:     strings.Title(name) + " CLI", //nolint:staticcheck
		Status:   "pass",
		Detail:   "available",
		Required: false,
	}
}

func countFiles(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() {
			count++
		}
	}
	return count
}

// countLearningFiles counts .md and .jsonl files in a directory,
// matching the counting method used by countArtifacts in metrics.go.
func countLearningFiles(dir string) int {
	mdFiles, _ := filepath.Glob(filepath.Join(dir, "*.md"))
	jsonlFiles, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	return len(mdFiles) + len(jsonlFiles)
}

// countCheckStatuses tallies pass, fail, and warn counts from checks.
func countCheckStatuses(checks []doctorCheck) (passes, fails, warns int) {
	for _, c := range checks {
		switch c.Status {
		case "pass":
			passes++
		case "fail":
			fails++
		case "warn":
			warns++
		}
	}
	return passes, fails, warns
}

// buildDoctorSummary constructs a human-readable summary from check tallies.
func buildDoctorSummary(passes, fails, warns, total int) string {
	switch {
	case fails == 0 && warns == 0:
		return fmt.Sprintf("%d/%d checks passed", passes, total)
	case fails == 0:
		summary := fmt.Sprintf("%d/%d checks passed, %d warning", passes, total, warns)
		if warns > 1 {
			summary += "s"
		}
		return summary
	default:
		parts := []string{fmt.Sprintf("%d/%d checks passed", passes, total)}
		if warns > 0 {
			w := fmt.Sprintf("%d warning", warns)
			if warns > 1 {
				w += "s"
			}
			parts = append(parts, w)
		}
		if fails > 0 {
			f := fmt.Sprintf("%d failed", fails)
			parts = append(parts, f)
		}
		return strings.Join(parts, ", ")
	}
}

func computeResult(checks []doctorCheck) doctorOutput {
	passes, fails, warns := countCheckStatuses(checks)
	total := len(checks)

	result := "HEALTHY"
	if fails > 0 {
		result = "UNHEALTHY"
	} else if warns > 0 {
		result = "DEGRADED"
	}

	return doctorOutput{
		Checks:  checks,
		Result:  result,
		Summary: buildDoctorSummary(passes, fails, warns, total),
	}
}
