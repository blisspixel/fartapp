# Experimental walk evidence carrier

The Go CLI retains one authored walk request and its complete successful witness
report in a `.fartevidence` file. It provides working storage and reconstruction
while the broader `.fart` identity and certification contracts remain open.

```console
go run ./cmd/fartapp evidence capture testdata/walk/ordinary-low-pressure.json --output ordinary.fartevidence
go run ./cmd/fartapp evidence inspect ordinary.fartevidence
go run ./cmd/fartapp evidence verify ordinary.fartevidence --format json
go run ./cmd/fartapp evidence replay ordinary.fartevidence
go run ./cmd/fartapp evidence reconstruct ordinary.fartevidence --format json
```

Replace an input filename with `-` for standard input. Capture requires an
explicit destination in an existing directory. Other operations never write an
archive. Replay emits the retained compact JSON report plus one newline and
accepts only JSON output.

## Operation evidence

| Operation | Work and evidence |
| --- | --- |
| `capture` | Calculates one legacy-method witness, validates it, and publishes a new file |
| `inspect` | Checks the carrier and reports hashes, binding, and evidence limits |
| `verify` | Checks integrity without time integration |
| `replay` | Checks integrity and returns the retained report without time integration |
| `reconstruct` | Checks the carrier, calculates again, and compares against the retained witness |

Verification checks lengths, SHA-256 hashes, complete report structure,
normalized authored-request binding, and the account's own witness. Initial
model-boundary validation is permitted; verification never steps the solver.
Retained runtime metadata stays intact.

Unkeyed hashes establish internal consistency. Anyone able to replace the
account and hashes can forge an internally consistent carrier. Tests demonstrate
a forged account that verifies and fails explicit reconstruction. Neither
success proves authenticity, a real occurrence, empirical validity, recurrence,
or approval by a certificate authority.

Reconstruction uses this compiler, platform, architecture, model, and numerical
implementation. A changed profile or calculation can produce an honest
`witness_mismatch`. The report retains both witnesses and the complete new
calculation, returns status 1, and leaves the carrier unchanged.

## Bounded provisional format

The `fart.walk-evidence/v0alpha1` envelope is uncompressed JSON with exactly
`schema`, `request`, and `report`. Each member has `byte_length`, lowercase
`sha256`, and canonical standard `base64`. Bytes stay exact. No member supplies
a filename, path, link, compression method, or extraction instruction.

| Boundary | Limit |
| --- | ---: |
| Carrier bytes | 24 MiB |
| Authored request bytes | 64 KiB |
| Retained report bytes | 16 MiB |
| History samples | 4097 |
| Components | 64 |
| Envelope JSON depth | 8 |
| Report JSON depth | 32 |
| Member-name bytes | 128 |

Typed envelope decoding precedes general shape inspection to avoid expanding an
illegal large array. Report collections are bounded before typed allocation.
Duplicate members, unknown or wrong-case keys, nulls, malformed Unicode,
noncanonical base64, digest mismatches, and request substitution are refused.
Valid numerical reports above the storage limit cannot be captured.

New captures use the compact Go JSON witness profile
`go-oracle.walk/v0alpha4`, with the legacy mass-step method. Revision 4 records
the corrected restriction temperature and small-Mach arithmetic; revision 3
recorded the corrected scaled reservoir arithmetic. Integrity inspection and
replay explicitly accept both revisions 3 and 4 under the unchanged report
structure. They retain the original bytes and implementation-bound witness.
Reconstruction uses revision 4 and therefore reports a mismatch against a
revision 3 witness even when the numerical values agree. Preserve that retained
evidence; use its matching source release to repeat the earlier implementation.
The separate `walk refine` method is never silently substituted. Other
implementation revisions remain unsupported.
This is not canonical scientific identity, RFC 8785, a migration mechanism,
or the planned certified `.fart` format.

## Publication and interruption

The writer anchors the caller-selected parent directory, creates an exclusive
private staging file, writes bounded chunks, syncs and closes it, then publishes
with an atomic hard link that cannot replace an existing name. Competing writers
yield one winner. Existing files, directories, and symlinks are preserved.
Filesystems without hard-link support fail explicitly.

The library checks a cancellation context during staging and before publication.
Cancellation observed before publication prevents the destination commit and
attempts to remove its staging file. Failed cleanup can leave that private file.
Cancellation racing with the final link can still yield successful publication;
a completed commit remains valid. Fault tests cover short writes, write and sync
failures, cancellation, and competing writers.

The CLI currently uses a background context. Abrupt termination can leave a
private staging file. This implementation claims neither directory-sync nor
power-loss durability. Stdout can fail after a successful commit; inspect the
destination before retrying. The no-clobber rule still applies.

## Next archive gates

Ratify scenario identity, canonical records, measurement interaction, journals,
optional ontology structures, and migrations before expanding the storage
contract. This carrier supplies tested boundaries and actual replay for that
work; it does not complete the [v0.9 milestone](../ROADMAP.md).
