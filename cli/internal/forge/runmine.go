// Package forge — runmine.go provides the in-process RunMinePass entry point
// used by Dream's INGEST stage to mine already-forged session files under
// .agents/sessions/ for decision/knowledge records.
//
// This file deliberately avoids any cobra, flag, stdout, or stderr coupling.
// All rendering stays in the caller (cli/cmd/ao/forge.go for the cobra flow,
// cli/internal/overnight/ingest.go for the Dream flow). Soft-fail notes are
// accumulated on MineReport.Degraded — hard failures surface as errors.
//
// Mirrors lifecycle.ExecuteCloseLoop from the parent Dream plan:
// the caller constructs MineOpts and calls RunMinePass; this file holds
// pure logic only.
package forge

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// MineOpts configures a single RunMinePass. Zero-valued knobs fall back to
// documented defaults in normalize() where a reasonable default exists.
type MineOpts struct {
	// SessionsDir is the absolute path to the directory that holds forged
	// session JSONL files (typically filepath.Join(cwd, ".agents",
	// "sessions")). Required — RunMinePass returns an error if unset or if
	// the directory does not exist.
	SessionsDir string

	// SinceTime filters sessions to those modified at or after this time.
	// Zero value means "scan everything".
	SinceTime time.Time

	// Quiet is a hook-friendly mode flag. RunMinePass itself prints nothing;
	// the field is carried on the report so callers can propagate it to
	// their own rendering layer.
	Quiet bool

	// MaxSessions caps the number of session files processed in one pass.
	// Zero means unlimited.
	MaxSessions int
}

// MineReport is the aggregate return value of RunMinePass. It is safe for
// callers to append to Degraded after the fact (e.g., to layer their own
// soft-fail notes on top of the lifecycle-internal ones).
type MineReport struct {
	// Learnings is the flat list of mined learnings across all session
	// files that were successfully parsed.
	Learnings []MinedLearning

	// SessionsRead is the number of session files that were opened and
	// parsed successfully (excluding malformed files counted in Degraded).
	SessionsRead int

	// Duration is wall-clock time spent inside RunMinePass. Populated on
	// success and on error paths that return a partial report.
	Duration time.Duration

	// Degraded collects soft-fail warnings produced during the run. Hard
	// failures are returned as errors; items here are non-fatal.
	Degraded []string
}

// MinedLearning is one learning record extracted from a forged session file.
// It is intentionally thin — the fields are what Dream's INGEST stage needs
// to report into its substage counts. Downstream consumers that want the
// full storage.Session shape should re-load the source file themselves.
type MinedLearning struct {
	// Title is a short human-readable label for the learning. Populated
	// from the session summary when available, falling back to the source
	// filename stem.
	Title string

	// Body is the learning content. For decision records this is the
	// decision text; for knowledge records this is the knowledge snippet.
	Body string

	// Kind describes the record type — currently either "decision" or
	// "knowledge".
	Kind string

	// Source is the basename of the session file the learning was mined
	// from (not a full path — callers can join against opts.SessionsDir if
	// they need to open the source).
	Source string

	// Extracted is the wall-clock time at which RunMinePass processed the
	// source file.
	Extracted time.Time
}

// minedSessionFile is the on-disk shape this file cares about. It is a
// narrow subset of storage.Session — only the fields RunMinePass needs —
// declared locally to avoid pulling the storage package into internal/forge
// (which would create an import cycle with cmd/ao).
type minedSessionFile struct {
	ID        string    `json:"id"`
	Date      time.Time `json:"date"`
	Summary   string    `json:"summary"`
	Decisions []string  `json:"decisions,omitempty"`
	Knowledge []string  `json:"knowledge,omitempty"`
}

// RunMinePass scans opts.SessionsDir for forged session JSONL files and
// returns every decision/knowledge record as a flat MinedLearning slice.
// The function never writes to stdout or stderr; soft-fail notes are
// accumulated on the returned MineReport.Degraded field, and hard failures
// (missing cwd, missing sessions dir, invalid opts) are surfaced as errors.
//
// Required opts: SessionsDir.
//
// Typical usage from Dream's INGEST stage:
//
//	report, err := forge.RunMinePass(cwd, forge.MineOpts{
//	    SessionsDir: filepath.Join(cwd, ".agents", "sessions"),
//	    SinceTime:   lastRunAt,
//	    Quiet:       true,
//	})
func RunMinePass(cwd string, opts MineOpts) (*MineReport, error) {
	start := time.Now()
	report := &MineReport{}

	if cwd == "" {
		return report, errors.New("forge: RunMinePass requires cwd")
	}
	if opts.SessionsDir == "" {
		return report, errors.New("forge: RunMinePass requires opts.SessionsDir")
	}

	info, err := os.Stat(opts.SessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			// No sessions directory yet — return an empty report with a
			// degraded note so the caller can surface it, but do not fail
			// the whole INGEST stage.
			report.Degraded = append(report.Degraded,
				fmt.Sprintf("forge: sessions dir %q does not exist", opts.SessionsDir))
			report.Duration = time.Since(start)
			return report, nil
		}
		return report, fmt.Errorf("forge: stat sessions dir: %w", err)
	}
	if !info.IsDir() {
		return report, fmt.Errorf("forge: sessions path %q is not a directory", opts.SessionsDir)
	}

	candidates, err := collectSessionCandidates(opts.SessionsDir, opts.SinceTime)
	if err != nil {
		return report, fmt.Errorf("forge: collect session candidates: %w", err)
	}

	if opts.MaxSessions > 0 && len(candidates) > opts.MaxSessions {
		candidates = candidates[:opts.MaxSessions]
	}

	now := time.Now()
	for _, cand := range candidates {
		learnings, readErr := mineSessionFile(cand, now)
		if readErr != nil {
			report.Degraded = append(report.Degraded,
				fmt.Sprintf("forge: %s: %v", filepath.Base(cand), readErr))
			continue
		}
		report.SessionsRead++
		report.Learnings = append(report.Learnings, learnings...)
	}

	report.Duration = time.Since(start)
	return report, nil
}

// collectSessionCandidates walks sessionsDir (non-recursive plus one level)
// for *.jsonl / *.json session files newer than sinceTime. Results are
// sorted by mod time (newest first) so MaxSessions trims the oldest.
func collectSessionCandidates(sessionsDir string, sinceTime time.Time) ([]string, error) {
	type candidate struct {
		path    string
		modTime time.Time
	}

	var cands []candidate
	walkErr := filepath.WalkDir(sessionsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Permission errors on a subdirectory should not abort the
			// whole walk — skip and continue.
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".jsonl" && ext != ".json" {
			return nil
		}
		info, infoErr := d.Info()
		if infoErr != nil {
			return nil
		}
		if !sinceTime.IsZero() && info.ModTime().Before(sinceTime) {
			return nil
		}
		cands = append(cands, candidate{path: path, modTime: info.ModTime()})
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	sort.Slice(cands, func(i, j int) bool {
		return cands[i].modTime.After(cands[j].modTime)
	})

	paths := make([]string, len(cands))
	for i, c := range cands {
		paths[i] = c.path
	}
	return paths, nil
}

// mineSessionFile opens a single forged session file and returns every
// decision/knowledge record as a MinedLearning. Files are expected in the
// storage.Session JSON shape (single object) or JSONL (one object per
// line); both are probed.
func mineSessionFile(path string, extractedAt time.Time) ([]MinedLearning, error) {
	f, err := os.Open(path) // #nosec G304 -- path came from directory walk
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer func() { _ = f.Close() }()

	var learnings []MinedLearning
	source := filepath.Base(path)

	// First, try reading the whole file as a single JSON object. This is
	// the common case for storage.FileStorage output.
	stat, statErr := f.Stat()
	if statErr == nil && stat.Size() > 0 && stat.Size() < 16*1024*1024 {
		buf := make([]byte, stat.Size())
		if _, readErr := f.Read(buf); readErr == nil {
			var single minedSessionFile
			if jerr := json.Unmarshal(buf, &single); jerr == nil && (len(single.Decisions) > 0 || len(single.Knowledge) > 0 || single.ID != "") {
				learnings = appendSessionLearnings(learnings, single, source, extractedAt)
				return learnings, nil
			}
		}
		// Fall through to JSONL mode if single-object parse failed.
		if _, seekErr := f.Seek(0, 0); seekErr != nil {
			return nil, fmt.Errorf("seek: %w", seekErr)
		}
	}

	// JSONL fallback: one session object per line.
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var sess minedSessionFile
		if jerr := json.Unmarshal(line, &sess); jerr != nil {
			continue // skip malformed lines, matches forge parser tolerance
		}
		learnings = appendSessionLearnings(learnings, sess, source, extractedAt)
	}
	if serr := scanner.Err(); serr != nil {
		return learnings, fmt.Errorf("scan: %w", serr)
	}

	return learnings, nil
}

// appendSessionLearnings flattens one session's decisions + knowledge into
// MinedLearning records.
func appendSessionLearnings(dst []MinedLearning, sess minedSessionFile, source string, extractedAt time.Time) []MinedLearning {
	title := sess.Summary
	if title == "" {
		title = sess.ID
	}
	if title == "" {
		title = source
	}
	for _, dec := range sess.Decisions {
		if dec == "" {
			continue
		}
		dst = append(dst, MinedLearning{
			Title:     title,
			Body:      dec,
			Kind:      "decision",
			Source:    source,
			Extracted: extractedAt,
		})
	}
	for _, k := range sess.Knowledge {
		if k == "" {
			continue
		}
		dst = append(dst, MinedLearning{
			Title:     title,
			Body:      k,
			Kind:      "knowledge",
			Source:    source,
			Extracted: extractedAt,
		})
	}
	return dst
}
