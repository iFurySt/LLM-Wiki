#!/usr/bin/env bash
set -euo pipefail

mkdir -p docs/test-results/perf
out="docs/test-results/perf/$(date +%Y%m%d-%H%M%S)-perf.txt"
go test ./tests/e2e -run '^$' -bench BenchmarkCreateAndGetDocument -benchmem -count=1 | tee "$out"
echo "perf results saved to $out"

