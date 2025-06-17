#!/bin/bash

# Check if AWS credentials are set
if [ -z "$AWS_ACCESS_KEY_ID" ] || [ -z "$AWS_SECRET_ACCESS_KEY" ]; then
    echo "Error: AWS credentials not set!"
    echo "Please export AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY"
    exit 1
fi

# Set default MCP config if not provided
if [ -z "$CLAUDE_MCP_CONFIG" ]; then
  CLAUDE_MCP_CONFIG='{"mcpServers":{"gamecode":{"command":"/usr/local/bin/gamecode-mcp2","args":["--tools-file","/app/mcp/tools.yaml"],"type":"stdio"}}}'
fi

docker run -p 8080:8080 \
  -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
  -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
  -e AWS_REGION=${AWS_REGION:-us-west-2} \
  -e CLAUDE_CODE_USE_BEDROCK=1 \
  -e ANTHROPIC_MODEL=${ANTHROPIC_MODEL:-us.anthropic.claude-sonnet-4-20250514-v1:0} \
  -e ANTHROPIC_SMALL_FAST_MODEL=${ANTHROPIC_SMALL_FAST_MODEL:-anthropic.claude-3-5-haiku-20241022-v1:0} \
  -e CLAUDE_ALLOWED_TOOLS=${CLAUDE_ALLOWED_TOOLS:-} \
  -e CLAUDE_DISALLOWED_TOOLS="${CLAUDE_DISALLOWED_TOOLS}" \
  -e CLAUDE_MCP_CONFIG="$CLAUDE_MCP_CONFIG" \
  -e LOG_LEVEL=${LOG_LEVEL:-debug} \
  claude-web-go
