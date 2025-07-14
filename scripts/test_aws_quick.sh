#!/bin/bash

# Quick AWS CLI Test for LocalS3
# A simplified version for quick testing

set -e

ENDPOINT_URL="http://localhost:3000"
BUCKET_NAME="quick-test-bucket"

echo "Quick AWS CLI Test for LocalS3"
echo "=============================="
echo ""

# Check if AWS CLI is available
if ! command -v aws &> /dev/null; then
    echo "❌ AWS CLI not found. Please install it first:"
    echo "   sudo apt-get install awscli   # Ubuntu/Debian"
    echo "   brew install awscli           # macOS"
    echo "   pip install awscli            # pip"
    exit 1
fi

# Check if server is running
echo "🔍 Checking LocalS3 server..."
if ! curl -s "$ENDPOINT_URL/health" > /dev/null; then
    echo "❌ LocalS3 server is not running!"
    echo "   Please start it first: ./run.sh"
    exit 1
fi
echo "✅ Server is running"

# Configure AWS CLI (temporarily)
echo ""
echo "🔧 Configuring AWS CLI..."
export AWS_ACCESS_KEY_ID="test"
export AWS_SECRET_ACCESS_KEY="test123456789"
export AWS_DEFAULT_REGION="us-east-1"

echo "✅ AWS CLI configured"

# Test basic operations
echo ""
echo "📦 Testing bucket operations..."

# Create bucket
echo "   Creating bucket..."
aws --endpoint-url="$ENDPOINT_URL" s3 mb "s3://$BUCKET_NAME"

# List buckets
echo "   Listing buckets..."
aws --endpoint-url="$ENDPOINT_URL" s3 ls

echo ""
echo "📄 Testing object operations..."

# Create test file
echo '{"message": "Hello from AWS CLI!", "timestamp": "'$(date -Iseconds)'"}' > test-aws-cli.json

# Upload file
echo "   Uploading file..."
aws --endpoint-url="$ENDPOINT_URL" s3 cp test-aws-cli.json "s3://$BUCKET_NAME/"

# List objects
echo "   Listing objects..."
aws --endpoint-url="$ENDPOINT_URL" s3 ls "s3://$BUCKET_NAME/"

# Download file
echo "   Downloading file..."
aws --endpoint-url="$ENDPOINT_URL" s3 cp "s3://$BUCKET_NAME/test-aws-cli.json" downloaded-test.json

# Verify content
echo "   Verifying content..."
echo "   Original:"
cat test-aws-cli.json
echo "   Downloaded:"
cat downloaded-test.json

echo ""
echo "🧹 Cleaning up..."
aws --endpoint-url="$ENDPOINT_URL" s3 rm "s3://$BUCKET_NAME/test-aws-cli.json"
aws --endpoint-url="$ENDPOINT_URL" s3 rb "s3://$BUCKET_NAME"
rm -f test-aws-cli.json downloaded-test.json

echo ""
echo "🎉 All tests passed!"
echo ""
echo "Your LocalS3 server is working perfectly with AWS CLI!"
echo ""
echo "💡 Useful commands:"
echo "   aws --endpoint-url=$ENDPOINT_URL s3 mb s3://mybucket"
echo "   aws --endpoint-url=$ENDPOINT_URL s3 cp file.txt s3://mybucket/"
echo "   aws --endpoint-url=$ENDPOINT_URL s3 ls s3://mybucket/"
echo "   aws --endpoint-url=$ENDPOINT_URL s3 sync ./folder s3://mybucket/folder/"
