package main

import (
	"os"
	"strings"
	"testing"
)

// TestMain clears AGENTOPS_RPI_RUNTIME* env vars before any test runs.
// This prevents host environment from leaking into test assertions,
// eliminating the need for per-test t.Setenv calls.
func TestMain(m *testing.M) {
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "AGENTOPS_RPI_RUNTIME") {
			key, _, _ := strings.Cut(env, "=")
			os.Unsetenv(key)
		}
	}
	os.Exit(m.Run())
}
