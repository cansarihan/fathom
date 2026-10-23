# Security Policy

Fathom reads the output of `go test -fuzz` and decodes crash inputs; it never
executes a crash input itself. It does shell out to the Go toolchain, so run it
only against code you trust to compile.

## Reporting

Please report a suspected vulnerability privately through GitHub's security
advisory flow for this repository rather than opening a public issue. Include a
minimal reproduction and the affected version.
