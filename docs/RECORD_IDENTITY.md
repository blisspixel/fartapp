# Local reservoir record fingerprints and reconstruction

This decision applies only to the native `reservoir-experiment/v0alpha1`
profile. It defines how its normalized requests, retained reports, journal, and
reconstruction comparison relate. It does not ratify the general
`RecordIdentity` RFC, allocate a universal case identity, or establish an
identity relation in a represented world. Those remain separate
[roadmap decisions](../ROADMAP.md) under the
[bounded universality contract](UNIVERSALITY.md#identity-and-impermanence).

The existing [play session](PLAY_SESSION.md) supplies the concrete command and
retention contract. Reconstruction is a new verification operation against
that retained evidence. This decision does not change the existing command,
baseline, transcript, or fingerprint wire revisions by implication.

## Decision

Use exact RFC 8785 canonical-value comparison under an explicit, current native
implementation profile for the first reservoir transcript reconstruction.
Its comparison profile is `fart.play.canonical-current-implementation/v0alpha1`;
the report schema is `fart.reservoir-play-reconstruction/v0alpha1`.
Validate the entire retained profile and integrity chain before numerical work.
Then admit the normalized baseline through the live numerical boundary and
recompute each accepted prediction request, including costed model refusals.
Compare the complete resulting transcript, not only the last successful account.

The retained transcript remains immutable. Reconstruction does not restore a
live writer, append to the retained journal, choose a fresh nonce, change the
recorded completion state, or claim another physical occurrence. A separate
service instance performs bounded verification work and returns a comparison.

## Identity separation

| Value | What it identifies in this profile | What it does not establish |
| --- | --- | --- |
| Authored JSONL bytes | The caller's original stream, if separately preserved | The transcript does not retain its exact whitespace, member order, or delivery history |
| Baseline fingerprint | Normalized model, SI quantities, and components in canonical ID order | A general scenario, calibrated biological event, or context-occurrence identity |
| Request fingerprint | The complete normalized accepted command, including actor, key, revision, and explicit action fields | A request's transport delivery, permission from another process, or scientific equivalence |
| Session reference | The normalized start envelope, including its caller-supplied nonce and policies | Freshness, secrecy, authentication, or globally unique live authority |
| Service revision | Accepted mutation order within this session | Elapsed time or physical ordering in the selected model |
| Account fingerprint | An explicit not-evaluated state, or the complete latest successful report bound to the baseline | Independent validity, authorship, or an approximate numerical equivalence class |
| Journal fingerprint | The ordered accepted requests and receipts, including costs and previous journal references | Read-only observations, repeat delivery of an accepted receipt, or an immutable external audit authority |
| Transcript value | The declared schema, fingerprint profile, genesis, accepted journal, and control summary | A certified `.fart` container, exact authored bytes, or a restorable live session |
| Reconstruction comparison | A new evaluation's agreement with the retained transcript under its named comparison policy | A replacement account, new scientific identity, or certified empirical truth |

The live service enforces one mutable writer per `PlayService` instance. Another
instance can deliberately start the same envelope and receive the same session
reference. No cross-process lock or external authorization mechanism is implied.
The `actor_id` check binds commands to the declared local operator; it does not
authenticate a person.

## Canonicalization and binding

The fingerprint profile is `fart.play.rfc8785-sha256/v0alpha1`. Its hash input is
an RFC 8785 serialization of an envelope containing `profile`, `domain`, and
`value`. Separate domains identify baseline, request, session, account, genesis,
and journal-entry material. The journal hashes each entry before adding that
entry's own resulting fingerprint; its receipt already binds the preceding
journal reference.

Typed parsing precedes canonicalization. It rejects duplicate members,
unsupported fields, invalid Unicode, nonfinite quantities, and the profile's
resource violations. Baseline components are ordered by their validated IDs;
play withdrawal negative zero normalizes to zero. Integer control fields remain
within the exact supported range. Canonicalization does not repair a refused
request or authorize a new operation.

The [RFC 8785 specification](https://www.rfc-editor.org/rfc/rfc8785) defines
primitive serialization and member ordering for supported JSON values. It does
not define model equivalence, measurement uncertainty, or scientific identity.
Equivalent supported JSON presentations can therefore share fingerprints while
their authored byte sequences differ.

An unkeyed SHA-256 digest detects inconsistency relative to the retained
comparison material. Anyone replacing the material can also recompute the
digest chain. Integrity verification alone does not authenticate a record or
establish that its predicted quantities came from the declared equations.

## Why exact comparison first

The existing account and journal already bind complete report values exactly,
including the implementation revision, claim metadata, diagnostics, and model
nonclaims. Canonical comparison follows that declared representation and detects
a changed intermediate account even when a later successful prediction restores
the same final account. It also compares refused predictions instead of silently
discarding them from verification.

The independent Go/Rust parity tests use a declared numerical allowance. That
is a different operation: it compares implementations while preserving their
different revision labels. It does not produce an interchangeable account
fingerprint. Importing that allowance into reconstruction would change the
comparison meaning and could hide a small, representable change in retained
evidence. In particular, a tolerance supplied inside the retained report cannot
grant permission to accept changed numbers.

Approximate agreement is also not generally transitive. It is suitable for a
named comparison with justified per-quantity rules, not an implicit identity
relation. A later tolerant reconstruction profile would need independent
tolerance ownership, exact structural and diagnostic comparisons, tiny-value
counterexamples, and an explicit relationship to the retained exact hashes.
That work is deferred; it is not a prerequisite for this bounded exact profile.

An exact mismatch does not prove that either calculation is physically wrong.
Rust documents unspecified precision for operations such as
[`f64::powf`](https://doc.rust-lang.org/std/primitive.f64.html#method.powf).
The native implementation revision alone does not promise identical floating
point results across every compiler, platform, optimization setting, or future
runtime. The comparison reports what this execution obtained and preserves the
distinction between exact agreement and cross-platform numerical portability.

## Execution and result boundaries

Import and replay validate the complete retained schema, implementation profile,
control chain, bindings, and declared report consistency without evaluating the
reservoir. A malformed or unsupported report anywhere in the journal must be
refused before reconstruction enters numerical work, even when earlier entries
would be valid.

Reconstruction then performs live baseline admission even if no prediction was
recorded. Typed baseline validity during replay does not establish that its
derived numerical state can be represented. A forged but consistently rehashed
genesis can therefore pass integrity and fail fresh admission. Such a result
must not be reported as a successful numerical match.

For an admitted baseline, every retained prediction is evaluated from that same
baseline. Accepted model refusals are part of the comparison. Read-only views
and accepted retries need no reconstruction because they did not create journal
entries. Finish repeats the retained control decision without adding a physical
calculation or inferring any missing action.

Matching and completion are independent. An unfinished or budget-truncated
transcript may reconstruct exactly. A finished transcript may mismatch. The
comparison preserves the retained `complete`, `terminated`, and `truncated`
values rather than using them as synonyms for verification success. A
zero-prediction transcript must disclose that no prediction was recomputed.

Retained account and journal references remain distinct from fresh comparison
references. A mismatch must remain inspectable without replacing the expected
evidence with the newly obtained values. A structural or profile refusal is
separate from a validly formed comparison that failed to match.

The comparison report retains the following distinctions:

| Field | Meaning |
| --- | --- |
| `status` | `matched`, `mismatched`, or `refused` under the named comparison profile |
| `retained_summary` | The original completion, budget, account, and journal summary |
| `reconstructed_transcript` | The newly obtained transcript when fresh baseline admission succeeds; it does not replace retained evidence |
| `prediction_attempts_recomputed` | Actual retained prediction attempts evaluated afresh, including costed model refusals |
| `first_difference` | A JSON Pointer to the first canonical-value difference, present only for a mismatch |
| `refusal` | The stable refusal when fresh baseline admission fails |

Successful retained validation remains `integrity: verified` and
`control_plane: verified`. `baseline_admission` separately reports `admitted` or
`refused`; `prediction_recomputed` states whether any prediction was actually
evaluated. An admitted zero-attempt transcript reports
`numerical_verification: no-prediction-attempts`, even when its canonical
comparison status is `matched`. Otherwise numerical verification is separately
`matched-current-implementation`, `mismatched-current-implementation`, or
`reconstruction-refused`. Authentication remains `not-established`.
The report does not duplicate the full retained transcript beside its fresh
result. The caller already supplies that expected evidence.

Canonical numeric spelling matters to the comparison policy: JSON `1` and
`1.0` have the same canonical value. A differing-field pointer must use the same
relation as the overall match decision, rather than flagging a harmless source
spelling before the actual changed field.

The existing bounds remain effective: 8 MiB of transcript input, at most 16
prediction attempts, at most 17 journal entries plus genesis, and bounded JSON
depth, nodes, member names, and component arrays. Reconstruction cannot acquire
an unlimited work budget from supplied metadata. Output limits and write
failures affect delivery; they cannot mutate the retained object or turn an
undelivered result into a reported match. The synchronous operation introduces
no task cancellation, checkpoint, atomic file publication, or durability claim.

## Verification obligations

The bounded implementation needs executable evidence for:

- A normal retained session, a costed model refusal, a zero-prediction finish,
  and incomplete and exhausted sessions, with match independent of completion.
- A positive numerical value altered and consistently rehashed so that replay
  accepts integrity while reconstruction reports a mismatch.
- An altered earlier prediction followed by an unchanged final account, proving
  the comparison covers the entire journal.
- Unsupported schemas, fingerprint profiles, and implementation revisions,
  including a late journal violation, before any numerical execution.
- A typed but numerically inadmissible rehashed baseline, proving fresh admission
  is not skipped for a zero-prediction transcript.
- Exactly one fresh evaluation for each retained prediction attempt, with none
  caused by import, replay, views, or accepted retries.
- Immutable retained bytes and fingerprints before and after matching,
  mismatching, and failed output delivery; bounded input and output refusal;
  help that never reads a transcript.

The [service](../crates/fart-services/src/play/mod.rs),
[transcript implementation](../crates/fart-services/src/play/transcript.rs),
[reconstruction implementation and tests](../crates/fart-services/src/play/reconstruction.rs),
[service integration tests](../crates/fart-services/tests/play.rs), and
[native CLI tests](../crates/fart-cli/tests/cli.rs) are the relevant implementation
and verification owners. The [session contract](PLAY_SESSION.md) records the
delivered API, wire schema, and command examples.
