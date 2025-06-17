#!/bin/bash

# Test Claude directly in the container
echo "Testing Claude CLI directly in the container..."

docker run --rm \
  -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
  -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
  -e AWS_REGION=${AWS_REGION:-us-west-2} \
  -e CLAUDE_CODE_USE_BEDROCK=1 \
  -e ANTHROPIC_MODEL=${ANTHROPIC_MODEL:-us.anthropic.claude-sonnet-4-20250514-v1:0} \
  claude-web-go \
  /bin/bash -c "claude --version && claude --model \$ANTHROPIC_MODEL -p 'Say hello'"