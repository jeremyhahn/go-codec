// Package releasetest holds the regression tests for this repository's
// `make release-alpha` target (issue #10). It has no production code: the
// tests copy the repository's own Makefile and the internal/cmd/assemble-changelog
// sources into a throwaway git repository with a local bare remote and a stub
// agent-git.sh on PATH, run the real target, and assert what reached the
// remote.
package releasetest
