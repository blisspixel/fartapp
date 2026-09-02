# The Reference Pfft and metrology contract

F.A.R.T. Lab begins with something intentionally unexceptional. The bare
`fart` command does not open a pressure vessel, select a disaster, or roll a
cosmic source. It stages one low-energy dry event from a synthetic,
biology-neutral calibration fixture under a published Earth-profile procedure.

The event is called the **Reference Pfft**, profile identifier
`ref:rp1:v1`. The name is an English working title; the identifier is opaque.
The measurement language is not comic. This profile is a project convention
used to compare models and interfaces. It is not an SI unit, a medical norm, a
claim about an average being, or a national or international measurement
standard.

The fixture contains no human, animal, anatomy, sex, gender, diet, nationality,
health state, bathroom, or species. It is a deformable finite emitter, compliant
interface, payload, exterior, and observer defined only as far as the equations
require. Biological and cultural scenarios are separate, optional realizations.

## What is standardized

The profile standardizes the conditions and procedure, not one supposedly
universal emission. Its initial conventional exterior is:

| Quantity | Conventional value | Status |
| --- | ---: | --- |
| Ambient pressure | `101 325 Pa` | Project reference value matching the standard atmosphere |
| Ambient temperature | `293.15 K` | Project reference value, equal to 20 degrees Celsius |
| Relative humidity | `50 %` | Project reference target |
| Gravitational acceleration | `9.806 65 m/s^2` | Conventional standard gravity |
| Spatial model | Three-dimensional Euclidean Earth continuum | Declared law-profile assumption |
| Exterior medium | Humid air mixture | Composition and property model must be versioned |
| Enclosure | Small reference geometry with declared boundaries | Project geometry, not a universal habitat or room function |
| Observer | Named position, orientation, and instrument response | Required for acoustic and sensory claims |

These values define an abstract scenario exactly. A physical realization in a
real room would measure pressure, temperature, humidity, geometry, and sensor
response with stated uncertainty. The software must never display the abstract
convention as if it were a fresh instrument reading.

Source parameters are deliberately not frozen as biological truth. They begin
as a conservative, low-energy, dry, finite-duration model inside cited empirical
bounds. Each parameter records whether it is measured, literature-derived,
elicited, fitted, conventional, or fictional. The release cannot call that
source “typical” until the relevant population, protocol, uncertainty, and
evidence support the word.

Published human evidence is sparse and heterogeneous. Daily collected gas
volume cannot be divided into one event without another model. Composition and
manometry samples do not provide the synchronized pressure, aperture, mass-flow,
and audio waveform needed for calibration. Initial biological priors therefore
use aggregate public evidence and broad uncertainty. No new human or animal
measurement program begins without independent ethics review and a justified
scientific need.

“Sea level” is shorthand in casual copy only. Scientific output says
`101 325 Pa reference ambient`. Actual sea-level pressure varies with weather,
elevation definitions, and location.

## Exact constant foundation

The candidate standard does not define its own second, metre, kilogram, kelvin,
or mole. It inherits the SI and makes that inheritance inspectable. Relevant
exact defining constants include:

| Constant | Exact SI value | Role in the candidate standard |
| --- | ---: | --- |
| Caesium-133 hyperfine frequency, `delta nu Cs` | `9 192 631 770 Hz` | Time coordinate |
| Speed of light in vacuum, `c` | `299 792 458 m/s` | Length coordinate |
| Planck constant, `h` | `6.626 070 15e-34 J s` | Mass and energy coordinates |
| Boltzmann constant, `k` | `1.380 649e-23 J/K` | Thermodynamic-temperature coordinate |
| Avogadro constant, `N_A` | `6.022 140 76e23 1/mol` | Amount-of-substance coordinate |

From them, the Lab can express every Earth-profile result in exact
dimensionless coordinates before applying measurement and model uncertainty:

```text
time coordinate        = t * delta_nu_Cs
length coordinate      = x * delta_nu_Cs / c
mass coordinate        = m * c^2 / (h * delta_nu_Cs)
energy coordinate      = E / (h * delta_nu_Cs)
temperature coordinate = k * T / (h * delta_nu_Cs)
amount coordinate      = n * N_A
pressure coordinate    = p * c^3 / (h * delta_nu_Cs^4)
```

These are not proposed as convenient display units. They are the exact chain
showing how displayed seconds, metres, kilograms, joules, kelvins, moles, and
pascals inherit the SI definitions. Human output continues to use practical SI
scales.

The distinction is essential: constants define the coordinate system. They do
not select a biologically meaningful pressure history, composition, aperture,
or duration. Deriving a “natural fart” by combining `c`, `h`, and `k` until a
funny-sized number appears would be dimensional numerology.

## Candidate definition and one-time realizations

The standards thought experiment therefore has two layers:

1. **`RP-1 definition event`:** an exact, deterministic analytical calibration
   fixture with a published law profile, conventional exterior, normalized
   geometry, finite source state, observer, equations, numerical policy, and
   dimensionless reference signature. It can be reconstructed to test software
   and instruments.
2. **`RP-1 everyday realization`:** a one-time event sampled from a published
   low-energy uncertainty model around the definition. It receives a fresh
   record nonce, develops its own imperfections, and is never called the same
   Lab record. The Earth profile separately defines its context-occurrence
   identity.

The definition event is the tuning fork. The everyday realization is the note
that occurs in one declared context, once.

The candidate dimensionless signature is not just a tuple of Mach, Reynolds,
and Strouhal numbers. It includes normalized source and interface histories,
geometry, boundary and initial conditions, composition and material functions,
closure identifiers, active dimensionless groups, observer transfer functions,
and tolerance rules. A short famous-number match is insufficient.

The exact conventional parameter values for the `RP-1 definition event` remain
unratified until the analytical dry-flow model, empirical bounds, sensitivity
analysis, and independent review exist. The project will not invent precision
in documentation before the evidence can choose a defensible candidate.

## RP-1 Earth-profile comparison vector

A decibel cannot describe deposited mass. Emitted mass cannot describe pitch.
Peak pressure cannot describe duration or impulse. F.A.R.T. Lab therefore does
not collapse an event into one proprietary magnitude.

There is no single fart unit. For RP-1, the minimum comparison vector contains:

- Emitted mass by phase and species.
- Total impulse and angular impulse.
- Duration under a declared threshold rule.
- Peak and integrated volume flow.
- Reservoir and interface work.
- Acoustic exposure at a named observer and bandwidth.
- Directional plume extent under a declared concentration threshold.
- Droplet and particle loading, transport, and deposition.
- Active dimensionless signature and regime history.
- Mass, momentum, and energy residuals.

Compact ratings may summarize selected components for play, but the underlying
vector stays visible and the rating policy is versioned. “Three pffts” is never
accepted as a unit expression.

## Traceability chain

Metrological traceability relates a result to a reference through a documented,
unbroken calibration chain, with every link contributing uncertainty. The Lab
applies that principle in layers:

```text
SI definitions and stated conventions
  -> calibrated environmental and geometry inputs
  -> versioned material-property and sensor models
  -> normalized scenario with units and uncertainty
  -> solver and numerical realization
  -> observer model and derived measurands
  -> certificate, archive, and comparison result
```

Every reported quantity identifies:

- The measurand and system boundary.
- Value, unit, coordinate frame, time basis, and sign convention.
- Source and transformation provenance.
- Calibration or project convention used.
- Measurement, parameter, model-form, discretization, iterative, and roundoff
  uncertainty where applicable.
- Validity domain and unsupported effects.

Calibration, adjustment, verification, and validation are separate operations.
Changing a sensor correction is not calibration. Passing a conservation check
is not empirical validation. Reproducing the same number is not evidence that
the number represents nature.

Each certificate separates exact definitions, project conventions, calibrated
measurements, literature estimates, fitted parameters, and stochastic latent
state. A displayed decimal never silently acquires the authority of the SI
constant chain above it.

## The Reference Pfft procedure

1. Resolve `ref:rp1:v1` and print its profile hash.
2. State every conventional and empirical input before the event.
3. Commit to the scenario seed and unique record nonce.
4. Run the lowest validated dry-flow fidelity that supports the scenario.
5. Sample the named observer without inventing unresolved measurements.
6. Close the mass, momentum, and energy ledgers.
7. Report uncertainty, validity, and any inconclusive claims.
8. Let the encounter pass unless recording was selected before emission.

The initial commands are planned as:

```console
fart
fart reference inspect
fart reference realize --environment measured-enclosure.toml
fart reference compare run.fart
fart reference uncertainty run.fart
fart reference trace run.fart --quantity observer.acoustic_exposure
```

The bare command favors immediacy and impermanence. It does not persist a full
archive unless the player selects recording. `fart event run scenario.toml
--output run.fart` remains the explicit laboratory path for durable work.

## Reference hierarchy

The project uses restrained names so comedy does not become a false credential:

- **Definition:** the schema and mathematical definition of a measurand.
- **Reference profile:** a versioned project procedure and conventional inputs.
- **Reference realization:** a result produced by a named solver build under
  the profile.
- **Working fixture:** a compact result used routinely in tests and adapter
  comparisons.
- **Transfer artifact:** an archive used to compare implementations or systems.

“Certified” always names the exact certificate claims. It never means approved
by BIPM, NIST, another metrology institute, a regulator, or a standards body.

## Quality gates

- Unit spellings and display systems cannot change normalized scenario identity.
- Every conventional value is visibly distinguished from a measurement.
- Every measured input carries uncertainty and calibration provenance.
- Observer-dependent quantities cannot appear without an observer definition.
- Repeated realizations compare distributions and uncertainty, not only means.
- The Reference Pfft cannot cross a laboratory-pressure or wetness boundary
  without a source-model change visible in the scenario diff.
- Reference updates create new profile versions and permanent migration notes.
- Documentation examples are executable fixtures once their command exists.

The research basis and the exact scope of each source are in
[RESEARCH.md](RESEARCH.md).
