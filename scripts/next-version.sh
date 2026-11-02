#!/usr/bin/env bash
set -euo pipefail
last=$(git describe --tags --abbrev=0 --match 'v[0-9]*.[0-9]*.[0-9]*' 2>/dev/null || true)
if [ -n "$last" ]; then
  [ "$(git rev-parse "$last^{commit}")" = "$(git rev-parse HEAD)" ] && exit 0
  log=$(git log --format=%B "$last..HEAD")
else
  log=$(git log --format=%B)
fi
if grep -Eq '^([a-z]+(\([^)]*\))?!:|BREAKING CHANGE)' <<<"$log"; then
  bump=major
elif grep -Eq '^feat(\([^)]*\))?:' <<<"$log"; then
  bump=minor
elif grep -Eq '^(fix|perf|revert)(\([^)]*\))?:' <<<"$log"; then
  bump=patch
else
  exit 0
fi
if [ -z "$last" ]; then
  echo 0.1.0
  exit
fi
IFS=. read -r major minor patch <<<"${last#v}"
case $bump in
  major) echo "$((major + 1)).0.0" ;;
  minor) echo "$major.$((minor + 1)).0" ;;
  patch) echo "$major.$minor.$((patch + 1))" ;;
esac
