FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o claude-web-server ./cmd/server

FROM ubuntu:22.04

# Install dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    curl \
    unzip \
    && rm -rf /var/lib/apt/lists/*

# Install AWS CLI for debugging
RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" && \
    unzip awscliv2.zip && \
    ./aws/install && \
    rm -rf awscliv2.zip aws/

# Install Node.js 18+
RUN curl -fsSL https://deb.nodesource.com/setup_18.x | bash - && \
    apt-get install -y nodejs && \
    rm -rf /var/lib/apt/lists/*

# Install Claude Code CLI
RUN npm install -g @anthropic-ai/claude-code

# Test claude installation
RUN which claude && claude --version || echo "Claude installation check failed"

WORKDIR /app

COPY --from=builder /app/claude-web-server .
COPY web ./web

# Create tmp directory for Claude output files
RUN mkdir -p /tmp && chmod 777 /tmp

# Create a home directory for the app
RUN mkdir -p /home/app && chmod 755 /home/app
ENV HOME=/home/app

EXPOSE 8080

CMD ["./claude-web-server"]