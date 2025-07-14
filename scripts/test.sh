#!/bin/bash

# LocalS3 Test Script

set -e

SERVER_URL="http://localhost:3000"
TEST_BUCKET="testbucket"
TEST_FILE="test.json"

echo "LocalS3 Test Script"
echo "==================="
echo ""

# Check if server is running
echo "1. Checking server health..."
if curl -s "$SERVER_URL/health" > /dev/null; then
    echo "   ✓ Server is running"
else
    echo "   ✗ Server is not running. Please start LocalS3 first."
    exit 1
fi

echo ""

# Test bucket creation
echo "2. Creating test bucket..."
if curl -s -X PUT "$SERVER_URL/$TEST_BUCKET" > /dev/null; then
    echo "   ✓ Bucket created successfully"
else
    echo "   ✗ Failed to create bucket"
    exit 1
fi

echo ""

# Test file upload
echo "3. Creating test file..."
echo '{"message": "Hello from LocalS3!", "timestamp": "'$(date -Iseconds)'"}' > $TEST_FILE

echo "4. Uploading test file..."
if curl -s -X PUT -T "$TEST_FILE" "$SERVER_URL/$TEST_BUCKET/$TEST_FILE" > /dev/null; then
    echo "   ✓ File uploaded successfully"
else
    echo "   ✗ Failed to upload file"
    exit 1
fi

echo ""

# Test file download
echo "5. Downloading test file..."
if curl -s "$SERVER_URL/$TEST_BUCKET/$TEST_FILE" -o "downloaded_$TEST_FILE"; then
    echo "   ✓ File downloaded successfully"
    echo "   Content:"
    cat "downloaded_$TEST_FILE"
    echo ""
else
    echo "   ✗ Failed to download file"
    exit 1
fi

echo ""

# Test bucket listing
echo "6. Listing buckets..."
echo "   Response:"
curl -s "$SERVER_URL/" | head -20
echo ""

echo ""

# Test object listing
echo "7. Listing objects in bucket..."
echo "   Response:"
curl -s "$SERVER_URL/$TEST_BUCKET" | head -20
echo ""

echo ""

# Cleanup
echo "8. Cleaning up..."
rm -f "$TEST_FILE" "downloaded_$TEST_FILE"
echo "   ✓ Test files cleaned up"

echo ""
echo "All tests completed successfully! 🎉"
echo ""
echo "Your LocalS3 server is working correctly."
echo "You can now use it with AWS CLI or any S3-compatible tool."
echo ""
echo "Example AWS CLI usage:"
echo "  aws --endpoint-url=$SERVER_URL s3 ls"
echo "  aws --endpoint-url=$SERVER_URL s3 mb s3://mybucket"
echo "  aws --endpoint-url=$SERVER_URL s3 cp file.txt s3://mybucket/"
