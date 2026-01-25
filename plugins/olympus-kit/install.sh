#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR/cli"

echo "Building ao CLI..."
go build -o ao ./cmd/ao

mkdir -p ~/.local/bin
mv ao ~/.local/bin/ao

echo "Installed ao to ~/.local/bin/ao"
echo "Make sure ~/.local/bin is in your PATH"
