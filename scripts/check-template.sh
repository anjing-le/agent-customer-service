#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

required_files=(
  "README.md"
  "go.mod"
  "package.json"
  "pnpm-workspace.yaml"
  "apps/console/package.json"
  "cmd/platform-all/main.go"
  "internal/platform/store/store.go"
  "infra/postgres/migrations/001_agent_customer_service.sql"
  "infra/local/docker-compose.yml"
  "contracts/platform-contract.json"
  "contracts/service-boundaries.json"
  "project_document/README.md"
  "project_document/PROJECT_CONSTRAINTS.md"
  "project_document/ROADMAP.md"
  "project_document/STATUS.md"
  "project_document/API_CONTRACT_GUIDE.md"
  "project_document/SERVICE_BOUNDARY_GUIDE.md"
  "project_document/SCAFFOLD_INHERITANCE.md"
  "project_document/LOCAL_STARTUP_GUIDE.md"
  "project_document/FINAL_ACCEPTANCE.md"
  "project_document/TEACHING_DELIVERY_CHECKLIST.md"
  "project_document/TEACHING_TALK_TRACK.md"
  "scripts/demo-classroom-local.sh"
  "scripts/check-teaching-delivery.sh"
)

for file in "${required_files[@]}"; do
  if [[ ! -f "$file" ]]; then
    echo "check-template: missing $file" >&2
    exit 1
  fi
done

if grep -R "Spring Boot\\|Vue 3\\|Element Plus\\|JPA\\|MySQL" README.md project_document contracts >/dev/null; then
  echo "check-template: old Java/Vue runtime wording remains in core docs" >&2
  exit 1
fi

if grep -R "m1.apifoxmock.com" apps/console/src contracts project_document README.md >/dev/null; then
  echo "check-template: production/runtime API must not point to Apifox mock" >&2
  exit 1
fi

echo "check-template: ok"
