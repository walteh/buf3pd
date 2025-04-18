#!/bin/bash

# Exit on error
set -e

# Create coverage directory if it doesn't exist
mkdir -p coverage

# Run tests with coverage
go test -race -coverprofile=coverage/coverage.out -covermode=atomic ./...

# Generate HTML coverage report
go tool cover -html=coverage/coverage.out -o coverage/coverage.html

# Print coverage percentage
echo "Coverage report:"
go tool cover -func=coverage/coverage.out | grep total

echo "HTML coverage report generated at coverage/coverage.html"
