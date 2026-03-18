---
name: bump-version
description: Bump the version and release. Use when the user says /bump-version, "bump version", "cut a release", or "release".
---

# Bump Version

Analyze the changes since the last release tag and cut an appropriate release.

## Steps

1. Find the latest version tag: `git tag --sort=-v:refname | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | head -1`
2. Review the commits since that tag: `git log <latest-tag>..HEAD --oneline`
3. Determine the bump type based on the changes:
   - **major**: breaking changes (removed commands, changed CLI flags/behavior, renamed packages, changed install paths)
   - **minor**: new features (new commands, new flags, new capabilities)
   - **patch**: bug fixes, documentation, refactoring, test changes, dependency updates
4. Tell the user what you found and what bump type you're choosing, with a brief explanation.
5. Run `./scripts/release.sh <type>` to tag and push. The script will prompt for confirmation.
