# Experimental scenario probe

The v0.7 scenario probe is a deliberately narrow, read-only bridge between law
catalog inspection and a future ratified scenario contract. It lets the CLI
answer one useful question now:

> Is this finite capability request structurally valid under one exact built-in
> law-context revision, and what does the catalog currently claim about every
> requested capability axis?

It is not a simulation command. It does not apply defaults, allocate a record
nonce, create context-occurrence identity claims, compute scenario identity,
inspect host resources, evaluate execution admission, or invoke a solver.

## Command

```console
fartapp scenario validate <scenario.json|-> [--format text|json]
```

`-` reads standard input. JSON is the only accepted authoring form in this
experimental slice. The implementation combines Go's standard JSON decoder
with a small bounded token preflight for duplicate members, nesting, and exact
Unicode checks. Adding a TOML dependency or inventing a separate partial value
parser is not justified for this envelope.

Input is limited to 65,536 bytes. The parser rejects empty input, invalid JSON,
trailing values, duplicate object members, unknown or overlong member names,
excessive nesting, wrong types, invalid machine tokens, unresolved scope
references, repeated capability requests, unknown exact law revisions, and
unknown capabilities.
Diagnostics have stable codes, stages, JSON Pointer paths, reason codes, and
bounded optional byte offsets.

A named path is opened once and read through that same handle. Paths otherwise
retain operating-system semantics, including symlinks, devices, FIFOs, and
network-backed files where the host supports them. The byte limit is not a read
deadline. Standard input and special or remote files can therefore wait on their
providers; the probe performs no hidden timeout or background read.

Portable named-file use is identical across supported systems:

```console
fartapp scenario validate scenario.json --format json
```

Standard input is explicit. POSIX shells can redirect it:

```console
fartapp scenario validate - --format json < scenario.json
```

PowerShell can pipe the same document:

```powershell
Get-Content -Raw scenario.json | fartapp scenario validate - --format json
```

## Candidate document

The wire schema is explicitly provisional:
`fart.scenario-probe/v0alpha1`.

```json
{
  "schema": "fart.scenario-probe/v0alpha1",
  "law_context_set": {
    "contexts": [
      {
        "id": "conformance.relation.atemporal",
        "version": "v0alpha1",
        "scope_id": "s0"
      }
    ]
  },
  "scope": {
    "id": "s0"
  },
  "capability_requests": [
    {
      "id": "catalog.inspect"
    }
  ]
}
```

The scope identifier is only an addressable application boundary. It does not
imply a participant, object, proposition, region, source, observer, dimension,
geometry, state, clock, recurrence relation, or ordering.

Exactly one law context is accepted in this revision. Accepting several bare
contexts before scope assignment and `InterLawCoupling` contracts are ratified
would imply compatibility the Lab cannot justify. Multi-law input therefore
returns `multi_law_not_supported`.

`capability_requests` is logically unordered after successful resolution.
Validation rejects duplicate IDs and sorts a fully resolved result set by stable
capability ID. For fully valid documents, request order, object-member order,
and insignificant whitespace do not change the typed report. Failure reports
retain source position where relevant: JSON Pointer array indices and optional
byte offsets may change when invalid input is reordered or reformatted. Neither
behavior creates a law ordering.

The document contains no localized presentation, quantities, units, seed,
time, source, emitter, interface, exterior, body, observer, geometry,
measurement, backend, resource budget, or solver configuration. Later schemas
may add these only through context-owned, capability-checked contracts.

## Minimal opaque conformance fixture

The narrower negative-space fixture is executable through both catalog and
scenario JSON surfaces:

```console
fartapp law inspect conformance.opaque.minimal@v0alpha1 --format json
fartapp scenario validate testdata/scenarios/minimal-opaque-probe.json --format json
```

The exact catalog entry declares no localized presentation object, optional
structural module, extension role, or domain capability. It requests only the
Lab-level `catalog.inspect` capability. Its minimal scenario selects no case
operation and creates no case record.

Only the dedicated JSON outputs are conformance evidence for absence of those
optional fields. Text output remains a current-English presentation. Stable
IDs, reason codes, and JSON member names are locale-invariant engineering
tokens; that does not make them natural language, language-neutral notation, or
universally shared meaning. The fixture proves a narrow software property, not
that any represented reality is structureless.

## Unresolved capability-reference outer envelope

The next exact negative fixture is:

```console
fartapp scenario validate testdata/scenarios/minimal-opaque-unresolved-capability.json --format json
```

Its syntax and schema stages are valid. The exact
`conformance.opaque.minimal@v0alpha1` catalog entry resolves, but the requested
opaque token `c0` is not a capability in that entry. Capability resolution ends
as `unresolved` with reason `capability_not_defined` and diagnostic
`FART-E-CAP-0001` at `/capability_requests/0/id`. Overall `document_status` is
`invalid`.

The report consults exactly `document_bytes` and `built_in_law_catalog`, with no
ambient inputs. It omits the root members `document_schema`, `law_context`,
`scope`, `capabilities`, and `evidence_registry`; no partial capability or
evidence object is fabricated. Because the schema is already known to contain
no operation field, selection is `not-declared`, while admission and execution
are `not-applicable`. This differs from a schema-stage failure, where all three
operation assessments remain `not-evaluated`.

This is capability-reference resolution evidence at the outer envelope. It is
not a trust, policy, resource, mapping, admission, execution, ontology, or
physics refusal. The report does not echo the unresolved token or resolved
context, so it remains dependent on its input document and is not a retained,
self-contained audit record.

## Multi-law probe limit fixture

The executable negative boundary is:

```console
fartapp scenario validate testdata/scenarios/multi-law-probe-limit.json --format json
```

It supplies two authored context entries naming version strings under the
current one-entry probe schema. The validator returns `FART-E-ONTOLOGY-0001`
with reason `multi_law_not_supported` at `/law_context_set/contexts` before
parsing either entry or reading the built-in catalog. Its consulted input set is
exactly `document_bytes`, its ambient-input set is empty, and its report omits
the root members `document_schema`, `law_context`, `scope`, `capabilities`, and
`evidence_registry`. Reversing the two context entries produces the same report
bytes, so the probe cannot silently select the first entry.

This fixture proves only noninterpretation at the provisional schema boundary.
It does not establish that either context applies, that they are compatible or
incompatible, that a bridge exists or is missing, or that either one defines
observers, recurrence, conservation, dimensionality, language, biology, time,
or spacetime. It is not the planned `inter_law_bridge_missing` counterexample,
because this schema cannot represent a bridge. Operation selection, admission,
and execution remain `not-evaluated` with `prior_stage_failed`; the report
creates no case and invokes no solver.

## Validation report

For fixed executable and catalog contents, the deterministic report schema is
`fart.scenario-validation/v0alpha2`. It separates:

- `document_status`, which answers whether syntax, schema, references, and
  capability names are valid.
- `validation_stages`, which reports syntax, schema, exact law resolution, and
  exact capability resolution independently. Every later stage remains
  `not-evaluated` after an earlier failure.
- `requested_case_operation.selection`, which is `not-declared` because this probe
  schema has no operation field.
- `requested_case_operation.admission` and
  `requested_case_operation.execution`, which
  are `not-applicable` because no operation was declared. This does not assume
  that the selected law context defines realization or source-law occurrence.
  Before the document schema is valid, all three operation assessments remain
  `not-evaluated` with reason `prior_stage_failed`; the validator never infers
  absence from unreadable or malformed input.
- All eight catalog axes for each requested capability, without collapsing
  them into one runnable Boolean.
- Capability `resolution: resolved`, which means only that the exact requested
  capability ID exists in the exact selected catalog entry. It does not replace
  or imply the independent `law_definition` result.
- `validation_environment`, which lists the inputs the validator consulted and
  an explicit empty ambient-input set.

Before document bytes exist, input failures distinguish `named_input` from
`standard_input` without copying a source path into the report. After a bounded
read succeeds, validation reports `document_bytes` and, only when resolution is
reached, `built_in_law_catalog`.

A request for `flow.subsonic` under
`earth.continuum.si@v0alpha1` is therefore a valid probe. Its capability result
still says implementation unavailable, closure and applicability undetermined,
evidence design-only, trust undetermined, and backend and resource feasibility
not applicable. Valid input is not a solver claim.

The validator has no clock, random, locale, environment, hostname, terminal,
network, or Earth-default provider. Its pure validation function receives only
document bytes and reads the immutable built-in law catalog during exact law and
capability resolution. Reports expose that read set. This is software and
schema-conformance evidence, not a metaphysical claim or physical validation.

## Diagnostic registry

| Code | Stage | Reason |
| --- | --- | --- |
| `FART-E-IO-0001` | input | `input_not_found`, `input_permission_denied`, or `input_unavailable` |
| `FART-E-INPUT-0001` | input | `input_too_large` |
| `FART-E-SYNTAX-0001` | syntax | `empty_input` or `malformed_json` |
| `FART-E-SYNTAX-0002` | syntax | `trailing_json_value` |
| `FART-E-SCHEMA-0001` | schema | `unsupported_schema` |
| `FART-E-SCHEMA-0002` | schema | `duplicate_member` |
| `FART-E-SCHEMA-0003` | schema | `unknown_member` |
| `FART-E-SCHEMA-0004` | schema | required collection or member missing |
| `FART-E-SCHEMA-0005` | schema | `wrong_type` |
| `FART-E-SCHEMA-0006` | schema | `invalid_token` |
| `FART-E-SCHEMA-0007` | schema | collection, nesting, or member-name limit exceeded |
| `FART-E-SCHEMA-0008` | schema | `scope_reference_unresolved` |
| `FART-E-SCHEMA-0009` | schema | `duplicate_capability_request` |
| `FART-E-ONTOLOGY-0001` | schema | `multi_law_not_supported` |
| `FART-E-LAW-0001` | law-resolution | `law_context_not_found` |
| `FART-E-CAP-0001` | capability-resolution | `capability_not_defined` |

The experimental CLI retains exit status 0 for a valid probe or help and 1 for
a controlled failure until the global exit-code RFC is ratified. A native Unix
`SIGPIPE` retains the operating system's pipeline status if it terminates an fd 1
write before Go can return `EPIPE`; detected broken-pipe errors end quietly. JSON
format returns a typed validation report even for invalid input. Text format
sends a concise diagnostic to standard error. Successful result data never
shares a stream with a diagnostic.

## Deliberate boundaries

The full scenario contract remains blocked on ratification of:

- Capability dependencies, context-owned input and output schemas, and actual
  implementation read sets.
- Multi-law scope assignment and inter-law coupling.
- Measurement interactions and their possible back-action.
- Canonicalization, scenario identity, optional seeds, record nonces,
  provenance, archives, trust aggregation, backend inventory, and resource
  admission.
- Law-specific physical inputs, closure evaluation, applicability evaluation,
  numerical policy, and solver evidence.

The probe will not grow around those decisions. A later schema replaces it only
after the contracts and counterexample suite are ready.
