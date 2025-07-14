#!/bin/bash

# LocalS3 Startup Script

set -e

echo "Starting LocalS3 Server..."
echo "=========================="

# Set default environment variables if not already set
export PORT=${PORT:-3000}
export DATA_DIR=${DATA_DIR:-./data}
export ACCESS_KEY=${ACCESS_KEY:-test}
export SECRET_KEY=${SECRET_KEY:-test123456789}
export REGION=${REGION:-us-east-1}
export LOG_LEVEL=${LOG_LEVEL:-info}
export BASE_DOMAIN=${BASE_DOMAIN:-localhost:$PORT}

echo "Configuration:"
echo "  Port: $PORT"
echo "  Data Directory: $DATA_DIR"
echo "  Access Key: $ACCESS_KEY"
echo "  Region: $REGION"
echo "  Log Level: $LOG_LEVEL"
echo "  Base Domain: $BASE_DOMAIN"
echo ""

# Create data directory if it doesn't exist
mkdir -p "$DATA_DIR"

# Check if binary exists, build if not
if [ ! -f "./locals3" ]; then
    echo "Building LocalS3..."
    go mod tidy
    go build -o locals3
    echo "Build complete!"
    echo ""
fi

echo "Starting server on http://localhost:$PORT"
echo "Health check: http://localhost:$PORT/health"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

# Start the server
./locals3
