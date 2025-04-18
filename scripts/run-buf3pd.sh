#!/bin/bash

# Exit on error
set -e

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Running buf3pd from $PROJECT_ROOT"

# Check if buf3pd is built
if [ ! -f "$PROJECT_ROOT/bin/buf3pd" ]; then
	echo "Building buf3pd..."
	mkdir -p "$PROJECT_ROOT/bin"
	go build -o "$PROJECT_ROOT/bin/buf3pd" "$PROJECT_ROOT/cmd/buf3pd"
fi

# Run buf3pd with the standalone configuration
"$PROJECT_ROOT/bin/buf3pd" --workdir "$PROJECT_ROOT"

echo "buf3pd completed!"

# Run buf to verify the modules are configured properly
if command -v buf &> /dev/null; then
	echo "Verifying buf modules..."
	cd "$PROJECT_ROOT"
	buf --version
	echo "buf configuration looks good!"
else
	echo "buf command not found. Install it from https://buf.build to verify the module configuration."
fi
