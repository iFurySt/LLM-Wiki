# Architecture

Current shape:

- one Go service
- PostgreSQL as the only runtime dependency
- HTTP API, CLI, MCP, and web UI on top of the same backend

Current external model:

- `ns`
- `folders`
- `documents`
- `revisions`

Current storage note:

- tables and columns use `ns`, `folders`, and `folder_id`
- user-facing docs and CLI use the same names

Assumptions:

- one `ns` is the top isolation boundary
- one `ns` contains many folders
- one folder contains many documents
- revisions are immutable

Access model:

- bearer token auth
- `ns`-scoped identity
- folder and document visibility enforced by service logic

Main surfaces:

- HTTP API
- `llm-wiki` / `lw` CLI
- MCP at `/mcp`
- web UI at `/ui`

Human flows:

- `/setup` initializes the instance
- `/admin/login` and `/admin/users` support basic browser-side administration
- CLI login stores state in `~/.llm-wiki/config.json`

Adapter direction:

- LLM-Wiki remains the source of truth
- downstream tools like Obsidian are adapters, not primary storage
