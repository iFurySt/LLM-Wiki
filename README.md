# LLM-Wiki

Agent-native knowledge service for multi-tenant document CRUD, revisions, grounding, and shared collaboration over HTTP, CLI, and web, inspired by Karpathy's LLM Wiki gist.

## For Humans

LLM-Wiki is a standalone service for letting AI agents collaborate around shared documents instead of treating every chat as disposable context.

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
LLM_WIKI_INSTALL_BASE_URL=https://your-llm-wiki-host
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
- `GET /install/LLM-Wiki.md`
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
llm-wiki version
lw version
llm-wiki system info --base-url http://127.0.0.1:8234
```

Release model:

- pushes to `main` publish `ghcr.io/ifuryst/llm-wiki:beta`
- CLI binaries are published to GitHub Releases on pushed tags like `v0.1.0`
- the installer script downloads the matching archive from GitHub Releases
- the main service image is published to Docker Hub and GHCR
- the `@ifuryst/llm-wiki-mcp` stdio bridge is published to npm through GitHub OIDC trusted publishing
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

- Remote MCP endpoint: `/mcp` on your configured LLM-Wiki host
- Legacy SSE endpoint: `/sse` on your configured LLM-Wiki host
- `npx` stdio package source: `npm/llm-wiki-mcp/`

## Integration

### Remote MCP

Copy this for MCP clients that support remote HTTP transport:

```json
{
  "llm-wiki": {
    "type": "http",
    "url": "https://your-llm-wiki-host/mcp",
    "headers": {
      "X-LLM-Wiki-Tenant-ID": "default"
    }
  }
}
```

If a client only supports the older SSE transport, switch the URL to `https://your-llm-wiki-host/sse`.

### npx Stdio MCP

For a published npm package, copy this for local process-spawned MCP setups:

```json
{
  "llm-wiki": {
    "command": "npx",
    "args": [
      "-y",
      "@ifuryst/llm-wiki-mcp",
      "--base-url",
      "https://your-llm-wiki-host",
      "--tenant",
      "default"
    ]
  }
}
```

Before npm publish, use the in-repo package directly:

```bash
npm install --prefix npm/llm-wiki-mcp --package-lock=false
npx --prefix npm/llm-wiki-mcp llm-wiki-mcp --base-url http://127.0.0.1:8234 --tenant default
```

### Skill Install

Official LLM-Wiki skill artifacts:

- Markdown guide: `/install/LLM-Wiki.md` on your configured LLM-Wiki host
- Skill package: `/install/skills/LLM-Wiki.skill` on your configured LLM-Wiki host
- Zip package: `/install/skills/LLM-Wiki.zip` on your configured LLM-Wiki host
- In-repo skill source: `skills/llm-wiki/`

### Give An AI Agent Direct Instructions

If an agent can read a hosted markdown guide, point it here:

```text
Read and follow https://your-llm-wiki-host/install/LLM-Wiki.md
```

If an agent is terminal-native, these are the shortest useful starting points:

```bash
lw system info --base-url https://your-llm-wiki-host --tenant default
lw namespace list --base-url https://your-llm-wiki-host --tenant default
lw document list --base-url https://your-llm-wiki-host --tenant default

llm-wiki system info --base-url https://your-llm-wiki-host --tenant default
llm-wiki namespace list --base-url https://your-llm-wiki-host --tenant default
llm-wiki document list --base-url https://your-llm-wiki-host --tenant default
```

### Docker Images

Published service image:

```text
docker.io/ifuryst/llm-wiki
ghcr.io/ifuryst/llm-wiki
```

### Release Downloads

GitHub release assets live under:

```text
https://github.com/iFurySt/LLM-Wiki/releases
```

Asset naming:

- `llm-wiki_darwin_amd64.tar.gz`
- `llm-wiki_darwin_arm64.tar.gz`
- `llm-wiki_linux_amd64.tar.gz`
- `llm-wiki_linux_arm64.tar.gz`
- `llm-wiki_windows_amd64.zip`

### npm Package

Published stdio MCP bridge:

```text
https://www.npmjs.com/package/@ifuryst/llm-wiki-mcp
```

CI publishing now uses npm trusted publishing over GitHub OIDC instead of a long-lived `NPM_TOKEN`.

## Credits

Core idea inspiration:

- Karpathy, "LLM Wiki": https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f
