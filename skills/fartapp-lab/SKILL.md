---
name: "fartapp-lab"
description: "Use F.A.R.T. Lab to inspect declared metadata, calculate explicit SI reservoir and restriction cases, or run a bounded native reservoir play session through the local CLI."
---

# F.A.R.T. Lab CLI

Run from the package root containing `plugin.json`, `go.mod`, and `testdata/`,
two directories above this skill. Use a matching `fartapp` executable, or
`go run ./cmd/fartapp` with Go 1.27.1 or later from that root. If neither is available,
report the missing prerequisite. Package installation does not supply a shell
or an executable, and the skill body has no automatic path-variable expansion.
Calculations need no account, network service, or model API.

Read [recipes.json](recipes.json) for tested Go argument arrays and representative
results. Arguments exclude the executable and are passed separately. Paths are
relative to the package root; `input_argument` is the zero-based fixture argument
to replace with the user's file. Replace that argument with `-` to send the same
JSON bytes on standard input. Keep user inputs within the 65,536-byte limit.

For example, using the source package:

```console
go run ./cmd/fartapp law list --format json
go run ./cmd/fartapp assurance inspect PHY-001 --format json
go run ./cmd/fartapp scenario validate testdata/scenarios/atemporal-probe.json --format json
go run ./cmd/fartapp restriction predict testdata/restriction/gamma15-choked.json --format json
go run ./cmd/fartapp walk explain testdata/walk/ordinary-low-pressure.json --format json
```

Choose the operation that answers the request:

- `assurance list` and `assurance inspect` expose declared invariant metadata and
  check references. They execute no test or solver, read no evidence files, and
  establish neither a passing check nor applicability to a supplied case.
- `law list` and `law inspect` expose declared contexts and separate capability
  statuses. An available catalog inspection is not evidence of an available
  solver. Scenario validation checks a probe; it does not admit or execute a case.
- `reservoir predict` calculates a specified finite-reservoir withdrawal.
  `restriction predict` calculates one converging-restriction state.
  `restriction history` integrates at most 256 samples with frozen stagnation.
- `walk predict` reports an endpoint; `walk explain` reports model assumptions;
  `walk simulate` and `walk inspect` expose a coupled numerical account.
  `walk branch` compares one prescribed-area variant. Numerical accuracy requires
  step refinement; arithmetic residuals alone do not establish convergence.
- `walk certify` checks arithmetic balances. `walk witness` hashes the versioned
  software account. To reconstruct, retain that digest, supply it as
  `expected_witness` in the same case, and run `walk reconstruct`. A mismatch is
  a failed comparison. These operations create no archive or certificate authority.
- `walk refine` accepts explicit `--relative-tolerance` and `--max-evaluations`
  and reports numerical quadrature estimates. Inspect `tolerance_satisfied` and
  `discharge_complete` separately. It has a separate implementation profile.
- `evidence capture` preserves a request and legacy-method report in a new
  `.fartevidence` carrier. Use it when the user requests retention, with an
  explicit destination. It refuses existing destinations. `inspect`, `verify`,
  and `replay` do not integrate the solver; `reconstruct` calculates again.
  Carrier inputs allow 24 MiB, distinct from the 64 KiB request limit.
  A committed carrier can survive a stdout failure; inspect it before retrying.

Use `--format json` and inspect both exit status and the report. Completed
operations return 0. Failures return 1; input/model failures can still produce
structured JSON on stdout. Usage or output failures can instead write stderr.
Do not replace a refused input with a default case to make the command succeed.

For a bounded local experiment, the separate native Rust `fart` executable
provides `play run` and `play replay`. Use a matching native executable or
`cargo run --locked -p fart-cli --` with the repository's pinned Rust toolchain.
The Go recipe arrays do not run this native service.

```console
cargo run --locked -p fart-cli -- play run testdata/play/reservoir-session.jsonl --format jsonl
mkdir -p artifacts
cargo run --locked -p fart-cli -- play run testdata/play/reservoir-session.jsonl --format transcript > artifacts/session.json
cargo run --locked -p fart-cli -- play replay artifacts/session.json --format json
```

Read the [play session contract](../../docs/PLAY_SESSION.md) before editing the
command stream. Every prediction uses the immutable authored baseline. Exact
accepted retries return their original receipt without another attempt charge.
`--format transcript` emits the retained object for shell redirection; the caller
owns the output path and replacement behavior. Replay consumes that object,
checks its integrity without numerical evaluation, and does not restore a live
session. EOF never creates a finish: an unfinished run returns 1 but can still
retain a transcript whose integrity replay returns 0. Check completion,
truncation, and exit status separately.

Keep the user's law context, closure, units, and input provenance explicit.
The numerical models require explicit SI inputs. An incompatible law context is
refused; omission of `law_context` in a walk case selects no context. The sample
cases are software examples, with no calibration to an ordinary biological event.
Report relevant assumptions, regime, stopping condition, and evidence limits with
the numerical result. The Go command subset has no PlayService; the native local
profile is documented separately above. Neither implements plume, physical audio,
ratified canonical case identity, an MCP server, or an A2A agent.

Use `fartapp help <group> <operation>` or `go run ./cmd/fartapp help <group> <operation>`
for the current command contract. Consult the
[walkthrough](../../docs/WALKTHROUGH.md) for the coupled model and the
[carrier contract](../../docs/WALK_EVIDENCE.md) for retained evidence. Consult the
[interface plan](../../docs/INTERFACES.md) when distinguishing current CLI behavior
from planned adapters.
