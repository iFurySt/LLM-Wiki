# AGENTS.md

This repository builds `LLM-Wiki`, an agent-native knowledge service for shared document collaboration across multiple `ns` scopes.

`AGENTS.md` is the table of contents, not the encyclopedia.

Keep detailed collaboration rules in `docs/`; update those docs when stable behavior changes.

## Start Here

- [docs/README.md](docs/README.md): docs map and update rules
- [docs/knowledge/product.md](docs/knowledge/product.md): product framing, scope, core concepts
- [docs/knowledge/architecture.md](docs/knowledge/architecture.md): system model and resource boundaries
- [docs/knowledge/repo-map.md](docs/knowledge/repo-map.md): repo structure and implementation status
- [docs/install/README.md](docs/install/README.md): install and distribution entrypoint

## Every Round

- [docs/REPO_COLLAB_GUIDE.md](docs/REPO_COLLAB_GUIDE.md): repo-level collaboration rules, required read order, doc sync, skills usage, and pre-commit expectations

## Before Commit

- [docs/HISTORY_GUIDE.md](docs/HISTORY_GUIDE.md): when to update worklog, decisions, and durable test records before commit

## Read By Task Type

- [skills/llm-wiki/SKILL.md](skills/llm-wiki/SKILL.md): required when changing the official skill, hosted skill downloads, or agent guidance around durable shared knowledge
- [skills/chrome-devtools-cli/SKILL.md](skills/chrome-devtools-cli/SKILL.md): required for browser automation, UI debugging, and browser-side acceptance work
- [docs/UI_AGENT_GUIDE.md](docs/UI_AGENT_GUIDE.md): required for `/ui`, `/setup`, `/install/*`, `/admin/*`, browser login, and other user-visible web changes
- [docs/install/README.md](docs/install/README.md): required for install, packaging, release, and hosted distribution changes
- [docs/decisions/README.md](docs/decisions/README.md): required when making or reviewing significant product or technical decisions
- [docs/test-results/README.md](docs/test-results/README.md): required when preserving durable validation outcomes

## Working Rules

- Durable repository knowledge belongs in `docs/`, not only in chat.
- Keep `AGENTS.md` short; add detail in the referenced docs.
- Treat plans and todos as first-class repo artifacts.
- Update docs whenever code, structure, or stable decisions change.

## Docs Areas

- `docs/knowledge/`: long-lived source-of-truth docs
- `docs/plans/`: execution plans, active and completed
- `docs/todos/`: fast-moving task backlogs and near-term work
- `docs/decisions/`: append-only significant decisions
- `docs/test-results/`: durable records of unit, e2e, and performance runs
- `docs/worklog/`: chronological milestone log
- `docs/references/`: distilled external references that affect repo practice
- `docs/install/`: durable install and release-distribution guidance

## Current Focus

- evolve the shared document model and revision workflows
- improve agent integration through MCP, CLI, and hosted install surfaces
- keep repo knowledge structured and current as the codebase grows
