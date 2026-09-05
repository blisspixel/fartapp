# One small coupled-oracle experiment

This walkthrough follows one explicit, low-energy synthetic reservoir through
the experimental Go CLI. It demonstrates a scientific account that can be
inspected, challenged, and checked against retained evidence.

Use the stable Go version in [go.mod](../go.mod). Run these commands from the
repository root. `go run ./cmd/fartapp` works without installing a binary.

## Read the authored case

[ordinary-low-pressure.json](../testdata/walk/ordinary-low-pressure.json)
declares a 100 mL rigid reservoir, temperature 293.15 K, back pressure 101325 Pa,
approximately 930 Pa initial overpressure, and a 1 mm2 restriction. Its
calorically perfect gas has explicit synthetic air-like properties. The
isothermal closure requires an accounted heat input as mass leaves.

These values are a software experiment. They are not a measured person,
biological population, hidden default, or the ratified Reference Pfft. All
inputs remain visible, including the step budget and a 2 mm2 counterfactual.

```console
go run ./cmd/fartapp --help
go run ./cmd/fartapp help walk
go run ./cmd/fartapp walk predict testdata/walk/ordinary-low-pressure.json
```

Prediction reports the initial flow regime and the thermodynamic equalization
endpoint. Reachability is separate: a closed restriction cannot reach that
endpoint, and a compliant opening with zero resting area approaches it only
asymptotically. An optional law context must match the implemented SI continuum
profile. The atemporal counterexample is rejected before simulation.

## Run and inspect

```console
go run ./cmd/fartapp walk simulate testdata/walk/ordinary-low-pressure.json
go run ./cmd/fartapp walk inspect testdata/walk/ordinary-low-pressure.json --format json
go run ./cmd/fartapp walk explain testdata/walk/ordinary-low-pressure.json
```

The case remains subsonic. At its current numerical settings, it emits about
1.1054 mg and reaches the stated pressure tolerance in about 55.5 ms. These are
step-dependent numerical results, not exact physical timing.

The JSON account contains the normalized inputs, model and implementation
revisions, binary64 numerical policy, compiler and platform, every retained
sample, component mass transfers, total source enthalpy flow, thrust and recoil,
termination reason, ledgers, and dry-flow signature. A completed withdrawal
always has a retained endpoint. The initial sample is present even when no
flow occurs.

Read the stop reason before interpreting the final state:

| Stop | Meaning |
| --- | --- |
| `no-flow` | The initial supported state has zero flow. |
| `equalized` | A positive-rest-area run reached its reported pressure tolerance. |
| `pressure-tolerance` | An asymptotically closing opening reached a numerical tolerance with positive overpressure. |
| `max-time` | The authored time budget ended the run. |
| `max-steps` | The authored work budget ended the run. |
| `no-progress` | Further positive progress could not be represented. |

Stopping at a budget does not mean the reservoir completed its discharge.
`L/D` uses a circular-equivalent reference diameter and integrated exit speed;
it does not resolve a plume or prove a vortex structure.

## Change one thing

```console
go run ./cmd/fartapp walk branch testdata/walk/ordinary-low-pressure.json
```

The authored variant doubles the prescribed area. For this constant-area model
and completed case, it reaches the same withdrawn mass in half the numerical
time. The report retains both accounts and their stopping reasons. A run that
stops at a fixed time can have a different withdrawn mass; the comparison must
say so.

This is the useful loop: make a prediction, inspect what changed, and follow the
equations to the explanation. The restriction controls rate. The reservoir
closure controls the thermodynamic path.

## Check arithmetic and retain a witness

```console
go run ./cmd/fartapp walk certify testdata/walk/ordinary-low-pressure.json
go run ./cmd/fartapp walk witness testdata/walk/ordinary-low-pressure.json --format json
```

The provisional `certify` operation checks numerical balance claims and shows
their residuals and tolerances. It grants no certificate authority, empirical
validation, independent error estimate, or case archive.

The witness is a versioned SHA-256 comparison of the normalized simulation inputs
and complete numerical account. It binds model, implementation, compiler,
platform, policy, history, transfers, and claims. The normalized-input digest
is reported separately. Neither digest is a ratified scientific case identity
or cryptographic signature.

Reconstruction requires a previously retained expected witness. It compares a
new calculation against that expectation. It does not generate its own expected
answer by running the same calculation twice.

In PowerShell 7:

```powershell
go run ./cmd/fartapp walk witness testdata/walk/ordinary-low-pressure.json --format json > walk-evidence.json
$walkEvidence = Get-Content -Raw walk-evidence.json | ConvertFrom-Json
$walkEvidence.inputs | Add-Member -NotePropertyName expected_witness -NotePropertyValue $walkEvidence.witness
$walkEvidence.inputs | ConvertTo-Json -Depth 20 | go run ./cmd/fartapp walk reconstruct - --format json
```

In a Unix shell with optional `jq`:

```sh
go run ./cmd/fartapp walk witness testdata/walk/ordinary-low-pressure.json --format json > walk-evidence.json
jq '.inputs + {expected_witness: .witness}' walk-evidence.json | go run ./cmd/fartapp walk reconstruct - --format json
```

The digest can also be copied into an `expected_witness` member by hand.
No JSON processor is required by the application. Keep the same compiler,
platform, architecture, implementation, and numerical policy for a matching
witness. A changed profile can produce an honest mismatch. The command retains
both digests on mismatch and returns a failing exit status.

## What the evidence earns next

The optional accurate operation uses the same authored case:

```console
go run ./cmd/fartapp walk refine testdata/walk/ordinary-low-pressure.json --relative-tolerance 1e-8 --max-evaluations 100000
```

It gives about 58.3942 ms for complete discharge, with explicit quadrature
estimates and work counters. Its independent analytical clock differs from the
legacy mass-step method's 55.5 ms approximation. The
[analytical reference](BLOWDOWN_REFERENCE.md) explains the endpoint
regularization and limits. Tolerance satisfaction does not imply empirical
validation or a completed discharge when the sample budget truncates a run.

The [evidence carrier](WALK_EVIDENCE.md) automates retention without manually
copying a digest. Capture retains the legacy method under its declared current
implementation revision, then inspect, verify, and replay run without time
integration. Reconstruction explicitly calculates again. Refinement is a
separate numerical profile.

The Reference Pfft, complete play loop, physical audio, certified archives, and trusted
`RES-002` benchmark retain their [roadmap gates](../ROADMAP.md). This walkthrough
provides an inspectable foundation for them.
