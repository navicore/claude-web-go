# Claude Web Go - Development Context

This document contains important context for Claude or other AI assistants working on this project.

## Project Overview

Claude Web Go is a web-based chat interface for interacting with Claude via AWS Bedrock. It wraps the Claude CLI tool, executing each interaction as a one-shot command while maintaining conversation context in the browser's local storage.

## Key Technical Decisions

### One-Shot Execution Model
- Each user message triggers a new `claude -p "prompt"` command
- No persistent Claude sessions - context is rebuilt for each interaction
- Context window is managed client-side to control token usage

### AWS Bedrock Integration
- Uses `CLAUDE_CODE_USE_BEDROCK=1` environment variable (must be "1", not "true")
- Automatically generates AWS session tokens from base credentials
- Session tokens are cached and refreshed as needed (12-hour validity)

### File Handling
- Claude creates files in `/tmp/<session-uuid>/` directories
- Files are automatically detected after command execution
- Temporary storage with 30-minute cleanup cycle
- Images (PNG/SVG) are displayed inline, other files get download links

## Current Issues and Workarounds

### AWS Bedrock 429 Errors
- Some AWS accounts return "429 Too many tokens" even for tiny requests
- This appears to be an AWS account/quota issue, not actual rate limiting
- The code is working correctly - confirmed by proper error messages when accessing regions without model access

### Model ID Format
- Use full Bedrock model IDs: `us.anthropic.claude-sonnet-4-20250514-v1:0`
- The "us." prefix is part of the model ID, not a region indicator
- Default model is Sonnet 4, with Haiku as the small/fast model option

## Code Structure

```
claude-web-go/
├── cmd/server/         # Main server entry point
├── internal/
│   ├── api/           # HTTP handlers and WebSocket
│   ├── auth/          # AWS authentication and session management
│   ├── claude/        # Claude CLI execution wrapper
│   ├── logger/        # Structured logging with logrus
│   ├── models/        # Data structures
│   └── storage/       # File management
├── web/               # Frontend (HTML/JS/CSS)
├── docker/            # Docker configuration
└── k8s/              # Kubernetes manifests
```

## Key Components

### Claude Executor (`internal/claude/executor.go`)
- Handles Claude CLI command execution
- Manages temporary directories for file output
- Implements timeout and error handling
- Builds prompts with conversation context

### Session Manager (`internal/auth/session.go`)
- Generates AWS STS session tokens from base credentials
- Caches tokens and refreshes before expiration
- Thread-safe token management

### Frontend (`web/app.js`)
- Manages conversation history in localStorage
- Implements rolling context window
- Handles file rendering and downloads
- Real-time message updates

## Testing and Debugging

### Environment Setup
```bash
export AWS_ACCESS_KEY_ID="your-key"
export AWS_SECRET_ACCESS_KEY="your-secret"
export AWS_REGION="us-west-2"
export CLAUDE_CODE_USE_BEDROCK="1"
export ANTHROPIC_MODEL="us.anthropic.claude-sonnet-4-20250514-v1:0"
export LOG_LEVEL="debug"
```

### Common Debug Commands
```bash
# Test Claude CLI directly
claude --model "$ANTHROPIC_MODEL" -p "Hello"

# Build and run with debug logging
./docker-build.sh
LOG_LEVEL=debug ./docker-run.sh

# Check AWS credentials
aws sts get-caller-identity

# List available Bedrock models
aws bedrock list-foundation-models --region $AWS_REGION
```

## Future Enhancements

1. **Context Compression**: Implement intelligent context summarization for longer conversations
2. **Multi-Session Support**: Allow multiple concurrent chat sessions
3. **File Persistence**: Optional S3 backend for permanent file storage
4. **Streaming Responses**: Implement SSE for real-time Claude output streaming
5. **Authentication**: Add user authentication for multi-tenant deployments

## Recent Updates

### MCP Integration (gamecode-mcp2)
- Added multi-stage Docker build to include gamecode-mcp2 Rust MCP server
- MCP configuration passed inline via CLAUDE_MCP_CONFIG environment variable
- Default tools.yaml created at /app/mcp/tools.yaml in container
- Example MCP config format:
```json
{
  "mcpServers": {
    "gamecode": {
      "command": "/usr/local/bin/gamecode-mcp2",
      "args": ["--tools-file", "/app/mcp/tools.yaml"],
      "type": "stdio"
    }
  }
}
```

### PlantUML Support
- PlantUML added to Docker image for diagram generation
- Requires Java runtime (included in Docker image)
- Claude can use PlantUML via CLAUDE_ALLOWED_TOOLS environment variable

### Disallowed Tools Support
- Added --disallowedTools parameter support via CLAUDE_DISALLOWED_TOOLS env var
- Default disallowed tools prevent Claude from attempting filesystem operations
- Default list: "Bash,Glob,Grep,LS,Read,Edit,MultiEdit,Write,NotebookRead,NotebookEdit,WebFetch,TodoRead,TodoWrite,Task"
- This helps Claude understand upfront which tools it cannot use, avoiding wasted attempts

## Important Notes

- Always use `CLAUDE_CODE_USE_BEDROCK=1` (not "true") for Bedrock mode
- The 429 errors are often AWS account issues, not code problems
- File cleanup happens automatically - don't rely on files persisting
- Context window size directly impacts token usage and costs
- Claude CLI must be installed in the Docker image via npm
- Model IDs differ between cross-region and region-specific models (watch the "us." prefix)