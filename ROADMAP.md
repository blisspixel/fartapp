# Roadmap

F.A.R.T. Lab grows in three complete products:

1. **1.0 CLI Lab:** the complete simulation, play, story, translation, and proof
   set declared by the ratified 1.0 capability and content manifest works
   headlessly on Windows, macOS, and Linux. It does not imply every later model.
2. **2.0 Terminal Lab:** an htop-style interface makes the same instrument live
   and legible without adding hidden capability.
3. **3.0 Native Lab:** a polished Godot application makes proven events spatial,
   audible, haptic, cinematic, and destructible without becoming a web wrapper.

Every physics, narrative, and gameplay feature ships CLI first, TUI second,
native third. A checkbox means merged, documented, and verified behavior.

Versions express logical dependency and earned capability, never elapsed time or
delivery estimates. The project is not racing to 1.0. A milestone advances only
when its exit evidence is complete, and scope moves later rather than weakening
a scientific, usability, accessibility, security, or quality gate.

## Completed seed

- [x] v0.1: Validate one integer intensity from 1 through 5.
- [x] v0.2: Map five intensities to deterministic emission strings.
- [x] v0.3: Add the rating layer.
- [x] v0.4: Cover levels and invalid input with tests.
- [x] v0.5: Ship the original end-to-end Go CLI.

## v0.6: Public and product foundation

- [x] Publish the reviewed repository with generated operational logs and their
  sensitive history removed.
- [x] Add cross-platform CI for formatting, build, vet, static analysis,
  vulnerability analysis, fuzzing, shuffled tests, race tests, and at least 90
  percent Go statement coverage.
- [x] Add security, contribution, conduct, and release-readiness policies.
- [x] License repository-owned source code, documentation, and approved project
  media under Apache License 2.0, with separate rights recorded for any
  third-party material.
- [ ] Ratify `Occurrence`, `EmissionAnalogue`, and Earth `DischargeEvent`
  boundaries, bounded representability, multi-law contexts and bridge rules,
  optional law structures, structured no-mapping results, modes, story-director rules,
  accessibility matrix, seed RFC, and archive threat model.
- [ ] Ratify event-identity, typed-provenance, law-capability, content-pack trust,
  performance-budget, and human-evaluation RFCs.
- [ ] Ratify measurement interactions separately from read-only views and
  presentation, plus canonical play actions, roles, knowledge policies,
  journals, revisions, idempotency, surface parity, and input timing.
- [ ] Ratify physical-audio, diagnostic-sonification, Symphony, radio-catalog,
  cultural-safeguard, and agent/spectator threat-model RFCs.
- [ ] Ratify Reference Pfft, occurrence, trace, reconstruction, re-enactment,
  record-nonce and `event_nonce` wire semantics, natural-language interpretation,
  Plumeprint, and Fartflake RFCs.
- [ ] Ratify locale-invariant semantic identifiers, language packs, typography,
  nonhuman observer communication, translation loss, and locale parity RFCs.
- [x] Version the current toy behavior as permanent exact-output fixtures, reject
  ambiguous and out-of-range input safely, test output failure, fuzz arguments,
  and smoke test the executable.
- [x] Lock Markdown tooling, pin CI actions to immutable revisions, and move
  dependency updates to a weekly schedule.
- [x] Protect the public default branch against force-push and deletion, require
  current CI and CodeQL checks for merges, and preserve a documented
  maintainer-only emergency bypass.

Exit gate: a clean clone passes CI, the public history scan is clean, repository
visibility is confirmed public, and planned systems are clearly labeled as
planned.

## v0.7: Law-context Go oracle and first Earth-continuum specialization

- [ ] Freeze biology-neutral `ref:rp1:v1`, its exact SI constant chain,
  conventional exterior, reference observer, comparison vector, uncertainty
  budget, traceability graph, and non-normative status.
- [ ] Define the exact `RP-1 definition event` separately from imperfect,
  one-time everyday realizations.
- [ ] Add record, optional context-occurrence identity claims, law-context, scope, provenance,
  measurement, view, comparison, numerical, scenario, history, and certificate
  schemas, with emitter,
  interface, exterior, and payload as Earth discharge extensions.
- [x] Add dependency-free `law list` and `law inspect` commands with versioned
  experimental localized-text and deterministic JSON fixtures. Keep law
  definition, implementation, closure, applicability, evidence, trust, backend
  feasibility, and resource feasibility separate without requiring Earth roles
  or localized prose.
- [ ] Ratify the capability-report wire schema before calling it canonical or
  freezing it for the Rust service and protocol adapters.
- [x] Add a dependency-free, read-only `scenario validate` probe for one exact
  law revision, one opaque scope, and explicit capability requests. Bound and
  strictly parse file or stdin JSON, reject duplicate and unknown members, keep
  document validity separate from unratified realization admission, and prove
  the atemporal fixture uses no ambient or Earth default.
- [x] Add dependency-free root and nested help for every and only implemented
  command path. Distinguish the permanent v0.6 oracle from experimental v0.7
  probes, label current English text as presentation, and keep JSON tokens
  locale-invariant without claiming shared language or meaning.
- [ ] Add profile-neutral `scenario validate` and scenario refusal inspection
  over the ratified full scenario report. Ship its minimal atemporal, observerless,
  nonconserving, recurrence-free, and no-implicit-bridge counterexamples in the
  same milestone.
- [ ] Add dimension diagnostics only where selected law contexts define
  dimensional quantities.
- [ ] Implement ideal-mixture finite-reservoir mass and energy balance.
- [ ] Implement prescribed area, simple compliance, subsonic flow, and the
  analytical choking boundary with explicit assumptions.
- [ ] Export source mass, momentum, enthalpy, composition, pressure, and recoil
  histories.
- [ ] Compute the active Earth dry-flow signature and conservation ledgers.
- [ ] Ship a Go walking skeleton for one ordinary pfft: predict, simulate,
  inspect, explain, branch one counterfactual, certify, witness, and reconstruct.
- [ ] Run compensated disabled-player review of the CLI walking skeleton and
  document remediation and retesting before freezing its interaction contract.

Exit gate: zero-flow, adiabatic, isothermal, choking, positivity, conservation,
and trusted blowdown fixtures pass. The biology-neutral reference fixture cannot
silently enter a laboratory pressure regime. The walking skeleton is already
fun and useful without Rust, a TUI, or a graphical application.
A first-time operator can discover every implemented command from root help
without receiving a planned command or an implicit Earth, body, observer,
scientific-language, or realization default. Current English remains explicitly
labeled as presentation.

## v0.8: Rust production core and typed CLI

- [ ] Create `fart-domain`, `fart-core`, `fart-services`, and `fart-cli` crates.
- [ ] Freeze `fart-compute`, precision, determinism, device-provenance,
  checkpoint, and backend-conformance contracts before writing GPU kernels.
- [ ] Enforce the language and dependency policy: safe Rust for pure core,
  narrow reviewed FFI only for a measured specialist need, committed lockfiles,
  least feature sets, dependency decision records, and no Python, JavaScript,
  JVM, browser, or dynamic-plugin runtime requirement.
- [ ] Implement the canonical in-process `PlayService` reducer, read-only view
  projector, measurement-interaction boundary, knowledge policy, and typed
  action model inside `fart-services`.
- [ ] Implement one `CapabilityService` and typed `CapabilityReport` separating
  law concepts, implementation, closure, applicability, evidence grade, trust,
  backend feasibility, and resource refusal for every adapter.
- [ ] Route the walking skeleton through `PlayService`; adapters cannot import
  solver mutation APIs directly.
- [ ] Match all versioned Go oracle fixtures within documented tolerances.
- [ ] Freeze the Go oracle's owned equations and fixture schema; require exact,
  manufactured, or independently derived references so two implementations do
  not merely repeat one derivation error.
- [ ] Add fixed-order stepping, deterministic reductions, counter-based random
  streams, cooperative cancellation, and separate physics and numerical hashes.
- [ ] Add typed `simulate`, `inspect`, and `verify` commands with human, JSON,
  JSONL, and CSV output contracts.
- [ ] Add the offline natural-language request grammar and `fart ask --dry-run`,
  producing typed proposals, assumptions, ambiguities, budgets, source-span
  mappings, and stable interpretation receipts before acceptance.
- [ ] Separate semantic keys from display strings, use locale-aware quantities,
  and test pseudo-locales, grapheme width, bidirectional text, and hostile
  translated input without changing event identity.
- [ ] Implement and canonical-JSON-round-trip tiny ontology fixtures for a
  non-gaseous transfer, a discrete graph, an atemporal or partially ordered
  law, a distributed nonlinguistic observer, an occurrence with no emitter or
  exterior, a state-altering measurement distinct from a passive view, a
  coupled multi-law occurrence with an explicit bridge, and inapplicable Earth
  quantities.
- [ ] Make CLI inspection capability-driven so an unsupported or inapplicable
  concept never appears as a fake zero, empty Earth panel, or sentinel value.
- [ ] Test Windows, macOS, and Linux with formatting, Clippy, docs, tests,
  properties, fuzzing, mutation analysis, at least 80 percent per-package
  coverage, and stricter core and changed-line gates.

Exit gate: a headless Rust run agrees with the Go oracle, behaves correctly in a
TTY and a pipe, reproduces under its declared determinism level, and passes all
ontology fixtures. Measurement interaction changes occurrence-result identity;
view and presentation changes do not.

## v0.9: Certified event archive

- [ ] Implement canonical scenario, manifest, history, and certificate records.
- [ ] Add play-session identity, ordered action journals, actor roles,
  checkpoints, branch lineage, transition receipts, and privacy-safe saves.
- [ ] Store occurrence, trace, reconstruction, and re-enactment lineage
  separately. Replay presents retained evidence and never claims recurrence.
- [ ] Support explicit ephemeral, witness-only, and full-reconstruction recording
  policies selected before an event.
- [ ] Begin with the simplest bounded streamable history representation; add
  Arrow or another columnar layer only after named profiles miss measured
  budgets.
- [ ] Add `.fart` atomic write, inspect, replay, verify, migrate, and export.
- [ ] Round-trip every v0.8 ontology fixture through `.fart`, proving that
  serialization and journal sequence do not become source-law temporal order.
- [ ] Add SHA-256 member hashes and canonical JSON metadata.
- [ ] Reject duplicate members, traversal, links, decompression bombs, oversized
  arrays, schema violations, nonfinite invalid JSON, and hash substitution.
- [ ] Separate replayable, internally consistent, code-verified,
  solution-verified, empirically validated, and fictional-law-consistent claims.
- [ ] Freeze archive envelopes only after optional law structures,
  state-altering measurement, read-only contextual views, and structured
  no-mapping results round-trip.
- [ ] Add fault injection for cancellation and interrupted writes.

Exit gate: archives round-trip across all supported systems, malicious fixtures
fail safely, and cancellation leaves either one valid archive or no archive.

## v0.10: One exceptional ordinary pfft

- [ ] Add a reduced starting-jet or puff model driven by source flux.
- [ ] Add recoil and observer sampling.
- [ ] Add deterministic procedural WAV from interface motion, compact sources,
  and a labeled stochastic turbulence closure.
- [ ] Keep physical acoustics and diagnostic sonification separate, with
  declared calibration, confidence, safety limiting, and synchronized text.
- [ ] Implement no-exterior-sound vacuum behavior and visible-emission
  requirements.
- [ ] Promote the v0.7 walking skeleton into `fart quick` with one authored
  world, source, cultural stake, consequence, causal explanation, runnable
  counterfactual, certificate, witness, reconstruction, and inspect path.
- [ ] Make bare `fart` always start with the banal Reference Pfft. Pressure
  vessels, extreme sources, graphic payloads, and catastrophe stay explicit.
- [ ] Pilot the ordinary-event comedy, comprehension, grossness-control, and
  accessibility evaluation without showing participants the project pitch.

Exit gate: one ordinary low-energy event is fun enough to replay, scientifically
inspectable, fully deterministic under its contract, and compelling without a
TUI or graphical application.

## v0.11: Seeded worlds and Broadcast

- [ ] Implement a 256-bit scenario seed, separate 256-bit record nonce,
  commitment policy, and versioned named substreams.
- [ ] Add a scope-first grammar for law contexts, realization, and provenance,
  with capability-selected habitat, source, measurement, view, agents, senses,
  situated context, institutions, positions, and social stakes. No optional
  layer is invented merely to complete a pipeline.
- [ ] Add a read-only context-scoped occurrence-claim API and authored storylets
  with claim provenance.
- [ ] Add `fart broadcast`, immutable retained-occurrence playback, transcript, and `.fartshow`
  episode archives.
- [ ] Add `fart chill` as a scoreless, sparse, offline ambient stream using the
  same generated occurrences, record nonces, radio, and recording policy.
- [ ] Maintain a curated seed museum and a regression corpus for dull, repeated,
  contradictory, unsafe, or scientifically unsupported episodes.
- [ ] Prove that narration, localization, terminal size, accessibility, camera,
  and presentation streams cannot change the occurrence-result identity.
- [ ] Add versioned communication profiles with explicit channel capabilities,
  ambiguity, loss, and refusal to map. Human language is one optional profile,
  not the reference class.
- [ ] Test distributed, continuous, non-symbolic, non-compositional, stateful,
  and observer-coupled communication without assuming a universal meaning graph.
- [ ] Prove a supplied scenario seed reconstructs the setting while a default
  re-enactment receives a fresh record nonce and record identity. Its
  context-occurrence identity claims follow only the selected law contexts.
- [ ] Run at least 10,000 deterministic seed cases with expressive-range reports.
- [ ] Run representative human reviews for humor, comprehension, pacing,
  grossness controls, science density, and ordinary-event replay value.
- [ ] Add regression gates for monoculture, colonial gaze, pseudo-language,
  sacred-practice targeting, sanitation stigma, coercion, and humiliation.

Exit gate: every seed terminates or fails explicitly within a fixed budget, no
valid event is rerolled for drama, and a recorded episode replays its resolved
evidence without claiming the occurrence happened again.

## v0.12: Freestyle, proof, and translation

- [ ] Add `freestyle`, sweeps, comparisons, branches, optimization, uncertainty,
  refinement, and custom challenge contracts.
- [ ] Define author-reviewed semantic Pi groups and machine-verify dimensions,
  dependencies, and possible omissions.
- [ ] Implement strict same-law, same-dimension similarity where dimensions
  apply.
- [ ] Add constrained approximate translation with explicit residuals,
  infeasibility, and Pareto choices.
- [ ] Bind translation witnesses to source and target `LawContextSet` hashes,
  scope assignments, inter-law bridge hashes, and mapped context-occurrence
  identities. A coupled occurrence has no global source identity unless a bridge
  defines one.
- [ ] Separate semantic translation, structural mapping, signal transcoding,
  and experience analogy, with typed reasons for incompatibility, absence,
  privacy, refusal, prohibition, undecidability, and unknown results.
- [ ] Add ordinary Everyday Phenomena and Wind Tunnel challenge sets.

Exit gate: exact translations round-trip, incompatible targets say why they fail,
presentation mappings never claim physical equivalence, and challenge scoring
rewards precision and verification as well as magnitude.

## v0.13: Sound, Symphony, and player soundtrack

- [ ] Complete qualified name, design-mark, and sound-mark clearance before
  locking names or mastering branded audio and station assets.
- [ ] Create `fart-audio`, `fart-score`, and `fart-radio` consumers over the
  authoritative occurrence-account and presentation services.
- [ ] Create `pressure-standard.v1`, an original six-second certified audio
  ident built from the fixed `RP-1 definition event`, score mapping, spatial render,
  haptic envelope, captions, safety report, and provenance.
- [ ] Add blind sound-mark similarity review and prohibit briefs, prompts, or
  reference audio that ask for imitation of an existing audio logo.
- [ ] Ship evidence, concert, and split Symphony modes with versioned mappings,
  fixed calibration, arbitrary tuning, pitch confidence, and explicit loss.
- [ ] Add Snowflake Etude, Comparative Canon, Similarity Fugue, and
  Conservation Concerto through CLI commands and semantic score export.
- [ ] Add a deterministic offline station catalog and scheduler, independent
  controls, synchronized lyrics, captions, and presentation-only identity.
- [ ] Author the first three station packs: Drift 93.7, Night Side 106.1, and
  The Local Medium 88.4, with music-first editorial and cultural review.
- [ ] Add a manual, spend-bounded Eleven Music v2 development pipeline that
  reads `ELEVENLABS_API_KEY` only from the environment, never runs in ordinary
  CI, and freezes approved bytes and hashes.
- [ ] Require a compact manifest, source and similarity review, and explicit
  rights approval before any music asset enters the public repository or a
  release package.
- [ ] Verify every ident, stem, representative mix, mode transition, and long
  Chill session against ITU-T H.872 or a documented stricter safe-listening
  profile, including loudness, true peak, transition jumps, bass, and limiting.
- [ ] Prove that radio, Symphony, mixing, captions, lyrics, and station controls
  cannot change physics, narrative canon, challenge scores, or replay identity.

Exit gate: one snowflake can be heard as physical evidence and as an inspectable
score, The Pressure Standard is recognizable and fully traceable without
imitating another mark, one Broadcast remains pleasant for a long listening
session, every lane works offline, and removing the radio catalog leaves the
complete game intact.

## v0.14: First-class agent and spectator play

- [ ] Freeze versioned observation, action, receipt, play-handle, role, budget,
  challenge-grade, match, and roster schemas.
- [ ] Ship exact human and machine CLI paths before protocol adapters, including
  strict JSONL, legal-random, greedy, notebook-planning, and scripted baselines.
- [ ] After JSONL freezes, add optional research-only Gymnasium single-agent and
  PettingZoo multi-agent bindings outside the shipped runtime. Test reset, step,
  spaces, seeding, termination versus truncation, and sequential or parallel
  semantics against `PlayService`.
- [ ] Add `scenario_draft`, `scenario_validate`, `scenario_accept`, and
  `scenario_explain_interpretation` to CLI and MCP over the same natural-language
  compiler and typed proposal schema.
- [ ] Test exact typed and natural-language tracks separately. No prose request
  can access hidden verifier state, bypass canonical action budgets, or mutate a
  session before proposal acceptance.
- [ ] Add G0 through G2 challenge generators, held-out variants, seed
  commitments, hash-chained journals, and reproducible benchmark reports.
- [ ] Define authoritative-evaluation signing: algorithm, evaluator identity,
  public-key distribution, rotation, revocation, clock policy, replay verification,
  and the separation between reproducible local and trusted official results.
- [ ] Add an MCP `2026-07-28` stdio adapter with explicit play handles,
  structured observations, bounded resources, Tasks, cancellation, and no
  hidden transport session.
- [ ] Add an A2A 1.0 Agent Card, role-bound tasks, streaming, artifacts,
  cancellation, inbound seats, outbound rosters, and read-only spectators.
- [ ] Add capability budgets and separate Researcher, Operator, Accessible
  Operator, Omnimodal, Consortium, and matched human tracks.
- [ ] Prove transition parity, information parity, idempotency, retry safety,
  role isolation, observation nonleakage, and exact CLI replay across adapters.
- [ ] Pass official MCP conformance and Inspector checks plus the A2A TCK for
  every advertised binding and cross-SDK ITK tests.
- [ ] Complete protocol-input, authorization, terminal-injection, archive,
  resource-exhaustion, SSRF, and remote-hosting security reviews.

Exit gate: a human through CLI, a semantic MCP agent, and an A2A participant can
complete the same Quick Play and representative challenge under declared
capability budgets. Canonical journals and results agree, unauthorized requests
fail safely, and human CLI play remains offline and requires no network.

## v0.15: Snowflake artifacts

- [ ] Freeze `ArtifactProjectionProfile`, target-medium, preserved-feature,
  distortion, accessibility, evidence-status, and loss metadata before 2D or 3D
  artifact formats.
- [ ] Implement versioned evidence and artifact views for Plumeprints.
- [ ] Map normalized source histories, interface modes, vortex and impulse
  skeletons, species, payload, deposition, observer response, active groups,
  uncertainty, and ledger state through typed provenance.
- [ ] Implement the independent Fartflake generator with deterministic topology,
  LOD, mesh, thumbnail, and terminal outputs.
- [ ] Encode only a safe content-addressed event reference in one declared
  orthographic verification view, with a conventional flat code fallback.
- [ ] Add `artifact plumeprint`, `artifact grow`, `artifact verify`, `artifact
  compare`, and GLB, STL, PNG, and SVG export through the CLI.
- [ ] Test manifold promises, degeneracy, normals, poly and byte budgets,
  scannability across scale, blur, lighting, and rasterization, alt text,
  privacy, hostile payloads, and evidence-versus-art labels.
- [ ] Prove artifact generation cannot change the occurrence account,
  certificates, scores, or identity sets.

Exit gate: one ordinary event produces a scientifically inspectable Plumeprint,
a distinct 3D Fartflake, a safe scannable reference, and a terminal rendering,
all from one provenance graph and without copying QR-Bloom code or assets. A
nonspatial conformance event either produces a declared lossy projection or a
precise unsupported result.

## v0.16: Brand, community, and public culture foundation

- [ ] Confirm, document, and monitor the professional word, common-law,
  app-store, domain, international, design-mark, and sound-mark clearance
  completed before asset lock.
- [ ] Optically draw and test the Open Isobar candidate at icon, terminal,
  print, embroidery, vinyl, engraving, enamel, light, dark, one-bit, and
  high-contrast sizes.
- [ ] Publish reviewed brand, trademark, community-kit, event-artifact,
  merchandise, and asset-rights policies with a machine-readable manifest.
- [ ] Keep Core Marks distinct from a CC0 community kit and a visibly different
  `UNOFFICIAL FIELD OBSERVATION` badge.
- [ ] Ship local Witness, Ledger, Boundary, and Postcard export cards with alt
  text, privacy lint, schema version, and `SIMULATED EVENT` status.
- [ ] Pilot The Standard Pfft, Grid Refinement Friday, and Same Pi, Different Sky
  without engagement traps or a proprietary social feed.
- [ ] Open official community spaces only after private reporting, human
  moderation ownership, enforcement, appeals, rate limits, and agent conduct
  controls exist.
- [ ] Sample official adult shirts, patches, stickers, notebooks, and observation
  prints only after legal, supplier, labeling, durability, privacy, accessibility,
  refund, and financial-transparency gates pass.

Exit gate: the identity is distinctive and reproducible, fans can remix an open
kit without implying endorsement, official artifacts remain trustworthy, and no
store or community service becomes a dependency of the application.

## v0.17: Verified field and accelerator laboratory

- [ ] Implement small scalar Rust references and a conservative C++20 Kokkos
  field solver behind `fart-compute`, starting with one-dimensional references
  and axisymmetric transient flow before quantitative three-dimensional claims.
- [ ] Add finite-volume reconstruction, stable Riemann fluxes, positivity
  protection, shock capturing, explicit boundary ownership, and separate time,
  grid, iterative, roundoff, and model-error accounts.
- [ ] Pass manufactured solutions, uniform advection, acoustic wave, Sod shock
  tube, shock-vortex, isentropic nozzle, finite blowdown, starting-vortex, and
  stopping-vortex benchmark gates from `docs/VERIFICATION.md`.
- [ ] Implement Kokkos Serial, Threads or OpenMP, CUDA, HIP, and SYCL as separate
  first-party builds behind a narrow versioned C ABI, with fixed-reduction and
  performance-reduction modes.
- [ ] Graduate the first optional NVIDIA CUDA build with CPU differential,
  sanitizer, device-loss, out-of-memory, compiler, precision, conservation,
  positivity, and performance evidence.
- [ ] Add checkpoint and restart, bounded memory planning, device inspection,
  backend selection, cancellation, and a safe CPU fallback or explicit refusal.
- [ ] Qualify AMD HIP and Intel SYCL against the same conformance suite. Keep
  Apple Metal limited to interactive preview unless its device capabilities can
  support the requested numerical claim; macOS verified runs retain CPU FP64.
- [ ] Evaluate Mojo 1.x using the project suite. Do not adopt it unless it reduces
  total maintained kernel and proof complexity while preserving native platform,
  packaging, precision, profiler, and validation requirements.
- [ ] Publish time-to-solution, cell-update, memory, transfer, energy where
  measurable, accuracy, and convergence reports with hardware, driver,
  compiler, flags, decomposition, and command provenance.
- [ ] Demonstrate one workstation multi-GPU decomposition only after
  single-device evidence is stable. Keep MPI and cluster libraries optional.

Exit gate: one nontrivial transient field case is scientifically inspectable
from the CLI, converges under independent refinements, conserves under its
declared tolerance, and runs faster on a named GPU without changing certified
observables beyond tolerance. The complete application remains CPU-capable.

## 1.0: Exceptional CLI Lab

- [ ] Ratify and ship a machine-readable 1.0 capability and content manifest.
  Its required scientific set is RP-1 and the supported Earth-discharge
  analytical and reduced models, named verified field cases earned before 1.0,
  and the ontology conformance packs from v0.8. Extreme planetary, stellar,
  relativistic, and broad fictional-law packs remain later capabilities unless
  separately promoted through every gate.
- [ ] Complete Quick Play, Broadcast, Freestyle, Challenge, Campaign,
  simulation, translation, proof, replay, and export command families.
- [ ] Complete Chill Mode as a calm append-only or live terminal experience with
  sparse ordinary events, long silence, optional radio, math-derived visuals,
  explicit recording, and no engagement mechanics.
- [ ] Make natural-language generation exceptional for humans and agents while
  preserving offline typed control, interpretation receipts, and explicit
  acceptance.
- [ ] Ship the generic Occurrence Card plus capability-selected Witness Card,
  Plumeprint, Fartflake, and print-ready local exports.
- [ ] Polish width-adaptive human output, errors, help, progress, plain mode,
  accessibility, hostile text, and pipe behavior.
- [ ] Ship named reviewed BCP 47 locale packs spanning left-to-right,
  right-to-left, and representative Han, kana, and Hangul script behavior, plus
  pseudo-locales and a complete locale-invariant machine surface.
- [ ] Freeze a consistent command grammar with short play modes, noun-then-verb
  scientific groups, contextual examples, stable error codes, suggestions that
  never auto-execute, `doctor`, configuration provenance, schemas, manuals, and
  shell completion.
- [ ] Ship `fart update check`, signed standalone `fart update`, package-manager
  delegation, dry-run, channel and version selection, atomic activation,
  interrupted-update safety, health check, receipt, and rollback.
- [ ] Protect update metadata and artifacts against rollback, freeze,
  mix-and-match, redirection, wrong-platform, version-confusion, size, signature,
  digest, and partial-write attacks. Updates never request elevation or run
  silently.
- [ ] Freeze a TUF-style root, timestamp, snapshot, targets, delegation,
  threshold, consistent-snapshot, trusted-time, rotation, revocation, mirror,
  transparency-provenance, and compromise-recovery architecture. Platform
  signing remains an additional check, not an alternative trust path.
- [ ] Publish p50 and p95 startup, first-output, analytical-run, cancellation,
  memory, and archive-size results on named Windows, macOS, and Linux systems.
- [ ] Publish a minimum CPU-only hardware, memory, storage, installation,
  ordinary-run, update, and optional-media budget. Ranked play fixes attempts,
  actions, branches, evaluation work, and exposed fidelity rather than wall time.
- [ ] Generate shell completions, manual pages, schemas, examples, and command
  reference from the typed command model.
- [ ] Ship deterministic offline audio without requiring an audio device.
- [ ] Ship physical audio, diagnostic sonification, Symphony, radio, semantic
  agent audio, and independent accessible controls without confusing their
  scientific status.
- [ ] Ship one G3 challenge and the `C-Sharp Correspondence` G4 research
  campaign with repeated-seed, transfer, recovery, and proof-quality reports.
- [ ] Publish agent schemas, protocol compatibility, threat models, conformance
  evidence, benchmark methodology, and matched human baselines.
- [ ] Add signed or notarized release artifacts, checksums, SBOM, provenance,
  install and uninstall tests, and named platform baselines.
- [ ] Publish the requirements trace, invariant registry, dependency inventory,
  hazard and threat analyses, SSDF mapping, independent verification evidence,
  reproducible-build comparison, and explicit non-claims as an assurance case.
- [ ] Complete archive fuzzing, dependency review, security scan, content review,
  accessibility review, and cross-platform replay tests.
- [ ] Freeze semantic versioning, command, JSON, JSONL, error-code, schema,
  archive-compatibility, deprecation, migration, supported-platform, minimum
  toolchain, release-line, and security-support contracts.
- [ ] Complete at least one nonsymbolic, non-acoustic play path through typed CLI
  and agent JSONL without fabricated social or Earth-discharge fields.
- [ ] Complete compensated disabled-player testing, remediation, and retesting
  for the full CLI golden path without scoring penalties for accessibility aids.

Exit gate: the CLI is a complete, exceptionally polished game and laboratory on
Windows, macOS, and Linux. No graphical interface is needed to experience the
product's identity.

## 2.0: Terminal Lab

- [ ] Add `fart-tui` with generic occurrence, participant, coupling, measurement, view,
  comparison, invariant, uncertainty, provenance, solver, and proof panes plus
  capability-selected Earth discharge, score, radio, Chill, agent, timeline,
  translation, and Broadcast views.
- [ ] Make the overview an htop-style live scientific instrument with stable
  spatial panes, dense scan paths, sortable instruments, focus, filter, freeze,
  compare, provenance drill-down, uncertainty, solver health, ledger closure,
  regime transitions, live Plumeprint, and copyable CLI equivalents.
- [ ] Support wide, standard, compact, and minimum layouts.
- [ ] Support Unicode and explicitly labeled ASCII transliteration or fallback
  locales, truecolor through reduced color, keyboard-only, remappable keys,
  optional mouse, and append-only screen-reader mode. ASCII never claims full
  locale parity when original-script text is unavailable.
- [ ] Restore the terminal after success, error, panic, resize, suspend, and
  interrupts.
- [ ] Add cell-buffer snapshots plus PTY and ConPTY smoke tests.
- [ ] Test `160x48`, `120x36`, `80x24`, `60x18`, below-minimum, Unicode, ASCII,
  grapheme, bidirectional text, hostile escape, high-contrast, reduced-color,
  append-only screen-reader, resize, suspend, interrupt, panic, and restoration
  cases.
- [ ] Publish measured p50 and p95 input-to-render latency, idle CPU, memory,
  allocation, and resize budgets on Windows, macOS, and Linux.
- [ ] Prove that every TUI action has a CLI, scenario, or archive equivalent.
- [ ] Run the same canonical action and observation conformance suite against
  TUI controls, live deltas, spectators, and station presentation.
- [ ] Complete compensated disabled-player testing, remediation, and retesting
  across interactive and append-only TUI paths.
- [ ] Open RP-1 and a non-Earth conformance occurrence with no gas, geometry,
  acoustics, mass, or localized emitter without empty Earth-only panes.

Exit gate: Terminal Lab has feature parity with CLI services, adds no solver or
story branches, and passes native terminal tests on all supported systems.

## 3.0: Native Reference Enclosure

- [ ] Create a native Godot app and thin Rust GDExtension over the existing core.
- [ ] Prohibit browser runtimes, embedded webviews, HTML UI, localhost UI
  servers, Electron, Tauri, and graphical-only simulation or gameplay logic.
- [ ] Build one biology-neutral reference enclosure with instruments and
  capability-driven projections. Ship the Tiled Chamber only as an optional
  authored Earth realization with responsive props, spatial audio, particles,
  deposition, recoil, haptics, and consequences.
- [ ] Ship one playable native inspection path for the v0.8 non-Earth,
  non-acoustic conformance occurrence. It must not fabricate a room, emitter,
  geometry, sound, haptics, or Earth-only pane.
- [ ] Build the full native Chill Mode with sparse generated events, restrained
  radio mixing, slow field and topology art, science overlays, world continuity,
  and independent motion, flashing, grossness, audio, and display-sleep controls.
- [ ] Use one curated and inspectable discharge path from pfft through supported
  dry, wet, choked, and underexpanded laboratory cases.
- [ ] Show live values, active groups, regime causes, uncertainty, and proof.
- [ ] Add remappable keyboard, mouse, and controller input plus first-run
  accessibility setup.
- [ ] Add real platform accessibility semantics and an opt-in, visibly indicated
  native automation driver for pixels-only and accessibility-assisted agents.
- [ ] Complete compensated disabled-player testing, remediation, and retesting
  before public native release.
- [ ] Prove screenshots, semantic nodes, input actions, audio captions, and the
  session reducer cite synchronized presentation revisions without exposing
  hidden state.
- [ ] Package, sign, smoke test, install, run, export, and cleanly uninstall on
  Windows, macOS, and Linux.

Exit gate: the native app reproduces the certified Reference Enclosure scenario
already available in CLI and TUI, labels every spatial view as a projection when
the source is lower-dimensional, higher-dimensional, discrete, or nonspatial,
and every consumer cites the same occurrence provenance.

## After the Reference Enclosure

### Native Broadcast and campaign

- [ ] Present Quick Play, Broadcast, Freestyle, and campaign in the native world.
- [ ] Add spatial staging, characters, translation cards, callbacks, and replay
  exhibits without native-only narrative truth.

### Extended field physics

- [ ] Extend the verified pre-1.0 solver to adaptive three-dimensional fields,
  moving interfaces, room interaction, and larger multi-device runs.
- [ ] Add MPI-based multi-node execution only as an optional headless profile
  with restartable checkpoints and complete evidence provenance.

### Multiphase and underwater packs

- [ ] Add parcels, evaporation, breakup, collision, deposition, bubble formation,
  Rayleigh-Plesset dynamics, and bubble acoustics with separate ledgers.

### Universal and extreme packs

- [ ] Add rarefied, orbital, planetary, reacting, plasma, MHD, stellar,
  relativistic, and gravitational profiles one verified model family at a time.
- [ ] Add alternate-dimensional and fictional-law profiles with explicit axioms,
  compatibility maps, and no false empirical-validation claims.

### Interdimensional Consortium

- [ ] Add G5 multi-agent campaigns with disjoint roles, communication budgets,
  agent dropout, held-out universes, and shared proof artifacts.
- [ ] Measure cold-start ability, improvement, retention, recovery, and transfer
  without collecting private reasoning traces.

### Apocalypse Mode

- [ ] Add laboratory, planetary, and stellar source budgets before adding damage.
- [ ] Add target material, self-gravity, fracture, radiation, and relativistic
  consequence systems only after each has a defensible contract.

## Definition of done

- Behavior is implemented, documented, and reproducible from a clean checkout.
- Formatting and lint are clean, tests pass, aggregate core statement coverage
  is at least 90 percent, every package remains above 80 percent, and critical
  mutants do not survive.
- Numerical work states applicable axioms, rules or equations, assumptions,
  units and dimensions where defined, validity, tolerances, verification, and
  validation status.
- New presentation consumes the authoritative occurrence provenance graph.
- New narrative claims cite context-scoped occurrence claims and cannot modify
  the occurrence account.
- New interfaces preserve CLI parity.
- Adapters over the canonical play service preserve transition and information
  parity for their declared track.
- Every milestone delivers at least one player-visible improvement, one better
  scientific claim or boundary, and one stronger verification result.
- Every named invariant maps to an executable check, owner, tolerance, and
  milestone before it can be marked implemented.
- Where a profile defines source energy or magnitude scaling, high-energy
  content changes the applicable source or budget model before presentation
  scale.
- Public changes pass secret, privacy, dependency, license, safety, and
  accessibility review appropriate to their surface.
