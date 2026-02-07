package main

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/boshu2/agentops/cli/internal/types"
	"github.com/spf13/cobra"
)

// flywheelCmd provides a convenient alias for flywheel status operations.
var flywheelCmd = &cobra.Command{
	Use:   "flywheel",
	Short: "Knowledge flywheel operations",
	Long: `Knowledge flywheel operations and status.

The flywheel equation:
  dK/dt = I(t) - δ·K + σ·ρ·K - B(K, K_crit)

Escape velocity: σρ > δ → Knowledge compounds

Commands:
  status   Show comprehensive flywheel health

Examples:
  ao flywheel status
  ao flywheel status -o json`,
}

func init() {
	rootCmd.AddCommand(flywheelCmd)

	// flywheel status subcommand
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show flywheel health status",
		Long: `Display comprehensive flywheel health status.

Shows:
  - Delta (δ): Knowledge decay rate
  - Sigma (σ): Retrieval effectiveness
  - Rho (ρ): Citation rate
  - Velocity: σρ - δ (net growth rate)
  - Status: COMPOUNDING / NEAR ESCAPE / DECAYING

Examples:
  ao flywheel status
  ao flywheel status --days 30
  ao flywheel status -o json`,
		RunE: runFlywheelStatus,
	}
	statusCmd.Flags().IntVar(&metricsDays, "days", 7, "Period in days for metrics calculation")
	flywheelCmd.AddCommand(statusCmd)
}

// runFlywheelStatus displays comprehensive flywheel health.
func runFlywheelStatus(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	metrics, err := computeMetrics(cwd, metricsDays)
	if err != nil {
		return fmt.Errorf("compute metrics: %w", err)
	}

	switch GetOutput() {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]interface{}{
			"status":      metrics.EscapeVelocityStatus(),
			"delta":       metrics.Delta,
			"sigma":       metrics.Sigma,
			"rho":         metrics.Rho,
			"sigma_rho":   metrics.SigmaRho,
			"velocity":    metrics.Velocity,
			"compounding": metrics.AboveEscapeVelocity,
			"metrics":     metrics,
		})

	case "yaml":
		enc := yaml.NewEncoder(os.Stdout)
		return enc.Encode(map[string]interface{}{
			"status":      metrics.EscapeVelocityStatus(),
			"delta":       metrics.Delta,
			"sigma":       metrics.Sigma,
			"rho":         metrics.Rho,
			"sigma_rho":   metrics.SigmaRho,
			"velocity":    metrics.Velocity,
			"compounding": metrics.AboveEscapeVelocity,
		})

	default:
		printFlywheelStatus(metrics)
	}

	return nil
}

// printFlywheelStatus prints a focused flywheel status display.
func printFlywheelStatus(m *types.FlywheelMetrics) {
	status := m.EscapeVelocityStatus()

	// Status indicator (ASCII for accessibility)
	var statusIcon string
	switch status {
	case "COMPOUNDING":
		statusIcon = "[COMPOUNDING]"
	case "NEAR ESCAPE":
		statusIcon = "[NEAR_ESCAPE]"
	default:
		statusIcon = "[DECAYING]"
	}

	fmt.Println()
	fmt.Printf("  Flywheel Status: %s\n", statusIcon)
	fmt.Println("  ═══════════════════════════════")
	fmt.Println()

	// Core equation
	fmt.Println("  EQUATION: dK/dt = I(t) - δ·K + σ·ρ·K")
	fmt.Println()

	// Parameters
	fmt.Printf("  δ (decay):      %.2f/week\n", m.Delta)
	fmt.Printf("  σ (retrieval):  %.2f (%d%% of artifacts surfaced)\n", m.Sigma, int(m.Sigma*100))
	fmt.Printf("  ρ (citation):   %.2f refs/artifact/week\n", m.Rho)
	fmt.Println()

	// Critical comparison
	fmt.Println("  ESCAPE VELOCITY CHECK:")
	fmt.Printf("    σ × ρ = %.3f\n", m.SigmaRho)
	fmt.Printf("    δ     = %.3f\n", m.Delta)
	fmt.Println("    ───────────────")

	if m.AboveEscapeVelocity {
		fmt.Printf("    σρ > δ ✓ (velocity: +%.3f/week)\n", m.Velocity)
		fmt.Println("    → Knowledge is COMPOUNDING")
	} else if m.Velocity > -0.05 {
		fmt.Printf("    σρ ≈ δ (velocity: %.3f/week)\n", m.Velocity)
		fmt.Println("    → NEAR escape velocity, keep building!")
	} else {
		fmt.Printf("    σρ < δ ✗ (velocity: %.3f/week)\n", m.Velocity)
		fmt.Println("    → Knowledge is DECAYING")
		fmt.Println()
		fmt.Println("  RECOMMENDATIONS:")
		if m.Sigma < 0.3 {
			fmt.Println("    • Improve retrieval: run 'ao inject' more often")
		}
		if m.Rho < 0.5 {
			fmt.Println("    • Cite more learnings: reference artifacts in your work")
		}
		if m.StaleArtifacts > 5 {
			fmt.Printf("    • Review %d stale artifacts (90+ days uncited)\n", m.StaleArtifacts)
		}
	}

	fmt.Println()
	fmt.Printf("  Period: %s to %s (%d days)\n",
		m.PeriodStart.Format("2006-01-02"),
		m.PeriodEnd.Format("2006-01-02"),
		metricsDays)
}
