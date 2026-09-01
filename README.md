# F.A.R.T. Lab

**Flatulence Aerodynamics Research & Testing**

*From pfft to planetary extinction.*

F.A.R.T. Lab is the world's most overengineered and fun fart app. It is a vulgar
comedy game, a serious simulation laboratory, and a Universal Flatulence
Translator for connecting beings, worlds, dimensions, and universes through the
one joke every civilization eventually discovers.

The familiar Earth profile starts with a pressure vessel, a compliant opening,
and an exterior. From one physical event it derives the plume, sound, vibration,
particles, recoil, room response, damage, and controller rumble. Other profiles
may describe a machine, colony, planet, star, distributed intelligence, or
fictional higher-dimensional source under explicitly different laws.

There is no menu of prerecorded farts. If the pressure history changes, every
observable changes with it.

## Project status

The repository currently contains the original Go CLI prototype. It accepts an
intensity from 1 to 5 and prints a deterministic, rated emission. The next goal
is not a graphical wrapper. It is a genuinely excellent, cross-platform command
line laboratory.

The wider product will be built in this order:

1. **CLI Lab:** every model, parameter, run, sweep, comparison, replay, and proof
   is available headlessly on Windows, macOS, and Linux.
2. **Terminal Lab:** an htop-style live instrument panel presents the same CLI
   capabilities without hiding or duplicating them.
3. **Native Lab:** a polished native desktop game adds spatial audio, fluid
   visualization, haptics, rooms, and destruction through the same physics core.

The desktop app will not be a browser shell, embedded webview, or local web
server. It will be a native Godot application backed by the production physics
core. Every future physics feature follows the same promotion path: CLI first,
terminal instrumentation second, native presentation third.

See the [roadmap](ROADMAP.md) for delivery gates.

## The scientific premise

The neutral event contract is:

> A finite-time state transition coupling an emitter domain to an exterior
> domain across an interface, under a versioned law profile.

The first validated specialization is delightfully anatomical:

> A pressure-driven discharge from a deformable reservoir through a compliant
> aperture into an exterior domain.

This supports real thermodynamics, compressible flow, turbulence,
aeroacoustics, multiphase transport, rigid-body recoil, and eventually extreme
gas dynamics. Earth continuum profiles use SI as their canonical unit system.
Other law profiles must declare their dimensions, constants, equations,
conserved currents, units, and validity limits rather than inheriting Earth
physics by accident.

| Domain | Representative inputs |
| --- | --- |
| Law profile | Dimensions, metric, constants, fields, equations, closures, and invariants |
| Emitter | Inventory, pressure or driving potential, temperature, volume, composition, compliance |
| Interface | Geometry, topology, elasticity, opening history, orientation, surface condition |
| Exterior | Pressure, gravity, humidity, temperature, medium or vacuum, boundaries, world geometry |
| Payload | Gas, liquid-droplet, and solid-particle mass fractions and size distributions |
| Observers | Sensor locations, bandwidth, sensory model, and translation rules |
| Numerics | Fidelity level, timestep policy, grid resolution, and random seed |

The simulator derives an active dimensionless signature from the selected
equations. An Earth gas event may activate pressure ratio, Mach, Reynolds,
Strouhal, Froude, Knudsen, Weber, Stokes, and particle-loading groups. Mach is
not invented where no sound speed exists, and Weber is not reported where there
is no material interface. The labels and jokes are consequences of calculated
state and a versioned classification policy, not hidden presets.

## One event, many consequences

```text
event parameters
      |
      v
reservoir -> compliant aperture -> jet and plume -> room or atmosphere
      |                |                 |
      +----------------+-----------------+
                       |
                       v
             shared event history
                       |
          +------------+------------+------------+
          |            |            |            |
        audio       visuals      haptics       damage
```

The same mass-flow and pressure history drives every branch. Interface layers
may render, sonify, and explain the result, but they must not invent a second
physics story.

## Regime classification

The classifications are independent axes, not one fake severity ladder. An
event can be wet and choked at the same time.

### Material regime

| Label | Model meaning |
| --- | --- |
| Dry fart | A predominantly single-phase turbulent gas jet |
| Wet fart | Gas carrying suspended droplets with breakup, evaporation, or deposition |
| Shart | Dense liquid or particle loading with meaningful surface deposition |
| Solid fart | `FLATULENCE BOUNDARY CROSSED: EVENT RECLASSIFIED AS FECAL EJECTA` |

### Flow regime

| Label | Model meaning |
| --- | --- |
| Subsonic | Aperture flow remains below Mach 1 |
| Choked | The critical pressure ratio produces sonic flow at the narrowest opening |
| Underexpanded or supersonic | A choked jet expands outside the aperture and may form shock cells |

### Energy and source regime

| Label | Model meaning |
| --- | --- |
| Thermal gas | Ordinary gas chemistry remains applicable |
| Plasma | Ionization is significant enough that the gas model must change |
| Solar | The biological source model is disabled and replaced by a stellar-energy release through an anatomically themed nozzle |

## The Flatulence Similarity Law

Two events are strictly similar only when their nondimensional governing
equations, closures, spatial dimension, normalized geometry, boundary and
initial conditions, material functions, and active dimensionless coefficients
match. Matching a short list of famous numbers is necessary in some models, but
not sufficient in general.

That conditional similarity protocol is Buckingham Pi analysis used as a game
mechanic and a testable engine invariant. It enables a challenge such as
reproducing an event at 1,000 times scale without merely multiplying every value
by 1,000.

It also powers the Universal Flatulence Translator:

1. **Strict translation** matches the complete active signature between
   compatible law profiles.
2. **Approximate translation** solves for selected invariants, reports residuals,
   and admits when no target solution exists.
3. **Comic translation** preserves declared observer experiences such as pitch,
   surprise, visible extent, social meaning, or perceived severity, and labels
   the result as presentation rather than physical similarity.

A vacuum cannot preserve external loudness. A universe without surface tension
cannot preserve Weber number. A different spatial dimension generally changes
wave spreading and jet behavior. `UNTRANSLATABLE` is a successful scientific
answer.

The lab gives memorable names to real constraints:

- **The Choked Cheek Criterion:** the aperture reaches sonic mass flow at the
  critical pressure ratio.
- **The Wetness Transition:** droplets break up, remain entrained, or deposit as
  Weber and Stokes behavior changes.
- **The No-Sound-in-Vacuum Lemma:** recoil and transported matter can remain, but
  an external acoustic wave cannot propagate without a medium. Visibility still
  requires particles, condensate, radiation, plasma, or another declared model.
- **Conservation of Ass:** expelled mass, momentum, and energy must be accounted
  for in the reservoir, plume, payload, boundaries, and environment.

## Ways to play

- **Quick Play:** one command creates one coherent, physically valid
  transmission with a visible seed, one consequence, and useful replay and
  inspection commands.
- **Broadcast:** a seeded universe, law profile, source, culture, situation, and
  event unfold like interdimensional television. The story director reacts to
  immutable simulation facts and never rerolls a polite result into a disaster.
- **Freestyle Lab:** edit every supported law, emitter, interface, payload,
  exterior, observer, and numerical setting with explicit units and proof.
- **Challenges and campaign:** solve constrained scientific puzzles and progress
  from ordinary pffts to universal translation and optional apocalypse physics.

The master episode seed derives named, versioned random streams for laws, worlds,
entities, physics, narration, audio, and presentation. Changing terminal width,
language, camera, or a joke line cannot change a particle trajectory or event
hash. A shareable episode archive preserves resolved story canon as well as the
certified event.

See [docs/GAMEPLAY.md](docs/GAMEPLAY.md) for the mode and story-director
contracts.

## CLI Lab

The command line is the primary scientific product, not a debug console for the
eventual app. The planned command surface includes:

```console
fart quick --seed F7-4PK9 --record run.fart
fart broadcast --seed 42 --length standard
fart freestyle bathroom.toml --set reservoir.pressure="106 kPa"
fart simulate bathroom.toml --output run.fart
fart inspect run.fart
fart sweep reservoir.pressure 105kPa..800kPa --steps 64
fart compare small.fart large.fart --nondimensional
fart verify run.fart --refine timestep
fart replay run.fart --check-hash
fart export run.fart --format json,csv,wav
```

Names and flags remain provisional until their schemas are implemented. The
contract is not provisional: commands must compose in scripts, support stable
machine-readable output, explain invalid physical states, and never require a
GUI. Human output should be useful at a terminal, while JSON, CSV, and event
archives make every calculation inspectable.

## Terminal Lab

After the CLI is complete and stable, the same core becomes an htop-style
terminal application. It will provide live panes for:

- Reservoir pressure, temperature, volume, mass, and energy.
- Aperture area, compliance, mass flow, Mach number, and structural margin.
- Plume and payload summaries, including an optional character-cell field view.
- Waveform, spectrum, pitch, and room-acoustic diagnostics.
- Dimensionless groups, regime transitions, and challenge targets.
- Mass, momentum, and energy ledgers with numerical residuals.

It will support modern terminals on Windows, macOS, and Linux, detect terminal
capabilities, and provide reduced-color and ASCII fallbacks. Every action must
have an equivalent CLI command or event-file edit so the TUI never becomes a
second, untestable control plane.

## Native Lab

Only after the CLI and Terminal Lab are excellent does the full native app
begin. Its first playable target is one tiled bathroom and one continuous
discharge control that sweeps a curated, physically consistent path from pfft
through turbulent, wet, choked, and underexpanded states. It is a linked
scenario path, not a claim that pressure alone changes moisture content.

During the sweep, the plume, procedural sound, room reflections, tile response,
particles, deposition, damage, and haptics respond continuously to the same
event history already inspectable in the CLI. The app adds a world and a feel.
It does not add secret physics.

## Player progression

1. **Bathroom Science:** shape loudness, pitch, duration, plume direction, and
   deposition in a tiled room.
2. **The Wind Tunnel:** discover vortex structures, turbulence, resonance,
   droplet breakup, and similarity scaling.
3. **Myth Lab:** test candle extinction, the brown note, vacuum propulsion, and
   supersonic discharge under controlled conditions.
4. **Orbital Flatulence:** use conservation of momentum for spacecraft attitude
   control and small but honest delta-v maneuvers.
5. **Extreme Gas Dynamics:** move underwater, into vacuum, onto Venus and
   Jupiter, and into extreme gravity.
6. **Universal Translation Office:** translate between different sources,
   observer senses, cultures, environments, scales, and compatible worlds.
7. **Speculative Laws:** enter explicitly fictional dimensions, constants,
   fields, and topologies with axiom-conformance certificates.
8. **Apocalypse Mode:** unlock non-biological energy sources, plasma models,
   relativistic warnings, and planet-scale consequences.

Challenge ideas include producing a perfect C-sharp without crossing the
wetness boundary, translating a ceremonial pfft between worlds, creating an
underwater bubble symphony, and proving through refinement that a `KABLAM`
classification is robust. An ordinary Earth-biological profile cannot become
choked just because a slider moved. Choking at sea-level ambient pressure for a
perfect gas with gamma 1.4 needs roughly 90 kPa gauge upstream pressure, far
above measured ordinary biological pressure excursions. Laboratory and extreme
source packs must say where that energy came from.

## Architecture direction

| Layer | Responsibility |
| --- | --- |
| Go oracle | Tiny, auditable reference equations, event fixtures, and diagnostics |
| Rust physics core and CLI | Deterministic production simulation, headless play, proof, archives, translation, and replay |
| Rust terminal UI | Cross-platform live instrumentation over the same commands and event state |
| Native Godot client | Native input, visualization, procedural audio, rooms, haptics, and game progression |

The Go implementation remains intentionally small. Once an analytical model and
its fixtures are trustworthy there, the Rust production core must match them
within documented tolerances before it can power the CLI, TUI, or native app.

The project will use three fidelity levels behind the same event contract:

1. An analytical real-time model for immediate simulation.
2. An interactive 2D or low-resolution fluid model for the wind tunnel.
3. A slower high-fidelity benchmark mode for convergence studies and reference
   results.

For the detailed scientific contract, see
[docs/SIMULATION.md](docs/SIMULATION.md). For the interface and release rules,
see [docs/INTERFACES.md](docs/INTERFACES.md). The research basis and model
boundaries are collected in [docs/RESEARCH.md](docs/RESEARCH.md).

## Proof, not vibes

Every completed simulation should be able to emit a certificate containing:

- Versioned inputs, solver version, fidelity, timestep or grid settings, and seed.
- Mass, momentum, and energy balance residuals for a declared system boundary.
- Convergence evidence when smaller timesteps or finer grids are requested.
- Pressure, density, temperature, and species-amount validity checks.
- Regime transitions and the dimensionless values that caused them.
- Evidence that audio, visuals, haptics, and damage used the same event history.
- A deterministic replay identifier and result hash.

Certificate claims are independent: replayable, internally consistent, code
verified, solution verified, empirically validated for a stated domain, or
fictional-law consistent. Conservation and convergence do not prove that the
chosen physical model is true. The certificate reports assumptions, tolerances,
unsupported effects, and inconclusive claims. It is evidence about the
simulation, not a claim that a game model is a medical or aerospace tool.

## Current CLI

### Build

```sh
go build -o fartapp .
```

### Run

```sh
./fartapp <intensity>
```

Intensity must be an integer from 1 to 5.

```console
$ ./fartapp 3
braaap (respectable)
```

### Test

```sh
go test ./...
```

| Intensity | Output |
| --- | --- |
| 1 | `pfft (gentle)` |
| 2 | `toot (respectable)` |
| 3 | `braaap (respectable)` |
| 4 | `blorp (respectable)` |
| 5 | `KABLAM (mighty)` |

The idea is deliberately silly. The implementation should be flawless.

## Contributing, security, and license

See [CONTRIBUTING.md](CONTRIBUTING.md),
[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md), and [SECURITY.md](SECURITY.md).

No open-source license has been granted yet. Public visibility allows reading
and forking through GitHub, but does not by itself grant broader reuse rights. A
deliberate license decision is tracked in the roadmap.
