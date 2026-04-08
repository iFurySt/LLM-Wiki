#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SRC="$ROOT/apps/obsidian-llm-wiki-live"
VAULT_PATH="${1:-$HOME/Documents/Obsidian Vault}"
PLUGIN_ID="llm-wiki-live"
DEST="$VAULT_PATH/.obsidian/plugins/$PLUGIN_ID"

mkdir -p "$DEST"
cp "$SRC/manifest.json" "$DEST/manifest.json"
cp "$SRC/main.js" "$DEST/main.js"
cp "$SRC/styles.css" "$DEST/styles.css"

if [ ! -f "$DEST/data.json" ]; then
  cat >"$DEST/data.json" <<'JSON'
{
  "autoRefreshSeconds": 15,
  "lastLoadedAt": "",
  "lastMirrorAt": ""
}
JSON
fi

COMMUNITY_PLUGINS="$VAULT_PATH/.obsidian/community-plugins.json"
if [ -f "$COMMUNITY_PLUGINS" ]; then
  node -e '
const fs = require("fs");
const path = process.argv[1];
const pluginId = process.argv[2];
const raw = fs.readFileSync(path, "utf8");
const parsed = JSON.parse(raw);
const list = Array.isArray(parsed) ? parsed : [];
if (!list.includes(pluginId)) list.push(pluginId);
fs.writeFileSync(path, JSON.stringify(list, null, 2));
' "$COMMUNITY_PLUGINS" "$PLUGIN_ID"
else
  cat >"$COMMUNITY_PLUGINS" <<JSON
[
  "$PLUGIN_ID"
]
JSON
fi

echo "Installed $PLUGIN_ID into $DEST"
