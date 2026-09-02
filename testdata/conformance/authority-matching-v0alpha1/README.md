# Authority-matching conformance corpus

Status: internal test metadata, not a product serialization.

`v0alpha1` versions only the manually authored Go-domain expectations in these
bounded TSV files. Column values are harness instructions, not protocol tokens.
No CLI, adapter, archive, or external implementation may consume these files as
a product contract.

The experiment counts exact scope-plus-authority self-bindings inside one
immutable, bounded, in-process multiset. Cardinalities mean exactly one, zero,
or multiple matches in that supplied snapshot. Cross-scope records and match
bindings are errors rather than cardinalities. Duplicate records for another
binding in the same snapshot do not prevent a unique target from producing a
one-match witness. Error cases use an explicit `not_applicable` sentinel rather
than encoding a match count.

Matching does not establish external catalog completeness, external reference
existence, attribution, legitimacy, trust, authorship, ownership, personhood,
agency, hierarchy, network access, source-law time, chronology, or truth.
Declaration-authority records are explicit corpus inputs and are never inferred
from registrations.

The cases are manually authored and stored separately from the Go constructors.
They are not independent implementation, catalog, authority, identity, or
scientific evidence. Each file accepts at most 65,536 UTF-8 bytes, 256 data rows,
its exact tab-separated field count, 128 bytes per field, and unique nonempty
fixture or case identifiers where applicable.
