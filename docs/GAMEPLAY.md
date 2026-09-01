# Gameplay and procedural broadcast contract

F.A.R.T. Lab treats physics as the straight man and comedy as the interpretation
layer. The result should be the world's most overengineered and fun fart app,
whether the source is a person, machine, civilization, planet, star, or thing
that does not fit into our universe's categories.

## Four layers of every episode

1. **Native event:** what happened under the source world's declared laws.
2. **Lab normalization:** the emitter, interface, exterior, transported
   quantities, observers, and invariants that make the event inspectable.
3. **Cultural interpretation:** what local beings or systems think the event
   means.
4. **Player translation:** the text, audio, graphics, and measurements through
   which a human player can experience it.

A player-facing sound may be a translation of vibration, chemistry,
electromagnetism, gravity, or another sensory channel. The interface says when
it is translating. It never implies that every universe communicates through
air pressure and ears.

## Quick Play

`fart` in an interactive terminal and `fart quick` should produce a meaningful
result without a setup questionnaire.

A Quick Play transmission:

1. Prints its title and replay seed immediately.
2. Introduces one world, one source, one cultural fact, and one question.
3. Runs one valid analytical event.
4. Renders the event and one consequence.
5. Explains one physical reason for the outcome.
6. Offers exactly three next actions: inspect, replay, or generate another.

The default mix favors approximately 60 percent ordinary or low-energy events,
25 percent unusual environments, 12 percent extreme sources, and 3 percent
cosmic events. This distribution is a tuning target, not a hidden physics input.
Graphic payload and extinction content remain opt-in.

Low magnitude never means low importance. A quiet pfft can complete a ritual,
make first contact, resolve a court case, expose an exquisite resonance, or fail
at precisely the funniest moment.

## Broadcast mode

`fart broadcast --seed <seed>` is the seeded interdimensional show. It generates
a law profile, world, source, morphology, senses, culture, situation, physical
scenario, and stakes question. It runs the simulation, freezes its identity,
and assembles a watchable episode from immutable facts.

A standard episode uses these beats:

1. Cold open.
2. World and sensory translation card.
3. Source introduction.
4. Cultural stakes.
5. Instrument setup and prediction.
6. Emission.
7. Honest readout.
8. Reactions and consequences.
9. Callback.
10. Certificate sting and replay code.

The story director may:

- Select a compatible template before simulation.
- Generate and validate initial conditions.
- Select cameras, instruments, narration, pacing, and display density.
- Pause or visibly compress playback time.
- Choose reaction storylets whose predicates match event facts.
- Let characters hold incorrect beliefs when dialogue marks them as beliefs.

The story director may not:

- Write simulation state after a scenario is accepted.
- Change law, pressure, payload, timestep, fidelity, or seed to rescue a story.
- Rerun a valid event because the result was insufficiently dramatic.
- Invent sound in vacuum, deposition without transported matter, or damage
  without an applicable transfer and target model.
- Hide failed or inconclusive verification.
- Present character interpretation as a measured fact.

If a predicted planet killer produces a polite ripple, the polite ripple is the
ending. Narration, witnesses, bureaucracy, and the callback make failure funny.
The simulation does not get another take.

## Freestyle Lab

Freestyle exposes the complete instrument:

- Start blank, from a physical preset, or from an archived episode.
- Edit values with explicit units and validate before running.
- Inspect the exact scenario diff produced by every convenience control.
- Branch an archive while preserving ancestry.
- Run sweeps, comparisons, translations, optimization, and refinement.
- Export tables, procedural audio, transcripts, plots, and certificates.
- Turn an experiment into a personal challenge.

Scientific commands are available from the beginning. Campaign progression
guides discovery and unlocks curated content, but does not hide the laboratory
behind playtime.

## Challenges

A challenge contract declares:

- Target observables.
- Physical, structural, material, time, and resource constraints.
- Allowed laws, models, fidelity, and source packs.
- Scoring vector and tie-breaking policy.
- Verification tolerances.
- A reference solution or proof of feasibility.
- Accessibility options and which scoring axes they affect.

Scores are vectors rather than explosive yield alone:

- Precision.
- Resource efficiency.
- Verification quality.
- Physical novelty.
- Translation fidelity.
- Cultural consequence.
- Cleanliness or deposition where relevant.

Examples include a perfect C-sharp without a wet transition, the lowest-energy
ceremonial translation between two worlds, a bubble symphony at depth, a
spacecraft attitude correction with honest impulse, and a refinement study that
shows `KABLAM` is a stable classification rather than a numerical artifact.

## Campaign progression

1. **Field Orientation:** observe, replay, inspect, and explain one ordinary
   pfft.
2. **Bathroom Science:** pressure, duration, direction, interface motion,
   composition, room response, and wetness.
3. **The Wind Tunnel:** starting jets, puffs, vortices, droplets, particles, and
   conditional similarity.
4. **Myth Lab:** candles, resonance myths, vacuum, choking, and careful verdicts.
5. **Orbital Flatulence:** recoil, torque, attitude control, and honest delta-v.
6. **Extreme Exteriors:** underwater, rarefied, planetary, and strong-gravity
   environments, one validated model pack at a time.
7. **Universal Translation Office:** source classes, observer senses, cultures,
   scales, and compatible law profiles.
8. **Speculative Laws:** explicitly fictional constants, dimensions, topology,
   fields, and conservation rules.
9. **Apocalypse Mode:** optional laboratory, stellar, relativistic, and
   planet-scale consequence packs with separate source models.

Rewards are field notes, proof stamps, curated profiles, instruments, broadcast
formats, translation packs, and archive exhibits. Avoid grind, daily chores,
fear of missing out, and a simple bigger-number ladder.

## Deterministic seed architecture

Each episode has one 256-bit master seed. Named, versioned substreams are derived
from stable data:

```text
episode format version
content pack hash
master seed
namespace
stable entity id
counter
```

Initial namespaces include:

```text
cosmos.laws
cosmos.topology
world.environment
culture.norms
culture.language
entity.morphology
entity.senses
entity.persona
source.parameters
event.payload
physics.turbulence
physics.parcels
physics.numerics
director.selection
director.pacing
language.realization
presentation.audio
presentation.visual
presentation.camera
```

Use a counter-based or equivalently splittable generator. Do not use one mutable
global random stream. Localization, terminal dimensions, subtitles, camera,
frame rate, and content filters cannot alter the physics event hash. Player
choices branch by stable beat and choice identifiers, never wall-clock input
timing.

The archive stores resolved content as well as seeds. An update to a grammar or
content pack creates a new episode identity and does not rewrite an old show.

## Content pipeline

```text
law profile
  -> habitat
  -> source capabilities
  -> observer senses
  -> culture
  -> situation and stakes
  -> valid physical scenario
  -> simulation
  -> immutable event facts
  -> reaction storylets
  -> player rendering
```

Generated cultures are connected systems, not random costumes and syllables.
Habitat, energy economics, coordination, sensory channels, and attitudes toward
venting constrain each other.

A storylet declares its identifier, version, episode phase, state and fact
predicates, knowledge requirements, exclusions, priority, cooldown, duration,
claims, opinions, state effects, localization keys, and content tags. Every lab
claim resolves to a scenario or telemetry path. Opinions are explicitly
diegetic.

Start with authored storylets and deterministic grammars. Runtime unrestricted
text generation would weaken replay, safety review, offline use, localization,
fact provenance, and archive longevity.

## Episode archives

A shareable episode bundle contains:

- Master seed and named-stream manifest.
- Resolved law, world, source, culture, names, and grammar choices.
- Complete scenario and authoritative event archive.
- Physics, narrative, and content-pack versions and hashes.
- Beat timeline with fact provenance.
- Certificate and event hash.
- Plain-text transcript.
- Optional procedural audio and native presentation assets.
- Parent episode and stable branch choice for remixes.

Archives exclude usernames, home paths, machine identifiers, and telemetry by
default. Importers treat archives and content packs as untrusted. They reject
path traversal, duplicate members, links, decompression bombs, oversized arrays,
schema violations, and hash substitution before consuming payloads.

## Tone

The comic formula is:

1. Absolute institutional sincerity.
2. A precise physical fact.
3. A disproportionate but internally coherent cultural consequence.

The joke targets pomposity, bureaucracy, overconfidence, and the refusal to stop
measuring. It does not target a real culture, identity, disability, body type, or
anatomy. Not every line is a punchline. Silence, anticipation, and a clinical
readout provide rhythm.

## Accessibility and content controls

Independent controls cover science density, narration verbosity, comedy density,
grossness, deposition detail, consequence intensity, audio, dynamic range,
motion, camera shake, flashing, haptics, playback speed, text, contrast, color,
subtitles, and screen narration.

Critical information remains available with color, sound, animation, haptics,
and Unicode disabled. The CLI supports plain text and redirected output. The TUI
has reduced-color, ASCII, keyboard-only, and append-only modes. Native
accessibility setup appears before the first audiovisual event.

## Acceptance gates

- At least 10,000 deterministic generated seeds terminate within a fixed step
  budget and pass schema and solver preflight.
- Every lab-fact sentence cites a scenario or telemetry field.
- Mutation tests catch sound in vacuum, impossible deposition, hidden rerolls,
  premature outcome knowledge, and presentation changes that alter physics.
- Ordinary low-energy events are at least half of default Quick Play and receive
  full setup, result, consequence, explanation, and callback treatment.
- Physics random streams are independent of narrative, presentation, execution
  order, terminal dimensions, localization, and accessibility settings.
- Archive round trips preserve event, story timeline, stream manifest,
  certificate, and hashes.
- Malicious archive tests fail safely.
- Every TUI or native action exposes an equivalent CLI command, scenario change,
  or archive operation.
- Disabled-player testing occurs before public native release.

The research basis is listed in [RESEARCH.md](RESEARCH.md).
