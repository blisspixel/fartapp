# Catalog-registration conformance corpus

Status: internal test metadata, not a product serialization.

`v0alpha1` versions only the manually authored Go-domain expectations in this
bounded TSV corpus. Column values are harness instructions, not protocol tokens.
No CLI, adapter, archive, or external implementation may consume this file as a
product contract.

The experiment constructs structural result values for one exact binding over a
catalog-scope reference, declaration-authority reference, subject reference, and
capability token. A caller may construct an evaluated registered result, an
evaluated not-registered result, or a not-evaluated result. The experiment does
not perform a lookup, prove that a catalog is closed, or verify the caller's
claim. Only the registered result can produce a structural registration value.

Registration does not establish declaration-authority standing, source-law
truth, scientific validity, implementation, applicability, evidence,
authorization, backend feasibility, resources, maturity, occurrence, or
operation status. Within this structural representation, `not_registered`
addresses only the exact binding. Revision is identity material, not source-law
time or an ordering claim.

The cases are manually authored and stored separately from the Go constructors.
This is not independent implementation or scientific evidence. The loader
accepts at most 65,536 UTF-8 bytes, 256 data rows, exactly ten tab-separated
fields per row, 128 bytes per field, and unique case identifiers.

The internal candidate token grammar accepts 1 through 128 ASCII bytes. The
first and last bytes must be lowercase letters or digits. Interior bytes may
also use `.`, `_`, or `-`, but separators may not be adjacent. Separators carry
no structural meaning. This grammar is not a product identifier contract.
