# Simulation contract

This document defines the scientific spine of F.A.R.T. Lab. It is a researched
contract. Most sections are design targets. The explicitly labeled v0.7
finite-reservoir endpoint subset is implemented; no other planned physics is
implied by that slice.

## Core invariant

Each accepted Lab case has one authoritative account under its declared law
contexts, accepted operations, numerical policies where applicable, and
measurement interactions, represented by an immutable typed provenance graph.
The graph records only the claims and relations the selected laws and
measurement contexts permit. It does not require that the source contexts
define an occurrence, realization, observer, simulation, or observer-independent
truth. Every visible, audible, haptic, narrative, destructive, or other
applicable derived effect must have a traversable derivation from that graph.

The graph is deliberately not one enormous master-rate array. Applicable solver
states, interface samples, measured signals, audio blocks, narrative claims,
and rendered frames use their declared ordering references, dependency
coordinates, clocks, and resolutions where those concepts exist. Every derived
node records its input identities, transformation version, applicable ordering
or dependency reference, resampling policy when relevant, uncertainty, and
content hash. Presentation can transform an outcome but cannot substitute an
unrelated preset or rewrite it.

### Contract ownership

The Lab keeps different claim owners separate. One owner cannot mint another
owner's result merely because the data appears in the same account.

| Owner | May establish | Cannot establish by itself |
| --- | --- | --- |
| Lab operation | Inputs consulted, stages performed, artifacts retained, and execution provenance | Source-law occurrence, physical truth, or permission |
| Law context | Applicable structures, rules, quantities, invariants, and supported conclusions | Implementation availability, authorization, or evidence sufficiency |
| Measurement interaction | Accessible observables, support, disturbance, noise, and uncertainty | Unmeasured state or observer-independent completeness |
| Mapping profile | Correspondence, preservation, residual, ambiguity, loss, and exact failure | Source identity, global compatibility, or semantic equivalence not declared by the witness |
| Play service | Legal actions, roles, knowledge, budgets, score, and journal transitions | Scientific state outside accepted operations or privileged hidden facts |
| Presentation and view | Authorized selection, layout, language, sonification, analogy, and accessibility transformation | New scientific geometry, probability, evidence, or source-law meaning |
| Narrative | Character, institution, and story claims tied to retained facts or declared fiction | Lab endorsement or scientific evidence |
| Evidence profile | What a test, proof, comparison, calibration, or review supports inside its scope | A stronger claim class, wider domain, or unrelated kind of evidence |

Provenance can record that one Lab claim was computed from another without
asserting that the first caused the second in the represented context.

## Minimum-assumption bounded record ontology

The general contract is:

> The Lab identifies a finite record under one or more versioned law contexts.
> Those contexts alone declare whether occurrence, realization, source-law
> identity, ordering, state, participants, couplings, locality, observables,
> transformations, invariants, and representations exist.

A declared classifier may map the supported claims in a Lab case to an
`EmissionAnalogue`, such as a release, transfer, relaxation, or boundary
crossing. `Occurrence` and `Realization` are context capabilities, not minimum
record fields. `DischargeEvent` is an Earth continuum specialization, not the
superclass of existence.

Only four top-level objects are required inside the Lab's bounded formalism:

- `RecordIdentity`: a unique Lab-local record identity. Optional
  `ContextOccurrenceIdentityClaims` contain identities only for scoped
  contexts that define occurrence identity and may therefore be empty. A
  context that does not define identity receives an explicit inapplicable
  result, never a fabricated identifier. A composite identity exists only when
  an `InterLawCoupling` supplies its equivalence, recurrence, cyclicity, or
  ordering semantics. Here `source` means the source law context, not a
  localized emitter, origin object, or causal source role.
- `LawContextSet`: one or more scoped `LawContext` entries. Each entry declares
  axioms or rules, supported structural modules, capabilities, interpretation or
  validity domain, implementation binding, trust class, and verification suite.
  Every `InterLawCoupling` owns its bridge rules, compatibility, ordering,
  representation conversion, and cross-context conservation claims.
- `Scope`: an addressable application boundary declaring whichever participants,
  relations, couplings, measurement interactions, or explicit absences apply.
  It need not contain spatial domains, localized objects, or any one of those
  optional categories.
- `ProvenanceGraph`: typed versioned edges connecting whichever authoritative
  states or relations, observations, claims, transformations, uncertainty,
  measurement back-action, or derived artifacts are applicable. These are
  Lab-evidence relations unless a law context separately licenses a source-law
  interpretation.

The formal boundary is explicit: a purported reality that cannot be identified,
finitely encoded, versioned, scoped, or related to anything expressible in this
schema is outside the representable ontology. The result code is
`outside_representable_ontology`, not `unknown`, `law_incompatible`, or a claim
of successful simulation.

A law context may independently declare modules for ordering or time, state or
object space, dimension, adjacency, topology, metric, locality, fields, units,
equations or transition rules, constants, symmetries, invariants, and conserved
currents. Any can be inapplicable. Schemas use typed option and capability
results, never sentinel Earth values or fabricated zeros.

Profiles may add `Occurrence`, `Realization`, `Participant`, `Coupling`, `StateRegion`,
`MeasurementInteractionProfile`, `ViewProfile`, `PresentationProfile`,
`Numerics`, and `ComparisonSignature` objects. The Earth discharge profile adds
`WorldProfile`, `Emitter`, `Interface`, `Exterior`, and `Payload` roles. Those
roles are not universal requirements.

A participant, where that concept exists, can be an organism, machine, colony,
ecosystem, spacecraft, planet, star, distributed intelligence, topological
structure, or something the Lab cannot classify further. A bathroom is one
Earth scenario. A human is one Earth preset. Neither is the engine's ontology.
The full boundary and negative-space conformance matrix are in
[UNIVERSALITY.md](UNIVERSALITY.md).

## First validation-target law profile

The initial target profile is `earth.continuum.si`, with three Euclidean spatial
dimensions, one time dimension, Newtonian mechanics, continuum gas dynamics,
and SI canonical storage. Under that profile, an emission analogue specializes to:

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

Every law context declares machine-readable concepts and capabilities. The
eventual canonical `CapabilityReport` keeps eight questions separate:

1. Does the law define the concept?
2. Does an implementation exist?
3. Does the requested model have an applicable closure?
4. Is it applicable to this scenario and any declared measurement interactions?
5. What verification and validation grade supports the requested claim?
6. Does policy permit and trust the pack and operation?
7. Can the selected backend satisfy the required implementation, precision, and
   determinism class?
8. Can available resources satisfy memory, storage, and work budgets?

Each entry returns typed availability or refusal evidence rather than one
Boolean. The capability vocabulary can represent continuum, compressible,
acoustic, multiphase, reacting, radiative, rarefied, plasma, MHD, relativistic,
and gravitational concepts without claiming an implementation exists.
Fictional-axiomatic packs declare their own concepts. Document validity remains
separate from operation selection, admission, and execution. A structurally
valid case can be unavailable, inapplicable, refused, infeasible, or
undetermined for a requested operation if it lacks a compatible law,
implementation, closure, verification grade, trust decision, backend, or
resource budget. Realization is only one law-selected operation kind.

Initial law packs are compiled in and reviewed. Future third-party packs are
declarative data with bounded resources. They cannot load native code, perform
network access, select arbitrary filesystem paths, or claim a capability without
the required schemas and tests. Executable extensions require a separate threat
model and are outside the initial public contract.

Authority classes never collapse into one impressive-sounding maturity label:

- An **empirical physical pack** may earn validation claims only inside a
  measured envelope.
- A **mathematical or formal pack** may earn proof and conformance claims, not
  empirical truth by construction.
- A **fictional axiomatic pack** may earn internal-consistency and
  implementation-conformance claims while remaining explicitly fictional.
- A **narrative pack** may shape story and presentation but cannot mint
  scientific state, evidence, or source-law meaning.

Every scientific component also carries one application-level
`ModelClaimClass`. This classification records the relationship between a
model and the Lab's reviewed evidence. It is not an axiom inside the source law
and does not claim that every context recognizes human scientific institutions:

- `established_physics` identifies a physical theory or model with empirical
  support for a named observable inside a stated envelope. It does not validate
  a new scenario merely because the governing equation is well established.
- `research_toy` identifies a published reduction or deliberately simplified
  problem that can be executed and studied without claiming world fidelity.
- `speculative_model` identifies a mathematically stated candidate physical
  model whose physical realization has not been established.
- `analogy` identifies a mapping that preserves only its enumerated structures,
  observables, or equations.
- `fictional_axiom` identifies deliberately invented rules that can earn only
  the declared internal-consistency and conformance claims.

Claim class, empirical support, scenario applicability, and implementation
assurance are independent fields. A verified implementation can solve a
speculative model correctly. An established equation can be extrapolated
outside its validated domain. A useful analogy can preserve one wave equation
without preserving dynamics, ontology, or empirical status. CLI, TUI, native,
agent, archive, and certificate surfaces retain these distinctions rather than
reducing them to one confidence or maturity score. Each status names its source
revision and Lab review date because scientific assessment can change without
changing a law-pack identity.

Several compatible models may address one requested observable. The Lab keeps
model definition, implementation, run, ensemble, and comparison identities
separate. Initial-condition, parameter, stochastic-closure, numerical,
structural, learned, and multi-model ensembles have different uncertainty
semantics and cannot be pooled by default. Agreement is not validation, and
disagreement is a reportable result. Learned surrogates, closures, hybrid
solvers, inverse estimators, and probabilistic generators retain their training
domain, calibration, physical constraints, and out-of-distribution refusal.
The full contract is in [MODELS.md](MODELS.md).

Every pack publishes the applicable subset of formal semantics, admissibility,
examples and counterexamples, implementation status, invariant obligations,
observable and measurement definitions, validity envelope, uncertainty, known
failure modes, unsupported claims, and versioned evidence. Advanced vocabulary
without a player action, an inspectable claim, or an executable obligation does
not enter the scientific surface.

## Measurement, view, and presentation profiles

A `MeasurementInteractionProfile` is an optional accepted scenario input. When
present, it declares
support or extent, an ordering or clock when one exists, accessible observables,
measurement operator, resolution, noise, and whether interaction is passive,
coupled, or state-altering. Any back-action enters the provenance graph as a
law-governed coupling and changes case-result identity.

A `ViewProfile` is a read-only knowledge, privacy, accessibility, and selection
projection over retained claims. A `PresentationProfile` controls locale,
layout, sonification, camera, and rendering. Neither can alter the Lab account
or its identity. Two views may expose different authorized subsets or
representations of the same retained context-scoped claims without creating two
accounts. A law context can declare no observer, no
localization, no sound, no language, or no proposition-like semantics.
Presentation adapters consume only the capabilities that exist.

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

#### Implemented v0.7 endpoint subset

The Go oracle now implements one narrow standalone model identified as
`continuum.rigid-calorically-perfect-ideal-mixture@v0alpha1`. It predicts an
endpoint after a prescribed composition-preserving finite withdrawal from a
rigid, homogeneous, nonreacting, single-gas-phase reservoir. Each component has
positive initial mass `m_i,0`, specific gas constant `R_i`, and constant
isochoric heat capacity `cv_i`. Volume `V`, initial temperature `T_0`, closure,
and withdrawal fraction `f` are explicit SI inputs. No Earth, source, exterior,
ambient, body, case, or occurrence is supplied.

With retained fraction `r = 1 - f`, the component masses and mixture properties
are:

```text
m_0    = sum(m_i,0)
m_i,1 = r * m_i,0
R_mix  = sum(m_i,0 * R_i) / sum(m_i,0)
cv_mix = sum(m_i,0 * cv_i) / sum(m_i,0)
cp_mix = cv_mix + R_mix
gamma  = cp_mix / cv_mix
```

For the rigid adiabatic closure:

```text
T_1   = T_0 * r^(R_mix / cv_mix)
p_1   = p_0 * r^gamma
H_out = m_0 * cp_mix * T_0 * (1 - r^gamma) / gamma
Q_in  = 0
```

For the prescribed-isothermal closure:

```text
T_1   = T_0
p_1   = p_0 * r
H_out = f * m_0 * cp_mix * T_0
Q_in  = f * m_0 * R_mix * T_0
```

Rigid boundary work is zero. The reported energy account uses the sign
convention:

```text
U_1 + H_out - Q_in - U_0 = 0
```

The implementation computes state relations, integrated transfers, component
mass balances, total mass balance, energy balance, and endpoint ideal-gas
reported-state equation-of-state consistency residuals through separately
expressed paths where practical.
It rejects nonfinite or nonpositive quantities, complete depletion, forged
states, and a positive fraction too small to produce representable progress.
An authored zero withdrawal is an exact identity rather than a missing value.

This is an analytical endpoint oracle, not a time-resolved blowdown model. It
has no aperture, discharge coefficient, flow rate, elapsed duration, exterior
pressure, choking, wall heat-transfer law, stratification, selective outflow,
reaction, phase change, recoil, plume, or acoustics. Its tests are code
verification against closed forms and properties. They are not empirical
validation. The complete `RES-002` benchmark remains open.

### Composition, chemistry, phase, and exposure layers

Composition is not one universal odor or danger slider. The Earth profile keeps
the following capabilities independent and couples them only through declared
ports:

- `BulkMixtureThermodynamics` owns the major species that determine mass,
  equation of state, heat capacity, enthalpy, density, and any supported
  transport properties. The first ideal-mixture reservoir belongs only to this
  layer and remains nonreacting, single-phase, and calorically perfect.
- `TraceSpeciesInventory` owns species-resolved trace masses or concentrations,
  analytical detection limits, uncertainty, and source basis. A trace species
  may be passive and one-way coupled only when a declared bound shows that its
  omission from bulk mass, energy, and properties is acceptable. Otherwise it
  enters the bulk mixture.
- `ReactionNetwork` declares `none`, `equilibrium`, or a versioned finite-rate
  mechanism, together with species closure, rate or equilibrium data, validity,
  reversibility, and coupling to energy and transport. An equilibrium result is
  not evidence for finite-rate evolution.
- `PhaseEquilibriumAndCondensation` owns saturation or fugacity assumptions,
  phase stability, nucleation policy, condensed-phase composition, and latent
  energy. Condensate formed after release remains distinct from liquid already
  present in the payload.
- `AerosolAndDropletChemistry` distinguishes primary emitted condensed material
  from secondary aerosol. Gas-particle partition, dissolution, evaporation,
  heterogeneous or aqueous reaction, coagulation, and deposition activate only
  when their own closures exist.
- `ExteriorChemistry` owns ambient composition, oxidants, humidity, radiation
  or photolysis fields, surfaces, mixing, and its mechanism revision. Source
  composition, transformed plume composition, and composition at a measurement
  support are separate states.
- `ObserverDetection` maps a physical field through an explicit sensor or
  observer model. Compound-dependent thresholds, adaptation, fatigue, masking,
  resolution, timing, and uncertainty belong here. Odor is not an intrinsic
  scalar property of a species or plume.
- `HazardAssessment` is a separate optional interpretation requiring species,
  concentration, averaging duration, route, endpoint, applicable population,
  authority, revision, and uncertainty. Missing context returns
  `insufficient_assessment_context`; the initial simulator does not invent a
  health score.

Carrier, trace, suspended, deposited, reacted, condensed, and exported ledgers
remain separately auditable. Detectability never implies toxicity, and lack of
detection never implies safety. Occupational or emergency exposure values are
context-scoped references, not universal game thresholds. A bioaerosol or
pathogen claim requires a separate biological capability and evidence review.
Composition cannot be used to infer identity, diet, disease, or protected
traits.

### 2. Interface

The first analytical model uses a prescribed effective area history and an
empirical discharge coefficient. It does not pretend to have solved a moving,
viscoelastic, separated flow boundary.

The next model adds a quasi-static compliance law. A later model may add a
damped interface oscillator with pressure work, contact, actuation, and stored
wall energy. Self-sustained oscillation requires aerodynamic phase feedback. A
one-mass forced oscillator alone is not enough to claim that behavior. Over one
candidate cycle, the diagnostic is:

```text
cycle_energy = fluid_work - structural_dissipation - fluid_dissipation
```

A stable tone can be classified as a self-excited limit cycle only after the
coupled model sustains positive growth and reaches a bounded attractor. A sound
designer cannot switch that label on.

### Source morphology and material parameter patches

The scientific source creator edits neutral source, interface, and material
parameters rather than selecting a hidden body type. Supported parameters may
include reservoir volume, shape, and compliance; opening area, equivalent or
hydraulic diameter, perimeter, aspect ratio, topology, and multiplicity;
channel length and wall thickness; lip curvature and separation geometry;
surface roughness and wettability; stiffness, damping, prestress, contact and
closure laws; surrounding baffle curvature and thickness; orientation; and
prescribed motion. Payload inventory remains a separate parameter group.

A humorous preset is a versioned pair:

- `PresetPresentation` contains a localized name, art, joke, cosmetic shell,
  and camera treatment. It never enters the scientific state or case hash.
- `MorphologyParameterPatch` is an explicit scenario diff with units and closure
  requirements. Applying it changes the scientific case identity and remains
  inspectable through CLI, TUI, native, agent, and archive surfaces.

Labels such as Big Butt, Small Butt, and Different Butt are optional localized
presentation aliases over named multidimensional patches. They are not
scientific types, a binary universal size axis, or evidence of a normal body.
Non-ordered variants may differ by topology, asymmetry, compliance, curvature,
or multiplicity. A patch cannot silently choose composition, odorants, wetness,
cleanliness, toxicity, diet, health, pressure, intelligence, culture, sex,
gender, race, species, or social value. Locale and cosmetic presentation cannot
change the patch, and two presentations over an identical patch must produce
identical scientific results. The biology-neutral reference interface remains
the default.

### 3. Restriction flow

For a quasi-steady, adiabatic, isentropic, calorically perfect gas through a
converging restriction, choking begins when:

```text
p_back / p_stagnation <= (2 / (gamma + 1))^(gamma / (gamma - 1))
```

The English presentation layer calls this the Choked Cheek Criterion. The
scientific state records a neutral choking-regime code. Choked means Mach 1
occurs at the controlling section. It does not by itself mean the plume is
supersonic, underexpanded, shock-containing, or screeching.

For gamma 1.4, the critical ratio is about 0.528. At Earth sea-level pressure,
the upstream reservoir needs about 192 kPa absolute, or 90 kPa gauge. Published
ambulatory measurements report ordinary flatus-associated pressure excursions
around 0.93 kPa. That optional Earth-biological preset therefore cannot drift
into choking. The biology-neutral calibration fixture, laboratory packs, and
fictional profiles each declare their own finite energy budgets.

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

Source closure is part of the event. Deceleration can generate a pressure-driven
stopping vortex and change impulse, recoil, trailing structures, and sound. The
solver therefore advances through opening, supply, deceleration, closure, and
post-closure evolution rather than deleting the jet when mass flow reaches
zero. The familiar vortex-ring formation number near `L / D = 4` is retained as
a canonical piston-cylinder benchmark, not a universal constant for arbitrary
apertures, compressibility, or source histories.

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
number, density and viscosity ratios, exposure time, strain history, droplet
Reynolds number, turbulence intermittency, and neighboring droplets also
matter. The English presentation alias Wetness Transition therefore describes a
probability distribution over regimes and breakup times, never one universal
Weber threshold. Stokes number
measures response to a declared flow timescale. It does not alone predict
deposition.

Primary atomization needs a liquid film or source model. Deposition needs a
surface-interaction model. Evaporated mass and latent energy are transferred
back into the carrier ledger rather than disappearing.

Evaporation progresses from a labeled isolated-droplet `d^2` limiting model,
through coupled heat and mass transfer, to multicomponent and vapor-shielded
models. The active closure is visible in every result. Dense parcels do not
silently inherit dilute assumptions.

Each surface owns a material state machine:

```text
approach -> impact -> rebound | stick | spread | splash
stick -> roll | slide | evaporate | dry | resuspend
```

Transitions cite impact kinematics, material properties, surface state, contact
angle, roughness, adhesion, temperature, and the applicable dimensionless
groups. Contact transfers mass, momentum, and energy instead of making parcels
vanish.

### 6. Underwater branch

Discharge into a liquid is not a gas jet with a blue color grade. The source
creates a connected gas cavity, neck, detaching bubbles, or a bubble cloud. The
first specialist model uses Rayleigh-Plesset dynamics within its incompressible,
spherical limits:

```text
rho_l * (R * d2R/dt2 + 3/2 * (dR/dt)^2)
  = p_bubble - p_infinity - 4 * mu_l * (dR/dt) / R - 2 * sigma / R
```

The small-amplitude isolated-bubble check uses the Minnaert frequency. Sound
amplitude is driven by volume and neck dynamics, not an arbitrary sine wave.
Strong compressibility, shocks, nearby walls, nonspherical collapse, and
interacting bubbles trigger a higher model or an explicit refusal. Later tiers
may use Keller-Miksis, Gilmore, boundary, and coupled-bubble models with separate
validation.

### 7. Audio and observers

Coarse pressure and mass-flow histories do not uniquely determine unresolved
turbulent waveform detail. The reduced audio model may combine:

- Resolved interface motion for tones and modulation.
- Volume-flow and loading source terms where their acoustic approximations hold.
- A documented state-conditioned turbulent spectrum.
- Versioned random phases from a dedicated deterministic audio stream.
- Parameterized room propagation or a declared empirical impulse response.

That is procedural audio derived from the Earth occurrence account plus a
stochastic closure. It is not a prerecorded emission. A resolved compressible solver can later provide
sources for an acoustic analogy or a directly resolved acoustic field.

The physical-audio source graph distinguishes interface loading and volume
velocity, turbulent quadrupoles, loading dipoles, mass or heat monopoles,
shock-associated broadband noise, shock-cell feedback tones, bubble-volume
sources, impacts, structural vibration, propagation, room response, and the
observer transfer. Each active edge records its approximation and validity
domain.

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

## Independent Earth-discharge regime axes

Neutral codes are facts, not a single severity ladder:

- Material: single-phase carrier, dilute dispersed loading, dense dispersed
  loading, gas-carried solids, continuous condensed transfer, or unsupported.
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

Optional English presentation aliases map single-phase carrier to dry, dilute
droplets to wet, dense depositional loading to shart, and continuous biological
condensed transfer to the Fecal Ejecta boundary notice. Those words never enter
scientific identifiers, archives, or cross-locale comparisons.

Plasma is not a single temperature threshold. Ionization and the valid model
depend on composition, density, equilibrium, collision scales, magnetization,
and characteristic length. Stellar events require radiation and plasma or MHD
profiles. Neutron-star conditions require relativistic hydrodynamics, spacetime
curvature, and a dense-matter equation of state. They are never ordinary slider
values in the Earth continuum gas solver.

## Relativity, compactification, strings, and cosmology

These profiles do not form one spectacle ladder. Relativistic flow in four
dimensional spacetime, a higher-dimensional classical field theory, perturbative
string theory, semiclassical gravity, a candidate quantum-gravity model, an
analogue system, and a fictional portal have different equations and evidence.
Selecting one never silently enables another.

### Compact dimensions

An exact compactification case declares all of the following before evaluation:

- Total spacetime dimension and the split between extended and compact factors.
- Compact topology, differentiable and metric structure where applicable,
  radii or moduli, orientation, and identifications.
- Complete field content, action or equations, couplings, localization, and
  gauge structure used by the calculation.
- Initial and boundary conditions, compact-coordinate boundary conditions, and
  any truncation of the mode tower.
- Which observer-accessible quantities and comparison maps are defined.

For a free scalar on a circle of radius `R`, a research-toy pack can verify the
Fourier basis and, in natural units, the Kaluza-Klein spectrum
`m_n^2 = m_0^2 + n^2 / R^2`. It can report degeneracy, truncation error,
effective lower-dimensional couplings, and the zero-mode limit. Energy can
leave a lower-dimensional sector only when the selected action supplies an
explicit coupling to bulk modes. The calculation does not establish that the
compact factor, mode, or coupling exists in nature. Arbitrary-dimensional fluid
dynamics and Kaluza-Klein field reduction are not automatically string theory.

### String and brane packs

A string pack identifies the exact formulation rather than exposing a generic
dimension slider. It declares whether it implements bosonic, Type I, Type II,
heterotic, or another stated theory; its critical dimension; background fields
and consistency conditions; string tension or `alpha_prime`; string coupling;
compactification; open or closed sectors; boundary conditions; gauge choice;
quantum state; and approximation order. A perturbative bosonic-string toy uses
26 spacetime dimensions and contains a tachyon; perturbative superstring packs
use their theory-specific ten-dimensional consistency conditions. An arbitrary
dimension or background is rejected unless a separate formal or fictional pack
defines it.

The smallest quantum string research toy is a free compact worldsheet boson. It
may calculate oscillator, momentum, and winding levels and test level matching
and the declared T-duality map `R <-> alpha_prime / R`. A classical
Nambu-Goto pack may calculate worldsheet motion, constraints, and conserved
quantities in a prescribed background. Neither turns a gas plume into a
fundamental string or creates physical audio. Any audible rendering of string
modes is an `analogy` or sonification with an explicit transfer function.

A brane pack declares dimension, embedding, worldvolume fields, charges,
boundary conditions, tension, bulk couplings, and effective action. A bounded
Dirac-Born-Infeld or linear worldvolume calculation is a research model, not a
claim that a brane is a material membrane, discovered portal, or route between
universes. The Standard Model, source, or payload remains brane-localized only
when the selected model says so.

### Gravity and black holes

Special and general relativity are physical theories with extensive empirical
support, but each calculation still declares its regime. A fixed-background
Schwarzschild or Kerr profile can calculate geodesics, redshift, causal reach,
capture, and radiation transport while explicitly neglecting backreaction. It
must refuse when the release materially changes the metric.

Black-hole formation cannot be certified from total energy or one compactness
number alone. Distribution, stress, pressure, angular momentum, geometry, and
the gravitational constraint and evolution equations matter. A formation claim
requires a dynamical spacetime solver, compatible matter model, constraint
convergence, horizon diagnostics, and an applicable validation argument.

Hawking temperature and evaporation are a separate semiclassical
quantum-field-on-curved-spacetime capability. They are not direct empirical
observations of an astrophysical evaporation event. String microstate claims
remain restricted to the exact supersymmetric or extremal configurations for
which the counting was derived. Low-energy quantum-gravity effective-field
calculations remain separate from unknown ultraviolet completion. At Planckian
curvature or energy, a physical pack without a justified model returns
`outside_validated_theory` rather than extrapolating.

### Cosmological initial data

A cosmology profile declares a spacetime model, metric, matter and radiation
components, equations of state, interactions, initial-data surface, perturbation
order, gauge, observables, and validity range. Homogeneous FLRW evolution can
calculate scale factor, density dilution, redshift, horizons, thermal history,
and interaction-rate freeze-out under those assumptions. Linear perturbation
theory activates only while its small-perturbation conditions hold.

A homogeneous Big Bang-scale case is not a pressure discharge from a localized
reservoir into an exterior. When a request replaces the exterior with the
initial state of the modeled cosmos, the engine reports:

```text
DISCHARGE BOUNDARY INVALID: CASE RECLASSIFIED AS COSMOLOGICAL INITIAL DATA
```

That case can evolve declared post-initial data. It cannot validate the
classical `t = 0` singularity, creation of spacetime, pre-Planck evolution, a
unique inflation model, or another universe. A localized release on an FLRW
background is admitted only when its perturbative or nonlinear gravity model is
explicit and applicable.

Covariant local stress-energy conservation does not guarantee a single global
energy in an arbitrary dynamical spacetime. A cosmological or gravitational
Conservation of Ass pane names the exact current, symmetry, boundary, or
asymptotic charge supporting each ledger. Otherwise the global scalar ledger is
`not_applicable`.

### Analogue gravity and fiction

An acoustic horizon may preserve the propagation equation for selected
perturbations in a moving medium. It does not reproduce dynamical Einstein
gravity, black-hole matter, entropy, or quantum-gravity evidence. Its
certificate enumerates the preserved and unpreserved structures. A portal,
cross-universe exhaust, arbitrary-law Big Bang, or universe-spawning release
belongs in `fiction.axiomatic.*`, with declared rules and no borrowed empirical
status.

## Comparison signatures and optional dimensional analysis

Every law context may define a `ComparisonSignature` from the supported relations,
invariants, observables, capability states, and applicability results needed for
its own comparisons. A `PiSignature` is an optional extension when the profile
declares dimensional quantities and equations. No dimensional-analysis field is
required for a graph, rule, or constraint system that lacks those concepts.

For a profile with base dimensions and dimension-checked equations, a dimension
matrix `D` has dimensionless monomials with exponent vectors `k` such that:

```text
D * k = 0
```

The implementation computes exact rational null spaces for discovery, but a raw
null-space basis is not a stable public contract because the basis is not
unique. Law-pack authors declare stable semantic groups obtained from
nondimensionalizing their equations and closures. Tooling proves that those
groups are dimensionless, checks their dependencies, and reports possible
omissions. It does not invent scientific meaning from a null space.

An Earth continuum active signature includes:

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

Strict similarity is defined by the selected `ComparisonSignature`. For the
Earth dimensional-continuum specialization, the source and target normalized
problems must be identical within declared tolerance, including compatible
equations, closures, dimensions, topology, normalized geometry, material
functions, initial and boundary conditions, active coefficients, and solution
branches. Other law contexts declare different comparison relations or report
that strict similarity is not applicable.

Matching Mach and Reynolds alone does not make two compliant multiphase events
equivalent. Buckingham Pi analysis identifies possible dimensionless
combinations. It does not prove that the chosen variables are complete or that
the model is true.

## Universal Flatulence Translator

The translator separates four operations:

1. `semantic_translation`: requires a declared shared semantic basis.
2. `structural_mapping`: preserves selected relations or invariants without
   claiming shared meaning.
3. `signal_transcoding`: preserves a supported signal under channel constraints.
4. `experience_analogy`: creates an explicitly comic or artistic presentation.

Compatible semantic and structural operations may use `strict` matching or
`approximate` optimization. Strict matching requires compatible law-context
hashes and complete declared comparison signatures. Approximate optimization
selects target parameters for requested invariants subject to validity and
safety constraints, then reports every residual and conflict.

Exact monomial matching can become a constrained linear problem in log space.
Approximate translation minimizes weighted log-ratio errors or exposes a Pareto
frontier. The preference objective chooses among solutions; it is not physics.

Each translation certificate records source and target `LawContextSet` hashes,
scope assignments, bridge hashes, applicable mapped context-occurrence
identities or their explicit absence, mode, requested invariants, achieved
residuals, discarded quantities, incompatible fields, and whether the result is
validated, extrapolated, or fictional.

No-mapping results use stable locale-invariant reason codes:

- `law_incompatible`
- `no_shared_observable`
- `no_common_semantic_basis`
- `target_channel_insufficient`
- `not_identifiable`
- `outside_validity`
- `loss_exceeds_policy`
- `withheld_by_source`
- `forbidden_by_policy`
- `undecidable_within_budget`
- `outside_representable_ontology`
- `unknown`

The result records the source and target `LawContextSet`s, scope assignments,
applicable bridge rules, requested preservation, evidence, residual or loss,
policy basis, and whether more information could change the answer. A
refusal is never rewritten as physical incompatibility. `UNTRANSLATABLE` is a
localized presentation label, not the scientific result identity.

Strict translation also emits a machine-checkable `CompatibilityWitness` that
identifies every applicable mapped axiom, rule, equation, dimension, closure,
topology, condition, semantic group, and tolerance. Approximate translation emits the
attempted witness plus unmatched capabilities, residuals, sensitivity,
nonuniqueness, and identifiability warnings. A close observable match is not
automatically a unique or physically equivalent inverse solution.

Different spatial dimensions generally change densities, interface measures,
wave propagation, turbulence, and geometric spreading. A literal 2D universe,
an axisymmetric reduction of 3D, and a 2D visualization are three different
things. Cross-dimensional translation is rejected unless an explicit mapping
exists.

## Identity and provenance

The project keeps ten identities separate:

1. **Scenario identity:** normalized requested contract, law contexts, scope,
   declared measurement interactions, inputs, and any declared scenario seed
   before any requested operation is admitted.
2. **Record identity:** one committed nonce and unique Lab capture or
   computation, independent of source-law time.
3. **Context-occurrence identity claims:** optional context-defined identities
   plus an optional composite identity supplied only by an inter-law coupling.
4. **Case-result identity:** authoritative Lab claim account under the declared
   law contexts, measurement, implementation, numerical, and tolerance
   contracts. It does not imply a source-law occurrence.
5. **Trace identity:** the observations, relations, fields, samples, and provenance that
   were actually retained from a case.
6. **View identity:** knowledge, privacy, accessibility, and selection filters
   over retained claims. It never changes case-result identity.
7. **Narrative identity:** resolved world, situated perspectives, storylets,
   facts, and narrative streams.
8. **Presentation identity:** language, layout, camera, audio-device path,
   accessibility, and rendering choices.
9. **Play-session identity:** rules, initial identity, roles, ordered canonical
   action journal, checkpoints, branches, and produced artifacts.
10. **Archive-byte identity:** exact serialized container bytes.

A view or presentation change must not alter scenario or case-result
identity. A measurement interaction is part of the accepted scenario and may
change that identity. An
archive may be migrated or recompressed without pretending that byte identity
was preserved. Cross-platform tolerant equality is a comparison result, not a
fake hash. Raw floating-point identity is promised only at a determinism level
that actually controls the arithmetic.

The candidate change matrix in
[INTERFACES.md](INTERFACES.md#identity-contract) records which of these
identities may change under locale, presentation, measurement, law, model, and
execution changes. It becomes normative only with the identity RFC and its
property suite.

Replay presents a trace. Numerical reconstruction computes a new realization
from retained inputs. Re-enactment uses a fresh record nonce. A bitwise-identical
reconstruction is still a new operation and encounter, not the original record.
Its relation to a source occurrence is decided only by the applicable law
contexts.

## Earth discharge fidelity ladder

All levels share Earth-discharge vocabulary, hashes, regime policy, and
certificate shape. They do not pretend to share one universal equation set.

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
- Trace chemistry, reaction, condensation, secondary aerosol, exterior
  chemistry, observer detection, and hazard interpretation as independent
  capabilities rather than one composition switch.
- Reacting, ionized, plasma, MHD, stellar, relativistic, gravitational, and
  cosmological profiles, each with its own equations and verification suite.
- Published compactification, string, and brane reductions labeled as
  `research_toy` or `speculative_model` at their actual evidential scope.
- Acoustic-horizon and other structural analogies with explicit preservation
  and non-preservation witnesses.
- `fiction.axiomatic.*` profiles that are internally consistent under declared
  rules but never called empirically validated.

## Closure registry and fidelity escalation

Every empirical or reduced closure has a registry entry containing its equation
or algorithm, implementation revision, provenance, dimensional inputs, validity
envelope, calibration data, uncertainty model, conserved transfers, verification
cases, known failure modes, and escalation condition. A certificate lists every
entry used.

Escalation is based on model applicability, not spectacle or platform:

- A reduced aperture model escalates when wall inertia, contact, separation, or
  fluid-structure feedback controls the requested observable.
- A dilute parcel model escalates when loading, collisions, neighbor effects,
  vapor shielding, or unresolved sheet topology becomes important.
- A continuum model escalates or refuses when its Knudsen criterion fails.
- A low-Mach or incompressible model refuses choking, shocks, and quantitative
  propagation acoustics.
- An isolated-bubble model escalates for interaction, boundaries, strong
  compressibility, or nonspherical collapse.
- A passive trace model escalates when trace mass, reaction heat, phase change,
  or property effects exceed its declared coupling bound.
- A fixed-spacetime model refuses when source backreaction is material.
- A local discharge model reclassifies when the requested exterior is instead a
  cosmological initial-data surface.
- A compact-mode or string truncation escalates when discarded modes or
  interactions control the requested observable.

Higher fidelity is not automatically more truthful. It earns authority only for
observables that pass stronger verification and validation evidence.

## Conservation of Ass

For profiles that define balances or conserved currents, Conservation of Ass is
the localized presentation alias for a serious ledger obligation. A profile
without conservation laws uses its declared invariant or consistency policy
instead and reports the conservation fields as `not_applicable`.

In curved or cosmological spacetime, a locally conserved stress-energy tensor
does not by itself authorize a global energy total. Every reported charge names
the applicable symmetry, current, boundary, or asymptotic construction. If the
selected profile defines none, the corresponding global ledger remains
`not_applicable` rather than displaying a fabricated total.

Each applicable Earth discharge run declares its control volume and double-entry
transfer ledgers:

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

## Determinism, scenario seeds, and record nonces

The public record schema field `event_nonce` has type `RecordNonce`. The comic
wire name is retained, but its semantics are precise: it changes Lab record
identity and therefore the committed case-result identity. It affects
realization identity only when the selected operation defines realization. Pure
scenario validation neither creates nor reads a record nonce or reconstruction
seed unless a later explicit record contract supplies one.
`ContextOccurrenceIdentityClaims` change only according to identity concepts
actually defined by the law contexts and any declared inter-law coupling. They
are absent when none of those contexts defines one.

Determinism has explicit levels:

- D0: the same normalized scenario and, when declared, the same scenario seed
  have the same scenario identity.
- D1: the same build, target, settings, thread policy, and retained record nonce
  reconstruct the numerical trace bitwise.
- D2: supported platforms match declared observables within tolerances.
- D3: cross-platform bitwise identity, claimed only if arithmetic, reductions,
  transcendentals, and DSP actually guarantee it.

The engine uses fixed iteration order, deterministic reductions, exact unit
normalization where possible, and counter-based or independently keyed random
streams. Physics, parcel, physical-audio closure, score, radio, narrative, and
presentation streams are separate. Scheduling, terminal width, localization,
camera, station, and subscriber choices cannot alter scenario or
case-result identity. A default Lab encounter is not reconstructable
after its record nonce is destroyed. This says nothing about recurrence under
the source laws. Scientific, test, and benchmark modes retain or explicitly
provide the nonce under their recording policy.

## Uncertainty, sensitivity, and identifiability

Certificates keep aleatory variability, parameter uncertainty, model-form
uncertainty, discretization error, iterative error, roundoff, and presentation
randomness separate. Inter-fidelity disagreement is measured rather than hidden
inside one confidence score.

Sensitivity reports state which uncertain inputs control each observable and
where interactions matter. Inverse and translation commands report families of
solutions when retained claims do not identify a unique underlying realization,
parameter set, or relation. Extrapolation
beyond a closure's validation domain is visible at the value, chart, and
certificate levels. One seeded realization is never presented as an uncertainty
study.

## Certificate claims

Every claim independently reports `pass`, `fail`, `inconclusive`, or
`not-applicable`:

- Replayable under a declared determinism level.
- Internally consistent under the declared axioms, rules, dimensions, or other
  applicable structures.
- Code verified against exact, manufactured, or trusted reference cases.
- Solution verified by timestep, grid, parcel, or ensemble refinement.
- Empirically validated for a stated physical domain.
- Fictional-law consistent for a stated axiom pack.

The generic certificate includes record, law, scope, provenance, validity, and
implementation claims plus the independent model-claim class, empirical-support
scope, scenario applicability, and review revision. Occurrence, state,
relation, observable, invariant, balance, and observation sections appear only
when selected capabilities define them.
Unknown, unsupported, and unverified concepts remain explicit without numeric
placeholders. Law-specific extensions may add
equation, closure, solver, numerics, comparison signature, extrema, positivity,
ledgers, refinement, regime, random-stream, and consumer hashes. A play receipt
separately binds the knowledge policy, action journal, budgets, branches, result
artifacts, and score vector.

## Verification gates

The progressive suite includes:

- Ontology conformance fixtures for a case with no occurrence, realization,
  emitter, interface, exterior, gas, geometry, acoustics, mass, canonical units,
  or localized observer, with schema, archive, certificate, and inspection round
  trips.
- A non-biological non-gaseous transfer, a discrete graph or non-Euclidean
  profile, a distributed nonlinguistic observer, and a profile where familiar
  Earth quantities are explicitly `not_applicable`.
- Distinct tests for semantic translation, structural mapping, signal
  transcoding, experience analogy, and every structured no-mapping reason.
- Measurement-interaction tests for passive, distributed, probabilistic, and
  state-altering measurement, including provenance of back-action and changed
  case-result identity.
- View and presentation tests proving that knowledge, privacy, accessibility,
  locale, layout, sonification, and camera changes do not alter the authoritative
  Lab account.
- Cross-surface tests proving that model-claim class, empirical support,
  applicability, assurance, nonclaims, and evidence revision survive CLI, TUI,
  native, agent, archive, and export paths.
- Morphology-patch tests proving exact diff and archive round trips, identical
  physics under cosmetic or locale changes, and no hidden composition, health,
  identity, or social-value parameter changes.
- Unit-expression invariance and dimensional homogeneity where units and
  dimensions exist.
- Exact null-space checks for discovered bases and dimensional checks for every
  authored semantic Pi group where dimensional analysis applies.
- Preflight rejection of unsupported capability combinations and untrusted
  executable content.
- Zero-area and equal-pressure zero-flow limits.
- Rigid adiabatic and isothermal reservoir limits.
- Continuity and the mass-flow plateau at the analytical choking boundary.
- Positive mass, species, density, pressure, temperature, and internal energy
  without silent clamping.
- Bulk-mixture and passive-trace coupling bounds; species and elemental closure
  for each reaction mechanism; no-reaction and zero-source limits; phase and
  dew-point references; latent-energy balance; and distinct primary, secondary,
  suspended, deposited, reacted, and exported material accounts.
- Detection-model tests that keep physical concentration, compound-dependent
  response, adaptation, and uncertainty separate from hazard interpretation,
  including explicit refusal when duration, route, endpoint, population, or
  authority context is missing.
- Exact discrete cancellation for reservoir, plume, deposition, and recoil
  transfers.
- Correct starting-jet and puff limits.
- Canonical starting-vortex formation and source-closure stopping-vortex cases,
  with geometry-specific validity stated.
- No exterior acoustic propagation in vacuum while structural vibration remains.
- Particle tracer, drag-relaxation, settling, ballistic, evaporation, and
  deposition balances.
- Breakup and droplet-size convergence above a declared reliable numerical
  cutoff. A peak tied to roughly three grid cells fails this gate.
- Dilute and dense spray cases distinguished, with parcel, grid, and ensemble
  refinement appropriate to the active model.
- Underwater hydrostatic, Laplace-pressure, Rayleigh-Plesset, and bubble-resonance
  references when that pack exists, followed by an open bubble-acoustics dataset
  comparison before empirical-validation claims.
- Sealed-system center-of-mass conservation and externally vented rocket limits.
- Deterministic random streams independent of execution order.
- Traversable provenance from every consumer artifact and lab-claim sentence to
  authoritative Lab-account nodes.
- Go-to-Rust oracle parity.
- Strict similarity scale families and exact translator round trips.
- Explicit infeasibility for incompatible targets and dimensions.
- Exact compact-circle spectra, degeneracies, zero-mode limits, and controlled
  truncation; free-string level matching and the selected momentum-winding
  duality; and brane small-slope or linear limits when those packs exist.
- Minkowski limits, fixed-background geodesic and redshift references,
  gravitational constraint convergence, and explicit backreaction refusal.
- Radiation, matter, and vacuum FLRW analytical solutions, continuity and
  redshift checks, perturbation applicability, and cosmological initial-data
  reclassification.
- Analogy tests that enumerate preserved equations and reject unstated
  gravitational, quantum, ontological, or empirical conclusions.
- Manufactured-solution order, shock relations, and separate time and grid
  refinement.
- Sod shock tube, shock-vortex interaction, isentropic nozzle, finite-reservoir
  blowdown, rectangular-room modes, and nested acoustic-surface comparisons.
- Ensemble and integral comparisons after chaotic fields decorrelate.

Aggregate non-generated core statement coverage remains at least 90 percent,
and every non-generated package remains above 80 percent. Higher-risk domain,
solver, archive, protocol, and changed code use stricter branch, property, fuzz,
mutation, and differential evidence. Coverage supports verification but never
replaces validation.

The benchmark inventory, required artifacts, tolerances, and claim vocabulary
are defined in [VERIFICATION.md](VERIFICATION.md).

## Safety boundary

F.A.R.T. Lab is a game and educational simulator, not a medical, biological,
pressure-vessel, ignition, aerospace, or planetary-safety tool. Public challenges
must not encourage real pressure-vessel abuse, ignition, inhalation, harmful
biological experimentation, or contamination.

Odor is never used as a safety alarm, and occupational or emergency limits are
never presented as universal safe thresholds. The simulator does not provide
real mixture recipes, exposure dosing, disease inference, hazardous-device
control, accelerator design, or engineering instructions for destructive
high-energy releases. Black-hole, string, extra-dimensional, and Big Bang-scale
results retain their model class and nonclaims in every surface.

The research sources and their exact scope are listed in
[RESEARCH.md](RESEARCH.md).
