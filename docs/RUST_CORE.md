# Rust reservoir foundation

The Rust workspace is an experimental, stateless reservoir foundation. It
implements the permanent five-level toy and one analytical ideal-mixture
reservoir endpoint operation. It is a concrete part of the v0.8 work, with its
own tests and implementation identity. It does not complete the v0.8 production
core milestone or provide `PlayService`, `CapabilityService`, sessions,
restriction-flow histories, walk witnesses, or archives. The broader Go CLI
remains available through [the project entry points](PROJECT_LAYOUT.md).

## Build and use

[rust-toolchain.toml](../rust-toolchain.toml) pins Rust 1.98.1 and the minimal
toolchain with Clippy, rustfmt, and LLVM tools. The workspace uses edition 2024.
Rust 1.98.1 was the current stable release when this foundation was introduced
in September 2026; it fixes the vtable-generation miscompilation in 1.98.0.
See the [official release announcement](https://blog.rust-lang.org/2026/09/03/Rust-1.98.1/)
and [edition guide](https://doc.rust-lang.org/edition-guide/rust-2024/index.html).

Run these commands from the repository root with Rustup and the platform's Rust
linker prerequisites installed:

```text
cargo build --locked --workspace
cargo run --locked -p fart-cli -- 3
cargo run --locked -p fart-cli -- --help
cargo run --locked -p fart-cli -- reservoir predict testdata/reservoir/synthetic-mixture-adiabatic.json --format json
cargo run --locked -p fart-cli -- reservoir predict testdata/reservoir/synthetic-mixture-isothermal.json
```

The native executable is `target/debug/fart`, or `target/debug/fart.exe` on
Windows, unless `CARGO_TARGET_DIR` selects another build directory. A release
build uses `cargo build --locked --release -p fart-cli` and the corresponding
`release` directory. The executable requires neither Node nor Python. Go is a
development dependency for comparison tests, not a Rust runtime dependency.

`reservoir predict` accepts a file path or `-` for standard input. Text is the
default; `--format json` emits one structured report followed by a newline.
Exit status 0 means the operation completed, and 1 means a usage, input, model,
or output failure. A refused JSON request still produces a refusal report.
Help reads no input. Toy intensities are canonical integers from 1 through 5;
their response strings retain the permanent legacy behavior.

## Four bounded layers

| Crate | Present responsibility | Dependencies |
| --- | --- | --- |
| [fart-domain](../crates/fart-domain/src/lib.rs) | Validated SI quantities, component IDs, immutable reservoir inputs, closures, withdrawal fractions, and toy intensity | Standard library |
| [fart-core](../crates/fart-core/src/lib.rs) | Pure endpoint equations, mixture summaries, component transfers, and arithmetic balances | Domain and standard library |
| [fart-services](../crates/fart-services/src/lib.rs) | Bounded request parsing, stateless prediction, structured refusals, and report rendering | Domain, core, Serde, and JSON |
| [fart-cli](../crates/fart-cli/src/lib.rs) | Arguments, bounded streams, files, output, and process-facing diagnostics | Services; JSON is also used by its tests |

All four crates inherit `publish = false` from the workspace manifest.

The accepted request schema is `fart.reservoir-prediction-request/v0alpha1`.
It requires the explicit model
`continuum.rigid-calorically-perfect-ideal-mixture@v0alpha1`, SI quantities,
component properties, a withdrawal fraction, and either `rigid-adiabatic` or
`rigid-isothermal` closure. The
[service parser](../crates/fart-services/src/parse.rs) limits encoded input to
65,536 bytes, JSON depth to 32, member names and component IDs to 128 bytes,
and the component collection to 64. Duplicate, unknown, case-alias, null, and
malformed Unicode members are refused. Component IDs determine normalized order.

The physical scope is a rigid, homogeneous, nonreacting, calorically perfect
ideal mixture with composition-preserving withdrawal. For retained mass
fraction `x`, the adiabatic temperature ratio is `x^(gamma-1)` and the pressure
ratio is `x^gamma`; isothermal temperature stays fixed and pressure scales with
`x`. The latter closure reports the heat that an ideal thermostat must supply.
It does not calculate a wall heat-transfer law, elapsed time, aperture flow,
momentum, a plume, physical sound, or a biological default. Arithmetic balance
claims do not establish empirical validity or Reference Pfft ratification.

## Comparison and independent evidence

The [native executable parity test](../crates/fart-cli/tests/parity.rs) compares
complete parsed JSON reports, including keys, types, array order, strings,
diagnostic paths and reasons, component identities, assumptions, and claim
metadata. Its one identity exception is `implementation_revision`: the native
report must declare `rust-reservoir/v0alpha1`, and the corrected Go reservoir
must declare `go-oracle.reservoir/v0alpha2`. Other revisions fail. The exception
does not remove the implementation field from either report.

For ordinary numeric fields, the permitted difference is
`64 * eps * max(abs(left), abs(right)) + tiny`, where `eps = 2^-52` and
`tiny = 2^-1074`, the smallest positive binary64 value. Claim residuals and the
corresponding balance fields must each satisfy their own report's tolerance;
their difference must also be within the sum of those two tolerances. Component
mass residuals use `64 * eps` times the largest initial, retained, or transferred
component mass, plus `tiny`, separately for each report. Comparator regression
tests reject discarded keys, changed diagnostics, array reordering, unreviewed
revisions, and erasing a representable `1e-100` quantity.

These are software comparison rules, not physical uncertainty estimates,
byte-for-byte JSON identity, or universal floating-point reproducibility. The
suite discovers the shared reservoir JSON fixtures and also exercises a
parameter grid, an exact underflow regression, hostile requests, and all five
toy intensities. A supplied `FARTAPP_GO_ORACLE` path selects the separately
built Go executable; an invalid explicit path fails the comparison. Without
that variable, ordinary Rust tests can run without Go, so that run alone does
not establish cross-language parity. CI supplies it explicitly.

Run the explicit comparison on a POSIX shell:

```sh
mkdir -p artifacts/bin
go build -o artifacts/bin/fartapp ./cmd/fartapp
export FARTAPP_GO_ORACLE="$PWD/artifacts/bin/fartapp"
cargo test --locked -p fart-cli --test parity -- --nocapture
```

Or in PowerShell:

```powershell
New-Item -ItemType Directory -Force artifacts/bin | Out-Null
go build -o artifacts/bin/fartapp.exe ./cmd/fartapp
$env:FARTAPP_GO_ORACLE = (Resolve-Path artifacts/bin/fartapp.exe).Path
cargo test --locked -p fart-cli --test parity -- --nocapture
```

[Independent core tests](../crates/fart-core/tests/reservoir.rs) anchor both
closures at `gamma = 1.5`, check zero withdrawal, sequential composition,
component ordering, and a parameter grid, and exercise extreme numerical
domains. The Go
[reservoir reference tests](../internal/idealmixturereservoir/reservoir_test.go)
provide a separately executable set of closed-form and invariant checks.

An exact powers-of-two regression exposed a defect shared by the original Go
and Rust arithmetic: a small intermediate product could underflow even though
the final mixture property or energy transfer was representable. Agreement
between those implementations would not have detected their shared mistake.
For `m = 2^-500`, `R = cv = 2^-560`, `V = 2^-60`, `T = 2^1000`, and isothermal
withdrawal `2^-20`, initial pressure is exactly 1 Pa, heat input is `2^-80` J,
and enthalpy out is `2^-79` J. Scaled arithmetic preserves those nonzero
transfers and the unchanged component properties. Evidence is retained in the
[Go scale regression](../internal/idealmixturereservoir/scale_test.go),
[Go product tests](../internal/floatmath/quotient_test.go), and
[Rust core regression](../crates/fart-core/tests/reservoir.rs). Truly
unrepresentable positive progress is refused rather than replaced by zero.

## Dependency and build-script scope

[Cargo.toml](../Cargo.toml) pins the two direct registry dependencies exactly,
disables their default features, and opts into `serde/std` and
`serde_json/std,float_roundtrip`. Serde's `derive` feature is not enabled;
the adapter uses its visitor interfaces directly. See the
[Serde feature contract](https://serde.rs/feature-flags.html) and
[the pinned package documentation](https://docs.rs/serde/1.0.229/serde/).

The active external graph contains these six packages, fixed in
[Cargo.lock](../Cargo.lock):

| Package | Version | Role |
| --- | --- | --- |
| `serde` | 1.0.229 | Serialization and deserialization interfaces |
| `serde_core` | 1.0.229 | Shared trait implementations, with `std` and `result` |
| `serde_json` | 1.0.151 | JSON parsing and rendering, with binary64 round-trip parsing |
| `itoa` | 1.0.18 | Integer formatting |
| `memchr` | 2.8.3 | Byte searching |
| `zmij` | 1.0.23 | Floating-point formatting |

The lockfile also contains `serde_derive`, `proc-macro2`, `quote`, `syn`, and
`unicode-ident`. Those five packages belong to inactive optional or
target-conditioned edges; their presence in a lockfile or an unfiltered Cargo
metadata graph does not mean they are compiled into this workspace. Inspect
the selected graph with `cargo tree --locked --edges normal,build` and inspect
features with `cargo tree --locked --edges features`.

All four workspace crates inherit `unsafe_code = "forbid"`. This policy covers
the project's Rust sources; it does not prohibit unsafe implementations inside
the standard library or external dependencies. Neither the dependency
allowlist nor the source lint proves whole-program memory safety.

There are no workspace build scripts. The four permitted dependency scripts
were reviewed at the locked versions:

| Package | Reviewed build-time behavior |
| --- | --- |
| `serde` | Writes a versioned private Rust module under `OUT_DIR`; probes `RUSTC --version` and selects compiler configuration |
| `serde_core` | Writes its private Rust module under `OUT_DIR`; probes compiler version and reads the Cargo target |
| `serde_json` | Selects arithmetic width from Cargo target architecture and pointer-width variables |
| `zmij` | Probes compiler version and reads Cargo optimization settings |

These script bodies contain no network client or native C/C++ build invocation.
Cargo and Rustup may separately download dependencies and toolchains. The
[Cargo build-script contract](https://doc.rust-lang.org/cargo/reference/build-scripts.html)
explains their execution before package compilation. Any package, feature,
version, or build-script change needs a corresponding policy review.

## Verification and promotion

[The dependency policy](../.cargo/deny.toml), enforced by `cargo-deny` 0.20.2,
allows only the six exact external package versions, reviewed feature sets,
four build scripts, and the crates.io registry. It rejects duplicate versions,
unreviewed registry or Git sources, and unapproved licenses; the accepted
license choices are Apache-2.0 and MIT. Advisory exceptions are empty. The
policy examines the workspace with all features enabled.

The implemented [Rust CI jobs](../.github/workflows/ci.yml) build and test on
Linux, macOS, and Windows. The quality job checks formatting, Clippy and rustdoc
with warnings denied, release builds, dependency policy, complete-report Go
parity, and LLVM source-line coverage. `cargo-llvm-cov` 0.9.0 generates coverage;
the repository's
[coverage checker](../internal/repoquality/rust_coverage.go) requires at least
90 percent aggregate and 80 percent in each of the four crates. It recomputes
counts from the actual `src` trees, rejects missing or duplicate evidence, and
does not count dependency or integration-test lines toward production coverage.

Promotion requires green checks on the reviewed commit and fresh green main CI
after merge. Local results do not substitute for that release gate. Each
release's evidence file identifies its exact source commit and successful CI.

Useful local checks are:

```text
cargo fmt --all --check
cargo clippy --locked --workspace --all-targets --all-features -- -D warnings
cargo test --locked --workspace --all-features
cargo deny --locked check
cargo llvm-cov --locked --workspace --all-features --json --summary-only --output-path artifacts/coverage/rust.json
go run ./tools/repoquality rust-coverage --profile artifacts/coverage/rust.json --aggregate 90 --package 80
```

Rustdoc also runs with `RUSTDOCFLAGS=-D warnings` in CI. Install the exact
quality-tool versions declared in that workflow before running the two Cargo
subcommands locally. Create `artifacts/coverage` before collecting coverage.
Set `FARTAPP_GO_ORACLE` to the built Go CLI when collecting
parity evidence. Broader service admission, cancellation, sessions, archives,
and protocol contracts remain separate [roadmap work](../ROADMAP.md).
