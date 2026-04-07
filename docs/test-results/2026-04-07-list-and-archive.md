# 2026-04-07 List And Archive Validation

## Commands

- Unit: `./scripts/test/run_unit.sh`
- E2E: `./scripts/test/run_e2e.sh`
- Perf: `./scripts/test/run_perf.sh`

## Result Files

- Unit: [20260407-142757-unit.txt](/Users/ifuryst/projects/github/DocMesh/docs/test-results/unit/20260407-142757-unit.txt)
- E2E: [20260407-142823-e2e.txt](/Users/ifuryst/projects/github/DocMesh/docs/test-results/e2e/20260407-142823-e2e.txt)
- Perf: [20260407-142840-perf.txt](/Users/ifuryst/projects/github/DocMesh/docs/test-results/perf/20260407-142840-perf.txt)

## Summary

- Unit tests passed after adding validation coverage.
- E2E tests passed after adding document listing, slug lookup, and archive flow coverage.
- Benchmark still passes after API surface expansion.

## Notable Perf Baseline

- Benchmark: `BenchmarkCreateAndGetDocument`
- Result: `715` iterations, `1510981 ns/op`, `36688 B/op`, `416 allocs/op`

## Notes

- During this round, a Gin wildcard route conflict surfaced between `/v1/namespaces/:id` and a nested slug route. The slug lookup API was changed to `/v1/document-by-slug?namespace_id=...&slug=...` to avoid ambiguous router shape.
