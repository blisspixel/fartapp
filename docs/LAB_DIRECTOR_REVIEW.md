# Fictional laboratory director review

This is a synthetic design exercise. The participants are fictional composites,
not real people, interviews, endorsements, affiliations, or credentials. Their
titles name the expertise being simulated so that the project can expose weak
claims before implementation.

The premise under review is deliberately absurd: treat every supported fart
analogue as an individual scientific case and engineer the instrument with more
care than the joke appears to deserve. The comedy comes from restraint,
specificity, and escalating rigor. It does not excuse fabricated science.

## Panel

- **Chair, experimental fluid dynamics:** source characterization, jets,
  diagnostics, uncertainty, and validation.
- **Director, scientific computing:** numerical methods, verification,
  reproducibility, and high-performance computing.
- **Principal chemist:** mixtures, trace species, phase change, reactions,
  detection, and hazards.
- **Professor, mathematical physics:** dimensional analysis, quantum models,
  gravity, extra dimensions, and limits of analogy.
- **Director, human-computer interaction:** CLI, terminal instruments,
  accessibility, learning, and native interaction.
- **Editor, play and culture:** comedy, localization, children, community, and
  misuse.
- **Staff engineer, dependable systems:** architecture, parsers, release
  engineering, security, and long-term maintenance.

## Review transcript

**Chair:** State the scientific object without performing the joke for us.

**Project:** A Lab case is an operation under declared law contexts, inputs,
measurement interactions, and applicability limits. In the familiar continuum
specialization, an emission is a pressure-driven discharge from a reservoir
through an interface into an exterior. Other contexts may omit any of those
concepts. “Fart” is the comic classifier, not a required ontology.

**Chair:** Good. What does the current reservoir command actually solve?

**Project:** An exact endpoint for finite withdrawal from a rigid, homogeneous,
nonreacting, calorically perfect ideal-gas mixture. Component fractions remain
fixed. The user selects adiabatic withdrawal or a prescribed isothermal path.
The command reports initial and final states, transferred mass and enthalpy,
heat transfer where applicable, and independently calculated mass, energy, and
reported-state equation-of-state consistency residuals.

**Chair:** Then do not call it blowdown yet. It has no aperture, time history,
external pressure, choking, heat-transfer law, mixing transient, stratification,
or wall model. An exact endpoint is useful, but its boundary must be more
prominent than its precision.

**Project:** Accepted. The command and documentation say endpoint prediction,
and the complete `RES-002` blowdown benchmark remains open.

**Scientific computing:** Are the balance checks circular?

**Project:** The implementation deliberately uses independent expressions where
practical. The adiabatic final temperature follows the state relation, while
withdrawn enthalpy is integrated separately. The isothermal heat input follows
the component mass and gas-constant ledger. Tests use analytical fixtures,
boundary properties, forged-state rejection, shuffled execution, fuzzing, and
cross-package CLI goldens.

**Scientific computing:** That is code verification, not empirical validation.
Keep those labels separate. When fields arrive, require refinement studies,
positivity evidence, conservation ledgers, and comparisons against independent
reference data. A GPU result is not higher fidelity merely because it was
expensive.

**Project:** Accepted. Backends may accelerate a named model but cannot promote
its claim class.

**Scientific computing:** Multiple models would improve the instrument. Use the
forecast-office pattern: analytical guidance, reduced physics, field models,
ensembles, and optional learned surrogates can disagree in public. Keep model,
implementation, fitted artifact, run, and ensemble identity separate. State
what each ensemble perturbs. Agreement is not validation, and incompatible
observables do not become a consensus average.

**Project:** Could ML make large parameter sweeps interactive?

**Scientific computing:** Eventually. Begin with a bounded surrogate for named
integral outputs of an already verified solver, retain the solver as fallback,
and require held-out shifts, calibration, physical residuals, and explicit OOD
refusal. A learned closure must be tested inside the solver. A training loss is
not a conservation proof, and plausible graphics are not validation.

**Chemist:** “Gas composition” is too coarse. What do you intend to represent?

**Project:** Separate capabilities for bulk-mixture thermodynamics, trace-species
inventory, reaction networks, phase equilibrium and condensation, aerosol and
droplet chemistry, exterior chemistry, observer detection, and hazard
assessment.

**Chemist:** Preserve those separations in the schema and UI. Odor thresholds,
instrument detection limits, toxicity thresholds, and flammability limits are
different measurements. Detectable does not mean hazardous. Undetectable does
not mean safe. Never let a comic odor score stand in for concentration or risk.

**Project:** Accepted. Any safety statement will name species, concentration,
exposure duration, pathway, model, source, uncertainty, and jurisdiction where
applicable.

**Chemist:** What does a “big butt” preset change?

**Project:** Nothing hidden. A localized comic label expands to an inspectable
source-morphology parameter patch such as reservoir compliance, interface
geometry, symmetry, orientation, and soft-structure properties. It cannot
silently change composition, health, sex, gender, identity, intelligence, or
social worth. The same patch can be selected without the joke label.

**Mathematical physics:** What does a fart mean in a seventh dimension?

**Project:** The Lab cannot answer until the model supplies the dimensionality,
topology, metric signature, compactification geometry and scales, fields,
action or evolution law, boundary conditions, observables, and a mapping back
to the requested presentation. If those are absent, the result is
`INSUFFICIENT_MODEL_DEFINITION`, not a decorated three-dimensional plume.

**Mathematical physics:** And what does string theory say about a fart in
superposition?

**Project:** Superposition belongs to quantum theory generally. A string model
would have to declare which states are superposed, their amplitudes, dynamics,
decoherence environment, and measurement operator. String theory is not an
extra-dimension switch and does not license arbitrary macroscopic claims. A
toy quantum source, a string-inspired compactification study, an analogy, and a
fictional axiom pack receive different labels.

**Mathematical physics:** What about a fart with the power of the Big Bang?

**Project:** That phrase fails before energy is assigned. The Big Bang is not a
localized discharge into an existing exterior. The Lab returns:

```text
DISCHARGE BOUNDARY INVALID:
CASE RECLASSIFIED AS COSMOLOGICAL INITIAL DATA
```

A selected cosmological model can then expose its actual variables and limits.
It must not preserve an emitter, nozzle, plume, exterior, sound field, or damage
radius merely to keep the gag visually familiar.

**Mathematical physics:** Good. Precision should make the refusal funnier.
Maintain distinct claim classes for established executable physics, research
toy models, speculative models, analogies, and fictional axioms. Also separate
empirical support, applicability, implementation assurance, and verification
evidence. One “confidence” percentage would erase the distinctions.

**Human-computer interaction:** Why CLI first?

**Project:** It forces every action, parameter, result, explanation, and proof
to exist without a mouse, image, or undocumented native state. It supports
automation, remote compute, accessibility workflows, scientific pipelines, and
agent play. It also makes the initial product small enough to polish deeply.

**Human-computer interaction:** CLI first must not mean CLI forever or a GUI
that shells out to a command. Put canonical operations in shared services. The
CLI, htop-style Terminal Lab, native Godot application, MCP, and A2A adapters
should exercise the same contracts. Each surface can add presentation, but no
surface gets secret scientific powers.

**Project:** Accepted. Native means a real Windows, macOS, and Linux application,
not a browser shell, embedded webview, or local web server.

**Human-computer interaction:** A polished CLI needs completion, structured
output, standard input support, stable exit codes, responsive cancellation,
actionable errors, examples, accessible plain output, and a safe updater. The
Terminal Lab needs reliable resize and terminal restoration. The native app
needs remapping, reduced motion, captions, independent audio channels, and an
accessible first-run path.

**Play editor:** You are in danger of building a dissertation browser. Where is
the fun?

**Project:** The default is a banal, ordinary, low-energy encounter. The player
can laugh immediately, then inspect why it sounded or moved that way. Quick
Play starts instantly. Broadcast creates an interdimensional episode. Freestyle
opens the laboratory. Chill makes the simulation ambient. Challenges turn
conservation, similarity, pitch, orbit, and regime boundaries into goals.

**Play editor:** The humor should be a read-only interpretation of facts, not a
second physics engine. Allow `no_joke`. Localize through reviewed transcreation,
not national stereotypes. Aim jokes at the Lab's excessive certainty,
bureaucratic gravity, and category failures. Children should be able to enjoy
the sound and motion without being steered toward humiliation or unsafe acts.

**Project:** Accepted. Explorer, Lab, and Research views will reveal different
depth over the same retained calculation. A locale pack cannot modify physical
inputs, claims, score, or seed semantics.

**Play editor:** “Every fart is a snowflake” is a good product truth if you do
not misuse reproducibility. A retained account can be recalculated or presented
again. A fresh enactment remains a new encounter. Make the Plumeprint and
Fartflake honest projections with declared loss, not mystical fingerprints.

**Dependable systems:** Why Go, Rust, and Godot?

**Project:** Go remains the tiny independent oracle and current CLI. Rust is the
candidate shared simulation and service core because it supports explicit data
models, memory safety, native libraries, terminal applications, and CPU/GPU
integration without a garbage-collected frame loop. Godot supplies a native
cross-platform presentation layer. Optional accelerator kernels are isolated
behind verified backends. The choice is provisional until small vertical
experiments prove interoperability, build quality, and maintenance cost.

**Dependable systems:** Good. Do not turn “best language” into language fashion.
The best stack is the smallest one that can meet accuracy, performance,
portability, auditability, and contributor constraints. Every dependency needs
an owner, reason, license, update path, and removal strategy. Reject god files,
schema duplication, generated-code drift, unbounded parsers, quiet fallbacks,
and tests that reproduce the implementation formula.

**Project:** The quality floor is 90 percent aggregate statement coverage and
80 percent per non-generated package. Numerical and protocol code also receives
properties, fuzzing, independent fixtures, race checks, static analysis,
cross-platform tests, and eventually mutation and differential testing.

**Dependable systems:** Coverage is a floor, not evidence that the right thing
was tested. Keep files cohesive, errors typed, formats bounded, pure logic apart
from adapters, and generated artifacts reproducible. Replace growing collections
of shell-specific repository checks with one dependency-free cross-platform
quality tool. Do not let automation become a second undocumented build system.

**Chair:** Is this a real scientific tool or a game?

**Project:** It can be both if each claim remains visible. Explorer may hide
detail, but never invent certainty. Research mode may expose residuals without
claiming validation. A joke may name a boundary, but never move it. The same
account supports play, education, comparison, and research only within the
capabilities it has actually earned.

**Chair:** Then the product proposition is credible. The premise creates an
unusually memorable way to practice careful modeling. The billion-dollar joke
is not a promise of market value. It is the willingness to examine a silly
problem until useful architecture, experiments, and questions fall out of it.

## Decisions accepted

1. Keep the current v0.7 predictor named and documented as an exact finite
   reservoir endpoint, not full blowdown.
2. Require an applicability envelope and explicit nonclaims beside every
   scientific result.
3. Split chemistry, observer response, and hazard assessment into independent
   capabilities.
4. Represent humorous source presets as inspectable parameter patches with
   neutral aliases.
5. Use five advanced-physics claim classes: established executable physics,
   research toy model, speculative model, analogy, and fictional axiom pack.
6. Reject undefined extra-dimensional and cosmological prompts precisely rather
   than filling them with Earth-continuum defaults.
7. Keep humor and locale read-only, versioned, optional, and unable to mutate
   science.
8. Deliver each scientific and play capability through CLI, then Terminal Lab,
   then native surfaces.
9. Keep the oracle independently formulated and make backend promotion depend
   on verification evidence rather than speed.
10. Treat coverage as a minimum and enforce structural, property, fuzz,
    cross-platform, security, and scientific evidence gates.
11. Keep project automation cross-platform and progressively consolidate it in
    a small dependency-free repository-quality executable.
12. Make public humor, branding, radio, and merchandise state-bound and clear
    of false certification, cultural stereotypes, and borrowed media identity.
13. Add plural model guidance only through separate model, implementation,
    artifact, run, ensemble, comparison, and localized-display identities.
14. Introduce scientific ML as an optional bounded accelerator with
    preregistered baselines, solver-owned invariants, shifted-domain evidence,
    and fallback or refusal.

## Unresolved challenges

- Define the first aperture and exterior model without overstating biological
  evidence.
- Ratify a source-morphology vocabulary that works for biological, mechanical,
  collective, distributed, and source-free cases.
- Select chemistry mechanisms and databases with explicit range, revision, and
  licensing policies.
- Design human playtests that measure comprehension and delight without making
  cultural universality claims.
- Prove cross-language semantic parity before adding a second production core.
- Establish when an expensive field calculation produces a materially stronger
  claim than the analytical model.
- Specify agent observations that are rich enough for play but do not leak
  privileged simulator state.
- Design an original audiovisual ident and radio catalog that remain enjoyable
  when the premise is only a subtle part of the world.

The panel's final standard is simple: an ordinary `pfft` must be immediately
fun, scientifically modest, mechanically complete, and inspectable all the way
down. Extreme regimes earn their absurdity by passing the same discipline.
