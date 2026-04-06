package rpi

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// PhaseArtifactNumberPattern matches phase-N in artifact filenames.
var PhaseArtifactNumberPattern = regexp.MustCompile(`phase-(\d+)`)

// ArtifactRef is a reference to an RPI artifact on disk.
type ArtifactRef struct {
	Path      string `json:"path"`
	Label     string `json:"label"`
	Kind      string `json:"kind"`
	Phase     int    `json:"phase,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
	SizeBytes int64  `json:"size_bytes,omitempty"`
}

// ArtifactContent holds the body of a read artifact.
type ArtifactContent struct {
	Path        string `json:"path"`
	Label       string `json:"label,omitempty"`
	Kind        string `json:"kind,omitempty"`
	ContentType string `json:"content_type"`
	UpdatedAt   string `json:"updated_at,omitempty"`
	SizeBytes   int64  `json:"size_bytes,omitempty"`
	Body        string `json:"body"`
	Truncated   bool   `json:"truncated,omitempty"`
}

// PathClean normalises a relative path to forward-slash form.
func PathClean(rel string) string {
	return filepath.ToSlash(filepath.Clean(filepath.FromSlash(strings.TrimSpace(rel))))
}

// IsSafeArtifactRelPath returns true when rel is a safe relative artifact path.
func IsSafeArtifactRelPath(rel string) bool {
	rel = PathClean(rel)
	if rel == "." || rel == "" {
		return false
	}
	if strings.HasPrefix(rel, "../") || rel == ".." || strings.HasPrefix(rel, "/") {
		return false
	}
	return true
}

// ClassifyRPIArtifact returns (kind, label, phase) for a relative artifact path.
// phasedStateFile and c2EventsFileName are passed in to avoid coupling to cmd/ao constants.
func ClassifyRPIArtifact(rel, phasedStateFile, c2EventsFileName string) (kind, label string, phase int) {
	base := filepath.Base(rel)
	phase = ArtifactPhaseNumber(base)

	switch {
	case strings.HasSuffix(rel, "execution-packet.json"):
		return "execution_packet", "Execution packet", 0
	case strings.HasSuffix(rel, filepath.ToSlash(filepath.Join(".agents", "rpi", phasedStateFile))):
		return "phased_state", "Phased state", 0
	case strings.HasSuffix(rel, c2EventsFileName):
		return "run_events", "Run events", 0
	case strings.HasSuffix(rel, "heartbeat.txt"):
		return "run_heartbeat", "Heartbeat", 0
	case strings.Contains(base, "-result.json"):
		return "phase_result", fmt.Sprintf("Phase %d result", phase), phase
	case strings.Contains(base, "-handoff.json"):
		return "phase_handoff", fmt.Sprintf("Phase %d handoff", phase), phase
	case strings.Contains(base, "-summary") && strings.HasSuffix(base, ".md"):
		return "phase_summary", fmt.Sprintf("Phase %d summary", phase), phase
	case strings.Contains(base, "-evaluator.json"):
		return "phase_evaluator", fmt.Sprintf("Phase %d evaluator", phase), phase
	case strings.Contains(rel, "/plans/"):
		return "plan", "Plan", 0
	case strings.Contains(rel, "/research/"):
		return "research", "Research", 0
	case strings.Contains(rel, "/council/") && strings.Contains(strings.ToLower(base), "pre-mortem"):
		return "council_pre_mortem", "Pre-mortem report", 0
	case strings.Contains(rel, "/council/") && strings.Contains(strings.ToLower(base), "post-mortem"):
		return "council_post_mortem", "Post-mortem report", 0
	case strings.Contains(rel, "/council/") && strings.Contains(strings.ToLower(base), "vibe"):
		return "council_vibe", "Vibe report", 0
	default:
		return "artifact", base, phase
	}
}

// ArtifactPhaseNumber extracts the phase number from a filename like "phase-2-result.json".
func ArtifactPhaseNumber(name string) int {
	matches := PhaseArtifactNumberPattern.FindStringSubmatch(name)
	if len(matches) != 2 {
		return 0
	}
	n, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0
	}
	return n
}

// ArtifactContentType returns the MIME type for an artifact path.
func ArtifactContentType(rel string) string {
	switch strings.ToLower(filepath.Ext(rel)) {
	case ".json", ".jsonl":
		return "application/json"
	case ".md", ".mdx":
		return "text/markdown"
	default:
		return "text/plain"
	}
}

// SortArtifactRefs sorts artifact refs by UpdatedAt (descending) then Path (ascending).
func SortArtifactRefs(refs []ArtifactRef) {
	for i := 0; i < len(refs); i++ {
		for j := i + 1; j < len(refs); j++ {
			if refs[j].UpdatedAt > refs[i].UpdatedAt ||
				(refs[j].UpdatedAt == refs[i].UpdatedAt && refs[j].Path < refs[i].Path) {
				refs[i], refs[j] = refs[j], refs[i]
			}
		}
	}
}
