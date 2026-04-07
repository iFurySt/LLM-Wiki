# AGENTS.md

This repository builds `DocMesh`, an agent-native knowledge service for multi-tenant document collaboration.

`AGENTS.md` is the table of contents, not the encyclopedia.

## Start Here

- [docs/README.md](/Users/ifuryst/projects/github/DocMesh/docs/README.md): docs map and update rules
- [docs/knowledge/product.md](/Users/ifuryst/projects/github/DocMesh/docs/knowledge/product.md): product framing, scope, core concepts
- [docs/knowledge/architecture.md](/Users/ifuryst/projects/github/DocMesh/docs/knowledge/architecture.md): system model and resource boundaries
- [docs/knowledge/repo-map.md](/Users/ifuryst/projects/github/DocMesh/docs/knowledge/repo-map.md): repo structure and implementation status

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

## Current Focus

- evolve the shared document model and revision workflows
- improve agent integration through MCP, CLI, and hosted install surfaces
- keep repo knowledge structured and current as the codebase grows
