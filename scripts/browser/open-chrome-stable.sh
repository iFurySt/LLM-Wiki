#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
PROFILE_DIR="${ROOT_DIR}/.llmwiki/chrome-profile"
PORT="${CHROME_REMOTE_DEBUG_PORT:-9222}"

mkdir -p "${PROFILE_DIR}"

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "This helper currently supports macOS only. Launch Chrome stable manually with:"
  echo "  --user-data-dir=\"${PROFILE_DIR}\" --remote-debugging-port=${PORT}"
  exit 1
fi

open -na "Google Chrome" --args \
  --user-data-dir="${PROFILE_DIR}" \
  --remote-debugging-port="${PORT}"
