#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Configuration
COVERAGE_FILE="coverage.out"
COVERAGE_HTML="coverage.html"
MIN_COVERAGE=75.0
PROJECT_NAME="Panoptic"

print_status "Starting code coverage analysis for $PROJECT_NAME"
print_status "=============================================="

# Clean up previous coverage
rm -f $COVERAGE_FILE $COVERAGE_HTML

# Run tests with coverage
print_status "Running tests with coverage profiling..."
if go test -coverprofile=$COVERAGE_FILE -covermode=atomic ./internal/... ./cmd/...; then
    print_success "Tests completed successfully"
else
    print_error "Tests failed - coverage analysis aborted"
    exit 1
fi

# Check if coverage file was created
if [ ! -f "$COVERAGE_FILE" ]; then
    print_error "Coverage file not generated"
    exit 1
fi

# Calculate total coverage
print_status "Calculating coverage metrics..."
total_coverage=$(go tool cover -func=$COVERAGE_FILE | tail -1 | awk '{print $3}')
print_status "Total coverage: $total_coverage"

# Check minimum coverage threshold
coverage_float=$(echo $total_coverage | sed 's/%//')
if (( $(echo "$coverage_float < $MIN_COVERAGE" | bc -l) )); then
    print_error "Coverage $total_coverage is below minimum threshold $MIN_COVERAGE%"
    echo
    print_status "Coverage details by package:"
    go tool cover -func=$COVERAGE_FILE | grep -v "total:"
    exit 1
else
    print_success "Coverage $total_coverage meets minimum threshold $MIN_COVERAGE%"
fi

# Generate HTML coverage report
print_status "Generating HTML coverage report..."
go tool cover -html=$COVERAGE_FILE -o $COVERAGE_HTML
print_success "HTML report generated: $COVERAGE_HTML"

# Coverage by package analysis
print_status "Coverage by package:"
echo
go tool cover -func=$COVERAGE_FILE | grep -v "total:" | while read line; do
    percentage=$(echo $line | awk '{print $3}')
    file=$(echo $line | awk '{print $1}')
    
    # Extract package name
    package=$(dirname $file | sed 's|^./||')
    
    # Color code based on coverage
    if [[ "$percentage" == *"100.0%"* ]]; then
        echo -e "${GREEN}✓ $package: $percentage${NC}"
    elif [[ $(echo "$percentage" | sed 's/%//') -ge 80 ]]; then
        echo -e "${GREEN}✓ $package: $percentage${NC}"
    elif [[ $(echo "$percentage" | sed 's/%//') -ge 60 ]]; then
        echo -e "${YELLOW}⚠ $package: $percentage${NC}"
    else
        echo -e "${RED}✗ $package: $percentage${NC}"
    fi
done

# Detailed coverage breakdown
echo
print_status "Detailed coverage breakdown:"
echo

# Find files with low coverage
print_status "Files needing attention (coverage < 75%):"
low_coverage_files=$(go tool cover -func=$COVERAGE_FILE | awk -v min=75.0 '$3 != "" && substr($3, 1, length($3)-1) < min && $1 != "total:" {print $0}')

if [ -n "$low_coverage_files" ]; then
    echo "$low_coverage_files" | while read line; do
        coverage=$(echo $line | awk '{print $3}')
        file=$(echo $line | awk '{print $1}')
        echo -e "${RED}  $file: $coverage${NC}"
    done
else
    print_success "All files meet coverage requirements!"
fi

echo

# Find files with excellent coverage
print_status "Files with excellent coverage (> 90%):"
high_coverage_files=$(go tool cover -func=$COVERAGE_FILE | awk -v min=90.0 '$3 != "" && substr($3, 1, length($3)-1) >= min && $1 != "total:" {print $0}')

if [ -n "$high_coverage_files" ]; then
    echo "$high_coverage_files" | while read line; do
        coverage=$(echo $line | awk '{print $3}')
        file=$(echo $line | awk '{print $1}')
        echo -e "${GREEN}  $file: $coverage${NC}"
    done
else
    print_warning "No files with >90% coverage found"
fi

# Coverage trend analysis (if previous coverage file exists)
PREVIOUS_COVERAGE_FILE="previous_coverage.out"
if [ -f "$PREVIOUS_COVERAGE_FILE" ]; then
    print_status "Coverage trend analysis:"
    previous_coverage=$(go tool cover -func=$PREVIOUS_COVERAGE_FILE | tail -1 | awk '{print $3}')
    
    # Remove % and convert to number for comparison
    prev_float=$(echo $previous_coverage | sed 's/%//')
    curr_float=$(echo $total_coverage | sed 's/%//')
    
    if (( $(echo "$curr_float > $prev_float" | bc -l) )); then
        improvement=$(echo "$curr_float - $prev_float" | bc)
        print_success "Coverage improved by ${improvement}% (from $previous_coverage to $total_coverage)"
    elif (( $(echo "$curr_float < $prev_float" | bc -l) )); then
        regression=$(echo "$prev_float - $curr_float" | bc)
        print_warning "Coverage decreased by ${regression}% (from $previous_coverage to $total_coverage)"
    else
        print_status "Coverage unchanged at $total_coverage"
    fi
    
    # Save current coverage for next comparison
    cp $COVERAGE_FILE $PREVIOUS_COVERAGE_FILE
else
    print_status "No previous coverage file found - saving current as baseline"
    cp $COVERAGE_FILE $PREVIOUS_COVERAGE_FILE
fi

# Recommendations
echo
print_status "Coverage recommendations:"
echo

if (( $(echo "$coverage_float < 60" | bc -l) )); then
    print_error "Critical: Coverage is very low. Consider adding tests for core functionality."
elif (( $(echo "$coverage_float < 80" | bc -l) )); then
    print_warning "Moderate: Coverage could be improved. Focus on untested code paths."
else
    print_success "Good: Coverage is solid. Consider edge case testing."
fi

# Check for untested packages
untested_packages=$(find ./internal ./cmd -name "*.go" -exec dirname {} \; | sort -u | while read pkg; do
    if [ -f "$pkg" ]; then
        # Check if package has test files
        test_files=$(find $pkg -name "*_test.go" 2>/dev/null | wc -l)
        if [ "$test_files" -eq 0 ]; then
            echo $pkg
        fi
    fi
done)

if [ -n "$untested_packages" ]; then
    print_warning "Packages without tests:"
    echo "$untested_packages" | while read pkg; do
        echo -e "${YELLOW}  $pkg${NC}"
    done
else
    print_success "All packages have test files"
fi

echo
print_status "Coverage analysis completed!"
print_status "HTML report: $COVERAGE_HTML"
print_status "Coverage data: $COVERAGE_FILE"
print_status "=============================================="