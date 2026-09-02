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

Every named product invariant receives a stable identifier, owner, test mapping,
tolerance, and milestone. Planned invariants remain clearly marked until an
executable check exists.

| ID | Invariant | Current evidence | Release direction |
| --- | --- | --- | --- |
| `CLI-001` | Valid toy levels have exact permanent output | Golden table and executable test | Preserve as Go and Rust oracle fixtures |
| `CLI-002` | Arbitrary input cannot panic or produce unbounded output | Seeded Go fuzz target | Nightly retained-corpus fuzzing |
| `CLI-003` | Diagnostics never contaminate successful stdout | Exact fixtures and fuzz property | JSONL framing and broken-pipe conformance |
| `CLI-004` | Root, family, leaf, and topic help expose only implemented command paths without semantic work | Byte-identical route fixtures, tripwire input, executable test, and fuzz property | Generate CLI, manual, and completion help from one typed Rust command model |
| `ONT-001` | Catalog inspection requires no Earth discharge role | Executable atemporal relation-only Go catalog fixture | Extend the proof to scenario, archive, certificate, and every adapter with dedicated counterexamples |
| `CAP-001` | Capability reports keep eight decision axes independent | Exact localized-text and deterministic JSON Go fixtures | Cross-adapter schema and refusal conformance |
| `SCN-001` | Scenario-probe validity does not imply realization or admission | Exact atemporal and unavailable-Earth text and JSON fixtures | Ratified scenario admission and refusal conformance |
| `SCN-002` | Atemporal scenario probing consults no ambient or Earth default | Validator over document bytes and the compiled-in catalog, explicit empty ambient-input report, and counterexample fixture | Provider tripwires and broader orthogonal ontology suite |
| `OBS-001` | Observation capabilities and back-action remain explicit | Design contract | Passive, distributed, and coupled observer tests |
| `ACC-001` | Every enabled projection cites one authoritative Lab account without inventing causal structure | Design contract | Typed claim and transformation traversal tests |
| `PHY-002` | Applicable conserved transfers close for the declared boundary | Design contract | Exact and tolerance-based ledger properties |
| `PHY-003` | Earth-profile exterior sound is absent in vacuum | Design contract | Analytical and field-solver fixtures |
| `ID-001` | Replay presents a record and never claims recurrence | Design contract | Archive and UI language conformance |
| `ID-002` | Re-enactment receives a fresh record nonce without inventing a source-law identity change | Design contract | Deterministic identity property tests |
| `DET-001` | Presentation streams cannot change physical identity | Design contract | Same-trace differential tests |
| `SURF-001` | CLI, TUI, native, MCP, and A2A lower to canonical actions | Design contract | Adapter golden and differential fuzz tests |
| `ART-001` | Spatial artifacts declare projection, distortion, and loss | Design contract | Feature and artifact manifest verification |
| `LOC-001` | Localization cannot change case results, applicable identity sets, or score | Design contract | Cross-locale same-trace tests |
| `CMP-001` | Backend choice cannot silently change laws, rules, equations, or precision class | Design contract | Capability refusal and scenario-digest tests |
| `CMP-002` | Accelerated results preserve certified observables within tolerance | Design contract | CPU, Kokkos, CUDA, HIP, and SYCL differential suite |

## Rust production gates

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
current research review.

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
