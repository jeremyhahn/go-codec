// Command releasability is the public-releasability gate: it fails when a
// tracked file carries a lab address, a fleet hostname, an owner path or an
// internal domain name, when a shell script or agent context is tracked, and
// when a tracked binary or oversize file is not listed in
// releasability-exempt.txt (see Scan).
//
// Usage:
//
//	go run ./internal/cmd/releasability [-C <repo>]   # scan git's tracked files
//	go run ./internal/cmd/releasability -hash <name>  # print a fleetHosts entry
//
// The published GitHub tree is byte-identical to the Gitea release tag's tree
// (version-control.md §5a), so anything internal must be scrubbed in the
// source before tagging; this gate is what proves the scrub. It reads the list
// of tracked files from git, so an untracked local file (such as an agent's
// CLAUDE.md) is not in scope, and it needs a git work tree to run.
//
// Exit status: 0 clean, 1 findings (one per line on stdout), 2 the scan could
// not run.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

const (
	exitClean    = 0
	exitFindings = 1
	exitError    = 2
)

// GitError is a failure to list the tracked files.
type GitError struct {
	Dir    string
	Stderr string
	Err    error
}

func (e *GitError) Error() string {
	msg := "releasability: git ls-files in " + e.Dir + ": " + e.Err.Error()
	if s := strings.TrimSpace(e.Stderr); s != "" {
		msg += ": " + s
	}
	return msg
}

func (e *GitError) Unwrap() error { return e.Err }

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("releasability", flag.ContinueOnError)
	fl.SetOutput(stderr)
	dir := fl.String("C", ".", "repository `dir`ectory to scan")
	hash := fl.String("hash", "", "print the fleetHosts digest for a hostname `name` and exit")
	if err := fl.Parse(args); err != nil {
		return exitError
	}
	if fl.NArg() != 0 {
		_, _ = fmt.Fprintln(stderr, "releasability: unexpected arguments:", strings.Join(fl.Args(), " "))
		return exitError
	}
	if *hash != "" {
		_, _ = fmt.Fprintln(stdout, HostHash(*hash))
		return exitClean
	}
	paths, err := trackedFiles(*dir)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return exitError
	}
	findings, err := Scan(os.DirFS(*dir), paths)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return exitError
	}
	for _, f := range findings {
		if f.Line == 0 { // a finding about the whole file
			_, _ = fmt.Fprintf(stdout, "%s: %s: %s\n", f.Path, f.Kind, f.Match)
			continue
		}
		_, _ = fmt.Fprintf(stdout, "%s:%d: %s: %s\n", f.Path, f.Line, f.Kind, f.Match)
	}
	if len(findings) > 0 {
		_, _ = fmt.Fprintf(stderr, "releasability: %d finding(s) in %d tracked file(s) scanned; scrub them before tagging (version-control.md §5a)\n", len(findings), len(paths))
		return exitFindings
	}
	_, _ = fmt.Fprintf(stdout, "releasability: %d tracked file(s) clean\n", len(paths))
	return exitClean
}

// trackedFiles lists the files in git's index, relative to dir.
func trackedFiles(dir string) ([]string, error) {
	cmd := exec.Command("git", "-C", dir, "ls-files", "-z") // #nosec G204 -- fixed git subcommand; dir is the operator's -C
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		return nil, &GitError{Dir: dir, Stderr: errOut.String(), Err: err}
	}
	var paths []string
	for _, p := range strings.Split(out.String(), "\x00") {
		if p != "" {
			paths = append(paths, p)
		}
	}
	if len(paths) == 0 {
		return nil, &GitError{Dir: dir, Err: errNoTrackedFiles}
	}
	return paths, nil
}

// errNoTrackedFiles refuses a vacuous pass: a scan of nothing proves nothing.
var errNoTrackedFiles = errors.New("no tracked files")
