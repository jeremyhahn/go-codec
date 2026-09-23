#!/bin/sh
# assemble-changelog.sh — fold the pending changelog into a release section.
#
# Usage: assemble-changelog.sh <CHANGELOG.md> <X.Y.Z-alpha> <YYYY-MM-DD> <fragment-dir>
#
# release-versioning.md §5: a PR adds changelog.d/<issue>-<slug>.md instead of
# editing CHANGELOG.md, and the release step folds the fragments into the new
# version's section as a COMMITTED step before tagging. This script is the file
# transform only — it touches no git state. `make release-alpha` (the
# release-alpha-changelog prerequisite) runs it, removes the fragments and
# commits both through agent-git.sh.
#
# Pending entries are everything under "## [Unreleased]" (content written there
# before fragments were adopted releases normally) followed by every
# <fragment-dir>/*.md in name order. They become
#
#   ## [Unreleased]
#
#   ## [X.Y.Z-alpha] - YYYY-MM-DD
#
#   <Unreleased content>
#
#   <fragments>
#
# Headings are recognised only outside fenced code blocks, with the same
# CommonMark fence rules as the release-alpha CHANGELOG guard.
#
# Writes the assembled file to stdout and exits:
#   0  assembled
#   3  nothing to assemble: no pending entries (the version's heading may or
#      may not exist already; release-alpha's own guard decides that)
#   1  refused, with the reason on stderr: no "## [Unreleased]" heading, an
#      empty fragment, or the version already has a section while entries are
#      still pending (releasing then would silently leave them out)
set -eu

die() { printf 'assemble-changelog: %s\n' "$1" >&2; exit 1; }

[ "$#" -eq 4 ] || die 'usage: assemble-changelog.sh <CHANGELOG.md> <X.Y.Z-alpha> <YYYY-MM-DD> <fragment-dir>'
changelog=$1 version=$2 date=$3 dir=$4
[ -f "$changelog" ] || die "no changelog at $changelog"
[ -n "$version" ] || die 'empty version'
printf '%s' "$date" | grep -Eq '^[0-9]{4}-[0-9]{2}-[0-9]{2}$' || die "date is not YYYY-MM-DD: $date"

frags=$(mktemp)
trap 'rm -f "$frags"' EXIT
n=0
for f in "$dir"/*.md; do
	[ -e "$f" ] || continue
	[ -n "$(tr -d '[:space:]' <"$f")" ] || die "fragment $f is empty"
	[ "$n" -eq 0 ] || printf '\n' >>"$frags"
	# Strip leading/trailing blank lines so fragments are separated by exactly one.
	awk 'NF { for (; b > 0; b--) print ""; print; seen = 1; next } seen { b++ }' "$f" >>"$frags"
	n=$((n + 1))
done

rc=0
awk -v ver="## [$version]" -v date="$date" -v frags="$frags" '
function fence(s,   i, n) {
	i = 1; while (i <= 3 && substr(s, i, 1) == " ") i++
	fch = substr(s, i, 1); if (fch != "~" && fch != "`") return 0
	n = 0; while (substr(s, i + n, 1) == fch) n++
	info = substr(s, i + n); return n >= 3 ? n : 0
}
{
	line = $0; heading = 0
	n = fence(line)
	if (open) {
		if (n >= open && fch == opench && info ~ /^[ \t]*$/) open = 0
	} else if (n && !(fch == "`" && index(info, "`"))) {
		open = n; opench = fch
	} else {
		heading = (substr(line, 1, 3) == "## ")
		if (index(line, ver) == 1) hasver = 1
	}
	if (state == 0) {
		pre[++npre] = line
		if (heading && line ~ /^## \[Unreleased\][ \t]*$/) state = 1
	} else if (state == 1 && heading) {
		state = 2; post[++npost] = line
	} else if (state == 1) {
		body[++nbody] = line
	} else {
		post[++npost] = line
	}
}
END {
	if (state == 0) { print "assemble-changelog: no \"## [Unreleased]\" heading outside a code block" > "/dev/stderr"; exit 1 }
	first = 0; last = 0
	for (i = 1; i <= nbody; i++) if (body[i] ~ /[^ \t]/) { if (!first) first = i; last = i }
	nfrag = 0
	while ((getline l < frags) > 0) fr[++nfrag] = l
	pending = first || nfrag
	if (hasver && pending) { print "assemble-changelog: CHANGELOG.md already has a \"" ver "\" section, yet entries are still pending under [Unreleased] or in changelog.d; releasing would leave them out. Fold them in by hand or pick the next version" > "/dev/stderr"; exit 1 }
	if (!pending) exit 3
	for (i = 1; i <= npre; i++) print pre[i]
	print ""
	print ver " - " date
	print ""
	if (first) for (i = first; i <= last; i++) print body[i]
	if (first && nfrag) print ""
	for (i = 1; i <= nfrag; i++) print fr[i]
	if (npost) { print ""; for (i = 1; i <= npost; i++) print post[i] }
}' "$changelog" || rc=$?
exit "$rc"
