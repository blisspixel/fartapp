# Evaluation-disposition conformance corpus

Status: internal test metadata, not a product serialization.

`v0alpha1` versions only the expectations in this bounded test corpus. The TSV
column values are harness instructions, not protocol tokens. No CLI, adapter,
archive, or external implementation may consume this file as a product
contract.

The corpus tests one shared invariant: an evaluated Lab procedure has exactly
one outcome validated by its owning profile, while a non-evaluated
procedure has no outcome. It defines no generic uncertainty, applicability,
refusal, evidence, trust, backend, resource, source-time, observer, physical,
geometric, language, or workflow-stage semantics.

The cases are manually authored expectations stored separately from the Go
constructors and profile validators. This is not independent implementation or
scientific evidence. The loader accepts at most 65,536 UTF-8 bytes, 256 data
rows, exactly five tab-separated fields per row, 128 bytes per field, and unique
case identifiers.
