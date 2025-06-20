#!/bin/bash

# Set mock AWS credentials for local testing
# Replace these with real credentials when running in production
export AWS_ACCESS_KEY_ID="mock-access-key"
export AWS_SECRET_ACCESS_KEY="mock-secret-key"
export AWS_REGION="us-east-1"
export CLAUDE_CODE_USE_BEDROCK="1"
export ANTHROPIC_MODEL="us.anthropic.claude-sonnet-4-20250514-v1:0"
export ANTHROPIC_SMALL_FAST_MODEL="anthropic.claude-3-5-haiku-20241022-v1:0"
export LOG_LEVEL="${LOG_LEVEL:-warn}"

echo "Starting Claude Web server with mock AWS credentials..."
echo "Note: Claude CLI commands will fail without real AWS Bedrock credentials"
echo "Server will be available at http://localhost:8080"

go run cmd/server/main.go
