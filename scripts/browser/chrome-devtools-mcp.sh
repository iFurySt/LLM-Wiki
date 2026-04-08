#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
PROFILE_DIR="${ROOT_DIR}/.llmwiki/chrome-profile"
PORT="${CHROME_REMOTE_DEBUG_PORT:-9222}"
BROWSER_URL="http://127.0.0.1:${PORT}"

if ! command -v chrome-devtools-mcp >/dev/null 2>&1; then
  echo "chrome-devtools-mcp is not installed. Run: npm i chrome-devtools-mcp@latest -g" >&2
  exit 1
fi

mkdir -p "${PROFILE_DIR}"

if curl -fsS "${BROWSER_URL}/json/version" >/dev/null 2>&1; then
  exec chrome-devtools-mcp --browserUrl "${BROWSER_URL}" "$@"
fi

exec chrome-devtools-mcp \
  --channel stable \
  --userDataDir "${PROFILE_DIR}" \
  "$@"
