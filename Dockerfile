FROM golang:1.21-alpine AS go-builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o claude-web-server ./cmd/server

# Rust builder stage for gamecode-mcp2
FROM rust:1.87 AS rust-builder

WORKDIR /build

# Install gamecode-mcp2 from crates.io
RUN cargo install gamecode-mcp2 || { echo "Failed to install gamecode-mcp2"; exit 1; }

# Verify installation
RUN ls -la /usr/local/cargo/bin/ && \
    test -f /usr/local/cargo/bin/gamecode-mcp2 || { echo "gamecode-mcp2 not found after installation"; exit 1; }

# Final stage
FROM ubuntu:22.04

# Install dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    curl \
    unzip \
    default-jre \
    graphviz \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

# Set timezone to UTC
ENV TZ=UTC
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Install ntpdate for time sync
RUN apt-get update && apt-get install -y ntpdate && rm -rf /var/lib/apt/lists/*

# Install PlantUML
RUN curl -L https://github.com/plantuml/plantuml/releases/download/v1.2024.3/plantuml-1.2024.3.jar -o /usr/local/bin/plantuml.jar && \
    echo '#!/bin/bash\njava -jar /usr/local/bin/plantuml.jar "$@"' > /usr/local/bin/plantuml && \
    chmod +x /usr/local/bin/plantuml

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

COPY --from=go-builder /app/claude-web-server .
COPY web ./web
COPY docker-entrypoint.sh .

# Copy gamecode-mcp2 from rust builder
COPY --from=rust-builder /usr/local/cargo/bin/gamecode-mcp2 /usr/local/bin/gamecode-mcp2
RUN chmod +x /usr/local/bin/gamecode-mcp2 && \
    ls -la /usr/local/bin/gamecode-mcp2 && \
    which gamecode-mcp2 && \
    gamecode-mcp2 --version || { echo "gamecode-mcp2 installation check failed"; exit 1; }

# Create tmp directory for Claude output files
RUN mkdir -p /tmp && chmod 777 /tmp

# Create a home directory for the app
RUN mkdir -p /home/app && chmod 755 /home/app
ENV HOME=/home/app

# Create directory for MCP configuration
RUN mkdir -p /app/mcp && chmod 755 /app/mcp

# Copy tools.yaml file
COPY tools.yaml /app/mcp/tools.yaml
RUN echo "Tools file created:" && cat /app/mcp/tools.yaml

EXPOSE 8080

ENTRYPOINT ["./docker-entrypoint.sh"]