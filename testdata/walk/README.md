# Coupled-oracle walkthrough fixtures

These are synthetic code-verification inputs for the experimental
`continuum.quasi-steady-coupled-blowdown@v0alpha1` model. None is the Reference
Pfft, a biological norm, an empirical validation case, or a certified archive.

| Fixture | Purpose |
| --- | --- |
| [ordinary-low-pressure.json](ordinary-low-pressure.json) | A 100 mL rigid reservoir at 293.15 K with approximately 930 Pa overpressure against an explicit 101325 Pa back pressure. A conventional air-like ideal gas, 1 square mm prescribed area, and a doubled-area comparison demonstrate subsonic flow and a useful numerical counterfactual. |
| [isothermal-choked.json](isothermal-choked.json) | A large synthetic gamma-1.5 calculation chosen for analytical checks. Its short time budget stops both area variants before equalization, so they emit different masses over the same duration. |
| [atemporal-no-dimension.json](atemporal-no-dimension.json) | A rejection fixture. The SI solver must refuse the incompatible atemporal law context and emit no physical result. |

The ordinary fixture's component identifier and constant properties describe a
project convention. They do not supply a versioned measured-air property model.
The isothermal closure prescribes constant temperature and reports the heat it
requires. It does not simulate a wall heat-transfer law.

## Try one calculation and one variation

From the repository root:

```console
fartapp walk predict testdata/walk/ordinary-low-pressure.json
fartapp walk explain testdata/walk/ordinary-low-pressure.json
fartapp walk branch testdata/walk/ordinary-low-pressure.json
fartapp walk certify testdata/walk/ordinary-low-pressure.json --format json
```

Both ordinary area variants reach the same equalization mass. Doubling the
constant area halves the duration under this model and shared withdrawal-step
policy. These durations depend on numerical refinement. A balanced ledger does
not bound elapsed-time error. The `certify` operation exposes arithmetic balance
checks with actual residuals and tolerances; it issues no certificate.

The JSON report retains normalized inputs, model and runtime profile, and bounded
history with the initial state and every completed step, including the final
state. History includes reservoir state, restriction regime, exit state, source
total enthalpy rate, thrust, recoil, and component masses and cumulative outflow.
Component identifiers are sorted before the calculation. Source total enthalpy
includes the energy represented as exit kinetic energy and must not be confused
with exit static enthalpy.

Closed restrictions preserve the initial state. A compliant restriction with
zero resting area approaches back pressure only asymptotically. A numerical
`pressure-tolerance` stop does not turn that limit into finite-time physical
equalization. Step and time budget stops remain distinct.

## Retain and compare a witness

First save evidence from one calculation:

```console
fartapp walk witness testdata/walk/ordinary-low-pressure.json --format json > walk-evidence.json
```

The `witness` is a SHA-256 digest under `fart.walk-witness/v0alpha1`. It binds the
complete simulation account, including normalized inputs, component identities,
history, balances, model revision, implementation revision, Go version, operating
system, and architecture. The separately versioned `input_digest` binds normalized
request values. JSON whitespace and object order do not affect either digest.
`expected_witness` is comparison metadata and is excluded from normalized inputs.

In PowerShell, reconstruct from the retained inputs and expected digest:

```powershell
$walkEvidence = Get-Content -Raw walk-evidence.json | ConvertFrom-Json
$walkEvidence.inputs | Add-Member -NotePropertyName expected_witness -NotePropertyValue $walkEvidence.witness
$walkEvidence.inputs | ConvertTo-Json -Depth 20 | ./fartapp walk reconstruct - --format json
```

In a shell with the optional `jq` utility:

```sh
jq '.inputs + {expected_witness: .witness}' walk-evidence.json | ./fartapp walk reconstruct - --format json
```

The same operation can be performed manually by copying the retained `inputs`
object to a JSON file, adding `expected_witness`, and passing that file to
`walk reconstruct`. Reconstruction performs one new calculation and compares
it with the retained expectation. Missing expectations fail before simulation;
mismatches return status 1 while retaining both digests and the new account.

Witnesses use versioned Go JSON bytes. They are not RFC 8785 canonical JSON,
scientific or occurrence identities, signatures, immutable build provenance, or
empirical proof. Matching requires the same declared implementation and runtime
profile. Cross-platform or cross-version bitwise reconstruction is not promised.

## Input and resource boundary

Each request explicitly names its model, SI quantity system, reservoir, closure,
restriction, and withdrawal-step policy. An optional `law_context` must be the
exact compatible `earth.continuum.si@v0alpha1` revision; omitting it selects no
context. Every operation rejects invalid policy fields, even when that operation
does not use elapsed-time stepping.

Input is limited to 65536 bytes, JSON nesting depth 32, member names of 128 bytes,
64 components, and 4096 simulation steps per calculation. History contains at
most 4097 samples per calculation. A branch includes at most two such histories.
`max_time_s` is a nonnegative numerical budget; zero or omission means no time
ceiling, with the step ceiling still enforced. Explicit null, unknown or
case-aliased members, duplicate members, trailing JSON values, malformed UTF-8,
and inactive area-law parameters are rejected.
