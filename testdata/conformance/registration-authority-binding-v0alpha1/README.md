# Registration-authority binding conformance corpus

Status: internal test metadata, not a product serialization.

`v0alpha1` versions only the manually authored Go-domain expectations in these
bounded TSV files. Column values are harness instructions, not protocol tokens.
No CLI, adapter, archive, or external implementation may consume these files as
a product contract.

The experiment compares one snapshot-resolved opaque authority reference with
the opaque declaration-authority reference retained by one positive structural
registration. The package-owned rule uses exact equality only. Catalog-scope
mismatch is an error. Same-scope authority-reference mismatch is a valid
pairwise no-match. Only equality produces an `ExactBindingWitness`.

The relation is not Declaration Attribution Scope. It does not identify an
agent, entity, creator, owner, speaker, lawgiver, source, or responsible party.
It performs no normalization, alias, delegation, hierarchy, inheritance,
provenance, trust, truth, time, network, or catalog-closure inference. Unequal
opaque references are not evidence that different external things exist.

Error and upstream non-resolution cases use explicit `not_applicable` sentinels
instead of encoding a comparison outcome or positive witness. Every comparison
input is already a positive structural `Registration`; a no-match can never be
registration absence.

The cases are manually authored and stored separately from the Go constructors.
They are not independent implementation, catalog, authority, identity,
attribution, provenance, or scientific evidence. Each file accepts at most
65,536 UTF-8 bytes, 256 data rows, its exact tab-separated field count, 128 bytes
per field, and unique nonempty fixture or case identifiers where applicable.
