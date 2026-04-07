# Repo Map

## Current Status

The repository is in bootstrap phase.

## Top-Level Layout

- `AGENTS.md`: short table of contents for agent operation
- `cmd/`: entrypoints for the HTTP service and the thin CLI wrapper
- `internal/`: private Go packages for app wiring, config, logging, HTTP server, CLI, and client code
- `deploy/dev/`: local development infrastructure definitions
- `docs/`: durable repository knowledge and execution artifacts
- `README.md`: quickstart and local setup notes
- `.env.example`: environment variable template
- `Makefile`: common development commands

## Docs Layout

- `docs/knowledge/`: long-lived product, architecture, and repo docs
- `docs/plans/active/`: current execution plans
- `docs/plans/completed/`: closed execution plans
- `docs/todos/`: near-term backlog
- `docs/decisions/`: major decisions
- `docs/test-results/`: durable validation and benchmark records
- `docs/worklog/`: chronological milestones
- `docs/references/`: distilled external references

## Planned Layout

Expected directories as implementation grows:

- `pkg/`: reusable public packages if needed
- `configs/`: local or example configuration
- `scripts/`: helper scripts
- `test/` or `tests/`: integration and end-to-end tests
- `internal/db/migrations/`: embedded SQL migrations used at startup

## Implementation Status

- naming: settled on `DocMesh`
- docs system: initialized and structured
- Go service scaffold: initialized
- thin HTTP CLI scaffold: initialized
- test-result tracking structure: initialized
- database schema: initial v0 migration implemented
- HTTP API: readiness, structured error responses, space list, namespace CRUD/list/archive, document CRUD/list/filter, slug lookup, and archive
- CLI: system, space, namespace, and document commands implemented, including list and archive flows
- UI: simple Gin-served HTML page for manual inspection and creation
- infra manifests: local development compose initialized
- dockerized dev entrypoint: `make dev`
