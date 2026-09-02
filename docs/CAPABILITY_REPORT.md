# Capability report ratification candidate

Status: design candidate, not ratified.

This document records the decomposition gate for a future singular
capability-report wire contract. The current Go oracle emits
`fart.law-capability-report/v0alpha1`. That token remains provisional. Neither
this document nor the current output is a stable cross-language contract.

The current report places eight separately reported, non-collapsing questions
about named capabilities under one exact catalog entry. Those questions do not
form one universal assessment algebra. They belong to several Lab-level
contracts that may be absent or unevaluated. The report is not a scenario,
operation request, admission decision, case, occurrence, realization, solver
result, certificate, or cross-context compatibility report.

## Decision: do not freeze the current report

The present report has valuable executable evidence, but ratifying it now would
freeze accidental limitations:

- collections and aggregate encoded size have no explicit report-level bounds;
- the machine-token grammar permits punctuation-only identifiers;
- law inspection preserves authored capability order while scenario validation
  sorts resolved capabilities;
- localized presentation uses a provisional locale grammar;
- evidence records can describe only repository-local Go tests;
- `validated/empirical_reference` is accepted even though no empirical evidence
  record can be represented;
- `none/no_evidence` and `unknown/not_evaluated` do not yet prohibit references;
- unreferenced evidence-registry entries are accepted;
- `unknown`, `undetermined`, and the process reason `not_evaluated` need a
  ratified distinction;
- `within-default-budget` does not identify the budget profile it assessed.

More fundamentally, the eight fields mix catalog authority, implementation
inventory, scientific-model assessment, case applicability, claim evidence,
authorization, execution planning, and resource admission. Requiring all eight
as a universal core would turn the current application workflow into an
ontology requirement.

The roadmap therefore separates a minimal evaluation disposition from
profile-owned outcomes and their enclosing contracts. Presentation, evidence
payloads, singular catalog binding, scenarios, and operations remain separate
versioned contracts. The existing eight-field report remains a useful alpha
projection while that decomposition is tested.

## Candidate evaluation disposition

The smallest shared candidate describes only whether a Lab evaluation produced
an outcome. A future enclosing profile must name the evaluator and exact input
binding:

- `not_evaluated`: the evaluator produced no outcome. This makes no claim that
  the question is meaningful, answerable, unknown, or ordered in source-law
  time.
- `evaluated`: the declared procedure completed and produced exactly one
  outcome owned by that evaluator's profile.

`not_evaluated` forbids an outcome. `evaluated` requires one. Explanatory
diagnostics may identify missing or unresolved prerequisites, but they do not
fabricate an outcome. `not_applicable` and
`no_supported_determination`, when an owning profile defines them, are evaluated
outcomes rather than aliases for skipped evaluation.

Generic `unknown` and `undetermined` are not part of this shared candidate.
Formal outcomes such as multiple admissible results, insufficient constraints,
failure of a declared decision procedure, or budget exhaustion belong to the
profile that defines their meaning and limits.

This candidate is Lab-process state only. It requires no source time, observer,
body, object, language, geometry, quantity, operation, backend, policy, or
budget. An implementation experiment must remain internal and non-wire until
its owning inputs and outcome types are explicit.

## Current axes and future owners

No current axis may be collapsed into or inferred solely from another. Any
cross-axis consistency constraint must be explicit and separately validated.
Their candidate ownership is:

| Current field | Candidate owning contract |
| --- | --- |
| `law_definition` | Catalog Registration Profile after separately ratified Declaration Authority Resolution, Declaration Attribution Scope, and Catalog Lookup Closure; maturity remains separate |
| `implementation` | Provider or build profile bound to implementation revision and platform contract |
| `closure` | No generic field; optional formal-system specification-sufficiency or law-specific closure profile |
| `applicability` | Case or operation evaluation with exact request binding and validity rules |
| `evidence` | Evidence about individually identified claims, not one capability-wide grade |
| `trust` | Split among provenance assurance and operation authorization or refusal |
| `backend_feasibility` | Execution plan bound to implementation, backend, precision, and determinism contract |
| `resource_feasibility` | Admission decision bound to an explicit resource-budget profile |

An optional `specification_sufficiency` profile may ask whether a declared
evaluator established that every capability-defined specification obligation
resolves and that the requested result relation is well formed. It cannot by
itself assert deductive closure, completeness, consistency, existence,
uniqueness, determinism, temporal initial-value structure, numerical
well-posedness, stability, computability, decidability, or empirical truth.
Continuum profiles may define their own narrower closure checks.

Case, measurement, operation, backend, policy, and budget inputs belong to
future enclosing contracts. Missing required authored values are validation
failures, not scientific insufficiency.

Omission from a capability collection means only “not reported.” It never means
`law_does_not_define`, impossible, incompatible, refused, or unavailable.

## Catalog registration experiment

The first owner-profile experiment in `internal/catalogregistration` remains
internal and non-wire. One four-part binding contains three exact revisioned
references plus one capability token:

- the catalog scope named by the binding;
- the declaration authority to which a registration is attributed;
- the subject revision for which registration is queried;
- the capability token.

Revision is identity material, not source-law time or chronological ordering.
Declaration authority is catalog attribution. It does not establish authorship,
personhood, legal ownership, legitimacy, jurisdiction, trust, or universal
standing. The current opaque reference does not distinguish a Lab application,
law context, or any other authority category. Those variants need separate
review before product integration.

The profile composes with the shared evaluation disposition:

- `not_evaluated` has no registration-presence outcome;
- evaluated `registered` can produce a positive structural `Registration` value;
- evaluated `not_registered` cannot produce that value and applies only to the
  exact binding.

A future product evaluator must not convert invalid input, unresolved authority,
ambiguous scope, incomplete lookup, or omission into `not_registered`. The
structural experiment validates binding syntax and result shape only. It does
not resolve an authority reference, establish declaration-attribution scope,
close a catalog lookup domain, or produce a closed-lookup witness. Those product
semantics must exist before a CLI or wire projection can be ratified.

The package defines no maturity value or ordering. It makes no source-law truth,
scientific validity, implementation, applicability, evidence, authorization,
backend, resource, case, occurrence, realization, or operation claim. Its token
grammar and TSV cases are bounded Go-domain candidates, not product protocol.

Current `candidate`, `declared`, `not-declared`, and
`not-applicable/application_capability` values are not mapped into this profile.
They currently conflate registration presence, authority, and maturity, so every
existing report byte remains provisional and unchanged.

## Finite catalog snapshot experiment

The internal, non-wire experiment in `internal/cataloglookup` stores a
bounded immutable copy of positive structural registrations under one exact
catalog-scope reference. `SnapshotLookup` accepts only an exact same-scope
binding, decides membership against every stored registration, and retains the
snapshot with the structural result.

Within this experiment, `registered` means the complete binding is in the
supplied in-process set. `not_registered` means only that the binding is absent
from that set. An empty snapshot can therefore produce `not_registered`, but
only relative to its explicitly empty set. A scope mismatch is an error and
never becomes absence. Input order has no semantic effect, duplicate entries
are rejected, and the package API does not expose the underlying structural
`catalogregistration.Result`. Callers must retain `SnapshotLookup` to validate
the snapshot-relative membership claim.

This finite-snapshot property does not establish that an external catalog was
fully supplied, that a catalog-scope or authority reference resolves, that an
attribution rule exists, or that any registration is true. It defines no source
time, chronology, person, institution, lawgiver, observer, object, quantity,
language, network, universe, operation, evidence, trust, or maturity. It adds no
wire form and changes no current output byte.

## Exact declaration-authority reference matching experiment

The internal, non-wire experiment in `internal/authoritymatching` stores a
bounded immutable copy of a caller-supplied multiset of positive
`DeclarationAuthorityRecord` values under one exact catalog-scope reference.
The snapshot preserves record multiplicity but assigns no semantic meaning to
input order.

For one exact `AuthorityMatchBinding`, the experiment counts records whose
catalog-scope and declaration-authority references both match exactly. It
reports `no_match_in_snapshot` for zero matches, `one_match_in_snapshot` for one,
and `multiple_matches_in_snapshot` for more than one. Only the one-match outcome
can produce a positive `SnapshotAuthorityMatch`. `AuthorityMatch` retains the
complete supplied snapshot and exposes the exact match count, and the package
exposes no detached generic result.

Repeated records do not invalidate the snapshot. Cardinality is query-specific:
multiple records for reference A do not prevent a one-match result for reference
B in the same snapshot. Multiple matches are a structural count only. They do
not establish semantic ambiguity, conflicting identity, invalid authority data,
mistrust, corruption, malice, or distinct entities in a represented context.
Unlike the registration snapshot, which rejects a duplicate exact registration,
this authority snapshot is deliberately a multiset so repeated record entries
remain observable. Its key is complete only because the current authority record
contains nothing beyond its self-binding. Any future field must be retained or
included in record identity.

The experiment performs exact token equality only. It does not normalize text,
follow aliases or delegation, infer hierarchy or authority category, select a
latest revision, dereference a URI, contact a network service, or consult ambient
state. Revision is identity material, not chronology. Neither a one-match nor a
no-match result establishes external catalog completeness, global uniqueness,
legitimacy, standing, permission, authorship, ownership, attribution scope,
trust, source truth, scientific truth, or registration presence. No match never
becomes registration `not_registered`.

This exact match-cardinality evidence satisfies only a structural precondition.
The candidate below defines one snapshot-relative rule for how that evidence
participates in a resolution decision. It does not ratify product resolution.
Accepting repeated records here does not commit a future wire format or
authority profile to accepting them.

## Declaration-authority resolution candidate

The internal, non-wire candidate in `internal/authorityresolution` consumes one
valid retained `AuthorityMatch`. It stores that complete match evidence rather
than copying a caller-supplied outcome, count, binding, or record. Resolution is
therefore deterministic for the retained snapshot and exact query, without
consulting ambient locale, clock, filesystem, network, process order, or
randomness.

The candidate owns exactly three snapshot-qualified outcomes:

- `not_resolved_no_match_in_snapshot` requires zero exact matching record
  entries;
- `resolved_one_match_in_snapshot` requires exactly one; and
- `not_resolved_multiple_matches_in_snapshot` requires more than one.

Only `resolved_one_match_in_snapshot` can produce a positive
`SnapshotResolvedAuthority`. That refinement retains the entire resolution and
upstream snapshot witness. Its record self-binding must equal the requested
scope-plus-authority binding. The package exposes no detached registration
presence or generic authority result.

The two negative outcomes remain distinct. No match means only that the exact
reference has no record entry in the supplied finite artifact. Multiple matches
means only that this exact profile cannot select one record entry. It does not
establish semantic ambiguity, conflict, corruption, mistrust, multiple
identities, or multiple entities. Multiplicity for one binding does not prevent
another binding in the same snapshot from resolving through its unique entry.

Even the positive refinement is a Lab software relation, not an external or
ontological claim. It establishes no external existence, external completeness,
global uniqueness, canonical identity, legitimacy, standing, jurisdiction,
permission, authorization, ownership, authorship, delegation, hierarchy,
attribution scope, trust, provenance assurance, source truth, scientific truth,
maturity, implementation, applicability, or registration presence. It requires
and infers no person, agent, organism, species, body, anatomy, institution,
object, observer, event, location, geometry, dimensionality, unit system,
language, network identity, universe, source-law time, chronology, recency,
supersession, revocation, or latest revision.

Invalid or cross-scope inputs never become resolution outcomes. The candidate
performs no aliasing, normalization, dereference, retrieval, registry access,
network operation, or I/O. Its Go identifiers and corpus labels are bounded Lab
engineering tokens, not natural language or universally shared meaning. Neither
negative outcome can become registration `not_registered`; that bridge still
requires separately ratified attribution scope and lookup closure.

## Exact registration-authority binding relation candidate

The internal, non-wire experiment in `internal/registrationauthoritybinding` implements
one pairwise exact-equality rule. It consumes a positive
`SnapshotResolvedAuthority` and a positive structural `Registration`, first
requires their exact catalog-scope references to match, and then compares only
their exact opaque declaration-authority references. Catalog-scope mismatch is
an input-composition error and never a no-match.

The rule identity `lab.exact-authority-binding` at revision `v0alpha1` is closed
inside the package. Callers cannot attach arbitrary rule labels to the fixed
behavior. Revision is rule identity material, not chronology, recency, maturity,
precedence, compatibility, or a latest-version selector. A future behavior or
revision requires a separate explicit factory and conformance decision.

For valid same-scope inputs, `exact_authority_binding_match` means only that the
two existing opaque reference values are exactly equal under that rule.
`no_exact_authority_binding_match` means only that they are unequal. It does not
establish that they denote different external things. No normalization, case
folding, decoding, alias, delegation, hierarchy, inheritance, or fallback is
performed. Subject and capability remain retained in the registration but do
not participate in this comparison.

The `Decision` retains the complete rule identity, resolution witness, and
positive structural registration without storing a caller-provided outcome or
copied binding. Only equality can produce an `ExactBindingWitness`. A no-match
decision still exposes its retained positive structural registration, making it
structurally distinct from registration absence. Product registration presence
still requires separately ratified catalog closure.

This relation is deliberately not called Declaration Attribution Scope. It
does not identify an agent, entity, creator, owner, speaker, lawgiver, source,
or responsible party, and it establishes no attribution, provenance,
responsibility, legitimacy, permission, authorization, ownership, authorship,
identity equivalence, trust, source truth, scientific truth, or catalog
completeness. It requires and infers no person, organism, species, body, anatomy,
institution, object, observer, event, location, geometry, dimensionality, unit
system, language, network identity, universe, source-law time, chronology, or
external existence.

The Go-domain decision is deterministic and performs no locale, clock,
filesystem, environment, network, randomness, or I/O operation. Its identifiers
and corpus labels are Lab engineering tokens, not natural language or universal
meaning. Collection-level rule selection, rule migration, attribution-scope
ratification, closed attribution domains, conflict handling, canonical bytes,
signatures, CLI projection, and wire representation remain planned.

## Positive snapshot-registration binding composition candidate

The internal, non-wire candidate in `internal/snapshotregistrationbinding`
composes two already-produced positive witnesses. It accepts a valid
`SnapshotLookup` that exposes a positive structural registration and a valid
`ExactBindingWitness` that exposes the same complete structural registration.
It constructs no new snapshot, invokes neither upstream constructor, and emits
no new membership or authority-comparison outcome. It defensively revalidates
both retained witnesses and their shared registration.

`SnapshotMemberExactBindingWitness` retains both complete upstream witnesses.
It therefore retains the finite registration snapshot, exact lookup query,
positive structural registration, resolved-authority snapshot, pairwise rule
identity, and exact authority-binding decision. Its validity requires the two
witnesses to expose equal complete `Registration` values, not merely the same
authority reference.

The candidate defines no absent, unequal, unresolved, invalid, non-attributed,
or not-registered outcome. A valid negative lookup is a typed construction error,
as are an invalid binding witness and two positive witnesses for different
registrations. Failure to produce the positive composition has no negative
external meaning.

This candidate performs no collection selection, partition, filtering,
enumeration, Declaration Attribution Scope, or Catalog Lookup Closure. It does
not ratify product `registered` or establish attribution, provenance,
responsibility, identity equivalence, ownership, authorship, legitimacy, trust,
truth, external existence, or catalog completeness. It introduces no person,
agent, organism, species, body, anatomy, institution, object, observer, event,
source, location, geometry, dimensionality, unit system, language, network,
universe, occurrence, operation, source-law time, ambient state, I/O, CLI, or
wire contract.

## Catalog lookup closure gate

Product meaning still requires separate ratification of all three contracts
below. Internal candidates and pairwise relations do not satisfy that gate:

- **Declaration Authority Resolution:** a profile-owned, witness-retaining
  decision over exact authority-reference match evidence. The current candidate
  resolves only an exact one-match snapshot result. Zero and multiple matches
  remain distinct snapshot-relative non-resolution outcomes. No outcome implies
  legitimacy, permission, jurisdiction, trust, authorship, ownership,
  personhood, agency, attribution scope, external completeness, or source-law
  standing.
- **Declaration Attribution Scope:** a rule-relative, witness-retaining decision
  over a collection of structural registration records and one
  snapshot-resolved authority. The current pairwise experiment is only a
  precondition, and positive snapshot-registration binding composition only
  joins two witnesses for one registration. Neither implements collection
  scope. The future contract must perform no implicit alias, delegation,
  hierarchy, inheritance, ownership, authorship, or cross-authority equivalence
  inference.
- **Catalog Lookup Closure:** a finite Lab software-enumeration property over one
  exact catalog-scope revision and attribution scope. It is not a claim that
  reality, possible capabilities, or an external catalog are finite or complete.

A future query-bound closed-lookup witness must retain the exact binding,
resolved authority record, exact attribution-rule identity, complete
attribution-scope witness, lookup-contract revision, and closed membership
domain. Only zero matching records in that domain may produce product
`not_registered`. A positive result must identify its exact matching
registration record. Invalid bindings, unresolved or ambiguous authority, no
exact authority match, multiple exact authority matches, pairwise
authority-reference inequality, any separately defined semantic ambiguity,
unresolved attribution scope, open or incomplete content, unavailable content,
and conflicting registration records remain typed owner-profile outcomes or
diagnostics. They are never shared evaluation-disposition kinds and never
aliases for absence.

## Current provisional tagged variants

The Go validator currently treats each status and reason combination as a
closed pair. A slash below separates `status` from the required `reason_code`.
For an entry without a slash, the current Go emitter omits an empty reason through
`omitempty`, while typed validation treats a decoded omitted reason and explicit
empty string identically. A future strict wire parser must require the reason for
slash pairs and reject explicit empty `reason_code` for no-slash pairs.

| Axis | Current accepted pairs |
| --- | --- |
| Law definition | `candidate`; `declared`; `not-applicable/application_capability`; `not-declared/law_does_not_define`; `unknown/not_evaluated` |
| Implementation | `available`; `unavailable/not_implemented`; `unknown/not_evaluated` |
| Closure | `available`; `unavailable/closure_unavailable`; `not-required`; `undetermined/scenario_not_evaluated` |
| Applicability | `applicable`; `not-applicable/law_does_not_define`; `not-applicable/outside_validity`; `undetermined/scenario_not_evaluated` |
| Evidence | `verified/software_fixture`; `validated/empirical_reference`; `design-only/implementation_evidence_unavailable`; `none/no_evidence`; `unknown/not_evaluated` |
| Trust | `built-in-candidate`; `permitted`; `refused/forbidden_by_policy`; `untrusted/untrusted_pack`; `undetermined/operation_not_evaluated` |
| Backend feasibility | `available`; `unavailable/backend_unavailable`; `not-required/application_capability`; `not-applicable/implementation_unavailable`; `undetermined/not_evaluated` |
| Resource feasibility | `within-default-budget`; `insufficient/resource_budget_exceeded`; `refused/resource_policy_refusal`; `not-applicable/implementation_unavailable`; `undetermined/scenario_not_evaluated` |

This table documents current behavior. It is not yet the v1 vocabulary. Before
ratification, complete report fixtures must witness every retained variant and
cross-axis mutations must demonstrate for corpus fixtures that a pair from one
axis is rejected by another.

The current intended distinction is provisional and cannot be carried into the
shared disposition unchanged:

- `unknown` means the question is meaningful but the Lab has no supported
  result from the declared inputs, but current `unknown/not_evaluated` conflates
  that epistemic claim with process non-evaluation.
- `undetermined` currently means a prerequisite Lab stage has not run. It does
  not establish mathematical underdetermination or source-law time.
- `not_evaluated` explains Lab process state. It is not evidence that the
  underlying answer is absent, unknowable, inapplicable, or false.

These definitions must be tested against every retained pair before they become
normative.

## Application capabilities

`catalog.inspect` currently reports:

```text
law_definition = not-applicable / application_capability
```

This means the Lab application owns the inspection operation. The selected law
context is its subject, not its defining authority. `candidate` and `declared`
instead attribute definition to the selected law context.

Application ownership cannot force the other seven current fields. Each future
profile must evaluate only its owned question from its declared binding.

## Presentation boundary

Localized presentation remains optional and non-authoritative. Its absence is
valid. Adding, removing, translating, or reordering presentation must not change
capability identity or any assessment.

Stable IDs, member names, statuses, and reason codes are locale-invariant Lab
protocol tokens. They are not natural language, language-neutral notation, or
universally shared meaning. The current locale label is a bounded application
token, not a claim of complete BCP 47 validation.

The evaluation-disposition candidate includes no presentation. A later singular
report may reference a separately versioned presentation profile.
The current text renderer prefers an English entry and otherwise selects the
first authored presentation. That fallback is order-sensitive. A future
unordered presentation profile must replace it with an explicit fallback rule.

## Evidence boundary

The current evidence registry is deliberately software-specific:

```json
{
  "id": "test:law-catalog-inspection",
  "scope": "software",
  "kind": "go-test",
  "go_package": "./internal/lawcatalog",
  "go_test": "TestBuiltInCatalog"
}
```

This locates an executable repository fixture. It is not a signed execution
receipt, empirical reference, calibration chain, proof object, dataset citation,
or physical validation claim.

Before a singular report can be v1:

- every evidence reference must be unique and resolve exactly once;
- evidence status and reason must permit the referenced evidence class;
- `none` and `unknown` must carry no evidence references;
- design-only claims may cite only an explicitly declared design class;
- every serialized registry entry must be referenced;
- empirical, formal, calibration, dataset, and software evidence payloads must
  use separately reviewed profiles with revision and provenance fields. These
  are initial extensible classes, not a universal evidence ontology.

For each document accepted by a sound reference-integrity validator, validation
can establish that a pointer resolves to the declared class. It cannot establish
that the evidence is true, sufficient, applicable, or current.
Both `evidence_references` and evidence payloads remain outside the shared
evaluation disposition. Evidence outcomes belong to a separately reviewed claim
evidence profile.

## Ordering and encoding

Capability, evidence-reference, evidence-record, and presentation collections
are intended to be semantically unordered. A future deterministic encoder should
sort them by a ratified bytewise key after validating uniqueness. Current Go
inspection output preserves authored capability order, while the scenario probe
sorts fully resolved capability results. That mismatch must be removed before a
cross-language contract freezes.

JSON object-member order is nonsemantic. Existing SHA-256 fixtures demonstrate
repeatable bytes only for the covered executions of one Go implementation. They
are not RFC 8785 canonical JSON, scientific identity, case identity, or evidence
of source-law recurrence.

No canonical ordering key has been selected or implemented yet. Bytewise stable
ID order is a candidate that must be tested against duplicate handling,
presentation locale rules, and future cross-language encoders.

## Parser and resource gate

Before v1, the protocol must publish and enforce the same limits in prose,
machine schema, Go, and the shared conformance corpus:

- total encoded bytes and defensive nesting depth;
- machine-token grammar and byte length;
- capability, presentation, evidence-reference, and evidence-record counts;
- locale-label and display-text byte lengths;
- evidence-locator lengths.

No numerical report, collection, or parser limit has been selected yet. Any
candidate limit must be justified against denial-of-service resistance, useful
scientific reports, and implementation cost before it enters a v1 artifact.

A consumer-facing parser must reject invalid UTF-8, duplicate members, unknown
members, trailing values, wrong types, explicit nulls for required values,
missing axes, invalid tagged pairs, duplicates, and dangling references. Go's
`DisallowUnknownFields` is insufficient by itself because it does not reject
duplicate JSON members.

The current official JSON Schema release is Draft 2020-12. A future committed
schema should declare that dialect, close every object, encode each axis as its
own tagged union, and remain paired with normative prose for constraints JSON
Schema cannot express directly. No runtime network fetch or general schema
dependency is required for the producer-only Go oracle.

## Evaluation-disposition experiment

The first executable experiment in `internal/evaluation` is internal and
non-wire. Its bounded, separately stored, manually authored TSV corpus includes:

- `evaluated` with exactly one profile-owned outcome;
- `not_evaluated` with no outcome;
- invalid profile outcomes, a missing validator, the invalid zero value, a forged
  unknown kind, and a forged non-evaluated value carrying an outcome;
- no shared kind or constructor for generic `unknown` or `undetermined`;
- fixtures requiring no source time, observer, physical quantity, geometry,
  language, or workflow-stage sequence.

The corpus records Go-domain behavior only. It does not create a product schema,
reserve a discriminator, or change the current report bytes. Its implementation
and authorization examples use different fixture-owned types so the shared
candidate cannot absorb a profile vocabulary by accident.

## Report ratification corpus

The future corpus must be implementation-independent and usable unchanged by Go,
Rust, CLI, TUI, native, MCP, and A2A adapters. It includes:

- a presentation-free minimum report;
- application-owned, law-candidate, law-declared, and law-not-declared fixtures;
- a complete report witness for every accepted status and reason pair;
- compatible software, design, empirical, formal, calibration, and dataset
  evidence examples only after their payload profiles exist;
- every required-field deletion and null substitution;
- unknown-member, duplicate-member, invalid-token, invalid-pair,
  dangling-reference, duplicate-ID, unreferenced-record, and limit-plus-one
  fixtures;
- permutations asserting that collection order has no semantic effect for the
  corpus fixtures;
- exact Go bytes kept separately from decoded semantic expectations;
- deterministic, bounded, panic-free fuzz properties.

Tests generated from the same lookup table as the validator are not independent
evidence. At least one committed corpus must state expectations outside the Go
implementation.

## Ratification exit gate

The shared evaluation disposition can leave internal-candidate status only when:

1. Its input binding and profile-owned outcome boundary are explicit.
2. Positive and negative fixtures cover every constructor invariant.
3. It cannot represent generic uncertainty, applicability, refusal, or failure
   as process non-evaluation.
4. Current Go law and scenario outputs remain byte-identical.
5. No source-law ordering, observer, object, quantity, geometry, or language is
   required.

Each owner profile can become v1 only after its authority, exact inputs,
outcomes, resource bounds, conformance corpus, and nonclaims are ratified. The
singular capability report can become v1 only after the required profiles,
law-context binding, ordering, optional presentation profile, evidence payload
profiles, referential closure, machine schema, semantic prose, and exact Go
fixtures agree.

No v1 capability-assessment or capability-report discriminator is reserved.

Passing those gates would provide evidence of wire conformance. For each
document accepted by a sound validator, validation could establish internal
reference closure. It would not establish that a law, implementation, closure,
applicability result, evidence claim, policy decision, backend assessment, or
resource assessment is correct. It would not select or admit an operation,
create a case, establish cross-context compatibility, validate physics, or make
Lab tokens universally meaningful.

## Standards basis

- [JSON Schema specification](https://json-schema.org/specification), whose
  current released version is Draft 2020-12.
- [JSON Schema Draft 2020-12](https://json-schema.org/draft/2020-12), including
  the published core, validation vocabulary, and meta-schema.
- [RFC 8259](https://www.rfc-editor.org/rfc/rfc8259), the JSON data interchange
  syntax referenced by JSON Schema.
