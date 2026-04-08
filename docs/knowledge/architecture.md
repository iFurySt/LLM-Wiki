# Architecture

## Current Direction

LLM-Wiki starts as a standalone Go service with:

- HTTP API
- thin CLI wrapper over HTTP with browser/device login and local profile storage
- MCP server over Streamable HTTP and SSE
- future protocol surface for third-party client implementations
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
- `tenant_memberships`
- `oauth_providers`
- `oauth_accounts`
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
- a signed-in user should be able to get a personal default tenant automatically
- a user may belong to multiple tenants and create additional tenants over time
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
- browser login with loopback callback for human CLI sessions
- device-code login as fallback for headless or constrained environments
- first-boot setup flow for choosing the default tenant and initial admin account
- username/password-backed browser approvals and device-code approvals for human login
- admin-configured Google and GitHub OAuth providers for human sign-in
- automatic user creation on first successful OAuth login when allowed by server policy
- automatic personal-tenant creation on first successful user login
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

The architecture should separate:

- the canonical backend storage model owned by LLM-Wiki
- client-facing access methods such as CLI, HTTP, MCP, and future protocol clients
- output adapters that sync or publish knowledge into external tools like Obsidian, Feishu, Notion, or IDE extensions

## Current Agent Integration Surfaces

LLM-Wiki currently exposes these agent-facing entry points:

- HTTP API for direct JSON integrations
- thin `llm-wiki` CLI for terminal-first workflows with `--token`, `--token-file`, env vars, and `~/.llm-wiki/config.json`
- remote MCP endpoint at `/mcp`
- legacy MCP SSE endpoint at `/sse`
- official in-repo `llm-wiki` skill
- `@ifuryst/llm-wiki-mcp` npm package for stdio MCP via `npx`

This keeps the same backend reachable from hosted agents, local coding agents, and process-spawned MCP clients.

Longer-term, these should converge on one public service contract so external clients can implement their own frontends without re-encoding product semantics per surface.

The first service-first downstream client contract is documented in [docs/decisions/2026-04-08-service-first-live-browse-protocol.md](/Users/bytedance/projects/github/LLM-Wiki/docs/decisions/2026-04-08-service-first-live-browse-protocol.md). The current Obsidian adapter now uses that same service model for file mirroring instead of keeping a dedicated live-browse view.

## Output Adapter Direction

LLM-Wiki should keep its own internal storage and revision model even when users prefer another tool as the reading or editing surface.

Planned adapter direction:

- Obsidian sync or plugin support as the first downstream target
- additional sync or publish targets such as Feishu and Notion
- IDE-facing integrations that expose tenant, namespace, and document workflows inside coding tools

The adapter layer should translate between LLM-Wiki's backend model and each target's file, page, or block model without making the external tool the source of truth.

Current implementation status:

- an Obsidian sync plugin mirrors the current LLM-Wiki tenant into vault files for native Files-pane browsing while keeping LLM-Wiki as the source of truth
- a CLI export command still exists as a fallback artifact bridge, but is not the preferred long-term adapter model
- current Obsidian sync is polling-based and read-only
- future Obsidian plugin and protocol work should build on the same service model rather than bypass it

For human operators, the service also exposes:

- `/setup` for one-time initialization
- `/admin/login` for browser login
- `/admin/users` for basic user administration
