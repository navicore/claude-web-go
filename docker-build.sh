#!/bin/bash

echo "Building Docker image..."
docker build -t claude-web-go .

echo "Build complete! To run:"
echo ""
echo "docker run -p 8080:8080 \\"
echo "  -e AWS_ACCESS_KEY_ID=\$AWS_ACCESS_KEY_ID \\"
echo "  -e AWS_SECRET_ACCESS_KEY=\$AWS_SECRET_ACCESS_KEY \\"
echo "  -e AWS_REGION=\${AWS_REGION:-us-east-1} \\"
echo "  -e CLAUDE_CODE_USE_BEDROCK=true \\"
echo "  -e ANTHROPIC_MODEL=\${ANTHROPIC_MODEL:-claude-3-5-sonnet-20241022} \\"
echo "  claude-web-go"