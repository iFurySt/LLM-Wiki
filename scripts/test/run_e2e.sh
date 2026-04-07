#!/usr/bin/env bash
set -euo pipefail

mkdir -p docs/test-results/e2e
out="docs/test-results/e2e/$(date +%Y%m%d-%H%M%S)-e2e.txt"
go test ./tests/e2e -count=1 | tee "$out"
echo "e2e results saved to $out"

