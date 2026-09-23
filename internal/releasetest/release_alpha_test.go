package releasetest

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const (
	testVersion = "v0.1.0-alpha"
	testHeading = "## [0.1.0-alpha]"
	unreleased  = "## [Unreleased]"
)

// stubAgentGit stands in for agent-git.sh: it asserts the repo argument,
// records every call, can be told to fail the release commit, and otherwise
// runs plain git (signing is agent-git.sh's contract, not the recipe's).
const stubAgentGit = `#!/bin/sh
[ "$1" = go-codec ] || { echo "stub agent-git.sh: called for repo '$1', want go-codec" >&2; exit 97; }
shift
printf '%s\n' "$*" >> "$STUB_LOG"
if [ "${STUB_FAIL:-}" = commit-release ] && [ "$1" = commit ]; then
	case "$*" in *"chore(release)"*) echo "stub agent-git.sh: refusing the release commit" >&2; exit 1 ;; esac
fi
exec git "$@"
`

// fixture is one isolated release scenario: a working clone on main, its bare
// remote, and a private HOME/git config.
type fixture struct {
	t      *testing.T
	root   string
	work   string
	remote string
	bin    string
	log    string
	env    []string
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "Makefile")); err != nil {
		t.Fatalf("no Makefile at repo root %s: %v", root, err)
	}
	return root
}

func copyFile(t *testing.T, src, dst string, mode os.FileMode) {
	t.Helper()
	b, err := os.ReadFile(src) // #nosec G304 -- fixed paths inside this repository
	if err != nil {
		t.Fatalf("read %s: %v", src, err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		t.Fatalf("mkdir for %s: %v", dst, err)
	}
	if err := os.WriteFile(dst, b, mode); err != nil {
		t.Fatalf("write %s: %v", dst, err)
	}
}

func writeFile(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("mkdir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// newFixture builds the scenario with the given CHANGELOG.md and fragments
// (name -> content) committed and pushed, so the tree is clean and the remote
// is level with HEAD before the recipe runs.
func newFixture(t *testing.T, changelog string, fragments map[string]string) *fixture {
	t.Helper()
	for _, tool := range []string{"git", "make"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Fatalf("%s is required to test the release recipe: %v", tool, err)
		}
	}
	root := t.TempDir()
	f := &fixture{
		t:      t,
		root:   root,
		work:   filepath.Join(root, "work"),
		remote: filepath.Join(root, "remote.git"),
		bin:    filepath.Join(root, "bin"),
		log:    filepath.Join(root, "agent-git.log"),
	}
	home := filepath.Join(root, "home")
	if err := os.MkdirAll(home, 0o750); err != nil {
		t.Fatal(err)
	}
	var env []string
	for _, kv := range os.Environ() {
		switch strings.SplitN(kv, "=", 2)[0] {
		case "HOME", "GIT_CONFIG_GLOBAL", "GIT_CONFIG_NOSYSTEM", "PATH", "AGENT_GIT", "VERSION",
			"REPO_NAME", "RELEASE_REMOTE", "MAKEFLAGS", "MFLAGS", "MAKELEVEL", "STUB_FAIL", "STUB_LOG":
			continue
		}
		env = append(env, kv)
	}
	f.env = append(env,
		"HOME="+home,
		"GIT_CONFIG_GLOBAL="+filepath.Join(home, "gitconfig"),
		"GIT_CONFIG_NOSYSTEM=1",
		"PATH="+f.bin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"STUB_LOG="+f.log,
	)
	writeFile(t, filepath.Join(f.bin, "agent-git.sh"), stubAgentGit, 0o700)

	for _, kv := range [][2]string{
		{"user.name", "fixture"},
		{"user.email", "fixture@example.invalid"},
		{"init.defaultBranch", "main"},
		{"commit.gpgSign", "false"},
		{"tag.gpgSign", "false"},
	} {
		f.git(root, "config", "--global", kv[0], kv[1])
	}
	f.git(root, "init", "-q", "--bare", f.remote)
	f.git(root, "init", "-q", f.work)
	f.git(f.work, "remote", "add", "origin", f.remote)

	src := repoRoot(t)
	copyFile(t, filepath.Join(src, "Makefile"), filepath.Join(f.work, "Makefile"), 0o600)
	copyFile(t, filepath.Join(src, "scripts", "assemble-changelog.sh"), filepath.Join(f.work, "scripts", "assemble-changelog.sh"), 0o700)
	writeFile(t, filepath.Join(f.work, "VERSION"), "0.0.1-alpha\n", 0o600)
	writeFile(t, filepath.Join(f.work, "CHANGELOG.md"), changelog, 0o600)
	writeFile(t, filepath.Join(f.work, "changelog.d", ".gitkeep"), "", 0o600)
	for name, body := range fragments {
		writeFile(t, filepath.Join(f.work, "changelog.d", name), body, 0o600)
	}
	f.git(f.work, "add", "-A")
	f.git(f.work, "commit", "-q", "-m", "fixture")
	f.git(f.work, "push", "-q", "origin", "main")
	return f
}

func (f *fixture) run(dir, name string, args ...string) (string, error) {
	f.t.Helper()
	cmd := exec.Command(name, args...) // #nosec G204 -- test drives git/make with fixed arguments
	cmd.Dir = dir
	cmd.Env = f.env
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

func (f *fixture) git(dir string, args ...string) string {
	f.t.Helper()
	out, err := f.run(dir, "git", args...)
	if err != nil {
		f.t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(out)
}

// release runs `make release-alpha` with the given extra arguments and extra
// environment, returning its combined output and whether it succeeded.
func (f *fixture) release(extraEnv []string, args ...string) (string, bool) {
	f.t.Helper()
	saved := f.env
	f.env = append(append([]string{}, saved...), extraEnv...)
	defer func() { f.env = saved }()
	out, err := f.run(f.work, "make", append([]string{"-s", "release-alpha"}, args...)...)
	return out, err == nil
}

func (f *fixture) remoteTags() []string {
	f.t.Helper()
	out := f.git(f.work, "ls-remote", "--tags", "--refs", "origin")
	var tags []string
	for _, line := range strings.Split(out, "\n") {
		if i := strings.Index(line, "refs/tags/"); i >= 0 {
			tags = append(tags, line[i+len("refs/tags/"):])
		}
	}
	return tags
}

func (f *fixture) read(rel string) string {
	f.t.Helper()
	b, err := os.ReadFile(filepath.Join(f.work, rel)) // #nosec G304 -- path inside the fixture
	if err != nil {
		f.t.Fatalf("read %s: %v", rel, err)
	}
	return string(b)
}

func (f *fixture) head() string { return f.git(f.work, "rev-parse", "HEAD") }

// expectRefused asserts the recipe failed for the given reason and left
// nothing behind: no commit, no local or remote tag, a clean tree and an
// untouched remote main.
func (f *fixture) expectRefused(out string, ok bool, reason string) {
	f.t.Helper()
	if ok {
		f.t.Fatalf("release-alpha succeeded, want refusal (%q)\n%s", reason, out)
	}
	if !strings.Contains(out, reason) {
		f.t.Fatalf("release-alpha refused for the wrong reason: want %q in\n%s", reason, out)
	}
	if got := f.git(f.work, "status", "--porcelain"); got != "" {
		f.t.Fatalf("working tree modified by a refused release:\n%s", got)
	}
	if got := f.git(f.work, "tag", "--list"); got != "" {
		f.t.Fatalf("a refused release created local tags: %s", got)
	}
	if tags := f.remoteTags(); len(tags) != 0 {
		f.t.Fatalf("a refused release published tags: %v", tags)
	}
	if local, remote := f.head(), f.git(f.work, "rev-parse", "origin/main"); local != remote {
		f.t.Fatalf("a refused release committed: HEAD %s, origin/main %s", local, remote)
	}
}

// expectReleased asserts exactly testVersion was published, annotated, on the
// release commit, with remote main level with it and VERSION unprefixed.
func (f *fixture) expectReleased(out string, ok bool) {
	f.t.Helper()
	if !ok {
		f.t.Fatalf("release-alpha refused, want success:\n%s", out)
	}
	if tags := f.remoteTags(); len(tags) != 1 || tags[0] != testVersion {
		f.t.Fatalf("remote tags = %v, want only %s", tags, testVersion)
	}
	head := f.head()
	if got := strings.Fields(f.git(f.work, "ls-remote", "origin", "refs/heads/main")); len(got) == 0 || got[0] != head {
		f.t.Fatalf("remote main %v is not the release commit %s", got, head)
	}
	if got := f.git(f.work, "cat-file", "-t", testVersion); got != "tag" {
		f.t.Fatalf("tag %s is a %s, want an annotated tag", testVersion, got)
	}
	if got := f.git(f.work, "rev-parse", testVersion+"^{commit}"); got != head {
		f.t.Fatalf("tag %s points at %s, want HEAD %s", testVersion, got, head)
	}
	if got := f.read("VERSION"); got != "0.1.0-alpha\n" {
		f.t.Fatalf("VERSION = %q, want %q", got, "0.1.0-alpha\n")
	}
	if got := f.git(f.work, "status", "--porcelain"); got != "" {
		f.t.Fatalf("working tree dirty after release:\n%s", got)
	}
	log := f.read("../agent-git.log")
	for _, want := range []string{
		"tag -a " + testVersion + " -m " + testVersion,
		"push --atomic --no-follow-tags origin HEAD refs/tags/" + testVersion,
	} {
		if !strings.Contains(log, want) {
			f.t.Fatalf("agent-git.sh was not called with %q; calls:\n%s", want, log)
		}
	}
}

const releasedChangelog = "# Changelog\n\n" + unreleased + "\n\n" + testHeading + " - 2026-01-01\n\n### Fixed\n\n- a fix\n"

func TestReleaseAlphaRequiresExplicitVersion(t *testing.T) {
	f := newFixture(t, releasedChangelog, nil)
	out, ok := f.release(nil)
	f.expectRefused(out, ok, "state the version explicitly")
}

func TestReleaseAlphaRejectsMalformedVersion(t *testing.T) {
	for _, v := range []string{"0.1.0-alpha", "v0.1.0", "v0.01.0-alpha", "v0.1.0-beta", "v0.1-alpha"} {
		t.Run(v, func(t *testing.T) {
			f := newFixture(t, releasedChangelog, nil)
			out, ok := f.release(nil, "VERSION="+v)
			f.expectRefused(out, ok, "is not of the form vX.Y.Z-alpha")
		})
	}
}

// grep is line-oriented, so a VERSION of a valid first line plus a newline and
// more text would pass a grep-only grammar check. With entries pending, the
// changelog step would then commit a multiline heading and delete the
// fragments before any later guard refused. Both version guards must reject
// it before anything is written.
func TestReleaseAlphaRejectsMultilineVersion(t *testing.T) {
	const changelog = "# Changelog\n\n" + unreleased + "\n\n- pending unreleased entry\n"
	for name, v := range map[string]string{
		"pending":          testVersion + "\nextra",
		"trailing-newline": testVersion + "\n",
		"leading-line":     "junk\n" + testVersion,
	} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t, changelog, map[string]string{"1-x.md": "### Added\n\n- x\n"})
			head := f.head()
			out, ok := f.release(nil, "VERSION="+v)
			f.expectRefused(out, ok, "is not of the form vX.Y.Z-alpha")
			if got := f.head(); got != head {
				t.Fatalf("HEAD moved from %s to %s on a refused release", head, got)
			}
			if got := f.read("CHANGELOG.md"); got != changelog {
				t.Fatalf("CHANGELOG.md changed by a refused release:\n%s", got)
			}
			if got := f.read("changelog.d/1-x.md"); got != "### Added\n\n- x\n" {
				t.Fatalf("fragment changed by a refused release: %q", got)
			}
			if _, err := os.Stat(f.log); !errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("agent-git.sh was called during a refused release (stat err %v)", err)
			}
		})
	}
}

func TestReleaseAlphaRefusesOffMain(t *testing.T) {
	f := newFixture(t, releasedChangelog, map[string]string{"1-x.md": "### Added\n\n- x\n"})
	f.git(f.work, "checkout", "-q", "-b", "feature")
	f.git(f.work, "push", "-q", "origin", "feature")
	out, ok := f.release(nil, "VERSION="+testVersion)
	f.expectRefused(out, ok, "releases are cut from main")
}

func TestReleaseAlphaRefusesWithoutAgentGit(t *testing.T) {
	f := newFixture(t, releasedChangelog, map[string]string{"1-x.md": "### Added\n\n- x\n"})
	if err := os.Remove(filepath.Join(f.bin, "agent-git.sh")); err != nil {
		t.Fatal(err)
	}
	out, ok := f.release(nil, "VERSION="+testVersion)
	f.expectRefused(out, ok, "AGENT_GIT does not point at an executable agent-git.sh")
}

func TestReleaseAlphaRefusesExistingRemoteTag(t *testing.T) {
	f := newFixture(t, "# Changelog\n\n"+unreleased+"\n", map[string]string{"1-x.md": "### Added\n\n- x\n"})
	f.git(f.work, "tag", "-a", testVersion, "-m", "earlier")
	f.git(f.work, "push", "-q", "origin", "refs/tags/"+testVersion)
	f.git(f.work, "tag", "-d", testVersion)
	head := f.head()
	out, ok := f.release(nil, "VERSION="+testVersion)
	if ok || !strings.Contains(out, "already exists on origin") {
		t.Fatalf("want refusal on the remote duplicate tag, got ok=%v:\n%s", ok, out)
	}
	if f.head() != head || f.git(f.work, "status", "--porcelain") != "" {
		t.Fatalf("the changelog was assembled or committed before the duplicate-tag refusal")
	}
}

// A failed VERSION commit must stop the release before any tag exists — the v1
// recipe's `|| true` tagged anyway.
func TestReleaseAlphaFailedCommitPreventsTagging(t *testing.T) {
	f := newFixture(t, releasedChangelog, nil)
	out, ok := f.release([]string{"STUB_FAIL=commit-release"}, "VERSION="+testVersion)
	if ok {
		t.Fatalf("release-alpha succeeded although the release commit failed:\n%s", out)
	}
	if !strings.Contains(out, "refusing the release commit") {
		t.Fatalf("release failed for another reason:\n%s", out)
	}
	if got := f.git(f.work, "tag", "--list"); got != "" {
		t.Fatalf("a tag was created after the release commit failed: %s", got)
	}
	if tags := f.remoteTags(); len(tags) != 0 {
		t.Fatalf("tags reached the remote after the release commit failed: %v", tags)
	}
}

// Only the named tag is published — the v1 recipe pushed every local tag — even
// with push.followTags=true and an unpublished annotated tag reachable from HEAD.
func TestReleaseAlphaPublishesOnlyTheNamedTag(t *testing.T) {
	f := newFixture(t, releasedChangelog, nil)
	f.git(f.work, "tag", "-a", "v0.0.9-unrelated", "-m", "unrelated")
	f.git(f.work, "tag", "stray-lightweight")
	f.git(f.root, "config", "--global", "push.followTags", "true")
	out, ok := f.release(nil, "VERSION="+testVersion)
	f.expectReleased(out, ok)
	if strings.Contains(f.read("../agent-git.log"), "docs(changelog)") {
		t.Fatalf("an already-released changelog with nothing pending was re-assembled")
	}
}

func TestReleaseAlphaAssemblesChangelog(t *testing.T) {
	changelog := "# Changelog\n\nIntro.\n\n" + unreleased + "\n\n### Changed\n\n- older entry\n\n## [0.0.1-alpha] - 2026-01-01\n\n- first\n"
	f := newFixture(t, changelog, map[string]string{
		"20-b.md": "\n### Fixed\n\n- fragment b\n\n",
		"10-a.md": "### Added\n\n```markdown\n## [9.9.9] quoted, not a heading\n```\n\n- fragment a\n",
	})
	out, ok := f.release(nil, "VERSION="+testVersion)
	f.expectReleased(out, ok)

	date := regexp.MustCompile(`## \[0\.1\.0-alpha\] - (\d{4}-\d{2}-\d{2})`).FindStringSubmatch(f.read("CHANGELOG.md"))
	if date == nil {
		t.Fatalf("no dated %s heading:\n%s", testHeading, f.read("CHANGELOG.md"))
	}
	want := "# Changelog\n\nIntro.\n\n" + unreleased + "\n\n" + testHeading + " - " + date[1] + "\n\n" +
		"### Changed\n\n- older entry\n\n" +
		"### Added\n\n```markdown\n## [9.9.9] quoted, not a heading\n```\n\n- fragment a\n\n" +
		"### Fixed\n\n- fragment b\n\n" +
		"## [0.0.1-alpha] - 2026-01-01\n\n- first\n"
	if got := f.read("CHANGELOG.md"); got != want {
		t.Fatalf("assembled CHANGELOG.md:\n%s\nwant:\n%s", got, want)
	}
	if got := f.git(f.work, "ls-files", "changelog.d"); got != "changelog.d/.gitkeep" {
		t.Fatalf("changelog.d after release = %q, want only the .gitkeep", got)
	}
	subjects := f.git(f.work, "log", "--format=%s", "-2")
	if subjects != "chore(release): "+testVersion+"\ndocs(changelog): assemble "+testVersion {
		t.Fatalf("release commits = %q, want the assembly commit then the release commit", subjects)
	}
	if got := f.git(f.work, "rev-parse", "HEAD~1:CHANGELOG.md"); got != f.git(f.work, "rev-parse", testVersion+":CHANGELOG.md") {
		t.Fatalf("the tagged tree's CHANGELOG.md is not the assembled one")
	}
}

// A version that already has a section while entries are pending would ship
// without them; the release must refuse before committing anything.
func TestReleaseAlphaRefusesPendingEntriesForExistingSection(t *testing.T) {
	f := newFixture(t, releasedChangelog, map[string]string{"1-x.md": "### Added\n\n- x\n"})
	out, ok := f.release(nil, "VERSION="+testVersion)
	f.expectRefused(out, ok, "entries are still pending")
}

func TestReleaseAlphaRefusesWithNothingToRelease(t *testing.T) {
	f := newFixture(t, "# Changelog\n\n"+unreleased+"\n\n## [0.0.1-alpha] - 2026-01-01\n\n- first\n", nil)
	out, ok := f.release(nil, "VERSION="+testVersion)
	f.expectRefused(out, ok, "CHANGELOG.md has no")
}

func TestReleaseAlphaRefusesEmptyFragment(t *testing.T) {
	f := newFixture(t, "# Changelog\n\n"+unreleased+"\n", map[string]string{"1-x.md": "\n  \n"})
	out, ok := f.release(nil, "VERSION="+testVersion)
	f.expectRefused(out, ok, "is empty")
}

// Command-line values are data, never Make source: none of these may execute
// (template v2.3.3/v2.3.7/v2.3.8), including through the AGENT_GIT PATH lookup.
func TestReleaseAlphaCommandLineInjection(t *testing.T) {
	f := newFixture(t, releasedChangelog, nil)
	canary := filepath.Join(f.root, "canary")
	payload := "$(shell touch " + canary + ")"
	for _, arg := range []string{
		"VERSION=" + payload,
		"RELEASE_REMOTE=" + payload + "origin",
		"REPO_NAME=" + payload + "go-codec",
		"AGENT_GIT=" + payload,
		"REL_VERSION=" + payload,
		"REL_REMOTE=" + payload,
		"REL_REPO=" + payload,
		"REL_AGENT_GIT=" + payload,
	} {
		// Every case also names a remote that does not exist, so none can
		// publish; the assertion is that the payload never runs.
		args := []string{"VERSION=" + testVersion, "RELEASE_REMOTE=nowhere", arg}
		out, ok := f.release(nil, args...)
		if _, err := os.Stat(canary); err == nil {
			t.Fatalf("%s executed its payload:\n%s", arg, out)
		}
		if ok {
			t.Fatalf("%s: release succeeded against a nonexistent remote:\n%s", arg, out)
		}
		if got := f.git(f.work, "status", "--porcelain"); got != "" {
			t.Fatalf("%s: working tree modified:\n%s", arg, got)
		}
	}
	out, _ := f.release([]string{"PATH=" + f.bin + string(os.PathListSeparator) + payload + string(os.PathListSeparator) + os.Getenv("PATH")},
		"VERSION=v0.2.0-alpha", "RELEASE_REMOTE=nowhere")
	if _, err := os.Stat(canary); err == nil {
		t.Fatalf("a PATH entry executed its payload:\n%s", out)
	}
}
