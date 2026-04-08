# 2026-04-08 Service-First Live Browse Protocol

## Status

Accepted.

## Context

LLM-Wiki's core value is the service itself, not any one access surface.

CLI, HTTP, MCP, future protocol clients, and downstream adapters like Obsidian should all project the same backend semantics instead of inventing separate data models per client.

The first external knowledge-surface integration is Obsidian. The user requirement is live browsing inside Obsidian, not periodic export into local markdown files.

## Decision

Define the first client-sync contract as a service-first live browse protocol over the existing authenticated HTTP API.

Protocol v0 rules:

- LLM-Wiki remains the source of truth for documents, revisions, `ns` scopes, and memberships.
- downstream clients may cache remote state for rendering, but should not treat local cache or vault files as canonical state
- the minimum browse contract is:
  - `GET /v1/auth/whoami`
  - `GET /v1/workspaces`
  - `POST /v1/auth/switch-tenant`
  - `GET /v1/namespaces`
  - `GET /v1/documents`
- auth uses bearer tokens, and desktop clients may reuse `~/.llm-wiki/config.json` as the default local credential source
- `ns` switching is token-based, using `switch-tenant` to mint an `ns`-bound token for the selected scope
- live sync in v0 is polling-based; clients should re-fetch list and document state on refresh or interval
- v0 is read-first; editing and bidirectional sync are deferred until revision-safe write semantics are specified for non-CLI clients

## Consequences

Positive:

- avoids export/import drift between LLM-Wiki and external tools
- keeps one backend semantic model across CLI, web, MCP, and Obsidian
- makes the first Obsidian adapter cheap to ship because it can build directly on existing APIs

Tradeoffs:

- polling is less efficient than a future event stream or subscription protocol
- current browse clients fetch the whole document list, which may need pagination or tree endpoints later
- write support is intentionally delayed to avoid creating unsafe edit races before revision-aware protocol semantics exist

## Follow-up

- add a dedicated public protocol document once live browse v0 stabilizes across more than one client
- introduce change feeds or subscriptions for lower-latency sync
- design write protocol semantics around explicit revision preconditions and conflict handling
