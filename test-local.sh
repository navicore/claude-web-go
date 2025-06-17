#!/bin/bash

# Test script to verify the app builds and starts correctly
# This will fail at runtime without real AWS credentials, but will verify the build

echo "Building Go application..."
go build -o claude-web-server ./cmd/server

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
    echo ""
    echo "To run locally with real AWS credentials:"
    echo "export AWS_ACCESS_KEY_ID=your-key"
    echo "export AWS_SECRET_ACCESS_KEY=your-secret"
    echo "./claude-web-server"
    echo ""
    echo "To build Docker image when Docker is running:"
    echo "./docker-build.sh"
else
    echo "❌ Build failed!"
    exit 1
fi