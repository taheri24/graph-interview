#!/bin/bash
# Development script for Task API

# Add Go binaries to PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Set default environment variables for development
export DB_HOST=${DB_HOST:-localhost}
export DB_USER=${DB_USER:-postgres}
export DB_PASSWORD=${DB_PASSWORD:-""}
export DB_NAME=${DB_NAME:-taskdb}
export SERVER_PORT=${SERVER_PORT:-8080}

echo "Starting development server with Air..."
echo "Database: $DB_HOST:$DB_PORT/$DB_NAME"
echo "Server: http://localhost:$SERVER_PORT"
echo "Press Ctrl+C to stop"

# Run Air
air