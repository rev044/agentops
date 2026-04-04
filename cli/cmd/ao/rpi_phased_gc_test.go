package main

import (
	"os"
	"testing"
	"time"
)

func TestGCExecutor_Name(t *testing.T) {
	e := &gcExecutor{}
	if e.Name() != "gc" {
		t.Errorf("gcExecutor.Name() = %q, want %q", e.Name(), "gc")
	}
}

func TestSelectExecutorFromCaps_GCBackend(t *testing.T) {
	caps := backendCapabilities{RuntimeMode: "gc"}
	opts := defaultPhasedEngineOptions()
	opts.WorkingDir = t.TempDir()

	executor, reason := selectExecutorFromCaps(caps, "", nil, opts)
	if executor.Name() != "gc" {
		t.Errorf("executor.Name() = %q, want %q", executor.Name(), "gc")
	}
	if reason != "runtime=gc" {
		t.Errorf("reason = %q, want %q", reason, "runtime=gc")
	}
}

func TestSelectExecutorFromCaps_GCFallbackToAuto(t *testing.T) {
	// When runtime is "auto", gc is NOT selected — stream is the default
	caps := backendCapabilities{RuntimeMode: "auto"}
	opts := defaultPhasedEngineOptions()

	executor, _ := selectExecutorFromCaps(caps, "", nil, opts)
	if executor.Name() == "gc" {
		t.Error("auto mode should not select gc executor")
	}
}

func TestGCExecutorAvailable_NoBinary(t *testing.T) {
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", t.TempDir())
	defer os.Setenv("PATH", origPath)

	if gcExecutorAvailable("/tmp") {
		t.Error("gcExecutorAvailable should return false when gc not on PATH")
	}
}

func TestGCExecutor_PhaseTimeout(t *testing.T) {
	e := &gcExecutor{
		phaseTimeout: 5 * time.Minute,
		pollInterval: 1 * time.Second,
	}
	if e.phaseTimeout != 5*time.Minute {
		t.Errorf("phaseTimeout = %v, want 5m", e.phaseTimeout)
	}
}

func TestValidateRuntimeMode_GC(t *testing.T) {
	if err := validateRuntimeMode("gc"); err != nil {
		t.Errorf("validateRuntimeMode(\"gc\") should succeed, got: %v", err)
	}
}

func TestGCCityPathFromOpts(t *testing.T) {
	opts := defaultPhasedEngineOptions()
	opts.GCCityPath = "/explicit/path"
	if got := gcCityPathFromOpts(opts); got != "/explicit/path" {
		t.Errorf("gcCityPathFromOpts with explicit = %q, want /explicit/path", got)
	}

	opts.GCCityPath = ""
	opts.WorkingDir = t.TempDir()
	// No city.toml exists, so should return empty
	if got := gcCityPathFromOpts(opts); got != "" {
		t.Errorf("gcCityPathFromOpts without city.toml = %q, want empty", got)
	}
}
