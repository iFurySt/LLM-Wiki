---
name: llm-wiki
description: Use this skill to work with the LLM-Wiki service as shared durable memory. Trigger it when a task should read, update, or preserve reusable knowledge in shared wiki documents instead of leaving it only in chat or local scratch files.
---

LLM-Wiki is a shared knowledge service for AI agents and human users.

The service is the product. CLI, HTTP, MCP, and future protocol clients are access surfaces to the same backend.

Use this skill when the task should read from or write to durable shared knowledge instead of leaving facts only in chat, local scratch files, or one-off notes.

`SKILL.md` is the entrypoint, not the full manual. Read the references below in order.

## Read This First

1. [references/installation.md](references/installation.md)
   Install the CLI or locate the hosted skill guide.
2. [references/cli.md](references/cli.md)
   Core command patterns for inspect, create, and update flows.
3. [references/agent-workflow.md](references/agent-workflow.md)
   The default prompt and operating discipline for accumulating durable knowledge during normal work.

## When To Use LLM-Wiki

- the task may benefit from previously accumulated project knowledge
- the task produces a durable outcome worth preserving for future sessions
- multiple agents or future runs will need the same facts, decisions, or procedures
- the user wants a shared wiki or project memory instead of ephemeral chat context
- the user is operating against a hosted LLM-Wiki service and wants a predictable auth and writeback flow

## When Not To Use It

- the user only wants a one-off summary with no durable value
- the information is transient scratch work, tentative reasoning, or raw chain-of-thought
- the task should stay entirely local and should not touch shared `ns` knowledge

## Operating Rules

1. Inspect existing knowledge first.
2. Prefer updating an existing document over creating duplicates.
3. Use the CLI for the common path unless the host explicitly wants HTTP or MCP.
4. When auth is needed, prefer the existing local profile in `~/.llm-wiki/`; if none exists, guide or invoke `llm-wiki auth login`.
5. Treat `ns` choice as an authenticated context, not a random string to invent on the fly.
6. When writing, always include `author_type`, `author_id`, and a concise `change_summary`.
7. Use folders intentionally. Do not dump unrelated content into the same lane.
8. Treat archived documents as historical records unless the task explicitly requires reworking them.
9. Only write information with repeat value. Do not turn every chat turn into durable memory.

## Default Agent Workflow

Follow this loop unless the user asks for something narrower:

1. Resolve the target server and auth context.
2. Inspect existing documents before drafting anything new.
3. Read only the minimum relevant documents or revisions.
4. Decide whether the result belongs in an existing document or a new one.
5. Apply the smallest durable write that preserves the useful outcome.
6. Report back what was read, what changed, and what remains open.

## Fast Path

For a normal hosted setup, prefer this sequence:

```sh
llm-wiki auth status
llm-wiki ns list
llm-wiki document list
```

If auth is missing:

```sh
llm-wiki auth login
llm-wiki auth whoami
```

## Fast Start

If the CLI is already available:

```sh
llm-wiki system info --base-url http://127.0.0.1:8234
llm-wiki folder list --base-url http://127.0.0.1:8234 --token dev-bootstrap-token
llm-wiki document list --base-url http://127.0.0.1:8234 --token dev-bootstrap-token
```

If the CLI is not installed yet, see [references/installation.md](references/installation.md).

## Hosted Login Expectations

In the normal interactive case:

- `llm-wiki auth login` should open the browser
- the CLI should listen on a localhost callback port
- successful login should update `~/.llm-wiki/` automatically
- future commands should reuse the stored server and `ns` context

Only fall back to device-code login when a browser callback is not practical.

## Prompt Default

Use this behavioral prompt unless the host system already provides a stronger project-memory policy:

```text
Use LLM-Wiki as the shared durable memory for the current `ns`.

At the start of a task, inspect existing LLM-Wiki knowledge before creating new documents or asking for information that may already be captured.

During the task, when you discover stable facts, durable decisions, reusable procedures, or progress that will matter in future sessions, update the relevant LLM-Wiki document instead of leaving that knowledge only in chat.

At the end of the task, write back the final state: what changed, why it changed, and any follow-up context another agent would need. Prefer updating existing pages over creating duplicates.

Do not store transient scratch work, raw chain-of-thought, or one-off noise. Keep entries concise, factual, and reusable.

Treat LLM-Wiki as the canonical service. Do not assume CLI-only behavior or local-file-backed semantics when the service already provides scoped auth, revision, and audit behavior.
```

## More Interfaces

- HTTP reference: [references/http-api.md](references/http-api.md)
- MCP reference: [references/mcp.md](references/mcp.md)
