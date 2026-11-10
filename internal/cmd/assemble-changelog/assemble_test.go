package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const header = "# Changelog\n\n## [Unreleased]\n"

func writeTree(t *testing.T, changelog string, frags map[string]string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	cl := filepath.Join(dir, "CHANGELOG.md")
	if err := os.WriteFile(cl, []byte(changelog), 0o600); err != nil {
		t.Fatal(err)
	}
	fd := filepath.Join(dir, "changelog.d")
	if err := os.MkdirAll(fd, 0o750); err != nil {
		t.Fatal(err)
	}
	for name, body := range frags {
		if err := os.WriteFile(filepath.Join(fd, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return cl, fd
}

func TestAssemble(t *testing.T) {
	tests := []struct {
		name      string
		changelog string
		frags     map[string]string
		want      string
	}{
		{
			name:      "fragments only, in name order, blank lines trimmed",
			changelog: header + "\n## [0.0.1-alpha] - 2026-01-01\n\n- old\n",
			frags: map[string]string{
				"20-b.md":   "\n\n### B\n\n- b\n\n\n",
				"10-a.md":   "### A\n- a",
				"notes.txt": "ignored",
			},
			want: header + "\n## [0.1.0-alpha] - 2026-09-25\n\n### A\n- a\n\n### B\n\n- b\n\n## [0.0.1-alpha] - 2026-01-01\n\n- old\n",
		},
		{
			name:      "unreleased body and fragments",
			changelog: header + "\n\n- pending\n\n",
			frags:     map[string]string{"1-x.md": "- frag\n"},
			want:      header + "\n## [0.1.0-alpha] - 2026-09-25\n\n- pending\n\n- frag\n",
		},
		{
			name:      "unreleased body only",
			changelog: header + "- pending\n## [0.0.1-alpha] - 2026-01-01\n",
			want:      header + "\n## [0.1.0-alpha] - 2026-09-25\n\n- pending\n\n## [0.0.1-alpha] - 2026-01-01\n",
		},
		{
			name:      "headings inside fences are content",
			changelog: header + "````md\n```\n## [0.0.9-alpha]\n````\n~~~\n## [0.1.0-alpha]\n~~~\n## [0.0.1-alpha]\n",
			want:      header + "\n## [0.1.0-alpha] - 2026-09-25\n\n````md\n```\n## [0.0.9-alpha]\n````\n~~~\n## [0.1.0-alpha]\n~~~\n\n## [0.0.1-alpha]\n",
		},
		{
			name:      "backtick info string with a backtick opens no fence",
			changelog: header + "``` a`b\n## [0.0.1-alpha]\n",
			want:      header + "\n## [0.1.0-alpha] - 2026-09-25\n\n``` a`b\n\n## [0.0.1-alpha]\n",
		},
		{
			name:      "no trailing newline",
			changelog: "## [Unreleased]\n- x",
			want:      "## [Unreleased]\n\n## [0.1.0-alpha] - 2026-09-25\n\n- x\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, fd := writeTree(t, tt.changelog, tt.frags)
			got, err := Assemble(cl, "0.1.0-alpha", "2026-09-25", fd)
			if err != nil {
				t.Fatalf("Assemble: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestAssembleRefusals(t *testing.T) {
	tests := []struct {
		name      string
		changelog string
		frags     map[string]string
		version   string
		date      string
		reason    string
	}{
		{"no unreleased heading", "# Changelog\n- x\n", nil, "0.1.0-alpha", "2026-09-25", `no "## [Unreleased]" heading`},
		{"unreleased only inside a fence", "```\n## [Unreleased]\n```\n", nil, "0.1.0-alpha", "2026-09-25", `no "## [Unreleased]" heading`},
		{"empty fragment", header, map[string]string{"1-x.md": " \n\t\n"}, "0.1.0-alpha", "2026-09-25", "is empty"},
		{"existing section with pending entries", header + "- p\n## [0.1.0-alpha] - 2026-01-01\n", nil, "0.1.0-alpha", "2026-09-25", "already has"},
		{"empty version", header, nil, "", "2026-09-25", "empty version"},
		{"bad date", header, nil, "0.1.0-alpha", "25/09/2026", "date is not YYYY-MM-DD"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, fd := writeTree(t, tt.changelog, tt.frags)
			_, err := Assemble(cl, tt.version, tt.date, fd)
			var re *RefusalError
			if !errors.As(err, &re) {
				t.Fatalf("err = %v, want *RefusalError", err)
			}
			if !strings.Contains(re.Error(), tt.reason) {
				t.Fatalf("err = %q, want it to contain %q", re.Error(), tt.reason)
			}
		})
	}
}

func TestAssembleNothingPending(t *testing.T) {
	cl, fd := writeTree(t, header+"\n  \n## [0.1.0-alpha] - 2026-01-01\n", nil)
	if _, err := Assemble(cl, "0.1.0-alpha", "2026-09-25", fd); !errors.Is(err, ErrNothingPending) {
		t.Fatalf("err = %v, want ErrNothingPending", err)
	}
}

func TestAssembleMissingChangelog(t *testing.T) {
	_, err := Assemble(filepath.Join(t.TempDir(), "absent.md"), "0.1.0-alpha", "2026-09-25", t.TempDir())
	var re *RefusalError
	if !errors.As(err, &re) || !strings.Contains(re.Reason, "no changelog at") {
		t.Fatalf("err = %v, want a no-changelog refusal", err)
	}
}

func TestAssembleUnreadableChangelog(t *testing.T) {
	dir := t.TempDir() // a directory is not a readable file
	_, err := Assemble(dir, "0.1.0-alpha", "2026-09-25", t.TempDir())
	var rerr *ReadError
	if !errors.As(err, &rerr) || rerr.Path != dir || errors.Unwrap(rerr) == nil {
		t.Fatalf("err = %v, want *ReadError for %s", err, dir)
	}
}

func TestRunExitCodes(t *testing.T) {
	cl, fd := writeTree(t, header+"- p\n", nil)
	var out, errOut bytes.Buffer
	if code := run([]string{cl, "0.1.0-alpha", "2026-09-25", fd}, &out, &errOut); code != exitAssembled {
		t.Fatalf("exit %d, stderr %q", code, errOut.String())
	}
	if !strings.Contains(out.String(), "## [0.1.0-alpha] - 2026-09-25") {
		t.Fatalf("stdout lacks the section:\n%s", out.String())
	}

	cl, fd = writeTree(t, header, nil)
	out.Reset()
	if code := run([]string{cl, "0.1.0-alpha", "2026-09-25", fd}, &out, &errOut); code != exitNothing || out.Len() != 0 {
		t.Fatalf("exit %d with %d bytes of output, want %d and none", code, out.Len(), exitNothing)
	}

	errOut.Reset()
	if code := run([]string{"too", "few"}, &out, &errOut); code != exitRefused || !strings.Contains(errOut.String(), "usage") {
		t.Fatalf("exit %d, stderr %q; want a usage refusal", code, errOut.String())
	}

	errOut.Reset()
	if code := run([]string{cl, "", "2026-09-25", fd}, &out, &errOut); code != exitRefused || !strings.Contains(errOut.String(), "empty version") {
		t.Fatalf("exit %d, stderr %q; want an empty-version refusal", code, errOut.String())
	}
}
