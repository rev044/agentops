#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR/cli"

echo "Building ol CLI..."
go build -o ol ./cmd/ol

mkdir -p ~/.local/bin
mv ol ~/.local/bin/ol

echo "Installed ol to ~/.local/bin/ol"
echo "Make sure ~/.local/bin is in your PATH"
