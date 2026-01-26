// Package ratchet implements the Brownian Ratchet workflow tracking.
// This file implements CASS (Contextual Agent Session Search) maturity transitions.
package ratchet

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/boshu2/agentops/cli/internal/types"
)

// MaturityTransitionResult contains the result of a maturity transition check.
type MaturityTransitionResult struct {
	// LearningID is the identifier of the learning.
	LearningID string `json:"learning_id"`

	// OldMaturity is the maturity before transition.
	OldMaturity types.Maturity `json:"old_maturity"`

	// NewMaturity is the maturity after transition.
	NewMaturity types.Maturity `json:"new_maturity"`

	// Transitioned indicates if a transition occurred.
	Transitioned bool `json:"transitioned"`

	// Reason explains why the transition did or didn't occur.
	Reason string `json:"reason"`

	// Utility is the current utility value.
	Utility float64 `json:"utility"`

	// Confidence is the current confidence value.
	Confidence float64 `json:"confidence"`

	// HelpfulCount is the number of helpful feedback events.
	HelpfulCount int `json:"helpful_count"`

	// HarmfulCount is the number of harmful feedback events.
	HarmfulCount int `json:"harmful_count"`

	// RewardCount is the total number of feedback events.
	RewardCount int `json:"reward_count"`
}

// CheckMaturityTransition evaluates if a learning should transition to a new maturity level.
// Transition rules:
//   - provisional → candidate: utility >= 0.7 AND reward_count >= 3
//   - candidate → established: utility >= 0.7 AND reward_count >= 5 AND helpful_count > harmful_count
//   - any → anti-pattern: utility <= 0.2 AND harmful_count >= 5
//   - established → candidate: utility < 0.5 (demotion)
//   - candidate → provisional: utility < 0.3 (demotion)
func CheckMaturityTransition(learningPath string) (*MaturityTransitionResult, error) {
	// Read the learning file
	content, err := os.ReadFile(learningPath)
	if err != nil {
		return nil, fmt.Errorf("read learning: %w", err)
	}

	// Parse JSONL (first line)
	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		return nil, fmt.Errorf("empty learning file")
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &data); err != nil {
		return nil, fmt.Errorf("parse learning: %w", err)
	}

	// Extract current values
	learningID := filepath.Base(learningPath)
	if id, ok := data["id"].(string); ok {
		learningID = id
	}

	currentMaturity := types.MaturityProvisional
	if m, ok := data["maturity"].(string); ok && m != "" {
		currentMaturity = types.Maturity(m)
	}

	utility := types.InitialUtility
	if u, ok := data["utility"].(float64); ok {
		utility = u
	}

	confidence := 0.5
	if c, ok := data["confidence"].(float64); ok {
		confidence = c
	}

	rewardCount := 0
	if rc, ok := data["reward_count"].(float64); ok {
		rewardCount = int(rc)
	}

	helpfulCount := 0
	if hc, ok := data["helpful_count"].(float64); ok {
		helpfulCount = int(hc)
	}

	harmfulCount := 0
	if hc, ok := data["harmful_count"].(float64); ok {
		harmfulCount = int(hc)
	}

	result := &MaturityTransitionResult{
		LearningID:   learningID,
		OldMaturity:  currentMaturity,
		NewMaturity:  currentMaturity,
		Transitioned: false,
		Utility:      utility,
		Confidence:   confidence,
		HelpfulCount: helpfulCount,
		HarmfulCount: harmfulCount,
		RewardCount:  rewardCount,
	}

	// Check for anti-pattern transition (takes priority)
	if utility <= types.MaturityAntiPatternThreshold && harmfulCount >= types.MinFeedbackForAntiPattern {
		result.NewMaturity = types.MaturityAntiPattern
		result.Transitioned = currentMaturity != types.MaturityAntiPattern
		result.Reason = fmt.Sprintf("utility %.2f <= %.2f and harmful_count %d >= %d",
			utility, types.MaturityAntiPatternThreshold, harmfulCount, types.MinFeedbackForAntiPattern)
		return result, nil
	}

	// Check transitions based on current maturity
	switch currentMaturity {
	case types.MaturityProvisional:
		// provisional → candidate
		if utility >= types.MaturityPromotionThreshold && rewardCount >= types.MinFeedbackForPromotion {
			result.NewMaturity = types.MaturityCandidate
			result.Transitioned = true
			result.Reason = fmt.Sprintf("utility %.2f >= %.2f and reward_count %d >= %d",
				utility, types.MaturityPromotionThreshold, rewardCount, types.MinFeedbackForPromotion)
		} else {
			result.Reason = "not enough positive feedback for promotion"
		}

	case types.MaturityCandidate:
		// candidate → established (promotion)
		if utility >= types.MaturityPromotionThreshold && rewardCount >= 5 && helpfulCount > harmfulCount {
			result.NewMaturity = types.MaturityEstablished
			result.Transitioned = true
			result.Reason = fmt.Sprintf("utility %.2f >= %.2f, reward_count %d >= 5, helpful > harmful (%d > %d)",
				utility, types.MaturityPromotionThreshold, rewardCount, helpfulCount, harmfulCount)
		} else if utility < types.MaturityDemotionThreshold {
			// candidate → provisional (demotion)
			result.NewMaturity = types.MaturityProvisional
			result.Transitioned = true
			result.Reason = fmt.Sprintf("utility %.2f < %.2f (demotion)",
				utility, types.MaturityDemotionThreshold)
		} else {
			result.Reason = "maintaining candidate status"
		}

	case types.MaturityEstablished:
		// established → candidate (demotion)
		if utility < 0.5 {
			result.NewMaturity = types.MaturityCandidate
			result.Transitioned = true
			result.Reason = fmt.Sprintf("utility %.2f < 0.5 (demotion from established)",
				utility)
		} else {
			result.Reason = "maintaining established status"
		}

	case types.MaturityAntiPattern:
		// Anti-patterns can be rehabilitated if utility improves significantly
		if utility >= 0.6 && helpfulCount > harmfulCount*2 {
			result.NewMaturity = types.MaturityProvisional
			result.Transitioned = true
			result.Reason = fmt.Sprintf("utility %.2f >= 0.6 and helpful > 2*harmful (%d > %d) - rehabilitation",
				utility, helpfulCount, harmfulCount*2)
		} else {
			result.Reason = "maintaining anti-pattern status"
		}
	}

	return result, nil
}

// ApplyMaturityTransition checks and applies a maturity transition to a learning file.
// Returns the transition result and updates the file if a transition occurred.
func ApplyMaturityTransition(learningPath string) (*MaturityTransitionResult, error) {
	result, err := CheckMaturityTransition(learningPath)
	if err != nil {
		return nil, err
	}

	if !result.Transitioned {
		return result, nil
	}

	// Read and update the file
	content, err := os.ReadFile(learningPath)
	if err != nil {
		return nil, fmt.Errorf("read learning for update: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty file")
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &data); err != nil {
		return nil, fmt.Errorf("parse learning for update: %w", err)
	}

	// Update maturity and timestamp
	data["maturity"] = string(result.NewMaturity)
	data["maturity_changed_at"] = time.Now().Format(time.RFC3339)
	data["maturity_reason"] = result.Reason

	// Write back
	newJSON, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal updated learning: %w", err)
	}

	lines[0] = string(newJSON)
	if err := os.WriteFile(learningPath, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return nil, fmt.Errorf("write updated learning: %w", err)
	}

	return result, nil
}

// ScanForMaturityTransitions scans a learnings directory and returns all pending transitions.
func ScanForMaturityTransitions(learningsDir string) ([]*MaturityTransitionResult, error) {
	files, err := filepath.Glob(filepath.Join(learningsDir, "*.jsonl"))
	if err != nil {
		return nil, fmt.Errorf("glob learnings: %w", err)
	}

	var results []*MaturityTransitionResult

	for _, file := range files {
		result, err := CheckMaturityTransition(file)
		if err != nil {
			continue // Skip files that can't be parsed
		}

		// Only include learnings that would transition
		if result.Transitioned {
			results = append(results, result)
		}
	}

	return results, nil
}

// GetAntiPatterns returns all learnings marked as anti-patterns.
func GetAntiPatterns(learningsDir string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(learningsDir, "*.jsonl"))
	if err != nil {
		return nil, fmt.Errorf("glob learnings: %w", err)
	}

	var antiPatterns []string

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		if scanner.Scan() {
			var data map[string]interface{}
			if err := json.Unmarshal(scanner.Bytes(), &data); err != nil {
				f.Close()
				continue
			}

			if maturity, ok := data["maturity"].(string); ok && maturity == string(types.MaturityAntiPattern) {
				antiPatterns = append(antiPatterns, file)
			}
		}
		f.Close()
	}

	return antiPatterns, nil
}

// GetEstablishedLearnings returns all learnings with established maturity.
func GetEstablishedLearnings(learningsDir string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(learningsDir, "*.jsonl"))
	if err != nil {
		return nil, fmt.Errorf("glob learnings: %w", err)
	}

	var established []string

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		if scanner.Scan() {
			var data map[string]interface{}
			if err := json.Unmarshal(scanner.Bytes(), &data); err != nil {
				f.Close()
				continue
			}

			if maturity, ok := data["maturity"].(string); ok && maturity == string(types.MaturityEstablished) {
				established = append(established, file)
			}
		}
		f.Close()
	}

	return established, nil
}

// MaturityDistribution represents the count of learnings at each maturity level.
type MaturityDistribution struct {
	Provisional int `json:"provisional"`
	Candidate   int `json:"candidate"`
	Established int `json:"established"`
	AntiPattern int `json:"anti_pattern"`
	Unknown     int `json:"unknown"`
	Total       int `json:"total"`
}

// GetMaturityDistribution returns the distribution of learnings across maturity levels.
func GetMaturityDistribution(learningsDir string) (*MaturityDistribution, error) {
	files, err := filepath.Glob(filepath.Join(learningsDir, "*.jsonl"))
	if err != nil {
		return nil, fmt.Errorf("glob learnings: %w", err)
	}

	dist := &MaturityDistribution{}

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		if scanner.Scan() {
			var data map[string]interface{}
			if err := json.Unmarshal(scanner.Bytes(), &data); err != nil {
				f.Close()
				dist.Unknown++
				dist.Total++
				continue
			}

			maturity, ok := data["maturity"].(string)
			if !ok || maturity == "" {
				maturity = string(types.MaturityProvisional)
			}

			switch types.Maturity(maturity) {
			case types.MaturityProvisional:
				dist.Provisional++
			case types.MaturityCandidate:
				dist.Candidate++
			case types.MaturityEstablished:
				dist.Established++
			case types.MaturityAntiPattern:
				dist.AntiPattern++
			default:
				dist.Unknown++
			}
			dist.Total++
		}
		f.Close()
	}

	return dist, nil
}
