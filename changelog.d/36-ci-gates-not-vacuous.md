### Fixed — local CI gates can no longer pass vacuously (#36)

- `make lint` now also runs `actionlint` over `.github/workflows/` and fails
  when `actionlint` is not installed, printing the `go install` hint.
- `make trivy` and `make trivy-image` exit non-zero with the install URL when
  `trivy` is missing, instead of printing "skipping" and exiting 0.
  `make trivy-image` also fails when `docker` is missing.
- `.github/workflows/ci.yml`: quoted the shell expansions that actionlint's
  shellcheck pass flagged (`basename "$pkg"`, `"$GITHUB_OUTPUT"`) and read the
  coverage threshold from the `COVERAGE_THRESHOLD` shell variable instead of
  inlining `${{ env.COVERAGE_THRESHOLD }}` into the `-lt` test (SC2170).
