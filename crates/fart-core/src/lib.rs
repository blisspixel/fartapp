//! Pure rigid-reservoir endpoint equations over validated SI inputs.
//!
//! The model assumes a homogeneous nonreacting calorically perfect ideal mixture,
//! rigid volume, and composition-preserving withdrawal. It supplies no aperture,
//! elapsed time, momentum, plume, acoustic, case-admission, or empirical claim.

use fart_domain::{
    Closure, Component, Mass, ModelError, ReservoirState, Temperature, WithdrawalFraction,
};

mod numeric;
use numeric::{product_over, sum_products_over};

/// Derived SI state quantities, checked for finite positive representability.
#[derive(Clone, Copy, Debug, PartialEq)]
pub struct StateSummary {
    /// Total retained mass in kilograms.
    pub total_mass_kg: f64,
    /// Rigid volume in cubic metres.
    pub volume_m3: f64,
    /// Homogeneous temperature in kelvin.
    pub temperature_k: f64,
    /// Mass-weighted gas constant in joules per kilogram kelvin.
    pub gas_constant: f64,
    /// Mass-weighted specific isochoric heat in joules per kilogram kelvin.
    pub heat_cv: f64,
    /// Specific isobaric heat in joules per kilogram kelvin.
    pub heat_cp: f64,
    /// Dimensionless ratio of isobaric to isochoric heat capacity.
    pub gamma: f64,
    /// Ideal-mixture pressure in pascals.
    pub pressure_pa: f64,
    /// Sensible internal energy in joules, using the model datum cv*T.
    pub internal_energy_j: f64,
}

/// One component's mass transfer and independent balance residual.
#[derive(Clone, Debug, PartialEq)]
pub struct ComponentTransfer {
    /// Declared component identifier.
    pub id: String,
    /// Mass leaving the reservoir in kilograms, including zero withdrawal.
    pub mass_out_kg: f64,
    /// Initial minus retained minus transferred mass in kilograms.
    pub residual_kg: f64,
}

/// Arithmetic evidence with a finite roundoff allowance, not empirical evidence.
#[derive(Clone, Copy, Debug, PartialEq)]
pub struct BalanceClaim {
    /// Stable identity of the checked equation.
    pub id: &'static str,
    /// Method used to form the residual.
    pub method: &'static str,
    /// Residual in the declared unit.
    pub residual: f64,
    /// 64 epsilon times the largest term, plus the smallest positive subnormal.
    pub tolerance: f64,
    /// SI unit of the residual and tolerance.
    pub unit: &'static str,
}

/// Complete endpoint and transfer account for one immutable request.
#[derive(Clone, Debug, PartialEq)]
pub struct Transition {
    /// Original validated input state.
    pub before: ReservoirState,
    /// Retained validated state after withdrawal.
    pub after: ReservoirState,
    /// Initial derived state quantities.
    pub initial: StateSummary,
    /// Final derived state quantities.
    pub final_state: StateSummary,
    /// Explicitly selected closure.
    pub closure: Closure,
    /// Requested component-wise fraction.
    pub withdrawal: WithdrawalFraction,
    /// Component transfers in identifier order.
    pub components: Vec<ComponentTransfer>,
    /// Sum of component mass transfers in kilograms.
    pub total_mass_out_kg: f64,
    /// Integrated stagnation enthalpy leaving the reservoir in joules.
    pub enthalpy_out_j: f64,
    /// Heat supplied by the isothermal thermostat, otherwise zero, in joules.
    pub heat_in_j: f64,
    /// Total mass, energy, initial EOS, and final EOS checks in that order.
    pub claims: [BalanceClaim; 4],
}

/// Derive a homogeneous state without changing its inputs.
pub fn summarize(state: &ReservoirState) -> Result<StateSummary, ModelError> {
    let parts = state.components();
    let mass = stable_sum(parts.iter().map(|part| part.mass().get()).collect());
    if !mass.is_finite() {
        return Err(ModelError::NonFinite);
    }
    let gas = sum_products_over(
        parts
            .iter()
            .map(|part| [part.mass().get(), part.gas_constant().get()]),
        mass,
    );
    let cv = sum_products_over(
        parts
            .iter()
            .map(|part| [part.mass().get(), part.heat_capacity().get()]),
        mass,
    );
    let temperature = state.temperature().get();
    let volume = state.volume().get();
    let cp = cv + gas;
    let gamma = cp / cv;
    let pressure = product_over(&[mass, gas, temperature], volume);
    let energy = sum_products_over(
        parts
            .iter()
            .map(|part| [part.mass().get(), part.heat_capacity().get(), temperature]),
        1.0,
    );
    for value in [mass, gas, cv, cp, gamma, pressure, energy] {
        if !value.is_finite() {
            return Err(ModelError::NonFinite);
        }
        if value <= 0.0 {
            return Err(ModelError::NonPositive);
        }
    }
    Ok(StateSummary {
        total_mass_kg: mass,
        volume_m3: volume,
        temperature_k: temperature,
        gas_constant: gas,
        heat_cv: cv,
        heat_cp: cp,
        gamma,
        pressure_pa: pressure,
        internal_energy_j: energy,
    })
}

/// Compute an analytical endpoint and independent mass/energy balance checks.
/// Positive unrepresentable progress, exhaustion, nonfinite transfers, and failed
/// arithmetic claims are exact refusals. The input state is never modified.
pub fn withdraw_fraction(
    before: &ReservoirState,
    withdrawal: WithdrawalFraction,
    closure: Closure,
) -> Result<Transition, ModelError> {
    let initial = summarize(before)?;
    let fraction = withdrawal.get();
    let retained = 1.0 - fraction;
    if fraction > 0.0 && retained == 1.0 {
        return Err(ModelError::NoRepresentableProgress);
    }
    let mut components = Vec::with_capacity(before.components().len());
    let mut retained_components = Vec::with_capacity(before.components().len());
    for component in before.components() {
        let old_mass = component.mass().get();
        let new_mass = old_mass * retained;
        let mass_out = old_mass - new_mass;
        if fraction > 0.0 && (new_mass <= 0.0 || new_mass >= old_mass || mass_out <= 0.0) {
            return Err(ModelError::NoRepresentableProgress);
        }
        retained_components.push(Component::new(
            component.id(),
            Mass::new(new_mass)?,
            component.gas_constant(),
            component.heat_capacity(),
        )?);
        components.push(ComponentTransfer {
            id: component.id().to_owned(),
            mass_out_kg: mass_out,
            residual_kg: signed_sum(vec![old_mass, -new_mass, -mass_out]),
        });
    }
    let log_retained = (-fraction).ln_1p();
    let temperature = match closure {
        Closure::RigidAdiabatic => {
            let decay = (initial.gas_constant / initial.heat_cv) * log_retained;
            let factor = decay.exp();
            if factor >= f64::MIN_POSITIVE {
                initial.temperature_k * factor
            } else {
                // A subnormal/zero decay factor can still yield a representable
                // temperature after multiplication by a large initial value.
                (initial.temperature_k.ln() + decay).exp()
            }
        }
        Closure::RigidIsothermal => initial.temperature_k,
    };
    let after = ReservoirState::new(
        retained_components,
        before.volume(),
        Temperature::new(temperature).map_err(|_| ModelError::NumericalDomain)?,
    )?;
    let final_state = summarize(&after).map_err(|_| ModelError::NumericalDomain)?;
    let mass_out = stable_sum(components.iter().map(|part| part.mass_out_kg).collect());
    let (enthalpy_out, heat_in) = if fraction == 0.0 {
        (0.0, 0.0)
    } else if closure == Closure::RigidAdiabatic {
        let energy_fraction = -(initial.gamma * log_retained).exp_m1();
        (
            product_over(&[initial.internal_energy_j, energy_fraction], 1.0),
            0.0,
        )
    } else {
        let enthalpy = before
            .components()
            .iter()
            .zip(&components)
            .map(|(part, transfer)| {
                [
                    transfer.mass_out_kg,
                    part.heat_capacity().get() + part.gas_constant().get(),
                    initial.temperature_k,
                ]
            });
        let heat = before
            .components()
            .iter()
            .zip(&components)
            .map(|(part, transfer)| {
                [
                    transfer.mass_out_kg,
                    part.gas_constant().get(),
                    initial.temperature_k,
                ]
            });
        (
            sum_products_over(enthalpy, 1.0),
            sum_products_over(heat, 1.0),
        )
    };
    if [mass_out, enthalpy_out, heat_in]
        .iter()
        .any(|value| !value.is_finite() || *value < 0.0)
    {
        return Err(ModelError::NumericalDomain);
    }
    if fraction > 0.0
        && (enthalpy_out == 0.0 || (closure == Closure::RigidIsothermal && heat_in == 0.0))
    {
        return Err(ModelError::NoRepresentableProgress);
    }
    let claims = [
        claim(
            "reservoir.total-mass-balance",
            "double-entry-balance",
            "kg",
            signed_sum(vec![
                final_state.total_mass_kg,
                mass_out,
                -initial.total_mass_kg,
            ]),
            &[initial.total_mass_kg, final_state.total_mass_kg, mass_out],
        )?,
        claim(
            "reservoir.energy-balance",
            "double-entry-balance",
            "J",
            signed_sum(vec![
                final_state.internal_energy_j,
                enthalpy_out,
                -heat_in,
                -initial.internal_energy_j,
            ]),
            &[
                initial.internal_energy_j,
                final_state.internal_energy_j,
                enthalpy_out,
                heat_in,
                0.0,
            ],
        )?,
        eos_claim("reservoir.initial-equation-of-state", initial)?,
        eos_claim("reservoir.final-equation-of-state", final_state)?,
    ];
    Ok(Transition {
        before: before.clone(),
        after,
        initial,
        final_state,
        closure,
        withdrawal,
        components,
        total_mass_out_kg: mass_out,
        enthalpy_out_j: enthalpy_out,
        heat_in_j: heat_in,
        claims,
    })
}

fn eos_claim(id: &'static str, state: StateSummary) -> Result<BalanceClaim, ModelError> {
    let pv = state.pressure_pa * state.volume_m3;
    let mrt = product_over(
        &[state.total_mass_kg, state.gas_constant, state.temperature_k],
        1.0,
    );
    claim(
        id,
        "derived-state-consistency-residual",
        "J",
        pv - mrt,
        &[pv, mrt],
    )
}

fn claim(
    id: &'static str,
    method: &'static str,
    unit: &'static str,
    residual: f64,
    terms: &[f64],
) -> Result<BalanceClaim, ModelError> {
    if terms.iter().any(|term| !term.is_finite()) {
        return Err(ModelError::InvariantViolation);
    }
    let magnitude = terms.iter().map(|term| term.abs()).fold(0.0_f64, f64::max);
    let tolerance = (64.0 * f64::EPSILON) * magnitude + f64::from_bits(1);
    if !residual.is_finite() || residual.abs() > tolerance {
        return Err(ModelError::InvariantViolation);
    }
    Ok(BalanceClaim {
        id,
        method,
        residual,
        tolerance,
        unit,
    })
}

fn stable_sum(mut values: Vec<f64>) -> f64 {
    values.sort_by(f64::total_cmp);
    let (mut sum, mut compensation) = (0.0, 0.0);
    for value in values {
        let adjusted = value - compensation;
        let next = sum + adjusted;
        compensation = (next - sum) - adjusted;
        sum = next;
    }
    sum
}

fn signed_sum(mut values: Vec<f64>) -> f64 {
    values.sort_by(|left, right| left.abs().total_cmp(&right.abs()));
    let (mut sum, mut compensation) = (0.0_f64, 0.0);
    for value in values {
        let next = sum + value;
        compensation += if sum.abs() >= value.abs() {
            (sum - next) + value
        } else {
            (value - next) + sum
        };
        sum = next;
    }
    sum + compensation
}
