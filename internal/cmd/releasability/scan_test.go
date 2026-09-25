package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"
)

// Fixtures are assembled at run time: a literal lab address or owner path in
// this file would make the gate fail on its own test.
var (
	labIPv4    = fmt.Sprintf("%d.%d.%d.%d", 192, 168, 7, 42)
	tenIPv4    = fmt.Sprintf("%d.%d.%d.%d", 10, 20, 30, 40)
	rfc1918b   = fmt.Sprintf("%d.%d.%d.%d", 172, 20, 0, 1)
	cgnatIPv4  = fmt.Sprintf("%d.%d.%d.%d", 100, 100, 1, 1)
	ulaIPv6    = "fd" + "12:3456::1"
	ownerPath  = "/" + "home/someone/sources/x"
	macPath    = "/" + "Users/someone/src"
	internalDN = "agent@example" + ".internal"
	arpaDN     = "nas.home" + ".arpa"
)

// withHost registers a synthetic fleet hostname for the test's duration.
func withHost(t *testing.T, name string) {
	t.Helper()
	h := HostHash(name)
	if _, ok := fleetHosts[h]; ok {
		t.Fatalf("%s is already a fleet host", name)
	}
	fleetHosts[h] = struct{}{}
	t.Cleanup(func() { delete(fleetHosts, h) })
}

func TestScanTextFindings(t *testing.T) {
	withHost(t, "testbox")
	withHost(t, "labnode")
	tests := []struct {
		line  string
		kind  Kind
		match string
	}{
		{"url: http://" + labIPv4 + ":3001/x", KindLabAddress, labIPv4},
		{"addr " + tenIPv4 + ".", KindLabAddress, tenIPv4},
		{"(" + rfc1918b + ")", KindLabAddress, rfc1918b},
		{"vpn " + cgnatIPv4, KindLabAddress, cgnatIPv4},
		{"v6 [" + ulaIPv6 + "]:443", KindLabAddress, ulaIPv6},
		{"prefix " + "fd00" + "::/8", KindLabAddress, "fd00" + "::"},
		{"cd " + ownerPath, KindOwnerPath, "/" + "home/someone"},
		{"at " + macPath, KindOwnerPath, "/" + "Users/someone"},
		{"mail " + internalDN, KindInternalDomain, "example" + ".internal"},
		{"smb://" + arpaDN + "/share", KindInternalDomain, arpaDN},
		{"ssh TestBox", KindFleetHost, "testbox"},
		{"/mnt/labnode-07/data", KindFleetHost, "labnode-07"},
	}
	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := ScanText("f.txt", []byte("clean\n"+tt.line+"\n"))
			if len(got) != 1 {
				t.Fatalf("got %d findings %+v, want 1", len(got), got)
			}
			want := Finding{Path: "f.txt", Line: 2, Kind: tt.kind, Match: tt.match}
			if got[0] != want {
				t.Fatalf("got %+v, want %+v", got[0], want)
			}
		})
	}
}

func TestScanTextClean(t *testing.T) {
	withHost(t, "testbox")
	for _, line := range []string{
		"loopback 127.0.0.1 and docs 192.0.2.1 and public 8.8.8.8",
		"version 1.10.0.1.2 and v" + tenIPv4 + ".5",
		"not an address 999.168.1.1 or 010.0.0.1",
		"public v6 2001:db8::1, link fe80::1, fd::1",
		"placeholder /home/<user>/sources and /home/ alone",
		"module github.com/jeremyhahn/go-codec/internal/releasetest",
		"time.Local and x.internalize",
		"testboxes and testbox2 and mytestbox",
		"testbox-dev",
	} {
		if got := ScanText("f", []byte(line)); len(got) != 0 {
			t.Errorf("%q: unexpected findings %+v", line, got)
		}
	}
}

func TestScanSkipsBinaryAndMissing(t *testing.T) {
	fsys := fstest.MapFS{
		"bin":   {Data: []byte("\x00" + labIPv4)},
		"a.txt": {Data: []byte("x " + labIPv4)},
	}
	got, err := Scan(fsys, []string{"bin", "gone.txt", "a.txt"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Path != "a.txt" {
		t.Fatalf("got %+v, want exactly the a.txt finding", got)
	}
}

func TestScanReadError(t *testing.T) {
	fsys := fstest.MapFS{"d/f": {Data: []byte("x")}}
	_, err := Scan(fsys, []string{"d"})
	var re *ReadError
	if !errors.As(err, &re) || re.Path != "d" || errors.Unwrap(re) == nil {
		t.Fatalf("err = %v, want *ReadError for d", err)
	}
	if !strings.HasPrefix(re.Error(), "releasability: read d:") {
		t.Fatalf("message %q", re.Error())
	}
}

func TestFleetHostsAreDigests(t *testing.T) {
	hex := regexp.MustCompile(`^[0-9a-f]{64}$`)
	if len(fleetHosts) == 0 {
		t.Fatal("fleetHosts is empty")
	}
	for h := range fleetHosts {
		if !hex.MatchString(h) {
			t.Errorf("fleetHosts entry %q is not a SHA-256 hex digest", h)
		}
	}
}

// gitRepo creates a repository with the given tracked files and one untracked
// file carrying a lab address, which the gate must ignore.
func gitRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Fatalf("git is required: %v", err)
	}
	dir := t.TempDir()
	gitCmd := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...) // #nosec G204 -- fixed test arguments
		cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	gitCmd("init", "-q")
	for name, body := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	gitCmd("add", "-A")
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("lab "+labIPv4+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestRunPassesOnCleanTree(t *testing.T) {
	dir := gitRepo(t, map[string]string{"README.md": "# ok\n", "pkg/a.go": "package a\n"})
	var out, errOut bytes.Buffer
	if code := run([]string{"-C", dir}, &out, &errOut); code != exitClean {
		t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out.String(), errOut.String())
	}
	if !strings.Contains(out.String(), "2 tracked file(s) clean") {
		t.Fatalf("stdout %q", out.String())
	}
}

func TestRunFailsOnPlantedLabAddress(t *testing.T) {
	dir := gitRepo(t, map[string]string{
		"README.md":   "# ok\n",
		"docs/cfg.md": "server: http://" + labIPv4 + ":3001\n",
	})
	var out, errOut bytes.Buffer
	if code := run([]string{"-C", dir}, &out, &errOut); code != exitFindings {
		t.Fatalf("exit %d, want %d\nstdout:\n%s", code, exitFindings, out.String())
	}
	want := "docs/cfg.md:1: lab address: " + labIPv4 + "\n"
	if out.String() != want {
		t.Fatalf("stdout %q, want %q", out.String(), want)
	}
	if !strings.Contains(errOut.String(), "1 finding(s)") {
		t.Fatalf("stderr %q", errOut.String())
	}
}

func TestRunErrors(t *testing.T) {
	var out, errOut bytes.Buffer
	notRepo := t.TempDir()
	if code := run([]string{"-C", notRepo}, &out, &errOut); code != exitError || !strings.Contains(errOut.String(), "git ls-files") {
		t.Fatalf("not a repo: exit %d, stderr %q", code, errOut.String())
	}

	var ge *GitError
	if _, err := trackedFiles(notRepo); !errors.As(err, &ge) || errors.Unwrap(ge) == nil {
		t.Fatalf("err = %v, want *GitError", err)
	}

	empty := gitRepo(t, nil)
	if _, err := trackedFiles(empty); !errors.Is(err, errNoTrackedFiles) {
		t.Fatalf("err = %v, want errNoTrackedFiles", err)
	}

	errOut.Reset()
	if code := run([]string{"stray"}, &out, &errOut); code != exitError || !strings.Contains(errOut.String(), "unexpected arguments") {
		t.Fatalf("stray arg: exit %d, stderr %q", code, errOut.String())
	}
	errOut.Reset()
	if code := run([]string{"-nosuchflag"}, &out, &errOut); code != exitError {
		t.Fatalf("bad flag: exit %d", code)
	}
}

func TestRunHash(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"-hash", "TestBox"}, &out, &errOut); code != exitClean {
		t.Fatalf("exit %d", code)
	}
	if got := strings.TrimSpace(out.String()); got != HostHash("testbox") {
		t.Fatalf("hash %q is not case-folded HostHash", got)
	}
}
