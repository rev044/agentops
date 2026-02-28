package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// goals_init.go — buildDefaultGoalFile (0% → higher)
// ---------------------------------------------------------------------------

func TestCov15_buildDefaultGoalFile(t *testing.T) {
	gf := buildDefaultGoalFile()
	if gf == nil {
		t.Fatal("buildDefaultGoalFile returned nil")
	}
	if gf.Version != 4 {
		t.Errorf("version: got %d, want 4", gf.Version)
	}
	if len(gf.Directives) == 0 {
		t.Error("expected at least one directive")
	}
	if len(gf.NorthStars) == 0 {
		t.Error("expected default north stars")
	}
	if len(gf.AntiStars) == 0 {
		t.Error("expected default anti stars")
	}
}

// ---------------------------------------------------------------------------
// goals_init.go — buildInteractiveGoalFile (0% → higher)
// ---------------------------------------------------------------------------

func TestCov15_buildInteractiveGoalFile_withInput(t *testing.T) {
	// Provide all prompts: mission, north stars, anti stars, directive title, directive desc
	input := "My mission statement\nNorth star 1,North star 2\nAnti star 1\nEstablish baseline\nGet all gates passing\n"
	gf, err := buildInteractiveGoalFile(strings.NewReader(input))
	if err != nil {
		t.Fatalf("buildInteractiveGoalFile: %v", err)
	}
	if gf.Mission != "My mission statement" {
		t.Errorf("mission: got %q, want 'My mission statement'", gf.Mission)
	}
	if len(gf.NorthStars) != 2 {
		t.Errorf("northStars: got %d, want 2", len(gf.NorthStars))
	}
	if len(gf.AntiStars) != 1 {
		t.Errorf("antiStars: got %d, want 1", len(gf.AntiStars))
	}
}

func TestCov15_buildInteractiveGoalFile_emptyUsesDefaults(t *testing.T) {
	// All empty inputs → all defaults are applied
	input := "\n\n\n\n\n"
	gf, err := buildInteractiveGoalFile(strings.NewReader(input))
	if err != nil {
		t.Fatalf("buildInteractiveGoalFile empty input: %v", err)
	}
	// Should have default north stars since we provided empty
	if len(gf.NorthStars) == 0 {
		t.Error("expected default northStars for empty input")
	}
	if len(gf.AntiStars) == 0 {
		t.Error("expected default antiStars for empty input")
	}
	if len(gf.Directives) == 0 {
		t.Error("expected default directive for empty input")
	}
}

// ---------------------------------------------------------------------------
// goals_init.go — splitCommaSeparated (0% → higher)
// ---------------------------------------------------------------------------

func TestCov15_splitCommaSeparated_empty(t *testing.T) {
	result := splitCommaSeparated("")
	if len(result) != 0 {
		t.Errorf("splitCommaSeparated empty: got %v", result)
	}
}

func TestCov15_splitCommaSeparated_whitespaceOnly(t *testing.T) {
	result := splitCommaSeparated("   ")
	if len(result) != 0 {
		t.Errorf("splitCommaSeparated whitespace: got %v", result)
	}
}

func TestCov15_splitCommaSeparated_multi(t *testing.T) {
	result := splitCommaSeparated("a, b,  c  ")
	if len(result) != 3 {
		t.Errorf("splitCommaSeparated multi: got %d items, want 3", len(result))
	}
}

func TestCov15_splitCommaSeparated_single(t *testing.T) {
	result := splitCommaSeparated("only one")
	if len(result) != 1 {
		t.Errorf("splitCommaSeparated single: got %d items, want 1", len(result))
	}
}

// ---------------------------------------------------------------------------
// goals_init.go — detectGates (0% → higher)
// ---------------------------------------------------------------------------

func TestCov15_detectGates_emptyDir(t *testing.T) {
	tmp := t.TempDir()
	gates := detectGates(tmp)
	if len(gates) != 0 {
		t.Errorf("detectGates empty dir: got %d gates, want 0", len(gates))
	}
}

func TestCov15_detectGates_rootGoMod(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	gates := detectGates(tmp)
	if len(gates) < 2 {
		t.Errorf("detectGates go.mod: got %d gates, want >= 2", len(gates))
	}
}

func TestCov15_detectGates_nestedGoMod(t *testing.T) {
	tmp := t.TempDir()
	cliDir := filepath.Join(tmp, "cli")
	if err := os.MkdirAll(cliDir, 0755); err != nil {
		t.Fatalf("mkdir cli: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cliDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0644); err != nil {
		t.Fatalf("write cli/go.mod: %v", err)
	}
	gates := detectGates(tmp)
	if len(gates) < 2 {
		t.Errorf("detectGates nested go.mod: got %d gates, want >= 2", len(gates))
	}
}

func TestCov15_detectGates_packageJson(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "package.json"), []byte(`{"name":"test"}`), 0644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}
	gates := detectGates(tmp)
	// Should have npm-test
	found := false
	for _, g := range gates {
		if g.ID == "npm-test" {
			found = true
		}
	}
	if !found {
		t.Errorf("detectGates package.json: expected npm-test gate, got %v", gates)
	}
}

func TestCov15_detectGates_cargoToml(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "Cargo.toml"), []byte("[package]\nname = \"test\"\n"), 0644); err != nil {
		t.Fatalf("write Cargo.toml: %v", err)
	}
	gates := detectGates(tmp)
	found := false
	for _, g := range gates {
		if g.ID == "cargo-test" {
			found = true
		}
	}
	if !found {
		t.Errorf("detectGates Cargo.toml: expected cargo-test gate, got %v", gates)
	}
}

func TestCov15_detectGates_pyprojectToml(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "pyproject.toml"), []byte("[tool.pytest]\n"), 0644); err != nil {
		t.Fatalf("write pyproject.toml: %v", err)
	}
	gates := detectGates(tmp)
	found := false
	for _, g := range gates {
		if g.ID == "python-test" {
			found = true
		}
	}
	if !found {
		t.Errorf("detectGates pyproject.toml: expected python-test gate, got %v", gates)
	}
}

func TestCov15_detectGates_makefile(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "Makefile"), []byte("build:\n\techo ok\n"), 0644); err != nil {
		t.Fatalf("write Makefile: %v", err)
	}
	gates := detectGates(tmp)
	found := false
	for _, g := range gates {
		if g.ID == "make-build" {
			found = true
		}
	}
	if !found {
		t.Errorf("detectGates Makefile: expected make-build gate, got %v", gates)
	}
}

// ---------------------------------------------------------------------------
// goals_init.go — loadTemplate (0% → higher)
// ---------------------------------------------------------------------------

func TestCov15_loadTemplate_goCLI(t *testing.T) {
	tmpl, err := loadTemplate("go-cli")
	if err != nil {
		t.Fatalf("loadTemplate go-cli: %v", err)
	}
	if tmpl.Name == "" {
		t.Error("expected non-empty template name for go-cli")
	}
	if len(tmpl.Gates) == 0 {
		t.Error("expected at least one gate in go-cli template")
	}
}

func TestCov15_loadTemplate_generic(t *testing.T) {
	tmpl, err := loadTemplate("generic")
	if err != nil {
		t.Fatalf("loadTemplate generic: %v", err)
	}
	if tmpl == nil {
		t.Fatal("expected non-nil template for generic")
	}
}

func TestCov15_loadTemplate_notFound(t *testing.T) {
	_, err := loadTemplate("nonexistent-template-xyz-abc")
	if err == nil {
		t.Fatal("expected error for nonexistent template, got nil")
	}
}

// ---------------------------------------------------------------------------
// goals_init.go — templateGatesToGoals (0% → higher)
// ---------------------------------------------------------------------------

func TestCov15_templateGatesToGoals_withGates(t *testing.T) {
	tmpl := &goalTemplate{
		Name: "test-template",
		Gates: []goalTemplateGate{
			{ID: "gate-a", Description: "Gate A", Check: "echo a", Weight: 3, Type: "health"},
			{ID: "gate-b", Description: "Gate B", Check: "echo b", Weight: 5, Type: "quality"},
		},
	}
	result := templateGatesToGoals(tmpl)
	if len(result) != 2 {
		t.Fatalf("templateGatesToGoals: got %d, want 2", len(result))
	}
	if result[0].ID != "gate-a" {
		t.Errorf("first gate ID: got %q, want 'gate-a'", result[0].ID)
	}
}

func TestCov15_templateGatesToGoals_empty(t *testing.T) {
	tmpl := &goalTemplate{Name: "empty", Gates: nil}
	result := templateGatesToGoals(tmpl)
	if len(result) != 0 {
		t.Errorf("templateGatesToGoals empty: got %d, want 0", len(result))
	}
}

// ---------------------------------------------------------------------------
// goals_init.go — goalsInitCmd.RunE paths (0% → higher)
// ---------------------------------------------------------------------------

func TestCov15_goalsInitCmd_nonInteractiveDryRun(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	origNonInteractive := goalsInitNonInteractive
	origTemplate := goalsInitTemplate
	origDryRun := dryRun
	defer func() {
		goalsInitNonInteractive = origNonInteractive
		goalsInitTemplate = origTemplate
		dryRun = origDryRun
	}()
	goalsInitNonInteractive = true
	goalsInitTemplate = ""
	dryRun = true

	cmd := goalsInitCmd
	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("goalsInitCmd non-interactive dry-run: %v", err)
	}
}

func TestCov15_goalsInitCmd_nonInteractiveWritesFile(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	origNonInteractive := goalsInitNonInteractive
	origTemplate := goalsInitTemplate
	origDryRun := dryRun
	defer func() {
		goalsInitNonInteractive = origNonInteractive
		goalsInitTemplate = origTemplate
		dryRun = origDryRun
	}()
	goalsInitNonInteractive = true
	goalsInitTemplate = ""
	dryRun = false

	cmd := goalsInitCmd
	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("goalsInitCmd non-interactive write: %v", err)
	}
	// Verify file was created
	if _, statErr := os.Stat(filepath.Join(tmp, "GOALS.md")); os.IsNotExist(statErr) {
		t.Error("expected GOALS.md to be created")
	}
}

func TestCov15_goalsInitCmd_fileAlreadyExists(t *testing.T) {
	tmp := t.TempDir()
	// Create GOALS.md first
	if err := os.WriteFile(filepath.Join(tmp, "GOALS.md"), []byte("# Goals\n"), 0644); err != nil {
		t.Fatalf("write GOALS.md: %v", err)
	}

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	origNonInteractive := goalsInitNonInteractive
	defer func() { goalsInitNonInteractive = origNonInteractive }()
	goalsInitNonInteractive = true

	cmd := goalsInitCmd
	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error for existing GOALS.md, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' in error, got: %v", err)
	}
}

func TestCov15_goalsInitCmd_jsonOutput(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	origNonInteractive := goalsInitNonInteractive
	origTemplate := goalsInitTemplate
	origDryRun := dryRun
	origGoalsJSON := goalsJSON
	defer func() {
		goalsInitNonInteractive = origNonInteractive
		goalsInitTemplate = origTemplate
		dryRun = origDryRun
		goalsJSON = origGoalsJSON
	}()
	goalsInitNonInteractive = true
	goalsInitTemplate = ""
	dryRun = false
	goalsJSON = true

	cmd := goalsInitCmd
	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("goalsInitCmd json output: %v", err)
	}
}

func TestCov15_goalsInitCmd_withTemplate(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	origNonInteractive := goalsInitNonInteractive
	origTemplate := goalsInitTemplate
	origDryRun := dryRun
	defer func() {
		goalsInitNonInteractive = origNonInteractive
		goalsInitTemplate = origTemplate
		dryRun = origDryRun
	}()
	goalsInitNonInteractive = true
	goalsInitTemplate = "go-cli" // use a known embedded template
	dryRun = true

	cmd := goalsInitCmd
	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("goalsInitCmd with template: %v", err)
	}
}

// ---------------------------------------------------------------------------
// goals_init.go — autoDetectTemplate (0% → higher)
// ---------------------------------------------------------------------------

func TestCov15_autoDetectTemplate_emptyDir(t *testing.T) {
	tmp := t.TempDir()
	result := autoDetectTemplate(tmp)
	// Empty dir → unknown type → returns "" (generic maps to "")
	_ = result // just verify no panic
}
