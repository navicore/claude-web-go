#!/bin/bash
set -e

# Log current time before sync
echo "Container time before sync: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"

# Try to sync time if possible (may fail in some Docker environments)
ntpdate -s time.google.com || true

# Log time after sync attempt
echo "Container time after sync: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"

# Start the application
exec ./claude-web-server "$@"