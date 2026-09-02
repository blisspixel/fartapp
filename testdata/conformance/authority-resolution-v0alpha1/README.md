# Authority-resolution conformance corpus

Status: internal test metadata, not a product serialization.

`v0alpha1` versions only the manually authored Go-domain expectations in these
bounded TSV files. Column values are harness instructions, not protocol tokens.
No CLI, adapter, archive, or external implementation may consume these files as
a product contract.

The candidate consumes exact-match evidence that already retains one bounded,
immutable, in-process authority snapshot. It maps zero, one, or multiple exact
record entries to three snapshot-qualified resolution outcomes. Only exactly
one match produces a positive resolved-authority witness. Multiple entries are
not called semantic ambiguity and do not prove multiple identities or entities.

Invalid and cross-scope inputs use explicit `not_applicable` sentinels rather
than encoding a resolution outcome, count, or positive witness. No negative
outcome is registration absence or evidence of external nonexistence.

Resolution does not establish external completeness, global identity,
attribution, legitimacy, trust, authorship, ownership, personhood, agency,
organism, species, body, anatomy, object, location, geometry, dimensionality,
unit system, language, observer, universe, source-law time, chronology, network
identity, source truth, scientific truth, or registration presence.

The cases are manually authored and stored separately from the Go constructors.
They are not independent implementation, catalog, authority, identity, or
scientific evidence. Each file accepts at most 65,536 UTF-8 bytes, 256 data rows,
its exact tab-separated field count, 128 bytes per field, and unique nonempty
fixture or case identifiers where applicable.
