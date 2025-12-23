#!/bin/bash

# Script to generate Swagger documentation using swag CLI

# Install swag CLI if not present
if ! command -v swag &> /dev/null; then
    echo "Installing swag CLI..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

# Add Go bin to PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Generate Swagger docs from the main API file
echo "Generating Swagger documentation..."
swag init -g cmd/api/main.go -o docs/

# Copy the generated swagger.json to the root directory
if [ -f docs/swagger.json ]; then
    echo "Swagger file generated and copied to root: swagger.json"
else
    echo "Error: swagger.json not found in docs/"
fi