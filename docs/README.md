# Docs

This repository treats `docs/` as the system of record for durable repository knowledge.

The structure follows two ideas:

- progressive disclosure: agents should start from a short map and drill down only where needed
- plans and todos are first-class artifacts, not chat residue

## Layout

- `knowledge/`
  Stable product, architecture, and repo knowledge.
- `plans/`
  Execution plans for substantial work. Split into `active/` and `completed/`.
- `todos/`
  Higher-churn near-term worklists and implementation backlog.
- `decisions/`
  Append-only major product and technical decisions.
- `test-results/`
  Durable records of repeatable test runs and benchmark outputs.
- `worklog/`
  Chronological notable milestones.
- `references/`
  Distilled notes from important external references that shape repo practice.
- `install/`
  Durable installation and release-distribution guidance for humans and AI agents.

## Update Rules

Update `knowledge/` when:

- repo structure changes
- architecture changes
- product scope or terminology changes

Update `plans/` when:

- a task is large enough to need tracked execution
- a plan materially changes state

Update `todos/` when:

- the near-term backlog changes
- priorities or sequencing change

Update `decisions/` when:

- a meaningful decision is made
- a previous decision is superseded

Update `worklog/` when:

- a milestone lands
- the repo gains a new subsystem or capability

Update `test-results/` when:

- a meaningful validation run is completed
- benchmark or quality trends should be preserved in-repo

## Reading Order

1. `knowledge/product.md`
2. `knowledge/architecture.md`
3. `knowledge/repo-map.md`
4. `plans/active/`
5. `todos/`
6. `test-results/`
7. `install/`
