# LocalS3 - Local S3 Compatible Server

LocalS3 is a lightweight, S3-compatible object storage server written in Go. It can be used as a drop-in replacement for MinIO or AWS S3 during development and testing.

## Features

- **S3 Compatible API**: Supports standard S3 operations including bucket and object management
- **AWS Signature V4 Authentication**: Compatible with AWS SDKs and tools
- **File System Storage**: Uses local file system for object storage
- **Multipart Upload Support**: Handles large file uploads efficiently
- **Configurable**: Easy configuration via environment variables
- **CORS Support**: Built-in CORS handling for web applications

## Supported Operations

### Bucket Operations
- List buckets (`GET /`)
- Create bucket (`PUT /{bucket}`)
- Delete bucket (`DELETE /{bucket}`)
- List objects (`GET /{bucket}`)

### Object Operations
- Put object (`PUT /{bucket}/{key}`)
- Get object (`GET /{bucket}/{key}`)
- Delete object (`DELETE /{bucket}/{key}`)
- Head object (`HEAD /{bucket}/{key}`)

### Multipart Upload
- Initiate multipart upload
- Upload part
- Complete multipart upload
- Abort multipart upload

## Quick Start

### Using Make (Recommended)

```bash
# Build and run
make build
make run

# Run tests
make test-all
make test-aws

# Setup AWS CLI
make setup

# Interactive test menu
make test-menu

# Quick demo
make demo
```

### Using Go

```bash
# Clone and build
git clone <repository>
cd LocalS3
go mod tidy
go build -o locals3

# Run with default settings
./locals3

# Or run directly
go run main.go
```

### Using Scripts

```bash
# Run specific script
./run-script test
./run-script setup_aws_cli

# Interactive test menu
./test-menu

# List all available scripts
make scripts
```

### Configuration

Configure using environment variables:

```bash
export PORT=3000                    # Server port (default: 3000)
export DATA_DIR=./data             # Data storage directory (default: ./data)
export ACCESS_KEY=test             # Access key (default: test)
export SECRET_KEY=test123456789    # Secret key (default: test123456789)
export REGION=us-east-1            # AWS region (default: us-east-1)
export LOG_LEVEL=info              # Log level (default: info)
export BASE_DOMAIN=localhost       # Base domain (default: localhost)
```

### Example Usage

The server will start on `http://localhost:3000` by default.

#### Using AWS CLI

```bash
# Quick setup (run once)
./setup_aws_cli.sh

# For development, use --no-sign-request flag:
aws --endpoint-url=http://localhost:3000 s3 mb s3://mybucket --no-sign-request
aws --endpoint-url=http://localhost:3000 s3 cp file.txt s3://mybucket/ --no-sign-request
aws --endpoint-url=http://localhost:3000 s3 ls s3://mybucket/ --no-sign-request

# Or use the convenient aliases (after running setup):
s3local mb s3://mybucket
s3local cp file.txt s3://mybucket/
s3local ls s3://mybucket/
s3local sync ./folder s3://mybucket/folder/

# Advanced operations:
s3localapi head-object --bucket mybucket --key file.txt
s3localapi list-objects-v2 --bucket mybucket
```

#### Using curl

```bash
# List buckets
curl http://localhost:3000/

# Create bucket
curl -X PUT http://localhost:3000/mybucket

# Upload object
curl -X PUT -T file.txt http://localhost:3000/mybucket/file.txt

# Download object
curl http://localhost:3000/mybucket/file.txt

# Delete object
curl -X DELETE http://localhost:3000/mybucket/file.txt
```

## Directory Structure

```
LocalS3/
├── main.go                 # Application entry point
├── go.mod                  # Go module file
├── internal/
│   ├── config/            # Configuration management
│   ├── auth/              # AWS V4 authentication
│   ├── storage/           # Storage backend implementation
│   └── handlers/          # HTTP request handlers
└── data/                  # Default data directory (created automatically)
```

## Authentication

LocalS3 supports AWS Signature Version 4 authentication, making it compatible with AWS SDKs and tools. For development purposes, authentication can be disabled, but it's recommended to use proper credentials in production-like environments.

Default credentials:
- Access Key: `test`
- Secret Key: `test123456789`

## Storage

Objects are stored in the local file system under the configured data directory. The structure follows:

```
data/
├── bucket1/
│   ├── object1
│   ├── object2
│   └── subfolder/
│       └── object3
└── bucket2/
    └── object4
```

Metadata is stored alongside objects in `.metadata` files.

## Health Check

The server provides a health check endpoint:

```bash
curl http://localhost:3000/health
```

## CORS Support

The server includes built-in CORS support for web applications, allowing cross-origin requests from browsers.

## Limitations

- No versioning support
- No server-side encryption
- No access control lists (ACLs)
- Simplified multipart upload implementation
- No lifecycle policies
- No replication

## Development

```bash
# Install dependencies
go mod tidy

# Run tests
go test ./...

# Build
go build -o locals3

# Run with debug logging
LOG_LEVEL=debug go run main.go
```

## Testing

LocalS3 includes comprehensive test scripts to verify functionality:

### Interactive Test Menu
```bash
# Run the interactive test menu
./test_menu.sh
```

### Basic Testing
```bash
# Test basic HTTP API
./test.sh

# Test with sample data (similar to your original request)
./test_sample.sh

# Test direct HTTP calls
./test_direct_http.sh
```

### AWS CLI Testing
```bash
# Quick AWS CLI test
./test_aws_quick.sh

# Comprehensive AWS CLI test suite
./test_aws_cli_complete.sh

# Set up AWS CLI for easy usage
./setup_aws_cli.sh
```

### Available Test Scripts
- `test_menu.sh` - Interactive test menu with all options
- `test.sh` - Basic functionality test using curl
- `test_sample.sh` - Tests with JSON data similar to your sample
- `test_direct_http.sh` - Direct HTTP API testing
- `test_aws_quick.sh` - Quick AWS CLI verification
- `test_aws_cli_complete.sh` - Full AWS CLI compatibility test
- `setup_aws_cli.sh` - Sets up AWS CLI with convenient aliases
```

## License

This project is open source and available under the MIT License.
