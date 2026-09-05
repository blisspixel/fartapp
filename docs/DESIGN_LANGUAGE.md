# CLI design language

The current product surface is the command line. Its design should make an
operation's result, supporting evidence, and limits easy to find. The joke is
the subject and the permanent toy output; ordinary controls and diagnostics
should read like a well-made scientific instrument.

This is a presentation standard for incremental implementation. The current
pass applies shared numeric formatting to Go reservoir, restriction,
restriction-history, and walk reports, including refinement, and to Rust
reservoir reports and play observations. It also supplies compact root help,
contextual native play and Go walk help, and applicable recovery hints for
prediction refusals. The Go assurance view identifies its output as declared
metadata. Other surfaces remain subject to their existing contracts.

The [candidate brand](BRAND.md) supplies the broader visual direction; the
[interface contract](INTERFACES.md) owns command behavior and evidence meaning.

## One account, several presentations

Presentation reads a retained typed result. It cannot rerun the model, choose a
new closure, repair an inadmissible input, change a claim disposition, or consume
an action budget. The same rule applies to summary and complete session views.

Permanent intensity 1 through 5 output remains exact. Versioned JSON, identifiers,
input schemas, retained evidence, and numerical policies do not change to improve
the appearance of human text. A display name never replaces a semantic ID in
machine output.

## Report hierarchy

Use this reading order when the operation supplies the corresponding facts:

1. Operation and outcome.
2. Completion or stopping condition, when distinct from the outcome.
3. Inputs and results needed to understand that outcome, with explicit units.
4. Supporting checks, residuals, tolerances, and retained references.
5. Relevant limitations and one useful next step.

An endpoint prediction has no elapsed time. A registry inspection has no test
result. A read-only observation has no new physical outcome. Omit inapplicable
sections instead of filling them with zeros or empty familiar concepts.

Keep one short heading, plain labels, two-space indentation, and one blank line
between sections. Use a consistent description column for short command lists;
put long descriptions on an indented continuation line. Aim for 80 columns of
ordinary prose. Preserve copyable commands, paths, identifiers, and hashes even
when they are longer.

Do not add boxes, repeated rules, required Unicode symbols, emoji, decorative
ASCII art, or terminal-control sequences to plain output. Do not print a brand
banner before each JSON record or mix progress into machine stdout.

## Outcome vocabulary

Use the result's actual disposition, not a generic success label:

| Situation | Human meaning |
| --- | --- |
| Predicted | The requested model calculation returned its account |
| Refused | The request could not be admitted or evaluated for the stated reason |
| Truncated | A declared budget or stopping limit ended the attempted work |
| Finished | The session was explicitly closed; this does not prove scientific success |
| Integrity verified | The declared structure, binding, and digest checks passed |
| Reconstruction matched | A new calculation matched the retained comparison target under the stated profile |
| Declared metadata | References were inspected; their checks were not executed |

Do not use these phrases as replacement machine tokens. Preserve the exact code
and JSON disposition owned by the relevant report schema.

Keep accuracy tolerance satisfaction separate from complete discharge. Keep
session closure separate from truncation: an operator can finish an exhausted
session, and its truncated status remains visible. Keep arithmetic consistency,
integrity, reconstruction, empirical validity, and authenticity separate.

## Numbers and units

Human reports should use six significant digits for calculated scalar values.
Choose notation after rounding: use scientific notation for nonzero magnitudes
below `1e-3` or at least `1e6`;
trim unnecessary trailing zeros and keep zero as `0`. This gives, for example:

| Retained value | Human display |
| --- | --- |
| `102255.00000000009 Pa` | `102255 Pa` |
| `0.00000000000005684341886080802 kg` | `5.68434e-14 kg` |
| `1.1595933153081862e-19 kg` | `1.15959e-19 kg` |

The presentation boundary table is:

| Input value | Display |
| --- | --- |
| Positive or negative zero | `0` |
| `0.000999999` | `9.99999e-4` |
| `0.0009999999` | `0.001` |
| `0.001` | `0.001` |
| `-0.001` | `-0.001` |
| `123.456789` | `123.457` |
| `999999` | `999999` |
| `999999.9` | `1e6` |
| `1000000` | `1e6` |
| `-1000000` | `-1e6` |
| `1e-100` | `1e-100` |
| Smallest positive binary64 value | `4.94066e-324` |
| Largest finite binary64 value | `1.79769e308` |

Exponents use lowercase `e` without a redundant plus sign or leading zeros.
The two rounding-crosses-boundary rows are part of the shared Go and Rust
presentation tests. The executable boundary corpus is
[numbers.json](../testdata/presentation/numbers.json), consumed by the
[Go tests](../internal/cli/presentation_test.go) and
[Rust formatter tests](../crates/fart-services/src/presentation.rs).

This is display rounding, not a new uncertainty estimate or a change to the
claim's test. Never display a nonzero residual as zero. Keep counts, revisions,
codes, IDs, and digests exact. A report using rounded scientific values states
that full numeric values remain available in JSON.

Use a space between value and unit, plain SI spellings such as `Pa`, `kg`,
`m^3`, and `N s`, and a leading zero before a decimal fraction. Do not silently
switch units, insert locale-dependent numeric separators, or infer an ambient
pressure. Distinguish source total enthalpy, exit static enthalpy, and kinetic
energy where the selected model reports them.

## Help and recovery

Root help is an index of implemented command groups. It provides a short usage
line, a concise purpose for each group, a few executable starting examples, and
the route to contextual help. Avoid a single deeply nested union containing
every leaf command.

Nested help describes the requested command. Show its grammar, required inputs,
supported options, one repository fixture example, standard-input behavior,
relevant bounds, and exit meaning. Do not advertise planned commands or repeat
the entire root help for a leaf topic. Help must not read an input document or
perform an operation.

A text diagnostic answers three things: what failed, where the report locates
the problem, and what the operator can do next. Retain the stable diagnostic
code and field path, then add a concise explanation and an applicable recovery
hint. For example, a missing input can suggest checking the path or using `-`
for standard input. An incompatible law reference names the exact supported
revision. It does not quietly select that revision.

Only suggest actions supported by the current command. Escape and bound
untrusted displayed text. Keep errors on stderr according to the command's
output contract, and keep structured refusals intact when JSON is requested.
An output failure after a committed file does not undo that file; its diagnostic
should preserve the operation's documented retry advice.

## Later surfaces

Terminal and native views will reuse this hierarchy and the same account. Color
can reinforce a named state, never replace it. The candidate brand's measurement,
ledger, and failure colors are conditional presentation tokens, not evidence of
measurement, provenance, or verification by themselves.

The default remains monochrome. The existing Open Isobar palette can supply an
optional accent when a surface supports it:

| Role | Existing token | Use |
| --- | --- | --- |
| Structure and text | Instrument / Reference | Primary hierarchy, values, and ordinary controls |
| Active scientific detail | Measurement | Applicable measurement or selected flow information |
| Evidence and uncertainty | Ledger | Retained references, tolerances, and explicit uncertainty |
| Failed verification | Failure | A stated failed check or unsafe state, never decorative emphasis |

A supported refusal is not automatically a failed verification or unsafe state.
Its text must identify the actual disposition. Native typography, theme contrast,
and focus behavior need their own checks; the CLI uses the operator's terminal
font and requires no new font, image, animation, or media asset.

Motion, audio, layout, disclosure depth, and accessibility controls stay outside
scientific and journal identity. Future panes remain conditional on supported
capabilities. A more elaborate interface must not invent a source, observer,
room, timeline, or physical sound to fill its layout.

## Verification

Presentation changes need exact-output checks for permanent toy behavior and
the affected human help or report, plus machine-output checks proving that
scientific values and schema bytes remain unchanged. Numeric formatting checks
cover zero, nonzero tiny values, notation boundaries, and finite extremes.
Help tests use an input reader that fails if anything is read. Session view tests
assert unchanged revision, budgets, retained accounts, and journal fingerprints.

Fixture and process checks establish implemented behavior. They do not replace
the planned user, accessibility, localization, or cross-platform interaction
reviews in the broader product contract.
