# Bootstrap V0 Plan

## Goal

Turn the current scaffold into a minimal but real LLM-Wiki service with PostgreSQL-backed namespace and document primitives.

## Scope

- service bootstrap
- first DB connection
- first migrations
- first document-oriented API surface
- thin CLI integration over HTTP
- repeatable test harnesses and stored validation results

## Current Steps

1. Done: initialize Go repo skeleton
2. Done: initialize docs system of record
3. Done: choose direct `pgx` SQL for v0
4. Done: add PostgreSQL connection and readiness checks
5. Done: create initial schema and migrations
6. Done: add namespace CRUD
7. Done: add document create/get/update with revision tracking
8. Done: add tests for bootstrap and first routes
9. Done: add e2e and perf harnesses with docs-backed result storage
10. Done: extend CRUD surface with list and lookup APIs
11. In progress: harden validation and error model
12. Done: add document archive lifecycle operation
13. Done: add namespace archive lifecycle operation
14. Done: add active-vs-archived listing filters
15. Done: add explicit spaces API
16. Done: add a minimal Gin-served HTML management page
17. Next: add UI actions beyond create-only flows

## Notes

- Keep v0 narrow.
- Do not build search, approval flow, or complex patching before the base document model is stable.
