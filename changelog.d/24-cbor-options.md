### Added — `cbor.NewWithOptions` for decode strictness (#24)

- `cbor.NewWithOptions[T](dec cbor.DecOptions, enc cbor.EncOptions)` builds
  the fxamacker/cbor decode and encode modes once and returns a codec that
  uses them for `Marshal`, `Unmarshal`, `Encode` and `Decode`. Callers
  decoding attacker-controlled CBOR can now refuse duplicate map keys,
  indefinite-length items, tags and invalid UTF-8, and set their own nesting
  bound.
- Invalid options fail at construction with a typed `*cbor.OptionsError`
  (`errors.Is(err, cbor.ErrInvalidOptions)`), never per call.
- `cbor.New[T]()` keeps fxamacker's default options, so existing behavior
  is unchanged.
