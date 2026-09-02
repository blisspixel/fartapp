# Roadmap

F.A.R.T. Lab grows in three complete products:

1. **1.0 CLI Lab:** the full simulation, play, story, translation, and proof
   instrument works headlessly on Windows, macOS, and Linux.
2. **2.0 Terminal Lab:** an htop-style interface makes the same instrument live
   and legible without adding hidden capability.
3. **3.0 Native Lab:** a polished Godot application makes proven events spatial,
   audible, haptic, cinematic, and destructible without becoming a web wrapper.

Every physics, narrative, and gameplay feature ships CLI first, TUI second,
native third. A checkbox means merged, documented, and verified behavior.

## Completed seed

- [x] v0.1: Validate one integer intensity from 1 through 5.
- [x] v0.2: Map five intensities to deterministic emission strings.
- [x] v0.3: Add the rating layer.
- [x] v0.4: Cover levels and invalid input with tests.
- [x] v0.5: Ship the original end-to-end Go CLI.

## v0.6: Public and product foundation

- [x] Publish the reviewed repository with generated operational logs and their
  sensitive history removed.
- [x] Add cross-platform CI for formatting, vet, tests, and at least 80 percent
  Go statement coverage.
- [x] Add security, contribution, conduct, and release-readiness policies.
- [ ] Make a deliberate license decision. Public visibility alone does not grant
  reuse rights.
- [ ] Ratify the event ontology, law profile, modes, story-director rules,
  accessibility matrix, seed RFC, and archive threat model.
- [ ] Ratify event-identity, typed-provenance, law-capability, content-pack trust,
  performance-budget, and human-evaluation RFCs.
- [ ] Version the current toy behavior as permanent oracle fixtures.

Exit gate: a clean clone passes CI, the public history scan is clean, repository
visibility is confirmed public, and planned systems are clearly labeled as
planned.

## v0.7: Earth dry-flow Go oracle

- [ ] Add quantity, unit, law, emitter, interface, exterior, observer, numerical,
  scenario, history, and certificate schemas.
- [ ] Add `scenario validate`, `law inspect`, and dimension diagnostics.
- [ ] Implement ideal-mixture finite-reservoir mass and energy balance.
- [ ] Implement prescribed area, simple compliance, subsonic flow, and the
  analytical choking boundary with explicit assumptions.
- [ ] Export source mass, momentum, enthalpy, composition, pressure, and recoil
  histories.
- [ ] Compute the active Earth dry-flow signature and conservation ledgers.
- [ ] Ship a Go walking skeleton for one ordinary pfft: predict, simulate,
  inspect, explain, branch one counterfactual, certify, and replay.

Exit gate: zero-flow, adiabatic, isothermal, choking, positivity, conservation,
and trusted blowdown fixtures pass. The ordinary Earth-biological preset cannot
silently enter a laboratory pressure regime. The walking skeleton is already
fun and useful without Rust, a TUI, or a graphical application.

## v0.8: Rust production core and typed CLI

- [ ] Create `fart-domain`, `fart-core`, `fart-services`, and `fart-cli` crates.
- [ ] Match all versioned Go oracle fixtures within documented tolerances.
- [ ] Freeze the Go oracle's owned equations and fixture schema; require exact,
  manufactured, or independently derived references so two implementations do
  not merely repeat one derivation error.
- [ ] Add fixed-order stepping, deterministic reductions, counter-based random
  streams, cooperative cancellation, and separate physics and numerical hashes.
- [ ] Add typed `simulate`, `inspect`, and `verify` commands with human, JSON,
  JSONL, and CSV output contracts.
- [ ] Test Windows, macOS, and Linux with formatting, Clippy, docs, tests, and at
  least 80 percent coverage for core packages.

Exit gate: a headless Rust run agrees with the Go oracle, behaves correctly in a
TTY and a pipe, and reproduces under its declared determinism level.

## v0.9: Certified event archive

- [ ] Implement canonical scenario, manifest, history, and certificate records.
- [ ] Begin with the simplest bounded streamable history representation; add
  Arrow or another columnar layer only after named profiles miss measured
  budgets.
- [ ] Add `.fart` atomic write, inspect, replay, verify, migrate, and export.
- [ ] Add SHA-256 member hashes and canonical JSON metadata.
- [ ] Reject duplicate members, traversal, links, decompression bombs, oversized
  arrays, schema violations, nonfinite invalid JSON, and hash substitution.
- [ ] Separate replayable, internally consistent, code-verified,
  solution-verified, empirically validated, and fictional-law-consistent claims.
- [ ] Add fault injection for cancellation and interrupted writes.

Exit gate: archives round-trip across all supported systems, malicious fixtures
fail safely, and cancellation leaves either one valid archive or no archive.

## v0.10: One exceptional ordinary pfft

- [ ] Add a reduced starting-jet or puff model driven by source flux.
- [ ] Add recoil and observer sampling.
- [ ] Add deterministic procedural WAV from interface motion, compact sources,
  and a labeled stochastic turbulence closure.
- [ ] Implement no-exterior-sound vacuum behavior and visible-emission
  requirements.
- [ ] Promote the v0.7 walking skeleton into `fart quick` with one authored
  world, source, cultural stake, consequence, causal explanation, runnable
  counterfactual, certificate, replay, and inspect path.

Exit gate: one ordinary low-energy event is fun enough to replay, scientifically
inspectable, fully deterministic under its contract, and compelling without a
TUI or graphical application.

## v0.11: Seeded worlds and Broadcast

- [ ] Implement a 256-bit master seed and versioned named substreams.
- [ ] Add law, habitat, source, senses, culture, situation, and scenario grammar
  stages with bounded generation.
- [ ] Add a read-only event-fact API and authored storylets with fact provenance.
- [ ] Add `fart broadcast`, immutable event playback, transcript, and `.fartshow`
  episode archives.
- [ ] Maintain a curated seed museum and a regression corpus for dull, repeated,
  contradictory, unsafe, or scientifically unsupported episodes.
- [ ] Prove that narration, localization, terminal size, accessibility, camera,
  and presentation streams cannot change the physical-result identity.
- [ ] Run at least 10,000 deterministic seed cases with expressive-range reports.
- [ ] Run representative human reviews for humor, comprehension, pacing,
  grossness controls, science density, and ordinary-event replay value.

Exit gate: every seed terminates or fails explicitly within a fixed budget, no
valid event is rerolled for drama, and the archived episode replays its resolved
world, story, event, and certificate.

## v0.12: Freestyle, proof, and translation

- [ ] Add `freestyle`, sweeps, comparisons, branches, optimization, uncertainty,
  refinement, and custom challenge contracts.
- [ ] Define author-reviewed semantic Pi groups and machine-verify dimensions,
  dependencies, and possible omissions.
- [ ] Implement strict same-law, same-dimension similarity and translation.
- [ ] Add constrained approximate translation with explicit residuals,
  infeasibility, and Pareto choices.
- [ ] Add comic observer translation as a separately labeled presentation layer.
- [ ] Add ordinary Bathroom Science and Wind Tunnel challenge sets.

Exit gate: exact translations round-trip, incompatible targets say why they fail,
presentation mappings never claim physical equivalence, and challenge scoring
rewards precision and verification as well as magnitude.

## 1.0: Exceptional CLI Lab

- [ ] Complete Quick Play, Broadcast, Freestyle, Challenge, Campaign,
  simulation, translation, proof, replay, and export command families.
- [ ] Polish width-adaptive human output, errors, help, progress, plain mode,
  accessibility, hostile text, and pipe behavior.
- [ ] Publish p50 and p95 startup, first-output, analytical-run, cancellation,
  memory, and archive-size results on named Windows, macOS, and Linux systems.
- [ ] Generate shell completions, manual pages, schemas, examples, and command
  reference from the typed command model.
- [ ] Ship deterministic offline audio without requiring an audio device.
- [ ] Add signed or notarized release artifacts, checksums, SBOM, provenance,
  install and uninstall tests, and named platform baselines.
- [ ] Complete archive fuzzing, dependency review, security scan, content review,
  accessibility review, and cross-platform replay tests.

Exit gate: the CLI is a complete, exceptionally polished game and laboratory on
Windows, macOS, and Linux. No graphical interface is needed to experience the
product's identity.

## 2.0: Terminal Lab

- [ ] Add `fart-tui` with overview, emitter, interface, plume, payload,
  acoustics, proof, timeline, translation, and Broadcast views.
- [ ] Support wide, standard, compact, and minimum layouts.
- [ ] Support Unicode and ASCII, truecolor through reduced color, keyboard-only,
  remappable keys, optional mouse, and append-only screen-reader mode.
- [ ] Restore the terminal after success, error, panic, resize, suspend, and
  interrupts.
- [ ] Add cell-buffer snapshots plus PTY and ConPTY smoke tests.
- [ ] Prove that every TUI action has a CLI, scenario, or archive equivalent.

Exit gate: Terminal Lab has feature parity with CLI services, adds no solver or
story branches, and passes native terminal tests on all supported systems.

## 3.0: Native Tiled Chamber

- [ ] Create a native Godot app and thin Rust GDExtension over the existing core.
- [ ] Build one tiled room with source, instruments, responsive props, replay,
  spatial audio, particles, deposition, recoil, haptics, and consequences.
- [ ] Use one curated and inspectable discharge path from pfft through supported
  dry, wet, choked, and underexpanded laboratory cases.
- [ ] Show live values, active groups, regime causes, uncertainty, and proof.
- [ ] Add remappable keyboard, mouse, and controller input plus first-run
  accessibility setup.
- [ ] Package, sign, smoke test, install, run, export, and cleanly uninstall on
  Windows, macOS, and Linux.

Exit gate: the native app reproduces the certified Tiled Chamber scenario already
available in CLI and TUI, and every consumer cites the same event history.

## After the Tiled Chamber

### Native Broadcast and campaign

- [ ] Present Quick Play, Broadcast, Freestyle, and campaign in the native world.
- [ ] Add spatial staging, characters, translation cards, callbacks, and replay
  exhibits without native-only narrative truth.

### Interactive field physics

- [ ] Add a conservative compressible finite-volume solver with positivity,
  shock capturing, manufactured solutions, and separate grid and timestep proof.
- [ ] Add axisymmetric physics before making quantitative 3D plume claims.

### Multiphase and underwater packs

- [ ] Add parcels, evaporation, breakup, collision, deposition, bubble formation,
  Rayleigh-Plesset dynamics, and bubble acoustics with separate ledgers.

### Universal and extreme packs

- [ ] Add rarefied, orbital, planetary, reacting, plasma, MHD, stellar,
  relativistic, and gravitational profiles one verified model family at a time.
- [ ] Add alternate-dimensional and fictional-law profiles with explicit axioms,
  compatibility maps, and no false empirical-validation claims.

### Apocalypse Mode

- [ ] Add laboratory, planetary, and stellar source budgets before adding damage.
- [ ] Add target material, self-gravity, fracture, radiation, and relativistic
  consequence systems only after each has a defensible contract.

## Definition of done

- Behavior is implemented, documented, and reproducible from a clean checkout.
- Formatting and lint are clean, tests pass, and core statement coverage is at
  least 80 percent.
- Numerical work states equations, assumptions, units, dimensions, validity,
  tolerances, verification, and validation status.
- New presentation consumes the authoritative event graph.
- New narrative claims cite event facts and cannot modify simulation state.
- New interfaces preserve CLI parity.
- Every milestone delivers at least one player-visible improvement, one better
  scientific claim or boundary, and one stronger verification result.
- High-energy content changes the source model before it changes the scale.
- Public changes pass secret, privacy, dependency, license, safety, and
  accessibility review appropriate to their surface.
