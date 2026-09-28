package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
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
	// KindUnreadable is a tracked file the gate cannot read as text, a binary
	// file or one over MaxFileSize, that ExemptFile does not list.
	KindUnreadable Kind = "unreadable file"
	// KindForbiddenFile is a file that never ships: a shell script or agent
	// context.
	KindForbiddenFile Kind = "forbidden file"
)

// Finding is one match in one tracked file. Line is 0 for a finding about
// the whole file: KindUnreadable and KindForbiddenFile.
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
	// host: name, hostname=name, server: "name", "hostname": "name".
	regexp.MustCompile(`\b(?:host|hostname|server)["']?\s*[:=]\s*["']?` + hostLabel),
	// name.lan, name.local, name.localdomain.
	regexp.MustCompile(`(?:^|[^a-z0-9_.@/\\-])` + hostLabel + `\.(?:lan|local|localdomain)(?:[^a-z0-9_.-]|$)`),
}

// hostCommand finds a remote-access or lookup command and captures its
// arguments up to the next shell separator; commandHostRefs parses them.
var hostCommand = regexp.MustCompile(`(?:^|[^a-z0-9_-])(ssh|scp|sftp|rsync|ping6?|mosh|telnet|ssh-copy-id|traceroute|nslookup|dig)\s+([^;|&]*)`)

// hostArgument is a command argument naming a host: [user@]name followed by
// the end of the argument, :path, :port or a local suffix.
var hostArgument = regexp.MustCompile(`^(?:[^@]*@)?` + hostLabel)

// commandSpec describes how a command takes its host.
type commandSpec struct {
	// valueFlags are the single-letter options that consume a value, either
	// attached (-p2222) or as the next argument (-p 2222).
	valueFlags string
	// copyOperands marks scp and rsync: every operand of the form
	// [user@]host:path names a host, and a bare operand is a local path.
	copyOperands bool
}

// hostCommands maps each command hostCommand recognizes to its option syntax,
// taken from the OpenSSH, rsync, iputils and bind manual pages. For the other
// commands the first operand is the host and later operands are a remote
// command or query arguments.
var hostCommands = map[string]commandSpec{
	"ssh":         {valueFlags: "BbcDEeFIiJLlmOopQRSWw"},
	"scp":         {valueFlags: "cDFiJlloPSX", copyOperands: true},
	"sftp":        {valueFlags: "BbcDFiJlloPRSs"},
	"rsync":       {valueFlags: "eBfMT", copyOperands: true},
	"ssh-copy-id": {valueFlags: "Fiop"},
	"mosh":        {valueFlags: "p"},
	"telnet":      {valueFlags: "blenSX"},
	"ping":        {valueFlags: "AcFIilMmNpQSstTWw"},
	"ping6":       {valueFlags: "AcFIilMmNpQSstTWw"},
	"traceroute":  {valueFlags: "AfgimNpstqwxz"},
	"nslookup":    {},
	"dig":         {valueFlags: "bcfkpqtxy"},
}

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
		spec := hostCommands[lower[loc[2]:loc[3]]]
		for _, arg := range commandHostArgs(spec, commandArgs(lower, loc[4], loc[5])) {
			if m := hostArgument.FindStringSubmatchIndex(lower[arg[0]:arg[1]]); m != nil {
				consider(arg[0]+m[2], arg[0]+m[3])
			}
		}
	}
	return out
}

// commandHostArgs walks a command's arguments past its options, however many
// there are, and returns the operands that name a host: the destination for
// ssh-style commands, every host:path operand for scp and rsync.
func commandHostArgs(spec commandSpec, args []commandArg) [][2]int {
	var hosts [][2]int
	options := true
	for i := 0; i < len(args); i++ {
		a := args[i]
		if options && a.text == "--" {
			options = false
			continue
		}
		if options && len(a.text) > 1 && a.text[0] == '-' {
			if a.text[1] != '-' && optionTakesNext(spec.valueFlags, a.text[1:]) {
				i++
			}
			continue
		}
		if !spec.copyOperands {
			return append(hosts, a.span)
		}
		if c := strings.IndexByte(a.text, ':'); c > 0 && !strings.Contains(a.text[:c], "/") {
			hosts = append(hosts, a.span)
		}
	}
	return hosts
}

// optionTakesNext reports whether a cluster of single-letter options, such as
// "vp" from -vp, ends in one that consumes the next argument as its value.
func optionTakesNext(valueFlags, cluster string) bool {
	for i := 0; i < len(cluster); i++ {
		if strings.IndexByte(valueFlags, cluster[i]) >= 0 {
			return i == len(cluster)-1
		}
	}
	return false
}

// commandArg is one whitespace-separated argument with surrounding quotes
// removed; span is its unquoted position in the line.
type commandArg struct {
	text string
	span [2]int
}

// commandArgs splits lower[start:end] on whitespace outside quotes, with no
// limit on the number of arguments.
func commandArgs(lower string, start, end int) []commandArg {
	var args []commandArg
	for i := start; i < end; {
		for i < end && (lower[i] == ' ' || lower[i] == '\t') {
			i++
		}
		j := i
		var quote byte
		for j < end && (quote != 0 || lower[j] != ' ' && lower[j] != '\t') {
			switch c := lower[j]; {
			case c == quote:
				quote = 0
			case quote == 0 && (c == '"' || c == '\''):
				quote = c
			}
			j++
		}
		s, e := i, j
		for e-s >= 2 && (lower[s] == '"' || lower[s] == '\'') && lower[e-1] == lower[s] {
			s, e = s+1, e-1
		}
		if e > s {
			args = append(args, commandArg{text: lower[s:e], span: [2]int{s, e}})
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

// ExemptFile is the committed list of tracked files the gate cannot read as
// text but that may still ship: binary files and files over MaxFileSize. Each
// entry is one line, the path followed by a reason; blank lines and lines
// starting with # are comments. The list is honored only when it is itself
// tracked, so a local edit cannot quietly widen it.
const ExemptFile = "releasability-exempt.txt"

// MaxFileSize is the size cap. A tracked file larger than this is a finding
// unless ExemptFile lists it.
const MaxFileSize = 1 << 20

// ExemptError is a malformed or stale ExemptFile entry.
type ExemptError struct {
	Line   int
	Reason string
}

func (e *ExemptError) Error() string {
	return fmt.Sprintf("releasability: %s:%d: %s", ExemptFile, e.Line, e.Reason)
}

// Scan reads every path from fsys and returns the findings. A path that no
// longer exists is skipped: it is a deletion the index has not recorded yet,
// and it cannot reach a published tree.
//
// The gate fails closed on what it cannot read as text: a binary file (see
// binary) or a file over MaxFileSize is a finding unless ExemptFile lists it.
// The text embedded in an exempt binary file is still scanned, see
// scanBinary, and an exempt oversize file is still scanned in full.
//
// A shell script, by extension or by shebang, and agent context (.claude/,
// .serena/, CLAUDE.md) are findings wherever they are tracked.
//
// A symbolic link is never followed: git publishes a link as its target text,
// so that text is what is scanned, whether or not the target exists here.
// fsys must implement fs.ReadLinkFS (os.DirFS does) for links to be seen as
// links.
func Scan(fsys fs.FS, paths []string) ([]Finding, error) {
	exempt, err := readExempt(fsys, paths)
	if err != nil {
		return nil, err
	}
	var out []Finding
	for _, p := range paths {
		if isAgentFile(p) {
			out = append(out, Finding{Path: p, Kind: KindForbiddenFile, Match: "agent context"})
		}
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
		if info.Size() > MaxFileSize && !exempt[p] {
			out = append(out, Finding{Path: p, Kind: KindUnreadable, Match: fmt.Sprintf("over %d bytes", MaxFileSize)})
			continue
		}
		data, err := fs.ReadFile(fsys, p)
		if err != nil {
			return nil, &ReadError{Path: p, Err: err}
		}
		if isShellScript(p, data) {
			out = append(out, Finding{Path: p, Kind: KindForbiddenFile, Match: "shell script"})
		}
		if binary(data) {
			if !exempt[p] {
				out = append(out, Finding{Path: p, Kind: KindUnreadable, Match: "binary file"})
				continue
			}
			out = append(out, scanBinary(p, data)...)
			continue
		}
		out = append(out, ScanText(p, data)...)
	}
	return out, nil
}

// readExempt parses ExemptFile when it is tracked. An entry without a reason,
// or naming a path that is not tracked, is an *ExemptError: a stale exemption
// would otherwise wait silently for a file of that name to reappear.
func readExempt(fsys fs.FS, paths []string) (map[string]bool, error) {
	tracked := make(map[string]bool, len(paths))
	for _, p := range paths {
		tracked[p] = true
	}
	exempt := map[string]bool{}
	if !tracked[ExemptFile] {
		return exempt, nil
	}
	data, err := fs.ReadFile(fsys, ExemptFile)
	if errors.Is(err, fs.ErrNotExist) {
		return exempt, nil
	}
	if err != nil {
		return nil, &ReadError{Path: ExemptFile, Err: err}
	}
	for i, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		path, reason, _ := strings.Cut(line, " ")
		if tab, r, ok := strings.Cut(path, "\t"); ok {
			path, reason = tab, r+" "+reason
		}
		if strings.TrimSpace(reason) == "" {
			return nil, &ExemptError{Line: i + 1, Reason: path + " has no reason"}
		}
		if !tracked[path] {
			return nil, &ExemptError{Line: i + 1, Reason: path + " is not a tracked file"}
		}
		exempt[path] = true
	}
	return exempt, nil
}

// isAgentFile reports whether a tracked path is agent context, which never
// ships: anything under a .claude or .serena directory, or a CLAUDE.md.
func isAgentFile(p string) bool {
	parts := strings.Split(p, "/")
	for _, d := range parts[:len(parts)-1] {
		if d == ".claude" || d == ".serena" {
			return true
		}
	}
	return parts[len(parts)-1] == "CLAUDE.md"
}

// shellExtensions are the file extensions of shell scripts.
var shellExtensions = []string{".sh", ".bash", ".zsh", ".ksh", ".csh", ".tcsh", ".fish"}

// shellShebang matches a shebang line that runs a shell, directly or through
// env: #!/bin/sh, #!/usr/bin/env bash, #!/usr/bin/env -S zsh -e.
var shellShebang = regexp.MustCompile(`^#!\s*\S*/(?:env\s+(?:-\S+\s+)*)?(?:sh|bash|dash|ash|zsh|ksh|mksh|csh|tcsh|fish|busybox)(?:\s|$)`)

// isShellScript reports whether a file is a shell script, by its extension or
// by the shebang on its first line.
func isShellScript(p string, data []byte) bool {
	name := strings.ToLower(p[strings.LastIndex(p, "/")+1:])
	for _, ext := range shellExtensions {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	first, _, _ := bytes.Cut(data, []byte("\n"))
	return shellShebang.Match(bytes.TrimRight(first, "\r"))
}

// scanBinary scans the text embedded in a binary file. Every byte outside
// printable ASCII, tab and newline becomes a separator, so each embedded
// string is scanned on its own and line numbers still count newline bytes.
// The data is scanned a second time with NUL bytes dropped, which turns
// UTF-16 text back into ASCII. A match found by both passes is reported once.
func scanBinary(path string, data []byte) []Finding {
	seen := map[Finding]bool{}
	var out []Finding
	for _, view := range [][]byte{printable(data), printable(bytes.ReplaceAll(data, []byte{0}, nil))} {
		for _, f := range ScanText(path, view) {
			if !seen[f] {
				seen[f] = true
				out = append(out, f)
			}
		}
	}
	return out
}

// binarySeparator replaces non-text bytes in a binary file; it is neither
// whitespace nor a host-name byte, so it ends a match without joining two
// embedded strings into one command line.
const binarySeparator = 0x1f

func printable(data []byte) []byte {
	out := make([]byte, len(data))
	for i, c := range data {
		if c == '\n' || c == '\t' || c >= 0x20 && c < 0x7f {
			out[i] = c
		} else {
			out[i] = binarySeparator
		}
	}
	return out
}
