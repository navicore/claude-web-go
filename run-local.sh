#!/bin/bash

# Set mock AWS credentials for local testing
# Replace these with real credentials when running in production
export AWS_ACCESS_KEY_ID="mock-access-key"
export AWS_SECRET_ACCESS_KEY="mock-secret-key"
export AWS_REGION="us-east-1"
export CLAUDE_CODE_USE_BEDROCK="true"
export ANTHROPIC_MODEL="claude-3-5-sonnet-20241022"

echo "Starting Claude Web server with mock AWS credentials..."
echo "Note: Claude CLI commands will fail without real AWS Bedrock credentials"
echo "Server will be available at http://localhost:8080"

go run cmd/server/main.go