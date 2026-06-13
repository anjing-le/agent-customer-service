#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

go test ./internal/platform/store -run TestAgentRegressionCases -count=1

echo "check-agent-regression: ok"
