#!/bin/bash

# Automated Load Testing Script for WhatsApp API
# This script runs fully automated load tests with thousands of users

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_HOST="${API_HOST:-http://localhost:8080}"
USERS="${USERS:-1000}"
SPAWN_RATE="${SPAWN_RATE:-50}"
RUN_TIME="${RUN_TIME:-5m}"
REPORT_DIR="load_test_reports"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}WhatsApp API - Automated Load Testing${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Function to check if service is running
check_service() {
    echo -e "${YELLOW}→ Checking if API is running...${NC}"
    if curl -s "${API_HOST}/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ API is running${NC}"
        return 0
    else
        echo -e "${RED}✗ API is not running at ${API_HOST}${NC}"
        echo -e "${YELLOW}Please start the API server first${NC}"
        return 1
    fi
}

# Function to check environment variable
check_env() {
    echo -e "${YELLOW}→ Checking APP_ENV setting...${NC}"
    
    # Check if .env file exists
    if [ -f ".env" ]; then
        if grep -q "APP_ENV=testing\|APP_ENV=development" .env; then
            echo -e "${GREEN}✓ APP_ENV is set correctly for testing${NC}"
        else
            echo -e "${YELLOW}⚠ APP_ENV should be 'testing' or 'development'${NC}"
            echo -e "${YELLOW}  Current .env setting may not enable test endpoint${NC}"
        fi
    else
        echo -e "${RED}✗ .env file not found${NC}"
    fi
}

# Function to check dependencies
check_dependencies() {
    echo -e "${YELLOW}→ Checking dependencies...${NC}"
    
    if ! command -v locust &> /dev/null; then
        echo -e "${RED}✗ Locust not installed${NC}"
        echo -e "${YELLOW}Installing Locust...${NC}"
        pip install -r requirements-test.txt
        echo -e "${GREEN}✓ Locust installed${NC}"
    else
        echo -e "${GREEN}✓ Locust is installed${NC}"
    fi
}

# Function to seed test sessions (optional)
seed_test_sessions() {
    echo -e "${YELLOW}→ Do you want to pre-seed test sessions? (y/N)${NC}"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        echo -e "${YELLOW}Seeding test sessions...${NC}"
        if command -v psql &> /dev/null; then
            psql -h localhost -U postgres -d whatsapp_api -f tests/seed_test_sessions.sql
            echo -e "${GREEN}✓ Test sessions seeded${NC}"
        else
            echo -e "${RED}✗ psql not found. Please seed manually${NC}"
        fi
    fi
}

# Function to create report directory
prepare_reports() {
    echo -e "${YELLOW}→ Preparing report directory...${NC}"
    mkdir -p "${REPORT_DIR}"
    TIMESTAMP=$(date +%Y%m%d_%H%M%S)
    REPORT_FILE="${REPORT_DIR}/load_test_${TIMESTAMP}.html"
    echo -e "${GREEN}✓ Reports will be saved to: ${REPORT_FILE}${NC}"
}

# Function to run load test
run_load_test() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Starting Load Test${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo -e "Host:       ${GREEN}${API_HOST}${NC}"
    echo -e "Users:      ${GREEN}${USERS}${NC}"
    echo -e "Spawn Rate: ${GREEN}${SPAWN_RATE}/sec${NC}"
    echo -e "Duration:   ${GREEN}${RUN_TIME}${NC}"
    echo -e "Report:     ${GREEN}${REPORT_FILE}${NC}"
    echo ""
    
    locust \
        -f locustfile_auto.py \
        --host="${API_HOST}" \
        --users="${USERS}" \
        --spawn-rate="${SPAWN_RATE}" \
        --run-time="${RUN_TIME}" \
        --headless \
        --html="${REPORT_FILE}" \
        --csv="${REPORT_DIR}/load_test_${TIMESTAMP}" \
        --logfile="${REPORT_DIR}/load_test_${TIMESTAMP}.log"
}

# Function to display results
show_results() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Load Test Completed${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo -e "${GREEN}✓ HTML Report: ${REPORT_FILE}${NC}"
    echo -e "${GREEN}✓ CSV Data:    ${REPORT_DIR}/load_test_${TIMESTAMP}_stats.csv${NC}"
    echo -e "${GREEN}✓ Log File:    ${REPORT_DIR}/load_test_${TIMESTAMP}.log${NC}"
    echo ""
    echo -e "${YELLOW}Open the HTML report in your browser:${NC}"
    echo -e "  xdg-open ${REPORT_FILE}  # Linux"
    echo -e "  open ${REPORT_FILE}      # Mac"
    echo ""
}

# Function to cleanup
cleanup_sessions() {
    echo -e "${YELLOW}→ Do you want to cleanup test sessions? (y/N)${NC}"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        echo -e "${YELLOW}Cleaning up test sessions...${NC}"
        if command -v psql &> /dev/null; then
            psql -h localhost -U postgres -d whatsapp_api -c "DELETE FROM sessions WHERE agent_id LIKE 'load_test_%';"
            echo -e "${GREEN}✓ Test sessions cleaned up${NC}"
        else
            echo -e "${YELLOW}⚠ Please cleanup manually using:${NC}"
            echo -e "  DELETE FROM sessions WHERE agent_id LIKE 'load_test_%';"
        fi
    fi
}

# Main execution
main() {
    check_dependencies
    check_service || exit 1
    check_env
    seed_test_sessions
    prepare_reports
    
    echo ""
    echo -e "${YELLOW}→ Ready to start load test. Continue? (y/N)${NC}"
    read -r response
    if [[ ! "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        echo -e "${RED}Load test cancelled${NC}"
        exit 0
    fi
    
    run_load_test
    show_results
    cleanup_sessions
    
    echo -e "${GREEN}Done!${NC}"
}

# Run main function
main
