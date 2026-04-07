#!/usr/bin/env bash
set -euo pipefail

mkdir -p docs/test-results/unit
out="docs/test-results/unit/$(date +%Y%m%d-%H%M%S)-unit.txt"
packages=$(go list ./... | grep -v '/tests/e2e$')
go test $packages -run Test -count=1 | tee "$out"
echo "unit results saved to $out"
