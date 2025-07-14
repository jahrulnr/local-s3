#!/bin/bash

# LocalS3 Test Menu
# Interactive menu for running different tests

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

print_header() {
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════╗"
    echo "║          LocalS3 Test Suite          ║"
    echo "╚══════════════════════════════════════╝"
    echo -e "${NC}"
}

print_menu() {
    echo ""
    echo -e "${BLUE}Available Tests:${NC}"
    echo ""
    echo -e "${GREEN}1.${NC} Basic HTTP API Test (curl)"
    echo -e "${GREEN}2.${NC} Sample Data Test (JSON upload)"
    echo -e "${GREEN}3.${NC} Direct HTTP Test"
    echo -e "${GREEN}4.${NC} AWS CLI Quick Test"
    echo -e "${GREEN}5.${NC} AWS CLI Complete Test"
    echo -e "${GREEN}6.${NC} Setup AWS CLI"
    echo -e "${GREEN}7.${NC} Run All Tests"
    echo ""
    echo -e "${YELLOW}Server Management:${NC}"
    echo -e "${GREEN}s.${NC} Start LocalS3 Server"
    echo -e "${GREEN}k.${NC} Stop LocalS3 Server"
    echo -e "${GREEN}r.${NC} Restart LocalS3 Server"
    echo ""
    echo -e "${GREEN}q.${NC} Quit"
    echo ""
}

check_server() {
    if curl -s "http://localhost:3000/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Server is running on port 3000"
        return 0
    elif curl -s "http://localhost:3001/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Server is running on port 3001"
        return 0
    else
        echo -e "${RED}✗${NC} Server is not running"
        return 1
    fi
}

start_server() {
    echo -e "${BLUE}Starting LocalS3 Server...${NC}"
    
    if check_server; then
        echo "Server is already running!"
        return 0
    fi
    
    echo "Starting server on port 3000..."
    cd .. && go run main.go &
    sleep 3
    
    if check_server; then
        echo -e "${GREEN}✓${NC} Server started successfully"
    else
        echo -e "${RED}✗${NC} Failed to start server"
    fi
}

stop_server() {
    echo -e "${BLUE}Stopping LocalS3 Server...${NC}"
    pkill -f "go run main.go" || true
    sleep 1
    echo -e "${GREEN}✓${NC} Server stopped"
}

restart_server() {
    stop_server
    sleep 2
    start_server
}

run_test() {
    local test_name="$1"
    local test_script="$2"
    
    echo ""
    echo -e "${CYAN}Running: $test_name${NC}"
    echo "=================================="
    
    if [ -f "$test_script" ]; then
        chmod +x "$test_script"
        if ./"$test_script"; then
            echo -e "${GREEN}✓ Test completed successfully${NC}"
        else
            echo -e "${RED}✗ Test failed${NC}"
        fi
    else
        echo -e "${RED}✗ Test script not found: $test_script${NC}"
    fi
    
    echo ""
    read -p "Press Enter to continue..."
}

run_all_tests() {
    echo -e "${CYAN}Running All Tests${NC}"
    echo "=================="
    
    if ! check_server; then
        echo "Starting server for tests..."
        start_server
    fi
    
    echo ""
    echo "Test 1/5: Basic HTTP API Test"
    ./test.sh || echo "Test 1 failed"
    
    echo ""
    echo "Test 2/5: Sample Data Test"
    ./test_sample.sh || echo "Test 2 failed"
    
    echo ""
    echo "Test 3/5: Direct HTTP Test"
    ./test_direct_http.sh || echo "Test 3 failed"
    
    echo ""
    echo "Test 4/5: AWS CLI Quick Test"
    ./test_aws_quick.sh || echo "Test 4 failed"
    
    echo ""
    echo "Test 5/5: AWS CLI Complete Test"
    ./test_aws_cli_complete.sh || echo "Test 5 failed"
    
    echo ""
    echo -e "${GREEN}All tests completed!${NC}"
    read -p "Press Enter to continue..."
}

main() {
    while true; do
        clear
        print_header
        check_server
        print_menu
        
        read -p "Choose an option: " choice
        
        case $choice in
            1)
                run_test "Basic HTTP API Test" "test.sh"
                ;;
            2)
                run_test "Sample Data Test" "test_sample.sh"
                ;;
            3)
                run_test "Direct HTTP Test" "test_direct_http.sh"
                ;;
            4)
                run_test "AWS CLI Quick Test" "test_aws_quick.sh"
                ;;
            5)
                run_test "AWS CLI Complete Test" "test_aws_cli_complete.sh"
                ;;
            6)
                echo ""
                echo -e "${BLUE}Setting up AWS CLI...${NC}"
                ./setup_aws_cli.sh
                echo ""
                read -p "Press Enter to continue..."
                ;;
            7)
                run_all_tests
                ;;
            s|S)
                start_server
                echo ""
                read -p "Press Enter to continue..."
                ;;
            k|K)
                stop_server
                echo ""
                read -p "Press Enter to continue..."
                ;;
            r|R)
                restart_server
                echo ""
                read -p "Press Enter to continue..."
                ;;
            q|Q)
                echo ""
                echo -e "${GREEN}Thank you for using LocalS3!${NC}"
                exit 0
                ;;
            *)
                echo ""
                echo -e "${RED}Invalid option. Please try again.${NC}"
                sleep 1
                ;;
        esac
    done
}

# Run main function
main
