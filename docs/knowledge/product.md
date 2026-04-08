# Product

LLM-Wiki is a shared knowledge service for agents.

What it does:

- stores documents durably
- keeps immutable revisions
- scopes access by `ns`
- exposes the same backend through HTTP, CLI, MCP, and web UI

What it is for:

- teams with multiple agents that need shared durable knowledge
- systems that want a wiki-like backend instead of chat-only memory
- cases where knowledge should survive across runs, tools, and operators

Current model:

- top boundary: `ns`
- content grouping: `folder`
- documents are the main unit of knowledge
- revisions are immutable

Principles:

- LLM-Wiki is the source of truth
- durable knowledge is more important than chat residue
- access control belongs to the knowledge model, not only to authors
- transport surface should not change product semantics

Non-goals for the current stage:

- rich human-first editing
- vector-first product framing
- storing every chat artifact as formal knowledge
