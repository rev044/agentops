package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var exportConstraintsCmd = &cobra.Command{
	Use:   "export-constraints",
	Short: "Export AO anti-patterns as OL constraints",
	Long: `Export learnings with type 'failure' or maturity 'anti-pattern' from
.agents/learnings/ as Olympus-compatible constraints.

Output format: .ol/constraints/quarantine.json

This enables the AO→OL constraint propagation bridge defined in
docs/ol-bridge-contracts.md (Section 1, AO → OL).

Status: STUB — not yet implemented. See bridge contracts doc for spec.`,
	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")
		if format != "ol" && format != "" {
			fmt.Fprintf(os.Stderr, "unsupported format: %s (only 'ol' is planned)\n", format)
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "ao export-constraints: not yet implemented")
		fmt.Fprintln(os.Stderr, "See docs/ol-bridge-contracts.md for the interchange spec.")
		os.Exit(2)
	},
}

func init() {
	exportConstraintsCmd.Flags().String("format", "ol", "Output format (only 'ol' supported)")
	rootCmd.AddCommand(exportConstraintsCmd)
}
