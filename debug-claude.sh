#!/bin/bash

echo "=== Debug Claude CLI ==="
echo "Running minimal Claude test in container..."

# Test 1: Basic version check
echo -e "\n1. Testing claude --version:"
docker run --rm \
  -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
  -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
  -e AWS_REGION=${AWS_REGION:-us-west-2} \
  -e CLAUDE_CODE_USE_BEDROCK=1 \
  claude-web-go \
  claude --version

# Test 2: Simple prompt without model specification
echo -e "\n2. Testing without model specification:"
docker run --rm \
  -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
  -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
  -e AWS_REGION=${AWS_REGION:-us-west-2} \
  -e CLAUDE_CODE_USE_BEDROCK=1 \
  claude-web-go \
  timeout 10 claude -p "Hi"

# Test 3: With model but no other flags
echo -e "\n3. Testing with just model:"
docker run --rm \
  -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
  -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
  -e AWS_REGION=${AWS_REGION:-us-west-2} \
  -e CLAUDE_CODE_USE_BEDROCK=1 \
  -e ANTHROPIC_MODEL=${ANTHROPIC_MODEL:-us.anthropic.claude-sonnet-4-20250514-v1:0} \
  claude-web-go \
  timeout 10 claude --model "\$ANTHROPIC_MODEL" -p "Hi"

# Test 4: Try with --no-stream
echo -e "\n4. Testing with --no-stream:"
docker run --rm \
  -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
  -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
  -e AWS_REGION=${AWS_REGION:-us-west-2} \
  -e CLAUDE_CODE_USE_BEDROCK=1 \
  -e ANTHROPIC_MODEL=${ANTHROPIC_MODEL:-us.anthropic.claude-sonnet-4-20250514-v1:0} \
  claude-web-go \
  timeout 10 claude --no-stream --model "\$ANTHROPIC_MODEL" -p "Hi"