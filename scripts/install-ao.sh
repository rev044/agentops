#!/bin/bash
# Install the ao CLI from source
# Usage: ./scripts/install-ao.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$REPO_ROOT/cli"

echo "Building ao CLI..."
go build -o ao ./cmd/ao

mkdir -p ~/.local/bin
mv ao ~/.local/bin/ao

echo "Installed ao to ~/.local/bin/ao"
echo "Make sure ~/.local/bin is in your PATH"
