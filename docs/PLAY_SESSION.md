# Experimental reservoir play session

The native Rust CLI implements a bounded reservoir experiment session with an
immutable baseline, explicit prediction attempts, read-only views, retained
receipts, integrity replay, and fresh reconstruction. This document describes that experimental
profile and its evidence limits.

The session keeps one authored reservoir baseline and lets an operator compare
explicit withdrawal fractions and closures. Every prediction begins from that
same baseline. A previous prediction never silently becomes the initial state
of the next one. The journal records Lab decisions, not elapsed source time or
a sequence of physical discharges.

The in-process service belongs to `fart-services`. The native CLI calls that
service directly through `fart play run <commands.jsonl|->`. The profile requires
no daemon, network connection, graphical surface, or external agent protocol.
It supplies a concrete part of the
[learning loop](GAMEPLAY.md#the-playable-learning-loop) and
[play contract](AGENT_PLAY.md#canonical-environment-contract).

## Run an experiment

From the repository root:

```console
cargo run --locked -p fart-cli -- play run testdata/play/reservoir-session.jsonl
cargo run --locked -p fart-cli -- play run testdata/play/reservoir-session.jsonl --format jsonl
mkdir -p artifacts
cargo run --locked -p fart-cli -- play run testdata/play/reservoir-session.jsonl --format transcript > artifacts/session.json
cargo run --locked -p fart-cli -- play replay artifacts/session.json --format json
cargo run --locked -p fart-cli -- play reconstruct artifacts/session.json --format json
```

The [checked-in session](../testdata/play/reservoir-session.jsonl) starts with a
four-attempt budget, predicts a 0.75 adiabatic withdrawal, observes the account,
retries that action, predicts a 0.5 isothermal withdrawal, discovers available
actions, explicitly finishes, and retries the earlier action again. It produces
eight replies and one end-of-input summary. The final summary has revision 3,
two attempts used, two remaining, `complete: true`, and `truncated: false`.
The latest account remains the isothermal prediction: 2 kg, 400 K, and
280000 Pa. Retrying the earlier receipt does not replace it.

The command source can be a filename or `-` for standard input. `play run`
accepts `--format text|jsonl|transcript`. Text is the default. JSONL contains one
reply for each processed command and a separate end-of-input summary containing
the retained transcript. Transcript output contains only that transcript, so it
can be redirected to a file without extracting an object from other output.
Framing uses LF; CRLF and a final line ending at EOF are accepted. Blank lines
are invalid commands. Unicode line separators inside JSON are data.

An incomplete session can still supply retained evidence. Transcript output
returns a failure exit status when no explicit finish was accepted; end of input
does not change its completion state. Shell redirection owns the destination
file and may replace it or leave partial output. This command does not supply
the Go carrier's atomic no-clobber file writer.

| Command | Exit 0 | Exit 1 |
| --- | --- | --- |
| `play run` | Input delivered, no rejected command, and explicit finish accepted | Rejected command, unfinished session, usage, input, transport, or output failure |
| `play replay` | Retained integrity verified, including an honestly unfinished session | Usage, input, retained-integrity, or output failure |
| `play reconstruct` | Exact current-implementation match, including an honestly unfinished session | Mismatch, fresh admission refusal, usage, input, retained-integrity, or output failure |

A costed model refusal is retained evidence rather than a rejected command.
It does not by itself make an explicitly finished run fail. A finish after
budget exhaustion can also return exit 0 while retaining `truncated: true`.

## Command profile

Every command carries `schema: fart.reservoir-play-command/v0alpha1`, an
`operation`, and `actor_id`. One JSON object occupies one line. Unknown
operations, fields, nulls, aliases, and incorrect types are refused.

| Operation | Additional required fields |
| --- | --- |
| `start` | `profile`, `role`, `session_nonce`, `idempotency_key`, `expected_revision`, `attempt_budget`, `measurement_interaction`, `knowledge_policy`, `termination_policy`, `baseline` |
| `predict` | `session_ref`, `idempotency_key`, `expected_revision`, `withdrawal_fraction`, `closure` |
| `observe` | `session_ref`, `view` |
| `actions` | `session_ref` |
| `finish` | `session_ref`, `idempotency_key`, `expected_revision` |

Start explicitly selects `reservoir-experiment/v0alpha1`, role `operator`,
measurement interaction `none`, knowledge policy `full-reservoir`, and termination
policy `explicit-finish-or-budget`. Its expected revision is zero and its
attempt budget is an integer from 1 through 16.

The baseline has schema `fart.reservoir-play-baseline/v0alpha1`, model
`continuum.rigid-calorically-perfect-ideal-mixture@v0alpha1`,
`quantity_system: si`, and explicit initial component, volume, and temperature
inputs. It contains neither a withdrawal fraction nor a
closure. Prediction requires those choices explicitly, using the existing
`rigid-adiabatic` or `rigid-isothermal` model closure for a supported calculation.

Observe selects `brief` or `research`. Both use the same full-reservoir knowledge
policy; research adds the complete retained report and normalized baseline.
Action discovery reports legal phases, argument requirements, attempt costs,
and remaining budget. Neither operation requires a mutation key or advances
revision.

Actor IDs, nonces, and idempotency keys are nonempty tokens of at most 64 bytes,
using lowercase ASCII letters, digits, and `._:-`. A session reference is
`sha256:` followed by 64 lowercase hexadecimal digits. Copy the reference from
the start receipt; it binds subsequent commands to that session, not to an
authenticated identity.

## Scope and boundaries

The selected mathematical model is the rigid, homogeneous, nonreacting,
calorically perfect ideal mixture already exposed by `reservoir predict`.
Its SI quantities and thermodynamic assumptions belong to that explicit model.
The session does not acquire an emitter, aperture, atmosphere, observer, sound,
or physical clock by recording decisions.

The operator has the full information supplied by this experimental profile.
Summary and complete views select from the same retained account. They do not
implement ranked secrecy, multiple seats, or a general knowledge-policy system.

This profile does not admit a universal Lab case or ratify `RecordIdentity`,
scenario identity, occurrence identity, measurement back-action, a certificate,
or the certified `.fart` archive. It does not replace the existing Go
[walk evidence carrier](WALK_EVIDENCE.md). The broader
[v0.8](../ROADMAP.md#v08-rust-production-core-and-typed-cli) and
[v0.9](../ROADMAP.md#v09-certified-case-archive) gates remain independent.
The [local identity decision](RECORD_IDENTITY.md) specifies this profile's
fingerprints and comparison policy without closing those broader gates.

## Authoritative state and views

An accepted transition owns its request, receipt, service revision, remaining
evaluation budget, and any retained model report. Those values belong to the
service. CLI formatting cannot change them.

The service has one writer per `PlayService` instance through Rust's mutable
borrow boundary. The public live service cannot be cloned or restored from a
transcript. This is not a cross-process lock: another instance can deliberately
start the same caller-supplied envelope and obtain the same session reference.

Read-only observation and action discovery must not:

- run the numerical model again;
- change the baseline or retained report;
- advance the service revision or consume an evaluation attempt;
- append a canonical journal entry;
- change the request, account, or journal fingerprint.

No result exists before the first successful prediction. A view must represent
that absence explicitly rather than inventing a zero-valued physical result.
The distinction between service ordering and source-law ordering follows the
[bounded universality contract](UNIVERSALITY.md#nothing-familiar-arrives-implicitly).

## Retries, refusals, and completion

An accepted idempotency key remains retained for the session's published
lifetime. An exact retry returns the original receipt, including after later
progress or completion. Reusing the key for a different canonical request is a
conflict. Idempotency lookup precedes stale-revision and terminal-state checks.
The returned revision, remaining budget, and attempt cost describe the original
acceptance. They are not a fresh observation or a second charge. Use `observe`
or the end-of-input summary for the current state.

Malformed, unauthorized, stale, and conflicting requests do not consume an
evaluation attempt or append an accepted transition. A well-formed prediction
that the numerical model refuses consumes one attempt and advances the service
revision while preserving the latest successful account. Diagnostic delivery
and retry counts do not become canonical journal identity material.

Unknown operations and missing, extra, null, or wrongly typed fields are parsing
refusals. An unsupported but well-formed closure token, or a finite fraction
outside the model's admitted range, is a costed prediction refusal. Its receipt
retains the failed attempt's diagnostic separately from the latest successful
account.

Reaching the evaluation limit is truncation, not scientific completion. The
operator can still finish after exhaustion, and the finished report retains
its truncated status. End of input does not silently invent a finish action.

The summary's `complete` field means that explicit finish was accepted:

| Service state | `complete` | `terminated` | `truncated` |
| --- | --- | --- | --- |
| Active | `false` | `false` | `false` |
| Evaluation budget exhausted, awaiting finish | `false` | `false` | `true` |
| Explicitly finished with attempts remaining | `true` | `true` | `false` |
| Explicitly finished after exhaustion | `true` | `false` | `true` |

## Identity and integrity

The caller supplies the session nonce explicitly. The service does not claim
that the nonce is fresh, random, secret, or globally unique. It is not a live
authority token and does not assert a new occurrence in the represented model.

The profile separates the session reference, accepted service revision,
normalized request fingerprint, retained account fingerprint, journal
fingerprint, and serialized bytes. Presentation selection cannot change the
authoritative account or action journal. The exact hashing and serialization
profile must be versioned with the implementation.

The fingerprint profile is `fart.play.rfc8785-sha256/v0alpha1`. The session
reference binds the normalized start command, including its nonce, actor, and
policies. The baseline fingerprint binds normalized typed inputs. The initial
account fingerprint identifies an explicit `not-evaluated` state; it does not
imply a calculated result. A successful prediction binds the complete retained
model report to that baseline. An accepted model refusal preserves the latest
successful account fingerprint while extending the journal with its own
request, diagnostic, cost, and receipt.

The selected serialization is the
[RFC 8785 JSON Canonicalization Scheme](https://www.rfc-editor.org/rfc/rfc8785).
It makes the supported JSON values suitable for reproducible hashing through
defined primitive serialization and property ordering. It does not preserve
authored bytes or define scientific equivalence.
Play-only normalization treats a withdrawal fraction of negative zero as zero;
it does not change the existing reservoir command's wire contract.

The transcript retains normalized commands and baseline components in canonical
ID order. It does not retain the exact authored JSONL bytes, whitespace, or key
order. Account hashing includes the complete model report and its implementation
revision. Exact digest comparison does not promise that independently rounded
reports from different runtimes will have identical hashes.

An unkeyed digest can establish internal consistency under its declared profile.
Someone who can replace a record can also recompute its digests. Hash agreement
does not establish authorship, authenticity, empirical validity, a real physical
event, or endorsement by a certificate authority.

## Retain and replay

The retained object has schema `fart.reservoir-play-transcript/v0alpha1` and
contains the fingerprint profile, genesis entry, accepted journal, and final
control summary. Read-only queries and accepted retries do not become new
journal entries. Rejection and delivery counters remain outside that identity.
`play replay` reads this object, not the mixed reply/summary JSONL stream.

Replay validates the current closed profile, command bindings, actor and session
references, revisions, accepted keys, attempt accounting, completion state, and
digest chain. Retained predictions must use the current implementation revision,
bind their authored inputs and explicit action, preserve ordered component
identities, and carry the exact current claim and nonclaim vocabulary. Retained
residuals must match their balance fields and fit their declared positive
tolerances. These are consistency checks on the supplied declarations; replay
does not derive the tolerances or independently establish the equations.

No reservoir evaluation runs during import or replay. The replay report states
`prediction_recomputed: false`, `numerical_verification: not-performed`, and
`authentication: not-established`. A record with altered positive numerical
results can pass integrity checks if its digests and retained declarations are
consistently recomputed. The test corpus preserves that counterexample.

## Reconstruct retained attempts

`play reconstruct <transcript.json|-> [--format text|json]` verifies the entire
retained profile and chain before numerical work. It then freshly admits the
normalized baseline and recomputes each retained prediction from that baseline,
including costed model refusals. It compares the complete resulting transcript,
so an altered earlier result cannot be hidden by an unchanged final account.

The report schema is `fart.reservoir-play-reconstruction/v0alpha1`, with comparison
profile `fart.play.canonical-current-implementation/v0alpha1`. `status` is
`matched`, `mismatched`, or `refused`. `retained_summary` preserves the expected
completion and fingerprint references. When fresh admission succeeds,
`reconstructed_transcript` retains all newly obtained reports and receipts.
A mismatch includes a `first_difference` JSON Pointer. Fresh admission failure
includes its stable `refusal`. Invalid retained structure or an unsupported
profile is refused before either fresh admission or prediction work.

`prediction_attempts_recomputed` reports the fresh work count. Verification
separately identifies retained integrity, control consistency, baseline admission,
prediction work, numerical comparison, and the absence of authentication:

| Condition | `baseline_admission` | `prediction_recomputed` | `numerical_verification` |
| --- | --- | --- | --- |
| Fresh baseline refused | `refused` | `false` | `reconstruction-refused` |
| Admitted baseline, zero prediction attempts | `admitted` | `false` | `no-prediction-attempts` |
| All retained prediction evidence matches | `admitted` | `true` | `matched-current-implementation` |
| Retained prediction evidence differs | `admitted` | `true` | `mismatched-current-implementation` |

An incomplete or truncated transcript can match. A finished transcript can
mismatch. `ReconstructionSummary::is_matched()` and `is_complete()` expose these
separate decisions; the latter reads the retained session's explicit finish.
The CLI returns 0 for a match, including a zero-prediction or unfinished session.
It does not infer completion from successful verification.

This exact comparison is distinct from tolerant Go/Rust numerical parity.
It does not promise identical floating-point results across platforms, compilers,
or implementations, and a mismatch alone does not prove a physical error.
Preserve the expected transcript and inspect the fresh evidence when they differ.
The [identity decision](RECORD_IDENTITY.md#why-exact-comparison-first) explains
the selected boundary.

`Transcript::reconstruct()` borrows immutable retained evidence and returns an
immutable comparison. It creates no live handle, new nonce, recorded occurrence,
or mutation of the expected transcript. Live-session restoration is not
implemented. The Go walk carrier's reconstruction remains a separate profile.

## Resource boundaries

| Boundary | Limit |
| --- | ---: |
| One JSONL command | 64 KiB |
| Entire command stream | 1 MiB |
| Commands in one CLI stream | 128 |
| Aggregate CLI output | 16 MiB |
| Evaluation attempts | At most 16 |
| Retained transcript | 8 MiB |
| JSON depth | 32 |
| Member-name bytes | 128 |
| JSON nodes in a command | 4096 |
| JSON nodes in a transcript | 65,536 |
| Accepted journal entries | 17 |
| Other JSON arrays | 64 entries |

The genesis entry is retained separately from the at-most-17-entry journal.
The command and output limits belong to transport. Hitting one reports
incomplete delivery and never invents a finish. A prediction may already be
accepted before its output fails; transport failure does not roll it back and
can prevent delivery of the final transcript. The transport failure itself is
not an additional prediction attempt.

The parser refuses an excess array element before decoding that element. The
existing reservoir model also bounds component count, numeric representability,
and model work. Resource refusal must occur before a partial authoritative
transition is committed. A bounded transcript does not imply a certified archive,
atomic file publication, or power-loss durability.

## Executable evidence

The [public service](../crates/fart-services/src/play/mod.rs),
[command parser](../crates/fart-services/src/play/wire.rs), and
[retained replay](../crates/fart-services/src/play/transcript.rs) own the
experimental contract. The [CLI transport](../crates/fart-cli/src/play.rs)
only frames input, calls the service, and delivers its projections.
The [reconstruction implementation and tests](../crates/fart-services/src/play/reconstruction.rs)
cover fresh numerical comparison, rehashed intermediate-result changes, costed
refusals, zero-attempt admission, and profile refusal before solver work.

The [service integration tests](../crates/fart-services/tests/play.rs) cover
baseline reuse, free views, retained retries after progress and finish, costed
model refusals, exhaustion, zero-attempt finish, hostile input, the literal
repository fixture, and rehashed evidence counterexamples. A separate
[execution-boundary test](../crates/fart-services/src/play/engine.rs) verifies
that import and replay enter numerical prediction zero times. The
[CLI tests](../crates/fart-cli/tests/cli.rs) cover framing, output formats,
help without reads, transport limits, and actual file/stdin behavior.
