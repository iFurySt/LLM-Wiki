# Agent Install

This document is the durable installation reference for LLM-Wiki.

It is written primarily for AI agents, but should also be readable by humans operating the service.

For any hosted install surfaces such as `/install/LLM-Wiki.md` and `/install/install-cli.sh`, the intended public host should come from `LLM_WIKI_INSTALL_BASE_URL`.
If that env var is unset, the service falls back to `LLM_WIKI_CLI_BASE_URL`.

## Distribution Channels

LLM-Wiki currently ships through four channels:

- GitHub Releases for cross-platform CLI binaries and hosted install assets
- Docker Hub and GHCR for the main `llm-wiki-server` image
- npm for the `@ifuryst/llm-wiki-mcp` stdio bridge package
- hosted install docs and skill archives served by a running LLM-Wiki instance

## CLI Install

The standard CLI installer downloads binaries from GitHub Releases:

```sh
curl -fsSL https://your-llm-wiki-host/install/install-cli.sh | sh
```

The installer places:

- `llm-wiki`
- `lw`

into the target install directory.

Useful overrides:

```sh
LLM_WIKI_VERSION=v0.1.0 curl -fsSL https://your-llm-wiki-host/install/install-cli.sh | sh
LLM_WIKI_RELEASE_REPO=iFurySt/LLM-Wiki curl -fsSL https://your-llm-wiki-host/install/install-cli.sh | sh
```

## Docker Install

LLM-Wiki publishes a main-service image only. Users are expected to provide PostgreSQL and Redis themselves.

Published images:

- `docker.io/ifuryst/llm-wiki`
- `ghcr.io/ifuryst/llm-wiki`

Minimal example:

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
  -e LLM_WIKI_REDIS_ADDR=host.docker.internal:6379 \
  docker.io/ifuryst/llm-wiki:latest
```

## Remote MCP

Preferred endpoint:

```text
https://your-llm-wiki-host/mcp
```

Legacy compatibility:

```text
https://your-llm-wiki-host/sse
```

Example config:

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

## npm Stdio MCP

Published package:

```sh
LLM_WIKI_TOKEN=<llm-wiki-token> npx -y @ifuryst/llm-wiki-mcp --base-url https://your-llm-wiki-host
```

In-repo package before publish:

```sh
npm install --prefix npm/llm-wiki-mcp --package-lock=false
LLM_WIKI_TOKEN=<llm-wiki-token> npx --prefix npm/llm-wiki-mcp llm-wiki-mcp --base-url https://your-llm-wiki-host
```

Example config:

```json
{
  "llm-wiki": {
    "command": "npx",
    "args": [
      "-y",
      "@ifuryst/llm-wiki-mcp",
      "--base-url",
      "https://your-llm-wiki-host"
    ],
    "env": {
      "LLM_WIKI_TOKEN": "<llm-wiki-token>"
    }
  }
}
```

## CLI Authentication

Preferred human flow:

```sh
lw auth login --base-url https://your-llm-wiki-host
lw auth status --base-url https://your-llm-wiki-host
lw namespace list --base-url https://your-llm-wiki-host
```

On a fresh instance, open `https://your-llm-wiki-host/setup` first to create the initial admin account and default tenant. After setup, the web admin console lives at `/admin/login` and `/admin/users`.

Explicit token flow:

```sh
lw namespace list --base-url https://your-llm-wiki-host --token <llm-wiki-token>
lw namespace list --base-url https://your-llm-wiki-host --token-file /var/run/secrets/llm-wiki/token
```

The CLI resolves credentials in this order:

1. explicit flags
2. environment variables such as `LLM_WIKI_TOKEN`
3. the active profile in `~/.llm-wiki/config.json`

Base URL and tenant resolve the same way. If tenant is not specified anywhere, the CLI falls back to the stored profile tenant and then `default`.

Device-code login always prints the approval URL and code. On a local machine it also tries to open the browser unless `--no-open` is passed.

## Hosted Skill Docs

Hosted guide:

- `https://your-llm-wiki-host/install/LLM-Wiki.md`

Hosted skill downloads:

- `https://your-llm-wiki-host/install/skills/LLM-Wiki.skill`
- `https://your-llm-wiki-host/install/skills/LLM-Wiki.zip`

## Guidance For AI Agents

If an agent can read markdown instructions from a URL, point it to:

```text
Read and follow https://your-llm-wiki-host/install/LLM-Wiki.md
```

If an agent is terminal-native and already has the CLI:

```sh
lw auth login --base-url https://your-llm-wiki-host
lw namespace list --base-url https://your-llm-wiki-host
lw document list --base-url https://your-llm-wiki-host
```
