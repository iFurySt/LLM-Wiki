---
name: docmesh
description: Use this skill to read and update shared knowledge in a DocMesh server over CLI or HTTP. Trigger it when the task involves tenant wiki documents, namespaces, revisions, or durable agent-maintained knowledge.
---

DocMesh is a shared knowledge backend for AI agents. Use it when the task should read from or write to a durable document system instead of leaving knowledge in a chat transcript or a local scratch file.

## Setup

If the `docmesh` CLI is not installed yet, see [references/installation.md](references/installation.md). Installation is a one-time prerequisite and is not part of the normal workflow.

## AI Workflow

1. Inspect existing knowledge first.
2. Prefer updating an existing document over creating duplicates.
3. When writing, always include `author_type`, `author_id`, and a concise `change_summary`.
4. Use namespaces intentionally. Do not dump unrelated content into the same lane.
5. Treat archived documents as historical records unless the task explicitly requires reworking them.

## CLI Usage

Use the CLI for the common path. See [references/cli.md](references/cli.md) for command patterns.

```sh
docmesh system info --base-url http://127.0.0.1:8234
docmesh namespace list --base-url http://127.0.0.1:8234 --tenant default
docmesh document list --base-url http://127.0.0.1:8234 --tenant default
```

## HTTP Usage

Use HTTP directly only when CLI wrapping is not appropriate. See [references/http-api.md](references/http-api.md).

## MCP Usage

If the agent runtime supports MCP, prefer the DocMesh MCP surfaces over raw HTTP.

See [references/mcp.md](references/mcp.md).

## Writing Discipline

- Reuse the same tenant consistently.
- Look up by slug before creating a new document if the namespace is already known.
- Use short, stable slugs.
- Keep change summaries factual.
- Avoid using DocMesh for transient scratch notes that do not need to survive the current task.
