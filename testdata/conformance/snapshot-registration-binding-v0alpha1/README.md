# Snapshot-registration binding conformance corpus

Status: internal test metadata, not a product serialization.

`v0alpha1` versions only the manually authored Go-domain expectations in this
bounded TSV file. Column values are harness instructions, not protocol tokens.
No CLI, adapter, archive, or external implementation may consume this file as a
product contract.

The candidate composes a positive finite-snapshot membership witness and a
positive exact registration-authority binding witness only when both retain the
same complete structural registration. It invokes neither upstream constructor
and creates no new membership, comparison, or negative result. It defensively
revalidates both witnesses. Absent or mismatched inputs are typed errors with no
composition witness.

The composition is not attribution, provenance, responsibility, identity,
trust, catalog completeness, Catalog Lookup Closure, or product registration
status. Failure to compose has no external meaning. The cases are manually
authored and are not independent implementation, catalog, authority, identity,
attribution, provenance, or scientific evidence.

The file accepts at most 65,536 UTF-8 bytes, 256 data rows, its exact
tab-separated field count, 128 bytes per field, and unique nonempty case
identifiers.
