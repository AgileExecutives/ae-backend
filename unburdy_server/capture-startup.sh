#!/bin/bash

# Capture server startup logs to file
echo "Starting server and capturing logs to startup.log..."
echo "Press Ctrl+C to stop the server"
echo ""

./bin/unburdy-server 2>&1 | tee startup.log
