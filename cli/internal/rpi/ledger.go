package rpi

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	// LedgerSchemaVersion is the current schema version for ledger records.
	LedgerSchemaVersion = 1
	// LedgerRelativePath is the repo-relative path to the RPI ledger file.
	LedgerRelativePath = ".agents/ledger/rpi-events.jsonl"
	// RunCacheRelativeDir is the repo-relative path for run cache files.
	RunCacheRelativeDir = ".agents/rpi/runs"
)

// LedgerRecord is a single append-only event in the RPI ledger.
type LedgerRecord struct {
	SchemaVersion int             `json:"schema_version"`
	EventID       string          `json:"event_id"`
	RunID         string          `json:"run_id"`
	TS            string          `json:"ts"`
	Phase         string          `json:"phase"`
	Action        string          `json:"action"`
	Details       json.RawMessage `json:"details"`
	PrevHash      string          `json:"prev_hash"`
	PayloadHash   string          `json:"payload_hash"`
	Hash          string          `json:"hash"`
}

// LedgerAppendInput contains fields needed for appending an event.
type LedgerAppendInput struct {
	RunID   string
	Phase   string
	Action  string
	Details any
}

// RunCache is a materialized cache of the latest state for one run.
type RunCache struct {
	RunID      string       `json:"run_id"`
	EventCount int          `json:"event_count"`
	Latest     LedgerRecord `json:"latest"`
	UpdatedAt  string       `json:"updated_at"`
}

// LedgerVerifyResult is the machine-readable verify output contract.
type LedgerVerifyResult struct {
	Pass             bool   `json:"pass"`
	RecordCount      int    `json:"record_count"`
	FirstBrokenIndex int    `json:"first_broken_index"`
	Message          string `json:"message,omitempty"`
}

// LedgerPayload is the hash-input subset of a ledger record.
type LedgerPayload struct {
	SchemaVersion int             `json:"schema_version"`
	EventID       string          `json:"event_id"`
	RunID         string          `json:"run_id"`
	TS            string          `json:"ts"`
	Phase         string          `json:"phase"`
	Action        string          `json:"action"`
	Details       json.RawMessage `json:"details"`
	PrevHash      string          `json:"prev_hash"`
}

// ValidateAppendInput validates required fields on the append input.
func ValidateAppendInput(input LedgerAppendInput) error {
	requiredFields := []struct {
		value string
		name  string
	}{
		{input.RunID, "run_id"},
		{input.Phase, "phase"},
		{input.Action, "action"},
	}
	for _, f := range requiredFields {
		if strings.TrimSpace(f.value) == "" {
			return fmt.Errorf("%s is required", f.name)
		}
	}
	return nil
}

// ValidateRunID checks that a run ID is non-empty and contains no path traversal.
func ValidateRunID(runID string) error {
	if strings.TrimSpace(runID) == "" {
		return fmt.Errorf("run_id is required")
	}
	if strings.Contains(runID, "/") || strings.Contains(runID, "\\") || strings.Contains(runID, "..") {
		return fmt.Errorf("run_id contains invalid path elements")
	}
	return nil
}

// ValidateLedgerRecord validates a single ledger record's structural integrity.
func ValidateLedgerRecord(record LedgerRecord) error {
	if record.SchemaVersion != LedgerSchemaVersion {
		return fmt.Errorf("schema_version mismatch: got %d want %d", record.SchemaVersion, LedgerSchemaVersion)
	}
	if err := validateLedgerRequiredFields(record); err != nil {
		return err
	}
	if err := ValidateLedgerTimestamp(record.TS); err != nil {
		return err
	}
	if _, err := NormalizeDetails(record.Details); err != nil {
		return err
	}
	return nil
}

func validateLedgerRequiredFields(record LedgerRecord) error {
	fields := []struct {
		value string
		name  string
	}{
		{record.EventID, "event_id"},
		{record.RunID, "run_id"},
		{record.Phase, "phase"},
		{record.Action, "action"},
		{record.TS, "ts"},
		{record.PayloadHash, "payload_hash"},
		{record.Hash, "hash"},
	}
	for _, f := range fields {
		if strings.TrimSpace(f.value) == "" {
			return fmt.Errorf("%s is required", f.name)
		}
	}
	return nil
}

// ValidateLedgerTimestamp checks that a timestamp is valid UTC RFC3339Nano.
func ValidateLedgerTimestamp(ts string) error {
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return fmt.Errorf("invalid ts: %w", err)
	}
	if t.UTC().Format(time.RFC3339Nano) != ts {
		return fmt.Errorf("ts must be UTC RFC3339Nano")
	}
	return nil
}

// ComputeLedgerHashes computes payload_hash and hash for a ledger record.
func ComputeLedgerHashes(record LedgerRecord) (payloadHash string, hashValue string, err error) {
	details, err := NormalizeDetails(record.Details)
	if err != nil {
		return "", "", err
	}
	payload := LedgerPayload{
		SchemaVersion: record.SchemaVersion,
		EventID:       record.EventID,
		RunID:         record.RunID,
		TS:            record.TS,
		Phase:         record.Phase,
		Action:        record.Action,
		Details:       details,
		PrevHash:      record.PrevHash,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", "", fmt.Errorf("marshal payload: %w", err)
	}
	payloadHash = HashHex(payloadBytes)
	hashValue = HashHex([]byte(payloadHash + "\n" + record.PrevHash))
	return payloadHash, hashValue, nil
}

// NormalizeDetails canonicalizes a details value to deterministic JSON.
func NormalizeDetails(details any) (json.RawMessage, error) {
	if details == nil {
		return json.RawMessage([]byte("{}")), nil
	}

	if raw, ok := details.(json.RawMessage); ok {
		details = []byte(raw)
	}

	switch v := details.(type) {
	case []byte:
		return normalizeDetailsBytes(v)
	default:
		return normalizeDetailsValue(v)
	}
}

func normalizeDetailsBytes(v []byte) (json.RawMessage, error) {
	if len(bytes.TrimSpace(v)) == 0 {
		return json.RawMessage([]byte("{}")), nil
	}
	return roundTripJSON(v)
}

func normalizeDetailsValue(v any) (json.RawMessage, error) {
	encoded, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal details: %w", err)
	}
	return roundTripJSON(encoded)
}

func roundTripJSON(data []byte) (json.RawMessage, error) {
	var parsed any
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("details must be valid JSON: %w", err)
	}
	normalized, err := json.Marshal(parsed)
	if err != nil {
		return nil, fmt.Errorf("marshal details: %w", err)
	}
	return json.RawMessage(normalized), nil
}

// VerifyLedgerChain verifies hashes and prev-hash links for all records.
func VerifyLedgerChain(records []LedgerRecord) error {
	prevHash := ""
	for i, record := range records {
		if err := ValidateLedgerRecord(record); err != nil {
			return fmt.Errorf("record %d: %w", i+1, err)
		}
		if record.PrevHash != prevHash {
			return fmt.Errorf("record %d: prev_hash mismatch: got %q want %q", i+1, record.PrevHash, prevHash)
		}

		payloadHash, hashValue, err := ComputeLedgerHashes(record)
		if err != nil {
			return fmt.Errorf("record %d: %w", i+1, err)
		}
		if record.PayloadHash != payloadHash {
			return fmt.Errorf("record %d: payload_hash mismatch", i+1)
		}
		if record.Hash != hashValue {
			return fmt.Errorf("record %d: hash mismatch", i+1)
		}
		prevHash = record.Hash
	}
	return nil
}

// VerifyLedgerDetailed verifies the chain and returns a structured result
// without failing for chain mismatches.
func VerifyLedgerDetailed(records []LedgerRecord) LedgerVerifyResult {
	result := LedgerVerifyResult{
		Pass:             true,
		RecordCount:      len(records),
		FirstBrokenIndex: -1,
	}

	prevHash := ""
	for i, record := range records {
		if err := ValidateLedgerRecord(record); err != nil {
			result.Pass = false
			result.FirstBrokenIndex = i + 1
			result.Message = err.Error()
			return result
		}
		if record.PrevHash != prevHash {
			result.Pass = false
			result.FirstBrokenIndex = i + 1
			result.Message = fmt.Sprintf("prev_hash mismatch: got %q want %q", record.PrevHash, prevHash)
			return result
		}

		payloadHash, hashValue, err := ComputeLedgerHashes(record)
		if err != nil {
			result.Pass = false
			result.FirstBrokenIndex = i + 1
			result.Message = err.Error()
			return result
		}
		if record.PayloadHash != payloadHash {
			result.Pass = false
			result.FirstBrokenIndex = i + 1
			result.Message = "payload_hash mismatch"
			return result
		}
		if record.Hash != hashValue {
			result.Pass = false
			result.FirstBrokenIndex = i + 1
			result.Message = "hash mismatch"
			return result
		}
		prevHash = record.Hash
	}

	return result
}

// FilterRunRecords returns the latest record and count for a given run ID.
func FilterRunRecords(records []LedgerRecord, runID string) (LedgerRecord, int) {
	var latest LedgerRecord
	count := 0
	for _, record := range records {
		if record.RunID != runID {
			continue
		}
		latest = record
		count++
	}
	return latest, count
}

// BuildRunCache constructs a RunCache struct from a latest record and count.
func BuildRunCache(runID string, latest LedgerRecord, count int) RunCache {
	return RunCache{
		RunID:      runID,
		EventCount: count,
		Latest:     latest,
		UpdatedAt:  time.Now().UTC().Format(time.RFC3339Nano),
	}
}

// NewLedgerEventID generates a random event ID.
func NewLedgerEventID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("evt-%d", time.Now().UnixNano())
	}
	return "evt-" + hex.EncodeToString(b[:])
}

// HashHex returns the hex-encoded SHA-256 hash of data.
func HashHex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// BuildLedgerRecordFromInput creates a complete LedgerRecord from input and previous hash.
func BuildLedgerRecordFromInput(input LedgerAppendInput, prevHash string) (LedgerRecord, error) {
	details, err := NormalizeDetails(input.Details)
	if err != nil {
		return LedgerRecord{}, err
	}

	record := LedgerRecord{
		SchemaVersion: LedgerSchemaVersion,
		EventID:       NewLedgerEventID(),
		RunID:         input.RunID,
		TS:            time.Now().UTC().Format(time.RFC3339Nano),
		Phase:         input.Phase,
		Action:        input.Action,
		Details:       details,
		PrevHash:      prevHash,
	}

	payloadHash, hashValue, err := ComputeLedgerHashes(record)
	if err != nil {
		return LedgerRecord{}, err
	}
	record.PayloadHash = payloadHash
	record.Hash = hashValue

	return record, nil
}

// ParseLedgerLine parses a single JSONL line into a LedgerRecord.
func ParseLedgerLine(line string) (LedgerRecord, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return LedgerRecord{}, fmt.Errorf("empty line")
	}
	var record LedgerRecord
	if err := json.Unmarshal([]byte(line), &record); err != nil {
		return LedgerRecord{}, err
	}
	return record, nil
}
