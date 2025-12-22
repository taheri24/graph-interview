#!/bin/bash
# Compact test script for Task Management API

echo "=================================="
echo "Task Management API Test Suite"
echo "=================================="

# Colors
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'

print_status() { echo -e "${1}${2}${NC}"; }

# Set test database environment variables
export TEST_DB_HOST=${TEST_DB_HOST:-localhost}
export TEST_DB_PORT=${TEST_DB_PORT:-5433}
export TEST_DB_USER=${TEST_DB_USER:-postgres}
export TEST_DB_PASSWORD=${TEST_DB_PASSWORD:-password}
export TEST_DB_NAME=${TEST_DB_NAME:-taskdb_test}

print_status $BLUE "Test DB: $TEST_DB_HOST:$TEST_DB_PORT/$TEST_DB_NAME"

# Track results
total=0; passed=0; failed=0

run_check() {
    local name=$1; local cmd=$2; total=$((total+1))
    print_status $BLUE "Running $name..."
    if eval "$cmd"; then
        print_status $GREEN "✓ $name passed"
        passed=$((passed+1))
    else
        print_status $RED "✗ $name failed"
        failed=$((failed+1))
    fi
    echo
}

# Linting
if ./lint.sh >/dev/null 2>&1; then
    print_status $GREEN "✓ Linting passed"
    passed=$((passed+1))
else
    print_status $RED "✗ Linting failed"
    ./lint.sh
    failed=$((failed+1))
    exit 1
fi
total=$((total+1))
echo

# Tests
test_cmd="go test ./..."
[ "$1" = "--short" ] && test_cmd="$test_cmd -short"

run_check "All Tests" "$test_cmd -v"

# Coverage
print_status $BLUE "Generating Coverage..."
if go test ./... -coverprofile=coverage.out >/dev/null 2>&1; then
    print_status $GREEN "✓ Coverage generated"
    echo

    # Get and color coverage percentage
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
        echo -n "Coverage: "
        print_status $color "${coverage_percent}%"
    else
        print_status $BLUE "Coverage Summary:"
        echo "$coverage_line"
    fi

    go tool cover -html=coverage.out -o coverage.html >/dev/null 2>&1
    print_status $GREEN "✓ HTML report: coverage.html"
    passed=$((passed+1))
else
    print_status $RED "✗ Coverage failed"
    failed=$((failed+1))
fi
total=$((total+1))

echo
echo "=================================="
print_status $BLUE "Summary: $passed/$total passed"
[ $failed -eq 0 ] && print_status $GREEN "🎉 All tests passed!" || print_status $RED "❌ $failed tests failed"
exit $failed