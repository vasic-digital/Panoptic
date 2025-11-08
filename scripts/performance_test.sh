#!/bin/bash

set -e

# Performance testing script for Panoptic
# Tests performance characteristics, resource usage, and scalability

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[PERF]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[PERF]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[PERF]${NC} $1"
}

print_error() {
    echo -e "${RED}[PERF]${NC} $1"
}

# Configuration
BUILD_DIR="perf_build"
RESULTS_DIR="perf_results"
TEST_DURATION=60
CONCURRENT_TESTS=5
MEMORY_LIMIT_MB=1024
CPU_LIMIT_PERCENT=80

# Parse command line arguments
QUICK_MODE=false
STRESS_MODE=false
BENCHMARK_MODE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --quick)
            QUICK_MODE=true
            TEST_DURATION=30
            CONCURRENT_TESTS=3
            shift
            ;;
        --stress)
            STRESS_MODE=true
            TEST_DURATION=120
            CONCURRENT_TESTS=10
            shift
            ;;
        --benchmark)
            BENCHMARK_MODE=true
            shift
            ;;
        --duration)
            TEST_DURATION="$2"
            shift 2
            ;;
        --concurrent)
            CONCURRENT_TESTS="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Performance testing options:"
            echo "  --quick          Quick performance test (30s, 3 concurrent)"
            echo "  --stress         Stress test (2min, 10 concurrent)"
            echo "  --benchmark       Benchmark mode (detailed metrics)"
            echo "  --duration N     Test duration in seconds"
            echo "  --concurrent N   Number of concurrent tests"
            echo "  -h, --help      Show this help message"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Performance test functions

run_memory_test() {
    print_status "Running memory usage test..."
    
    # Build application for performance testing
    go build -o $BUILD_DIR/panoptic-perf main.go
    
    # Create test config for memory testing
    cat > memory_test.yaml << EOF
name: "Memory Performance Test"
apps:
  - name: "Memory Test App"
    type: "web"
    url: "https://httpbin.org/delay/5"
actions:
  - name: "navigate"
    type: "navigate"
    value: "https://httpbin.org/delay/5"
  - name: "wait"
    type: "wait"
    wait_time: 2
  - name: "screenshot"
    type: "screenshot"
  - name: "wait"
    type: "wait"
    wait_time: 2
EOF

    # Monitor memory usage during test
    start_time=$(date +%s)
    memory_log="$RESULTS_DIR/memory_usage.log"
    
    # Start memory monitoring in background
    (
        while [ $(($(date +%s) - start_time)) -lt $TEST_DURATION ]; do
            if pgrep -f "panoptic-perf" > /dev/null; then
                pid=$(pgrep -f "panoptic-perf")
                memory_kb=$(ps -p $pid -o rss= 2>/dev/null || echo "0")
                timestamp=$(date +%s.%N)
                echo "$timestamp,$memory_kb" >> "$memory_log"
            fi
            sleep 1
        done
    ) &
    monitor_pid=$!
    
    # Run the test
    timeout $TEST_DURATION $BUILD_DIR/panoptic-perf run memory_test.yaml --output $RESULTS_DIR/memory_test > $RESULTS_DIR/memory_output.log 2>&1 || true
    
    # Stop monitoring
    kill $monitor_pid 2>/dev/null || true
    
    # Analyze memory results
    if [ -f "$memory_log" ]; then
        max_memory=$(awk -F, 'NR==1 || $2>max {max=$2} END {print max}' "$memory_log")
        avg_memory=$(awk -F, '{sum+=$2} END {print int(sum/NR)}' "$memory_log")
        
        print_success "Memory analysis complete:"
        echo "  Peak memory: ${max_memory} KB"
        echo "  Average memory: ${avg_memory} KB"
        
        # Check against limit
        max_memory_mb=$((max_memory / 1024))
        if [ $max_memory_mb -gt $MEMORY_LIMIT_MB ]; then
            print_warning "Memory usage exceeds limit: ${max_memory_mb}MB > ${MEMORY_LIMIT_MB}MB"
        else
            print_success "Memory usage within limits: ${max_memory_mb}MB <= ${MEMORY_LIMIT_MB}MB"
        fi
        
        # Generate memory chart
        python3 - << EOF
import csv
import matplotlib.pyplot as plt

times = []
memories = []

with open('$memory_log', 'r') as f:
    reader = csv.reader(f)
    for row in reader:
        if len(row) == 2:
            times.append(float(row[0]) - float(open('$memory_log').readline().split(',')[0]))
            memories.append(int(row[1]))

if times and memories:
    plt.figure(figsize=(10, 6))
    plt.plot(times, memories)
    plt.xlabel('Time (seconds)')
    plt.ylabel('Memory Usage (KB)')
    plt.title('Panoptic Memory Usage Over Time')
    plt.grid(True)
    plt.savefig('$RESULTS_DIR/memory_chart.png')
    plt.close()
EOF
    else
        print_warning "No memory data collected"
    fi
}

run_cpu_test() {
    print_status "Running CPU usage test..."
    
    # Create CPU-intensive test config
    cat > cpu_test.yaml << EOF
name: "CPU Performance Test"
apps:
  - name: "CPU Test App 1"
    type: "web"
    url: "https://httpbin.org/html"
  - name: "CPU Test App 2"
    type: "web"
    url: "https://httpbin.org/json"
actions:
  - name: "concurrent_navigate"
    type: "navigate"
    value: "https://httpbin.org/delay/1"
  - name: "quick_screenshot"
    type: "screenshot"
  - name: "wait"
    type: "wait"
    wait_time: 1
EOF

    # Monitor CPU usage
    start_time=$(date +%s)
    cpu_log="$RESULTS_DIR/cpu_usage.log"
    
    # Start CPU monitoring
    (
        while [ $(($(date +%s) - start_time)) -lt $TEST_DURATION ]; do
            if pgrep -f "panoptic-perf" > /dev/null; then
                pid=$(pgrep -f "panoptic-perf")
                cpu_percent=$(ps -p $pid -o %cpu= 2>/dev/null | tr -d ' ' || echo "0")
                timestamp=$(date +%s.%N)
                echo "$timestamp,$cpu_percent" >> "$cpu_log"
            fi
            sleep 0.5
        done
    ) &
    monitor_pid=$!
    
    # Run CPU-intensive test
    timeout $TEST_DURATION $BUILD_DIR/panoptic-perf run cpu_test.yaml --output $RESULTS_DIR/cpu_test --verbose > $RESULTS_DIR/cpu_output.log 2>&1 || true
    
    # Stop monitoring
    kill $monitor_pid 2>/dev/null || true
    
    # Analyze CPU results
    if [ -f "$cpu_log" ]; then
        max_cpu=$(awk -F, 'NR==1 || $2>max {max=$2} END {print max}' "$cpu_log")
        avg_cpu=$(awk -F, '{sum+=$2} END {print int(sum/NR)}' "$cpu_log")
        
        print_success "CPU analysis complete:"
        echo "  Peak CPU: ${max_cpu}%"
        echo "  Average CPU: ${avg_cpu}%"
        
        # Check against limit
        if (( $(echo "$max_cpu > $CPU_LIMIT_PERCENT" | bc -l) )); then
            print_warning "CPU usage exceeds limit: ${max_cpu}% > ${CPU_LIMIT_PERCENT}%"
        else
            print_success "CPU usage within limits: ${max_cpu}% <= ${CPU_LIMIT_PERCENT}%"
        fi
    else
        print_warning "No CPU data collected"
    fi
}

run_concurrent_test() {
    print_status "Running concurrent test performance..."
    
    # Create multiple test configs
    for i in $(seq 1 $CONCURRENT_TESTS); do
        cat > concurrent_test_${i}.yaml << EOF
name: "Concurrent Test $i"
apps:
  - name: "Concurrent App $i"
    type: "web"
    url: "https://httpbin.org/delay/2"
actions:
  - name: "navigate_${i}"
    type: "navigate"
    value: "https://httpbin.org/delay/2"
  - name: "screenshot_${i}"
    type: "screenshot"
  - name: "wait_${i}"
    type: "wait"
    wait_time: 1
EOF
    done
    
    # Track concurrent execution
    start_time=$(date +%s.%N)
    concurrent_log="$RESULTS_DIR/concurrent_execution.log"
    
    # Launch all tests in background
    pids=()
    for i in $(seq 1 $CONCURRENT_TESTS); do
        timeout $TEST_DURATION $BUILD_DIR/panoptic-perf run concurrent_test_${i}.yaml --output $RESULTS_DIR/concurrent_${i} > $RESULTS_DIR/concurrent_${i}_output.log 2>&1 &
        pids+=($!)
        echo "$(date +%s.%N),START,$i" >> "$concurrent_log"
    done
    
    # Monitor concurrent execution
    while [ ${#pids[@]} -gt 0 ]; do
        active_pids=()
        for pid in "${pids[@]}"; do
            if kill -0 $pid 2>/dev/null; then
                active_pids+=($pid)
            else
                wait $pid
                echo "$(date +%s.%N),END,$?" >> "$concurrent_log"
            fi
        done
        pids=("${active_pids[@]}")
        sleep 0.1
    done
    
    end_time=$(date +%s.%N)
    duration=$(echo "$end_time - $start_time" | bc)
    
    print_success "Concurrent test complete:"
    echo "  Duration: ${duration} seconds"
    echo "  Concurrent tests: $CONCURRENT_TESTS"
    
    # Analyze results
    successful_tests=0
    for i in $(seq 1 $CONCURRENT_TESTS); do
        if [ -f "$RESULTS_DIR/concurrent_${i}/report.html" ]; then
            ((successful_tests++))
        fi
    done
    
    success_rate=$((successful_tests * 100 / CONCURRENT_TESTS))
    echo "  Successful tests: $successful_tests/$CONCURRENT_TESTS (${success_rate}%)"
    
    if [ $success_rate -lt 90 ]; then
        print_warning "Low success rate in concurrent execution: ${success_rate}%"
    else
        print_success "Good success rate in concurrent execution: ${success_rate}%"
    fi
    
    # Cleanup test files
    rm -f concurrent_test_*.yaml
}

run_load_test() {
    print_status "Running load test..."
    
    # Create load test scenario
    cat > load_test.yaml << EOF
name: "Load Test Scenario"
apps:
  - name: "Load Test App"
    type: "web"
    url: "https://httpbin.org/html"
    timeout: 10
actions:
  - name: "navigate_load"
    type: "navigate"
    value: "https://httpbin.org/html"
  - name: "screenshot_load"
    type: "screenshot"
  - name: "wait_load"
    type: "wait"
    wait_time: 1
  - name: "navigate_json"
    type: "navigate"
    value: "https://httpbin.org/json"
  - name: "wait_json"
    type: "wait"
    wait_time: 1
EOF

    # Run load test with timing
    load_log="$RESULTS_DIR/load_performance.log"
    
    for i in $(seq 1 10); do
        start_time=$(date +%s.%N)
        timeout $TEST_DURATION $BUILD_DIR/panoptic-perf run load_test.yaml --output $RESULTS_DIR/load_test_$i > /dev/null 2>&1
        end_time=$(date +%s.%N)
        duration=$(echo "$end_time - $start_time" | bc)
        echo "$i,$duration" >> "$load_log"
        echo "Load test iteration $i completed in ${duration}s"
    done
    
    # Analyze load test results
    if [ -f "$load_log" ]; then
        avg_duration=$(awk -F, '{sum+=$2} END {print sum/NR}' "$load_log")
        min_duration=$(awk -F, 'NR==1 || $2<min {min=$2} END {print min}' "$load_log")
        max_duration=$(awk -F, 'NR==1 || $2>max {max=$2} END {print max}' "$load_log")
        
        print_success "Load test analysis:"
        echo "  Average duration: ${avg_duration}s"
        echo "  Min duration: ${min_duration}s"
        echo "  Max duration: ${max_duration}s"
        
        # Check performance consistency
        variance=$(echo "$max_duration - $min_duration" | bc)
        if (( $(echo "$variance > 5.0" | bc -l) )); then
            print_warning "High variance in load test: ${variance}s"
        else
            print_success "Consistent performance in load test: variance ${variance}s"
        fi
    fi
}

run_benchmark_test() {
    if [ "$BENCHMARK_MODE" != true ]; then
        return
    fi
    
    print_status "Running detailed benchmark tests..."
    
    # Benchmark individual operations
    cat > benchmark.yaml << EOF
name: "Benchmark Test"
apps:
  - name: "Benchmark App"
    type: "web"
    url: "https://httpbin.org/html"
actions:
  - name: "benchmark_navigate"
    type: "navigate"
    value: "https://httpbin.org/html"
  - name: "benchmark_screenshot"
    type: "screenshot"
  - name: "benchmark_wait"
    type: "wait"
    wait_time: 1
EOF

    # Run benchmark with detailed timing
    benchmark_log="$RESULTS_DIR/benchmark.log"
    
    echo "timestamp,operation,duration_ms" > "$benchmark_log"
    
    # Run multiple iterations for each operation
    operations=("navigate" "screenshot" "wait")
    
    for op in "${operations[@]}"; do
        for i in $(seq 1 20); do
            start_time=$(date +%s.%N%3 | sed 's/...$//')
            timeout 30 $BUILD_DIR/panoptic-perf run benchmark.yaml --output $RESULTS_DIR/benchmark_$op > /dev/null 2>&1
            end_time=$(date +%s.%N%3 | sed 's/...$//')
            duration_ms=$(echo "($end_time - $start_time) * 1000" | bc)
            echo "$(date +%s),$op,$duration_ms" >> "$benchmark_log"
        done
        echo "Benchmark for $op completed"
    done
    
    # Generate benchmark report
    python3 - << EOF
import csv
import statistics

operations = {}
with open('$benchmark_log', 'r') as f:
    reader = csv.DictReader(f)
    for row in reader:
        op = row['operation']
        duration = float(row['duration_ms'])
        if op not in operations:
            operations[op] = []
        operations[op].append(duration)

print("Benchmark Results:")
for op, durations in operations.items():
    print(f"\\n{op}:")
    print(f"  Average: {statistics.mean(durations):.2f}ms")
    print(f"  Median:  {statistics.median(durations):.2f}ms")
    print(f"  Min:     {min(durations):.2f}ms")
    print(f"  Max:     {max(durations):.2f}ms")
    print(f"  StdDev:  {statistics.stdev(durations):.2f}ms")
EOF
}

generate_performance_report() {
    print_status "Generating performance report..."
    
    cat > "$RESULTS_DIR/performance_report.md" << EOF
# Panoptic Performance Test Report

## Test Configuration
- Test Duration: ${TEST_DURATION}s
- Concurrent Tests: $CONCURRENT_TESTS
- Memory Limit: ${MEMORY_LIMIT_MB}MB
- CPU Limit: ${CPU_LIMIT_PERCENT}%

## Test Results

### Memory Usage
- Peak Memory: $(cat $RESULTS_DIR/memory_usage.log 2>/dev/null | awk -F, 'NR==1 || $2>max {max=$2} END {print max/1024}' | cut -d. -f1 || echo "N/A")MB
- Average Memory: $(cat $RESULTS_DIR/memory_usage.log 2>/dev/null | awk -F, '{sum+=$2} END {print int(sum/NR/1024)}' || echo "N/A")MB

### CPU Usage
- Peak CPU: $(cat $RESULTS_DIR/cpu_usage.log 2>/dev/null | awk -F, 'NR==1 || $2>max {max=$2} END {print max}' || echo "N/A")%
- Average CPU: $(cat $RESULTS_DIR/cpu_usage.log 2>/dev/null | awk -F, '{sum+=$2} END {print int(sum/NR)}' || echo "N/A")%

### Concurrent Performance
- Concurrent Tests: $CONCURRENT_TESTS
- Duration: $(cat $RESULTS_DIR/concurrent_execution.log 2>/dev/null | awk -F, 'NR==1 {start=$1} END {print $1-start}' || echo "N/A")s

### Load Performance
- Average Duration: $(cat $RESULTS_DIR/load_performance.log 2>/dev/null | awk -F, '{sum+=$2} END {print sum/NR}' || echo "N/A")s

## Performance Charts
- Memory usage chart: memory_chart.png
- CPU usage chart: cpu_chart.png

## Recommendations
EOF

    # Add recommendations based on results
    if [ -f "$RESULTS_DIR/memory_usage.log" ]; then
        max_memory_mb=$(awk -F, 'NR==1 || $2>max {max=$2} END {print max/1024}' "$RESULTS_DIR/memory_usage.log" | cut -d. -f1)
        if [ "$max_memory_mb" -gt $((MEMORY_LIMIT_MB * 80 / 100)) ]; then
            echo "- Consider memory optimization - usage approaching limits" >> "$RESULTS_DIR/performance_report.md"
        else
            echo "- Memory usage is within acceptable limits" >> "$RESULTS_DIR/performance_report.md"
        fi
    fi
    
    if [ -f "$RESULTS_DIR/cpu_usage.log" ]; then
        max_cpu=$(awk -F, 'NR==1 || $2>max {max=$2} END {print max}' "$RESULTS_DIR/cpu_usage.log")
        if (( $(echo "$max_cpu > $((CPU_LIMIT_PERCENT * 80 / 100))" | bc -l) )); then
            echo "- Consider CPU optimization - usage approaching limits" >> "$RESULTS_DIR/performance_report.md"
        else
            echo "- CPU usage is within acceptable limits" >> "$RESULTS_DIR/performance_report.md"
        fi
    fi
    
    print_success "Performance report generated: $RESULTS_DIR/performance_report.md"
}

cleanup() {
    print_status "Cleaning up test artifacts..."
    rm -f *_test.yaml
    rm -rf $BUILD_DIR
}

# Main execution
main() {
    print_status "Starting Panoptic performance testing"
    print_status "=================================="
    
    # Setup
    mkdir -p $RESULTS_DIR
    mkdir -p $BUILD_DIR
    
    # Build performance test binary
    print_status "Building performance test binary..."
    go build -o $BUILD_DIR/panoptic-perf main.go
    
    if [ "$STRESS_MODE" = true ]; then
        print_status "Running stress tests..."
        run_memory_test
        run_cpu_test
        run_concurrent_test
        run_load_test
    elif [ "$QUICK_MODE" = true ]; then
        print_status "Running quick performance tests..."
        run_memory_test
        run_cpu_test
    else
        print_status "Running standard performance tests..."
        run_memory_test
        run_cpu_test
        run_concurrent_test
        run_load_test
    fi
    
    run_benchmark_test
    generate_performance_report
    cleanup
    
    print_status "Performance testing completed"
    print_status "=================================="
    print_status "Results available in: $RESULTS_DIR"
}

# Check dependencies
if ! command -v bc &> /dev/null; then
    print_error "bc calculator is required for performance testing"
    exit 1
fi

if ! command -v python3 &> /dev/null && [ "$BENCHMARK_MODE" = true ]; then
    print_error "Python3 is required for benchmark mode"
    exit 1
fi

# Run main function
main "$@"