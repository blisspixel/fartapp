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

## Current accuracy and termination contract

The stepper uses the left-endpoint restriction rate and an exact finite
thermodynamic withdrawal. It has first-order time accuracy away from pressure
equalization. At equalization, `dt/dm` has an integrable
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

The next numerical gate is controlled fixed-time and event-time error, including
choking and compliance-cap transitions. A regularized variable such as
`z=sqrt(m-me)` for positive-rest-area tails can remove the endpoint singularity.
That improvement needs fresh convergence evidence before it replaces this
method. Full trusted `RES-002`, a selected wall heat-transfer closure, RP-1
ratification, and empirical validation remain open.
