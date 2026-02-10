#!/usr/bin/env bash
set -euo pipefail

# AgentOps Installer
# Usage: Download and run this script, or execute via npx skills@latest add boshu2/agentops --all -g

echo "Installing AgentOps..."

# Check prerequisites
command -v npm >/dev/null 2>&1 || { echo "Error: npm required. Install Node.js first."; exit 1; }
command -v claude >/dev/null 2>&1 || { echo "Error: Claude Code CLI required. See https://docs.anthropic.com/en/docs/claude-code"; exit 1; }

# Step 1: Install plugin (skills + hooks + agents)
echo "Step 1/3: Installing plugin..."
npx skills@latest add boshu2/agentops --all -g

# Step 2: Install CLI (optional — enhances with knowledge flywheel)
if command -v brew >/dev/null 2>&1; then
    echo "Step 2/3: Installing CLI via Homebrew..."
    brew tap boshu2/agentops https://github.com/boshu2/homebrew-agentops 2>/dev/null || true
    brew install agentops 2>/dev/null || brew upgrade agentops 2>/dev/null || true

    # Step 3: Install hooks
    if command -v ao >/dev/null 2>&1; then
        echo "Step 3/3: Installing hooks..."
        ao hooks install

        # Optional health check
        ao doctor 2>/dev/null && echo "Health check: PASS" || echo "Health check: run 'ao doctor' after setup"
    fi
else
    echo "Step 2/3: Skipping CLI (Homebrew not found). Install manually: brew install agentops"
    echo "Step 3/3: Skipped (CLI needed for hooks)"
fi

echo ""
echo "Done! Start with: /quickstart"
