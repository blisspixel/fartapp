# Assurance registry

The Go CLI can inspect the repository's declared invariant obligations and their
evidence references without a source checkout, input document or test run:

```sh
fartapp assurance list
fartapp assurance inspect PHY-004
fartapp assurance inspect ONT-001 --format json
```

From a source checkout, use `go run ./cmd/fartapp` in place of `fartapp`.
The [generated invariant reference](INVARIANTS.md) contains the same records.
Their single authored source is
[internal/assurance/registry.json](../internal/assurance/registry.json), embedded
in the Go executable. A previously built executable describes the registry from
its build, so rebuild after changing the source.

## What inspection means

Inspection reports metadata. It does not execute a check, determine whether an
invariant applies to a case, or issue a passing result. JSON always exposes
`evidence_status: "not-executed"` and
`applicability_status: "not-evaluated"`. Text states that distinction once before
the records. Both formats use stable ID ordering and expose the same declared
scope, tolerances, references and open work.

The initial registry retains all 34 identifiers from the former
[quality invariant table](QUALITY.md#invariant-registry): 23 executable
candidates with narrowly scoped Go checks and 11 design candidates. Existing
experimental replay, service or archive work does not silently promote a broader
planned product invariant. For example, `ID-001` remains planned even though
the candidate walk evidence carrier already distinguishes retained replay from
reconstruction.

Invariant IDs and [verification benchmark IDs](VERIFICATION.md) occupy separate
namespaces. The invariant `ONT-001` covers Go catalog inspection and provisional
scenario probing. The verification benchmark `ONT-001` requires a minimum
bounded admitted case with archive and certificate round trips. Its relationship
is explicitly `partial-support`; matching ID text is not equivalence. Likewise,
the two-context schema limit in `SCN-003` supplies no semantic no-bridge proof for
verification benchmark `ONT-003`.

## Provisional metadata contract

The source schema is `fart.assurance-registry/v0alpha1`. Inspection uses
`fart.assurance-list/v0alpha1` or `fart.assurance-inspection/v0alpha1`. These are
internal candidate metadata contracts with no public stability or ratification
promise. They contain no command language.

| Field | Meaning |
| --- | --- |
| `id` | Original stable invariant identifier, such as `PHY-004` |
| `statement` | The obligation being tracked |
| `owner` | Declared repository responsibility role, such as `go-continuum-oracles`; not a person, approving authority or law declaration |
| `applicability` | Authored scope and exclusions of the registered evidence; not an evaluated case result |
| `tolerance` | Named comparison profile and its explicit description; repeated profile IDs must have identical definitions |
| `lifecycle` | `design-candidate` or `executable-candidate` under the rules below |
| `checks` | Exact Go package, source file and test or fuzz declaration, with a stable check reference ID |
| `evidence` | Repository files and descriptions of their declared evidence or owning design |
| `counterexamples` | Files containing named negative or boundary evidence; empty for planned rows |
| `milestone` | The next relevant planned release group in the roadmap; not a completion claim |
| `direction` | Remaining work before the broader obligation can advance |
| `related_benchmarks` | Separately namespaced verification references with `partial-support` or `planned-conformance` and a scope description |

A design candidate must have an owning design reference, explicit empty check
and counterexample arrays, and the `planned-v0alpha1` tolerance profile. An
executable candidate must have declared checks, counterexamples and a comparison
profile. Both retain scope and open work. These are source-validation rules,
not proof that the implementation satisfies its statement.

This revision refuses every other lifecycle value, including `passing`,
`ratified-internal` and `stable-public`. Promotion into the broader
[contract lifecycle](QUALITY.md#contract-lifecycle-and-evidence-debt) requires a
separately reviewed schema and its required decision evidence. Metadata cannot
substitute for owner review. Changing a design row to an executable candidate
still requires positive and negative evidence under the declared scope, followed
by normal code review and execution of the relevant checks. No history of
approvals or lifecycle transitions is invented by this snapshot.

Scientific profiles retain fixture-specific units and comparisons. For example,
`PHY-004` distinguishes legacy fixed-step convergence from the opt-in adaptive
clock's estimated quadrature error. Neither profile claims a universal physical
validation or a rigorous error bound across every accepted finite input. The
[blowdown reference](BLOWDOWN_REFERENCE.md) derives the numerical scope.

## Source validation and maintenance

Repository policy runs the assurance gate automatically. It can also be run on
its own:

```sh
go run ./tools/repoquality assurance
go run ./tools/repoquality repository
```

The gate checks exact source references without running registry-supplied code:

- The JSON parser rejects unknown or duplicate members, malformed Unicode,
  explicit nulls, omitted required collections, invalid identifiers and paths,
  conflicting profile or check definitions, and invalid lifecycle claims.
- Every evidence and counterexample path must identify a bounded regular file
  inside the repository. Paths must be canonical portable relative identities;
  external paths, traversal and symlink aliases are refused.
- Each declared Go check must name a nonempty top-level `Test...` or `Fuzz...`
  function in an actual `_test.go` file, with exactly one `*testing.T` or
  `*testing.F` parameter, respectively. Named aliases of the standard `testing`
  import work; methods, function-valued variables, fake imports, duplicate
  declarations, generic functions and wrong signatures do not.
- Registered checks must participate in Windows, macOS and Linux amd64 Go builds
  without cgo or custom tags. Ignored filenames, platform-only suffixes and
  excluded or malformed build constraints are refused in this revision.
- Separate benchmark IDs must resolve to exactly one benchmark table row in
  `VERIFICATION.md`. An invariant with the same ID cannot satisfy that reference.
- Each declared milestone must resolve to one release heading in `ROADMAP.md`.
- The generated document must match the canonical registry. Only Git worktree
  CRLF versus LF line endings are normalized; all other content is exact.

These checks establish metadata integrity and source declaration existence.
They do not compile the referenced package, detect every semantic weakness of a
test body, or prove that a runtime test is applicable or unskipped. The normal
test, race, fuzz, coverage and CI gates supply separate execution evidence. This
revision registers executable Go checks only. A reference to Rust documentation
does not imply Rust execution by the assurance checker.

Encoded input is limited to 1 MiB, 256 invariants, 32 entries per reference
collection, 12 JSON nesting levels and 64-byte member names. Descriptions are
limited to 4096 bytes and reject control and format characters; source paths
are at most 240 ASCII bytes. The checker reads at most 512 distinct source files,
4 MiB per source file and 32 MiB total. It loads the registry itself under the
stricter 1 MiB limit. No remote source is fetched.

To change an obligation or add evidence, edit the canonical registry, update the
declared tests or design files, and regenerate the reference:

```sh
go run ./tools/repoquality assurance --write
go test ./internal/assurance ./internal/repoquality
go run ./tools/repoquality repository
```

`--write` first validates every declared reference, then replaces only
`docs/INVARIANTS.md` using a temporary file in the same directory. Ordinary
inspection and repository checking never regenerate it. Review both the
authored metadata diff and the generated reference before merging. Run the
affected executable checks through the repository's normal test commands;
referential validity alone does not establish a passing implementation.

The parser tests preserve the original IDs, planned statuses, immutable access,
separate benchmark relationships and deterministic output. Checker tests cover
false declarations, excluded builds, malformed sources, missing references,
drift, source bounds and link escapes, including panic-bearing source fixtures
that must never execute during metadata inspection.
