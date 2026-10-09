### Changed — docs index, status banners and doc corrections (#33)

- Added `docs/README.md`, an index of the seven per-format codec guides.
- Every `docs/*.md` now carries a REFERENCE status banner and a
  `Last verified:` stamp, checked against `pkg/`.
- `docs/protobuf.md`: the example now uses `protobuf.New[*User]()` and a
  non-nil destination message for `Unmarshal`. It documents that
  `factory.NewProtoBuf` returns `(codec.Codec[T], error)` with an always-nil
  error, and that importing `pkg/factory` links all seven codecs.
- `docs/msgpack.md`: removed the unmeasured "30-50% smaller" claim.
- Documentation only; no code or behavior change.
