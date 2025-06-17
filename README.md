# Claude Web Go

A web-based chat application that provides a clean interface for interacting with Claude via AWS Bedrock. Each interaction runs Claude CLI in one-shot mode with context maintained in the browser's local storage.

## Features

- 🗨️ Clean chat interface with markdown and code syntax highlighting
- 🖼️ Automatic file rendering for Claude-generated diagrams (PNG/SVG)
- 💾 Local storage for conversation history with configurable rolling context window
- 🔐 AWS Bedrock authentication using temporary session tokens
- 📁 Automatic file detection and download links for generated content
- 🚀 Docker-ready for easy deployment

## Prerequisites

- Docker (for containerized deployment)
- AWS credentials with Bedrock access
- Access to Claude models in AWS Bedrock

## Quick Start

1. Clone the repository:
```bash
git clone <repository-url>
cd claude-web-go
```

2. Build the Docker image:
```bash
./docker-build.sh
```

3. Set your AWS credentials and run:
```bash
export AWS_ACCESS_KEY_ID="your-key"
export AWS_SECRET_ACCESS_KEY="your-secret"
export AWS_REGION="us-west-2"  # or your preferred region
./docker-run.sh
```

4. Open http://localhost:8080 in your browser

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `AWS_ACCESS_KEY_ID` | Your AWS access key | Required |
| `AWS_SECRET_ACCESS_KEY` | Your AWS secret key | Required |
| `AWS_REGION` | AWS region for Bedrock | us-west-2 |
| `CLAUDE_CODE_USE_BEDROCK` | Enable Bedrock mode | 1 |
| `ANTHROPIC_MODEL` | Claude model to use | us.anthropic.claude-sonnet-4-20250514-v1:0 |
| `CLAUDE_ALLOWED_TOOLS` | Tools Claude can use (e.g., "Task") | "" (empty - no tools) |
| `LOG_LEVEL` | Logging verbosity | info |

## How It Works

1. **One-Shot Execution**: Each message creates a new Claude CLI process with `-p` flag
2. **Context Management**: Previous messages are stored in browser localStorage and included in prompts
3. **Session Directories**: Each interaction creates a `/tmp/<uuid>` directory for Claude's output files
4. **File Detection**: Any files created by Claude are automatically detected and made available for download
5. **AWS Authentication**: The server automatically generates AWS session tokens from your credentials

## Context Window Management

- Messages are stored in browser localStorage
- Configurable context window size (default: 20 messages)
- Only the last N messages are sent to Claude to manage token usage
- Full history remains visible in the UI

## File Handling

When Claude generates files (diagrams, code, etc.):
- Files are created in a temporary session directory
- The web UI automatically detects and displays images
- Download links are provided for all file types
- Files are cleaned up after 30 minutes

## Development

### Local Development (without Docker)

```bash
# Install dependencies
go mod download

# Run with mock credentials (UI testing only)
./run-local.sh
```

### Building from Source

```bash
# Build the Go binary
go build -o claude-web-server ./cmd/server

# Run directly
AWS_ACCESS_KEY_ID=xxx AWS_SECRET_ACCESS_KEY=xxx ./claude-web-server
```

## Troubleshooting

### 429 Too Many Tokens Error
- This often indicates AWS Bedrock quota issues rather than actual token limits
- Verify your AWS account has proper Bedrock model access
- Check your service quotas in the AWS console

### Claude Not Responding
- Ensure `CLAUDE_CODE_USE_BEDROCK=1` is set
- Verify AWS credentials are valid
- Check that your region has access to the specified model

### Debug Mode
Run with `LOG_LEVEL=debug` for detailed logging:
```bash
LOG_LEVEL=debug ./docker-run.sh
```

## Architecture

- **Backend**: Go server that wraps Claude CLI
- **Frontend**: Vanilla JavaScript with local storage
- **Authentication**: AWS STS for temporary credentials
- **File Storage**: Temporary directories with automatic cleanup
- **Context**: Client-side storage with server-side prompt building