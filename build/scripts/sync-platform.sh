#!/usr/bin/env bash
#
# Publish the vendored platform/ sub-projects OUT to their standalone repos.
#
# platform/<name>/ is the source of truth: the sub-projects are developed here,
# in the monorepo, alongside the Go code they talk to. This pushes each one to a
# local checkout of its public repo so the standalone repos stay in step.
#
# Only files tracked by THIS repo are published, so a sub-project's build output
# and node_modules (gitignored under platform/) never leak into the public repo.
#
#   make sync-platform                       # write, review, commit yourself
#   make sync-platform PLATFORM_COMMIT=1     # also commit and push
#   make sync-platform PLATFORM=osmedeus-registry
#   make sync-platform PLATFORM_DEST=/path/to/checkouts
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Default to the directory holding this repo, matching sync-skills' `../` default,
# rather than hardcoding a developer-specific path.
DEST_BASE="${PLATFORM_DEST:-$(cd "$REPO_ROOT/.." && pwd)}"
GIT_ORG="${PLATFORM_GIT_ORG:-git@github.com:osmedeus}"
DO_COMMIT="${PLATFORM_COMMIT:-}"

ALL_PLATFORMS=(osmedeus-dashboard osmedeus-registry osmedeus-workflow)
if [ -n "${PLATFORM:-}" ]; then
    PLATFORMS=("$PLATFORM")
else
    PLATFORMS=("${ALL_PLATFORMS[@]}")
fi

cd "$REPO_ROOT"

for name in "${PLATFORMS[@]}"; do
    src="$REPO_ROOT/platform/$name"
    dest="$DEST_BASE/$name"

    if [ ! -d "$src" ]; then
        echo "SKIP: platform/$name does not exist"
        continue
    fi

    echo "==> Syncing $name"

    if [ ! -d "$dest" ]; then
        echo "    Cloning $GIT_ORG/$name.git -> $dest"
        git clone -q "$GIT_ORG/$name.git" "$dest"
    fi

    if [ ! -d "$dest/.git" ]; then
        echo "    ERROR: $dest is not a git checkout, refusing to write"
        exit 1
    fi

    if [ "$(git ls-files "platform/$name" | wc -l | tr -d ' ')" -eq 0 ]; then
        echo "    ERROR: no tracked files under platform/$name, refusing to sync"
        exit 1
    fi

    # --delete propagates removals. Excluded paths are protected on the receiving
    # side too, so the destination keeps its own .git, node_modules and build
    # output — these are the same paths gitignored under platform/ here.
    rsync -a --delete \
        --exclude='.git/' \
        --exclude='node_modules/' \
        --exclude='.next/' \
        --exclude='build/' \
        --exclude='out/' \
        --exclude='coverage/' \
        --exclude='*.tsbuildinfo' \
        --exclude='.env*' \
        --exclude='.DS_Store' \
        "$src/" "$dest/"

    if [ -z "$DO_COMMIT" ]; then
        changed=$(git -C "$dest" status --porcelain | wc -l | tr -d ' ')
        echo "    $changed change(s) staged in $dest - review and commit there"
        continue
    fi

    if [ -n "$(git -C "$dest" status --porcelain)" ]; then
        git -C "$dest" add -A
        git -C "$dest" commit -q -m "sync: update from osmedeus monorepo ($(date +%Y-%m-%d))"
        git -C "$dest" push -q origin
        echo "    Committed and pushed $name"
    else
        echo "    No changes for $name"
    fi
done

echo "==> Done"
