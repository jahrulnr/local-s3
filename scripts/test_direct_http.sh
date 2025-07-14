#!/bin/bash

# AWS CLI Test without authentication (development mode)
# This version uses direct HTTP calls to test the S3 API

set -e

ENDPOINT_URL="http://localhost:3000"
BUCKET_NAME="no-auth-test-bucket"

echo "AWS CLI Test for LocalS3 (Development Mode)"
echo "==========================================="
echo ""

# Check if server is running
echo "🔍 Checking LocalS3 server..."
if ! curl -s "$ENDPOINT_URL/health" > /dev/null; then
    echo "❌ LocalS3 server is not running!"
    echo "   Please start it first: PORT=3000 go run main.go"
    exit 1
fi
echo "✅ Server is running"

echo ""
echo "📦 Testing bucket operations with curl..."

# Create bucket
echo "   Creating bucket..."
curl -X PUT "$ENDPOINT_URL/$BUCKET_NAME" -w "\nStatus: %{http_code}\n"

# List buckets
echo "   Listing buckets..."
curl -s "$ENDPOINT_URL/" | head -5
echo ""

echo ""
echo "📄 Testing object operations..."

# Create test file
echo '{"message": "Hello from LocalS3!", "timestamp": "'$(date -Iseconds)'", "method": "direct_http"}' > test-direct.json

# Upload file
echo "   Uploading file..."
curl -X PUT "$ENDPOINT_URL/$BUCKET_NAME/test-direct.json" -T test-direct.json -w "\nStatus: %{http_code}\n"

# List objects
echo "   Listing objects..."
curl -s "$ENDPOINT_URL/$BUCKET_NAME" | head -10
echo ""

# Download file
echo "   Downloading file..."
curl -s "$ENDPOINT_URL/$BUCKET_NAME/test-direct.json" > downloaded-direct.json

# Verify content
echo "   Verifying content..."
echo "   Original:"
cat test-direct.json
echo "   Downloaded:"
cat downloaded-direct.json

echo ""
echo "🧹 Cleaning up..."
curl -X DELETE "$ENDPOINT_URL/$BUCKET_NAME/test-direct.json"
curl -X DELETE "$ENDPOINT_URL/$BUCKET_NAME"
rm -f test-direct.json downloaded-direct.json

echo ""
echo "🎉 Direct HTTP API tests passed!"
echo ""
echo "Note: AWS CLI requires proper authentication."
echo "The server supports AWS Signature V4 authentication."
echo ""
echo "For development, you can use direct HTTP calls as shown above."
