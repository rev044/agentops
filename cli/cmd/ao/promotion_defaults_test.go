package main

import (
	"testing"
	"time"
)

func TestPromotionDefaults_AutoPromoteThresholdParseable(t *testing.T) {
	// Verify the default constant is a valid Go duration string.
	d, err := time.ParseDuration(defaultAutoPromoteThreshold)
	if err != nil {
		t.Fatalf("defaultAutoPromoteThreshold %q is not a valid duration: %v", defaultAutoPromoteThreshold, err)
	}
	if d <= 0 {
		t.Errorf("defaultAutoPromoteThreshold duration = %v, want positive", d)
	}
}

func TestPromotionDefaults_AutoPromoteThresholdIs24h(t *testing.T) {
	if defaultAutoPromoteThreshold != "24h" {
		t.Errorf("defaultAutoPromoteThreshold = %q, want %q", defaultAutoPromoteThreshold, "24h")
	}
}
