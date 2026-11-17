#!/bin/sh
# kitchen-worker has no web port; it's healthy if it wrote a Redis
# heartbeat in the last 20 seconds (it writes one every 10s).
set -eu
HOST="${REDIS_HOST:-redis}"
PORT="${REDIS_PORT:-6379}"
LAST=$(redis-cli -h "$HOST" -p "$PORT" GET kitchen-worker:heartbeat 2>/dev/null || echo 0)
NOW=$(($(date +%s%N) / 1000000))
AGE=$(( (NOW - LAST) / 1000 ))
if [ "$AGE" -lt 20 ]; then
  exit 0
else
  exit 1
fi
