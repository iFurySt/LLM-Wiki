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
- source-aware document create flows now exist for inline text, local files, and URL imports, with structured provenance stored on documents and revisions
- HTTP API, CLI, MCP, and web UI all exist
- auth, setup, and browser/device login exist
- hosted install docs and CLI installer exist
- Obsidian mirror adapter exists
- the hosted `/dashboard` page now uses a small embedded React island with `@mui/x-tree-view` for the left file tree while keeping the rest of the page server-rendered
- the hosted `/dashboard` page now swaps workspace and install content through fragment fetches plus `history.pushState`, so document selection and filter changes do not require a full page reload
- document selection inside the hosted `/dashboard` tree now updates the reader panel through a dedicated reader fragment, so the sidebar and outer workspace shell stay mounted during normal file browsing
- the hosted `/dashboard` knowledge page now uses a three-column workspace layout: left rail for ns switching, search, and the existing tree; center column for recent activity or document content; right rail for workspace stats when idle and revision history when a document is open
- revision selection in the hosted `/dashboard` now renders historical document bodies in the center reader without remounting the left tree shell
- legacy `/ui` and `/ui/install` endpoints remain as compatibility redirects to the matching `/dashboard` routes

Current boundaries:

- source of truth is LLM-Wiki
- runtime dependency is PostgreSQL
- user-facing terminology is `ns` and `folder`
