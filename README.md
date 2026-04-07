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

Useful local endpoints:

- `GET /healthz`
- `GET /readyz`
- `GET /v1/spaces`
- `GET /v1/namespaces`
- `GET /v1/documents`

Useful local commands:

```bash
go run ./cmd/cli system info
go run ./cmd/cli space list
go run ./cmd/cli namespace list
go run ./cmd/cli document list
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

## Credits

Core idea inspiration:

- Karpathy, "LLM Wiki": https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f
