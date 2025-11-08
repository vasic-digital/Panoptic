#!/bin/bash

# Panoptic Test Automation Dashboard
# Provides real-time monitoring and reporting of test execution

set -e

# Colors and formatting
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m'

BOLD='\033[1m'
DIM='\033[2m'
UNDERLINE='\033[4m'
BLINK='\033[5m'
REVERSE='\033[7m'

# Dashboard configuration
REFRESH_INTERVAL=5
DASHBOARD_LOG="/tmp/panoptic_dashboard.log"
METRICS_FILE="/tmp/panoptic_metrics.json"
ALERT_THRESHOLDS=100
HISTORY_FILE="/tmp/panoptic_history.csv"

# Create logs directory
mkdir -p "$(dirname "$DASHBOARD_LOG")"
mkdir -p "$(dirname "$METRICS_FILE")"

# Dashboard functions

show_header() {
    clear
    echo -e "${BLUE}${BOLD}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}${BOLD}║                    PANOPTIC TEST AUTOMATION DASHBOARD              ║${NC}"
    echo -e "${BLUE}${BOLD}║                         Real-Time Monitoring                       ║${NC}"
    echo -e "${BLUE}${BOLD}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

show_system_status() {
    echo -e "${CYAN}${BOLD}SYSTEM STATUS${NC}"
    echo -e "${CYAN}──────────────────────────────────────────────────────────────────${NC}"
    
    # CPU Usage
    cpu_usage=$(top -l 1 | grep "CPU usage" | awk '{print $3}' | sed 's/%//')
    echo -e "${WHITE}CPU Usage:${NC} ${cpu_usage}%"
    
    # Memory Usage
    memory_pressure=$(memory_pressure | grep "System-wide memory free percentage" | awk '{print $5}' | sed 's/%//')
    echo -e "${WHITE}Memory Usage:${NC} ${memory_pressure}% free"
    
    # Disk Usage
    disk_usage=$(df -h / | awk 'NR==2 {print $5}')
    echo -e "${WHITE}Disk Usage:${NC} ${disk_usage}"
    
    # Network Status
    if ping -c 1 8.8.8.8 &>/dev/null; then
        network_status="${GREEN}Connected${NC}"
    else
        network_status="${RED}Disconnected${NC}"
    fi
    echo -e "${WHITE}Network:${NC} ${network_status}"
    
    # Panoptic Process Status
    if pgrep -f "panoptic" > /dev/null; then
        panoptic_status="${GREEN}Running${NC}"
        panoptic_count=$(pgrep -f "panoptic" | wc -l)
        echo -e "${WHITE}Panoptic Processes:${NC} ${panoptic_count} (${panoptic_status})"
    else
        echo -e "${WHITE}Panoptic Processes:${NC} ${RED}Not Running${NC}"
    fi
    
    echo ""
}

show_test_metrics() {
    echo -e "${YELLOW}${BOLD}TEST METRICS${NC}"
    echo -e "${YELLOW}──────────────────────────────────────────────────────────────────${NC}"
    
    # Initialize metrics file if it doesn't exist
    if [ ! -f "$METRICS_FILE" ]; then
        echo '{"total_tests": 0, "passed_tests": 0, "failed_tests": 0, "coverage": 0, "last_run": null}' > "$METRICS_FILE"
    fi
    
    # Parse metrics
    total_tests=$(jq -r '.total_tests' "$METRICS_FILE" 2>/dev/null || echo "0")
    passed_tests=$(jq -r '.passed_tests' "$METRICS_FILE" 2>/dev/null || echo "0")
    failed_tests=$(jq -r '.failed_tests' "$METRICS_FILE" 2>/dev/null || echo "0")
    coverage=$(jq -r '.coverage' "$METRICS_FILE" 2>/dev/null || echo "0")
    last_run=$(jq -r '.last_run' "$METRICS_FILE" 2>/dev/null || echo "Never")
    
    # Calculate success rate
    if [ "$total_tests" -gt 0 ]; then
        success_rate=$(echo "scale=1; $passed_tests * 100 / $total_tests" | bc)
    else
        success_rate="0"
    fi
    
    # Display metrics with colors based on thresholds
    echo -e "${WHITE}Total Tests:${NC} ${total_tests}"
    
    if [ "$passed_tests" -eq "$total_tests" ] && [ "$total_tests" -gt 0 ]; then
        echo -e "${WHITE}Passed Tests:${NC} ${GREEN}${passed_tests}${NC}"
    else
        echo -e "${WHITE}Passed Tests:${NC} ${GREEN}${passed_tests}${NC}"
    fi
    
    if [ "$failed_tests" -gt 0 ]; then
        echo -e "${WHITE}Failed Tests:${NC} ${RED}${failed_tests}${NC}"
    else
        echo -e "${WHITE}Failed Tests:${NC} ${GREEN}${failed_tests}${NC}"
    fi
    
    # Success rate color coding
    if (( $(echo "$success_rate >= 90" | bc -l) )); then
        success_color="${GREEN}"
    elif (( $(echo "$success_rate >= 70" | bc -l) )); then
        success_color="${YELLOW}"
    else
        success_color="${RED}"
    fi
    echo -e "${WHITE}Success Rate:${NC} ${success_color}${success_rate}%${NC}"
    
    # Coverage color coding
    if (( $(echo "$coverage >= 80" | bc -l) )); then
        coverage_color="${GREEN}"
    elif (( $(echo "$coverage >= 60" | bc -l) )); then
        coverage_color="${YELLOW}"
    else
        coverage_color="${RED}"
    fi
    echo -e "${WHITE}Code Coverage:${NC} ${coverage_color}${coverage}%${NC}"
    
    echo -e "${WHITE}Last Test Run:${NC} ${last_run}"
    
    echo ""
}

show_active_tests() {
    echo -e "${PURPLE}${BOLD}ACTIVE TEST SESSIONS${NC}"
    echo -e "${PURPLE}──────────────────────────────────────────────────────────────────${NC}"
    
    # Find active panoptic processes
    active_processes=$(pgrep -f "panoptic" | head -5)
    
    if [ -n "$active_processes" ]; then
        echo "${WHITE}Active Panoptic processes:${NC}"
        echo ""
        
        # Show details for each active process
        for pid in $active_processes; do
            if [ -f "/proc/$pid/cmdline" ]; then
                cmdline=$(tr '\0' ' ' < "/proc/$pid/cmdline")
                start_time=$(ps -p $pid -o lstart= | awk '{$1=$2=""; print $0}' | xargs)
                
                echo -e "${DIM}PID: ${pid}${NC}"
                echo -e "${DIM}Command: ${cmdline}${NC}"
                echo -e "${DIM}Started: ${start_time}${NC}"
                echo ""
            fi
        done
    else
        echo -e "${DIM}${WHITE}No active test sessions found${NC}"
        echo ""
    fi
}

show_recent_history() {
    echo -e "${GREEN}${BOLD}RECENT TEST HISTORY${NC}"
    echo -e "${GREEN}──────────────────────────────────────────────────────────────────${NC}"
    
    # Create history file if it doesn't exist
    if [ ! -f "$HISTORY_FILE" ]; then
        echo "timestamp,total,passed,failed,coverage,success_rate" > "$HISTORY_FILE"
    fi
    
    # Show last 5 entries
    if [ -f "$HISTORY_FILE" ]; then
        tail -n 6 "$HISTORY_FILE" | while IFS= read -r line; do
            if [ "$line" != "timestamp,total,passed,failed,coverage,success_rate" ]; then
                timestamp=$(echo "$line" | cut -d',' -f1)
                total=$(echo "$line" | cut -d',' -f2)
                passed=$(echo "$line" | cut -d',' -f3)
                failed=$(echo "$line" | cut -d',' -f4)
                coverage=$(echo "$line" | cut -d',' -f5)
                success_rate=$(echo "$line" | cut -d',' -f6)
                
                # Format timestamp
                formatted_time=$(date -r "$timestamp" "+%H:%M:%S %Y-%m-%d" 2>/dev/null || echo "$timestamp")
                
                echo -e "${WHITE}${formatted_time}${NC} - Total: ${total}, Passed: ${GREEN}${passed}${NC}, Failed: ${RED}${failed}${NC}, Coverage: ${coverage}%, Success Rate: ${success_rate}%"
            fi
        done
    else
        echo -e "${DIM}${WHITE}No test history available${NC}"
    fi
    
    echo ""
}

show_alerts() {
    echo -e "${RED}${BOLD}ALERTS & NOTIFICATIONS${NC}"
    echo -e "${RED}──────────────────────────────────────────────────────────────────${NC}"
    
    alerts=()
    
    # Check for test failures
    if [ -f "$METRICS_FILE" ]; then
        failed_tests=$(jq -r '.failed_tests' "$METRICS_FILE" 2>/dev/null || echo "0")
        if [ "$failed_tests" -gt 0 ]; then
            alerts+=("❌ ${failed_tests} test failures detected")
        fi
    fi
    
    # Check for low coverage
    if [ -f "$METRICS_FILE" ]; then
        coverage=$(jq -r '.coverage' "$METRICS_FILE" 2>/dev/null || echo "0")
        if (( $(echo "$coverage < 70" | bc -l) )); then
            alerts+=("⚠️  Low code coverage: ${coverage}%")
        fi
    fi
    
    # Check system resources
    cpu_usage=$(top -l 1 | grep "CPU usage" | awk '{print $3}' | sed 's/%//')
    if (( $(echo "$cpu_usage > 80" | bc -l) )); then
        alerts+=("🔥 High CPU usage: ${cpu_usage}%")
    fi
    
    memory_pressure=$(memory_pressure | grep "System-wide memory free percentage" | awk '{print $5}' | sed 's/%//')
    if [ "$memory_pressure" -lt 20 ]; then
        alerts+=("🧠 Low memory: ${memory_pressure}% free")
    fi
    
    # Check disk space
    disk_usage=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
    if [ "$disk_usage" -gt 85 ]; then
        alerts+=("💾 Low disk space: ${disk_usage}% used")
    fi
    
    if [ ${#alerts[@]} -gt 0 ]; then
        for alert in "${alerts[@]}"; do
            echo -e "${WHITE}${alert}${NC}"
        done
    else
        echo -e "${WHITE}${GREEN}✅ All systems operating normally${NC}"
    fi
    
    echo ""
}

show_actions_menu() {
    echo -e "${CYAN}${BOLD}QUICK ACTIONS${NC}"
    echo -e "${CYAN}──────────────────────────────────────────────────────────────────${NC}"
    
    echo -e "${WHITE}[1]${NC} Run Unit Tests"
    echo -e "${WHITE}[2]${NC} Run Integration Tests"
    echo -e "${WHITE}[3]${NC} Run E2E Tests"
    echo -e "${WHITE}[4]${NC} Run All Tests"
    echo -e "${WHITE}[5]${NC} Generate Coverage Report"
    echo -e "${WHITE}[6]${NC} View Test Logs"
    echo -e "${WHITE}[7]${NC} Clear History"
    echo -e "${WHITE}[8]${NC} Export Metrics"
    echo -e "${WHITE}[q]${NC} Exit Dashboard"
    echo ""
    
    echo -e "${DIM}Press key to execute action (dashboard will refresh every ${REFRESH_INTERVAL}s)${NC}"
}

update_metrics() {
    # This would be called by test scripts to update metrics
    local total=$1
    local passed=$2
    local failed=$3
    local coverage=$4
    
    local timestamp=$(date +%s)
    
    jq --arg total "$total" --arg passed "$passed" --arg failed "$failed" \
       --arg coverage "$coverage" --arg timestamp "$timestamp" \
       '.total_tests = ($total | tonumber) | .passed_tests = ($passed | tonumber) | 
        .failed_tests = ($failed | tonumber) | .coverage = ($coverage | tonumber) | 
        .last_run = $timestamp' "$METRICS_FILE" > "${METRICS_FILE}.tmp" && \
       mv "${METRICS_FILE}.tmp" "$METRICS_FILE"
    
    # Add to history
    if [ "$total" -gt 0 ]; then
        success_rate=$(echo "scale=1; $passed * 100 / $total" | bc)
        echo "${timestamp},${total},${passed},${failed},${coverage},${success_rate}" >> "$HISTORY_FILE"
    fi
}

handle_user_input() {
    local key=$1
    
    case $key in
        1)
            echo -e "${WHITE}Running unit tests...${NC}"
            ./scripts/test.sh --skip-integration --skip-e2e &
            ;;
        2)
            echo -e "${WHITE}Running integration tests...${NC}"
            ./scripts/test.sh --skip-e2e &
            ;;
        3)
            echo -e "${WHITE}Running E2E tests...${NC}"
            ./scripts/test.sh --skip-integration &
            ;;
        4)
            echo -e "${WHITE}Running all tests...${NC}"
            ./scripts/test.sh &
            ;;
        5)
            echo -e "${WHITE}Generating coverage report...${NC}"
            ./scripts/coverage.sh &
            ;;
        6)
            echo -e "${WHITE}Opening test logs...${NC}"
            if [ -d "output/logs" ]; then
                tail -f output/logs/panoptic.log
            else
                echo -e "${RED}No test logs found${NC}"
            fi
            ;;
        7)
            echo -e "${WHITE}Clearing history...${NC}"
            > "$HISTORY_FILE"
            echo "timestamp,total,passed,failed,coverage,success_rate" > "$HISTORY_FILE"
            ;;
        8)
            echo -e "${WHITE}Exporting metrics...${NC}"
            cp "$METRICS_FILE" "panoptic_metrics_$(date +%Y%m%d_%H%M%S).json"
            cp "$HISTORY_FILE" "panoptic_history_$(date +%Y%m%d_%H%M%S).csv"
            ;;
        q|Q)
            echo -e "${WHITE}Exiting dashboard...${NC}"
            exit 0
            ;;
    esac
}

# Main dashboard loop
dashboard_main() {
    echo -e "${BLUE}Starting Panoptic Test Automation Dashboard...${NC}"
    echo -e "${BLUE}Refresh interval: ${REFRESH_INTERVAL}s${NC}"
    echo ""
    
    # Initialize terminal
    stty -echo 2>/dev/null || true
    tput civis 2>/dev/null || true
    
    # Main loop
    while true; do
        show_header
        show_system_status
        show_test_metrics
        show_active_tests
        show_recent_history
        show_alerts
        show_actions_menu
        
        # Wait for user input or timeout
        read -t $REFRESH_INTERVAL -n 1 key
        
        if [ $? -eq 0 ]; then
            handle_user_input "$key"
            sleep 1  # Brief pause after action
        fi
    done
}

# Cleanup function
cleanup() {
    echo -e "${WHITE}Shutting down dashboard...${NC}"
    stty echo 2>/dev/null || true
    tput cnorm 2>/dev/null || true
    exit 0
}

# Signal handlers
trap cleanup SIGINT SIGTERM

# Check dependencies
if ! command -v jq &> /dev/null; then
    echo -e "${RED}Error: jq is required for dashboard functionality${NC}"
    echo -e "${WHITE}Install with: brew install jq (macOS) or apt-get install jq (Linux)${NC}"
    exit 1
fi

if ! command -v bc &> /dev/null; then
    echo -e "${RED}Error: bc is required for dashboard functionality${NC}"
    echo -e "${WHITE}Install with: brew install bc (macOS) or apt-get install bc (Linux)${NC}"
    exit 1
fi

# Create help mode
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo "Panoptic Test Automation Dashboard"
    echo ""
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  --help, -h     Show this help message"
    echo "  --update        Update metrics (internal use)"
    echo "  --metrics N     Set current metrics (internal use)"
    echo ""
    echo "Interactive Mode:"
    echo "  Press number keys to execute actions"
    echo "  Press 'q' to quit"
    echo "  Dashboard refreshes every ${REFRESH_INTERVAL}s"
    exit 0
fi

# Command line mode for metrics updates
if [ "$1" = "--update" ]; then
    update_metrics "$2" "$3" "$4" "$5"
    exit 0
fi

# Start dashboard
dashboard_main