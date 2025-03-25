#!/bin/bash

# Get ports from command line argument or use default
PORTS=${1:-"8787,8788,9229,9230"}

# Kill processes on the ports
echo "Killing processes on ports: $PORTS"
lsof -ti:$PORTS | xargs kill -9 2>/dev/null || true 