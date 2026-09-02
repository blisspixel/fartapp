# Finite catalog-lookup conformance corpus

Status: internal test metadata, not a product serialization.

`v0alpha1` versions only the manually authored Go-domain expectations in these
bounded TSV files. Column values are harness instructions, not protocol tokens.
No CLI, adapter, archive, or external implementation may consume these files as
a product contract.

The experiment records only that an exact binding was found or not found among
every positive registration stored in one immutable, bounded, in-process
snapshot. It does not prove that the caller supplied every record from an
external catalog, that a reference resolves, that an authority has standing, or
that any registration is true. A mismatched catalog scope is an error and never
becomes a negative lookup result.

The cases are manually authored and stored separately from the Go constructors.
They are not independent implementation, catalog, authority, or scientific
evidence. Each file accepts at most 65,536 UTF-8 bytes, 256 data rows, its exact
tab-separated field count, 128 bytes per field, and unique nonempty fixture or
case identifiers where applicable.
