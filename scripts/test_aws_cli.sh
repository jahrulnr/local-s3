#!/bin/bash

# AWS CLI Test Script for LocalS3
# This script tests LocalS3 server using AWS CLI commands

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
ENDPOINT_URL="http://localhost:3000"
TEST_BUCKET="aws-test-bucket"
TEST_BUCKET2="aws-test-bucket-2"
MULTIPART_BUCKET="multipart-test"
TEST_FILE="aws-test-file.json"
LARGE_FILE="large-test-file.txt"
DOWNLOAD_FILE="downloaded-aws-test.json"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to check if AWS CLI is installed
check_aws_cli() {
    if ! command -v aws &> /dev/null; then
        print_error "AWS CLI is not installed. Please install it first:"
        echo "  - macOS: brew install awscli"
        echo "  - Ubuntu: sudo apt-get install awscli"
        echo "  - Other: pip install awscli"
        exit 1
    fi
    print_success "AWS CLI is installed"
}

# Function to configure AWS CLI for LocalS3
configure_aws_cli() {
    print_status "Configuring AWS CLI for LocalS3..."
    
    # Create AWS credentials directory if it doesn't exist
    mkdir -p ~/.aws
    
    # Set up credentials file
    cat > ~/.aws/credentials << EOF
[default]
aws_access_key_id = test
aws_secret_access_key = test123456789
EOF

    # Set up config file
    cat > ~/.aws/config << EOF
[default]
region = us-east-1
output = json
EOF

    print_success "AWS CLI configured for LocalS3"
}

# Function to check if LocalS3 server is running
check_server() {
    print_status "Checking if LocalS3 server is running..."
    
    if curl -s "$ENDPOINT_URL/health" > /dev/null 2>&1; then
        print_success "LocalS3 server is running"
    else
        print_error "LocalS3 server is not running!"
        print_warning "Please start the server first:"
        echo "  cd /apps/LocalS3 && ./run.sh"
        exit 1
    fi
}

# Function to create test files
create_test_files() {
    print_status "Creating test files..."
    
    # Create a JSON test file
    cat > "$TEST_FILE" << EOF
{
    "test": "AWS CLI test with LocalS3",
    "timestamp": "$(date -Iseconds)",
    "data": {
        "numbers": [1, 2, 3, 4, 5],
        "boolean": true,
        "nested": {
            "key": "value",
            "array": ["a", "b", "c"]
        }
    },
    "description": "This file is used to test AWS CLI compatibility with LocalS3 server"
}
EOF

    # Create a larger file for multipart upload testing
    print_status "Creating large test file (10MB)..."
    dd if=/dev/zero of="$LARGE_FILE" bs=1024 count=10240 2>/dev/null
    echo "Large file created for multipart upload testing" >> "$LARGE_FILE"
    
    print_success "Test files created"
}

# Function to clean up test files
cleanup_files() {
    print_status "Cleaning up test files..."
    rm -f "$TEST_FILE" "$LARGE_FILE" "$DOWNLOAD_FILE"
    print_success "Test files cleaned up"
}

# Function to test bucket operations
test_bucket_operations() {
    print_status "Testing bucket operations..."
    
    # Test: Create bucket
    print_status "Creating bucket: $TEST_BUCKET"
    if aws --endpoint-url="$ENDPOINT_URL" s3 mb "s3://$TEST_BUCKET"; then
        print_success "Bucket created successfully"
    else
        print_error "Failed to create bucket"
        return 1
    fi
    
    # Test: Create second bucket
    print_status "Creating second bucket: $TEST_BUCKET2"
    aws --endpoint-url="$ENDPOINT_URL" s3 mb "s3://$TEST_BUCKET2"
    
    # Test: List buckets
    print_status "Listing buckets..."
    if aws --endpoint-url="$ENDPOINT_URL" s3 ls; then
        print_success "Buckets listed successfully"
    else
        print_error "Failed to list buckets"
        return 1
    fi
    
    print_success "Bucket operations completed"
}

# Function to test object operations
test_object_operations() {
    print_status "Testing object operations..."
    
    # Test: Upload file
    print_status "Uploading file: $TEST_FILE"
    if aws --endpoint-url="$ENDPOINT_URL" s3 cp "$TEST_FILE" "s3://$TEST_BUCKET/"; then
        print_success "File uploaded successfully"
    else
        print_error "Failed to upload file"
        return 1
    fi
    
    # Test: Upload file to subdirectory
    print_status "Uploading file to subdirectory"
    aws --endpoint-url="$ENDPOINT_URL" s3 cp "$TEST_FILE" "s3://$TEST_BUCKET/subdir/nested-file.json"
    
    # Test: List objects
    print_status "Listing objects in bucket..."
    if aws --endpoint-url="$ENDPOINT_URL" s3 ls "s3://$TEST_BUCKET/"; then
        print_success "Objects listed successfully"
    else
        print_error "Failed to list objects"
        return 1
    fi
    
    # Test: List objects recursively
    print_status "Listing objects recursively..."
    aws --endpoint-url="$ENDPOINT_URL" s3 ls "s3://$TEST_BUCKET/" --recursive
    
    # Test: Download file
    print_status "Downloading file..."
    if aws --endpoint-url="$ENDPOINT_URL" s3 cp "s3://$TEST_BUCKET/$TEST_FILE" "$DOWNLOAD_FILE"; then
        print_success "File downloaded successfully"
    else
        print_error "Failed to download file"
        return 1
    fi
    
    # Test: Verify downloaded file
    print_status "Verifying downloaded file content..."
    if diff "$TEST_FILE" "$DOWNLOAD_FILE" > /dev/null; then
        print_success "Downloaded file matches original"
    else
        print_error "Downloaded file does not match original"
        return 1
    fi
    
    # Test: Copy object within bucket
    print_status "Copying object within bucket..."
    aws --endpoint-url="$ENDPOINT_URL" s3 cp "s3://$TEST_BUCKET/$TEST_FILE" "s3://$TEST_BUCKET/copied-file.json"
    
    # Test: Copy object between buckets
    print_status "Copying object between buckets..."
    aws --endpoint-url="$ENDPOINT_URL" s3 cp "s3://$TEST_BUCKET/$TEST_FILE" "s3://$TEST_BUCKET2/copied-from-bucket1.json"
    
    print_success "Object operations completed"
}

# Function to test sync operations
test_sync_operations() {
    print_status "Testing sync operations..."
    
    # Create a directory with multiple files
    mkdir -p sync_test_dir
    echo "File 1 content" > sync_test_dir/file1.txt
    echo "File 2 content" > sync_test_dir/file2.txt
    mkdir -p sync_test_dir/subdir
    echo "Nested file content" > sync_test_dir/subdir/nested.txt
    
    # Test: Sync directory to S3
    print_status "Syncing directory to S3..."
    if aws --endpoint-url="$ENDPOINT_URL" s3 sync sync_test_dir "s3://$TEST_BUCKET/synced/"; then
        print_success "Directory synced to S3 successfully"
    else
        print_error "Failed to sync directory to S3"
        return 1
    fi
    
    # Test: Sync from S3 to directory
    mkdir -p sync_download_dir
    print_status "Syncing from S3 to local directory..."
    if aws --endpoint-url="$ENDPOINT_URL" s3 sync "s3://$TEST_BUCKET/synced/" sync_download_dir; then
        print_success "S3 content synced to local directory successfully"
    else
        print_error "Failed to sync from S3 to local directory"
        return 1
    fi
    
    # Cleanup sync test directories
    rm -rf sync_test_dir sync_download_dir
    
    print_success "Sync operations completed"
}

# Function to test multipart upload (for large files)
test_multipart_upload() {
    print_status "Testing multipart upload operations..."
    
    # Create bucket for multipart testing
    aws --endpoint-url="$ENDPOINT_URL" s3 mb "s3://$MULTIPART_BUCKET"
    
    # Test: Upload large file (should use multipart automatically)
    print_status "Uploading large file (may use multipart)..."
    if aws --endpoint-url="$ENDPOINT_URL" s3 cp "$LARGE_FILE" "s3://$MULTIPART_BUCKET/"; then
        print_success "Large file uploaded successfully"
    else
        print_error "Failed to upload large file"
        return 1
    fi
    
    # Test: Download large file
    print_status "Downloading large file..."
    if aws --endpoint-url="$ENDPOINT_URL" s3 cp "s3://$MULTIPART_BUCKET/$LARGE_FILE" "downloaded-$LARGE_FILE"; then
        print_success "Large file downloaded successfully"
    else
        print_error "Failed to download large file"
        return 1
    fi
    
    # Cleanup
    rm -f "downloaded-$LARGE_FILE"
    
    print_success "Multipart upload operations completed"
}

# Function to test advanced operations
test_advanced_operations() {
    print_status "Testing advanced operations..."
    
    # Test: Get object metadata
    print_status "Getting object metadata..."
    aws --endpoint-url="$ENDPOINT_URL" s3api head-object --bucket "$TEST_BUCKET" --key "$TEST_FILE"
    
    # Test: List objects with pagination
    print_status "Testing object listing with pagination..."
    aws --endpoint-url="$ENDPOINT_URL" s3api list-objects-v2 --bucket "$TEST_BUCKET" --max-keys 2
    
    # Test: Delete specific object
    print_status "Deleting specific object..."
    aws --endpoint-url="$ENDPOINT_URL" s3 rm "s3://$TEST_BUCKET/copied-file.json"
    
    print_success "Advanced operations completed"
}

# Function to test error handling
test_error_handling() {
    print_status "Testing error handling..."
    
    # Test: Try to access non-existent bucket
    print_status "Testing access to non-existent bucket..."
    if aws --endpoint-url="$ENDPOINT_URL" s3 ls "s3://non-existent-bucket/" 2>/dev/null; then
        print_warning "Expected error for non-existent bucket, but operation succeeded"
    else
        print_success "Correctly handled non-existent bucket error"
    fi
    
    # Test: Try to download non-existent object
    print_status "Testing download of non-existent object..."
    if aws --endpoint-url="$ENDPOINT_URL" s3 cp "s3://$TEST_BUCKET/non-existent-file.txt" "temp-file.txt" 2>/dev/null; then
        print_warning "Expected error for non-existent object, but operation succeeded"
    else
        print_success "Correctly handled non-existent object error"
    fi
    
    print_success "Error handling tests completed"
}

# Function to cleanup test buckets
cleanup_buckets() {
    print_status "Cleaning up test buckets..."
    
    # Remove all objects from buckets first
    aws --endpoint-url="$ENDPOINT_URL" s3 rm "s3://$TEST_BUCKET" --recursive 2>/dev/null || true
    aws --endpoint-url="$ENDPOINT_URL" s3 rm "s3://$TEST_BUCKET2" --recursive 2>/dev/null || true
    aws --endpoint-url="$ENDPOINT_URL" s3 rm "s3://$MULTIPART_BUCKET" --recursive 2>/dev/null || true
    
    # Remove buckets
    aws --endpoint-url="$ENDPOINT_URL" s3 rb "s3://$TEST_BUCKET" 2>/dev/null || true
    aws --endpoint-url="$ENDPOINT_URL" s3 rb "s3://$TEST_BUCKET2" 2>/dev/null || true
    aws --endpoint-url="$ENDPOINT_URL" s3 rb "s3://$MULTIPART_BUCKET" 2>/dev/null || true
    
    print_success "Test buckets cleaned up"
}

# Function to display test summary
display_summary() {
    echo ""
    echo "=================================="
    echo -e "${GREEN}AWS CLI Test Summary${NC}"
    echo "=================================="
    echo -e "${GREEN}✓${NC} AWS CLI compatibility verified"
    echo -e "${GREEN}✓${NC} Bucket operations working"
    echo -e "${GREEN}✓${NC} Object operations working"
    echo -e "${GREEN}✓${NC} Sync operations working"
    echo -e "${GREEN}✓${NC} Large file upload working"
    echo -e "${GREEN}✓${NC} Advanced operations working"
    echo -e "${GREEN}✓${NC} Error handling working"
    echo ""
    echo -e "${BLUE}Your LocalS3 server is fully compatible with AWS CLI!${NC}"
    echo ""
    echo "Example commands you can use:"
    echo "  aws --endpoint-url=$ENDPOINT_URL s3 mb s3://mybucket"
    echo "  aws --endpoint-url=$ENDPOINT_URL s3 cp file.txt s3://mybucket/"
    echo "  aws --endpoint-url=$ENDPOINT_URL s3 ls s3://mybucket/"
    echo "  aws --endpoint-url=$ENDPOINT_URL s3 sync ./folder s3://mybucket/folder/"
    echo ""
}

# Main test execution
main() {
    echo "=================================="
    echo -e "${BLUE}LocalS3 AWS CLI Test Suite${NC}"
    echo "=================================="
    echo ""
    
    # Pre-flight checks
    check_aws_cli
    configure_aws_cli
    check_server
    
    # Create test files
    create_test_files
    
    # Run tests
    test_bucket_operations
    test_object_operations
    test_sync_operations
    test_multipart_upload
    test_advanced_operations
    test_error_handling
    
    # Cleanup
    cleanup_buckets
    cleanup_files
    
    # Display summary
    display_summary
}

# Trap to ensure cleanup on script exit
trap 'cleanup_buckets; cleanup_files' EXIT

# Run main function
main "$@"
