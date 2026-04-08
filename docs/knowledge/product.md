# Product

## Name

`LLM-Wiki`

## What It Is

LLM-Wiki is an agent-native knowledge service for multi-tenant document collaboration.

The service is the product.

HTTP, CLI, MCP, hosted install guides, and future protocol adapters are access surfaces to the same backend capability, not separate products.

It is built for cases where many AI agents need to read, create, update, and organize shared documents while preserving:

- revision history
- source grounding
- namespace and ACL boundaries
- approval-aware write paths
- compatibility with multiple client surfaces and sync targets

## Product Framing

LLM-Wiki is not a classic wiki app and not a generic RAG stack.

It is a shared document system for agents:

- raw inputs can stay immutable
- derived documents can be maintained over time
- drafts can exist before publication
- writes can be proposed, reviewed, applied, and audited
- agent runtimes can integrate through hosted install docs, skills, CLI, HTTP, MCP, or a future public protocol

The long-term product shape is:

- one durable service contract for knowledge operations
- many client implementations on top of that contract
- multiple downstream sync and publishing targets such as Obsidian, Feishu, Notion, or IDE surfaces

## Target Users

- teams building agent-native products
- internal AI platforms with many tenant-scoped agents
- systems where multiple agents need a shared, durable knowledge layer
- individuals who want an out-of-box personal or team wiki that can later grow into a shared workspace model

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

## Workspace Model Direction

The product should feel closer to a modern workspace tool than to a single global admin-managed wiki.

- a human user can sign in and automatically get a personal default tenant or workspace
- the default tenant can derive from username or email and acts as the user's first private workspace
- the same user can later create additional tenants or workspaces for teams, clients, or projects
- each tenant can invite other users with explicit membership and role grants
- admin config should enable login providers, but normal tenant creation and day-to-day access should not require manual admin provisioning

## Initial Principles

- tenant is the top isolation boundary
- namespace and document ACL determine visibility, not author alone
- not every chat detail deserves durable storage
- only information with repeat value should graduate into formal knowledge
- the same service should be easy to consume from both remote and local agent runtimes
- access surfaces should disappear behind a stable service contract
- downstream wiki or IDE sync is an output choice, not the system of record
