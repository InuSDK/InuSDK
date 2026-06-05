#!/usr/bin/env bash
set -e

SHIMS_DIR="$HOME/.inusdk/shims"

echo "Building InuSDK..."
go build -o build/inusdk .

echo "Copying to shims..."
mkdir -p "$SHIMS_DIR"
cp build/inusdk "$SHIMS_DIR/inusdk"
chmod +x "$SHIMS_DIR/inusdk"

echo "Done ; Run 'inusdk' in a new terminal."
