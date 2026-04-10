package corpus

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFile is a small helper that writes a file under dir, creating the
// parent directory tree as needed.
func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	full := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(full), err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", full, err)
	}
}

func TestCompute_MissingAgentsDir(t *testing.T) {
	tmp := t.TempDir()
	vec, _, err := Compute(tmp)
	if err == nil {
		t.Fatalf("expected error for missing .agents, got nil")
	}
	if vec == nil {
		t.Fatalf("expected non-nil vector even on error")
	}
	if vec.MaturityProvisional != 0 || vec.CitationCoverage != 0 {
		t.Fatalf("expected zero-value vector on missing .agents, got %+v", vec)
	}
}

func TestCompute_EmptyLearningsAndFindings(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	vec, degraded, err := Compute(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vec.MaturityProvisional != 0 || vec.UnresolvedFindings != 0 {
		t.Fatalf("expected zeros, got %+v", vec)
	}
	if len(degraded) == 0 {
		t.Fatalf("expected degraded notes on empty corpus")
	}
}

func TestCompute_MaturityFraction(t *testing.T) {
	tmp := t.TempDir()
	// 2 provisional, 1 accepted, 1 missing maturity => 3/4 = 0.75
	writeFile(t, tmp, ".agents/learnings/a.md", "---\nmaturity: provisional\n---\nbody\n")
	writeFile(t, tmp, ".agents/learnings/b.md", "---\nmaturity: Provisional\n---\nbody\n")
	writeFile(t, tmp, ".agents/learnings/c.md", "---\nmaturity: accepted\n---\nbody\n")
	writeFile(t, tmp, ".agents/learnings/d.md", "---\ntitle: no maturity here\n---\nbody\n")

	vec, _, err := Compute(tmp)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	want := 0.75
	if vec.MaturityProvisional != want {
		t.Fatalf("MaturityProvisional = %v, want %v", vec.MaturityProvisional, want)
	}
}

func TestCompute_CitationCoverage(t *testing.T) {
	tmp := t.TempDir()
	// 2 with source_bead, 1 without => 2/3 ≈ 0.667
	writeFile(t, tmp, ".agents/learnings/a.md", "---\nsource_bead: bd-123\n---\nbody\n")
	writeFile(t, tmp, ".agents/learnings/b.md", "---\nsource_bead: bd-456\n---\nbody\n")
	writeFile(t, tmp, ".agents/learnings/c.md", "---\ntitle: no citation\n---\nbody\n")

	vec, _, err := Compute(tmp)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	got := vec.CitationCoverage
	if got < 0.666 || got > 0.667 {
		t.Fatalf("CitationCoverage = %v, want approx 0.667", got)
	}
}

func TestCompute_UnresolvedFindingsCount(t *testing.T) {
	tmp := t.TempDir()
	// 5 findings: 2 resolved, 3 unresolved
	writeFile(t, tmp, ".agents/findings/f-001.md", "# Finding\n\n**Resolved:** yes\n")
	writeFile(t, tmp, ".agents/findings/f-002.md", "---\nresolved: true\n---\n# Finding\n")
	writeFile(t, tmp, ".agents/findings/f-003.md", "# Finding\n\nOpen question.\n")
	writeFile(t, tmp, ".agents/findings/f-004.md", "# Finding\n\nAnother open item.\n")
	writeFile(t, tmp, ".agents/findings/f-005.md", "---\ntitle: five\n---\n# Finding\n")

	vec, _, err := Compute(tmp)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	if vec.UnresolvedFindings != 3 {
		t.Fatalf("UnresolvedFindings = %d, want 3", vec.UnresolvedFindings)
	}
}

func TestCompute_InjectVisibility_ExcludesSuperseded(t *testing.T) {
	tmp := t.TempDir()
	// 3 learnings: 2 normal, 1 superseded => visible 2/3 ≈ 0.667
	writeFile(t, tmp, ".agents/learnings/a.md", "---\ntitle: one\n---\nbody\n")
	writeFile(t, tmp, ".agents/learnings/b.md", "---\ntitle: two\n---\nbody\n")
	writeFile(t, tmp, ".agents/learnings/c.md", "---\ntitle: three\nsuperseded: true\n---\nbody\n")

	vec, _, err := Compute(tmp)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	got := vec.InjectVisibility
	if got < 0.666 || got > 0.667 {
		t.Fatalf("InjectVisibility = %v, want approx 0.667", got)
	}
}

func TestCompute_Deterministic(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, tmp, ".agents/learnings/a.md", "---\nmaturity: provisional\nsource_bead: bd-1\n---\nbody\n")
	writeFile(t, tmp, ".agents/learnings/b.md", "---\nmaturity: stable\nsource_bead: bd-2\n---\nbody\n")
	writeFile(t, tmp, ".agents/findings/f-001.md", "# Finding\n\n**Resolved:** yes\n")
	writeFile(t, tmp, ".agents/findings/f-002.md", "# Finding\n\nOpen.\n")

	v1, d1, err := Compute(tmp)
	if err != nil {
		t.Fatalf("compute1: %v", err)
	}
	v2, d2, err := Compute(tmp)
	if err != nil {
		t.Fatalf("compute2: %v", err)
	}
	// Ignore ComputedAt.
	v1.ComputedAt = v2.ComputedAt
	if *v1 != *v2 {
		t.Fatalf("vectors differ across runs:\n  v1=%+v\n  v2=%+v", v1, v2)
	}
	if len(d1) != len(d2) {
		t.Fatalf("degraded counts differ: %d vs %d", len(d1), len(d2))
	}
	for i := range d1 {
		if d1[i] != d2[i] {
			t.Fatalf("degraded[%d] differs: %q vs %q", i, d1[i], d2[i])
		}
	}
}
