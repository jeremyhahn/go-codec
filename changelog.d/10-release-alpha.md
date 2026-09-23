### Changed — `make release-alpha` replaced with handbook template v2.3.9 (#10)

- The version is now explicit: `make release-alpha VERSION=vX.Y.Z-alpha`. The
  old recipe bumped PATCH only, made lightweight unsigned tags as the local git
  identity, pushed every local tag with `--tags` to a `dev` remote this repo
  does not have, and used `|| true` to hide a failed VERSION commit.
- The new recipe signs every git write through `agent-git.sh` (found on
  `PATH`, or pass `AGENT_GIT=`) and creates an annotated tag. It pushes only the
  release commit and that one tag to `origin`, atomically. It releases only
  from `main`, and it refuses to cut a version that ends up with no
  `## [X.Y.Z-alpha]` section in `CHANGELOG.md`.
- Releasing now also assembles the changelog, as `release-versioning.md` §5
  requires: before tagging, `make release-alpha` moves everything under
  `## [Unreleased]` and every `changelog.d/*.md` fragment into a dated
  `## [X.Y.Z-alpha]` section (`scripts/assemble-changelog.sh`). It then deletes
  the fragments and commits the result through `agent-git.sh`. It refuses when
  the version already has a section but entries are still pending.
- Regression tests for the recipe (`internal/releasetest`) run this repo's
  own Makefile against a throwaway remote.
- A `VERSION=` value containing a newline is refused before anything is
  committed. The template's line-oriented `grep` check accepted
  `vX.Y.Z-alpha` followed by a newline and more text.
