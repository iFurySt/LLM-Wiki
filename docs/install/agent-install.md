# Agent Install

This document is the durable installation reference for DocMesh.

It is written primarily for AI agents, but should also be readable by humans operating the service.

## Distribution Channels

DocMesh currently ships through four channels:

- GitHub Releases for cross-platform CLI binaries and hosted install assets
- Docker Hub and GHCR for the main `docmesh-server` image
- npm for the `docmesh-mcp` stdio bridge package
- hosted install docs and skill archives served by a running DocMesh instance

## CLI Install

The standard CLI installer downloads binaries from GitHub Releases:

```sh
curl -fsSL https://docmesh.amoylab.com/install/install-cli.sh | sh
```

The installer places:

- `docmesh`
- `dm`

into the target install directory.

Useful overrides:

```sh
DOCMESH_VERSION=v0.1.0 curl -fsSL https://docmesh.amoylab.com/install/install-cli.sh | sh
DOCMESH_RELEASE_REPO=iFurySt/DocMesh curl -fsSL https://docmesh.amoylab.com/install/install-cli.sh | sh
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
https://docmesh.amoylab.com/mcp
```

Legacy compatibility:

```text
https://docmesh.amoylab.com/sse
```

Example config:

```json
{
  "docmesh": {
    "type": "http",
    "url": "https://docmesh.amoylab.com/mcp",
    "headers": {
      "X-DocMesh-Tenant-ID": "default"
    }
  }
}
```

## npm Stdio MCP

Published package:

```sh
npx -y docmesh-mcp --base-url https://docmesh.amoylab.com --tenant default
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
      "https://docmesh.amoylab.com",
      "--tenant",
      "default"
    ]
  }
}
```

## Hosted Skill Docs

Hosted guide:

- `https://docmesh.amoylab.com/install/DocMesh.md`

Hosted skill downloads:

- `https://docmesh.amoylab.com/install/skills/DocMesh.skill`
- `https://docmesh.amoylab.com/install/skills/DocMesh.zip`

## Guidance For AI Agents

If an agent can read markdown instructions from a URL, point it to:

```text
Read and follow https://docmesh.amoylab.com/install/DocMesh.md
```

If an agent is terminal-native and already has the CLI:

```sh
dm system info --base-url https://docmesh.amoylab.com --tenant default
dm namespace list --base-url https://docmesh.amoylab.com --tenant default
dm document list --base-url https://docmesh.amoylab.com --tenant default
```
