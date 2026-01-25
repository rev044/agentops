package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/boshu2/agentops/cli/internal/storage"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize AgentOps storage structure",
	Long: `Create the .agents/ao directory structure for knowledge storage.

This creates:
  .agents/ao/sessions/    - Session markdown and JSONL files
  .agents/ao/index/       - Session index for quick lookup
  .agents/ao/provenance/  - Provenance tracking graph

Run this in your project root to enable knowledge compounding.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	// Use current directory as base
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	baseDir := filepath.Join(cwd, storage.DefaultBaseDir)

	// Check if already initialized
	if _, err := os.Stat(baseDir); err == nil {
		VerbosePrintf("Directory %s already exists\n", baseDir)
	}

	// Create storage and initialize
	fs := storage.NewFileStorage(storage.WithBaseDir(baseDir))
	if err := fs.Init(); err != nil {
		return fmt.Errorf("initialize storage: %w", err)
	}

	fmt.Printf("✓ Initialized AgentOps storage at %s\n", baseDir)
	fmt.Println()
	fmt.Println("Created directories:")
	fmt.Printf("  %s/sessions/    - Session files\n", storage.DefaultBaseDir)
	fmt.Printf("  %s/index/       - Session index\n", storage.DefaultBaseDir)
	fmt.Printf("  %s/provenance/  - Provenance graph\n", storage.DefaultBaseDir)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  ao forge transcript <path.jsonl>  - Extract knowledge from transcript")

	return nil
}
