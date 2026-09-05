# Engineering quality contract

Eighty percent test coverage is the floor beneath the floor. F.A.R.T. Lab aims
for software that remains trustworthy when the joke, codebase, numerical risk,
platform count, and protocol surface all become much larger.

Coverage is evidence that code executed. It is not evidence that assertions are
meaningful, mutants are killed, equations are correct, archives are safe, or a
model represents nature. Quality gates therefore combine example tests,
properties, fuzzing, mutation, differential comparison, static analysis,
performance budgets, security review, and scientific verification.

## Gates in force for the Go oracle

- Formatting, build, `go vet`, tests, and race tests on supported runners.
- Exact stdout, stderr, and exit-code fixtures for every valid level and invalid
  input class.
- Checked domain boundaries with no out-of-range panic path.
- Deliberate behavior for broken output streams.
- Fuzzing for arbitrary command input, determinism, bounded output, and stream
  separation.
- A subprocess test of the real executable.
- At least 90 percent aggregate statement coverage now.
- At least 80 percent statement coverage in every non-generated package as the
  repository gains packages.
- The dependency-free coverage-profile audit enforces both thresholds in CI and
  excludes only files carrying the standard Go generated-code marker.
- Dependency vulnerability and code scanning, public-checkpoint review of secret
  alerts, and a locked development-dependency source, script, integrity, and
  license allowlist. Production dependency manifests expand as languages arrive.

The current tooling gate pins `staticcheck`, `govulncheck`, workflow lint, and
Markdown tooling. The next gate adds mutation testing and changed-line coverage.
Core domain, solver, archive, replay, and protocol code targets at least 95
percent changed-line coverage and 90 percent mutation score. No critical mutant
may survive merely because an aggregate score passes.

## Current stable toolchain policy

Language and tool upgrades use published stable releases, explicit patch
versions, and the same numerical and cross-platform checks as a feature change.
Preview compilers, release candidates, and draft protocols do not become an
implicit production requirement. These pins were reviewed on 2026-09-04:

| Component | Reviewed version | Role and source |
| --- | --- | --- |
| Go | 1.27.1 | Runtime and checker compiler, selected by [go.mod](../go.mod); [official release history](https://go.dev/doc/devel/release) |
| Node.js | 26.8.1 | Current stable documentation toolchain, selected by [.node-version](../.node-version); [official release index](https://nodejs.org/dist/index.json) |
| npm | 12.0.2 | Documentation package manager, selected by [package.json](../package.json); [official release](https://github.com/npm/cli/releases/tag/v12.0.2) |
| Markdownlint CLI2 | 0.23.2 | Exact npm development dependency in the committed lockfile |
| Staticcheck | 2026.2.1 / module v0.8.1 | Pinned static analysis |
| Govulncheck | v1.7.0 | Pinned vulnerability analysis |
| Actionlint | v1.7.12 | Pinned workflow lint |
| Rust | 1.98.1, edition 2024 | Native workspace; [toolchain pin](../rust-toolchain.toml) and [official release](https://blog.rust-lang.org/2026/09/03/Rust-1.98.1/) |
| cargo-deny | 0.20.2 | Pinned Rust advisory, license, source, version, feature, and build-script policy |
| cargo-llvm-cov | 0.9.0 | Pinned source-coverage collection with the matching compiler's LLVM tools |

The application and repository checker retain zero external Go dependencies.
Node is used only for development documentation tooling. CI reads language and
package-manager pins from their source files. GitHub Actions retain immutable
commit pins.
The latest stable Node Current release is deliberate for this development-only
toolchain; its release status is distinct from LTS. Go 1.27 raises the macOS
floor to macOS 13, as recorded in its
[platform release notes](https://go.dev/doc/go1.27).

Go, npm, and Actions dependency updates are reviewed on the existing weekly
schedule. Language patch pins and standalone Go analysis-tool pins receive the
same explicit review, including their version support and full CI result.
Cargo dependency updates use that schedule too. Rust and its two quality tools
were reviewed on 2026-09-05. The active Rust dependency and build-script decisions
are recorded in [RUST_CORE.md](RUST_CORE.md).

## Cross-platform repository automation

Repository policy is enforced by a standard-library-only Go checker. Domain
logic lives in `internal/repoquality`. The thin command in `tools/repoquality`
only forwards arguments and exit status:

```text
go run ./tools/repoquality repository
go run ./tools/repoquality coverage --profile artifacts/coverage/go.out --aggregate 90 --package 80
go run ./tools/repoquality rust-coverage --profile artifacts/coverage/rust.json --aggregate 90 --package 80
go run ./tools/repoquality fuzz --time 5s
```

`repository` checks npm and Go dependency policy, local Markdown links,
media manifests, and the portable agent package. `coverage` enforces the
aggregate and per-package statement floors. `rust-coverage` recomputes line
coverage from bounded LLVM summary JSON, requires every Rust `src/*.rs` file
including nested modules, and enforces 90 percent aggregate and 80 percent in
each of the four crates. Tests and dependency sources do not increase those
totals. Missing, duplicate, escaping, and malformed source evidence fails.
`fuzz` runs the declared Go fuzz targets. CI runs the repository check
on Ubuntu, macOS, and Windows, and runs coverage and fuzz in the Linux quality
job. Optional PowerShell scripts under `scripts/` are wrappers only; they
contain no validation policy.

The checker bounds policy JSON, Markdown, covered Go source, and SVG inputs to
4 MiB; coverage profiles and media assets to 32 MiB. Policy JSON allows at most
64 levels and 4096-byte member names. It rejects malformed or ambiguous policy
documents, source traversal, symlink escapes, nonfinite coverage thresholds,
overflowing counters, empty evidence, and undeclared fuzz-target drift. Media
header validation is format-specific; WebP container validation is not a claim
to have decoded its image pixels.

New repository-policy logic belongs in this Go checker, not in another
shell-specific implementation. GitHub Actions YAML, locked npm installation,
`npm audit`, and Markdown lint remain direct tools. The project does not add
Make, Task, Just, or another runner merely to rename these commands.

## Structural and assisted-code controls

Code acceptance depends on evidence, not on whether a human or an automated
tool typed the first draft. Tool use does not lower review, security, license,
traceability, or scientific standards, and it does not require assistant names,
generator bylines, or attribution comments in source or product output.

Every meaningful package, file, type, and function owns a named responsibility.
Reviews reject mixed wire-parser-solver-report files, circular dependency
pressure, adapters leaking into the core, speculative extension points, unused
configuration, redundant wrappers, comments that merely restate code, and
generic abstractions whose inputs do not share one semantic contract. The
smallest coherent split is preferred over both a god file and fragmentation for
its own sake.

Validation and parsing are reused only when their contracts are identical.
Versioned wire formats require explicit tests for duplicate members, unknown
members, omissions, source positions, bounds, deterministic ordering, and exact
bytes where promised. Similar-looking formats with different ownership rules
remain separate rather than sharing a misleading parser.

Coverage-only tests are not evidence. A test must be capable of failing under a
meaningful defect and should use an independent oracle, analytical result,
metamorphic property, or externally fixed fixture where appropriate. Reviews
reject assertions that recompute an answer through the production path,
tautological residual checks, brittle source-string surgery, and negative cases
that never reach the behavior they claim to exercise.

This policy follows current evidence without turning one study into a universal
claim. [NIST SP 800-218A](https://doi.org/10.6028/NIST.SP.800-218A) supplies an
institutional comparison for producers of generative AI and dual-use foundation
models. It is not a study of code-assistant output. A 2025 randomized controlled
trial of experienced open-source developers found a 19 percent slowdown under
its particular mature-repository and early-2025 tool conditions, despite
participants expecting a speedup. That bounded result supports measuring
outcomes rather than assuming them. See
[Becker et al., arXiv:2507.09089](https://arxiv.org/abs/2507.09089).

## Contract lifecycle and evidence debt

Ratification is a compatibility decision, not praise, scientific validation, or
general maturity. Contract artifacts use these lifecycle states:

| State | Meaning |
| --- | --- |
| Idea | Discussion with no implementation or compatibility claim |
| Design candidate | Scoped prose with named owner, semantics, nonclaims, and open questions |
| Executable candidate | Internal implementation and positive and negative fixtures with no public wire promise |
| Review candidate | Normative semantics, limits, conformance corpus, versioning, migration, security, and owner review are complete enough for a ratification decision |
| Ratified internal | Approved for repository-internal dependency under its exact revision, but not advertised as a public compatibility contract |
| Public provisional | Versioned external surface that may still make documented breaking changes before stability |
| Stable public | Published compatibility, deprecation, migration, security-support, and conformance commitments are in force |
| Deprecated | Still supported for a declared transition policy but no longer preferred |
| Retired | No supported execution; historical readers and migration behavior follow the published policy |

A source comment or roadmap checkbox cannot promote a state. The evidence bundle
records the approving owner, exact artifact revision, fixtures, review decisions,
and remaining nonclaims. Scientific evidence, accessibility, localization,
performance, and security retain owner-specific statuses rather than being
inferred from this contract lifecycle.

Every experimental capability also carries explicit evidence debt. Debt is a
set of missing obligations, not a story-point total or one severity score. It
may include absent counterexamples, independent oracle, refinement, empirical
data, uncertainty, calibration, cross-platform parity, security review,
accessibility test, locale review, adapter parity, or recovery evidence. A
feature may remain available under an honest experimental label, but it cannot
cross a promotion gate whose required debt is open. Shipping another feature
does not pay unrelated evidence debt.

## Critical mutant policy

The following mutations are release blockers regardless of aggregate score:

- Input bounds or unit dimensions change.
- A conservation sign, transfer direction, or system boundary changes.
- Positivity, finite-value, or capability checks disappear.
- A hash field, seed namespace, random counter, or identity boundary changes.
- Sound propagates in vacuum without a declared medium.
- Archive path, size, member, or decompression defenses weaken.
- Role authorization, observation filtering, idempotency, or retry behavior
  weakens.
- A presentation stream changes authoritative physics or scoring.

## Invariant registry

The [canonical machine-readable registry](../internal/assurance/registry.json)
owns every original invariant ID, responsibility owner, applicability scope,
tolerance profile, candidate lifecycle, declared Go check, evidence reference,
counterexample, milestone and remaining direction. The
[complete reference](INVARIANTS.md) is generated from it; the Go CLI embeds the
same metadata for `fartapp assurance list` and `fartapp assurance inspect <ID>`.

Repository policy validates the bounded schema, exact Go test declarations,
source containment, separate verification-benchmark references and generated
reference drift. It rejects an executable candidate without checks or
counterexamples and refuses lifecycle promotions beyond this provisional
metadata schema. Planned rows remain planned. Inspection does not execute
checks, evaluate case applicability or report a passing invariant.

Read [assurance semantics and maintenance](ASSURANCE.md) for limits, lifecycle
rules, source validation and regeneration. These invariant IDs remain separate
from the broader benchmark namespace in [VERIFICATION.md](VERIFICATION.md);
matching text does not establish equivalent scope or promote a benchmark.

## Rust production gates

The v0.8 alpha2 subset includes the stateless reservoir predictor and a bounded
local `PlayService`. It enforces formatting, warnings-denied Clippy and rustdoc,
locked debug/release builds, tests on Linux/macOS/Windows, full-report Go
reservoir comparisons, and complete-source line coverage. Session evidence
includes immutable-baseline experiments, exact accepted retries, revision and
actor checks, charged model refusals, bounded journals, exhaustion followed by
finish, and transcript integrity without numerical reevaluation. Read the
[session contract](PLAY_SESSION.md) for its exact limits and nonclaims.

The dependency gate is `cargo deny --locked check` with the reviewed
[configuration](../.cargo/deny.toml). Four direct registry pins select 17 active
external packages across supported targets, including target-conditioned
ARM64 `libc`; five dependency build scripts are permitted. The exact versions,
resolved features, inactive lockfile nodes, and reviewed script behavior are
recorded in [RUST_CORE.md](RUST_CORE.md#dependency-and-build-script-scope).
Coverage is collected with:

```console
cargo llvm-cov --locked --workspace --all-features --json --summary-only --output-path artifacts/coverage/rust.json
go run ./tools/repoquality rust-coverage --profile artifacts/coverage/rust.json --aggregate 90 --package 80
```

CI sets `FARTAPP_GO_ORACLE` to a freshly built Go executable so the cross-language
test cannot silently skip. Optional local Rust-only tests may omit that path.
The comparison scope remains the reservoir endpoint and permanent toy output;
full Go command parity, general `PlayService` admission, and `CapabilityService`
remain open. The following broader production requirements apply as their
features arrive; mutation analysis, Miri, Loom, and full ontology comparison
are not claimed by this alpha release.

The crate graph enforces `domain <- core <- services <- adapters`. Pure domain
and core crates cannot depend on terminal, Godot, network, protocol, or ambient
filesystem packages. Pure crates forbid unsafe code. Necessary platform and
GDExtension boundaries are isolated, reviewed, and tested.

Per pull request:

- Rust formatting, Clippy with warnings denied, rustdoc warnings denied, all
  targets, all features, and minimal features.
- Go-oracle parity against frozen independent fixtures.
- Property tests for units, positivity, zero-flow limits, choking continuity,
  conservation, similarity, identity, and translation round trips.
- Dependency advisory, license, source, ban, and duplicate-version policy.
- Architecture dependency tests and generated-schema drift checks.

Scheduled and release gates add mutation testing, longer fuzzing, Miri, Loom,
sanitizers, deterministic repeats under varied thread counts, and numerical
benchmark comparisons.

## Native field and accelerator gates

The optional C++20 Kokkos field library is a separately owned high-risk
component. It has warnings-as-errors builds, formatting and static analysis,
address, undefined-behavior, thread, and device sanitizers where supported,
focused fuzz harnesses at the C ABI, and explicit exception containment. No C++
exception, STL object, allocator ownership, or implicit thread lifetime crosses
the ABI.

Each Serial, Threads or OpenMP, CUDA, HIP, and SYCL build passes the applicable
benchmark registry and cross-backend comparisons. The suite covers precision,
fast-math, FMA, denormal, optimization, reduction order, device loss,
out-of-memory, cancellation, checkpoint restart, domain decomposition, and
compiler variation. Apple Metal remains a non-certifying preview backend unless
the hardware and implementation pass the requested precision class.

Mojo is measured through the same suite as an isolated research backend. A new
language cannot graduate through microbenchmarks alone. Native Windows support,
stable packages, offline reproducible builds, redistribution, debugging,
profiling, sanitizers, FP64, multi-device behavior, and total maintenance cost
are release gates.

## Numerical verification and validation

Every solver feature states applicable axioms, rules or equations, closures,
dimensions and units where defined, assumptions, validity, discretization,
tolerances, reference cases, and validation status. The progression is:

1. Exact limits and independently derived analytical cases.
2. Manufactured solutions and observed order of accuracy.
3. Conservation, positivity, symmetry, and invariance properties.
4. Timestep, grid, parcel, and ensemble refinement where applicable.
5. Cross-implementation comparison without shared derivation errors.
6. Empirical validation against data whose population and uncertainty are
   actually relevant.

No single green badge combines code verification, solution verification, and
empirical validation.

## Surface-specific proof

- CLI examples execute in CI and machine output receives schema tests.
- TUI rendering is a pure cell buffer with wide, standard, compact, minimum,
  Unicode, ASCII, color, and screen-reader snapshots.
- PTY and ConPTY tests prove resize, interruption, panic, and terminal restoration.
- Native tests bind frame, audio, haptic, caption, accessibility, and action
  revisions to the same Lab account.
- MCP and A2A run official conformance suites plus retry, cancellation,
  idempotency, authorization, leak, backpressure, and resource-limit tests.
- Archives receive round-trip, migration, malformed-input, decompression,
  traversal, short-write, disk-full, cancellation, and atomicity tests.

## Performance and reproducibility

Budgets name hardware, operating system, build, dataset, warm-up, sample count,
and percentile. Releases report startup, first output, analytical runtime,
cancellation latency, TUI input and render latency, native frame pacing, audio
underruns, memory, and archive size.

Determinism claims use the levels in [SIMULATION.md](SIMULATION.md). Cross-platform
bit equality is never promised by aspiration. A comparison job must inspect
artifacts produced independently on Windows, macOS, and Linux.

## Documentation and media

Current gates:

- Internal Markdown links and the permanent toy CLI example are checked on
  every change.
- Screenshots identify shipped behavior or carry a visible `PLANNED CONCEPT`
  label.
- The visual manifest verifies file hashes, actual dimensions, byte budgets,
  alt text, references, source revision or generation provenance,
  post-processing, rights review, brand clearance state, and replacement
  history.
- No screenshot contains a username, host, secret, notification, private path,
  or prohibited attribution.

Planned gates as typed interfaces and scheduled automation exist:

- External research links are checked on a non-blocking schedule with archived
  review results rather than making ordinary builds depend on the network.
- Schemas, examples, help, completions, and manuals derive from typed models.
- Reproducible screenshot fixtures pin terminal size, theme, font, seed,
  platform, command, and source revision.

Quality requirements become stricter as risk grows. A new surface cannot lower
the confidence earned by the core beneath it.

## Institutional assurance target

The project is not NASA software and never implies NASA approval. It does use
public institutional guidance as a demanding comparison point. NASA-STD-8739.8B
defines a systematic lifecycle approach to software assurance, software safety,
and independent verification and validation. NIST SP 800-218 defines final SSDF
1.1 secure-development practices, while SSDF 1.2 remains a draft as of the
current research review. NIST SP 800-218A adds a final AI-specific secure
development community profile for generative AI and dual-use foundation models
without replacing the base SSDF.

The project should publish an assurance case that maps:

- Approved and traced requirements to architecture, hazards, code, tests, and
  release evidence.
- Safety-relevant states, transitions, commands, failures, and recovery to
  analysis and tests.
- Independent reference derivations and review to high-consequence numerical
  claims.
- Open-source and reused components to acquisition, license, vulnerability,
  maintenance, and replacement evidence.
- Threat models and SSDF practices to implemented controls and response plans.
- Released binaries to source, builder identity, locked dependencies, SBOM,
  checksums, signatures, attestations, and reproducibility comparisons.

Every release states which evidence exists and which institutional practices are
not claimed. A funny user interface never turns an educational simulator into a
medical, flight, pressure-vessel, or mission-safety tool.
