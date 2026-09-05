# Coupled blowdown analytical reference

The experimental Go oracle couples a finite reservoir to a restriction. Its
mass and energy ledgers check the thermodynamic path. Independent solutions
check whether that path is traversed at the correct rate. These are different
proof obligations.

## Supported model and reference boundary

The reservoir is rigid, homogeneous, nonreacting, and calorically perfect.
Component mass fractions remain constant. The bulk reservoir velocity is
negligible, back pressure is constant, and the restriction is a one-way,
quasi-steady, isentropic converging section with an explicit discharge
coefficient. Adiabatic and prescribed-isothermal closures are ideal limits.
They do not supply a wall heat-transfer law. Spatial equilibration inside the
reservoir is an assumption; this model does not resolve or verify it.

The constant-area references additionally require positive, constant area and
discharge coefficient. Their synthetic parameters are not biological defaults.
The restriction equations follow the
[NASA mass-flow function](https://www.grc.nasa.gov/www/k-12/BGP/mflchk.html) and
[isentropic relations](https://www.grc.nasa.gov/www/k-12/airplane/isentrop.html).
[Dutton and Coverdill (1997)](https://www.ijee.ie/articles/Vol13-2/ijee924.pdf),
printed pages 124-125, supply compatible published discharge equations. The
tests compare to those analytical equations, not to the paper's measurements.
This is code verification and literature-relative analytical consistency.

## Thermodynamic path and time

Let `q` be positive outward mass flow, `x=m/m0`, and `n=1` for prescribed
isothermal or `n=gamma` for adiabatic closure. Constant mixture properties obey
`cp=cv+R` and `gamma=cp/cv`.

```text
dm/dt = -q
d(m cv T)/dt = Qdot - q cp T
T/T0 = x^(n-1)
P/P0 = x^n

adiabatic:  Hout = m0 cv T0 (1-x^gamma), Qin = 0
isothermal: Hout = m0 cp T0 (1-x), Qin = m0 R T0 (1-x)
Ufinal + Hout - Qin - Uinitial = 0
```

For `r=Pb/P`, define:

```text
alpha = (2/(gamma+1))^(gamma/(gamma-1))
Phi_choked = sqrt(gamma) * (2/(gamma+1))^((gamma+1)/(2*(gamma-1)))
Phi_subsonic = sqrt(2 gamma/(gamma-1) *
                   (r^(2/gamma)-r^((gamma+1)/gamma)))
q = Cd A P / sqrt(R T) * Phi(r)
B = V/(Cd A sqrt(R T0))
k = Phi_choked/B
```

Choking requires `r<=alpha`. While that condition holds, direct integration
gives:

```text
isothermal: x = exp(-k t)
adiabatic:  x = [1+(gamma-1)k t/2]^(-2/(gamma-1))
```

For an initially choked case, `r0=Pb/P0`, `xc=(r0/alpha)^(1/n)`, and:

```text
tc = -log(xc)/k                              if n=1
tc = 2*(xc^(-(n-1)/2)-1)/((n-1)*k)           otherwise
xe = r0^(1/n)
```

An initially subsonic case starts its subsonic integration at `x=1,t=0`.
Continuing the choked law to equalization would give the wrong elapsed time.

## Independently derived subsonic primitive

For `gamma=3/2`, substitution in the mass-flow equation gives elementary time
primitives independent of the production restriction and reservoir functions.
The test checks these alongside published `gamma=7/5` equations.

For isothermal closure, set `s=sqrt(1-(r0/x)^(1/3))`:

```text
Jiso(s) = B sqrt(6) * [s/(4*(1-s^2)^2)
                      +3*s/(8*(1-s^2)) +3*atanh(s)/8]
```

For adiabatic closure, set `b=r0^(1/3)` and `w=sqrt(sqrt(x)-b)`:

```text
Jadi(w) = 4 B (w^3/3+b*w)/(sqrt(6)*r0^(2/3))
```

Both primitives vanish at the equalization mass. In each case:

```text
t(x) = tc + J(xc) - J(x)
te = tc + J(xc)
```

For `m0=1.5625 kg`, `R=200 J/(kg K)`, `cv=400 J/(kg K)`, `T0=400 K`,
`V=1 m^3`, `A=0.01 m^2`, `Cd=1`, `Pb=50000 Pa`, and `P0=125000 Pa`:

| Closure | Choking ends (s) | Equalization (s) |
| --- | ---: | ---: |
| Prescribed isothermal | 0.12449023055628441 | 0.6136028612846427 |
| Adiabatic | 0.08472445964707728 | 0.44905436251851805 |

These digits describe analytical software references, not measured precision.

## Energy, recoil, and history

The source total enthalpy is `cp*T`. At the control section it separates into
static enthalpy `cp*Te` and kinetic energy `u^2/2`. Frozen-stagnation restriction
histories report all three integrals. A coupled reservoir exports total
enthalpy, which must close its internal-energy account with any prescribed
heat input.

Thrust includes momentum and pressure terms:

```text
F = q u + (Pe-Pb) A
recoil = -F
```

For the choked constant-area reference, with
`u0=sqrt(2 gamma R T0/(gamma+1))`:

```text
I = A P0 alpha (Cd gamma+1) *
    2*(1-x^((n+1)/2))/(k*(n+1)) - A Pb t
L = -(u0/k)*log(x)
```

These expressions independently check the impulse and stroke approximation.
Recoil cancellation alone verifies a sign convention. It does not independently
establish a resolved exterior momentum balance.

The retained history has one initial sample and one sample for every completed
withdrawal, including a final sample when the run stops at its step or time
budget. Component masses and cumulative component transfers remain inspectable.
Positive flow that cannot be represented is a numerical failure or a
`no-progress` stop, not a zero-flow identity.

## Legacy accuracy and termination contract

`Simulate` and the existing `walk simulate`, `walk witness`, and retained
evidence operations use the left-endpoint restriction rate and an exact finite
thermodynamic withdrawal. Their versioned numerical profiles remain unchanged.
The method has first-order time accuracy away from pressure equalization.
At equalization, `dt/dm` has an integrable
`1/sqrt(m-me)` singularity. The complete-discharge time consequently converges
only as `O(sqrt(max_withdrawal_fraction_per_step))`.

The executable checks in
[reference_test.go](../internal/coupledblowdown/reference_test.go) cover:

- Choked histories for both closures and `gamma=1.4`, `1.5`, and `5/3`, with
  decreasing fixed-time errors under successive step refinement.
- Independently derived and published subsonic relations, both initially choked
  and initially subsonic, plus transition bracketing.
- Constant-area endpoint and elapsed-time refinement from 128 to 2048 aligned
  mass intervals. The finest time error must be below 1.5 percent and decrease
  at every refinement. Observed errors are about 1.26 percent isothermal and
  1.33 percent adiabatic at 2048 intervals.
- Component, total-mass, and thermodynamic closure, independent choked impulse
  and stroke, finite values, and the zero-rest-area compliance boundary.

`equalized` means the reported positive-rest-area endpoint is within the
reported binary64 pressure tolerance, without clamping its pressure.
`max-time`, `max-steps`, and `no-progress` remain distinct from completion.
With zero prescribed area and positive linear compliance, the flow scales as
`(P-Pb)^(3/2)` near closure and exact equalization takes infinite model time.
Such a run can report `pressure-tolerance`, with positive residual overpressure;
it cannot claim a finite exact equalization event.

## Opt-in accurate integration

`walk refine` uses `SimulateAccurate` and a separate implementation revision,
`go-oracle.walk-refine/v0alpha2`. Revision 2 includes the corrected restriction
temperature and small-Mach arithmetic. It reports estimated numerical errors and work
alongside the thermodynamic evidence. Supply the relative tolerance and
evaluation budget explicitly:

```sh
fartapp walk refine testdata/walk/ordinary-low-pressure.json --relative-tolerance 1e-8 --max-evaluations 100000
fartapp walk refine testdata/walk/ordinary-low-pressure.json --relative-tolerance 1e-8 --max-evaluations 100000 --absolute-time-tolerance 0 --format json
```

From a source checkout, replace `fartapp` with `go run ./cmd/fartapp`.
This operation does not replace the legacy solver or its witness profiles.

### Regularized thermodynamic coordinate

Define the dimensionless mass coordinate `z` by:

```text
xe = (Pb/P0)^(1/n)
d = 1-xe
x = xe+d*z^2
z = 1 initially; z = 0 at analytical equalization

Jt(z) = 2*m0*d*z/q
JI(z) = 2*m0*d*z*u + (Pe-Pb)*A*Jt(z)
JL(z) = u*Jt(z)
t(z) = integral from z to 1 of Jt(v) dv
```

The corresponding integrals of `JI` and `JL` give thrust impulse and stroke.
With positive rest area, both `q` and `dm/dz` approach zero in proportion to
`z`, leaving a finite time integrand. The implementation cancels that factor
analytically instead of subtracting nearly equal pressures at quadrature
points. For the subsonic branch:

```text
v = d*z^2/xe
L = log(P/Pb) = n*log1p(v)
a = (gamma-1)/gamma
(u/z)^2 = 2*R*T*n*d/xe * (log1p(v)/v) * (-expm1(-a*L)/(a*L))
```

The last two ratios each approach one at the endpoint. Mass and temperature
samples are reconstructed from the initial constant-composition path. Their
cumulative mass, enthalpy, and heat accounts do not depend on the number of
quadrature subdivisions.

Integration splits at the choking pressure and the cap transition of
`A=min(Amax,A0+C*(P-Pb))`. It also splits where the compliant opening equals
the rest area, helping resolve a narrow near-closure contribution. Each smooth
piece uses an embedded 7/15 Gauss-Kronrod rule with bounded subdivision. The
nodes, weights, and local error rescaling follow
[QUADPACK DQK15](https://www.netlib.org/quadpack/dqk15.f).

### Accuracy evidence and supported scope

The reported time request is:

```text
estimated_time_error <= absolute_time_tolerance
                        + relative_tolerance * elapsed_time
```

Impulse and stroke each use the relative tolerance alone. The relative request
must be between `1e-12` and `0.1`; the absolute time tolerance must be finite
and nonnegative and defaults to zero in the CLI. Error estimates are
**a posteriori estimates, not rigorous bounds**. They exclude input uncertainty,
floating-point representation of the thermodynamic path and endpoint, and
physical model error. A closed ledger does not establish an error bound.

Successful reports expose `accuracy.tolerance_satisfied` separately from
`accuracy.discharge_complete`. A `max-steps` stop can satisfy the requested
quadrature accuracy for the traversed portion while leaving discharge
incomplete. The sample fraction selects retained mass samples; it does not
set integration accuracy. Every completed mass step retains its endpoint.

The work budget must be between 15 and 1,000,000 integrand evaluations. The
report gives evaluations, accepted intervals, and refinements. Subdivision
also has a hard depth limit of 48. Retained history remains bounded by 4096
mass steps and 4097 samples. Exhausted work or unresolved numerical accuracy
produces an explicit refusal, with work counters and no error estimates or
convergence claim.

Flowing cases support both closures, positive prescribed rest area, and either
constant area or capped linear compliance. They require `step.max_time_s=0`.
An authored positive time limit is explicitly unsupported by this method;
controlled inversion to a prescribed time remains a future numerical gate.
Zero-rest-area compliant discharge is also refused because its exact
equalization time is infinite. Valid no-flow identities remain exact, including
closed area or initially equal pressure.

Very small representable initial overpressure does not automatically establish
a meaningful accuracy request. If the total equalization mass fraction is
below `128*epsilon`, where binary64 `epsilon=2^-52`, the method refuses the
roundoff-dominated clock. Other unrepresentable states, totals, signatures, or
unresolved narrow tails also refuse. Finite inputs do not guarantee that every
derived integral can be represented.

For a completed run, the clock reaches the analytical equalization endpoint.
The retained binary64 state can have a small positive pressure gap within the
reported pressure tolerance. Its pressure is never clamped below back pressure.
The quadrature estimate does not bound this endpoint representation effect.

### Independent verification of the accurate method

[accurate_test.go](../internal/coupledblowdown/accurate_test.go) checks the
independently derived `gamma=3/2` and published `gamma=7/5` time solutions for
both closures. With the reference parameters above, `cv=R/(gamma-1)`, and a
coarse retained mass fraction of `0.4`, observed complete times are:

| Gamma | Closure | Computed time (s) | Mass steps | Evaluations |
| --- | --- | ---: | ---: | ---: |
| 1.4 | Prescribed isothermal | 0.6210815056104255 | 2 | 45 |
| 1.4 | Adiabatic | 0.4803111200222318 | 2 | 45 |
| 1.5 | Prescribed isothermal | 0.6136028612846429 | 2 | 45 |
| 1.5 | Adiabatic | 0.4490543625185178 | 2 | 45 |

The local binary64 comparisons differ from the independent evaluations by at
most `1.2e-16 s`. The tests require relative time error below `2e-11`; these
results are numerical observations, not additional guaranteed precision.
With initial back-pressure ratio `0.001` and one retained mass step, tightening
the requested tolerance increases or preserves work and reduces or preserves
the actual time error down to rounding. Separate choked truncation tests check
the exact time, pressure-force impulse, recoil, and stroke relations.

For the ordinary low-pressure case, the accurate time is
`0.05839420446440555 s`, compared with the legacy
`0.05550629534650945 s`. The published initially-subsonic primitive independently
checks the accurate value. The original sample fraction `0.00005` produces
183 mass steps and 2745 evaluations; a fraction of `0.01` produces one mass
step and 15 evaluations with the same computed time. At relative tolerance
`1e-10`, the reported time-error estimate is about `6.48e-16 s` in both runs.

A further independently derived reference chooses `A0=C*Pb`, so the uncapped
area becomes `A=A0*P/Pb`. With `B0=V/(Cd*A0*sqrt(R*T0))`, the subsonic tail
primitives for `gamma=3/2` are:

```text
isothermal: J(s) = B0*sqrt(6)*s

adiabatic:  b = r0^(1/3), w = sqrt(sqrt(x)-b)
            J(w) = 4*B0*b/sqrt(6) *
                   [w/(2*b*(w^2+b)) + atan(w/sqrt(b))/(2*b^(3/2))]
```

Here `s` is the isothermal coordinate defined earlier. In the uncapped choked
phase, `dx/dt=-k0*x^((3*n+1)/2)/r0`, with
`k0=A0*sqrt(R*T0)*Phi_choked/V`. Combining these expressions with the capped
constant-area phase checks cap transitions before and after choking ends.
Additional tests cover doubled area, preserved component fractions, finite
integrals across extreme scales and subnormal source rates, exact no-flow,
strict history and evaluation budgets, and explicit unresolved-tail refusal.

Controlled fixed-time inversion, a rigorous error certificate, full trusted
`RES-002`, a selected wall heat-transfer closure, RP-1 ratification, and
empirical validation remain open.
