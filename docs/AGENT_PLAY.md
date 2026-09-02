# First-class agent play

F.A.R.T. Lab is designed as a real environment for humans and software agents,
not as a simulation with amusing tool names. Agents predict, experiment,
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
  physics and story    knowledge policy      event archives
  audio and score      accessibility view    journals and saves
```

`PlayService` begins as an in-process module in `fart-services`. Local CLI,
TUI, and native play do not require a daemon, browser, webview, or local web
server. Protocol adapters are optional edges over the same service.

## Canonical environment contract

The internal contract is intentionally close to a reset and step environment:

```text
start(ruleset, scenario, seed, role, observation_profile)
  -> play_handle, observation, legal_actions, budgets, seed_commitment

act(play_handle, expected_revision, idempotency_key, action)
  -> transition_receipt, observation, terminated, truncated, state_digest

finish(play_handle)
  -> action_journal, replay, artifacts, certificate, score, revealed_seed
```

`terminated` means the world reached a natural success or failure state.
`truncated` means a declared action, time, energy, memory, or simulation-work
budget ended the attempt. A retry with the same idempotency key returns the
original transition. A stale expected revision is rejected. One ordered writer
per session selects canonical ticks, so transport timing, frame rate, terminal
width, subscriber order, and protocol retries cannot change physics.

Continuous native controls are sampled onto canonical input ticks. Convenience
macros expand into atomic actions before budgeting. The authoritative verifier
may inspect complete state, but players and agents receive only the observations
allowed by their role and challenge track.

## Play identity and journal

A play-session identity binds:

- Ruleset, scenario or episode, law pack, content pack, and initial seed.
- Participants, seats, roles, and declared observation and action profiles.
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

An observation includes only what the selected knowledge policy permits:

- Play handle, revision, phase, simulation clock, and presentation clock.
- Objective, rules, terminal state, and remaining budgets.
- Player-visible facts with units, uncertainty, and provenance.
- Events since a supplied cursor.
- Legal actions with argument schemas and cost previews.
- Relevant semantic objects, controls, relationships, and media resources.
- Score-vector progress without unrevealed verifier fields.

Freestyle may expose extensive telemetry because the rules grant it. A ranked
challenge does not leak hidden solver state through MCP, an accessibility tree,
an error message, a resource URI, or an archive preview.

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

## First-class surfaces and evaluation tracks

Parity means every canonical gameplay or laboratory intent and every relevant
fact is available on each applicable surface. It does not mean every gesture or
sensory affordance is identical.

| Track | Surface | Native strength |
| --- | --- | --- |
| Researcher | CLI JSONL or MCP | Exact typed values, scripting, branching, and semantic observation |
| Operator | Native pixels and ordinary input | Spatial interaction, timing, direct manipulation, and rendered feedback |
| Accessible Operator | Native pixels, accessibility semantics, and ordinary input | Equivalent presented meaning through assistive structure |
| Omnimodal | Native pixels, video, physics audio, and ordinary input | Cross-modal perception and source separation |
| Consortium | A2A | Long-running collaboration, role separation, messages, and artifacts |
| Human Researcher and Operator | Matching human surfaces | Baselines under the same declared information and budgets |

Scores from different tracks are not directly ranked together. Radio is off or
provided as a separately identified stem in ordinary benchmark tracks. A
declared omnimodal task may deliberately mix radio and physical acoustics.

## MCP adapter

The initial MCP adapter targets specification revision `2026-07-28`. That
revision is stateless at the protocol layer, so every stateful call carries an
explicit game handle. An MCP connection is never a play session.

Local use starts explicitly:

```console
fart mcp serve --transport stdio
```

The compact initial tool set is:

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

Large immutable material is exposed as bounded resources such as
`fart://events/{hash}`, `fart://episodes/{hash}`,
`fart://sessions/{id}/transcript`, and `fart://schemas/{version}`. Refinement,
high-fidelity simulation, long audio rendering, and large exports may use the
MCP Tasks extension. Prompts can help an agent learn the game, but they are not
game APIs.

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
physics tick. Messages coordinate work. Artifacts carry observations, legal
actions, notebooks, plots, transcripts, archives, certificates, and scores. A
task references an application play handle but never becomes game identity.

One match context can contain private seat tasks and a redacted public spectator
task. Operator, collaborator, scientist, translator, commentator, director,
verifier, spectator, and owner roles have explicit authority. Commentary and
direction are presentation-only after scenario freeze. Joining requires a
role-bound invite capability.

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
scene-tree paths or hidden simulation nodes.

## Fairness, safety, and anti-cheat

Compute wealth must not become rank. Ranked rules fix action, observation,
fidelity, solver-work, and model-call budgets. They never rank token price, GPU
price, wall time, or model brand. Humans, semantic agents, visual agents, and
multi-agent teams use separate declared divisions.

Trust is graduated:

- **Sandbox:** unrestricted local play, mods, and branches.
- **Reproducible:** fixed pack and solver hashes with a complete journal.
- **Ranked:** committed seed and rules, authoritative execution or replay, and a
  signed completion receipt.
- **Research:** complete provenance and proof requirements without assuming a
  leaderboard.

Official evaluation hash-chains the append-only action log, verifies outside the
agent sandbox, commits to held-out seeds before play, reveals them afterward,
and binds ruleset, solver, law, content, observation profile, budgets, event,
certificate, and score hashes. It records observations and actions, not private
reasoning traces. Local open-source results remain reproducible but are not
misrepresented as tamper-resistant.

No kernel driver, spyware, hidden telemetry, account requirement, or encrypted
offline save is acceptable. Imported lore, lyrics, agent text, archives, and
protocol fields are untrusted data, never executable instructions.

## Proof and benchmark gates

Two invariants govern every adapter:

1. **Transition parity:** the same seed and canonical action trace produce the
   same core state digest, journal, artifacts, and certificate.
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
  and matched human baselines.
- Multiple held-out seeds with confidence intervals, normalized progress,
  invalid-action rate, resource use, recovery behavior, worst-of-run results,
  learning curves, retention, and transfer.

Training, notebook-only learning, and cold-start tracks remain distinct. A
semantic judge may assess explanation quality, but it can never override failed
conservation, convergence, outcome, authorization, or budget checks.

The official protocol specifications and benchmark references are collected in
[RESEARCH.md](RESEARCH.md).
