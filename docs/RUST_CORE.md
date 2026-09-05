# Rust reservoir and local session foundation

The v0.8 alpha2 Rust increment supplies the permanent five-level toy, a
stateless analytical ideal-mixture reservoir predictor, and a bounded local
`PlayService` for reservoir experiments. Sessions retain an authored baseline,
costed decisions, model accounts, and an integrity-checked transcript. This is
a concrete part of the v0.8 work with its own tests and implementation identity.
General `PlayService` admission, `CapabilityService`, full Go command parity,
restriction-flow histories, walk witnesses, and certified archives remain open.
The broader Go CLI remains available through
[the project entry points](PROJECT_LAYOUT.md).

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
cargo run --locked -p fart-cli -- play run testdata/play/reservoir-session.jsonl
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
| [fart-services](../crates/fart-services/src/lib.rs) | Bounded parsing, stateless prediction, local experiment sessions, transcript integrity, refusals, and report views | Domain, core, Serde, JSON, canonicalization, and SHA-256 |
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

## Bounded local sessions

The [local service](../crates/fart-services/src/play/mod.rs) accepts explicit
`start`, `predict`, `observe`, `actions`, and `finish` commands under
`reservoir-experiment/v0alpha1`. It has one operator and one immutable authored
reservoir baseline. Every prediction starts from that baseline; service
revision and journal order do not describe elapsed physical time. The command,
baseline, transcript, and fingerprint profiles retain their own `v0alpha1`
revisions independently of the alpha2 release number.

Mutations require the declared actor, session reference, expected revision, and
idempotency key. An exact accepted retry returns its retained receipt without
charging another attempt, including after later progress or finish. A changed
request under the same key is refused. Well-formed model refusals consume one
attempt and preserve the latest successful account. Parsing and control
refusals do not consume an attempt. Read-only observation and action discovery
do not call the solver or extend the canonical journal.

The session admits at most 16 prediction attempts. Exhaustion remains visible
as truncation even when the operator subsequently finishes. End of input never
invents a finish. The CLI limits one command to 65,536 bytes, a command stream
to 1 MiB and 128 commands, and output to 16 MiB. A retained transcript is
limited to 8 MiB. The [session contract](PLAY_SESSION.md) specifies the complete
grammar, bounds, retry order, and completion semantics.

Retained fingerprints use `fart.play.rfc8785-sha256/v0alpha1`, which binds typed
finite JSON values through RFC 8785 canonicalization and SHA-256. Transcript
import verifies the schema, control chain, bindings, and digests. Replay
projects retained evidence without rerunning the numerical model and cannot
restore live writer authority. An unkeyed digest does not authenticate the
record or verify its physical claims. Scientific reconstruction and the
certified `.fart` archive remain separate work.

On Linux or macOS, retain and replay a complete fixture session with:

```sh
mkdir -p artifacts
cargo run --locked -p fart-cli -- play run testdata/play/reservoir-session.jsonl --format transcript > artifacts/session.json
cargo run --locked -p fart-cli -- play replay artifacts/session.json --format json
```

The shell owns this output file and can replace it or leave partial output.
This transport does not provide atomic archive publication. Text and JSONL
views are also available; human numeric display uses six significant digits
while the retained JSON values remain unchanged. The
[service tests](../crates/fart-services/tests/play.rs) exercise baseline reuse,
idempotency, stale and unauthorized requests, charged model refusals, budget
exhaustion, terminal states, bounded parsing, and transcript tampering.

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

In PowerShell, use the `.exe` suffix and set `$env:FARTAPP_GO_ORACLE` to the
executable's absolute path before running the same Cargo test command.

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

Small component transfers are calculated directly from mass and the declared
withdrawal fraction. Subtracting nearly equal endpoint masses loses relative
precision and can differ when a compiler contracts floating-point operations.
Independent small-transfer anchors retain the roundoff in the balance residual
without widening the full-report comparison tolerance. Subnormal reference
powers in Rust tests use exact bit patterns because integer-power evaluation
can itself underflow through an intermediate reciprocal.

## Dependency and build-script scope

[Cargo.toml](../Cargo.toml) pins four direct registry dependencies exactly and
disables their default features at the workspace declaration:

| Direct dependency | Version | Explicit workspace features |
| --- | --- | --- |
| `serde` | 1.0.229 | `std` |
| `serde_json` | 1.0.151 | `std`, `float_roundtrip` |
| `serde_json_canonicalizer` | 0.3.2 | None |
| `sha2` | 0.11.0 | None |

Canonicalization enables the transitive `default` feature of `serde` and
`serde_json`; each of those defaults enables only `std` at the pinned version.
The policy permits that resolved feature set explicitly. Serde's `derive`
feature is not enabled; the adapter uses its visitor interfaces directly.
SHA-256 has no enabled named features. See the
[Serde feature contract](https://serde.rs/feature-flags.html) and
[the pinned package documentation](https://docs.rs/serde/1.0.229/serde/).

The active external graph across supported targets contains these 17 packages,
fixed in [Cargo.lock](../Cargo.lock). Windows x64 selects 16; `libc` is selected
for the ARM64 Linux and Apple CPU-detection paths:

| Package | Version | Role |
| --- | --- | --- |
| `serde` | 1.0.229 | Serialization and deserialization interfaces |
| `serde_core` | 1.0.229 | Shared trait implementations, with `std` and `result` |
| `serde_json` | 1.0.151 | JSON parsing and rendering, with binary64 round-trip parsing |
| `itoa` | 1.0.18 | Integer formatting |
| `memchr` | 2.8.3 | Byte searching |
| `zmij` | 1.0.23 | Floating-point formatting |
| `serde_json_canonicalizer` | 0.3.2 | RFC 8785 serialization of strict, owned JSON values |
| `ryu-js` | 1.0.3 | ECMAScript-compatible number serialization for canonicalization |
| `sha2` | 0.11.0 | SHA-256 fingerprints |
| `cfg-if` | 1.0.4 | Compile-time platform selection |
| `cpufeatures` | 0.3.1 | Platform CPU-feature detection for SHA-256 |
| `digest` | 0.11.3 | Digest interfaces, with `default` and `block-api` |
| `block-buffer` | 0.12.1 | Digest block buffering |
| `crypto-common` | 0.2.2 | Shared cryptographic traits |
| `hybrid-array` | 0.4.14 | Fixed-size arrays for the digest implementation |
| `typenum` | 1.20.1 | Type-level integers, with `const-generics` |
| `libc` | 0.2.189 | Target-conditioned CPU-detection FFI |

The lockfile also contains `serde_derive`, `proc-macro2`, `quote`, `syn`, and
`unicode-ident`. Those five packages belong to inactive optional or
target-conditioned edges; their presence in the 22-package external lockfile
graph or an unfiltered metadata graph does not mean they are compiled into
this workspace. Inspect the host graph with
`cargo tree --locked --edges normal,build`, or select a supported target with
`--target aarch64-apple-darwin`. Inspect features with
`cargo tree --locked --edges features`. A blanket `--target all` also includes
inactive conditional edges and is not an active-build inventory.

The canonicalizer receives strict, owned, finite `serde_json::Value` data after
duplicate-member and shape validation. The service does not use upstream
streaming helpers that lose duplicate members or unchecked UTF-8 string
helpers. CPU detection chooses a SHA implementation; it does not supply
scientific model inputs.

All four workspace crates inherit `unsafe_code = "forbid"`. This policy covers
the project's Rust sources; it does not prohibit unsafe implementations inside
the standard library or external dependencies. Neither the dependency
allowlist nor the source lint proves whole-program memory safety.

There are no workspace build scripts. The five permitted dependency scripts
were reviewed at the locked versions:

| Package | Reviewed build-time behavior |
| --- | --- |
| `serde` | Writes a versioned private Rust module under `OUT_DIR`; probes `RUSTC --version` and selects compiler configuration |
| `serde_core` | Writes its private Rust module under `OUT_DIR`; probes compiler version and reads the Cargo target |
| `serde_json` | Selects arithmetic width from Cargo target architecture and pointer-width variables |
| `zmij` | Probes compiler version and reads Cargo optimization settings |
| `libc` | Reads compiler and target configuration, probes `rustc --version` through any configured wrapper, attempts `emcc -dumpversion`, and probes `freebsd-version` only when `LIBC_CI` is set |

These script bodies contain no network client or native C/C++ build invocation.
The `libc` script's `emcc` call is a version query, not a compilation step; the
script is active only where its target-conditioned package is selected.
Cargo and Rustup may separately download dependencies and toolchains. The
[Cargo build-script contract](https://doc.rust-lang.org/cargo/reference/build-scripts.html)
explains their execution before package compilation. Any package, feature,
version, or build-script change needs a corresponding policy review.

## Verification and promotion

[The dependency policy](../.cargo/deny.toml), enforced by `cargo-deny` 0.20.2,
allows only the 17 exact external package versions, reviewed feature sets,
five build scripts, and the crates.io registry. It rejects duplicate versions,
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
parity evidence. General `PlayService` admission, `CapabilityService`,
cancellation, multiple-operator sessions, full Go command parity, certified
archives, and external protocol contracts remain separate
[roadmap work](../ROADMAP.md).
