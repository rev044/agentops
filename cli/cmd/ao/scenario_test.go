package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScenarioInit_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer os.Chdir(origDir)

	out, err := executeCommand("scenario", "init")
	if err != nil {
		t.Fatalf("scenario init failed: %v", err)
	}

	holdoutDir := filepath.Join(".agents", "holdout")
	if _, err := os.Stat(holdoutDir); os.IsNotExist(err) {
		t.Fatal("holdout directory not created")
	}
	if _, err := os.Stat(filepath.Join(holdoutDir, "README.md")); os.IsNotExist(err) {
		t.Fatal("README.md not created")
	}
	if !strings.Contains(out, "Initialized holdout directory") {
		t.Errorf("expected init confirmation, got: %s", out)
	}
}

func TestScenarioInit_Idempotent(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer os.Chdir(origDir)

	// Run twice
	executeCommand("scenario", "init")
	_, err := executeCommand("scenario", "init")
	if err != nil {
		t.Fatalf("second init should not error: %v", err)
	}
}

func TestScenarioList_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer os.Chdir(origDir)

	os.MkdirAll(filepath.Join(".agents", "holdout"), 0755)

	out, err := executeCommand("scenario", "list")
	if err != nil {
		t.Fatalf("list should not error on empty dir: %v", err)
	}
	if !strings.Contains(out, "No scenarios found") {
		t.Fatalf("expected 'No scenarios found', got: %s", out)
	}
}

func TestScenarioList_NoDirectory(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer os.Chdir(origDir)

	out, err := executeCommand("scenario", "list")
	if err != nil {
		t.Fatalf("list should not error when dir missing: %v", err)
	}
	if !strings.Contains(out, "No holdout directory found") {
		t.Fatalf("expected missing-dir message, got: %s", out)
	}
}

func TestScenarioList_WithScenarios(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer os.Chdir(origDir)

	holdoutDir := filepath.Join(".agents", "holdout")
	os.MkdirAll(holdoutDir, 0755)

	scenario := map[string]interface{}{
		"id": "s-2026-04-05-001", "version": 1, "date": "2026-04-05",
		"goal": "test goal", "narrative": "test narrative",
		"expected_outcome": "test outcome", "satisfaction_threshold": 0.8,
		"status": "active",
	}
	data, _ := json.Marshal(scenario)
	os.WriteFile(filepath.Join(holdoutDir, "s-2026-04-05-001.json"), data, 0644)

	out, err := executeCommand("scenario", "list")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if !strings.Contains(out, "test goal") {
		t.Fatalf("expected scenario in output, got: %s", out)
	}
}

func TestScenarioList_FilterByStatus(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer os.Chdir(origDir)

	holdoutDir := filepath.Join(".agents", "holdout")
	os.MkdirAll(holdoutDir, 0755)

	active := map[string]interface{}{
		"id": "s-2026-04-05-001", "version": 1, "date": "2026-04-05",
		"goal": "active goal", "narrative": "n", "expected_outcome": "o",
		"satisfaction_threshold": 0.8, "status": "active",
	}
	draft := map[string]interface{}{
		"id": "s-2026-04-05-002", "version": 1, "date": "2026-04-05",
		"goal": "draft goal", "narrative": "n", "expected_outcome": "o",
		"satisfaction_threshold": 0.5, "status": "draft",
	}
	d1, _ := json.Marshal(active)
	d2, _ := json.Marshal(draft)
	os.WriteFile(filepath.Join(holdoutDir, "s1.json"), d1, 0644)
	os.WriteFile(filepath.Join(holdoutDir, "s2.json"), d2, 0644)

	out, err := executeCommand("scenario", "list", "--status", "draft")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if !strings.Contains(out, "draft goal") {
		t.Fatalf("expected draft scenario, got: %s", out)
	}
	if strings.Contains(out, "active goal") {
		t.Fatalf("should not contain active scenario when filtering by draft")
	}
}

func TestScenarioValidate_ValidSchema(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer os.Chdir(origDir)

	holdoutDir := filepath.Join(".agents", "holdout")
	os.MkdirAll(holdoutDir, 0755)

	scenario := map[string]interface{}{
		"id": "s-2026-04-05-001", "version": 1, "date": "2026-04-05",
		"goal": "test", "narrative": "test", "expected_outcome": "test",
		"satisfaction_threshold": 0.7, "status": "active", "source": "human",
	}
	data, _ := json.Marshal(scenario)
	os.WriteFile(filepath.Join(holdoutDir, "test.json"), data, 0644)

	out, err := executeCommand("scenario", "validate")
	if err != nil {
		t.Fatalf("validate should pass: %v", err)
	}
	if !strings.Contains(out, "all pass") {
		t.Fatalf("expected 'all pass', got: %s", out)
	}
}

func TestScenarioValidate_InvalidSchema(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer os.Chdir(origDir)

	holdoutDir := filepath.Join(".agents", "holdout")
	os.MkdirAll(holdoutDir, 0755)

	// Missing required fields and bad id pattern
	scenario := map[string]interface{}{"id": "bad-id"}
	data, _ := json.Marshal(scenario)
	os.WriteFile(filepath.Join(holdoutDir, "bad.json"), data, 0644)

	_, err := executeCommand("scenario", "validate")
	if err == nil {
		t.Fatal("validate should fail for invalid schema")
	}
}

func TestScenarioValidate_MalformedJSON(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer os.Chdir(origDir)

	holdoutDir := filepath.Join(".agents", "holdout")
	os.MkdirAll(holdoutDir, 0755)

	os.WriteFile(filepath.Join(holdoutDir, "bad.json"), []byte("{invalid"), 0644)

	_, err := executeCommand("scenario", "validate")
	if err == nil {
		t.Fatal("validate should fail for malformed JSON")
	}
}

func TestScenarioValidate_NoDirectory(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer os.Chdir(origDir)

	out, err := executeCommand("scenario", "validate")
	if err != nil {
		t.Fatalf("validate should not error when dir missing: %v", err)
	}
	if !strings.Contains(out, "No holdout directory found") {
		t.Fatalf("expected missing-dir message, got: %s", out)
	}
}

func TestScenarioValidate_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer os.Chdir(origDir)

	os.MkdirAll(filepath.Join(".agents", "holdout"), 0755)

	out, err := executeCommand("scenario", "validate")
	if err != nil {
		t.Fatalf("validate should not error on empty dir: %v", err)
	}
	if !strings.Contains(out, "No scenario files found") {
		t.Fatalf("expected empty message, got: %s", out)
	}
}
