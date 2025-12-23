#!/bin/bash
# Code Coverage Display Script

echo "=================================="
echo "Code Coverage Report"
echo "=================================="

# Colors
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'

print_status() { echo -e "${1}${2}${NC}"; }

# Set test database environment variables (if needed)
export TEST_DB_HOST=${TEST_DB_HOST:-localhost}
export TEST_DB_PORT=${TEST_DB_PORT:-5433}
export TEST_DB_USER=${TEST_DB_USER:-postgres}
export TEST_DB_PASSWORD=${TEST_DB_PASSWORD:-password}
export TEST_DB_NAME=${TEST_DB_NAME:-taskdb_test}

# Generate coverage profile
print_status $BLUE "Generating coverage profile..."
if go test ./... -coverprofile=coverage.out >/dev/null 2>&1; then
    print_status $GREEN "✓ Coverage profile generated"
    echo

    # Display coverage summary
    print_status $BLUE "Coverage Summary:"
    go tool cover -func=coverage.out

    echo
    # Get total coverage percentage
    coverage_line=$(go tool cover -func=coverage.out | grep total)
    coverage_percent=$(echo "$coverage_line" | grep -o '[0-9]\+\.[0-9]\+' | head -1)

    if [ -n "$coverage_percent" ]; then
        coverage_num=$(echo "$coverage_percent" | cut -d. -f1)
        if [ "$coverage_num" -ge 80 ]; then
            color=$GREEN
        elif [ "$coverage_num" -ge 50 ]; then
            color=$YELLOW
        else
            color=$RED
        fi
        echo -n "Total Coverage: "
        print_status $color "${coverage_percent}%"
    fi

    # Generate HTML report
    go tool cover -html=coverage.out -o coverage.html >/dev/null 2>&1
    print_status $GREEN "✓ HTML report generated: coverage.html"
else
    print_status $RED "✗ Failed to generate coverage"
    exit 1
fi