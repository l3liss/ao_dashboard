#!/bin/bash

# ---------------------------------------------------
# Anarchy Online Dashboard Launcher
# Starts backend (Go) and frontend (Python) together
# ---------------------------------------------------

# Color codes
GREEN="\\033[0;32m"
CYAN="\\033[0;36m"
RED="\\033[0;31m"
NC="\\033[0m" # No Color

# Go to script directory to keep paths consistent
cd "$(dirname "$0")"

echo -e "${CYAN}Starting AO Tracker Backend...${NC}"

# Move to backend and run Go tracker with all files
cd backend
go run . &
BACKEND_PID=$!
cd ..

sleep 1
echo -e "${GREEN}Backend running (PID $BACKEND_PID)${NC}"

# Launch Python GUI frontend
cd frontend
echo -e "${CYAN}Starting AO Tracker Frontend...${NC}"
python3 main.py
cd ..

# Cleanup
# Kill backend and any child processes
echo -e "${RED}Frontend closed. Killing backend...${NC}"
pkill -P $BACKEND_PID 2>/dev/null
kill $BACKEND_PID 2>/dev/null
wait $BACKEND_PID 2>/dev/null

echo -e "${CYAN}Shutdown complete.${NC}"
