# Simulation contract

This document defines the scientific spine of F.A.R.T. Lab. It is a researched
design target, not a description of physics already implemented in this
repository.

## Core invariant

One event produces one authoritative physical truth represented by an immutable,
typed provenance graph. Every visible, audible, haptic, narrative, and
destructive effect must have a traversable derivation from that graph.

The graph is deliberately not one enormous master-rate array. Solver states,
interface samples, observer signals, audio blocks, narrative facts, and rendered
frames use their appropriate clocks and resolutions. Every derived node records
its input identities, transformation version, clock, resampling policy,
uncertainty, and content hash. Presentation can transform an outcome but cannot
substitute an unrelated preset or rewrite it.

## Source-neutral event ontology

The general contract is:

> A finite-time state transition coupling an emitter domain to an exterior
> domain across an interface, under a versioned law profile.

It is represented by these top-level objects:

- `LawProfile`: dimensions, metric or geometry, fields, governing equations,
  closures, constants, symmetries, conserved currents, sources, canonical units,
  capability set, compatibility class, validity domain, trust class, and
  verification suite.
- `WorldProfile`: boundaries, exterior phases, composition, radiation, gravity
  or other fields, initial state, and observer locations.
- `Emitter`: stored inventories and energy, state equations, containment,
  inertia, structural limits, forcing, and one or more boundary ports.
- `Interface`: topology, geometry, orientation, measure, compliance, actuation,
  contact, and surface behavior.
- `Payload`: carrier and dispersed phases, composition, sizes, loading, charge,
  reactivity, and material properties.
- `Observers`: sensors, bandwidth, sensory model, and presentation translation.
- `Numerics`: solver family, discretization, tolerances, refinement, thread
  policy, random streams, and determinism contract.
- `EventSignature`: normalized source histories, regime sequence, active
  dimensionless groups, impulse, spectra, loading, and directional structure.
- `EventGraph`: authoritative states and typed derived artifacts connected by
  versioned provenance edges.

A source can be an organism, machine, colony, ecosystem, spacecraft, planet,
star, distributed intelligence, topological structure, or unknown. A bathroom
is one scenario. A human is one preset. Neither is the engine's ontology.

## First validated law profile

The initial profile is `earth.continuum.si`, with three Euclidean spatial
dimensions, one time dimension, Newtonian mechanics, continuum gas dynamics,
and SI canonical storage. Under that profile, a fart event specializes to:

> A pressure-driven discharge from a deformable reservoir through a compliant
> aperture into an exterior domain.

Authoring formats may accept documented units. The normalized scenario is
stored in the law profile's canonical units. Dimensionless signatures remain
independent of the author's unit choice.

The input schema must not independently prescribe an inconsistent mass,
composition, pressure, temperature, and volume. It selects an independent
thermodynamic state set, an equation of state, and explicit validation
constraints. Derived quantities remain derived.

## Capabilities and trust

Every law profile declares a machine-readable `CapabilitySet`. Initial
capabilities include continuum, compressible, acoustic, multiphase, reacting,
radiative, rarefied, plasma, MHD, relativistic, gravitational, and
fictional-axiomatic support. A scenario is rejected before simulation if a
requested effect lacks a compatible implementation, closure, and verification
suite.

Initial law packs are compiled in and reviewed. Future third-party packs are
declarative data with bounded resources. They cannot load native code, perform
network access, select arbitrary filesystem paths, or claim a capability without
the required schemas and tests. Executable extensions require a separate threat
model and are outside the initial public contract.

## Coupled Earth-profile pipeline

### 1. Reservoir

Track species mass, internal energy, pressure, temperature, and deformable
volume. A baseline balance is:

```text
dm_i/dt = -Y_i,e * mass_flow + species_source_i

dU/dt = heat_flow - mass_flow * outlet_enthalpy
        - reservoir_pressure * dV/dt + energy_source
```

The first ideal-mixture closure may use:

```text
pV = m * R_mix * T
R_mix = sum(Y_i * R_i)
```

Every source term is named. Earth-biological presets use finite, plausible
budgets. Laboratory, spacecraft, planetary, and stellar profiles must select a
different source model rather than hiding impossible energy inside a body.

### 2. Interface

The first analytical model uses a prescribed effective area history and an
empirical discharge coefficient. It does not pretend to have solved a moving,
viscoelastic, separated flow boundary.

The next model adds a quasi-static compliance law. A later model may add a
damped interface oscillator with pressure work, contact, actuation, and stored
wall energy. Self-sustained oscillation requires aerodynamic phase feedback. A
one-mass forced oscillator alone is not enough to claim that behavior.

### 3. Restriction flow

For a quasi-steady, adiabatic, isentropic, calorically perfect gas through a
converging restriction, choking begins when:

```text
p_back / p_stagnation <= (2 / (gamma + 1))^(gamma / (gamma - 1))
```

This is the Choked Cheek Criterion. Choked means Mach 1 occurs at the
controlling section. It does not by itself mean the plume is supersonic,
underexpanded, shock-containing, or screeching.

For gamma 1.4, the critical ratio is about 0.528. At Earth sea-level pressure,
the upstream reservoir needs about 192 kPa absolute, or 90 kPa gauge. Published
ambulatory measurements report ordinary flatus-associated pressure excursions
around 0.93 kPa. An ordinary Earth-biological preset therefore cannot drift into
choking. A laboratory or fictional source pack must provide the missing energy.

The analytical mass-flow model uses a declared discharge coefficient, gas
properties, stagnation state, area history, and back pressure. Its full thrust
control surface includes both momentum and pressure thrust:

```text
thrust = mass_flow * exit_velocity + (exit_pressure - ambient_pressure) * exit_area
```

### 4. Starting jet, puff, and exterior

Ordinary finite events are starting jets or puffs, not eternal steady jets. The
source history exports mass, momentum, enthalpy, composition, area, and pressure
flux. After cutoff, the exterior evolves a finite impulse. The active signature
may include nondimensional stroke:

```text
L / D = integral(exit_velocity dt) / D
```

Developed round-jet centerline decay and entrainment laws are allowed only when
their far-field, unconfined assumptions apply. A tiled room with return flow and
walls must stop claiming infinite-domain similarity.

Open boundaries record exported mass, momentum, and energy. Room boundaries
account for frequency-dependent acoustic impedance, reflection, scattering,
deposition, heat transfer, and structural coupling as supported by fidelity.

### 5. Droplets and particles

Droplets and solid particles have separate mass and momentum accounts. The first
model uses deterministic seeded parcel populations with a documented drag law.
Later models add evaporation, collision, breakup, splash, adhesion, rebound,
and two-way fluid coupling.

Weber number concerns deformation and breakup after a droplet exists. Ohnesorge
number, density ratio, exposure time, acceleration history, droplet Reynolds
number, and the breakup correlation also matter. Stokes number measures response
to a declared flow timescale. It does not alone predict deposition.

Primary atomization needs a liquid film or source model. Deposition needs a
surface-interaction model. Evaporated mass and latent energy are transferred
back into the carrier ledger rather than disappearing.

### 6. Audio and observers

Coarse pressure and mass-flow histories do not uniquely determine unresolved
turbulent waveform detail. The reduced audio model may combine:

- Resolved interface motion for tones and modulation.
- Volume-flow and loading source terms where their acoustic approximations hold.
- A documented state-conditioned turbulent spectrum.
- Versioned random phases from a dedicated deterministic audio stream.
- Parameterized room propagation or a declared empirical impulse response.

That is procedural audio derived from event state plus a stochastic closure. It
is not a prerecorded emission. A resolved compressible solver can later provide
sources for an acoustic analogy or a directly resolved acoustic field.

Physical acoustics is distinct from diagnostic sonification, Symphony Mode, and
radio. Sonification declares a data mapping. Symphony declares an artistic
mapping. Radio is independent presentation. None can overwrite the physical
history or borrow scientific status from it. The complete audio contract is in
[AUDIO.md](AUDIO.md).

Vacuum has no propagating exterior acoustic wave. Structural vibration, sound
inside a pressurized vehicle, ballistic payload, and recoil can remain. A clean,
cold gas exhaust is not automatically visible. Visibility requires particles,
droplets, condensate, scattering, luminescence, plasma, radiation, or an
explicit scientific visualization.

## Independent regime axes

Labels are facts, not a single severity ladder:

- Material: dry, wet, shart, gas-carried solids, or reclassified fecal ejecta.
- Discharge: subsonic, choked, underexpanded, locally supersonic,
  shock-containing, or another supported regime.
- Thermochemical: ordinary gas, reacting gas, ionized equilibrium, plasma, or
  unsupported.
- Source: biological, laboratory, machine, spacecraft, planetary, stellar,
  distributed, fictional, or custom.
- Exterior: atmosphere, liquid, rarefied gas, vacuum, field-only, or fictional.

A gas jet may legitimately carry solid particles. Reclassification as fecal
ejecta occurs only when a versioned policy finds that continuous or plug-like
condensed material controls the event. Every label cites calculated state,
policy version, thresholds, and uncertainty.

Plasma is not a single temperature threshold. Ionization and the valid model
depend on composition, density, equilibrium, collision scales, magnetization,
and characteristic length. Stellar events require radiation and plasma or MHD
profiles. Neutron-star conditions require relativistic hydrodynamics, spacetime
curvature, and a dense-matter equation of state. They are never ordinary slider
values in the bathroom solver.

## Active dimensionless signature

A law profile declares base dimensions and dimension-checked equations. For a
dimension matrix `D`, dimensionless monomials have exponent vectors `k` such
that:

```text
D * k = 0
```

The implementation computes exact rational null spaces for discovery, but a raw
null-space basis is not a stable public contract because the basis is not
unique. Law-pack authors declare stable semantic groups obtained from
nondimensionalizing their equations and closures. Tooling proves that those
groups are dimensionless, checks their dependencies, and reports possible
omissions. It does not invent scientific meaning from a null space.

The active signature includes:

- Law, equation, closure, dimension, and topology identities.
- Deterministic reference scales.
- Active physical coefficients and semantic Pi groups.
- Normalized geometry, boundary conditions, forcing histories, material
  functions, and payload distributions.
- Validity indicators.
- A separate numerical signature for CFL, grid, timestep, limiter, tolerances,
  solver version, and thread policy.

An Earth gas profile may activate Mach, Reynolds, Strouhal, Euler, Froude,
Knudsen, Prandtl, gamma, pressure ratio, temperature ratio, compliance, and
normalized reservoir capacity. Multiphase models may add Weber, Stokes,
Ohnesorge, Bond, density and viscosity ratios, loading, volatility, size-shape,
and closure identities.

## Flatulence Similarity Law

Strict similarity means the source and target normalized problems are identical
within declared tolerance. It requires compatible equations, closures,
dimensions, topology, normalized geometry, material functions, initial and
boundary conditions, active coefficients, and solution branches.

Matching Mach and Reynolds alone does not make two compliant multiphase events
equivalent. Buckingham Pi analysis identifies possible dimensionless
combinations. It does not prove that the chosen variables are complete or that
the model is true.

## Universal Flatulence Translator

The translator offers three honest modes:

1. `strict`: solve only between compatible law-profile hashes with matching
   active signatures and normalized histories.
2. `approximate`: optimize target parameters for selected invariants, subject to
   validity and safety constraints, then report every residual and conflict.
3. `comic`: preserve declared observer experiences or cultural meaning and mark
   the result as a presentation translation.

Exact monomial matching can become a constrained linear problem in log space.
Approximate translation minimizes weighted log-ratio errors or exposes a Pareto
frontier. The preference objective chooses among solutions; it is not physics.

Each translation certificate records source and target law hashes, mode,
requested invariants, achieved residuals, discarded quantities, incompatible
fields, and whether the result is validated, extrapolated, or fictional.

Strict translation also emits a machine-checkable `CompatibilityWitness` that
identifies the mapped equations, dimensions, closures, topology, normalized
conditions, semantic groups, and tolerances. Approximate translation emits the
attempted witness plus unmatched capabilities, residuals, sensitivity,
nonuniqueness, and identifiability warnings. A close observable match is not
automatically a unique or physically equivalent inverse solution.

Different spatial dimensions generally change densities, interface measures,
wave propagation, turbulence, and geometric spreading. A literal 2D universe,
an axisymmetric reduction of 3D, and a 2D visualization are three different
things. Cross-dimensional translation is rejected unless an explicit mapping
exists.

## Identity and provenance

The project keeps six identities separate:

1. **Scenario identity:** normalized author intent, law profile, inputs, and
   seed before solving.
2. **Physical-result identity:** authoritative event graph under a solver,
   numerical contract, and declared tolerance policy.
3. **Narrative identity:** resolved world, situated perspectives, storylets,
   facts, and narrative streams.
4. **Presentation identity:** language, layout, camera, audio-device path,
   accessibility, and rendering choices.
5. **Play-session identity:** rules, initial identity, roles, ordered canonical
   action journal, checkpoints, branches, and produced artifacts.
6. **Archive-byte identity:** exact serialized container bytes.

A presentation change must not alter scenario or physical-result identity. An
archive may be migrated or recompressed without pretending that byte identity
was preserved. Cross-platform tolerant equality is a comparison result, not a
fake hash. Raw floating-point identity is promised only at a determinism level
that actually controls the arithmetic.

## Fidelity ladder

All levels share event vocabulary, hashes, regime policy, and certificate shape.
They do not pretend to share one universal equation set.

### A. Canonical analytical oracle

- Ideal-mixture finite-reservoir mass and energy balance.
- Prescribed area and simple compliance.
- Quasi-steady subsonic and choked restriction flow.
- Reduced starting-jet or puff integrals.
- Source flux, recoil, active signature, and global ledgers.

### B. Reduced coupled real time

- Interface oscillator with explicit fluid work.
- Integral plume, room response, and deterministic parcel statistics.
- State-conditioned procedural audio with a versioned stochastic closure.
- Simplified evaporation and deposition with explicit domains.

### C. Interactive field solver

- Conservative compressible finite-volume gas solver.
- Positivity protection and shock-capturing fluxes.
- Axisymmetric reduction before claiming quantitative 3D behavior.
- Moving or immersed boundaries and seeded particle parcels.
- Higher-order and LES options only with stated filters and validation.

An incompressible 2D solver cannot validate choking, shocks, propagation audio,
or compressible pressure damage.

### D. Specialist law packs

- Underwater bubble formation, hydrostatics, interface dynamics, and bubble
  acoustics.
- Rarefied and vacuum transport with continuum-breakdown criteria and ballistic
  or DSMC models where needed.
- Orbital recoil, torque, mass loss, and delta-v.
- Reacting, ionized, plasma, MHD, stellar, relativistic, and gravitational
  profiles, each with its own equations and verification suite.
- `fiction.axiomatic.*` profiles that are internally consistent under declared
  rules but never called empirically validated.

## Conservation of Ass

Each run declares its control volume and double-entry transfer ledgers:

- Mass or applicable conserved current by carrier, species, liquid, solid,
  deposited material, source, and open-boundary export.
- Momentum in the emitter, emitted phases, structures, boundaries, and export,
  with external forces identified.
- Internal, kinetic, potential, elastic, plastic, fracture, thermal, radiation,
  and exported energy as supported by the law profile.

For each conserved quantity `Q`:

```text
residual = Q_end - Q_start + outward_flux - declared_sources
```

The certificate reports absolute and normalized residuals. A diagnostic acoustic
energy estimate is not counted as transferred energy unless it is conservatively
removed from the fluid model. A wrong model can conserve perfectly, so
conservation is necessary but not sufficient.

## Determinism and seeds

Determinism has explicit levels:

- D0: same normalized scenario and seed have the same identity.
- D1: same build, target, settings, and thread policy replay bitwise.
- D2: supported platforms match declared observables within tolerances.
- D3: cross-platform bitwise identity, claimed only if arithmetic, reductions,
  transcendentals, and DSP actually guarantee it.

The engine uses fixed iteration order, deterministic reductions, exact unit
normalization where possible, and counter-based or independently keyed random
streams. Physics, parcel, physical-audio closure, score, radio, narrative, and
presentation streams are separate. Scheduling, terminal width, localization,
camera, station, and subscriber choices cannot alter the event identity.

## Uncertainty, sensitivity, and identifiability

Certificates keep aleatory variability, parameter uncertainty, model-form
uncertainty, discretization error, iterative error, roundoff, and presentation
randomness separate. Inter-fidelity disagreement is measured rather than hidden
inside one confidence score.

Sensitivity reports state which uncertain inputs control each observable and
where interactions matter. Inverse and translation commands report families of
solutions when observations do not identify a unique source. Extrapolation
beyond a closure's validation domain is visible at the value, chart, and
certificate levels. One seeded realization is never presented as an uncertainty
study.

## Certificate claims

Every claim independently reports `pass`, `fail`, `inconclusive`, or
`not-applicable`:

- Replayable under a declared determinism level.
- Internally consistent and dimensionally homogeneous.
- Code verified against exact, manufactured, or trusted reference cases.
- Solution verified by timestep, grid, parcel, or ensemble refinement.
- Empirically validated for a stated physical domain.
- Fictional-law consistent for a stated axiom pack.

The certificate includes scenario, law, equation, closure, solver, numerics, and
consumer hashes; active signature; extrema and positivity; local and global
ledgers; refinement; uncertainty channels; regime evidence; random stream keys;
unsupported effects; and provenance for audio, visuals, haptics, narrative, and
damage. A play receipt separately binds the knowledge policy, action journal,
budgets, branches, result artifacts, and score vector.

## Verification gates

The progressive suite includes:

- Unit-expression invariance and dimensional homogeneity.
- Exact null-space checks for discovered bases and dimensional checks for every
  authored semantic Pi group.
- Preflight rejection of unsupported capability combinations and untrusted
  executable content.
- Zero-area and equal-pressure zero-flow limits.
- Rigid adiabatic and isothermal reservoir limits.
- Continuity and the mass-flow plateau at the analytical choking boundary.
- Positive mass, species, density, pressure, temperature, and internal energy
  without silent clamping.
- Exact discrete cancellation for reservoir, plume, deposition, and recoil
  transfers.
- Correct starting-jet and puff limits.
- No exterior acoustic propagation in vacuum while structural vibration remains.
- Particle tracer, settling, ballistic, evaporation, and deposition balances.
- Underwater hydrostatic, Laplace-pressure, Rayleigh-Plesset, and bubble-resonance
  references when that pack exists.
- Sealed-system center-of-mass conservation and externally vented rocket limits.
- Deterministic random streams independent of execution order.
- Traversable provenance from every consumer artifact and lab-fact sentence to
  authoritative event nodes.
- Go-to-Rust oracle parity.
- Strict similarity scale families and exact translator round trips.
- Explicit infeasibility for incompatible targets and dimensions.
- Manufactured-solution order, shock relations, and separate time and grid
  refinement.
- Ensemble and integral comparisons after chaotic fields decorrelate.

Core solver and archive packages maintain at least 80 percent statement
coverage. Coverage supports verification but never replaces validation.

## Safety boundary

F.A.R.T. Lab is a game and educational simulator, not a medical, biological,
pressure-vessel, ignition, aerospace, or planetary-safety tool. Public challenges
must not encourage real pressure-vessel abuse, ignition, inhalation, harmful
biological experimentation, or contamination.

The research sources and their exact scope are listed in
[RESEARCH.md](RESEARCH.md).
