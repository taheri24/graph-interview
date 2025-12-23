#!/bin/bash
# Seeder script for Task Management API

# Add Go binaries to PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Set default environment variables for seeding
export DB_HOST=${DB_HOST:-localhost}
export DB_USER=${DB_USER:-postgres}
export DB_PASSWORD=${DB_PASSWORD:-password}
export DB_NAME=${DB_NAME:-taskdb}
export SERVER_PORT=${SERVER_PORT:-8080}

echo "Starting database seeder..."
echo "Database: $DB_HOST:$DB_PORT/$DB_NAME"

# Run the seeder
go run cmd/seeder/main.go

if [ $? -eq 0 ]; then
    echo "Seeding completed successfully!"
else
    echo "Seeding failed!"
    exit 1
fi