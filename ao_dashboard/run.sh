#!/bin/bash

# ---------------------------------------------------
# Anarchy Online Dashboard Launcher
# Starts backend (Go) and frontend (Python) together
# ---------------------------------------------------

# Colors
GREEN="\\033[0;32m"
CYAN="\\033[0;36m"
RED="\\033[0;31m"
NC="\\033[0m" # No Color

echo -e "${CYAN}Starting AO Tracker Backend...${NC}"

# Move to backend and run Go backend
cd backend
go run main.go &
BACKEND_PID=$!
sleep 1

echo -e "${GREEN}Backend running (PID $BACKEND_PID)${NC}"

# Move to frontend and run Python GUI
cd ../frontend
echo -e "${CYAN}Starting AO Tracker Frontend...${NC}"
python3 main.py

# When frontend exits, kill backend
echo -e "${RED}Frontend closed. Killing backend...${NC}"
kill $BACKEND_PID
wait $BACKEND_PID 2>/dev/null

echo -e "${CYAN}Shutdown complete.${NC}"
