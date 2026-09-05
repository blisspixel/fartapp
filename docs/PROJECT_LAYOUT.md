# Project layout

The root is an entry point and workspace boundary. Source belongs with its
package, tests beside that source, and shared fixtures under `testdata/`.
This follows the Go project's
[module layout guidance](https://go.dev/doc/modules/layout) and Rust's
[Cargo workspace convention](https://doc.rust-lang.org/book/ch14-03-cargo-workspaces.html).
Neither ecosystem prescribes a maximum number of root files.

| Location | Responsibility |
| --- | --- |
| `cmd/fartapp/` | Thin Go executable: process arguments, streams, and exit status |
| `internal/cli/` | Go command routing, options, presentation, and CLI tests |
| `internal/` | Private Go domains, numerical oracles, document adapters, evidence, and repository checks |
| `crates/fart-domain/` | Validated Rust quantities and inputs |
| `crates/fart-core/` | Pure Rust numerical calculations |
| `crates/fart-services/` | Bounded prediction and local play services, decoding, reports, and retained replay |
| `crates/fart-cli/` | Native Rust command presentation and process tests |
| `testdata/` | Shared authored cases, conformance corpora, and fixed expected outputs |
| `docs/` | Contracts, decisions, scientific references, walkthroughs, and release evidence |
| `brand/` and `docs/media/` | Reviewed assets and provenance manifests |
| `skills/` | Portable agent instructions and executable recipes |
| `tools/` | Thin development commands |
| `scripts/` | Optional PowerShell argument-forwarding wrappers |
| `.github/` | CI, dependency updates, and repository templates |
| `.cargo/` | Rust dependency review configuration |
| `artifacts/` | Ignored local build, coverage, and experimental outputs |
| `target/` and `node_modules/` | Ignored tool-managed outputs |

Keep files at the root when tools or contributors expect to discover them
there: language manifests and lockfiles, toolchain pins, Git and Markdown
settings, portable `plugin.json`, README, roadmap, license, security policy,
and contribution policies. Moving these into an arbitrary configuration folder
would add special invocation requirements.

## Dependency direction

Go commands call `internal/cli`. Adapters call application/report packages,
which call validated domain and numerical packages. Numerical packages do not
import CLI or filesystem presentation code.

Rust uses `fart-cli -> fart-services -> fart-core -> fart-domain`.
The reservoir predictor and bounded local reservoir session are implemented
subsets. General play contracts, capability composition, and a model registry
remain separate roadmap work. The assurance registry lives in
`internal/assurance/`; its generated reference belongs in `docs/INVARIANTS.md`.

Keep one coherent responsibility per file. Split parsing, calculation, and
presentation when they have different invariants or reasons to change. Avoid
empty scaffolding and one-function packages that add navigation without owning
a useful contract.

## Development commands

From the repository root:

```console
go run ./cmd/fartapp --help
go test ./internal/cli
go test ./...
cargo run --locked -p fart-cli -- --help
cargo test --locked --workspace
go run ./tools/repoquality repository
```

The earlier `go run .` and `go test .` development paths moved to
`./cmd/fartapp` and `./internal/cli` respectively. Installed Go command behavior
and permanent intensity outputs stay intact. Fixture arguments are still
relative to the repository root.

Do not commit binaries, raw coverage, staging files, credentials, local caches,
or machine-specific receipts. Keep compact release evidence under
`docs/releases/` and downloadable artifacts in GitHub Releases.
