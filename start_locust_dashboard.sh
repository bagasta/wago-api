#!/bin/bash

# Start Locust Web Dashboard with Virtual Environment Support
# Access dashboard at: http://localhost:8089

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Starting Locust Web Dashboard${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if virtual environment exists
if [ ! -d "venv" ]; then
    echo -e "${YELLOW}Virtual environment not found!${NC}"
    echo -e "${YELLOW}Creating virtual environment...${NC}"
    python3 -m venv venv || {
        echo -e "${RED}Failed to create venv. Please install python3-venv:${NC}"
        echo -e "${YELLOW}  sudo apt install python3-venv python3-full${NC}"
        exit 1
    }
    echo -e "${GREEN}✓ Virtual environment created${NC}"
fi

# Activate virtual environment
echo -e "${YELLOW}Activating virtual environment...${NC}"
source venv/bin/activate

# Check if locust is installed in venv
if ! command -v locust &> /dev/null; then
    echo -e "${YELLOW}Locust not found in venv. Installing...${NC}"
    pip install -r requirements-test.txt || {
        echo -e "${RED}Failed to install Locust${NC}"
        deactivate
        exit 1
    }
fi

echo -e "${GREEN}✓ Locust ready in virtual environment${NC}"
echo ""
echo -e "${BLUE}Dashboard will open at:${NC}"
echo -e "${GREEN}  http://localhost:8089${NC}"
echo ""
echo -e "${YELLOW}In the dashboard, you can set:${NC}"
echo -e "  • Number of users (e.g., 1000)"
echo -e "  • Spawn rate (e.g., 50 users/second)"
echo -e "  • Host URL (e.g., http://localhost:8080)"
echo ""
echo -e "${YELLOW}Starting Locust...${NC}"
echo -e "${BLUE}(Press Ctrl+C to stop)${NC}"
echo ""

# Start Locust with Web UI
locust -f locustfile_auto.py

# Deactivate when done
deactivate
