#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

required_files=(
  "README.md"
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
)

for file in "${required_files[@]}"; do
  if [[ ! -f "$file" ]]; then
    echo "check-template: missing $file" >&2
    exit 1
  fi
done

if grep -R "agent-dev-scaffolding" backend/pom.xml frontend/package.json backend/src/main/resources/application.yml backend/src/main/java/com/anjing/config/properties/FeatureProperties.java >/dev/null; then
  echo "check-template: old agent-dev-scaffolding identity remains in active project metadata" >&2
  exit 1
fi

if grep -R "m1.apifoxmock.com" frontend/.env.production frontend/src/api >/dev/null; then
  echo "check-template: production/runtime API must not point to Apifox mock" >&2
  exit 1
fi

echo "check-template: ok"
