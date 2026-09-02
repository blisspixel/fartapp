# Verification, validation, and uncertainty

This document defines the evidence required before F.A.R.T. Lab calls a result
trustworthy. It is a plan and benchmark registry. The current toy CLI has not
implemented or passed the planned physical benchmarks.

The governing distinction is strict:

- Code verification asks whether the equations were solved as intended.
- Rule conformance asks whether a non-equational law or axiom pack was executed
  as declared.
- Solution verification estimates numerical error for one calculation.
- Validation asks whether a model represents physical observations within a
  stated domain.
- Uncertainty quantification reports what remains unknown.
- Software assurance asks whether the implementation and release process are
  controlled enough to trust those claims.

Conservation is necessary, but a perfectly conservative wrong model is still
wrong. Grid convergence is necessary for field claims, but a converged answer
can still converge to an unsuitable equation or closure.

## Claim vocabulary

Every certificate claim is one of `pass`, `fail`, `inconclusive`, or
`not-applicable`. It also contains:

```text
claim_id
observable, relation, or invariant and applicable units
axiom, rule, equation, closure, and implementation revisions as applicable
applicable parameter envelope
reference type and source
comparison method and tolerance
computed error and uncertainty
refinement or repetition evidence
known discrepancy
review status
```

`Validated` without a phenomenon, range, observable, reference, and uncertainty
is invalid wording. `NASA-grade`, `flight-qualified`, `medical-grade`, and
similar institutional claims are prohibited unless a competent external body
actually grants the applicable status. NASA standards and NIST guidance are
process references, not endorsements.

## Benchmark registry

Tolerances below are candidate acceptance gates for implementation planning.
Each must be ratified with precision, discretization, parameter envelope, and
reference uncertainty before the benchmark is marked active.

| ID | Model or invariant | Reference and observable | Candidate acceptance evidence |
| --- | --- | --- | --- |
| ONT-001 | Minimum bounded case | Case with no occurrence, realization, emitter, interface, exterior, gas, geometry, units, or presentation text | Schema, archive, certificate, and CLI round-trip without sentinel, fake-zero, or implicit operation fields |
| ONT-002 | Optional law structures | Discrete graph, atemporal, cyclic, and partially ordered constraint fixtures | Only declared ordering, state, metric, dimensional, and invariant capabilities appear; serialization order creates no source time |
| ONT-003 | Multiple law contexts | Two scoped contexts joined only by an explicit inter-law bridge | Bridge compatibility, representation, ordering, and conservation claims match the declared coupling; an undeclared bridge fails |
| OBS-001 | Measurement contract | Passive, distributed, probabilistic, and state-altering fixtures | Accessible results and back-action match the declared measurement operator and provenance; coupled measurement changes case-result identity |
| OBS-002 | Read-only view contract | Knowledge, privacy, accessibility, locale, and layout variations over one retained account | Permitted claims change as declared while scenario and case-result identities remain fixed |
| MAP-001 | Mapping taxonomy | Semantic, structural, signal, analogy, and every no-mapping reason | Stable code, evidence, loss, policy, and changeability fields agree across locales and surfaces |
| ALG-001 | Unit and coordinate invariance | Equivalent authored units and transformed coordinates | Canonical scenario and observables agree to the declared arithmetic contract |
| ALG-002 | Dimensional group basis | Exact rational dimension matrix | Exact null space, dimensionless authored groups, rank and basis recorded |
| RES-001 | Zero driving force | Equal reservoir and exterior pressure | Zero net mass flow and impulse within roundoff-scaled tolerance |
| RES-002 | Finite blowdown | Closed-form ideal adiabatic and isothermal limits | Mass, pressure, temperature, and energy errors below ratified analytical tolerance |
| NOZ-001 | Isentropic nozzle | Analytical subsonic mass flux | Integral mass flux within 1 percent inside the declared quasi-steady domain |
| NOZ-002 | Choked Cheek Criterion | Analytical critical pressure ratio | Continuous transition and mass-flux plateau within 1 percent |
| FSI-001 | Free compliant interface | Mass-spring-damper frequency and decay | Frequency and logarithmic decrement within ratified tolerance |
| FSI-002 | Coupled interface passivity | No active source below instability onset | No self-excited energy growth and complete fluid-structure energy account |
| JET-001 | Starting vortex | Canonical piston-cylinder formation case | Circulation and pinch-off trend agree inside that geometry's uncertainty |
| JET-002 | Source closure | Canonical stopping-vortex case | Deceleration and circulation trend reproduced, including closure impulse |
| GAS-001 | Uniform advection | Exact transported state | Observed spatial and temporal order match the scheme's stated order |
| GAS-002 | Acoustic wave | Linear wave speed, phase, and amplitude | Convergent dispersion and dissipation over a declared resolved band |
| GAS-003 | Sod shock tube | Exact Riemann solution | Correct wave ordering, shock speed, positivity, and convergent integral error |
| GAS-004 | Shock-vortex interaction | Published canonical solution | Conserved totals and expected resolved interaction without negative state |
| GAS-005 | Manufactured solution | Symbolically forced smooth field | Observed order reaches the documented asymptotic range |
| GAS-006 | Curved or adaptive mesh | Free stream and conservative transfer | No metric-generated source, positive state, conserved refinement transfer |
| PAR-001 | Particle response | Analytical Stokes drag relaxation | Tracer and ballistic limits plus convergent relaxation time |
| PAR-002 | Deposition ledger | Closed material balance | Emitted equals airborne plus deposited plus exported plus transformed mass |
| DRO-001 | Isolated evaporation | Declared `d^2` limiting case | Radius-squared slope and latent-energy transfer agree within closure domain |
| DRO-002 | Breakup resolution | Grid, parcel, and ensemble refinement | No claim below a declared reliable feature size; a grid-tied three-cell peak fails |
| DRO-003 | Dense loading | Published spray trend | Two-way coupling and vapor shielding compared only in matching regimes |
| BUB-001 | Static bubble | Hydrostatic and Laplace pressure | Radius and pressure balance within ratified tolerance |
| BUB-002 | Spherical dynamics | Rayleigh-Plesset reference | Radius history converges inside incompressible spherical assumptions |
| BUB-003 | Small oscillation | Minnaert frequency | Frequency within 1 percent in the isolated linear limit |
| BUB-004 | Detachment acoustics | Open 2026 bubble dataset | Radius, detachment timing, and pressure trace compared with experimental uncertainty |
| ACO-001 | Compact source | Analytical monopole and dipole fields | Level, phase, distance law, and directivity within the valid far field |
| ACO-002 | Acoustic analogy | Nested permeable control surfaces | Observer spectra agree within a ratified band, initially targeted near 1 dB |
| ACO-003 | Rectangular room | Analytical modal frequencies | Resolved modes within 1 percent before lossy-wall validation |
| ACO-004 | Vacuum lemma | Exterior with no material medium | Exactly no exterior propagating acoustic channel; recoil and internal vibration remain |
| SYS-001 | Conservation of Ass | Closed and open control volumes | Double-entry mass, species, momentum, and energy ledgers close to stated tolerance |
| SYS-002 | Cross-projection account coherence | Lab-account consumer audit | Every enabled audio, visual, haptic, damage, and fact output cites compatible retained claims and transformations; causal language appears only where declared |
| SYS-003 | Oracle parity | Go fixtures and Rust implementation | Declared observables match within exact or numeric tolerance across platforms |
| SYS-004 | Similarity family | Scaled same-law cases | Normalized observables agree only where full active similarity conditions hold |
| ARC-001 | Case archive | Round trip and adversarial corpus | Canonical members survive; traversal, bombs, duplicates, links, and corruption fail closed |
| ADP-001 | Surface transition parity | One canonical action journal through every adapter | Core revision, result, scientific artifacts, and certificate digests agree |
| ADP-002 | Surface information parity | Track-specific observation audit | No surface leaks verifier state or bypasses its declared sensor and budget policy |

The executable minimal opaque probe currently satisfies only the catalog and
pre-admission scenario-schema slice of `ONT-001`. It selects no case operation
and creates no record, archive, or certificate. Those retained-case round trips
remain planned and the benchmark stays a candidate.

The executable multi-law probe-limit fixture is evidence only for `SCN-003` in
the quality registry. It rejects a two-entry `contexts` array before parsing
either entry or beginning catalog resolution. It therefore does not satisfy
`ONT-003`, test bridge absence, or establish context compatibility. The later
full scenario contract must supersede this boundary rejection with scoped
contexts and explicit coupling evidence.

The executable minimal opaque unresolved-capability fixture satisfies `SCN-004`
only. It reaches exact law resolution and then ends at the minimum outer
envelope without fabricating a capability result or evidence record. Because
the report remains input-dependent and selects, admits, and executes no requested
case operation, it is not retained-case, archive, admission-refusal, policy, or
physical evidence.

## Numerical proof protocol

For a field result, the project stores at least three independently selected
resolutions in the asymptotic study when practical. Time, grid, parcel, and
ensemble effects are varied separately. Reports include observed order, a Grid
Convergence Index or justified alternative, iterative error, roundoff policy,
feature-resolution cutoff, and the exact observables used for comparison.

Chaotic fields are not required to preserve pointwise phase indefinitely.
Comparison then moves to conserved integrals, spectra, distributions, event
timing, topology summaries, and ensemble statistics. The switch is declared
before inspecting which metric makes a run look best.

Positivity must come from a scheme with stated admissibility conditions. Silent
clamping of negative density, pressure, mass, temperature, radius, or
concentration is a failed run. Recent positivity and entropy-stability methods
are research candidates, not automatically adopted dependencies:

- [Sayyari and Yamaleev, 2026](https://arxiv.org/abs/2608.20103) extends a
  positivity-preserving entropy-stable approach to implicit BDF2 dual time
  stepping for three-dimensional compressible viscous flow.
- [Yang and Fu, 2026](https://arxiv.org/abs/2604.21600) treats positivity,
  conservation, entropy behavior, curvilinear meshes, and adaptive refinement,
  while documenting a tradeoff between one mortar construction and provable
  entropy stability.
- [The Atomizing Pulsed Jet](https://arxiv.org/abs/2405.01959) reports a
  nonconvergent small-droplet peak near three grid cells, which becomes a
  permanent artifact-detection regression case.

## Physical validation protocol

Every empirical source enters an evidence ledger with publication state,
geometry, materials, parameter range, instruments, raw-data availability,
reported uncertainty, closure supported, transfer limitation, permitted claim,
and prohibited claim. Fitting and evaluation data remain separated.

Biological measurements define uncertain priors, not a universal human norm.
Daily gas volume is not a single-event source volume. A small composition or
manometry study cannot calibrate an unsampled pressure, aperture, flow, and
audio waveform. No human or animal data collection begins without independent
ethics review, informed consent, data minimization, and a real scientific need.

Underwater validation uses the open data associated with the 2026 detaching
bubble work before claiming physical audio validation. Shock, particle, spray,
room, and compliant-interface cases are matched to their actual geometry and
range. Analogy is documented as analogy.

## Software assurance gates

The active implementation must provide:

- Requirements and named invariants traced to tests and evidence.
- At least 90 percent aggregate statement coverage, at least 80 percent per
  package, and higher changed-core targets defined in [QUALITY.md](QUALITY.md).
- Mutation testing for conservation, archive, unit, regime, and protocol logic.
- Property, metamorphic, fuzz, differential, race, cancellation, and fault tests.
- Static analysis, dependency review, vulnerability review, SBOM, provenance,
  signed releases, and reproducible-build comparison.
- Cross-platform fixtures on Windows, macOS, and Linux.
- Threat, hazard, failure-mode, and recovery analyses proportionate to the
  affected surface.
- Independent review before elevating any empirical-validation or high-energy
  safety claim.

The assurance structure is informed by
[NASA-STD-8739.8B](https://standards.nasa.gov/standard/NASA/NASA-STD-87398),
[NIST SSDF 1.1](https://csrc.nist.gov/pubs/sp/800/218/final), and
[NISTIR 8298](https://doi.org/10.6028/NIST.IR.8298). Conformance mappings will
identify applicable, non-applicable, implemented, and missing practices. They
will not imply NASA, NIST, aerospace, medical, or regulatory approval.

## Evidence storage and release

Benchmark inputs, reference licenses, scripts, raw outputs, environment facts,
compiler and solver versions, command lines, hashes, and reports are content
addressed. Large public datasets may be fetched through a pinned manifest rather
than committed. A release claim cites the exact evidence bundle that earned it.

A failed or inconclusive benchmark is retained as evidence. No result is hidden
because it spoils a joke, an image, or a planned version number.
