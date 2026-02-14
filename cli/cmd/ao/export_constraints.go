package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var exportConstraintsCmd = &cobra.Command{
	Use:   "export-constraints",
	Short: "Export GOALS.yaml constraints for external consumers",
	Long: `Export the project's GOALS.yaml fitness constraints in machine-readable format.

Outputs goal definitions with check commands, weights, and pillar assignments
for consumption by CI pipelines, dashboards, or external validation tools.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("export-constraints: not yet implemented")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(exportConstraintsCmd)
}
