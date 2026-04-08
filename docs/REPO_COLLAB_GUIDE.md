# Repo Collaboration Guide

This document defines the default collaboration contract for `LLM-Wiki`.

## Core Rules

- Read the durable docs before changing code, prompts, install flows, or repo structure.
- Treat `docs/` as the durable source of truth; do not leave stable repository knowledge only in chat.
- Keep `AGENTS.md` short. Put detailed rules in `docs/` and let `AGENTS.md` link to them.
- Prefer small, coherent changes that update code and durable docs in the same pass.
- Do not overwrite or revert unrelated user changes in the working tree.

## Default Read Order

Read these at the start of a normal coding round:

1. `docs/README.md`
2. `docs/knowledge/product.md`
3. `docs/knowledge/architecture.md`
4. `docs/knowledge/repo-map.md`
5. Relevant active plans in `docs/plans/active/`
6. Relevant near-term backlog in `docs/todos/`

## Skills

- `skills/llm-wiki/SKILL.md` is the entrypoint for the official LLM-Wiki skill.
- Read the skill before changing `skills/llm-wiki/`, hosted skill downloads, install prompts, or agent guidance that tells users how to work with shared durable knowledge.
- When a task should preserve reusable facts or procedures, prefer writing them into durable docs or LLM-Wiki materials instead of leaving them only in chat.

## Documentation Sync

Update docs in the same change whenever the code alters stable behavior:

- update `docs/knowledge/` when terminology, architecture, interfaces, or repo layout changes
- update `docs/plans/` when an execution plan starts, changes materially, or completes
- update `docs/todos/` when near-term priorities change
- update `docs/decisions/` when a durable product or technical decision is made, replaced, or superseded
- update `docs/worklog/` when a milestone lands that future contributors will need to understand
- update `docs/test-results/` when a validation run is worth preserving as a durable record
- update `docs/install/` when install, auth, packaging, release, or hosted distribution flows change

## Task-Specific Reads

Read these before making related changes:

- `docs/install/README.md`: install surface, packaging, hosted distribution, release-facing changes
- `skills/llm-wiki/SKILL.md`: official skill content, skill packaging, agent prompt guidance, CLI/HTTP/MCP usage guidance
- `skills/chrome-devtools-cli/SKILL.md` plus `docs/UI_AGENT_GUIDE.md`: browser automation, UI debugging, setup/install/admin UI changes, and browser-side acceptance
- `docs/decisions/README.md` plus the current monthly log: decision review or new significant direction changes
- `docs/worklog/README.md` plus the current monthly log: milestone logging or historical tracing work
- `docs/test-results/README.md`: durable validation writeback or result curation

## Commit Expectations

Before commit:

1. Re-read `docs/HISTORY_GUIDE.md`.
2. Check whether the change requires doc updates in `knowledge`, `plans`, `todos`, `decisions`, `worklog`, `test-results`, or `install`.
3. Ensure the commit contains only intended files and does not bundle unrelated local edits.
4. Make sure any user-visible behavior change is reflected in the durable docs that future agents will read first.
