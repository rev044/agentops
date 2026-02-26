package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// deprecatedAlias creates a hidden command that forwards execution to a new command path.
// It prints a deprecation warning to stderr before executing.
func deprecatedAlias(oldUse, newPath string, target *cobra.Command) *cobra.Command {
	alias := &cobra.Command{
		Use:    oldUse,
		Hidden: true,
		Short:  fmt.Sprintf("DEPRECATED: use '%s' instead", newPath),
		Long:   fmt.Sprintf("This command has been moved to '%s'.\nPlease update your scripts and workflows.", newPath),
		// Silence usage on error — the alias should be transparent
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(os.Stderr, "DEPRECATED: 'ao %s' has moved to '%s'\n", oldUse, newPath)
			// Forward to target's RunE
			if target.RunE != nil {
				return target.RunE(target, args)
			}
			if target.Run != nil {
				target.Run(target, args)
				return nil
			}
			return fmt.Errorf("target command '%s' has no run function", newPath)
		},
	}

	// Copy flags from target so the alias accepts the same flags
	alias.Flags().AddFlagSet(target.Flags())
	alias.PersistentFlags().AddFlagSet(target.PersistentFlags())

	return alias
}
