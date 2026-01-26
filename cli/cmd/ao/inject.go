package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/boshu2/agentops/cli/internal/ratchet"
	"github.com/boshu2/agentops/cli/internal/types"
)

const (
	// DefaultInjectMaxTokens is the default token budget for injection (~1500 tokens ≈ 6KB)
	DefaultInjectMaxTokens = 1500

	// InjectCharsPerToken is the approximate characters per token (conservative estimate)
	InjectCharsPerToken = 4

	// MaxLearningsToInject is the maximum number of learnings to include
	MaxLearningsToInject = 10

	// MaxPatternsToInject is the maximum number of patterns to include
	MaxPatternsToInject = 5

	// MaxSessionsToInject is the maximum number of recent sessions to summarize
	MaxSessionsToInject = 5
)

var (
	injectMaxTokens int
	injectContext   string
	injectFormat    string
	injectSessionID string
	injectNoCite    bool
	injectApplyDecay bool
)

type injectedKnowledge struct {
	Learnings []learning `json:"learnings,omitempty"`
	Patterns  []pattern  `json:"patterns,omitempty"`
	Sessions  []session  `json:"sessions,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
	Query     string     `json:"query,omitempty"`
}

type learning struct {
	ID             string  `json:"id"`
	Title          string  `json:"title"`
	Summary        string  `json:"summary"`
	Source         string  `json:"source,omitempty"`
	FreshnessScore float64 `json:"freshness_score,omitempty"`
	AgeWeeks       float64 `json:"age_weeks,omitempty"`
	Utility        float64 `json:"utility,omitempty"`         // MemRL utility value
	CompositeScore float64 `json:"composite_score,omitempty"` // Two-Phase ranking score
	Superseded     bool    `json:"-"`                         // Internal flag - not serialized
}

type pattern struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	FilePath    string `json:"file_path,omitempty"`
}

type session struct {
	Date    string `json:"date"`
	Summary string `json:"summary"`
}

var injectCmd = &cobra.Command{
	Use:   "inject [context]",
	Short: "Output relevant knowledge for session injection",
	Long: `Inject searches and outputs relevant knowledge for session startup.

This command is designed to be called from a SessionStart hook to
inject prior learnings, patterns, and context into new sessions.

Searches:
  1. Recent learnings (.agents/learnings/*.md)
  2. Active patterns (.agents/patterns/*.md)
  3. Recent session summaries (.agents/ao/sessions/)

Uses file-based search with Two-Phase retrieval (freshness + utility scoring).
CASS integration adds maturity weighting and confidence decay.

Examples:
  ao inject                     # Inject general knowledge
  ao inject "authentication"    # Inject knowledge about auth
  ao inject --max-tokens 2000   # Larger budget
  ao inject --format json       # JSON output
  ao inject --no-cite           # Skip citation recording
  ao inject --apply-decay       # Apply confidence decay before ranking`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInject,
}

func init() {
	rootCmd.AddCommand(injectCmd)
	injectCmd.Flags().IntVar(&injectMaxTokens, "max-tokens", DefaultInjectMaxTokens, "Maximum tokens to output")
	injectCmd.Flags().StringVar(&injectContext, "context", "", "Context query for filtering (alternative to positional arg)")
	injectCmd.Flags().StringVar(&injectFormat, "format", "markdown", "Output format: markdown, json")
	injectCmd.Flags().StringVar(&injectSessionID, "session", "", "Session ID for citation tracking (auto-generated if empty)")
	injectCmd.Flags().BoolVar(&injectNoCite, "no-cite", false, "Disable citation recording")
	injectCmd.Flags().BoolVar(&injectApplyDecay, "apply-decay", false, "Apply confidence decay before ranking")
}

func runInject(cmd *cobra.Command, args []string) error {
	// Get context query
	query := injectContext
	if len(args) > 0 {
		query = args[0]
	}

	if GetDryRun() {
		fmt.Printf("[dry-run] Would inject knowledge")
		if query != "" {
			fmt.Printf(" filtered by: %s", query)
		}
		fmt.Printf(" (max %d tokens)\n", injectMaxTokens)
		return nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	// Calculate character budget from tokens
	charBudget := injectMaxTokens * InjectCharsPerToken

	// Get or generate session ID for citation tracking
	sessionID := canonicalSessionID(injectSessionID)

	// Collect knowledge
	knowledge := &injectedKnowledge{
		Timestamp: time.Now(),
		Query:     query,
	}

	// Search learnings
	learnings, err := collectLearnings(cwd, query, MaxLearningsToInject)
	if err != nil {
		VerbosePrintf("Warning: failed to collect learnings: %v\n", err)
	}
	knowledge.Learnings = learnings

	// Record citations for retrieved learnings (Phase 0: Critical for MemRL feedback loop)
	if !injectNoCite && len(learnings) > 0 {
		if err := recordCitations(cwd, learnings, sessionID, query); err != nil {
			VerbosePrintf("Warning: failed to record citations: %v\n", err)
		} else {
			VerbosePrintf("Recorded %d citations for session %s\n", len(learnings), sessionID)
		}
	}

	// Search patterns
	patterns, err := collectPatterns(cwd, query, MaxPatternsToInject)
	if err != nil {
		VerbosePrintf("Warning: failed to collect patterns: %v\n", err)
	}
	knowledge.Patterns = patterns

	// Search recent sessions
	sessions, err := collectRecentSessions(cwd, query, MaxSessionsToInject)
	if err != nil {
		VerbosePrintf("Warning: failed to collect sessions: %v\n", err)
	}
	knowledge.Sessions = sessions

	// Format output
	var output string
	if injectFormat == "json" {
		data, err := json.MarshalIndent(knowledge, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal json: %w", err)
		}
		output = string(data)
	} else {
		output = formatKnowledgeMarkdown(knowledge)
	}

	// Trim to budget if needed
	if len(output) > charBudget {
		output = trimToCharBudget(output, charBudget)
	}

	fmt.Println(output)
	return nil
}

// collectLearnings finds recent learnings from .agents/learnings/
// Implements MemRL Two-Phase retrieval: Phase A (similarity/freshness) + Phase B (utility-weighted)
// With CASS integration: applies confidence decay when --apply-decay is set
func collectLearnings(cwd, query string, limit int) ([]learning, error) {
	learningsDir := filepath.Join(cwd, ".agents", "learnings")
	if _, err := os.Stat(learningsDir); os.IsNotExist(err) {
		// Try rig root
		learningsDir = findAgentsSubdir(cwd, "learnings")
		if learningsDir == "" {
			return nil, nil // No learnings directory
		}
	}

	files, err := filepath.Glob(filepath.Join(learningsDir, "*.md"))
	if err != nil {
		return nil, err
	}

	// Also check .jsonl files
	jsonlFiles, _ := filepath.Glob(filepath.Join(learningsDir, "*.jsonl"))
	files = append(files, jsonlFiles...)

	var learnings []learning
	queryLower := strings.ToLower(query)
	now := time.Now()

	for _, file := range files {
		l, err := parseLearningFile(file)
		if err != nil {
			continue
		}

		// F3: Skip superseded learnings (superseded_by field set)
		if l.Superseded {
			VerbosePrintf("Skipping superseded learning: %s\n", l.ID)
			continue
		}

		// Filter by query if provided
		if query != "" {
			content := strings.ToLower(l.Title + " " + l.Summary)
			if !strings.Contains(content, queryLower) {
				continue
			}
		}

		// Calculate freshness score: exp(-ageWeeks * decayRate)
		// decayRate = 0.17/week (literature default)
		info, _ := os.Stat(file)
		if info != nil {
			ageHours := now.Sub(info.ModTime()).Hours()
			ageWeeks := ageHours / (24 * 7)
			l.AgeWeeks = ageWeeks
			l.FreshnessScore = freshnessScore(ageWeeks)
		} else {
			l.FreshnessScore = 0.5 // Default for missing stat
		}

		// Set default utility if not set (for markdown files)
		if l.Utility == 0 {
			l.Utility = types.InitialUtility
		}

		// Apply confidence decay if requested (CASS feature)
		if injectApplyDecay {
			l = applyConfidenceDecay(l, file, now)
		}

		learnings = append(learnings, l)
	}

	// Phase B: Calculate composite scores with z-normalization
	// Score = z_norm(freshness) + λ × z_norm(utility)
	applyCompositeScoring(learnings, types.DefaultLambda)

	// Sort by composite score (highest first) - Two-Phase retrieval
	sort.Slice(learnings, func(i, j int) bool {
		return learnings[i].CompositeScore > learnings[j].CompositeScore
	})

	// Limit results
	if len(learnings) > limit {
		learnings = learnings[:limit]
	}

	return learnings, nil
}

// applyConfidenceDecay applies time-based confidence decay to a learning.
// Confidence decays at 10%/week for learnings that haven't received recent feedback.
// Formula: confidence *= exp(-weeks_since_last_feedback * ConfidenceDecayRate)
func applyConfidenceDecay(l learning, filePath string, now time.Time) learning {
	// Read the file to get last_decay_at and confidence
	content, err := os.ReadFile(filePath)
	if err != nil {
		return l
	}

	// Parse to extract CASS fields
	if strings.HasSuffix(filePath, ".jsonl") {
		lines := strings.Split(string(content), "\n")
		if len(lines) == 0 {
			return l
		}

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(lines[0]), &data); err != nil {
			return l
		}

		// Get confidence (default to 0.5)
		confidence := 0.5
		if c, ok := data["confidence"].(float64); ok && c > 0 {
			confidence = c
		}

		// Get last_decay_at or last_reward_at
		var lastInteraction time.Time
		if lda, ok := data["last_decay_at"].(string); ok && lda != "" {
			lastInteraction, _ = time.Parse(time.RFC3339, lda)
		} else if lra, ok := data["last_reward_at"].(string); ok && lra != "" {
			lastInteraction, _ = time.Parse(time.RFC3339, lra)
		}

		// Calculate decay
		if !lastInteraction.IsZero() {
			weeksSinceInteraction := now.Sub(lastInteraction).Hours() / (24 * 7)
			if weeksSinceInteraction > 0 {
				// Apply decay: confidence *= exp(-weeks * decayRate)
				decayFactor := math.Exp(-weeksSinceInteraction * types.ConfidenceDecayRate)
				newConfidence := confidence * decayFactor

				// Clamp to minimum of 0.1
				if newConfidence < 0.1 {
					newConfidence = 0.1
				}

				VerbosePrintf("Applied confidence decay to %s: %.3f -> %.3f (%.1f weeks)\n",
					l.ID, confidence, newConfidence, weeksSinceInteraction)

				// Update the learning's composite score weight
				// (actual file update happens in separate decay command)
				l.Utility = l.Utility * (newConfidence / confidence)
			}
		}
	}

	return l
}

// freshnessScore calculates decay-adjusted score: exp(-ageWeeks * decayRate)
// Based on knowledge decay rate δ = 0.17/week (Darr et al.)
func freshnessScore(ageWeeks float64) float64 {
	const decayRate = 0.17
	score := math.Exp(-ageWeeks * decayRate)
	// Clamp to [0.1, 1.0] - old knowledge still has some value
	if score < 0.1 {
		return 0.1
	}
	return score
}

// applyCompositeScoring implements MemRL Two-Phase scoring.
// Score = z_norm(freshness) + λ × z_norm(utility)
// This combines recency (Phase A) with learned utility (Phase B).
func applyCompositeScoring(learnings []learning, lambda float64) {
	if len(learnings) == 0 {
		return
	}

	// Calculate means and standard deviations for z-normalization
	var sumF, sumU float64
	for _, l := range learnings {
		sumF += l.FreshnessScore
		sumU += l.Utility
	}
	n := float64(len(learnings))
	meanF := sumF / n
	meanU := sumU / n

	// Calculate standard deviations
	var varF, varU float64
	for _, l := range learnings {
		varF += (l.FreshnessScore - meanF) * (l.FreshnessScore - meanF)
		varU += (l.Utility - meanU) * (l.Utility - meanU)
	}
	stdF := math.Sqrt(varF / n)
	stdU := math.Sqrt(varU / n)

	// Avoid division by zero - use minimum of 0.001
	if stdF < 0.001 {
		stdF = 0.001
	}
	if stdU < 0.001 {
		stdU = 0.001
	}

	// Apply z-normalization and calculate composite scores
	for i := range learnings {
		zFresh := (learnings[i].FreshnessScore - meanF) / stdF
		zUtility := (learnings[i].Utility - meanU) / stdU

		// Composite score: z_norm(freshness) + λ × z_norm(utility)
		learnings[i].CompositeScore = zFresh + lambda*zUtility
	}
}

// frontMatter holds parsed YAML front matter fields
type frontMatter struct {
	SupersededBy string
}

// parseFrontMatter extracts YAML front matter from markdown content
func parseFrontMatter(lines []string) (frontMatter, int) {
	var fm frontMatter
	endLine := 0

	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return fm, 0
	}

	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "---" {
			endLine = i + 1
			break
		}
		if strings.HasPrefix(line, "superseded_by:") || strings.HasPrefix(line, "superseded-by:") {
			fm.SupersededBy = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
		}
	}
	return fm, endLine
}

// extractSummary finds the first paragraph after headings
func extractSummary(lines []string, startIdx int) string {
	for i := startIdx; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "---") {
			continue
		}
		// Take first paragraph (up to 3 lines)
		summary := line
		for j := i + 1; j < len(lines) && j < i+3; j++ {
			nextLine := strings.TrimSpace(lines[j])
			if nextLine == "" || strings.HasPrefix(nextLine, "#") {
				break
			}
			summary += " " + nextLine
		}
		return truncateText(summary, 200)
	}
	return ""
}

// parseLearningFile extracts learning info from a file
// Sets Superseded=true if superseded_by field is found
func parseLearningFile(path string) (learning, error) {
	l := learning{
		ID:     filepath.Base(path),
		Source: path,
	}

	if strings.HasSuffix(path, ".jsonl") {
		return parseLearningJSONL(path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return l, err
	}

	lines := strings.Split(string(content), "\n")

	// Parse front matter
	fm, contentStart := parseFrontMatter(lines)
	if fm.SupersededBy != "" && fm.SupersededBy != "null" && fm.SupersededBy != "~" {
		l.Superseded = true
		return l, nil
	}

	// Parse body content
	for i := contentStart; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		if strings.HasPrefix(line, "# ") && l.Title == "" {
			l.Title = strings.TrimPrefix(line, "# ")
		} else if (strings.HasPrefix(line, "ID:") || strings.HasPrefix(line, "id:")) && l.ID == filepath.Base(path) {
			l.ID = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
		}
	}

	l.Summary = extractSummary(lines, contentStart)

	if l.Title == "" {
		l.Title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}

	return l, nil
}

// parseLearningJSONL extracts learning from JSONL file
// Returns empty learning (with Superseded=true) if superseded_by field is set
func parseLearningJSONL(path string) (learning, error) {
	l := learning{
		ID:      filepath.Base(path),
		Source:  path,
		Utility: types.InitialUtility, // Default to 0.5
	}

	f, err := os.Open(path)
	if err != nil {
		return l, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		var data map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &data); err == nil {
			// F3: Filter superseded learnings - skip if superseded_by is set
			if supersededBy, ok := data["superseded_by"]; ok && supersededBy != nil && supersededBy != "" {
				l.Superseded = true
				return l, nil // Return early, will be filtered out
			}

			if id, ok := data["id"].(string); ok {
				l.ID = id
			}
			if title, ok := data["title"].(string); ok {
				l.Title = title
			}
			if summary, ok := data["summary"].(string); ok {
				l.Summary = truncateText(summary, 200)
			}
			if content, ok := data["content"].(string); ok && l.Summary == "" {
				l.Summary = truncateText(content, 200)
			}
			// Parse MemRL utility value
			if utility, ok := data["utility"].(float64); ok && utility > 0 {
				l.Utility = utility
			}
		}
	}

	return l, nil
}

// collectPatterns finds patterns from .agents/patterns/
func collectPatterns(cwd, query string, limit int) ([]pattern, error) {
	patternsDir := filepath.Join(cwd, ".agents", "patterns")
	if _, err := os.Stat(patternsDir); os.IsNotExist(err) {
		// Try rig root
		patternsDir = findAgentsSubdir(cwd, "patterns")
		if patternsDir == "" {
			return nil, nil
		}
	}

	files, err := filepath.Glob(filepath.Join(patternsDir, "*.md"))
	if err != nil {
		return nil, err
	}

	// Sort by modification time
	sort.Slice(files, func(i, j int) bool {
		infoI, _ := os.Stat(files[i])
		infoJ, _ := os.Stat(files[j])
		if infoI == nil || infoJ == nil {
			return false
		}
		return infoI.ModTime().After(infoJ.ModTime())
	})

	var patterns []pattern
	queryLower := strings.ToLower(query)

	for _, file := range files {
		if len(patterns) >= limit {
			break
		}

		p, err := parsePatternFile(file)
		if err != nil {
			continue
		}

		// Filter by query
		if query != "" {
			content := strings.ToLower(p.Name + " " + p.Description)
			if !strings.Contains(content, queryLower) {
				continue
			}
		}

		patterns = append(patterns, p)
	}

	return patterns, nil
}

// parsePatternFile extracts pattern info from a markdown file
func parsePatternFile(path string) (pattern, error) {
	p := pattern{
		Name:     strings.TrimSuffix(filepath.Base(path), ".md"),
		FilePath: path,
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return p, err
	}

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Extract name from title
		if strings.HasPrefix(line, "# ") {
			p.Name = strings.TrimPrefix(line, "# ")
			continue
		}

		// First paragraph as description
		if p.Description == "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "---") && line != "" {
			desc := line
			for j := i + 1; j < len(lines) && j < i+2; j++ {
				nextLine := strings.TrimSpace(lines[j])
				if nextLine == "" || strings.HasPrefix(nextLine, "#") {
					break
				}
				desc += " " + nextLine
			}
			p.Description = truncateText(desc, 150)
			break
		}
	}

	return p, nil
}

// collectRecentSessions finds recent session summaries
func collectRecentSessions(cwd, query string, limit int) ([]session, error) {
	sessionsDir := filepath.Join(cwd, ".agents", "ao", "sessions")
	if _, err := os.Stat(sessionsDir); os.IsNotExist(err) {
		return nil, nil
	}

	files, err := filepath.Glob(filepath.Join(sessionsDir, "*.jsonl"))
	if err != nil {
		return nil, err
	}

	// Also include markdown summaries
	mdFiles, _ := filepath.Glob(filepath.Join(sessionsDir, "*.md"))
	files = append(files, mdFiles...)

	// Sort by modification time (newest first)
	sort.Slice(files, func(i, j int) bool {
		infoI, _ := os.Stat(files[i])
		infoJ, _ := os.Stat(files[j])
		if infoI == nil || infoJ == nil {
			return false
		}
		return infoI.ModTime().After(infoJ.ModTime())
	})

	var sessions []session
	queryLower := strings.ToLower(query)

	for _, file := range files {
		if len(sessions) >= limit {
			break
		}

		s, err := parseSessionFile(file)
		if err != nil || s.Summary == "" {
			continue
		}

		// Filter by query
		if query != "" && !strings.Contains(strings.ToLower(s.Summary), queryLower) {
			continue
		}

		sessions = append(sessions, s)
	}

	return sessions, nil
}

// parseSessionFile extracts session summary from a file
func parseSessionFile(path string) (session, error) {
	s := session{}

	info, err := os.Stat(path)
	if err != nil {
		return s, err
	}
	s.Date = info.ModTime().Format("2006-01-02")

	if strings.HasSuffix(path, ".jsonl") {
		f, err := os.Open(path)
		if err != nil {
			return s, err
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		if scanner.Scan() {
			var data map[string]interface{}
			if err := json.Unmarshal(scanner.Bytes(), &data); err == nil {
				if summary, ok := data["summary"].(string); ok {
					s.Summary = truncateText(summary, 150)
				}
			}
		}
	} else {
		// Markdown - extract first paragraph
		content, err := os.ReadFile(path)
		if err != nil {
			return s, err
		}
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "---") {
				s.Summary = truncateText(line, 150)
				break
			}
		}
	}

	return s, nil
}

// findAgentsSubdir looks for .agents/{subdir}/ walking up to rig root
func findAgentsSubdir(startDir, subdir string) string {
	dir := startDir
	for {
		candidate := filepath.Join(dir, ".agents", subdir)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}

		// Check if we're at rig root (has .beads, crew, or polecats)
		markers := []string{".beads", "crew", "polecats"}
		atRigRoot := false
		for _, marker := range markers {
			if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
				atRigRoot = true
				break
			}
		}
		if atRigRoot {
			break
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// formatKnowledgeMarkdown formats knowledge as markdown
func formatKnowledgeMarkdown(k *injectedKnowledge) string {
	var sb strings.Builder

	sb.WriteString("## Injected Knowledge (ol inject)\n\n")

	if len(k.Learnings) > 0 {
		sb.WriteString("### Recent Learnings\n")
		for _, l := range k.Learnings {
			if l.Summary != "" {
				sb.WriteString(fmt.Sprintf("- **%s**: %s\n", l.ID, l.Summary))
			} else {
				sb.WriteString(fmt.Sprintf("- **%s**: %s\n", l.ID, l.Title))
			}
		}
		sb.WriteString("\n")
	}

	if len(k.Patterns) > 0 {
		sb.WriteString("### Active Patterns\n")
		for _, p := range k.Patterns {
			if p.Description != "" {
				sb.WriteString(fmt.Sprintf("- **%s**: %s\n", p.Name, p.Description))
			} else {
				sb.WriteString(fmt.Sprintf("- **%s**\n", p.Name))
			}
		}
		sb.WriteString("\n")
	}

	if len(k.Sessions) > 0 {
		sb.WriteString("### Recent Sessions\n")
		for _, s := range k.Sessions {
			sb.WriteString(fmt.Sprintf("- [%s] %s\n", s.Date, s.Summary))
		}
		sb.WriteString("\n")
	}

	if len(k.Learnings) == 0 && len(k.Patterns) == 0 && len(k.Sessions) == 0 {
		sb.WriteString("*No prior knowledge found.*\n\n")
	}

	sb.WriteString(fmt.Sprintf("*Last injection: %s*\n", k.Timestamp.Format(time.RFC3339)))

	return sb.String()
}

// trimToCharBudget truncates output to fit character budget
func trimToCharBudget(output string, budget int) string {
	if len(output) <= budget {
		return output
	}

	// Try to truncate at a section boundary
	lines := strings.Split(output, "\n")
	var result strings.Builder
	for _, line := range lines {
		if result.Len()+len(line)+1 > budget-50 { // Leave room for truncation marker
			break
		}
		result.WriteString(line)
		result.WriteString("\n")
	}

	result.WriteString("\n*[truncated to fit token budget]*\n")
	return result.String()
}

// truncateText truncates a string to max length with ellipsis
func truncateText(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// canonicalSessionID normalizes session IDs to a consistent format.
// Addresses pre-mortem C2: session ID format mismatch causing zero citation matches.
// Format: session-YYYYMMDD-HHMMSS (auto-generated if empty or random string).
func canonicalSessionID(raw string) string {
	if raw == "" {
		// Generate new session ID with timestamp
		return fmt.Sprintf("session-%s", time.Now().Format("20060102-150405"))
	}

	// Check if already in canonical format
	canonicalPattern := regexp.MustCompile(`^session-\d{8}-\d{6}$`)
	if canonicalPattern.MatchString(raw) {
		return raw
	}

	// Check for UUID format (e.g., from Claude sessions)
	uuidPattern := regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)
	if uuidPattern.MatchString(raw) {
		// Convert UUID to canonical by prepending "session-" and using current timestamp
		return fmt.Sprintf("session-%s", time.Now().Format("20060102-150405"))
	}

	// Return as-is for other formats (e.g., user-provided IDs)
	return raw
}

// recordCitations records citation events for retrieved learnings.
// This is critical for closing the MemRL feedback loop (Phase 0).
// Citations link: session → learning → feedback → utility update.
func recordCitations(baseDir string, learnings []learning, sessionID, query string) error {
	for _, l := range learnings {
		event := types.CitationEvent{
			ArtifactPath: l.Source,
			SessionID:    sessionID,
			CitedAt:      time.Now(),
			CitationType: "retrieved", // Will be upgraded to "applied" if session succeeds
			Query:        query,
		}

		if err := ratchet.RecordCitation(baseDir, event); err != nil {
			return fmt.Errorf("record citation for %s: %w", l.ID, err)
		}
	}
	return nil
}
