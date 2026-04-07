# Agent Install

This document is the durable installation reference for DocMesh.

It is written primarily for AI agents, but should also be readable by humans operating the service.

For any hosted install surfaces such as `/install/DocMesh.md` and `/install/install-cli.sh`, the intended public host should come from `DOCMESH_INSTALL_BASE_URL`.
If that env var is unset, the service falls back to `DOCMESH_CLI_BASE_URL`.

## Distribution Channels

DocMesh currently ships through four channels:

- GitHub Releases for cross-platform CLI binaries and hosted install assets
- Docker Hub and GHCR for the main `docmesh-server` image
- npm for the `docmesh-mcp` stdio bridge package
- hosted install docs and skill archives served by a running DocMesh instance

## CLI Install

The standard CLI installer downloads binaries from GitHub Releases:

```sh
curl -fsSL https://your-docmesh-host/install/install-cli.sh | sh
```

The installer places:

- `docmesh`
- `dm`

into the target install directory.

Useful overrides:

```sh
DOCMESH_VERSION=v0.1.0 curl -fsSL https://your-docmesh-host/install/install-cli.sh | sh
DOCMESH_RELEASE_REPO=iFurySt/DocMesh curl -fsSL https://your-docmesh-host/install/install-cli.sh | sh
```

## Docker Install

DocMesh publishes a main-service image only. Users are expected to provide PostgreSQL and Redis themselves.

Published images:

- `docker.io/ifuryst/docmesh`
- `ghcr.io/ifuryst/docmesh`

Minimal example:

```sh
docker run --rm -p 8234:8234 \
  -e DOCMESH_SERVER_HOST=0.0.0.0 \
  -e DOCMESH_SERVER_PORT=8234 \
  -e DOCMESH_POSTGRES_HOST=host.docker.internal \
  -e DOCMESH_POSTGRES_PORT=5432 \
  -e DOCMESH_POSTGRES_USER=docmesh \
  -e DOCMESH_POSTGRES_PASSWORD=docmesh \
  -e DOCMESH_POSTGRES_DATABASE=docmesh \
  -e DOCMESH_POSTGRES_SSLMODE=disable \
  -e DOCMESH_REDIS_ADDR=host.docker.internal:6379 \
  docker.io/ifuryst/docmesh:latest
```

## Remote MCP

Preferred endpoint:

```text
https://your-docmesh-host/mcp
```

Legacy compatibility:

```text
https://your-docmesh-host/sse
```

Example config:

```json
{
  "docmesh": {
    "type": "http",
    "url": "https://your-docmesh-host/mcp",
    "headers": {
      "X-DocMesh-Tenant-ID": "default"
    }
  }
}
```

## npm Stdio MCP

Published package:

```sh
npx -y docmesh-mcp --base-url https://your-docmesh-host --tenant default
```

Example config:

```json
{
  "docmesh": {
    "command": "npx",
    "args": [
      "-y",
      "docmesh-mcp",
      "--base-url",
      "https://your-docmesh-host",
      "--tenant",
      "default"
    ]
  }
}
```

## Hosted Skill Docs

Hosted guide:

- `https://your-docmesh-host/install/DocMesh.md`

Hosted skill downloads:

- `https://your-docmesh-host/install/skills/DocMesh.skill`
- `https://your-docmesh-host/install/skills/DocMesh.zip`

## Guidance For AI Agents

If an agent can read markdown instructions from a URL, point it to:

```text
Read and follow https://your-docmesh-host/install/DocMesh.md
```

If an agent is terminal-native and already has the CLI:

```sh
dm system info --base-url https://your-docmesh-host --tenant default
dm namespace list --base-url https://your-docmesh-host --tenant default
dm document list --base-url https://your-docmesh-host --tenant default
```
