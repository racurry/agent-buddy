#!/usr/bin/env bash
set -euo pipefail

usage() {
    echo "Usage: ./scripts/release.sh [major|minor|patch]"
    exit 1
}

[[ $# -ne 1 ]] && usage

bump="$1"
[[ "$bump" != "major" && "$bump" != "minor" && "$bump" != "patch" ]] && usage

# Get the latest version tag, default to v0.0.0
latest=$(git tag --sort=-v:refname | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | head -1)
latest="${latest:-v0.0.0}"

IFS='.' read -r major minor patch <<<"${latest#v}"

case "$bump" in
major)
    major=$((major + 1))
    minor=0
    patch=0
    ;;
minor)
    minor=$((minor + 1))
    patch=0
    ;;
patch) patch=$((patch + 1)) ;;
esac

next="v${major}.${minor}.${patch}"

echo "Current: $latest"
echo "Next:    $next"
echo ""
read -p "Tag and push $next? [y/N] " confirm
[[ "$confirm" != "y" && "$confirm" != "Y" ]] && echo "Aborted." && exit 0

git tag "$next"
git pub --tags
echo "Tagged and pushed $next — GitHub Actions will create the release."
