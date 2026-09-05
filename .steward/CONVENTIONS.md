# CONVENTIONS

## Current implementation

- Module: `github.com/blisspixel/fartapp` on the Go version declared in `go.mod`.
- `cmd/fartapp/main.go` adapts process arguments, streams, and exit status.
- `internal/cli/` owns Go command routing, presentation, and CLI tests.
- Private Go domains, numerical oracles, and evidence adapters live under
  `internal/`; shared authored fixtures live under `testdata/`.
- Four `crates/fart-*` workspace members implement the experimental Rust
  domain, core, stateless reservoir service, and native CLI.
- Generated local outputs belong under ignored `artifacts/`, `target/`, or
  the tool's configured cache. Conventional manifests remain at the root.
- Go and Rust changes pass the checks in [the quality contract](../docs/QUALITY.md):
  at least 90 percent aggregate coverage and 80 percent in each Go package and
  Rust crate, alongside formatting, analysis, and platform tests.

## Planned system boundaries

- The Go oracle remains small, auditable, and independent.
- Rust owns production domain types, solvers, ledgers, archives, services, CLI,
  terminal UI, audio, scores, radio, agent protocols, and native extension
  adapter.
- Godot owns native presentation and input, never authoritative physics or story
  facts.
- Stored contracts are versioned, hashed, unit-aware, and explicit about law and
  dimension.
- Random systems use named substreams and cannot depend on execution order.
- Every interface can export an archive reproducible by the CLI.
- Every interface lowers canonical intents through `PlayService`; adapters do
  not own gameplay state or import solver mutation APIs.
- No embedded browser, webview, local HTTP service, or network dependency is part
  of the desktop architecture.

## Definition of done

A change is implemented, documented, test-backed, accessible through the CLI,
reproducible from a clean checkout, and included in the appropriate proof path.
Formatting and lint are clean, tests and CI pass, the enforced coverage floors
hold, and new presentation or narrative cannot invent an independent event.

## Evidence

- [Roadmap](../ROADMAP.md)
- [Interface strategy](../docs/INTERFACES.md)
