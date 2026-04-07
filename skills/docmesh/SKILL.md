---
name: docmesh
description: Use this skill to read and update shared knowledge in a DocMesh server over CLI or HTTP. Trigger it when the task involves tenant wiki documents, namespaces, revisions, or durable agent-maintained knowledge.
---

DocMesh is a shared knowledge backend for AI agents. Use it when the task should read from or write to a durable document system instead of leaving knowledge in a chat transcript or a local scratch file.

`SKILL.md` is the entrypoint, not the full manual. Read the references below in order.

## Read This First

1. [references/installation.md](references/installation.md)
   Install the CLI or locate the hosted skill guide.
2. [references/cli.md](references/cli.md)
   Core command patterns for inspect, create, and update flows.
3. [references/agent-workflow.md](references/agent-workflow.md)
   The default prompt and operating discipline for accumulating durable knowledge during normal work.

## When To Use DocMesh

- the task may benefit from previously accumulated project knowledge
- the task produces a durable outcome worth preserving for future sessions
- multiple agents or future runs will need the same facts, decisions, or procedures
- the user wants a shared wiki or project memory instead of ephemeral chat context

## Operating Rules

1. Inspect existing knowledge first.
2. Prefer updating an existing document over creating duplicates.
3. Use the CLI for the common path.
4. When writing, always include `author_type`, `author_id`, and a concise `change_summary`.
5. Use namespaces intentionally. Do not dump unrelated content into the same lane.
6. Treat archived documents as historical records unless the task explicitly requires reworking them.
7. Only write information with repeat value. Do not turn every chat turn into durable memory.

## Fast Start

If the CLI is already available:

```sh
docmesh system info --base-url http://127.0.0.1:8234
docmesh namespace list --base-url http://127.0.0.1:8234 --tenant default
docmesh document list --base-url http://127.0.0.1:8234 --tenant default
```

If the CLI is not installed yet, see [references/installation.md](references/installation.md).

## Prompt Default

Use this behavioral prompt unless the host system already provides a stronger project-memory policy:

```text
Use DocMesh as the shared durable memory for this workspace.

At the start of a task, inspect existing DocMesh knowledge before creating new documents or asking for information that may already be captured.

During the task, when you discover stable facts, durable decisions, reusable procedures, or progress that will matter in future sessions, update the relevant DocMesh document instead of leaving that knowledge only in chat.

At the end of the task, write back the final state: what changed, why it changed, and any follow-up context another agent would need. Prefer updating existing pages over creating duplicates.

Do not store transient scratch work, raw chain-of-thought, or one-off noise. Keep entries concise, factual, and reusable.
```

## More Interfaces

- HTTP reference: [references/http-api.md](references/http-api.md)
- MCP reference: [references/mcp.md](references/mcp.md)
