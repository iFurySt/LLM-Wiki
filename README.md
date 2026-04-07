# DocMesh

Agent-native knowledge service for multi-tenant document CRUD, revisions, grounding, and shared collaboration over HTTP, CLI, and web, inspired by Karpathy's LLM Wiki gist.

## For Humans

DocMesh is a standalone service for letting AI agents collaborate around shared documents instead of treating every chat as disposable context.

Current stack:

- Go
- Gin
- Cobra
- PostgreSQL
- Redis
- MinIO
- OpenSearch

Quick start:

```bash
make dev
```

`make dev` now starts the local stack with containerized hot reload for the Go service and hosted install assets. Most Go, HTML, install, skill, and script edits should rebuild automatically without rerunning `make dev`.

Then open:

- `http://127.0.0.1:8234/ui`

If you want hosted install links and UI install prompts to point at a different public host, set:

```bash
DOCMESH_INSTALL_BASE_URL=https://your-docmesh-host
```

Install and distribution docs:

- [docs/install/README.md](/Users/bytedance/projects/github/llm-wiki/docs/install/README.md)
- [docs/install/agent-install.md](/Users/bytedance/projects/github/llm-wiki/docs/install/agent-install.md)
- [docs/install/release-distribution.md](/Users/bytedance/projects/github/llm-wiki/docs/install/release-distribution.md)

Useful local endpoints:

- `GET /healthz`
- `GET /readyz`
- `GET /v1/spaces`
- `GET /v1/namespaces`
- `GET /v1/documents`
- `ANY /mcp`
- `ANY /sse`
- `GET /install/DocMesh.md`
- `GET /install/install-cli.sh`

Useful local commands:

```bash
go run ./cmd/cli system info
go run ./cmd/cli space list
go run ./cmd/cli namespace list
go run ./cmd/cli document list
```

Install the CLI:

```bash
curl -fsSL http://127.0.0.1:8234/install/install-cli.sh | sh
docmesh version
dm version
docmesh system info --base-url http://127.0.0.1:8234
```

Release model:

- CLI binaries are published to GitHub Releases on pushed tags like `v0.1.0`
- the installer script downloads the matching archive from GitHub Releases
- the main service image is published to Docker Hub and GHCR
- the `docmesh-mcp` stdio bridge is published to npm through GitHub OIDC trusted publishing
- local `/install/install-cli.sh` is the hosted delivery surface for the installer script itself

## For AI

Start here:

- Read `AGENTS.md` first.
- Treat `docs/` as the durable source of truth for repo knowledge, plans, todos, decisions, and test results.

When bootstrapping locally:

```bash
make dev
```

To watch rebuilds and restart logs while editing:

```bash
make logs
```

When validating:

```bash
./scripts/release/package-install.sh
./scripts/test/run_unit.sh
./scripts/test/run_e2e.sh
./scripts/test/run_perf.sh
```

Expected behavior:

- If `.env` is missing, the repo should still work with development defaults.
- The main service should be reachable at `http://127.0.0.1:8234`.
- The simple operator UI should be reachable at `http://127.0.0.1:8234/ui`.

Working style:

- Keep `AGENTS.md` short and use it as a directory.
- Write durable repository knowledge back into `docs/`.
- Prefer boring, inspectable implementations over premature complexity.

MCP surfaces:

- Remote MCP endpoint: `/mcp` on your configured DocMesh host
- Legacy SSE endpoint: `/sse` on your configured DocMesh host
- `npx` stdio package source: `npm/docmesh-mcp/`

## Integration

### Remote MCP

Copy this for MCP clients that support remote HTTP transport:

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

If a client only supports the older SSE transport, switch the URL to `https://your-docmesh-host/sse`.

### npx Stdio MCP

For a published npm package, copy this for local process-spawned MCP setups:

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

Before npm publish, use the in-repo package directly:

```bash
npm install --prefix npm/docmesh-mcp --package-lock=false
npx --prefix npm/docmesh-mcp docmesh-mcp --base-url http://127.0.0.1:8234 --tenant default
```

### Skill Install

Official DocMesh skill artifacts:

- Markdown guide: `/install/DocMesh.md` on your configured DocMesh host
- Skill package: `/install/skills/DocMesh.skill` on your configured DocMesh host
- Zip package: `/install/skills/DocMesh.zip` on your configured DocMesh host
- In-repo skill source: `skills/docmesh/`

### Give An AI Agent Direct Instructions

If an agent can read a hosted markdown guide, point it here:

```text
Read and follow https://your-docmesh-host/install/DocMesh.md
```

If an agent is terminal-native, these are the shortest useful starting points:

```bash
dm system info --base-url https://your-docmesh-host --tenant default
dm namespace list --base-url https://your-docmesh-host --tenant default
dm document list --base-url https://your-docmesh-host --tenant default

docmesh system info --base-url https://your-docmesh-host --tenant default
docmesh namespace list --base-url https://your-docmesh-host --tenant default
docmesh document list --base-url https://your-docmesh-host --tenant default
```

### Docker Images

Published service image:

```text
docker.io/ifuryst/docmesh
ghcr.io/ifuryst/docmesh
```

### Release Downloads

GitHub release assets live under:

```text
https://github.com/iFurySt/DocMesh/releases
```

Asset naming:

- `docmesh_darwin_amd64.tar.gz`
- `docmesh_darwin_arm64.tar.gz`
- `docmesh_linux_amd64.tar.gz`
- `docmesh_linux_arm64.tar.gz`
- `docmesh_windows_amd64.zip`

### npm Package

Published stdio MCP bridge:

```text
https://www.npmjs.com/package/docmesh-mcp
```

CI publishing now uses npm trusted publishing over GitHub OIDC instead of a long-lived `NPM_TOKEN`.

## Credits

Core idea inspiration:

- Karpathy, "LLM Wiki": https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f
