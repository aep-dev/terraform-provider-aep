#!/usr/bin/env bash
set -euo pipefail

AEPBASE_PORT="${AEPBASE_PORT:-8081}"
AEPBASE_URL="http://localhost:${AEPBASE_PORT}"
AEPBASE_DIR="${AEPBASE_DIR:-/tmp/aepbase}"

# Clone aepbase if not already present
if [ ! -d "${AEPBASE_DIR}" ]; then
  echo "Cloning aepbase..."
  git clone https://github.com/rambleraptor/aepbase.git "${AEPBASE_DIR}"
fi

# Build aepbase
echo "Building aepbase..."
(cd "${AEPBASE_DIR}" && go build -o aepbase ./)

# Start aepbase
echo "Starting aepbase on port ${AEPBASE_PORT}..."
"${AEPBASE_DIR}/aepbase" -db ":memory:" -port "${AEPBASE_PORT}" &
SERVER_PID=$!

# Wait for server to be ready
echo "Waiting for aepbase to be ready..."
for i in $(seq 1 30); do
  if curl -sf "${AEPBASE_URL}/openapi.json" > /dev/null 2>&1; then
    echo "aepbase is ready (PID ${SERVER_PID}) at ${AEPBASE_URL}"
    echo "${SERVER_PID}" > /tmp/aepbase.pid
    exit 0
  fi
  if ! kill -0 "$SERVER_PID" 2>/dev/null; then
    echo "aepbase process exited unexpectedly."
    exit 1
  fi
  sleep 1
done

echo "aepbase failed to start within 30 seconds."
kill "$SERVER_PID" 2>/dev/null || true
exit 1
