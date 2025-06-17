#!/bin/bash
set -e

echo "Building Docker image..."
docker build -t claude-web-go .

echo "Build complete! To run:"
echo ""
echo "docker run -p 8080:8080 \\"
echo "  -e AWS_ACCESS_KEY_ID=\$AWS_ACCESS_KEY_ID \\"
echo "  -e AWS_SECRET_ACCESS_KEY=\$AWS_SECRET_ACCESS_KEY \\"
echo "  -e AWS_REGION=\${AWS_REGION:-us-west-2} \\"
echo "  -e CLAUDE_CODE_USE_BEDROCK=1 \\"
echo "  -e ANTHROPIC_MODEL=\${ANTHROPIC_MODEL:-us.anthropic.claude-sonnet-4-20250514-v1:0} \\"
echo "  -e ANTHROPIC_SMALL_FAST_MODEL=\${ANTHROPIC_SMALL_FAST_MODEL:-us.anthropic.claude-3-5-haiku-20241022-v1:0} \\"
echo "  claude-web-go"
