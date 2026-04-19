#!/bin/sh

set -eu

if [ ! -f /app/config.json ]; then
  echo "Missing /app/config.json." >&2
  echo "Mount a runtime config file, for example:" >&2
  echo "  docker run --rm -p 8080:8080 \\" >&2
  echo "    -e ROBOCALL_SESSION_SECRET=change-me \\" >&2
  echo "    -v \$(pwd)/config.json:/app/config.json:ro \\" >&2
  echo "    robocall" >&2
  exit 1
fi

exec /app/robocall "$@"
