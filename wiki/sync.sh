#!/usr/bin/env bash
# Sync wiki/*.md to the GitHub wiki repo.
#
# Usage:
#   ./wiki/sync.sh            # commit + push
#   DRY_RUN=1 ./wiki/sync.sh  # commit only, no push (still touches a temp dir)
#
# Prereqs: the wiki must be bootstrapped via the GitHub UI once (create any
# first page). See wiki/README.md.

set -euo pipefail

REPO_OWNER="dipjyotimetia"
REPO_NAME="openapi-go-mcp"
WIKI_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}.wiki.git"

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
work_dir="$(mktemp -d -t openapi-go-mcp-wiki-XXXXXX)"
trap 'rm -rf "${work_dir}"' EXIT

echo "→ Cloning ${WIKI_URL}"
if ! git clone --depth 1 "${WIKI_URL}" "${work_dir}/wiki" 2>/dev/null; then
    cat >&2 <<EOF
Failed to clone the wiki repo. The most likely cause is that the wiki has
not been initialized yet.

To initialize:
  1. Open https://github.com/${REPO_OWNER}/${REPO_NAME}/wiki
  2. Click "Create the first page" and save any placeholder content.
  3. Re-run this script.
EOF
    exit 1
fi

echo "→ Copying wiki sources"
# Remove every tracked .md from the wiki except .git internals; rsync from source.
find "${work_dir}/wiki" -maxdepth 1 -type f -name '*.md' -delete
cp "${script_dir}"/*.md "${work_dir}/wiki/"

cd "${work_dir}/wiki"

if [[ -z "$(git status --porcelain)" ]]; then
    echo "✓ Wiki is already up to date."
    exit 0
fi

git add -A
commit_msg="docs(wiki): sync from main @ $(cd "${script_dir}/.." && git rev-parse --short HEAD)"
git commit -m "${commit_msg}"

if [[ "${DRY_RUN:-}" == "1" ]]; then
    echo "✓ DRY_RUN=1 — committed locally in ${work_dir}/wiki, not pushing."
    echo "  Inspect with: git -C ${work_dir}/wiki log -1 -p"
    trap - EXIT
    exit 0
fi

echo "→ Pushing to ${WIKI_URL}"
git push origin HEAD
echo "✓ Wiki synced."
