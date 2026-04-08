# Architecture

## Current Direction

LLM-Wiki starts as a standalone Go service with:

- HTTP API
- thin CLI wrapper over HTTP with browser/device login and local profile storage
- MCP server over Streamable HTTP and SSE
- PostgreSQL persistence
- hosted install surfaces under `/install/*`
- an npm stdio MCP bridge for `npx`-style local integrations

Supporting infrastructure already wired for local development:

- Redis for cache and coordination
- MinIO for object storage
- OpenSearch for search and indexing

## Core Resource Model

The current planned resource model is:

- `tenants`
- `spaces`
- `namespaces`
- `documents`
- `revisions`
- `patches`
- `sources`
- `citations`
- `jobs`
- `agents`

## Initial Model Assumptions

- one tenant maps to one primary space in v1
- a space contains many namespaces
- documents belong to a namespace
- revisions are immutable
- writes should prefer patch and revision semantics over in-place mutation
- sources remain immutable and serve as evidence

## Access Model

Access is expected to be evaluated by:

- tenant boundary
- namespace policy
- document ACL
- caller identity

Current implementation direction:

- bearer token auth for HTTP, CLI, and MCP
- service principals with tenant-scoped fine-grained tokens
- browser and device-code login for human CLI sessions
- first-boot setup flow for choosing the default tenant and initial admin account
- username/password-backed browser approvals and device-code approvals for human login
- cookie-backed web admin sessions for browser management pages
- audit metadata derived from authenticated principal context when possible

Caller identity should carry fields like:

- `tenant_id`
- `user_id`
- `agent_id`
- `team_ids`
- `agent_kind`

## Knowledge Layering

The service should preserve separation between:

1. Raw sources
2. Derived knowledge
3. Drafts and candidate work
4. System control data

## Integration Intent

The service is intended to integrate with agent platforms like `as-next` as a shared knowledge substrate:

- chat and task flows can query it
- agents can propose patches into it
- approved results can be written back into it
- admin surfaces can inspect revision and audit history

## Current Agent Integration Surfaces

LLM-Wiki currently exposes these agent-facing entry points:

- HTTP API for direct JSON integrations
- thin `llm-wiki` CLI for terminal-first workflows with `--token`, `--token-file`, env vars, and `~/.llm-wiki/config.json`
- remote MCP endpoint at `/mcp`
- legacy MCP SSE endpoint at `/sse`
- official in-repo `llm-wiki` skill
- `@ifuryst/llm-wiki-mcp` npm package for stdio MCP via `npx`

This keeps the same backend reachable from hosted agents, local coding agents, and process-spawned MCP clients.

For human operators, the service also exposes:

- `/setup` for one-time initialization
- `/admin/login` for browser login
- `/admin/users` for basic user administration
