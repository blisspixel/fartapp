# Model registry, ensembles, and scientific machine learning

Status: design contract. No ensemble or machine-learning command is implemented
yet.

Weather forecasting made model pluralism part of ordinary conversation. People
compare a deterministic run, an ensemble, a regional model, and “the European
model” without expecting them to be the same calculation. F.A.R.T. Lab can use
that familiar forecast-desk grammar for an absurdly rigorous purpose: several
declared models may examine one compatible question, disagree visibly, and earn
different kinds of trust.

The joke is model guidance delivered with grave institutional composure. The
science is an extensible registry, explicit comparison semantics, calibrated
ensembles, and a hard boundary between equations, learned approximations, and
presentation.

## Non-negotiable distinctions

The Lab does not define “the best model” as one universal scalar. It keeps six
objects separate:

1. A `ModelDefinition` states equations, rules, learned mapping, assumptions,
   inputs, outputs, domain, and revision.
2. A `ModelImplementation` binds that definition to code, precision, backend,
   dependencies, and verification evidence.
3. A `ModelArtifact` identifies immutable trained weights, fitted coefficients,
   lookup data, preprocessing, serialization, and digest independently of code.
4. A `ModelRun` applies one implementation and any required artifact to one
   compatible request with complete provenance.
5. An `Ensemble` is a declared population of related runs with a sampling and
   interpretation contract.
6. A `ModelComparison` maps compatible observables into a common comparison
   space and reports agreement, disagreement, and missing comparability.

A model can be mathematically respectable but inapplicable to the case. An
implementation can be verified while the model remains empirically weak. Two
models can agree and both be wrong. A fast learned surrogate can be useful
without becoming a governing law. None of those axes becomes a single
“confidence” percentage.

The [simulation claim classes](SIMULATION.md) remain in force:
`established_physics`, `research_toy`, `speculative_model`, `analogy`, and
`fictional_axiom`. Machine learning is an implementation or modeling method,
not a sixth epistemic class.

## Registry contract

Every registered model version declares at least:

```text
model_id and model_version
display-name keys, never one mandatory human-language name
provider identity and authority basis
law-context compatibility
supported operations
model claim class
equations, rules, architecture, or mapping revision
required inputs and produced observables
units, dimensions, coordinates, and reference frames where applicable
applicability and extrapolation envelope
closures and unresolved terms
uncertainty contract
verification, validation, and calibration evidence
known failures and explicit nonclaims
implementation and backend inventory
determinism and random-stream contract
resource estimates and admission limits
license, source, weight, data, and artifact provenance
security and trust status
```

Candidate Earth-continuum display families illustrate the tone without freezing
new scientific IDs:

| Stable scientific role | Candidate display family |
| --- | --- |
| Current rigid ideal-mixture endpoint | Pfft Reference |
| Future reduced source and puff model | Breeze |
| Future axisymmetric Euler field model | Gust |
| Future compressible viscous field model | Squall |
| Future Euler-Lagrange multiphase model | Drizzle |
| Future learned integral-output surrogate | Shortcut, always marked `SURROGATE` |

These are presentation candidates, not implemented model registrations or
claims that fidelity forms a universal ladder.

Stable model identifiers are locale-invariant Lab protocol tokens. Display
names, explanations, forecast-desk nicknames, and pronunciation belong to
reviewed presentation packs. An English name cannot select a different model.
The same model can have French, Japanese, Arabic, Spanish, Hindi, or another
reviewed presentation without changing its bytes, run identity, or score.

Real external model and provider names are factual registry metadata when an
adapter legitimately supports them. The project does not borrow institutional
logos or imply partnership. Built-in comic forecast copy uses original fictional
names such as `Continental Guidance`, `Local High Resolution`, and `The Wetness
Ensemble`, subject to cultural and trademark review.

## Ensemble taxonomy

“Ensemble” is not one operation. Each ensemble declares what varies:

| Kind | What varies | What spread may indicate |
| --- | --- | --- |
| Initial-condition | Uncertain admitted starting state | Sensitivity to initial uncertainty |
| Parameter | Priors or calibrated uncertain parameters | Parametric uncertainty inside the selected model |
| Stochastic-closure | Named random terms in one closure | Conditional variability introduced by that closure |
| Numerical | Mesh, timestep, precision, algorithm, or backend | Numerical sensitivity, not physical probability by default |
| Structural | Closure or model family | Model-form disagreement |
| Learned | Stochastic samples or independently trained members | Learned conditional distribution or training sensitivity, as declared |
| Multi-model | Different compatible model definitions | Inter-model guidance, not automatically one probability distribution |

The system never treats a numerical-refinement set as a physical ensemble or a
collection of unrelated models as equally likely samples. Weighting, exchangeability,
dependence, calibration, and probability semantics must be declared before an
ensemble can produce probabilities. Otherwise it produces a comparison of
members and spread only.

An ensemble report includes member failures and refusals. It does not discard a
valid but inconvenient member, reroll for a funnier outcome, or hide a model
that crosses a boundary. Missing members change the ensemble identity and are
visible in every summary.

## Forecast Desk presentation

The planned Forecast Desk is a read-only view over registered runs. In an Earth
continuum case it may show a regime plume, uncertainty cone, spaghetti traces,
probability bands, and model disagreement. In an atemporal, nonspatial, or
nonprobabilistic context it substitutes an applicable relation or comparison
view. It never fabricates a map, clock, geography, probability, or “forecast.”

Illustrative presentation copy may say:

```text
MODEL GUIDANCE
Analytical Endpoint       applicable, exact under declared assumptions
Reduced Puff Model        applicable, verification grade V1
Local High Resolution     unavailable, field backend not installed
Wetness Ensemble          not applicable, no multiphase capability

CONSENSUS WITHHELD
The models do not currently share the requested observable.
```

When facts support it, a humor pack may render a line such as “The Continental
Model remains stubbornly subsonic.” The underlying report still contains the
neutral model identifier, actual regime claim, frame, evidence, and limits. A
joke cannot manufacture consensus, probability, or a wetness boundary.

Useful visual grammar includes:

- A cone only for a declared distribution over a meaningful coordinate.
- Spaghetti traces for individual members, never as a replacement for summary
  calibration.
- Median, intervals, exceedance probability, and member count only when the
  ensemble defines their statistics.
- Side-by-side residuals and conservation errors for deterministic comparisons.
- A model-disagreement alert that explains which assumptions or closures differ.
- A `NO COMMON OBSERVABLE` result instead of forcing incomparable outputs onto
  one chart.

## Scientific machine-learning roles

ML enters only through a named role:

- **Surrogate or emulator:** approximates outputs of a specified simulator or
  experiment inside a declared domain.
- **Learned closure:** supplies an unresolved constitutive, turbulence,
  chemistry, interface, or subgrid term inside governing equations.
- **Hybrid solver:** combines an explicit numerical dynamical core with learned
  components.
- **Inverse estimator:** infers parameters or latent state from declared
  observations.
- **Reduced-order model:** advances a compressed state with a stated lifting and
  projection error.
- **Probabilistic generator:** samples a declared conditional distribution of
  states or outcomes.
- **Classifier or detector:** predicts a regime, anomaly, applicability warning,
  or failure condition without replacing the associated physical calculation.
- **Presentation model:** produces captions, jokes, narration, or visual style
  and has no scientific authority.

The role determines what can be claimed. A surrogate trained on CFD predicts
that CFD implementation's outputs, not nature directly. A learned turbulence
closure remains part of the model and must be tested solver-in-the-loop. A
classifier can warn of choking but cannot certify it without its own evidence
contract. A language model may draft a scenario proposal, but only accepted
typed fields reach the scientific service.

Scientific ML enters progressively. The first candidate approximates selected
scalar or integral outputs of an already verified solver for sweeps. Later
candidates may emulate property tables, advance a reduced trajectory, or supply
a solver-in-the-loop closure. Probabilistic field generation comes only after
the narrower roles have trustworthy data, baselines, stability, and calibration
evidence. The verified solver remains available as an authority and fallback.

## Learned-model evidence

Every scientific learned model adds a versioned model card with:

```text
training objective and loss terms
architecture and parameter count
training-code, framework, and compiler revisions
weight digest and serialization format
training, validation, calibration, and held-out test datasets
dataset origin, license, consent, filtering, and known gaps
input normalization and output transformation
physical invariances and constraints encoded
random initialization and training nondeterminism record
hyperparameter and selection procedure
in-domain and shifted-domain metrics
calibration method and reliability evidence
out-of-distribution detector and refusal policy
failure corpus and adversarial tests
compute, energy where measurable, and hardware provenance
permitted claims and explicit nonclaims
```

Training and test leakage is a release blocker. Model selection uses validation
data; final claims use untouched test and challenge sets. Data generated by one
solver cannot independently validate that solver. Biological data requires
ethics, consent, minimization, and privacy review. Ordinary play is local and
does not become training data unless a separate informed opt-in contract exists.

Large weights need not live in the source repository. An optional model pack
uses signed metadata, immutable digests, size limits, license records, platform
compatibility, and explicit installation. The base CLI, exact oracle, and
ordinary offline play remain useful with no model download, cloud account, or
accelerator.

## Physical constraints and failure policy

Conservation penalties are not conservation proofs. A learned output may earn a
conservation claim only if the complete coupled update closes the declared
ledger within tolerance. Projection onto an admissible manifold can enforce a
named constraint, but it does not establish accuracy, uniqueness, stability, or
validity of the remaining state.

Where applicable, learned components must address:

- Positive density, pressure, temperature, mass, concentration, and other
  admissibility conditions without silent clamping.
- Rotational, translational, reflection, permutation, gauge, unit, and coordinate
  behavior required by the selected model.
- Conservation, entropy, causality, realizability, and symmetry constraints
  actually owned by the law and closure.
- Autoregressive stability, spectral behavior, phase error, and long-horizon
  drift.
- Geometry, topology, Reynolds, Mach, composition, pressure-ratio, resolution,
  and boundary-condition shifts.
- Uncertainty calibration, sharpness, coverage, dependence, and rare-regime
  performance rather than average point error alone.

An out-of-domain request returns a typed refusal or falls back to an applicable
verified model. It never quietly extrapolates because the generated field looks
plausible. A learned model cannot be the sole judge of its own applicability,
uncertainty, or correctness.

The OOD gate reports exactly `in_domain`, `near_boundary`, `out_of_domain`, or
`undetermined`. It combines schema and law checks, explicit dimensional and
dimensionless bounds, distance from supported training data, ensemble
disagreement, invariant residuals, and rollout drift as applicable. No one
learned score can overrule a structural incompatibility. Near-boundary use is
visible and policy-controlled; out-of-domain use falls back or refuses.

Model disagreement uses typed causes such as `initial_state_sensitive`,
`parameter_sensitive`, `closure_sensitive`, `resolution_sensitive`,
`model_structure_split`, `surrogate_out_of_domain`, `no_shared_observable`, and
`unresolved`. Presentation may dramatize those results but cannot combine them
into an unexplained confidence meter.

## Verification and comparison gates

Before scientific promotion, a learned or hybrid model must pass:

1. Schema, weight, preprocessing, and reference-vector conformance.
2. Deterministic or explicitly bounded stochastic replay at the declared level.
3. Independent analytical, manufactured, or empirical tests appropriate to its
   claim, not only agreement with training data.
4. Baseline comparisons against the simplest adequate analytical and numerical
   models, including accuracy and total cost.
5. Conservation, admissibility, symmetry, dimensional, and limiting-behavior
   tests where applicable.
6. Solver-in-the-loop stability for learned closures and hybrid components.
7. Held-out geometry, parameter, regime, and distribution-shift challenges.
8. Calibration and ensemble scoring for probabilistic claims.
9. CPU reference execution or an explicit accelerator-only capability report,
   plus backend differential tests.
10. Dependency, weight-format, deserialization, supply-chain, privacy, license,
    and model-extraction threat review.

ML is promoted only when it improves a named product or scientific objective,
such as latency, fidelity at fixed cost, uncertainty quality, inverse inference,
or adaptive control. A larger model with a better demo image is not sufficient.

## Planned command surface

The typed CLI precedes TUI, native, MCP, and A2A adapters:

```console
fart model list --operation case.run
fart model inspect continuum.reduced-puff@v1
fart model explain continuum.reduced-puff@v1 --field applicability
fart ensemble plan reference-enclosure.toml --profile initial-condition.v1
fart ensemble run plan.json --members 64 --output guidance.fart
fart ensemble inspect guidance.fart --view research
fart model compare guidance.fart --observable interface.mach
fart model compare run-a.fart run-b.fart --nondimensional
```

`ensemble plan` resolves member construction, named random streams, model
versions, budget, expected outputs, comparison space, and failure policy before
execution. `--dry-run` performs no sampling. Long runs support progress,
cancellation, checkpoints, and resumable task receipts. Structured output uses
the same model and ensemble schemas exposed through MCP and A2A.

The Terminal Lab adds sortable guidance panes, member brushing, disagreement
drill-down, and terminal-safe plots. The native app adds spatial projections and
animation only for compatible cases. Neither surface receives a private model
or hidden ensemble member.

## Agent play

Models create a strong advanced-agent loop:

1. Discover available models, costs, observables, and evidence.
2. State a hypothesis and select a comparison or ensemble plan.
3. Spend a bounded experiment budget.
4. Observe only role-authorized summaries or artifacts.
5. Diagnose disagreement, request a refinement, or branch a parameter.
6. Submit a conclusion with cited run and evidence identifiers.

Agents can optimize ensemble value, choose fidelity under budget, detect model
failure, transfer a learned closure to a held-out universe, or prove that a
spectacular result is a numerical or learned artifact. The evaluator retains
hidden truth where a challenge requires it. MCP, A2A, CLI JSONL, TUI, and native
automation expose equal canonical actions and role-appropriate information.
Agent ergonomics are specified in [AGENT_PLAY.md](AGENT_PLAY.md).

## Research basis

The weather analogy is bounded but useful:

- [ECMWF operational AIFS ensemble](https://www.ecmwf.int/en/about/media-centre/news/2025/ecmwfs-ensemble-ai-forecasts-become-operational)
  demonstrates physics-based and learned ensemble systems operating side by
  side rather than forcing one method to erase the other.
- [ECMWF forecast uncertainty](https://www.ecmwf.int/en/research/modelling-and-prediction/quantifying-forecast-uncertainty)
  distinguishes uncertainty from initial conditions and model formulation and
  motivates recording what each ensemble actually varies.
- [GenCast](https://www.nature.com/articles/s41586-024-08252-9) demonstrates a
  learned probabilistic weather model evaluated as an ensemble, including
  calibration and joint spatiotemporal behavior rather than point error alone.
- [NeuralGCM](https://www.nature.com/articles/s41586-024-07744-y) demonstrates a
  hybrid numerical dynamical core with learned components and also documents
  extrapolation limits under changed climate conditions.

The fluid-modeling constraints come from:

- [Duraisamy, 2021](https://doi.org/10.1103/PhysRevFluids.6.050504), which
  emphasizes model-consistent training, physical constraints, and limited
  generalization for ML-augmented turbulence closures.
- [Shankar et al., 2025](https://doi.org/10.1103/PhysRevFluids.10.024605), which
  reports solver-in-the-loop differentiable closure training and ensemble
  uncertainty evaluation across changed flow conditions.
- [Taira, Rigas, and Fukami, 2025](https://doi.org/10.1103/8t52-mtb9), which
  gives a current critical assessment of machine learning in fluid dynamics and
  its remaining generalization, data, and community-validation problems.
- [Physics-consistent output projection, 2025](https://www.nature.com/articles/s42005-025-02329-1),
  which demonstrates strong enforcement of selected physical manifolds while
  leaving the broader accuracy and applicability questions separate.

These works motivate the architecture. They do not validate a learned fart
model, endorse the project, or justify transferring weather skill claims to
compressible multiphase discharge.

## Promotion gate

The first model-comparison feature can ship only when:

- Two independently useful compatible models are already verified through the
  CLI.
- Their common observable and comparison transform are explicit.
- Different model, implementation, run, ensemble, and presentation identities
  survive archive round trips.
- Incompatible requests produce `no_shared_observable` or another exact refusal.
- English and at least one structurally different reviewed locale select the
  same model IDs and results.
- One ordinary case makes disagreement understandable and funny without hiding
  uncertainty or pretending consensus is truth.

The first scientific ML feature ships only after it beats a named non-ML
baseline on a preregistered objective, passes the learned-model evidence gate,
works as an optional model pack, and can be removed without breaking the base
instrument.
