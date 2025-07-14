#!/bin/bash

# LocalS3 AWS CLI Setup Script
# This script sets up your environment for easy AWS CLI usage with LocalS3

ENDPOINT_URL="http://localhost:3001"
PORT="3001"

echo "LocalS3 AWS CLI Setup"
echo "===================="
echo ""

# Function to create alias
create_alias() {
    echo "Creating convenient alias..."
    
    # Add to current session
    alias s3local="aws --endpoint-url=$ENDPOINT_URL s3 --no-sign-request"
    alias s3localapi="aws --endpoint-url=$ENDPOINT_URL s3api --no-sign-request"
    
    # Add to shell configuration file
    if [ -f ~/.bashrc ]; then
        SHELL_RC=~/.bashrc
    elif [ -f ~/.zshrc ]; then
        SHELL_RC=~/.zshrc
    else
        SHELL_RC=~/.profile
    fi
    
    echo "" >> $SHELL_RC
    echo "# LocalS3 aliases" >> $SHELL_RC
    echo "alias s3local='aws --endpoint-url=$ENDPOINT_URL s3 --no-sign-request'" >> $SHELL_RC
    echo "alias s3localapi='aws --endpoint-url=$ENDPOINT_URL s3api --no-sign-request'" >> $SHELL_RC
    
    echo "✅ Aliases added to $SHELL_RC"
    echo "   s3local    - for S3 commands"
    echo "   s3localapi - for S3 API commands"
}

# Function to create example script
create_examples() {
    cat > aws_cli_examples.sh << 'EOF'
#!/bin/bash

# LocalS3 AWS CLI Examples
# This script demonstrates common operations

ENDPOINT="http://localhost:3001"

echo "LocalS3 AWS CLI Examples"
echo "======================="
echo ""

# Basic commands
echo "📦 Bucket Operations:"
echo "  aws --endpoint-url=$ENDPOINT s3 mb s3://mybucket --no-sign-request"
echo "  aws --endpoint-url=$ENDPOINT s3 ls --no-sign-request"
echo "  aws --endpoint-url=$ENDPOINT s3 rb s3://mybucket --no-sign-request"
echo ""

echo "📄 Object Operations:"
echo "  aws --endpoint-url=$ENDPOINT s3 cp file.txt s3://mybucket/ --no-sign-request"
echo "  aws --endpoint-url=$ENDPOINT s3 ls s3://mybucket/ --no-sign-request"
echo "  aws --endpoint-url=$ENDPOINT s3 cp s3://mybucket/file.txt downloaded.txt --no-sign-request"
echo "  aws --endpoint-url=$ENDPOINT s3 rm s3://mybucket/file.txt --no-sign-request"
echo ""

echo "🔄 Sync Operations:"
echo "  aws --endpoint-url=$ENDPOINT s3 sync ./folder s3://mybucket/folder/ --no-sign-request"
echo "  aws --endpoint-url=$ENDPOINT s3 sync s3://mybucket/folder/ ./downloaded/ --no-sign-request"
echo ""

echo "🔍 Advanced Operations:"
echo "  aws --endpoint-url=$ENDPOINT s3api head-object --bucket mybucket --key file.txt --no-sign-request"
echo "  aws --endpoint-url=$ENDPOINT s3api list-objects-v2 --bucket mybucket --no-sign-request"
echo ""

echo "If you've set up aliases (run ./setup_aws_cli.sh), you can use:"
echo "  s3local mb s3://mybucket"
echo "  s3local cp file.txt s3://mybucket/"
echo "  s3local ls s3://mybucket/"
echo "  s3localapi head-object --bucket mybucket --key file.txt"
EOF

    chmod +x aws_cli_examples.sh
    echo "✅ Created aws_cli_examples.sh with example commands"
}

# Function to create test data
create_test_data() {
    mkdir -p test_data
    
    echo '{"name": "test.json", "type": "JSON test file"}' > test_data/test.json
    echo "Hello World!" > test_data/hello.txt
    echo "# LocalS3 Test Data" > test_data/README.md
    
    mkdir -p test_data/images
    echo "Binary data placeholder" > test_data/images/placeholder.bin
    
    echo "✅ Created test_data/ directory with sample files"
}

# Check if server is running
echo "🔍 Checking if LocalS3 server is running..."
if curl -s "$ENDPOINT_URL/health" > /dev/null; then
    echo "✅ LocalS3 server is running on port $PORT"
else
    echo "❌ LocalS3 server is not running"
    echo "   Start it with: PORT=$PORT go run main.go"
    echo ""
fi

# Setup AWS CLI
echo ""
echo "🔧 Setting up AWS CLI configuration..."

mkdir -p ~/.aws

cat > ~/.aws/credentials << EOF
[default]
aws_access_key_id = test
aws_secret_access_key = test123456789

[locals3]
aws_access_key_id = test
aws_secret_access_key = test123456789
EOF

cat > ~/.aws/config << EOF
[default]
region = us-east-1
output = json

[profile locals3]
region = us-east-1
output = json
EOF

echo "✅ AWS CLI configured"

# Create aliases
echo ""
create_alias

# Create examples
echo ""
create_examples

# Create test data
echo ""
create_test_data

echo ""
echo "🎉 Setup complete!"
echo ""
echo "Quick start:"
echo "  1. Make sure LocalS3 server is running: PORT=$PORT go run main.go"
echo "  2. Try: s3local mb s3://testbucket"
echo "  3. Try: s3local cp test_data/test.json s3://testbucket/"
echo "  4. Try: s3local ls s3://testbucket/"
echo ""
echo "For more examples, run: ./aws_cli_examples.sh"
echo ""
echo "Note: Restart your terminal or run 'source ~/.bashrc' to use aliases"
