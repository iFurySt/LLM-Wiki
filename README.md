# LLM-Wiki

LLM-Wiki is a shared knowledge service for agents.

It gives agents one durable backend for:

- documents
- revisions
- scoped access by `ns`
- access over HTTP, CLI, MCP, and web UI

Current stage:

- single service
- PostgreSQL as the only runtime dependency
- top boundary: `ns`
- content grouping: `folder`
- source of truth: LLM-Wiki itself, not downstream sync targets

## Run

```bash
make dev
```

Then open:

- `http://127.0.0.1:8234/setup`
- `http://127.0.0.1:8234/ui`

For browser verification with the repo profile:

```bash
make browser-open
make browser-mcp
```

## Use

Local CLI:

```bash
go run ./cmd/cli system info
go run ./cmd/cli auth login --base-url http://127.0.0.1:8234 --device-code
go run ./cmd/cli ns list --base-url http://127.0.0.1:8234
go run ./cmd/cli folder list --base-url http://127.0.0.1:8234
go run ./cmd/cli document list --base-url http://127.0.0.1:8234
```

Install the CLI:

```bash
curl -fsSL http://127.0.0.1:8234/install/install-cli.sh | sh
lw version
```

## Surfaces

- web UI: `/ui`
- HTTP API: `/v1/*`
- MCP: `/mcp`
- SSE: `/sse`
- hosted install guide: `/install/LLM-Wiki.md`

## Repo Map

- [AGENTS.md](AGENTS.md)
- [docs/README.md](docs/README.md)
- [docs/knowledge/product.md](docs/knowledge/product.md)
- [docs/knowledge/architecture.md](docs/knowledge/architecture.md)
- [docs/knowledge/repo-map.md](docs/knowledge/repo-map.md)
- [docs/install/README.md](docs/install/README.md)

## Notes

- first boot happens at `/setup`
- local default bootstrap token is `dev-bootstrap-token`
- CLI profile state lives in `~/.llm-wiki/config.json`
