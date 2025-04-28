#!/bin/bash

# Start Go backend
cd backend
go run main.go &
BACKEND_PID=$!

# Start Python frontend
cd ../frontend
python3 main.py

# Kill backend when GUI closes
kill $BACKEND_PID
