#!/bin/bash
# Build and deploy script for AgentHub

set -e

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "=== AgentHub Build Script ==="

# Build the frontend
echo "Building frontend..."
cd web
npm install --prefer-offline
npm run build
cd "$SCRIPT_DIR"

# Build the server
echo "Building server..."
go build -o agenthub ./cmd/agenthub

# Build the CLI
echo "Building CLI..."
go build -o agenthub-cli ./cmd/agenthub-cli

echo ""
echo "=== Build Complete ==="
echo ""
echo "To run the server:"
echo "  ./agenthub"
echo ""
echo "Then open http://localhost:9093 in your browser"
