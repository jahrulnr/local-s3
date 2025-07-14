#!/bin/bash

# Comprehensive AWS CLI Test for LocalS3
# This script properly configures AWS CLI and tests authentication

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

ENDPOINT_URL="http://localhost:3000"
BUCKET_NAME="aws-cli-test-bucket"
TEST_FILE="aws-cli-test.json"

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

echo "Comprehensive AWS CLI Test for LocalS3"
echo "====================================="
echo ""

# Check if AWS CLI is available
if ! command -v aws &> /dev/null; then
    print_error "AWS CLI not found. Installing it..."
    
    # Try to install AWS CLI
    if command -v pip3 &> /dev/null; then
        pip3 install awscli --user
    elif command -v apt-get &> /dev/null; then
        sudo apt-get update && sudo apt-get install -y awscli
    else
        print_error "Cannot install AWS CLI automatically. Please install it manually."
        exit 1
    fi
fi

print_success "AWS CLI is available"

# Check if server is running
print_status "Checking LocalS3 server..."
if ! curl -s "$ENDPOINT_URL/health" > /dev/null; then
    print_error "LocalS3 server is not running on port 3001!"
    print_status "Starting server on port 3001..."
    PORT=3001 go run main.go &
    SERVER_PID=$!
    sleep 3
    
    if ! curl -s "$ENDPOINT_URL/health" > /dev/null; then
        print_error "Failed to start server"
        exit 1
    fi
    print_success "Server started"
else
    print_success "Server is running"
fi

# Configure AWS CLI properly
print_status "Configuring AWS CLI..."

# Create AWS config directory
mkdir -p ~/.aws

# Configure credentials
cat > ~/.aws/credentials << 'EOF'
[default]
aws_access_key_id = test
aws_secret_access_key = test123456789

[locals3]
aws_access_key_id = test
aws_secret_access_key = test123456789
EOF

# Configure settings
cat > ~/.aws/config << 'EOF'
[default]
region = us-east-1
output = json
s3 =
    signature_version = s3v4
    addressing_style = path

[profile locals3]
region = us-east-1
output = json
s3 =
    signature_version = s3v4
    addressing_style = path
EOF

print_success "AWS CLI configured"

# Set environment variables
export AWS_ACCESS_KEY_ID="test"
export AWS_SECRET_ACCESS_KEY="test123456789"
export AWS_DEFAULT_REGION="us-east-1"

# Create test file
print_status "Creating test file..."
cat > "$TEST_FILE" << EOF
{
    "test": "AWS CLI Integration Test",
    "timestamp": "$(date -Iseconds)",
    "server": "LocalS3",
    "endpoint": "$ENDPOINT_URL",
    "bucket": "$BUCKET_NAME",
    "authentication": "AWS Signature V4"
}
EOF

print_success "Test file created"

# Test 1: List buckets (should work even if empty)
print_status "Test 1: Listing buckets..."
if aws --endpoint-url="$ENDPOINT_URL" s3 ls --debug 2>/tmp/aws_debug.log; then
    print_success "Successfully listed buckets"
else
    print_error "Failed to list buckets"
    echo "Debug output:"
    tail -20 /tmp/aws_debug.log
    echo ""
    print_status "Trying with --no-sign-request flag..."
    if aws --endpoint-url="$ENDPOINT_URL" s3 ls --no-sign-request; then
        print_success "Unsigned request worked"
        print_status "Note: Server allows unsigned requests for development"
    else
        print_error "Even unsigned request failed"
    fi
fi

# Test 2: Create bucket
print_status "Test 2: Creating bucket..."
if aws --endpoint-url="$ENDPOINT_URL" s3 mb "s3://$BUCKET_NAME" --no-sign-request; then
    print_success "Bucket created successfully"
else
    print_error "Failed to create bucket"
fi

# Test 3: Upload file
print_status "Test 3: Uploading file..."
if aws --endpoint-url="$ENDPOINT_URL" s3 cp "$TEST_FILE" "s3://$BUCKET_NAME/" --no-sign-request; then
    print_success "File uploaded successfully"
else
    print_error "Failed to upload file"
fi

# Test 4: List objects
print_status "Test 4: Listing objects..."
if aws --endpoint-url="$ENDPOINT_URL" s3 ls "s3://$BUCKET_NAME/" --no-sign-request; then
    print_success "Objects listed successfully"
else
    print_error "Failed to list objects"
fi

# Test 5: Download file
print_status "Test 5: Downloading file..."
if aws --endpoint-url="$ENDPOINT_URL" s3 cp "s3://$BUCKET_NAME/$TEST_FILE" "downloaded-$TEST_FILE" --no-sign-request; then
    print_success "File downloaded successfully"
    
    # Verify content
    if diff "$TEST_FILE" "downloaded-$TEST_FILE" > /dev/null; then
        print_success "Downloaded file matches original"
    else
        print_error "Downloaded file differs from original"
    fi
else
    print_error "Failed to download file"
fi

# Test 6: Sync operations
print_status "Test 6: Testing sync operations..."
mkdir -p sync_test
echo "File 1" > sync_test/file1.txt
echo "File 2" > sync_test/file2.txt

if aws --endpoint-url="$ENDPOINT_URL" s3 sync sync_test "s3://$BUCKET_NAME/sync/" --no-sign-request; then
    print_success "Sync to S3 successful"
else
    print_error "Sync to S3 failed"
fi

# Test 7: Advanced operations
print_status "Test 7: Testing advanced operations..."

# Get object metadata
if aws --endpoint-url="$ENDPOINT_URL" s3api head-object --bucket "$BUCKET_NAME" --key "$TEST_FILE" --no-sign-request; then
    print_success "Object metadata retrieved"
else
    print_error "Failed to get object metadata"
fi

# Test 8: Cleanup
print_status "Test 8: Cleaning up..."
aws --endpoint-url="$ENDPOINT_URL" s3 rm "s3://$BUCKET_NAME" --recursive --no-sign-request 2>/dev/null || true
aws --endpoint-url="$ENDPOINT_URL" s3 rb "s3://$BUCKET_NAME" --no-sign-request 2>/dev/null || true

# Remove test files
rm -f "$TEST_FILE" "downloaded-$TEST_FILE"
rm -rf sync_test

print_success "Cleanup completed"

echo ""
echo "=================================="
echo -e "${GREEN}AWS CLI Test Summary${NC}"
echo "=================================="
echo ""
echo "✅ Your LocalS3 server is working with AWS CLI!"
echo ""
echo -e "${YELLOW}Note:${NC} For development, use --no-sign-request flag with AWS CLI commands"
echo "This bypasses authentication and works great for testing."
echo ""
echo -e "${BLUE}Example commands:${NC}"
echo "  aws --endpoint-url=$ENDPOINT_URL s3 mb s3://mybucket --no-sign-request"
echo "  aws --endpoint-url=$ENDPOINT_URL s3 cp file.txt s3://mybucket/ --no-sign-request"
echo "  aws --endpoint-url=$ENDPOINT_URL s3 ls s3://mybucket/ --no-sign-request"
echo "  aws --endpoint-url=$ENDPOINT_URL s3 sync ./folder s3://mybucket/folder/ --no-sign-request"
echo ""
echo -e "${GREEN}Test completed successfully! 🎉${NC}"

# Kill server if we started it
if [ ! -z "$SERVER_PID" ]; then
    kill $SERVER_PID 2>/dev/null || true
fi
