#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

expected_name="${ANJING_DELIVERY_USER_NAME:-安静}"
expected_email="${ANJING_DELIVERY_USER_EMAIL:-245548353+anjing-le@users.noreply.github.com}"
remote="${ANJING_DELIVERY_REMOTE:-origin}"
allow_dirty="${ANJING_DELIVERY_ALLOW_DIRTY:-false}"

fail() {
  echo "check-teaching-delivery: $*" >&2
  exit 1
}

require_file() {
  local file="$1"
  [[ -f "$file" ]] || fail "missing $file"
}

require_grep() {
  local pattern="$1"
  local file="$2"
  grep -q "$pattern" "$file" || fail "missing pattern '$pattern' in $file"
}

if [[ "$allow_dirty" != "true" && -n "$(git status --porcelain)" ]]; then
  fail "worktree is not clean; commit or stash changes before delivery check"
fi

actual_name="$(git config --get user.name || true)"
actual_email="$(git config --get user.email || true)"
[[ "$actual_name" == "$expected_name" ]] || fail "git user.name is '$actual_name', expected '$expected_name'"
[[ "$actual_email" == "$expected_email" ]] || fail "git user.email is '$actual_email', expected '$expected_email'"

author_name="$(git log -1 --pretty=%an)"
author_email="$(git log -1 --pretty=%ae)"
committer_name="$(git log -1 --pretty=%cn)"
committer_email="$(git log -1 --pretty=%ce)"
[[ "$author_name" == "$expected_name" ]] || fail "latest author name is '$author_name', expected '$expected_name'"
[[ "$author_email" == "$expected_email" ]] || fail "latest author email is '$author_email', expected '$expected_email'"
[[ "$committer_name" == "$expected_name" ]] || fail "latest committer name is '$committer_name', expected '$expected_name'"
[[ "$committer_email" == "$expected_email" ]] || fail "latest committer email is '$committer_email', expected '$expected_email'"

head_sha="$(git rev-parse HEAD)"
origin_main_sha="$(git rev-parse refs/remotes/origin/main 2>/dev/null || true)"
origin_master_sha="$(git rev-parse refs/remotes/origin/master 2>/dev/null || true)"
[[ -n "$origin_main_sha" ]] || fail "missing refs/remotes/origin/main; run git fetch first"
[[ -n "$origin_master_sha" ]] || fail "missing refs/remotes/origin/master; run git fetch first"
[[ "$head_sha" == "$origin_main_sha" ]] || fail "HEAD ($head_sha) does not match origin/main ($origin_main_sha)"
[[ "$head_sha" == "$origin_master_sha" ]] || fail "HEAD ($head_sha) does not match origin/master ($origin_master_sha)"

remote_refs="$(git ls-remote "$remote" refs/heads/main refs/heads/master)"
remote_main_sha="$(printf '%s\n' "$remote_refs" | awk '$2 == "refs/heads/main" { print $1 }')"
remote_master_sha="$(printf '%s\n' "$remote_refs" | awk '$2 == "refs/heads/master" { print $1 }')"
[[ -n "$remote_main_sha" ]] || fail "remote main not found on $remote"
[[ -n "$remote_master_sha" ]] || fail "remote master not found on $remote"
[[ "$head_sha" == "$remote_main_sha" ]] || fail "HEAD ($head_sha) does not match remote main ($remote_main_sha)"
[[ "$head_sha" == "$remote_master_sha" ]] || fail "HEAD ($head_sha) does not match remote master ($remote_master_sha)"

required_files=(
  "README.md"
  "apps/console/package.json"
  "cmd/platform-all/main.go"
  "contracts/api-contract.json"
  "contracts/service-boundaries.json"
  "infra/postgres/migrations/001_agent_customer_service.sql"
  "internal/platform/store/store.go"
  "project_document/DEMO_FLOW.md"
  "project_document/FINAL_ACCEPTANCE.md"
  "project_document/PRODUCTION_EXTENSION_GUIDE.md"
  "project_document/SCAFFOLD_INHERITANCE.md"
  "project_document/STATUS.md"
  "project_document/TEACHING_DELIVERY_CHECKLIST.md"
  "scripts/demo-classroom-local.sh"
  "scripts/demo-classroom-smoke.sh"
  "scripts/quality-gate.sh"
)

for file in "${required_files[@]}"; do
  require_file "$file"
done

require_grep "DVSkyFolding" "README.md"
require_grep "React + TypeScript + Vite" "README.md"
require_grep "Go + \`net/http\` / \`ServeMux\`" "README.md"
require_grep "PostgreSQL" "README.md"
require_grep "TEACHING_DELIVERY_CHECKLIST.md" "project_document/README.md"
require_grep "PRODUCTION_EXTENSION_GUIDE.md" "project_document/ROADMAP.md"
require_grep "完成度：约 96%" "project_document/STATUS.md"
require_grep "\"verify\": \"./scripts/quality-gate.sh\"" "package.json"
require_grep "\"demo:classroom\": \"./scripts/demo-classroom-smoke.sh\"" "package.json"
require_grep "\"demo:classroom:local\": \"./scripts/demo-classroom-local.sh\"" "package.json"
require_grep "\"check:delivery\": \"./scripts/check-teaching-delivery.sh\"" "package.json"

echo "check-teaching-delivery: ok"
