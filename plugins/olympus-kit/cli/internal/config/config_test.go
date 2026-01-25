package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Output != "table" {
		t.Errorf("Default Output = %q, want %q", cfg.Output, "table")
	}
	if cfg.BaseDir != ".agents/olympus" {
		t.Errorf("Default BaseDir = %q, want %q", cfg.BaseDir, ".agents/olympus")
	}
	if cfg.Verbose {
		t.Error("Default Verbose = true, want false")
	}
	if cfg.Search.DefaultLimit != 10 {
		t.Errorf("Default Search.DefaultLimit = %d, want %d", cfg.Search.DefaultLimit, 10)
	}
	if !cfg.Search.UseSmartConnections {
		t.Error("Default Search.UseSmartConnections = false, want true")
	}
}

func TestMerge(t *testing.T) {
	dst := Default()
	src := &Config{
		Output:  "json",
		BaseDir: "/custom/path",
	}

	result := merge(dst, src)

	if result.Output != "json" {
		t.Errorf("merge Output = %q, want %q", result.Output, "json")
	}
	if result.BaseDir != "/custom/path" {
		t.Errorf("merge BaseDir = %q, want %q", result.BaseDir, "/custom/path")
	}
	// Defaults should be preserved when not overridden
	if result.Search.DefaultLimit != 10 {
		t.Errorf("merge preserved DefaultLimit = %d, want %d", result.Search.DefaultLimit, 10)
	}
}

func TestMerge_BooleanOverride(t *testing.T) {
	dst := Default()
	if !dst.Search.UseSmartConnections {
		t.Fatal("Precondition: default UseSmartConnections should be true")
	}

	// Test explicit false override
	src := &Config{
		Search: SearchConfig{
			UseSmartConnections:    false,
			UseSmartConnectionsSet: true,
		},
	}

	result := merge(dst, src)

	if result.Search.UseSmartConnections {
		t.Error("merge should override UseSmartConnections to false")
	}
	if !result.Search.UseSmartConnectionsSet {
		t.Error("merge should set UseSmartConnectionsSet")
	}
}

func TestMerge_BooleanNotSet(t *testing.T) {
	dst := Default()
	src := &Config{
		Output: "json",
		// UseSmartConnectionsSet is false (default)
	}

	result := merge(dst, src)

	// Should keep default (true) since not explicitly set
	if !result.Search.UseSmartConnections {
		t.Error("merge should preserve default UseSmartConnections when not set")
	}
}

func TestApplyEnv(t *testing.T) {
	// Save and restore env
	origOutput := os.Getenv("OLYMPUS_OUTPUT")
	origVerbose := os.Getenv("OLYMPUS_VERBOSE")
	origNoSC := os.Getenv("OLYMPUS_NO_SC")
	defer func() {
		os.Setenv("OLYMPUS_OUTPUT", origOutput)
		os.Setenv("OLYMPUS_VERBOSE", origVerbose)
		os.Setenv("OLYMPUS_NO_SC", origNoSC)
	}()

	os.Setenv("OLYMPUS_OUTPUT", "yaml")
	os.Setenv("OLYMPUS_VERBOSE", "true")
	os.Setenv("OLYMPUS_NO_SC", "1")

	cfg := Default()
	cfg = applyEnv(cfg)

	if cfg.Output != "yaml" {
		t.Errorf("applyEnv Output = %q, want %q", cfg.Output, "yaml")
	}
	if !cfg.Verbose {
		t.Error("applyEnv Verbose = false, want true")
	}
	if cfg.Search.UseSmartConnections {
		t.Error("applyEnv UseSmartConnections = true, want false")
	}
	if !cfg.Search.UseSmartConnectionsSet {
		t.Error("applyEnv should set UseSmartConnectionsSet when OLYMPUS_NO_SC is set")
	}
}

func TestLoadFromPath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write test config
	content := `
output: json
base_dir: /custom/olympus
verbose: true
search:
  default_limit: 20
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadFromPath(configPath)
	if err != nil {
		t.Fatalf("loadFromPath() error = %v", err)
	}

	if cfg.Output != "json" {
		t.Errorf("loadFromPath Output = %q, want %q", cfg.Output, "json")
	}
	if cfg.BaseDir != "/custom/olympus" {
		t.Errorf("loadFromPath BaseDir = %q, want %q", cfg.BaseDir, "/custom/olympus")
	}
	if !cfg.Verbose {
		t.Error("loadFromPath Verbose = false, want true")
	}
	if cfg.Search.DefaultLimit != 20 {
		t.Errorf("loadFromPath DefaultLimit = %d, want %d", cfg.Search.DefaultLimit, 20)
	}
}

func TestLoadFromPath_NotExists(t *testing.T) {
	cfg, err := loadFromPath("/nonexistent/config.yaml")
	// Should return nil config and error, but not panic
	if cfg != nil {
		t.Errorf("loadFromPath for nonexistent file should return nil config")
	}
	if err == nil {
		t.Errorf("loadFromPath for nonexistent file should return error")
	}
}

func TestLoadFromPath_Empty(t *testing.T) {
	cfg, err := loadFromPath("")
	if cfg != nil || err != nil {
		t.Errorf("loadFromPath(\"\") = %v, %v; want nil, nil", cfg, err)
	}
}

func TestResolve(t *testing.T) {
	// Test that flag overrides take precedence
	rc := Resolve("json", "/flag/path", true)

	if rc.Output.Value != "json" {
		t.Errorf("Resolve Output.Value = %v, want %q", rc.Output.Value, "json")
	}
	if rc.Output.Source != SourceFlag {
		t.Errorf("Resolve Output.Source = %v, want %v", rc.Output.Source, SourceFlag)
	}
	if rc.BaseDir.Value != "/flag/path" {
		t.Errorf("Resolve BaseDir.Value = %v, want %q", rc.BaseDir.Value, "/flag/path")
	}
	if rc.Verbose.Value != true {
		t.Errorf("Resolve Verbose.Value = %v, want true", rc.Verbose.Value)
	}
}
