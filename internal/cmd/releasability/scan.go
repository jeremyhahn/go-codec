package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"net/netip"
	"regexp"
	"strings"
)

// Kind classifies a finding.
type Kind string

const (
	// KindLabAddress is a private-use IP address: RFC 1918 or RFC 6598 IPv4,
	// or an RFC 4193 unique-local IPv6 address.
	KindLabAddress Kind = "lab address"
	// KindFleetHost is the name of a machine in the company fleet.
	KindFleetHost Kind = "fleet hostname"
	// KindOwnerPath is an absolute path inside a named user's home directory.
	KindOwnerPath Kind = "owner path"
	// KindInternalDomain is a name under a private-use DNS suffix.
	KindInternalDomain Kind = "internal domain"
)

// Finding is one match in one tracked file.
type Finding struct {
	Path  string
	Line  int
	Kind  Kind
	Match string
}

// ReadError is a failure to read a tracked file.
type ReadError struct {
	Path string
	Err  error
}

func (e *ReadError) Error() string { return "releasability: read " + e.Path + ": " + e.Err.Error() }

func (e *ReadError) Unwrap() error { return e.Err }

// hashSalt prefixes every hostname before hashing. See fleetHosts.
const hashSalt = "go-codec/releasability:"

// HostHash returns the fleetHosts entry for a hostname or a hostname family
// stem. `go run ./internal/cmd/releasability -hash <name>` prints it.
func HostHash(name string) string {
	sum := sha256.Sum256([]byte(hashSalt + strings.ToLower(name)))
	return hex.EncodeToString(sum[:])
}

var (
	ipv4Pattern = regexp.MustCompile(`[0-9]{1,3}(?:\.[0-9]{1,3}){3}`)
	// Only the RFC 4193 unique-local range is private in IPv6; a candidate
	// must start with its fc or fd prefix before netip is asked to parse it.
	ipv6Pattern     = regexp.MustCompile(`(?i)\bf[cd][0-9a-f]{0,2}:[0-9a-f:.]*`)
	ownerPathRegexp = regexp.MustCompile(`/(?:home|Users)/[A-Za-z0-9_][A-Za-z0-9._-]*`)
	domainPattern   = regexp.MustCompile(`(?i)\b[a-z0-9-]+(?:\.[a-z0-9-]+)*\.(?:internal|home\.arpa)\b`)
	tokenPattern    = regexp.MustCompile(`[a-z0-9]+(?:-[a-z0-9]+)*`)
	familyPattern   = regexp.MustCompile(`^(.+)-[0-9]+$`)
	// cgnat is RFC 6598 shared address space, the range overlay VPNs hand to
	// lab hosts; netip's IsPrivate does not cover it.
	// It is built from octets so this file does not itself trip the gate.
	cgnat = netip.PrefixFrom(netip.AddrFrom4([4]byte{100, 64, 0, 0}), 10)
)

// ScanText returns every finding in one file's content.
func ScanText(path string, data []byte) []Finding {
	var out []Finding
	for i, line := range strings.Split(string(data), "\n") {
		for _, f := range scanLine(line) {
			f.Path, f.Line = path, i+1
			out = append(out, f)
		}
	}
	return out
}

func scanLine(line string) []Finding {
	var out []Finding
	add := func(k Kind, m string) { out = append(out, Finding{Kind: k, Match: m}) }
	for _, loc := range ipv4Pattern.FindAllStringIndex(line, -1) {
		if !ipv4Bounded(line, loc[0], loc[1]) {
			continue
		}
		m := line[loc[0]:loc[1]]
		if a, err := netip.ParseAddr(m); err == nil && (a.IsPrivate() || cgnat.Contains(a)) {
			add(KindLabAddress, m)
		}
	}
	for _, m := range ipv6Pattern.FindAllString(line, -1) {
		if a, ok := parseIPv6(m); ok && a.Is6() && a.IsPrivate() {
			add(KindLabAddress, m)
		}
	}
	for _, m := range ownerPathRegexp.FindAllString(line, -1) {
		add(KindOwnerPath, m)
	}
	for _, m := range domainPattern.FindAllString(line, -1) {
		add(KindInternalDomain, m)
	}
	lower := strings.ToLower(line)
	for _, tok := range tokenPattern.FindAllString(lower, -1) {
		if isFleetHost(tok) {
			add(KindFleetHost, tok)
		}
	}
	for _, h := range contextHostRefs(lower) {
		add(KindFleetHost, h)
	}
	return out
}

// hostLabel is one DNS label; every hostContextPatterns entry captures the
// candidate host as its first group.
const hostLabel = `([a-z0-9](?:[a-z0-9-]*[a-z0-9])?)`

// hostContextPatterns are the places where a lowercased line uses a word as a
// host name. contextHosts entries match only here.
var hostContextPatterns = []*regexp.Regexp{
	// URL authority: smb://name/share, https://user@name:3000.
	regexp.MustCompile(`[a-z][a-z0-9+.-]*://(?:[^\s/@]+@)?` + hostLabel),
	// UNC path: \\name\share.
	regexp.MustCompile(`\\\\` + hostLabel),
	// user@name, git@name:owner/repo.
	regexp.MustCompile(`[a-z0-9._%+-]+@` + hostLabel),
	// name:/path, name:~, name:port.
	regexp.MustCompile(`(?:^|[^a-z0-9_.@/\\-])` + hostLabel + `:(?:~|/[^/]|[0-9]{2,5}(?:[^0-9]|$))`),
	// ssh_config Host or HostName line.
	regexp.MustCompile(`^\s*host(?:name)?\s+` + hostLabel),
	// host: name, hostname=name, server: "name".
	regexp.MustCompile(`\b(?:host|hostname|server)\s*[:=]\s*["']?` + hostLabel),
	// name.lan, name.local, name.localdomain.
	regexp.MustCompile(`(?:^|[^a-z0-9_.@/\\-])` + hostLabel + `\.(?:lan|local|localdomain)(?:[^a-z0-9_.-]|$)`),
}

// hostCommand finds a remote-access or lookup command and captures its
// arguments up to the next shell separator; commandHostArgs checks the first
// maxCommandArgs of them.
var hostCommand = regexp.MustCompile(`(?:^|[^a-z0-9_-])(?:ssh|scp|sftp|rsync|ping6?|mosh|telnet|ssh-copy-id|traceroute|nslookup|dig)\s+([^;|&]*)`)

// hostArgument is a command argument naming a host: [user@]name followed by
// the end of the argument, :path, :port or a local suffix.
var hostArgument = regexp.MustCompile(`^(?:[^@]*@)?` + hostLabel)

const maxCommandArgs = 6

// localSuffixes are the private DNS suffixes a bare host name may carry and
// still be a fleet reference; name.com and the like are public names.
var localSuffixes = []string{"lan", "local", "localdomain", "internal", "home.arpa"}

// contextHostRefs returns every contextHosts name the lowercased line uses as
// a host, once per position.
func contextHostRefs(lower string) []string {
	var out []string
	seen := map[int]bool{}
	consider := func(start, end int) {
		if seen[start] || !hostEnds(lower, end) {
			return
		}
		name := lower[start:end]
		if _, ok := contextHosts[HostHash(name)]; ok {
			seen[start] = true
			out = append(out, name)
		}
	}
	for _, re := range hostContextPatterns {
		for _, loc := range re.FindAllStringSubmatchIndex(lower, -1) {
			consider(loc[2], loc[3])
		}
	}
	for _, loc := range hostCommand.FindAllStringSubmatchIndex(lower, -1) {
		for _, arg := range commandArgs(lower, loc[2], loc[3]) {
			if lower[arg[0]] == '-' {
				continue
			}
			if m := hostArgument.FindStringSubmatchIndex(lower[arg[0]:arg[1]]); m != nil {
				consider(arg[0]+m[2], arg[0]+m[3])
			}
		}
	}
	return out
}

// commandArgs splits lower[start:end] on whitespace and returns the bounds of
// at most maxCommandArgs arguments.
func commandArgs(lower string, start, end int) [][2]int {
	var args [][2]int
	for i := start; i < end && len(args) < maxCommandArgs; {
		for i < end && (lower[i] == ' ' || lower[i] == '\t') {
			i++
		}
		j := i
		for j < end && lower[j] != ' ' && lower[j] != '\t' {
			j++
		}
		if j > i {
			args = append(args, [2]int{i, j})
		}
		i = j
	}
	return args
}

// hostEnds reports whether the label ending at end is a whole host name: the
// label is not continued, and a dot after it ends the sentence or starts one of
// localSuffixes rather than a public domain or a Go selector such as name.FS.
func hostEnds(lower string, end int) bool {
	if end == len(lower) {
		return true
	}
	switch c := lower[end]; {
	case c == '-' || c == '_':
		return false
	case c != '.':
		return true
	}
	rest := lower[end+1:]
	if rest == "" || !isLabelByte(rest[0]) {
		return true
	}
	for _, suf := range localSuffixes {
		if strings.HasPrefix(rest, suf) && (len(rest) == len(suf) || !isLabelByte(rest[len(suf)])) {
			return true
		}
	}
	return false
}

func isLabelByte(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' || c == '_'
}

// ipv4Bounded rejects a dotted quad that is part of a longer dotted number,
// such as a five-part version string.
func ipv4Bounded(line string, start, end int) bool {
	digit := func(c byte) bool { return c >= '0' && c <= '9' }
	if start > 0 && (digit(line[start-1]) || line[start-1] == '.') {
		return false
	}
	if end < len(line) && digit(line[end]) {
		return false
	}
	return end+1 >= len(line) || line[end] != '.' || !digit(line[end+1])
}

// parseIPv6 parses a candidate, dropping trailing colons or dots that belong to
// the surrounding prose rather than the address.
func parseIPv6(m string) (netip.Addr, bool) {
	for m != "" {
		if a, err := netip.ParseAddr(m); err == nil {
			return a, true
		}
		if !strings.HasSuffix(m, ":") && !strings.HasSuffix(m, ".") {
			break
		}
		m = m[:len(m)-1]
	}
	return netip.Addr{}, false
}

func isFleetHost(tok string) bool {
	if _, ok := fleetHosts[HostHash(tok)]; ok {
		return true
	}
	if m := familyPattern.FindStringSubmatch(tok); m != nil {
		_, ok := fleetHosts[HostHash(m[1])]
		return ok
	}
	return false
}

// binary reports whether data looks like a binary file, the way git decides:
// a NUL byte in the first 8000 bytes.
func binary(data []byte) bool {
	head := data
	if len(head) > 8000 {
		head = head[:8000]
	}
	return bytes.IndexByte(head, 0) >= 0
}

// Scan reads every path from fsys and returns the findings. A path that no
// longer exists is skipped: it is a deletion the index has not recorded yet,
// and it cannot reach a published tree. Binary files are skipped.
//
// A symbolic link is never followed: git publishes a link as its target text,
// so that text is what is scanned, whether or not the target exists here.
// fsys must implement fs.ReadLinkFS (os.DirFS does) for links to be seen as
// links.
func Scan(fsys fs.FS, paths []string) ([]Finding, error) {
	var out []Finding
	for _, p := range paths {
		info, err := fs.Lstat(fsys, p)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, &ReadError{Path: p, Err: err}
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			target, err := fs.ReadLink(fsys, p)
			if err != nil {
				return nil, &ReadError{Path: p, Err: err}
			}
			out = append(out, ScanText(p, []byte(target))...)
			continue
		}
		data, err := fs.ReadFile(fsys, p)
		if err != nil {
			return nil, &ReadError{Path: p, Err: err}
		}
		if binary(data) {
			continue
		}
		out = append(out, ScanText(p, data)...)
	}
	return out, nil
}
