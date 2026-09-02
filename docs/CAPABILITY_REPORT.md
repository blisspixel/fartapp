# Capability report ratification candidate

Status: design candidate, not ratified.

This document records the gate for a future singular capability-report wire
contract. The current Go oracle emits
`fart.law-capability-report/v0alpha1`. That token remains provisional. Neither
this document nor the current output is a stable cross-language contract.

The intended report answers eight separately reported, non-collapsing questions
about named capabilities under one exact catalog entry. It is not a scenario,
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

The roadmap therefore separates the assessment algebra from the surrounding
report. The eight-axis core is the first ratification candidate. Presentation,
evidence payloads, singular catalog binding, scenarios, and operations remain
separate versioned contracts.

## Candidate assessment core

A reported capability has one machine identifier and exactly eight required
assessment objects:

1. `law_definition`: whether the selected law context defines the capability.
2. `implementation`: whether the Lab has an implementation.
3. `closure`: whether the declared formal system contains sufficient relations
   and conditions to determine the capability, including governing equations,
   constitutive or closure relations, and applicable boundary, initial, and
   interaction conditions where those concepts are meaningful.
4. `applicability`: whether the capability applies to the declared subject and
   inputs, if applicability has been evaluated.
5. `evidence`: what kind of support exists for the exact claim.
6. `trust`: whether provenance and policy permit the requested use.
7. `backend_feasibility`: whether a selected execution backend can meet the
   declared compute contract, where a backend is required.
8. `resource_feasibility`: whether an identified resource budget can support
   the requested evaluation or use.

No axis may be collapsed into or inferred solely from another. Any cross-axis
consistency constraint must be explicit and separately validated. A law can
define a capability with no implementation. An implementation can exist without
closure for a case. A backend can be available while policy refuses an
operation. An application-level operation can be outside source-law definition
while still needing trust, backend, and budget evaluation.

Case, measurement, operation, backend, policy, and budget inputs belong to a
future enclosing evaluation contract. The assessment core can report their
outcomes but cannot by itself prove that those inputs existed or were evaluated.
Missing authored input values are a separate validation concern. Their absence
must not be reported as failure of scientific or formal closure.

Omission from a capability collection means only “not reported.” It never means
`law_does_not_define`, impossible, incompatible, refused, or unavailable.

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

The intended distinction is also provisional:

- `unknown` means the question is meaningful but the Lab has no supported
  result from the declared inputs.
- `undetermined` means the question cannot be answered until a named later
  evaluation, such as scenario applicability, has occurred.
- `not_evaluated` is a reason explaining process state, not evidence that the
  underlying answer is absent or unknowable.

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

Application ownership cannot force the other seven answers. Future Lab
operations may still require closure, evidence, trust, a backend, or an explicit
budget.

## Presentation boundary

Localized presentation remains optional and non-authoritative. Its absence is
valid. Adding, removing, translating, or reordering presentation must not change
capability identity or any assessment.

Stable IDs, member names, statuses, and reason codes are locale-invariant Lab
protocol tokens. They are not natural language, language-neutral notation, or
universally shared meaning. The current locale label is a bounded application
token, not a claim of complete BCP 47 validation.

The first assessment-core schema should not include presentation. A later
singular report may reference a separately versioned presentation profile.
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
Both `evidence_references` and evidence payloads remain outside the assessment
core. The core retains only the tagged evidence assessment.

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

## Ratification corpus

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

The assessment core can become v1 only when:

1. Every term and tagged variant has one unambiguous definition.
2. The identifier grammar and every resource bound are explicit.
3. A bounded strict parser and independent positive and negative corpus agree.
4. Current Go law and scenario projections conform without hidden defaults.
5. Presentation and evidence payloads remain outside the core.

The singular capability report can become v1 only after the assessment core,
law-context binding, ordering, optional presentation profile, evidence payload
profiles, referential closure, machine schema, semantic prose, and exact Go
fixtures agree.

`fart.capability-assessment/v1` and `fart.law-capability-report/v1` are working
candidate discriminators only. They are inactive and unreserved until their
normative artifacts and conformance suites land together.

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
