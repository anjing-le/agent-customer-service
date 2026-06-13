#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "quality-gate: template and contract checks"
./scripts/check-template.sh
./scripts/check-contracts.sh

echo "quality-gate: backend package"
go test ./...

echo "quality-gate: frontend build"
pnpm build:console

echo "quality-gate: ok"
