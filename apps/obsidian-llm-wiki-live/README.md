# LLM-Wiki Live Obsidian Plugin

This plugin continuously mirrors the current LLM-Wiki `ns` into the active Obsidian vault.

Current behavior:

- reads base URL and bearer token from `~/.llm-wiki/config.json`
- resolves the active `ns` from `GET /v1/auth/whoami` instead of trusting a local config field
- mirrors the current `ns` into `LLM-Wiki/<ns>/...` inside the vault
- writes folders as folders and documents as markdown files
- overwrites mirrored files when remote content changes
- syncs once on plugin load and then on a configurable interval
- exposes a manual command to sync immediately

Current scope:

- read-only mirror into vault files
- service-first source of truth
- no local editing back to LLM-Wiki yet
