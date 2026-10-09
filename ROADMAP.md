# go-codec roadmap

> **Status: REFERENCE** · Owner: Software Department · Last verified: 2026-10-08
> Changes are proposed by docs-only PR.

## Mission

Layer 0 library: multi-format serialization behind one common Codec interface.
Formats: JSON, YAML, TOML, MessagePack, BSON, CBOR and Protobuf.
Gives every consuming project consistent encoding with no in-house dependencies.

## Current milestone

[Release candidate (M2)](../../milestone/2), target kind RC.
Projected: 2026-10-11.

## Now

- [#26](../../issues/26) public release readiness landed via PR #29 — in review
- [#33](../../issues/33) docs carry index, status banners and stamps — open
- readiness: merge readiness PR #29 — in review
- readiness: CI green on main — done

## Next

Nothing queued. New work arrives as issues.

## Later

Nothing accepted. New work arrives as issues.

## Dependencies

- Up: none; no in-house modules are pinned in go.mod.
- Down: consumed across the portfolio for serialization, per [docs/02](../handbook/software/docs/02-dependency-graph.md).

## Non-goals

- Anything outside its Layer 0 serialization row in [docs/00](../handbook/software/docs/00-ownership-and-sdk-boundary.md).
