# Contributing

Thanks for taking a look. Fathom is a small, single-binary Go project.

## Before a change

- `make lint` and `make test` must pass. CI runs both on every push.
- `gofmt` is enforced; run `make fmt`.
- New behaviour comes with a test. The parser in `internal/progress` and the
  crash grouping in `internal/crash` are pure and easy to test — add cases there
  rather than reaching for an integration test when you can.

## Running against real fuzzing

Tests that shell out to the Go toolchain are skipped under `go test -short`.
Run the full suite with `make test` before sending a change.

## Commits

Conventional commits (`feat:`, `fix:`, `test:`, `docs:`, `build:`, `ci:`),
one logical change each.
