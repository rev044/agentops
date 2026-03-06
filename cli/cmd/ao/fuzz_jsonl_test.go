package main

import (
	"os"
	"path/filepath"
	"testing"
)

// FuzzParseLearningJSONL fuzzes the JSONL learning parser with arbitrary data.
// parseLearningJSONL reads from a file path, so we write fuzz data to a temp file.
func FuzzParseLearningJSONL(f *testing.F) {
	// Seed corpus with realistic learning formats
	f.Add(`{"summary":"test learning","confidence":0.8,"tags":["test"]}`)
	f.Add(`{}`)
	f.Add(``)
	f.Add(`{"summary":""}`)
	f.Add(`{"summary":"multi-line\nlearning","confidence":1.0,"tags":["a","b"],"utility":0.9}`)
	f.Add(`{"superseded_by":"other-id"}`)
	f.Add(`{"summary":"has source","source_bead":"bd-123","source_phase":"research","maturity":"validated"}`)
	f.Add(`not json at all`)
	f.Add(`{"summary":"` + string(make([]byte, 1024)) + `"}`)

	f.Fuzz(func(t *testing.T, data string) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.jsonl")
		if err := os.WriteFile(path, []byte(data), 0644); err != nil {
			return
		}
		// Must never panic — errors are acceptable
		_, _ = parseLearningJSONL(path)
	})
}

// FuzzParseJSONLFirstLine fuzzes the first-line JSONL parser used by feedback.
func FuzzParseJSONLFirstLine(f *testing.F) {
	// Seed corpus with realistic JSONL first lines
	f.Add(`{"utility":0.5,"reward_count":3}`)
	f.Add(`{}`)
	f.Add(``)
	f.Add(`{"key":"value","nested":{"a":1}}`)
	f.Add("not json\n{\"second\":\"line\"}")
	f.Add(`{"utility":0.5}` + "\n" + `{"second":"line"}`)
	f.Add(`null`)
	f.Add(`[]`)

	f.Fuzz(func(t *testing.T, data string) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.jsonl")
		if err := os.WriteFile(path, []byte(data), 0644); err != nil {
			return
		}
		// Must never panic — errors are acceptable
		_, _, _ = parseJSONLFirstLine(path)
	})
}

// FuzzParseJSONLSessionSummary fuzzes the session summary JSONL parser.
func FuzzParseJSONLSessionSummary(f *testing.F) {
	// Seed corpus with realistic session formats
	f.Add(`{"summary":"Session about testing fuzz targets"}`)
	f.Add(`{}`)
	f.Add(``)
	f.Add(`{"summary":""}`)
	f.Add(`{"summary":"a]very long summary that exceeds the typical truncation length and should be handled gracefully by the parser without any panics or crashes regardless of content"}`)
	f.Add(`{"no_summary_field":true}`)
	f.Add(`not json`)
	f.Add(`{"summary":123}`)

	f.Fuzz(func(t *testing.T, data string) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.jsonl")
		if err := os.WriteFile(path, []byte(data), 0644); err != nil {
			return
		}
		// Must never panic — errors are acceptable
		_, _ = parseJSONLSessionSummary(path)
	})
}
