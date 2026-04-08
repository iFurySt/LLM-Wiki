# Agent Install

This is the main install and connection reference.

## CLI

Install:

```sh
curl -fsSL https://your-llm-wiki-host/install/install-cli.sh | sh
```

Check:

```sh
lw version
lw auth login --base-url https://your-llm-wiki-host
lw auth whoami --base-url https://your-llm-wiki-host
lw ns list --base-url https://your-llm-wiki-host
lw folder list --base-url https://your-llm-wiki-host
lw document create text --base-url https://your-llm-wiki-host --folder-id 1 --title "Hello" --content "First note"
```

Notes:

- first boot is `https://your-llm-wiki-host/setup`
- CLI state lives in `~/.llm-wiki/config.json`
- `lw auth login --ns <target>` chooses the login target `ns`; other commands use the token's bound `ns`
- switch between accessible spaces with `lw auth switch <ns>`
- manage invites with `lw ns invite list|create|accept`
- hosted install links should resolve from `LLM_WIKI_INSTALL_BASE_URL`

## Docker

LLM-Wiki publishes one main service image. PostgreSQL is the external dependency.

```sh
docker run --rm -p 8234:8234 \
  -e LLM_WIKI_SERVER_HOST=0.0.0.0 \
  -e LLM_WIKI_SERVER_PORT=8234 \
  -e LLM_WIKI_POSTGRES_HOST=host.docker.internal \
  -e LLM_WIKI_POSTGRES_PORT=5432 \
  -e LLM_WIKI_POSTGRES_USER=llmwiki \
  -e LLM_WIKI_POSTGRES_PASSWORD=llmwiki \
  -e LLM_WIKI_POSTGRES_DATABASE=llmwiki \
  -e LLM_WIKI_POSTGRES_SSLMODE=disable \
  docker.io/ifuryst/llm-wiki:latest
```

Published images:

- `docker.io/ifuryst/llm-wiki`
- `ghcr.io/ifuryst/llm-wiki`

## MCP

Preferred endpoint:

```text
https://your-llm-wiki-host/mcp
```

Legacy endpoint:

```text
https://your-llm-wiki-host/sse
```

Example:

```json
{
  "llm-wiki": {
    "type": "http",
    "url": "https://your-llm-wiki-host/mcp",
    "headers": {
      "Authorization": "Bearer <llm-wiki-token>"
    }
  }
}
```

## npx MCP

```sh
LLM_WIKI_TOKEN=<llm-wiki-token> npx -y @ifuryst/llm-wiki-mcp --base-url https://your-llm-wiki-host
```

## Hosted Agent Guide

Point agents here:

```text
https://your-llm-wiki-host/install/LLM-Wiki.md
```

Hosted downloads:

- `/install/skills/LLM-Wiki.skill`
- `/install/skills/LLM-Wiki.zip`
