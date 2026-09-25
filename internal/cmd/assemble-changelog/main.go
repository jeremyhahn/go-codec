// Command assemble-changelog folds the pending changelog into a release
// section.
//
// Usage: go run ./internal/cmd/assemble-changelog <CHANGELOG.md> <X.Y.Z-alpha> <YYYY-MM-DD> <fragment-dir>
//
// release-versioning.md §5: a PR adds changelog.d/<issue>-<slug>.md instead of
// editing CHANGELOG.md, and the release step folds the fragments into the new
// version's section as a COMMITTED step before tagging. This program is the
// file transform only — it touches no git state. `make release-alpha` (the
// release-alpha-changelog prerequisite) runs it, removes the fragments and
// commits both through agent-git.sh.
//
// Pending entries are everything under "## [Unreleased]" followed by every
// <fragment-dir>/*.md in name order. They become
//
//	## [Unreleased]
//
//	## [X.Y.Z-alpha] - YYYY-MM-DD
//
//	<Unreleased content>
//
//	<fragments>
//
// Headings are recognized only outside fenced code blocks, with the same
// CommonMark fence rules as the release-alpha CHANGELOG guard.
//
// It writes the assembled file to stdout and exits:
//
//	0  assembled
//	3  nothing to assemble: no pending entries
//	1  refused, with the reason on stderr: no "## [Unreleased]" heading, an
//	   empty fragment, or the version already has a section while entries are
//	   still pending (releasing then would silently leave them out)
//
// It replaces scripts/assemble-changelog.sh: the public release ships no shell
// scripts (owner ruling on go-codec#26).
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	exitAssembled = 0
	exitRefused   = 1
	exitNothing   = 3
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 4 {
		_, _ = fmt.Fprintln(stderr, "assemble-changelog: usage: assemble-changelog <CHANGELOG.md> <X.Y.Z-alpha> <YYYY-MM-DD> <fragment-dir>")
		return exitRefused
	}
	out, err := Assemble(args[0], args[1], args[2], args[3])
	if errors.Is(err, ErrNothingPending) {
		return exitNothing
	}
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return exitRefused
	}
	if _, err := io.WriteString(stdout, out); err != nil {
		_, _ = fmt.Fprintln(stderr, "assemble-changelog: write output:", err)
		return exitRefused
	}
	return exitAssembled
}
