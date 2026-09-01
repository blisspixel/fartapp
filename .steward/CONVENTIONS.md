# CONVENTIONS

## Current Go seed

- Module: `github.com/blisspixel/fartapp` on the Go version declared in `go.mod`.
- `main.go` owns CLI parsing and process behavior.
- `fart.go` owns current reference emission behavior.
- `fart_test.go` owns focused behavior and CLI tests.
- Go changes are formatted, pass vet and tests, and maintain at least 80 percent
  statement coverage.

## Planned system boundaries

- The Go oracle remains small, auditable, and independent.
- Rust owns production domain types, solvers, ledgers, archives, services, CLI,
  terminal UI, and native extension adapter.
- Godot owns native presentation and input, never authoritative physics or story
  facts.
- Stored contracts are versioned, hashed, unit-aware, and explicit about law and
  dimension.
- Random systems use named substreams and cannot depend on execution order.
- Every interface can export an archive reproducible by the CLI.
- No embedded browser, webview, local HTTP service, or network dependency is part
  of the desktop architecture.

## Definition of done

A change is implemented, documented, test-backed, accessible through the CLI,
reproducible from a clean checkout, and included in the appropriate proof path.
Formatting and lint are clean, tests and CI pass, core statement coverage is at
least 80 percent, and new presentation or narrative cannot invent an independent
event.

## Evidence

- [Roadmap](../ROADMAP.md)
- [Interface strategy](../docs/INTERFACES.md)
