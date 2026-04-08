# 2026-04-07 Bootstrap Validation

## Commands

- Unit: `./scripts/test/run_unit.sh`
- E2E: `./scripts/test/run_e2e.sh`
- Perf: `./scripts/test/run_perf.sh`

## Result Files

- Unit: [20260407-124905-unit.txt](unit/20260407-124905-unit.txt)
- E2E: [20260407-125018-e2e.txt](e2e/20260407-125018-e2e.txt)
- Perf: [20260407-125005-perf.txt](perf/20260407-125005-perf.txt)

## Summary

- Unit tests passed.
- E2E tests passed against real PostgreSQL containers.
- The initial benchmark for create-plus-get document flow completed successfully.

## Notable Perf Baseline

- Benchmark: `BenchmarkCreateAndGetDocument`
- Result: `877` iterations, `1305579 ns/op`, `33237 B/op`, `411 allocs/op`

## Notes

- The first e2e attempt exposed an early Postgres startup race. The DB open path was hardened with retry-based pinging, and the subsequent e2e run passed.
