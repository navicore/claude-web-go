FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o claude-web-server ./cmd/server

FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://repo.anthropic.com/deb/anthropic.gpg | apt-key add - && \
    echo "deb https://repo.anthropic.com/deb/ stable main" > /etc/apt/sources.list.d/anthropic.list && \
    apt-get update && \
    apt-get install -y claude && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/claude-web-server .
COPY web ./web

EXPOSE 8080

CMD ["./claude-web-server"]