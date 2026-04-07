# Product

## Name

`DocMesh`

## What It Is

DocMesh is an agent-native knowledge service for multi-tenant document collaboration.

It is built for cases where many AI agents need to read, create, update, and organize shared documents while preserving:

- revision history
- source grounding
- namespace and ACL boundaries
- approval-aware write paths
- compatibility with HTTP, CLI, and later MCP

## Product Framing

DocMesh is not a classic wiki app and not a generic RAG stack.

It is a shared document system for agents:

- raw inputs can stay immutable
- derived documents can be maintained over time
- drafts can exist before publication
- writes can be proposed, reviewed, applied, and audited

## Target Users

- teams building agent-native products
- internal AI platforms with many tenant-scoped agents
- systems where multiple agents need a shared, durable knowledge layer

## Early Non-Goals

- building a rich human-first editor before the service model is stable
- building a chat memory archive
- storing every conversation artifact as durable knowledge
- centering the product on vector retrieval

## Knowledge Domains

Within a tenant space, documents are expected to live in domains such as:

- `org/`: tenant-wide stable knowledge
- `projects/`: project and initiative knowledge
- `people/` or `accounts/`: sensitive customer or stakeholder knowledge
- `drafts/`: candidate and pre-publication material
- `agents/`: optional agent-private or agent-owned working areas

## Initial Principles

- tenant is the top isolation boundary
- namespace and document ACL determine visibility, not author alone
- not every chat detail deserves durable storage
- only information with repeat value should graduate into formal knowledge
