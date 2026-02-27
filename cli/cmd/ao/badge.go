package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/boshu2/agentops/cli/internal/storage"
	"github.com/boshu2/agentops/cli/internal/types"
)

var badgeCmd = &cobra.Command{
	Use:   "badge",
	Short: "Display knowledge flywheel health badge",
	Long: `Display a visual badge showing knowledge flywheel health status.

The badge shows:
  - Session and artifact counts
  - Core flywheel parameters (σ, ρ, δ)
  - Escape velocity calculation and status

Status levels:
  🚀 ESCAPE VELOCITY  - σ×ρ > δ (knowledge compounds)
  ⚡ APPROACHING      - σ×ρ > δ×0.8 (almost there)
  📈 BUILDING         - σ×ρ > δ×0.5 (making progress)
  🌱 STARTING         - σ×ρ ≤ δ×0.5 (early stage)

Example:
  ao badge`,
	RunE: runBadge,
}

func init() {
	badgeCmd.GroupID = "core"
	rootCmd.AddCommand(badgeCmd)
}

func runBadge(cmd *cobra.Command, args []string) error {
	cwd, err := resolveProjectDir()
	if err != nil {
		return err
	}

	// Compute metrics (reuse existing logic)
	metrics, err := computeMetrics(cwd, 7)
	if err != nil {
		VerbosePrintf("Warning: compute metrics: %v\n", err)
	}

	// Count sessions mined
	sessionsMined := countSessions(cwd)

	// Draw the badge
	printBadge(sessionsMined, metrics)
	return nil
}

// countSessions counts mined transcript sessions.
func countSessions(baseDir string) int {
	sessionsDir := filepath.Join(baseDir, storage.DefaultBaseDir, storage.SessionsDir)
	if _, err := os.Stat(sessionsDir); os.IsNotExist(err) {
		return 0
	}
	files, _ := filepath.Glob(filepath.Join(sessionsDir, "*.jsonl"))
	return len(files)
}

// printBadge prints the visual badge.
func printBadge(sessions int, m *FlywheelMetrics) {
	if m == nil {
		m = &FlywheelMetrics{Delta: types.DefaultDelta}
	}

	// Calculate status
	status, statusIcon := getEscapeStatus(m.SigmaRho, m.Delta)

	// Progress bars (10 chars width)
	sigmaBar := makeProgressBar(m.Sigma, 10)
	rhoBar := makeProgressBar(m.Rho, 10)
	deltaBar := makeProgressBar(m.Delta, 10)

	// Learnings count (from tier counts)
	learnings := m.TierCounts["learning"]
	patterns := m.TierCounts["pattern"]

	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════╗")
	fmt.Println("║         🏛️  AGENTOPS KNOWLEDGE             ║")
	fmt.Println("╠═══════════════════════════════════════════╣")
	fmt.Printf("║  Sessions Mined    │  %-19d ║\n", sessions)
	fmt.Printf("║  Learnings         │  %-19d ║\n", learnings)
	fmt.Printf("║  Patterns          │  %-19d ║\n", patterns)
	fmt.Printf("║  Citations         │  %-19d ║\n", m.CitationsThisPeriod)
	fmt.Println("╠═══════════════════════════════════════════╣")
	fmt.Printf("║  Retrieval (σ)     │  %.2f  %s ║\n", m.Sigma, sigmaBar)
	fmt.Printf("║  Citation Rate (ρ) │  %.2f  %s ║\n", m.Rho, rhoBar)
	fmt.Printf("║  Decay (δ)         │  %.2f  %s ║\n", m.Delta, deltaBar)
	fmt.Println("╠═══════════════════════════════════════════╣")

	// Final status line
	sigmaRhoStr := fmt.Sprintf("%.2f", m.SigmaRho)
	comparison := ">"
	if m.SigmaRho <= m.Delta {
		comparison = "≤"
	}
	statusLine := fmt.Sprintf("σ×ρ = %s %s δ", sigmaRhoStr, comparison)
	fmt.Printf("║  %-17s │  %s %-13s║\n", statusLine, statusIcon, status)
	fmt.Println("╚═══════════════════════════════════════════╝")
	fmt.Println()
}

// getEscapeStatus returns status text and icon based on velocity.
func getEscapeStatus(sigmaRho, delta float64) (string, string) {
	if sigmaRho > delta {
		return "ESCAPE VELOCITY", "🚀"
	}
	if sigmaRho > delta*0.8 {
		return "APPROACHING", "⚡"
	}
	if sigmaRho > delta*0.5 {
		return "BUILDING", "📈"
	}
	return "STARTING", "🌱"
}

// makeProgressBar creates a visual progress bar.
func makeProgressBar(value float64, width int) string {
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}

	filled := int(value * float64(width))
	empty := width - filled

	var sb strings.Builder
	for range filled {
		sb.WriteString("█")
	}
	for range empty {
		sb.WriteString("░")
	}
	return sb.String()
}

// FlywheelMetrics is imported from types but we use a local alias for brevity
type FlywheelMetrics = types.FlywheelMetrics
