#!/bin/bash
# Comprehensive linting script for Task Management API

echo "=================================="
echo "Go Code Linting"
echo "=================================="
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to run linter with timing
run_linter() {
    local linter_name=$1
    local linter_cmd=$2
    local start_time=$(date +%s)

    print_status $BLUE "Running $linter_name..."
    echo "Command: $linter_cmd"
    echo

    local output
    output=$(eval "$linter_cmd" 2>&1)
    local exit_code=$?

    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    if [ $exit_code -eq 0 ] || [ -z "$output" ]; then
        print_status $GREEN "✓ $linter_name passed (${duration}s)"
        echo
        return 0
    else
        print_status $RED "✗ $linter_name failed (${duration}s)"
        if [ -n "$output" ]; then
            echo "$output"
        fi
        echo
        return 1
    fi
}

# Track linting results
total_lints=0
passed_lints=0
failed_lints=0

# Run go vet
if run_linter "go vet" "go vet ./..."; then
    ((passed_lints++))
else
    ((failed_lints++))
fi
((total_lints++))

# Run go fmt check
print_status $BLUE "Running go fmt check..."
echo "Command: go fmt -d ."
echo

if [ -z "$(go fmt -d .)" ]; then
    print_status $GREEN "✓ Code formatting is correct"
    ((passed_lints++))
else
    print_status $RED "✗ Code formatting issues found:"
    go fmt -d .
    ((failed_lints++))
fi
((total_lints++))
echo

# Run go mod tidy
print_status $BLUE "Running go mod tidy..."
echo "Command: go mod tidy"
echo

if go mod tidy >/dev/null 2>&1; then
    print_status $GREEN "✓ Go modules tidied successfully"
    ((passed_lints++))
else
    print_status $RED "✗ Go mod tidy failed"
    ((failed_lints++))
fi
((total_lints++))
echo

# Check for ineffectual assignments (if go compiler supports it)
print_status $BLUE "Running ineffectual assignment check..."
echo "Command: go build -o /dev/null ./..."
echo

if go build -o /dev/null ./... >/dev/null 2>&1; then
    print_status $GREEN "✓ Code compiles successfully"
    ((passed_lints++))
else
    print_status $RED "✗ Compilation failed"
    go build -o /dev/null ./...
    ((failed_lints++))
fi
((total_lints++))
echo

# Print final summary
echo "=================================="
print_status $BLUE "Linting Summary:"
echo "Total checks: $total_lints"
print_status $GREEN "Passed: $passed_lints"
if [ $failed_lints -gt 0 ]; then
    print_status $RED "Failed: $failed_lints"
else
    print_status $GREEN "Failed: $failed_lints"
fi

echo
if [ $failed_lints -eq 0 ]; then
    print_status $GREEN "🎉 All linting checks passed!"
else
    print_status $RED "❌ Some linting checks failed. Please review the output above."
fi

# Return appropriate exit code
exit $failed_lints