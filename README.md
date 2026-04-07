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

Then open:

- `http://127.0.0.1:8234/ui`

Install surfaces:

- `http://127.0.0.1:8234/install/DocMesh.md`
- `http://127.0.0.1:8234/install/install-cli.sh`
- `http://127.0.0.1:8234/install/skills/DocMesh.skill`
- `http://127.0.0.1:8234/install/skills/DocMesh.zip`

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
docmesh system info --base-url http://127.0.0.1:8234
```

## For AI

Start here:

- Read `AGENTS.md` first.
- Treat `docs/` as the durable source of truth for repo knowledge, plans, todos, decisions, and test results.

When bootstrapping locally:

```bash
make dev
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

- Remote MCP endpoint: `http://127.0.0.1:8234/mcp`
- Legacy SSE endpoint: `http://127.0.0.1:8234/sse`
- `npx` stdio package source: `npm/docmesh-mcp/`

## Integration

### Remote MCP

Copy this for MCP clients that support remote HTTP transport:

```json
{
  "docmesh": {
    "type": "http",
    "url": "http://127.0.0.1:8234/mcp",
    "headers": {
      "X-DocMesh-Tenant-ID": "default"
    }
  }
}
```

If a client only supports the older SSE transport, switch the URL to `http://127.0.0.1:8234/sse`.

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
      "http://127.0.0.1:8234",
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

- Markdown guide: `http://127.0.0.1:8234/install/DocMesh.md`
- Skill package: `http://127.0.0.1:8234/install/skills/DocMesh.skill`
- Zip package: `http://127.0.0.1:8234/install/skills/DocMesh.zip`
- In-repo skill source: `skills/docmesh/`

### Give An AI Agent Direct Instructions

If an agent can read a hosted markdown guide, point it here:

```text
Read and follow http://127.0.0.1:8234/install/DocMesh.md
```

If an agent is terminal-native, these are the shortest useful starting points:

```bash
docmesh system info --base-url http://127.0.0.1:8234 --tenant default
docmesh namespace list --base-url http://127.0.0.1:8234 --tenant default
docmesh document list --base-url http://127.0.0.1:8234 --tenant default
```

## Credits

Core idea inspiration:

- Karpathy, "LLM Wiki": https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f
