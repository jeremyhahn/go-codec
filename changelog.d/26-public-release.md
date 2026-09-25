### Release notes — 1.5.0 is the first public release since 1.3.0 (#26)

- `v1.4.0` was cut on the internal forge only and never published to GitHub,
  whose newest tag is `v1.3.0`. Public consumers therefore move from 1.3.0
  straight to 1.5.0 and receive 1.4.0's changes with it.
- **BREAKING for consumers of `v1.3.0`:** the Avro codec (`pkg/avro`, the
  `codec.Avro` constant and the `codec_avro` build tag) was removed in 1.4.0;
  see that section for the reasons. It shipped as a MINOR bump under the
  pre-publication exception (release-versioning.md §1). From this first public
  release on, a breaking change follows the §8 protocol and takes a MAJOR
  version.
- New public API since 1.3.0: `cbor.NewWithOptions` (#24).

### Changed — public-release readiness (#26)

- `scripts/assemble-changelog.sh` is replaced by a Go program,
  `internal/cmd/assemble-changelog`, with the same inputs, output and exit
  codes; `make release-alpha` builds and runs it. The repository ships no
  shell scripts.
- Added a public-releasability gate, `make releasability`
  (`internal/cmd/releasability`), run by `make ci` and by a new `Releasability`
  CI job that the CI Gate requires. It fails when a tracked file carries a
  private-use IP address, a company fleet hostname, a path inside a named
  user's home directory, or an `.internal`/`.home.arpa` domain name.
- Agent context (`.claude/`, `.serena/`, `CLAUDE.md`) is no longer tracked and
  is ignored, so it never reaches the published tree.
