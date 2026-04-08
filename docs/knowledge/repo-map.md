# Repo Map

## Current Status

The repository is in bootstrap phase.

## Top-Level Layout

- `AGENTS.md`: short table of contents for agent operation
- `cmd/`: entrypoints for the HTTP service and the thin CLI wrapper
- `internal/`: private Go packages for app wiring, config, logging, HTTP server, CLI, and client code
- `deploy/dev/`: local development infrastructure definitions
- `Dockerfile`: production-oriented container build for the main DocMesh service
- `.github/workflows/`: CI and release automation
- `docs/`: durable repository knowledge and execution artifacts
- `install/`: hosted install docs and scripts served from `:8234/install/*`
- `skills/`: official DocMesh agent skill source files
- `npm/`: publishable npm packages maintained in-repo
- `scripts/`: helper scripts for testing and release packaging
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
- `docs/install/`: durable install and release-distribution references

## Planned Layout

Expected directories as implementation grows:

- `pkg/`: reusable public packages if needed
- `configs/`: local or example configuration
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
- MCP: streamable HTTP and legacy SSE endpoints with DocMesh tools and resources
- CLI: system, space, namespace, and document commands implemented, including list and archive flows
- UI: retro wiki-style Gin-served HTML page for browsing, creating, editing, archiving, and install flows
- infra manifests: local development compose initialized
- dockerized dev entrypoint: `make dev` with containerized hot reload for the app service
- install surfaces: hosted markdown guide, shell installer, and skill package download endpoints implemented
- release packaging: multi-platform CLI archives and packaged skill assets generated into `dist/install/`
- GitHub release automation: pushed tags publish CLI and install assets to GitHub Releases, push the main service image to Docker Hub and GHCR, and publish `docmesh-mcp` to npm
- main-branch beta automation: pushes to `main` publish `ghcr.io/ifuryst/docmesh:beta` and can SSH-deploy that image to amoylab
- npm stdio bridge: `docmesh-mcp` package source added for `npx`-style MCP usage
