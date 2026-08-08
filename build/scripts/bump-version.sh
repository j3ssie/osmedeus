#!/usr/bin/env bash
set -euo pipefail

# Bump the single source-of-truth version in internal/core/constants.go.
#
# That `VERSION` constant drives everything downstream: the Makefile's VERSION,
# goreleaser builds, `osmedeus version`, the DefaultUA string, and the
# @j3ssie/osmedeus npm package + its per-platform sub-packages. npm versions are
# immutable, so every release needs a new, unique number here before
# `make npm-publish`.
#
# Usage (normally via `make bump-version`):
#   bump-version.sh [part]
#
# Env / make vars:
#   PART    = patch (default) | minor | major | pre | release
#   LABEL   = override the prerelease label (e.g. LABEL=beta -> -beta)
#   SET     = set an explicit version (e.g. SET=v5.1.0-rc.1); skips computation
#   DRY_RUN = 1 to preview the change without writing the file
#
# Examples (current -> new):
#   v5.0.3       PART=patch    ->  v5.0.4   (default)
#   v5.0.3       PART=minor    ->  v5.1.0
#   v5.0.3       PART=major    ->  v6.0.0
#   v5.0.3       LABEL=beta    ->  v5.0.4-beta   (default patch + label)
#   v5.0.4-beta  PART=pre      ->  v5.0.4-beta.1
#   v5.0.4-beta  PART=release  ->  v5.0.4        (drop prerelease)

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
VERSION_FILE="$ROOT/internal/core/constants.go"

PART="${PART:-${1:-patch}}"
LABEL="${LABEL:-}"
SET="${SET:-}"
DRY_RUN="${DRY_RUN:-}"

die()  { printf '\033[31m[!] %s\033[0m\n' "$*" >&2; exit 1; }
info() { printf '\033[36m[*]\033[0m %s\n' "$*"; }

[ -f "$VERSION_FILE" ] || die "version file not found: $VERSION_FILE"

current="$(grep -E '^[[:space:]]*VERSION[[:space:]]*=' "$VERSION_FILE" | head -1 | cut -d '"' -f 2)"
[ -n "$current" ] || die "could not parse VERSION from $VERSION_FILE"

# Optionally v-prefixed semver with an optional prerelease label.
semver_re='^v?([0-9]+)\.([0-9]+)\.([0-9]+)(-[0-9A-Za-z][0-9A-Za-z.-]*)?$'

if [ -n "$SET" ]; then
  new="$SET"
  [[ "$new" =~ $semver_re ]] || die "SET='$new' is not a valid (optionally v-prefixed) semver"
else
  [[ "$current" =~ $semver_re ]] || die "current version '$current' is not parseable semver"
  major="${BASH_REMATCH[1]}"
  minor="${BASH_REMATCH[2]}"
  patch="${BASH_REMATCH[3]}"
  label="${BASH_REMATCH[4]#-}"   # prerelease without the leading '-' ('' if none)

  case "$PART" in
    major)   major=$((major + 1)); minor=0; patch=0 ;;
    minor)   minor=$((minor + 1)); patch=0 ;;
    patch)   patch=$((patch + 1)) ;;
    release) label="" ;;          # promote to a stable (non-prerelease) version
    pre)
      [ -n "$label" ] || die "PART=pre needs an existing prerelease label; set one with LABEL=alpha"
      if [[ "$label" =~ ^(.*[^0-9.])\.?([0-9]+)$ ]]; then
        label="${BASH_REMATCH[1]}.$((BASH_REMATCH[2] + 1))"
      else
        label="${label}.1"
      fi
      ;;
    *) die "unknown PART='$PART' (use: patch|minor|major|pre|release)" ;;
  esac

  # LABEL overrides the prerelease label (ignored when PART=release clears it).
  if [ -n "$LABEL" ] && [ "$PART" != "release" ]; then
    label="$LABEL"
  fi

  new="v${major}.${minor}.${patch}"
  [ -n "$label" ] && new="${new}-${label}"
fi

# constants.go keeps a leading 'v' — normalize so SET= without it still matches.
case "$new" in v*) ;; *) new="v$new" ;; esac
[[ "$new" =~ $semver_re ]] || die "computed version '$new' failed validation"
[ "$new" != "$current" ] || die "version unchanged ($current) — nothing to bump (PART=$PART)"

info "version: $current  ->  $new"

if [ "$DRY_RUN" = "1" ]; then
  info "DRY_RUN=1 — $VERSION_FILE left unchanged"
  exit 0
fi

# Rewrite only the `VERSION = "..."` line. The new value is passed as ARGV[0]
# and shifted off in BEGIN so perl does not treat it as a file to edit.
perl -0pi -e 'BEGIN{$n=shift @ARGV} s/^(\s*VERSION\s*=\s*)"[^"]*"/$1"$n"/m' "$new" "$VERSION_FILE"

after="$(grep -E '^[[:space:]]*VERSION[[:space:]]*=' "$VERSION_FILE" | head -1 | cut -d '"' -f 2)"
[ "$after" = "$new" ] || die "rewrite failed (file still shows '$after')"

info "updated $VERSION_FILE"
info "next: review the diff & commit, then 'make npm-publish' (auto-rebuilds binaries for $new)"
