#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

go test ./internal/platform/store -run TestQualityEvaluationCases -count=1

echo "check-agent-quality: ok"
