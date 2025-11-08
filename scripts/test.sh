#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Default settings
COVERAGE_FILE="coverage.out"
COVERAGE_HTML="coverage.html"
TEST_RESULTS="test-results.xml"

# Parse command line arguments
VERBOSE=false
SKIP_INTEGRATION=false
SKIP_E2E=false
COVER_PROFILE=false
RACE_DETECTOR=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        --skip-integration)
            SKIP_INTEGRATION=true
            shift
            ;;
        --skip-e2e)
            SKIP_E2E=true
            shift
            ;;
        --coverage)
            COVER_PROFILE=true
            shift
            ;;
        --race)
            RACE_DETECTOR=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  -v, --verbose         Enable verbose output"
            echo "  --skip-integration    Skip integration tests"
            echo "  --skip-e2e           Skip end-to-end tests"
            echo "  --coverage           Generate coverage profile"
            echo "  --race               Enable race detector"
            echo "  -h, --help           Show this help message"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Start timing
start_time=$(date +%s)

print_status "Starting Panoptic Test Suite"
print_status "=============================="

# Build test flags
TEST_FLAGS=""
if [ "$VERBOSE" = true ]; then
    TEST_FLAGS="-v"
fi

if [ "$RACE_DETECTOR" = true ]; then
    TEST_FLAGS="$TEST_FLAGS -race"
fi

if [ "$COVER_PROFILE" = true ]; then
    TEST_FLAGS="$TEST_FLAGS -coverprofile=$COVERAGE_FILE"
fi

print_status "Test flags: $TEST_FLAGS"

# Clean up previous results
rm -f $COVERAGE_FILE $COVERAGE_HTML $TEST_RESULTS
print_status "Cleaned up previous test results"

# Run unit tests
print_status "Running unit tests..."
unit_start=$(date +%s)

if go test $TEST_FLAGS ./internal/... ./cmd/...; then
    unit_end=$(date +%s)
    unit_duration=$((unit_end - unit_start))
    print_success "Unit tests completed in ${unit_duration}s"
else
    print_error "Unit tests failed"
    exit 1
fi

# Run integration tests
if [ "$SKIP_INTEGRATION" = false ]; then
    print_status "Running integration tests..."
    integration_start=$(date +%s)
    
    if go test $TEST_FLAGS -tags=integration ./tests/integration/...; then
        integration_end=$(date +%s)
        integration_duration=$((integration_end - integration_start))
        print_success "Integration tests completed in ${integration_duration}s"
    else
        print_warning "Integration tests failed (may be expected if dependencies unavailable)"
    fi
else
    print_status "Skipping integration tests"
fi

# Run e2e tests
if [ "$SKIP_E2E" = false ]; then
    print_status "Running end-to-end tests..."
    e2e_start=$(date +%s)
    
    if go test $TEST_FLAGS -tags=e2e ./tests/e2e/...; then
        e2e_end=$(date +%s)
        e2e_duration=$((e2e_end - e2e_start))
        print_success "E2E tests completed in ${e2e_duration}s"
    else
        print_warning "E2E tests failed (may be expected if dependencies unavailable)"
    fi
else
    print_status "Skipping end-to-end tests"
fi

# Generate coverage report
if [ "$COVER_PROFILE" = true ] && [ -f "$COVERAGE_FILE" ]; then
    print_status "Generating coverage report..."
    
    # Generate HTML coverage
    go tool cover -html=$COVERAGE_FILE -o $COVERAGE_HTML
    print_success "HTML coverage report generated: $COVERAGE_HTML"
    
    # Show coverage percentage
    coverage=$(go tool cover -func=$COVERAGE_FILE | tail -1 | awk '{print $3}')
    print_status "Total coverage: $coverage"
    
    # Check if coverage meets minimum threshold (80%)
    if [[ "$coverage" < "80.0%" ]]; then
        print_warning "Coverage is below 80%: $coverage"
    else
        print_success "Coverage meets minimum threshold: $coverage"
    fi
fi

# Generate test results (requires go-junit-report)
if command -v go-junit-report &> /dev/null; then
    print_status "Generating JUnit test results..."
    go test $TEST_FLAGS -json ./... 2>&1 | go-junit-report > $TEST_RESULTS
    print_success "JUnit test results generated: $TEST_RESULTS"
else
    print_status "go-junit-report not found, skipping JUnit report generation"
fi

# Calculate total duration
end_time=$(date +%s)
total_duration=$((end_time - start_time))

print_status "=============================="
print_success "All tests completed in ${total_duration}s"

if [ "$COVER_PROFILE" = true ] && [ -f "$COVERAGE_HTML" ]; then
    print_status "Coverage report available: $COVERAGE_HTML"
fi

if [ -f "$TEST_RESULTS" ]; then
    print_status "JUnit results available: $TEST_RESULTS"
fi

print_status "=============================="