#!/bin/bash
# Load Testing Script for Task Management API with pprof profiling
#
# Usage:
#   1. Start the application with pprof enabled:
#      PPROF_ENABLED=true docker-compose up -d
#      # or
#      PPROF_ENABLED=true ./dev.sh
#
#   2. Run the load test:
#      ./test-load.sh
#
#   3. View results in generated profile files:
#      cpu_report.txt, mem_report.txt, goroutine_report.txt
#
# Configuration:
#   - API_URL: API endpoint (default: http://localhost:8080)
#   - CONCURRENT_REQUESTS: Number of concurrent workers (default: 5)
#   - TOTAL_REQUESTS: Total requests to make (default: 200)
#   - REQUEST_DELAY: Delay between requests in seconds (default: 0.1)

echo "=================================="
echo "Task Management API Load Test"
echo "=================================="

# Colors
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'

print_status() { echo -e "${1}${2}${NC}"; }

# Configuration (can be overridden with environment variables)
API_URL="${API_URL:-http://localhost:8080}"
PPROF_PORT="${PPROF_PORT:-6060}"
CONCURRENT_REQUESTS="${CONCURRENT_REQUESTS:-5}"
TOTAL_REQUESTS="${TOTAL_REQUESTS:-200}"
REQUEST_DELAY="${REQUEST_DELAY:-0.1}"

# Set environment variables for pprof
export PPROF_ENABLED=true

print_status $BLUE "Configuration:"
echo "API URL: $API_URL"
echo "Concurrent Requests: $CONCURRENT_REQUESTS"
echo "Total Requests: $TOTAL_REQUESTS"
echo "Request Delay: ${REQUEST_DELAY}s"
echo "Pprof Port: $PPROF_PORT"
echo

# Function to check if service is healthy
check_health() {
    local max_attempts=30
    local attempt=1

    print_status $BLUE "Checking API health..."

    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "$API_URL/health" >/dev/null 2>&1; then
            print_status $GREEN "✓ API is healthy"
            return 0
        fi

        print_status $YELLOW "Waiting for API to be ready (attempt $attempt/$max_attempts)..."
        sleep 2
        ((attempt++))
    done

    print_status $RED "✗ API failed to become healthy"
    return 1
}

# Function to create test tasks
create_test_tasks() {
    local num_tasks=$1
    local task_ids=()

    print_status $BLUE "Creating $num_tasks test tasks..."

    for i in $(seq 1 $num_tasks); do
        response=$(curl -s -X POST "$API_URL/tasks" \
            -H "Content-Type: application/json" \
            -d "{
                \"title\": \"Load Test Task $i\",
                \"description\": \"Description for load test task $i\",
                \"status\": \"pending\",
                \"assignee\": \"loadtest@example.com\"
            }")

        task_id=$(echo "$response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
        if [ -n "$task_id" ]; then
            task_ids+=("$task_id")
        fi
    done

    echo "${task_ids[@]}"
}

# Function to run load test
run_load_test() {
    local task_ids=($1)
    local num_tasks=${#task_ids[@]}

    print_status $BLUE "Starting load test with $CONCURRENT_REQUESTS concurrent requests..."

    # Create temporary script for parallel execution
    cat > /tmp/load_test_worker.sh << 'EOF'
#!/bin/bash
API_URL=$1
TASK_IDS=$2
WORKER_ID=$3
TOTAL_REQUESTS=$4
REQUEST_DELAY=$5

for i in $(seq 1 $TOTAL_REQUESTS); do
    # Random operation selection
    operation=$((RANDOM % 4))

    case $operation in
        0) # GET all tasks
            curl -s "$API_URL/tasks?page=1&limit=10" >/dev/null
            ;;
        1) # GET single task
            if [ ${#TASK_IDS[@]} -gt 0 ]; then
                random_index=$((RANDOM % ${#TASK_IDS[@]}))
                task_id=${TASK_IDS[$random_index]}
                curl -s "$API_URL/tasks/$task_id" >/dev/null
            fi
            ;;
        2) # POST create task
            curl -s -X POST "$API_URL/tasks" \
                -H "Content-Type: application/json" \
                -d "{\"title\":\"Worker $WORKER_ID Task $i\",\"status\":\"pending\"}" >/dev/null
            ;;
        3) # PUT update task
            if [ ${#TASK_IDS[@]} -gt 0 ]; then
                random_index=$((RANDOM % ${#TASK_IDS[@]}))
                task_id=${TASK_IDS[$random_index]}
                curl -s -X PUT "$API_URL/tasks/$task_id" \
                    -H "Content-Type: application/json" \
                    -d "{\"status\":\"in_progress\"}" >/dev/null
            fi
            ;;
    esac

    # Small delay between requests
    sleep $REQUEST_DELAY
done

echo "Worker $WORKER_ID completed"
EOF

    chmod +x /tmp/load_test_worker.sh

    # Start background workers
    for worker_id in $(seq 1 $CONCURRENT_REQUESTS); do
        /tmp/load_test_worker.sh "$API_URL" "${task_ids[*]}" "$worker_id" "$((TOTAL_REQUESTS / CONCURRENT_REQUESTS))" "$REQUEST_DELAY" &
        pids[$worker_id]=$!
    done

    # Wait for all workers to complete
    for pid in "${pids[@]}"; do
        wait $pid
    done

    print_status $GREEN "✓ Load test completed"
}

# Function to collect pprof data
collect_pprof() {
    print_status $BLUE "Collecting pprof profiles..."

    # CPU profile
    print_status $YELLOW "Collecting CPU profile (10 seconds)..."
    curl -s "$API_URL/debug/pprof/profile?seconds=10" -o cpu.prof

    # Memory profile
    print_status $YELLOW "Collecting memory profile..."
    curl -s "$API_URL/debug/pprof/heap" -o mem.prof

    # Goroutine profile
    print_status $YELLOW "Collecting goroutine profile..."
    curl -s "$API_URL/debug/pprof/goroutine" -o goroutine.prof

    print_status $GREEN "✓ Pprof profiles collected"
}

# Function to generate pprof reports
generate_reports() {
    print_status $BLUE "Generating pprof reports..."

    if command -v go >/dev/null 2>&1 && [ -f cpu.prof ]; then
        # CPU report
        print_status $YELLOW "Generating CPU profile report..."
        if go tool pprof -text -output=cpu_report.txt cpu.prof >/dev/null 2>&1; then
            print_status $GREEN "✓ CPU report generated"
        else
            print_status $YELLOW "! Failed to generate CPU report"
        fi

        # Memory report
        if [ -f mem.prof ]; then
            print_status $YELLOW "Generating memory profile report..."
            if go tool pprof -text -output=mem_report.txt mem.prof >/dev/null 2>&1; then
                print_status $GREEN "✓ Memory report generated"
            else
                print_status $YELLOW "! Failed to generate memory report"
            fi
        fi

        # Goroutine report
        if [ -f goroutine.prof ]; then
            print_status $YELLOW "Generating goroutine profile report..."
            if go tool pprof -text -output=goroutine_report.txt goroutine.prof >/dev/null 2>&1; then
                print_status $GREEN "✓ Goroutine report generated"
            else
                print_status $YELLOW "! Failed to generate goroutine report"
            fi
        fi

        print_status $GREEN "✓ Profile reports generated (if Go toolchain available)"
    else
        print_status $YELLOW "! Go toolchain not available or no profiles collected"
        print_status $YELLOW "Raw profiles saved: cpu.prof, mem.prof, goroutine.prof"
    fi
}

# Function to show summary
show_summary() {
    echo
    echo "=================================="
    print_status $BLUE "Load Test Summary"
    echo "=================================="
    echo "Total Requests: $TOTAL_REQUESTS"
    echo "Concurrent Workers: $CONCURRENT_REQUESTS"
    echo "Average RPS: $(echo "scale=2; $TOTAL_REQUESTS / ($CONCURRENT_REQUESTS * $REQUEST_DELAY * $TOTAL_REQUESTS / $CONCURRENT_REQUESTS)" | bc 2>/dev/null || echo "N/A")"

    if [ -f cpu_report.txt ]; then
        echo
        print_status $BLUE "Top CPU Consumers:"
        head -10 cpu_report.txt
    fi

    if [ -f mem_report.txt ]; then
        echo
        print_status $BLUE "Top Memory Consumers:"
        head -10 mem_report.txt
    fi

    print_status $GREEN "Load test completed successfully!"
}

# Main execution
main() {
    # Check if API is running
    if ! check_health; then
        print_status $RED "API is not available. Please start the application first with pprof enabled:"
        echo "  # For Docker Compose:"
        echo "  PPROF_ENABLED=true docker-compose up -d"
        echo "  # Wait for services to be ready, then run:"
        echo "  ./test-load.sh"
        echo ""
        echo "  # For local development:"
        echo "  PPROF_ENABLED=true ./dev.sh"
        echo "  # In another terminal:"
        echo "  ./test-load.sh"
        exit 1
    fi

    # Create test data
    task_ids=$(create_test_tasks 20)

    # Start pprof collection in background
    collect_pprof &
    pprof_pid=$!

    # Run load test
    run_load_test "$task_ids"

    # Wait for pprof collection to complete
    wait $pprof_pid

    # Generate reports
    generate_reports

    # Show summary
    show_summary

    # Cleanup
    rm -f /tmp/load_test_worker.sh
}

# Handle cleanup on exit
trap 'echo -e "\nCleaning up..."; rm -f /tmp/load_test_worker.sh; exit' INT TERM

# Run main function
main "$@"