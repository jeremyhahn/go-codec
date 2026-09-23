# Style and Conventions

- Standard Go conventions (gofmt enforced)
- Go generics: `Codec[T]` interface pattern
- No codec build tags; the import is the opt-in (codec registers in `init()`)
- Each codec in its own sub-package under `pkg/`
- Factory pattern in `pkg/factory/` for runtime codec selection
- Race detection enabled in tests (`-race` flag)
- Coverage threshold: 95% enforced
- golangci-lint for advanced linting
- gosec + govulncheck for security scanning
