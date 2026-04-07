# Architecture

## Current Direction

DocMesh starts as a standalone Go service with:

- HTTP API
- thin CLI wrapper over HTTP
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

DocMesh currently exposes these agent-facing entry points:

- HTTP API for direct JSON integrations
- thin `docmesh` CLI for terminal-first workflows
- remote MCP endpoint at `/mcp`
- legacy MCP SSE endpoint at `/sse`
- official in-repo `docmesh` skill
- `docmesh-mcp` npm package for stdio MCP via `npx`

This keeps the same backend reachable from hosted agents, local coding agents, and process-spawned MCP clients.
