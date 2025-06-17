#!/bin/bash

# This script gets AWS credentials from your local AWS CLI configuration
# and outputs them as docker run environment variables

echo "Getting AWS credentials from AWS CLI..."

# Get credentials using AWS CLI
CREDS=$(aws configure export-credentials 2>/dev/null)

if [ $? -ne 0 ]; then
    echo "Error: Failed to get AWS credentials. Make sure you're logged in with AWS CLI."
    exit 1
fi

# Parse the JSON output
AWS_ACCESS_KEY_ID=$(echo $CREDS | jq -r '.AccessKeyId')
AWS_SECRET_ACCESS_KEY=$(echo $CREDS | jq -r '.SecretAccessKey')
AWS_SESSION_TOKEN=$(echo $CREDS | jq -r '.SessionToken // empty')
AWS_REGION=$(aws configure get region || echo "us-east-1")

# Output the docker run command with credentials
echo ""
echo "Run the following command:"
echo ""
echo "docker run -p 8080:8080 \\"
echo "  -e AWS_ACCESS_KEY_ID=\"$AWS_ACCESS_KEY_ID\" \\"
echo "  -e AWS_SECRET_ACCESS_KEY=\"$AWS_SECRET_ACCESS_KEY\" \\"
if [ -n "$AWS_SESSION_TOKEN" ]; then
    echo "  -e AWS_SESSION_TOKEN=\"$AWS_SESSION_TOKEN\" \\"
fi
echo "  -e AWS_REGION=\"$AWS_REGION\" \\"
echo "  -e CLAUDE_CODE_USE_BEDROCK=true \\"
echo "  -e ANTHROPIC_MODEL=\"us.anthropic.claude-3-5-sonnet-20241022-v2:0\" \\"
echo "  -e ANTHROPIC_SMALL_FAST_MODEL=\"us.anthropic.claude-3-5-haiku-20241022-v1:0\" \\"
echo "  claude-web-go"