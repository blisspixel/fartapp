# First-class agent play

F.A.R.T. Lab is designed as a real environment for humans and software agents,
not as a thin simulation wrapper with amusing tool names. Agents predict, experiment,
observe, keep evidence, recover from mistakes, cooperate, and submit verifiable
results. There is no `make_optimal_fart` action and no hidden shortcut around
the game.

CLI first describes delivery order, not a privileged control plane. Every
surface consumes one versioned play contract:

```text
CLI       TUI       Godot       MCP       A2A       native automation
 |         |          |          |         |                |
 +---------+----------+----------+---------+----------------+
                            |
                    canonical PlayService
                            |
        +-------------------+-------------------+
        |                   |                   |
  session reducer    observation projector   artifact store
        |                   |                   |
  case evaluation      knowledge policy      case archives
  audio and score      accessibility view    journals and saves
```

`PlayService` begins as an in-process module in `fart-services`. Local CLI,
TUI, and native play do not require a daemon, browser, webview, or local web
server. Protocol adapters are optional edges over the same service.

## Canonical environment contract

The implemented [local reservoir profile](PLAY_SESSION.md) is a bounded subset:
one explicit operator, immutable baseline predictions, no measurement
interaction, full-reservoir knowledge, costed attempts, exact retries, and
retained integrity replay. The generalized contract below remains a design
target. Local session references are instance correlation data, not the
expiring authenticated play handles described for future remote operation.

The internal contract is intentionally close to a reset and step environment.
Its revisions and transitions belong to the PlayService control plane. They do
not imply discrete source-law time, Markov state, or familiar causality:

```text
start(ruleset, scenario, optional_seed_request, operation_policy, role,
      measurement_interaction_profile, view_profile)
  -> play_handle, observation, legal_actions, budgets,
     optional_generation_commitment, optional_record_commitment

act(play_handle, expected_revision, idempotency_key, action)
  -> transition_receipt, observation, terminated, truncated, account_digest

finish(play_handle)
  -> action_journal, trace, artifacts, certificate, score, reconstruction_policy
```

`terminated` means the play objective reached a declared terminal condition.
`truncated` means a declared action, time, energy, memory, operation-work, or
evaluation-work budget ended the attempt. A retry with the same idempotency key returns the
original transition. A stale expected revision is rejected. One ordered writer
per session selects canonical service revisions, so transport timing, frame
rate, terminal width, subscriber order, and protocol retries cannot change the
Lab account. A source-law clock or physical tick exists only when a
selected evolution profile defines it.

For generated public play, `optional_seed_request` may contain a
player-selected seed and the response may reveal it. A nongenerated case omits
the seed. When a ranked case uses generation, the evaluator selects a hidden
seed and record nonce. Before play, it returns commitments only:

```text
generation_commitment = H(seed || normalized_scenario_digest || event_nonce)
record_commitment = H(normalized_scenario_digest || event_nonce)
```

A nongenerated ranked case has only the record commitment when recording is
selected.

Here `event_nonce` is the public wire name for a `RecordNonce`. It commits the
Lab record and case result. It affects realization identity only when a selected
operation defines realization, and it never changes context-occurrence identity
claims by itself.

The evaluator reveals the committed values after completion. A ranked caller
cannot choose or inspect them before play.

For an evolution profile with continuous controls, native input is sampled onto
canonical input ticks. Convenience
macros expand into atomic actions before budgeting. The authoritative verifier
may inspect complete state, but players and agents receive only the observations
allowed by their role and challenge track.

## Play identity and journal

A play-session identity binds:

- Ruleset, scenario or episode, law pack, content pack, and any applicable
  initial seed.
- Participants, seats, roles, and declared measurement, view, and action profiles.
- The ordered canonical action journal and every transition receipt.
- Checkpoints, parent session, branch point, and produced artifact identities.
- Completion state, budgets, score vector, and verification receipt.

Transport connection IDs, MCP requests, A2A task IDs, window focus, volume,
camera motion, caption style, and terminal layout are not game identity. A
resume continues one lineage. Acting from an old immutable state deliberately
creates a named branch.

The game issues opaque, expiring, role-bound play handles. Handles are authority,
not portable archive identifiers, and are never written into public replays.

## Actions and observations

A canonical action has a stable semantic ID, typed arguments, actor role,
expected revision, idempotency key, and declared cost. It never contains a
Godot node path, terminal coordinate, translated label, shell command, arbitrary
URL, or arbitrary filesystem path.

Malformed and unauthorized requests are rejected before transition, consume no
action budget, and do not advance revision. A well-formed but law-inadmissible
or out-of-scope action returns a canonical no-op receipt, consumes the declared
attempt cost, and advances the service revision or applicable decision ordering
without claiming a physical state transition. A relation or constraint query
can be a legal action without pretending it caused an occurrence transition.
A stale revision returns the current revision without transition. Exact
duplicates return the original receipt. Reports count each category separately,
and every adapter implements the same policy.

An observation includes only what the selected knowledge policy permits:

- Play handle, service revision, phase, and presentation ordering, plus a
  source ordering, simulation clock, or dependency coordinate only when the law
  supplies one.
- Objective, rules, terminal state, and remaining budgets.
- Player-visible facts with units, uncertainty, and provenance.
- Observations and service changes since a supplied cursor.
- Legal actions with argument schemas and cost previews.
- Relevant semantic objects, controls, relationships, and media resources.
- Score-vector progress without unrevealed verifier fields.

Freestyle may expose extensive telemetry because the rules grant it. A ranked
challenge does not leak hidden solver state through MCP, an accessibility tree,
an error message, a resource URI, or an archive preview.

## Agent Quality of Life contract

This is the planned usability contract for advanced G5-style agents and simpler
automation. It is not a claim that the current CLI already implements
`PlayService`, MCP, A2A, task orchestration, or every operation below. A feature
becomes current only when the applicable capability manifest, command help,
schemas, examples, conformance tests, and product status documentation agree.

The goal is not merely to make the game callable. An agent should be able to
discover what exists, form a valid request, understand a bounded response,
conduct long research work, recover from interruption, and collaborate without
reverse-engineering prose or scraping presentation output.

### Discovery, schemas, and examples

Every applicable surface must expose one versioned capability catalog generated
from the canonical service definitions. Discovery includes:

- Stable semantic action, observation, artifact, role, and error identifiers.
- Argument and result schemas, required capabilities, legal phases, authority,
  declared cost dimensions, limits, and content or safety boundaries.
- Supported protocol revision, adapter revision, optional extensions, and
  negotiated feature flags.
- Minimal, typical, boundary, refusal, and recovery examples using canonical
  typed values. Localized labels may accompany examples but are never schema
  keys.
- Pagination, filters, exact lookup, and a catalog digest so an agent can cache
  safely and detect a changed contract.

Unsupported capabilities are omitted from legal actions and remain discoverable
as unavailable only where revealing their existence is allowed. Discovery never
reveals a hidden value, reference solution, private role, seed, nonce, or
verifier-only field.

Natural language is a drafting convenience over this catalog. A request first
produces a typed, non-mutating proposal with assumptions, ambiguities, rejected
effects, source spans where applicable, and schema references. The actor then
accepts an exact proposal revision and digest. Drafting never implies acceptance
and cannot silently choose a hidden default that changes scientific or gameplay
state.

### Adjustable observations

An agent can request `brief`, `standard`, or `research` observation density, or
use a field mask and declared response budget. It can ask for a full snapshot or
a delta after an observation cursor. Concision changes representation only. It
does not remove terminal status, remaining applicable budgets, uncertainty,
provenance, safety information, required acknowledgments, or the means to
discover legal actions.

Truncation is explicit and returns omitted-field classes plus bounded continuation
or artifact references. Summaries cite the retained facts and artifacts from
which they were produced. An agent never has to parse decorative CLI tables,
localized narration, terminal escape sequences, screenshots, or audio to obtain
canonical machine information.

Planned subscriptions deliver ordered observation deltas, task progress, newly
available artifacts, messages, and terminal changes without client polling.
Each delivery carries a cursor and service revision. Reconnect resumes from an
acknowledged cursor when retention permits, otherwise it returns a typed
`cursor_expired` result and the smallest authorized resynchronization path.
Coalescing may reduce delivery frequency but cannot reorder canonical changes or
hide a terminal transition.

### Long-running work and interruption

Long simulation, refinement, verification, export, batch, and campaign work uses
an explicit task handle distinct from play identity. A task reports a typed
phase, completed and remaining work where knowable, budget use, latest safe
checkpoint, produced artifacts, warnings, and failure state. It reports a
percentage only when the denominator is defined. Otherwise it reports named
milestones and measured work without inventing certainty.

Cancellation is cooperative, bounded, and acknowledged. The receipt states
whether work stopped before acceptance, at a safe checkpoint, after a committed
transition, or after completion. Cancellation never silently rewinds the Lab
account, discards already committed evidence, or converts partial work into a
successful result.

Checkpoints are immutable and content-addressed. Resume identifies the exact
checkpoint, parent task, service and pack compatibility, budgets already spent,
and required role. An incompatible resume fails with a typed explanation and
does not approximate state. Acting from an older checkpoint creates an explicit
branch with a parent reference and branch-point digest. It never overwrites the
original lineage.

### Durable notebooks and artifacts

An agent notebook is an append-only workspace for hypotheses, plans,
observations, citations, artifact references, and declared conclusions. It is
not authoritative simulation state and cannot turn an opinion into evidence.
Entries cite stable observation revisions or content-addressed artifacts. A
notebook can be private to a seat, shared with named roles, or published by an
explicit action. Export and import preserve authorship as a role-scoped record
without exposing credentials or private reasoning.

Large arrays, fields, plots, audio, transcripts, certificates, archives, and
proof material return durable bounded artifact references instead of flooding
an observation. Metadata states media type, schema, byte size, digest,
provenance, authorization, retention, and available ranges or projections.
References do not grant more authority than the requesting handle and do not
smuggle hidden state through names, sizes, previews, or error behavior.

### Batch, sweep, and ensemble work

Batch, parameter-sweep, and ensemble actions accept typed member definitions or
a bounded generator, declared aggregation, concurrency ceiling, failure policy,
and total work budget. The service validates the complete envelope before
acceptance and exposes member costs or upper bounds. It never converts one
budgeted action into unbounded hidden solver work.

Members have stable identities and individual receipts. Partial failure remains
visible, successful members remain referenceable, and retries address only
failed or explicitly selected members. Aggregation never erases missing,
refused, invalid, or inconclusive members. Execution order, worker count, and
transport delivery order cannot change the defined account result. Numerical
equivalence uses the solver's declared reproducibility contract rather than an
unsupported promise of bitwise equality.

### Recovery and deterministic local baselines

Every failure returns a stable code, failing stage, human-readable explanation,
machine details allowed by the role, retryability, budget effect, and one or
more valid recovery actions when they exist. Revision conflicts return expected
and current revisions. Schema failures identify exact fields. Capability
failures name the required capability without revealing protected state. A
retryable transport failure is not confused with a rejected game action.

All mutating requests require an idempotency key. Repeating an accepted key with
the same canonical request returns the original receipt. Reusing it for a
different request is rejected. The service retains accepted keys for the
published lifetime of the live play or task. An adapter may expire transport
cache entries, but it cannot silently resubmit an expired key as a new mutation.
An expired handle returns a typed recovery or terminal result.

The planned SDK and protocol examples ship with offline deterministic local
fixtures for discovery, drafting, start, observe, act, task progress,
cancellation, checkpoint, resume, branch, notebook, batch, finish, and error
recovery. These fixtures provide a stable control-plane baseline and declared
scientific tolerances. They do not claim that every high-fidelity backend or
hardware target is bitwise identical.

### Fair roles and multi-agent coordination

The knowledge policy is identical for equivalent roles across direct service,
CLI JSONL, MCP `2026-07-28`, A2A `1.0`, TUI semantics, and native accessibility
semantics. Concision, subscriptions, artifact metadata, progress, timing, and
errors are tested as possible side channels. A richer transport cannot reveal
more task-relevant state than its declared track permits.

Multi-agent sessions provide explicit roles, inbox and outbox messages, shared
artifact references, communication budgets, handoff receipts, role withdrawal,
and deterministic conflict rules. Private seat observations and notebooks do
not enter shared context without an authorized action. Agents may delegate work
or transfer a role only when the ruleset grants that authority. A coordinator
cannot bypass another role's observation or action policy.

MCP and A2A provide parity of canonical intent and authorized information, not
identical protocol shapes. MCP exposes bounded tools, resources,
resource and task subscriptions, and long-running task controls for one
agent-service relation. Its no-polling quality path uses negotiated
`subscriptions/listen` notifications; Tasks polling remains a compatibility
fallback for clients that do not negotiate subscriptions. A2A exposes the same
underlying play and artifact semantics through A2A 1.0 tasks, messages,
artifacts, role coordination, `SubscribeToTask` or streaming, and cancellation.
Neither adapter shells out to the `fart` CLI, parses its output, or treats a CLI
subprocess as the service. Both call the same in-process service boundary used
by the CLI.

There is no required dashboard, browser login, localhost website, hidden
webview, or web-only setup step. A capable agent can complete every applicable
machine workflow through an applicable direct-contract, CLI JSONL, MCP, or A2A
surface. The TUI and native application remain first-class optional play
surfaces, not secret control planes.

### Quality gates for this contract

Before an adapter claims Agent Quality of Life conformance, tests must prove:

- Capability catalogs, schemas, examples, legal actions, and error codes agree
  for shared canonical intents across direct service, CLI JSONL, MCP
  2026-07-28, and A2A 1.0, while accurately declaring adapter-specific
  capabilities.
- Observation density, field masks, deltas, truncation, subscriptions, and
  artifact projections preserve authorized information and reveal no hidden
  fields.
- Disconnect, duplicate delivery, retry, stale revision, cancellation, resumed
  subscription, expired cursor, checkpoint, resume, and branch cases recover
  without duplicate transitions or silent data loss.
- Notebook and artifact authorization survives export, import, role handoff,
  expiry, and attempted cross-seat access.
- Batch, sweep, and ensemble results are invariant to allowed scheduling and
  report every member disposition.
- Deterministic offline fixtures pass without network access, a browser, a
  daemon, a CLI subprocess inside an adapter, or optional presentation assets.
- A hostile parity suite probes timing, error, progress, size, resource, and
  accessibility channels for hidden-state leaks.
- At least one sustained G5-style campaign completes with multiple roles,
  disjoint observations, interruption, resume, branching, notebook handoff,
  long-running work, and final verification.

## Natural-language generation

Agents and humans may describe a desired case or result in ordinary language. The system
compiles the request into the same typed scenario proposal and canonical action
plan used by every exact interface. It is meant to be one of the most expressive
fart generators available without weakening scientific control.

The proposal response contains applicable typed law contexts, scope, optional
structures, capabilities, measurement or view profiles, resolved values,
assumptions, ambiguities, unsupported effects, safety and content boundaries,
field-level source spans, budgets, and a stable interpretation receipt. Units,
source, exterior, and observer fields appear only when supported. Drafting does
not mutate a play session. The actor must accept a specific proposal revision
before it becomes a scenario.

The MCP surface may expose:

- `scenario_draft`
- `scenario_validate`
- `scenario_accept`
- `scenario_explain_interpretation`

Natural language never becomes a privileged action such as “make this optimal.”
The accepted proposal expands into budgeted canonical actions. Benchmarks can
use exact typed input, natural-language input, or both as separately reported
tracks. Hidden verifier state and reference solutions never enter a language
prompt or interpretation error.

## F.A.R.T. Challenge Grades

There is no universal research meaning for agent difficulty level 4. The project
therefore publishes its own versioned, measurable grades. Every manifest also
declares expected horizon, partial observability, branching, stochasticity,
irreversibility, proof burden, modalities, budgets, and team size.

| Grade | Operational definition | Example |
| --- | --- | --- |
| G0 Interface Check | 1 to 5 actions, full observation, no irreversible choice | Start, change one value, observe, and replay |
| G1 Pfft | 5 to 20 actions, one main control, dense feedback | Reach a target loudness without crossing wetness |
| G2 Lab Bench | 20 to 100 actions, coupled controls, noisy sensing | Infer an aperture response and tune a note |
| G3 Wind Tunnel | 100 to 300 actions, sparse milestones, constrained resources | Build and verify a stable vortex sequence |
| G4 Research Campaign | 500 to 2,000 actions across experiments, hidden parameters, delayed effects, persistent notebook, and consequential choices | Reproduce an event at 1,000 times scale and pass convergence checks |
| G5 Interdimensional Consortium | 1,000 or more actions, multiple roles with disjoint observations, communication cost, held-out worlds, and shared proof | Cooperatively derive and verify a cross-law translation |
| Open Research | No known optimal policy, declared generator family, held-out cases, verifiable outcome, and Pareto scoring | Discover a stable acoustic or transport regime under a fixed budget |

`C-Sharp Correspondence` is the target G4 demonstration: infer an unfamiliar
law profile, produce a dry event within a declared tuning tolerance, reproduce
its active similarity signature at another scale, and prove the result is not a
numerical artifact.

Every generator publishes versioned train and validation partitions plus the
hash of a private held-out partition. A private feasibility oracle audits
solvability without exposing a solution path. Challenge reports disclose the
generator family, partition, duplicate and contamination checks, and a fixed
normalized-progress function. Resetting or repeating an easy subgoal cannot
increase progress.

## First-class surfaces and evaluation tracks

Parity means every canonical gameplay or laboratory intent and every relevant
fact is available on each applicable surface. It does not mean every gesture or
sensory affordance is identical.

| Track | Surface | Native strength |
| --- | --- | --- |
| Researcher | CLI JSONL or MCP | Exact typed values, scripting, branching, and semantic observation |
| Operator | Native pixels and the declared keyboard, pointer, or controller contract | Spatial interaction, timing, direct manipulation, and rendered feedback |
| Accessibility-Semantics Operator | Native pixels, accessibility semantics, and assistive input contract | Equivalent presented meaning through assistive structure |
| Omnimodal | Native pixels, video, applicable physical audio, and declared input contract | Cross-modal perception and source separation |
| Consortium | A2A | Long-running collaboration, role separation, messages, and artifacts |
| Human Researcher and Operator | Matching human surfaces | Baselines under the same declared information and budgets |

Scores from different tracks are not directly ranked together. Radio is off or
provided as a separately identified stem in ordinary benchmark tracks. A
declared omnimodal task may deliberately mix radio and physical acoustics.

When an evolution profile defines paused and real-time semantics, each
applicable track also declares a clock division:

- **Paused decision:** operation evaluation and presentation stop at decision
  boundaries. Model inference latency is reported but excluded from any
  applicable represented evolution.
- **Real time:** the world continues during inference, deadlines are canonical,
  and latency affects the result.

Paused and real-time scores are never ranked together.

Human baselines include disabled participants using their chosen assistive
technology under the same declared information, action, and work budgets.

## MCP adapter

The initial MCP adapter targets specification revision `2026-07-28`. That
revision is stateless at the protocol layer, so every stateful call carries an
explicit game handle. An MCP connection is never a play session.

Every request carries the required protocol and client metadata for that
revision. A later HTTP transport supplies `Mcp-Method` and `Mcp-Name` for routing
and authorization. Remote authorization validates issuers and uses Client ID
Metadata Documents rather than adding new Dynamic Client Registration work.
The September 2026 review found a stable Rust SDK, `rmcp` 3.2.0, with Tier 1
support for this revision. Stability does not substitute for the adapter's
revision-specific conformance gate. The implementation decision and current
portable skill are recorded in [AGENT_INTEGRATION.md](AGENT_INTEGRATION.md).

Local use starts explicitly:

```console
fart mcp serve --transport stdio
```

The compact initial tool set is:

- `scenario_draft`
- `scenario_validate`
- `scenario_accept`
- `scenario_explain_interpretation`
- `play_start`
- `play_observe`
- `play_actions`
- `play_act`
- `play_checkpoint`
- `play_finish`
- `lab_simulate`
- `lab_explain`
- `lab_verify`
- `lab_translate`

The four `lab_*` operations are unrestricted only in Sandbox or Research play.
In a ranked challenge they are absent unless the manifest exposes the same
capability as a costed canonical action. They cannot reveal hidden state, run a
private feasibility oracle, obtain an unbudgeted refinement, or bypass the
action ledger.

Large immutable material is exposed as bounded resources such as
`fart://records/{hash}`, `fart://episodes/{hash}`,
`fart://sessions/{id}/transcript`, and `fart://schemas/{version}`. Refinement,
high-fidelity simulation, long audio rendering, and large exports may use the
MCP Tasks extension. Prompts can help an agent learn the game, but they are not
state-mutating game APIs. Natural-language drafting remains an explicit bounded
tool operation with typed output and a separate acceptance step.

New code does not depend on deprecated MCP Roots, Sampling, Logging, or legacy
HTTP+SSE. Structured semantic audio lets an agent receive propagation, pitch,
confidence, sonification, score, station, and synchronized speech facts without
pretending it heard PCM. Raw bounded audio remains available in audio-capable
tracks.

The local profile has no shell, arbitrary network, credential, or unrelated
filesystem authority. A separately enabled network transport requires
authentication, per-role authorization, quotas, expiry, request limits, and a
remote-hosting threat review.

## A2A adapter

A2A complements MCP. MCP connects an agent to tools and resources. A2A
coordinates autonomous agents through longer, stateful tasks. Implementation is
pinned to the latest reviewed A2A 1.0 patch while advertising wire version
`1.0`.

An Agent Card may advertise these implemented skills:

- `fartlab.play.v1`
- `fartlab.spectate.v1`
- `fartlab.science.v1`
- `fartlab.translate.v1`
- `fartlab.commentate.v1`
- `fartlab.direct.v1`

An A2A Task represents a campaign, match, experiment, or team role, not every
backend step. Messages coordinate work. Artifacts carry observations, legal
actions, notebooks, plots, transcripts, archives, certificates, and scores. A
task references an application play handle but never becomes game identity.

One match context can contain private seat tasks and a redacted public spectator
task. Operator, collaborator, scientist, translator, commentator, director,
verifier, spectator, and owner roles have explicit authority. Commentary and
direction are presentation-only after scenario freeze. Joining requires a
role-bound invite capability.

Each match declares sequential or simultaneous turns. Sequential turns bind an
actor to an exact observation revision. Simultaneous turns collect commitments
until a deadline, reveal together, and resolve conflicts with a versioned
deterministic rule. Deadlines, missing actions, disconnect grace, role handoff,
agent dropout, ties, and cancellation are part of the ruleset. A2A cancellation
maps explicitly to a natural termination, budget truncation, role withdrawal,
or no-op. It never silently rewinds the authoritative Lab account. Requests
carry `A2A-Version: 1.0`.

The first binding is authenticated JSON-RPC over explicitly started loopback
HTTP with streaming. HTTP+JSON and gRPC are advertised only after equivalent
behavior passes independent tests. Push notifications remain disabled until
webhook authentication, duplicate delivery, HTTPS, SSRF, and retry-budget
controls exist. There is no invented A2A stdio binding.

## Native and vision-agent play

Godot remains a real visual play surface. An opt-in native automation driver
tests the rendered application through captured frames, platform accessibility
semantics, focus, window metrics, keyboard, pointer, controller, and
accessibility-invoke actions. It does not call solver mutation APIs.

Automation begins only through an explicit launch mode, uses inherited standard
I/O where practical, displays a visible indication, and does not open an
always-listening port. Screenshots and accessibility trees cite the same
presentation revision. Semantic nodes reveal only presented information, never
scene-tree paths or hidden evaluation state.

## Fairness, safety, and anti-cheat

Compute wealth must not become rank. Ranked rules fix action, observation,
fidelity, solver-work, and model-call budgets. They never rank token price, GPU
price, wall time, or model brand. Humans, semantic agents, visual agents, and
multi-agent teams use separate declared divisions.

Trust is graduated:

- **Sandbox:** unrestricted local play, mods, and branches.
- **Reproducible:** fixed pack and solver hashes with a complete journal.
- **Ranked:** committed rules and any applicable seed, authoritative execution
  or replay, and a signed completion receipt.
- **Research:** complete provenance and proof requirements without assuming a
  leaderboard.

Official evaluation hash-chains the append-only action log, verifies outside the
agent sandbox, commits to held-out seeds before play, reveals them afterward,
and binds ruleset, solver, law, content, measurement and view profiles, budgets, record,
certificate, and score hashes. It records observations and actions, not private
reasoning traces. Local open-source results remain reproducible but are not
misrepresented as tamper-resistant.

The ranked receipt specifies signature algorithm, evaluator identity, public-key
distribution, key rotation and revocation, timestamp and clock policy, and the
command that verifies a replay. Reproducible local and trusted official results
are separate result classes.

No kernel driver, spyware, hidden telemetry, account requirement, or encrypted
offline save is acceptable. Imported lore, lyrics, agent text, archives, and
protocol fields are untrusted data, never executable instructions.

## Proof and benchmark gates

Two invariants govern every adapter:

1. **Transition parity:** the same accepted initial case, any applicable seed,
   and canonical action trace produce the same core account digest, journal,
   case result, scientific artifacts, and
   certificate. Presentation artifacts may differ but cite the same canonical
   presentation revision.
2. **Information parity:** surfaces in one track expose equivalent
   task-relevant information and no hidden verifier state.

Required evidence includes:

- Model-based reducer tests over randomized legal and illegal action sequences.
- Same-trace direct-service, CLI JSONL, MCP, A2A, and native action-lowering
  tests.
- Revision conflict, duplicate request, idempotency, retry, reconnect,
  cancellation, checkpoint, resume, and branch tests.
- Cross-platform save, replay, and digest tests.
- Knowledge-policy and observation-leak tests.
- Spectator read-only and role-authorization tests.
- Accessibility-tree, semantic-to-pixel, visible-window, and raw-input tests.
- Radio and Symphony isolation tests.
- Official MCP conformance and Inspector checks.
- Official A2A TCK checks for every advertised binding and ITK checks across at
  least three maintained SDK peers.
- Legal-random, greedy, notebook-planning, scripted reference, hostile-agent,
  and matched human baselines. Scripted references are labeled non-leaderboard
  upper bounds; every baseline publishes its action-sampling policy.
- Grade-specific repeated held-out runs with a preregistered minimum sample
  count, bootstrap confidence intervals, arithmetic mean, median, variance,
  `pass^k`, worst-of-n, tie policy, normalized progress, invalid-action classes,
  resource use, recovery, learning curves, retention, and transfer.
- Agent build, harness, system and task prompts where publishable, notebook and
  memory policy, model settings, token use, wall latency, and monetary cost.
  Cost remains metadata unless a challenge explicitly ranks it.
- Human baseline recruitment, novice or practiced status, practice allowance,
  sample size, applicable clock division, measurement and view profiles, and primitive-action
  accounting.

Training, notebook-only learning, and cold-start tracks remain distinct. A
semantic judge may assess explanation quality, but it can never override failed
conservation, convergence, outcome, authorization, or budget checks.

The official protocol specifications and benchmark references are collected in
[RESEARCH.md](RESEARCH.md).
