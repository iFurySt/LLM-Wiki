# Agent Workflow

This reference defines how an LLM or coding agent should use LLM-Wiki during normal work, not only for explicit wiki-editing tasks.

## Default Prompt

Use or adapt this prompt in agent systems that support custom instructions:

```text
You have access to LLM-Wiki, a shared durable knowledge system for this workspace.

Your job is not only to complete the current task, but also to keep important project knowledge up to date in LLM-Wiki.

Behavior:
- At task start, inspect LLM-Wiki for relevant existing documents, plans, decisions, or procedures.
- During the task, when you learn something stable and reusable, update the corresponding LLM-Wiki document.
- At task end, write back durable outcomes so the next agent does not need to rediscover them.
- Prefer editing an existing document over creating a near-duplicate.
- Use concise factual summaries, stable slugs, and the correct namespace.
- Always include `author_type`, `author_id`, and `change_summary` on writes.
- Do not store transient notes, hidden reasoning, or low-value chat residue.

Only information with repeat value should graduate into LLM-Wiki.
```

## When To Read

Read LLM-Wiki early when:

- the task touches an existing project, subsystem, customer, or workflow
- the user asks about prior decisions, known constraints, or current status
- you are about to create a plan, TODO, runbook, or project summary
- the task is recurring and prior work may already be documented

## When To Write

Write to LLM-Wiki when the task produces durable knowledge such as:

- project plans, execution status, and milestones
- stable architecture notes or implementation constraints
- decisions and their rationale
- operator runbooks and CLI usage patterns
- distilled research or external references that will be reused
- clean summaries of what changed in a project area

## When Not To Write

Do not write:

- temporary scratch notes for the current turn only
- speculative thoughts that have not been validated
- sensitive data unless the target namespace is explicitly appropriate
- raw logs or large transient command outputs without distillation
- every conversational detail from the session

## Document Pattern

Use a simple, reusable page shape:

```md
# Title

## Summary
Short durable overview.

## Current State
What is true now.

## Key Details
The facts, commands, decisions, or references worth preserving.

## Next Steps
Only if they are still relevant after this session ends.
```

## Namespace Guidance

- `org/`: tenant-wide stable knowledge and shared standards
- `projects/`: project-specific plans, status, architecture, and runbooks
- `drafts/`: candidate material not ready to be treated as settled
- `agents/`: agent-owned working memory when it must persist across sessions

## Recommended Rhythm

1. Inspect likely namespaces and documents first.
2. Reuse an existing slug if the page already exists.
3. Create a page only when the knowledge does not already have a natural home.
4. Update during the task when meaningful state changes happen.
5. Do a final write-back before finishing if the task produced durable outcomes.

## CLI Shape

Typical loop:

```sh
llm-wiki namespace list --base-url http://127.0.0.1:8234 --token dev-bootstrap-token
llm-wiki document list --base-url http://127.0.0.1:8234 --token dev-bootstrap-token --namespace-id 1
llm-wiki document get-by-slug --base-url http://127.0.0.1:8234 --token dev-bootstrap-token 1 launch-plan
```

If the page exists, update it by document id. If not, create it with a stable slug.
