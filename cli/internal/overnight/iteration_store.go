package overnight

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// iterFilenameRe matches persisted iteration filenames of the form
// "iter-<N>.json" where N is a 1-based positive decimal index.
var iterFilenameRe = regexp.MustCompile(`^iter-(\d+)\.json$`)

// writeIterationAtomic persists a single IterationSummary to
// dir/iter-<Index>.json using a temp-file + rename pattern so a crash in
// the middle of a write never leaves the on-disk history half-formed.
//
// Steps (in order; each step is load-bearing):
//  1. os.MkdirAll on the parent dir.
//  2. json.MarshalIndent first so an encoding error never leaves a partial file.
//  3. Create a sibling temp file in the same dir so os.Rename is atomic
//     within the same filesystem.
//  4. fsync the file so bytes hit disk before the rename.
//  5. os.Rename into the final path.
//  6. fsync the parent directory so the rename itself is durable.
func writeIterationAtomic(dir string, iter IterationSummary) error {
	// 1. Ensure parent exists.
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("overnight: mkdir iteration dir: %w", err)
	}

	// 2. Marshal first so an encoding error never leaves a partial file.
	data, err := json.MarshalIndent(iter, "", "  ")
	if err != nil {
		return fmt.Errorf("overnight: marshal iter-%d: %w", iter.Index, err)
	}

	// 3. Write to a sibling temp file in the same dir so os.Rename is
	//    atomic within the same filesystem.
	f, err := os.CreateTemp(dir, fmt.Sprintf(".iter-%d.*.json.tmp", iter.Index))
	if err != nil {
		return fmt.Errorf("overnight: create temp for iter-%d: %w", iter.Index, err)
	}
	tmpPath := f.Name()
	cleanupTmp := true
	defer func() {
		if cleanupTmp {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return fmt.Errorf("overnight: write iter-%d: %w", iter.Index, err)
	}
	// 4. fsync the file so bytes hit disk before the rename.
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return fmt.Errorf("overnight: fsync iter-%d: %w", iter.Index, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("overnight: close iter-%d: %w", iter.Index, err)
	}

	// 5. Atomic rename into final path.
	finalPath := filepath.Join(dir, fmt.Sprintf("iter-%d.json", iter.Index))
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return fmt.Errorf("overnight: rename iter-%d: %w", iter.Index, err)
	}
	cleanupTmp = false

	// 6. fsync the parent directory so the rename itself is durable. On
	//    platforms where directory fsync is a no-op (Windows), this is
	//    best-effort and returns nil without raising.
	if err := fsyncDir(dir); err != nil {
		return fmt.Errorf("overnight: fsync iteration dir: %w", err)
	}

	return nil
}

// fsyncDir opens the directory and calls Sync on it. On platforms where
// directory fsync is not meaningful (Windows), this is a best-effort
// no-op and returns nil.
func fsyncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	if err := d.Sync(); err != nil {
		// Linux/Darwin: real durability signal.
		// Windows: may return EINVAL on directory handles; swallow it.
		if runtime.GOOS == "windows" {
			return nil
		}
		return err
	}
	return nil
}

// writeCommittedButFlaggedMarker writes a sentinel file next to
// iter-<N>.json announcing that the iteration committed successfully
// but was flagged (post-commit regression halt) and the loop stopped.
// Operators can find flagged iterations by directory listing without
// parsing every iter-<N>.json file.
//
// The marker is empty (zero bytes); its presence is the signal.
// Filename: committed-but-flagged.iter-<N>.marker
//
// Uses the same atomic write pattern as writeIterationAtomic:
// CreateTemp → Sync → Rename → fsyncDir. Failure to write the marker
// is NOT a hard error — the caller surfaces the strip via Degraded.
func writeCommittedButFlaggedMarker(dir string, iterIndex int) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("overnight: mkdir marker dir: %w", err)
	}
	f, err := os.CreateTemp(dir, fmt.Sprintf(".committed-but-flagged.iter-%d.*.marker.tmp", iterIndex))
	if err != nil {
		return fmt.Errorf("overnight: create temp marker: %w", err)
	}
	tmpPath := f.Name()
	cleanupTmp := true
	defer func() {
		if cleanupTmp {
			_ = os.Remove(tmpPath)
		}
	}()
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return fmt.Errorf("overnight: fsync marker: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("overnight: close marker: %w", err)
	}
	finalPath := filepath.Join(dir, fmt.Sprintf("committed-but-flagged.iter-%d.marker", iterIndex))
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return fmt.Errorf("overnight: rename marker: %w", err)
	}
	cleanupTmp = false
	return fsyncDir(dir)
}

// ListCommittedButFlaggedMarkers returns the iteration indices that
// have a committed-but-flagged marker file in dir. Used by operators
// and downstream tooling to quickly find flagged iterations without
// parsing JSON. Returns an empty slice (not nil error) when the dir
// does not exist.
func ListCommittedButFlaggedMarkers(dir string) ([]int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("overnight: readdir marker dir: %w", err)
	}
	var out []int
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		// Pattern: committed-but-flagged.iter-<N>.marker
		const prefix = "committed-but-flagged.iter-"
		const suffix = ".marker"
		if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, suffix) {
			continue
		}
		numStr := name[len(prefix) : len(name)-len(suffix)]
		n, err := strconv.Atoi(numStr)
		if err != nil || n <= 0 {
			continue
		}
		out = append(out, n)
	}
	sort.Ints(out)
	return out, nil
}

// LoadIterations reads every valid iter-<N>.json from dir in ascending
// index order and returns the slice. A "valid" file is one where:
//   - the filename matches iter-<N>.json with N a positive decimal
//   - the JSON parses into an IterationSummary
//   - the embedded Index matches the filename's N
//   - the embedded ID's "<run-id>-iter-<N>" prefix matches the expected runID
//
// Any file that fails those checks is returned in the "rejected" slice
// along with a human-readable reason, so the caller can surface the gap
// via the Degraded list without halting the resume.
//
// Boundary behavior:
//   - dir does not exist           -> ([]IterationSummary{}, nil, nil)
//   - dir exists but is empty       -> ([]IterationSummary{}, nil, nil)
//   - dir contains gaps (1,2,4)    -> returns 1..2, rejects 4 as "gap in indices"
//   - corrupt JSON                  -> the file is rejected, load continues
//   - wrong runID in embedded ID    -> the file is rejected, load continues
func LoadIterations(dir string, expectedRunID string) ([]IterationSummary, []string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []IterationSummary{}, nil, nil
		}
		return nil, nil, fmt.Errorf("overnight: read iteration dir: %w", err)
	}

	type parsed struct {
		index int
		iter  IterationSummary
	}
	var good []parsed
	var rejected []string

	// Collect filenames to iterate deterministically for reject messages too.
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		m := iterFilenameRe.FindStringSubmatch(name)
		if m == nil {
			// Not an iter-<N>.json file; skip silently (temp files, etc.
			// starting with '.' are already skipped by the regex).
			continue
		}
		idx, convErr := strconv.Atoi(m[1])
		if convErr != nil || idx <= 0 {
			rejected = append(rejected, fmt.Sprintf("%s: invalid index", name))
			continue
		}
		path := filepath.Join(dir, name)
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			rejected = append(rejected, fmt.Sprintf("%s: read: %v", name, readErr))
			continue
		}
		var iter IterationSummary
		if parseErr := json.Unmarshal(raw, &iter); parseErr != nil {
			rejected = append(rejected, fmt.Sprintf("%s: parse: %v", name, parseErr))
			continue
		}
		if iter.Index != idx {
			rejected = append(rejected, fmt.Sprintf("%s: index mismatch (filename=%d, embedded=%d)", name, idx, iter.Index))
			continue
		}
		wantPrefix := fmt.Sprintf("%s-iter-%d", expectedRunID, idx)
		if string(iter.ID) != wantPrefix && !strings.HasPrefix(string(iter.ID), expectedRunID+"-iter-") {
			rejected = append(rejected, fmt.Sprintf("%s: runID mismatch (id=%q, want prefix %q-iter-)", name, iter.ID, expectedRunID))
			continue
		}
		// Stricter: require exact match of "<runID>-iter-<N>".
		if string(iter.ID) != wantPrefix {
			rejected = append(rejected, fmt.Sprintf("%s: id mismatch (id=%q, want %q)", name, iter.ID, wantPrefix))
			continue
		}
		good = append(good, parsed{index: idx, iter: iter})
	}

	sort.Slice(good, func(i, j int) bool { return good[i].index < good[j].index })

	// Enforce contiguous indices starting at 1. The first gap terminates
	// the "accepted" prefix; everything after is rejected as "gap in indices".
	out := make([]IterationSummary, 0, len(good))
	expected := 1
	gapHit := false
	for _, p := range good {
		if gapHit {
			rejected = append(rejected, fmt.Sprintf("iter-%d.json: gap in indices (expected %d)", p.index, expected))
			continue
		}
		if p.index != expected {
			// Record the gap for the current element and every subsequent
			// element in the good list.
			rejected = append(rejected, fmt.Sprintf("iter-%d.json: gap in indices (expected %d)", p.index, expected))
			gapHit = true
			continue
		}
		out = append(out, p.iter)
		expected++
	}

	return out, rejected, nil
}
