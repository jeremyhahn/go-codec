package main

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// ErrNothingPending reports that there is nothing to assemble: no content
// under "## [Unreleased]" and no fragment. The version's heading may or may not
// exist already; release-alpha's own CHANGELOG guard decides that.
var ErrNothingPending = errors.New("assemble-changelog: nothing to assemble")

// RefusalError is a refusal to assemble, with the reason for the operator.
type RefusalError struct {
	Reason string
}

func (e *RefusalError) Error() string { return "assemble-changelog: " + e.Reason }

// ReadError is a failure to read an input file.
type ReadError struct {
	Path string
	Err  error
}

func (e *ReadError) Error() string {
	return "assemble-changelog: read " + e.Path + ": " + e.Err.Error()
}

func (e *ReadError) Unwrap() error { return e.Err }

var (
	datePattern       = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)
	unreleasedPattern = regexp.MustCompile(`^## \[Unreleased\][ \t]*$`)
	nonBlankPattern   = regexp.MustCompile(`[^ \t]`)
)

// Assemble reads the changelog at path and the fragments in dir and returns the
// assembled changelog. It touches nothing on disk.
func Assemble(path, version, date, dir string) (string, error) {
	if version == "" {
		return "", &RefusalError{Reason: "empty version"}
	}
	if !datePattern.MatchString(date) {
		return "", &RefusalError{Reason: "date is not YYYY-MM-DD: " + date}
	}
	b, err := os.ReadFile(path) // #nosec G304 G703 -- the operator names the changelog on the command line
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", &RefusalError{Reason: "no changelog at " + path}
		}
		return "", &ReadError{Path: path, Err: err}
	}
	frags, err := readFragments(dir)
	if err != nil {
		return "", err
	}
	return assembleText(string(b), version, date, frags)
}

// readFragments returns the lines of every dir/*.md in name order, each with
// its leading and trailing blank lines stripped, separated by exactly one
// empty line.
func readFragments(dir string) ([]string, error) {
	names, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return nil, &RefusalError{Reason: "bad fragment directory " + dir}
	}
	sort.Strings(names)
	var out []string
	for i, name := range names {
		b, err := os.ReadFile(name) // #nosec G304 G703 -- a fragment inside the operator-named directory
		if err != nil {
			return nil, &ReadError{Path: name, Err: err}
		}
		if strings.TrimSpace(string(b)) == "" {
			return nil, &RefusalError{Reason: "fragment " + name + " is empty"}
		}
		if i > 0 {
			out = append(out, "")
		}
		out = append(out, trimBlankLines(splitLines(string(b)))...)
	}
	return out, nil
}

// splitLines splits text into lines the way awk reads records: a final
// newline does not start another, empty, line.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}

// trimBlankLines drops leading and trailing lines that hold only whitespace.
func trimBlankLines(lines []string) []string {
	blank := func(l string) bool { return strings.TrimSpace(l) == "" }
	for len(lines) > 0 && blank(lines[0]) {
		lines = lines[1:]
	}
	for len(lines) > 0 && blank(lines[len(lines)-1]) {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// fence reports the fence run length of line (0 when it is no fence), the
// fence character, and the info string after the run. CommonMark: 0-3 spaces
// of indent, then 3+ backticks or tildes.
func fence(line string) (n int, ch byte, info string) {
	i := 0
	for i < 3 && i < len(line) && line[i] == ' ' {
		i++
	}
	if i >= len(line) || (line[i] != '~' && line[i] != '`') {
		return 0, 0, ""
	}
	ch = line[i]
	for i+n < len(line) && line[i+n] == ch {
		n++
	}
	info = line[i+n:]
	if n < 3 {
		return 0, ch, info
	}
	return n, ch, info
}

// assembleText folds the [Unreleased] body and the fragments into a new
// "## [version] - date" section. Headings are recognized only outside fenced
// code blocks, with the same CommonMark fence rules as the release-alpha
// CHANGELOG guard.
func assembleText(changelog, version, date string, frags []string) (string, error) {
	ver := "## [" + version + "]"
	var pre, body, post []string
	state := 0 // 0 before [Unreleased], 1 in its body, 2 after it
	open, openCh := 0, byte(0)
	hasVer := false
	for _, line := range splitLines(changelog) {
		heading := false
		n, ch, info := fence(line)
		switch {
		case open > 0:
			if n >= open && ch == openCh && strings.Trim(info, " \t") == "" {
				open = 0
			}
		case n > 0 && (ch != '`' || !strings.Contains(info, "`")):
			open, openCh = n, ch
		default:
			heading = strings.HasPrefix(line, "## ")
			if strings.HasPrefix(line, ver) {
				hasVer = true
			}
		}
		switch {
		case state == 0:
			pre = append(pre, line)
			if heading && unreleasedPattern.MatchString(line) {
				state = 1
			}
		case state == 1 && heading:
			state = 2
			post = append(post, line)
		case state == 1:
			body = append(body, line)
		default:
			post = append(post, line)
		}
	}
	if state == 0 {
		return "", &RefusalError{Reason: `no "## [Unreleased]" heading outside a code block`}
	}
	first, last := -1, -1
	for i, l := range body {
		if nonBlankPattern.MatchString(l) {
			if first < 0 {
				first = i
			}
			last = i
		}
	}
	pending := first >= 0 || len(frags) > 0
	if hasVer && pending {
		return "", &RefusalError{Reason: `CHANGELOG.md already has a "` + ver + `" section, yet entries are still pending under [Unreleased] or in changelog.d; releasing would leave them out. Fold them in by hand or pick the next version`}
	}
	if !pending {
		return "", ErrNothingPending
	}
	out := append([]string{}, pre...)
	out = append(out, "", ver+" - "+date, "")
	if first >= 0 {
		out = append(out, body[first:last+1]...)
	}
	if first >= 0 && len(frags) > 0 {
		out = append(out, "")
	}
	out = append(out, frags...)
	if len(post) > 0 {
		out = append(out, "")
		out = append(out, post...)
	}
	return strings.Join(out, "\n") + "\n", nil
}
