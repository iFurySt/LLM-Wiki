# Repo Map

The repo is still in an early service-first stage.

Top-level layout:

- `cmd/`: server and CLI entrypoints
- `internal/`: service, HTTP, CLI, config, repo, auth, UI
- `internal/db/migrations/`: schema migrations
- `docs/`: durable product and repo knowledge
- `deploy/`: local and production deployment files
- `install/`: hosted install assets
- `skills/`: official LLM-Wiki and browser automation skills
- `npm/`: stdio MCP bridge package
- `apps/`: downstream adapters such as the Obsidian plugin
- `scripts/`: test and release helpers

Current implementation status:

- PostgreSQL-backed service is working
- folder and document CRUD exist
- revisions exist
- HTTP API, CLI, MCP, and web UI all exist
- auth, setup, and browser/device login exist
- hosted install docs and CLI installer exist
- Obsidian mirror adapter exists
- the hosted `/ui` page now uses a small embedded React island with `@mui/x-tree-view` for the left file tree while keeping the rest of the page server-rendered
- the hosted `/ui` page now swaps wiki and install content through fragment fetches plus `history.pushState`, so document selection and filter changes do not require a full page reload

Current boundaries:

- source of truth is LLM-Wiki
- runtime dependency is PostgreSQL
- user-facing terminology is `ns` and `folder`
