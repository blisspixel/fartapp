use fart_domain::restriction::{FlowError, Request};

use super::Regime;

use crate::{
    BalanceClaim,
    numeric::{product_over, product_ratio},
};

/// Complete finite control-section account over explicit validated inputs.
#[derive(Clone, Debug, PartialEq)]
pub struct FlowResult {
    /// Original immutable request.
    pub request: Request,
    /// Control-section regime.
    pub regime: Regime,
    /// Effective open area in square metres.
    pub effective_area_m2: f64,
    /// Sonic static-to-total pressure ratio.
    pub critical_pressure_ratio: f64,
    /// Authored back-to-total pressure ratio.
    pub back_pressure_ratio: f64,
    /// Control-section Mach number, exactly one for choked flow.
    pub throat_mach: f64,
    /// Control-section static pressure in pascals.
    pub exit_pressure_pa: f64,
    /// Control-section static temperature in kelvin.
    pub exit_temperature_k: f64,
    /// Control-section speed in metres per second.
    pub exit_speed_m_per_s: f64,
    /// Discharge-scaled mass flow in kilograms per second.
    pub mass_flow_kg_per_s: f64,
    /// Sonic comparison rate for the same open area, when representable.
    pub sonic_mass_flow_kg_per_s: f64,
    /// Momentum plus pressure thrust in newtons.
    pub thrust_n: f64,
    /// Equal and opposite source recoil in newtons.
    pub recoil_n: f64,
    /// Residual of discharge-scaled mass flux in kilograms per second.
    pub mass_flow_residual_kg_per_s: f64,
    /// Control-surface thrust residual in newtons.
    pub thrust_residual_n: f64,
    /// Action-reaction residual in newtons.
    pub recoil_residual_n: f64,
    /// Mass flux, thrust, and recoil roundoff evidence, in that order.
    pub claims: [BalanceClaim; 3],
}

/// Evaluate a quasi-steady calorically perfect isentropic converging restriction.
/// Positive unrepresentable flow is refused; it never becomes an authored closure.
pub fn evaluate(request: &Request) -> Result<FlowResult, FlowError> {
    let stagnation = request.stagnation();
    let (p0, t0, gas, gamma) = (
        stagnation.pressure().get(),
        stagnation.temperature().get(),
        stagnation.gas_constant().get(),
        stagnation.gamma().get(),
    );
    let back = request.back_pressure().get();
    let area = request.area().effective(p0 - back)?.get();
    let critical = stagnation.critical_pressure_ratio();
    let ratio = back / p0;
    if !ratio.is_finite() || ratio < 0.0 {
        return Err(FlowError::NoRepresentableFlow);
    }
    let cd = request.discharge_coefficient().get();
    let mut result = FlowResult {
        request: *request,
        regime: Regime::NoFlow,
        effective_area_m2: area,
        critical_pressure_ratio: critical,
        back_pressure_ratio: ratio,
        throat_mach: 0.0,
        exit_pressure_pa: back,
        exit_temperature_k: t0,
        exit_speed_m_per_s: 0.0,
        mass_flow_kg_per_s: 0.0,
        sonic_mass_flow_kg_per_s: 0.0,
        thrust_n: 0.0,
        recoil_n: 0.0,
        mass_flow_residual_kg_per_s: 0.0,
        thrust_residual_n: 0.0,
        recoil_residual_n: 0.0,
        claims: [BalanceClaim {
            id: "",
            method: "",
            residual: 0.0,
            tolerance: 0.0,
            unit: "",
        }; 3],
    };
    // A closed restriction has no reverse-flow operation to evaluate.
    if area == 0.0 {
        return with_claims(result);
    }
    if back > p0 {
        return Err(FlowError::AdversePressure);
    }
    let sonic = sonic_rate(cd, area, p0, t0, gas, gamma, critical);
    if back == p0 {
        result.sonic_mass_flow_kg_per_s = sonic.unwrap_or(0.0);
        return with_claims(result);
    }
    result.sonic_mass_flow_kg_per_s = sonic?;
    // Dividing adjacent pressures by p0 can erase which side of the sonic
    // section they lie on. Compare pressures directly so pressure thrust never
    // reverses solely because the reported ratios rounded to the same value.
    let (mach, pressure, temperature) = if back <= p0 * critical {
        result.regime = Regime::Choked;
        (1.0, p0 * critical, sonic_temperature(t0, gamma))
    } else {
        result.regime = Regime::Subsonic;
        // expm1 preserves a representable small pressure gap; square-root
        // factors preserve Mach if only its square is below binary64 range.
        let logarithm = ((p0 - back) / back).ln_1p();
        let expansion = (((gamma - 1.0) / gamma) * logarithm).exp_m1();
        let coefficient = 2.0 / (gamma - 1.0);
        let mach_squared = coefficient * expansion;
        let mut mach = mach_squared.sqrt();
        if mach_squared < f64::MIN_POSITIVE || !mach.is_finite() {
            mach = product_over(&[coefficient.sqrt(), expansion.sqrt()], 1.0);
        }
        if mach >= 1.0 {
            if mach > 1.0 + 8.0 * f64::EPSILON {
                return Err(FlowError::NoRepresentableFlow);
            }
            mach = 1.0_f64.next_down();
        }
        (mach, back, t0 / (1.0 + (gamma - 1.0) / 2.0 * mach * mach))
    };
    if !positive(pressure) || !positive(temperature) || !positive(mach) {
        return Err(FlowError::NoRepresentableFlow);
    }
    let (speed, mass, mass_definition) =
        transport(cd, area, pressure, temperature, gas, gamma, mach);
    let pressure_thrust = (pressure - back) * area;
    let momentum = product_over(&[mass, speed], 1.0);
    let thrust = momentum + pressure_thrust;
    if ![speed, mass, result.sonic_mass_flow_kg_per_s]
        .into_iter()
        .all(positive)
        || !thrust.is_finite()
        || !pressure_thrust.is_finite()
    {
        return Err(FlowError::NoRepresentableFlow);
    }
    result.throat_mach = mach;
    result.exit_pressure_pa = pressure;
    result.exit_temperature_k = temperature;
    result.exit_speed_m_per_s = speed;
    result.mass_flow_kg_per_s = mass;
    result.mass_flow_residual_kg_per_s = crate::signed_sum(vec![mass, -mass_definition]);
    result.thrust_n = thrust;
    result.recoil_n = -thrust;
    result.thrust_residual_n = crate::signed_sum(vec![thrust, -momentum, -pressure_thrust]);
    result.recoil_residual_n = crate::signed_sum(vec![-thrust, thrust]);
    with_claims(result)
}

fn with_claims(mut result: FlowResult) -> Result<FlowResult, FlowError> {
    let flow = result.mass_flow_kg_per_s;
    result.claims = [
        crate::claim(
            "restriction.mass-flow-definition",
            "cd-scaled-exit-mass-flux",
            "kg/s",
            result.mass_flow_residual_kg_per_s,
            &[flow, result.sonic_mass_flow_kg_per_s],
        )
        .map_err(|_| FlowError::InvariantViolation)?,
        crate::claim(
            "restriction.thrust-control-surface",
            "momentum-and-pressure-thrust",
            "N",
            result.thrust_residual_n,
            &[
                result.thrust_n,
                product_over(&[flow, result.exit_speed_m_per_s], 1.0),
            ],
        )
        .map_err(|_| FlowError::InvariantViolation)?,
        crate::claim(
            "restriction.recoil-action-reaction",
            "equal-and-opposite-force",
            "N",
            result.recoil_residual_n,
            &[result.recoil_n, result.thrust_n],
        )
        .map_err(|_| FlowError::InvariantViolation)?,
    ];
    Ok(result)
}

fn sonic_temperature(total: f64, gamma: f64) -> f64 {
    let ordinary = total * 2.0 / (gamma + 1.0);
    if positive(ordinary) {
        ordinary
    } else {
        product_over(&[total, 2.0], gamma + 1.0)
    }
}

fn sonic_rate(
    cd: f64,
    area: f64,
    pressure: f64,
    temperature: f64,
    gas: f64,
    gamma: f64,
    critical: f64,
) -> Result<f64, FlowError> {
    let pressure = pressure * critical;
    let temperature = sonic_temperature(temperature, gamma);
    if !positive(pressure) || !positive(temperature) {
        return Err(FlowError::NoRepresentableFlow);
    }
    let (speed, mass, _) = transport(cd, area, pressure, temperature, gas, gamma, 1.0);
    if !positive(speed) || !mass.is_finite() || mass < 0.0 {
        return Err(FlowError::NoRepresentableFlow);
    }
    Ok(mass)
}

fn transport(
    cd: f64,
    area: f64,
    pressure: f64,
    temperature: f64,
    gas: f64,
    gamma: f64,
    mach: f64,
) -> (f64, f64, f64) {
    let gamma_gas = gamma * gas;
    let squared_sound_speed = gamma_gas * temperature;
    let mut speed = mach * squared_sound_speed.sqrt();
    if !positive(speed) || gamma_gas < f64::MIN_POSITIVE || squared_sound_speed < f64::MIN_POSITIVE
    {
        speed = product_over(&[mach, gamma.sqrt(), gas.sqrt(), temperature.sqrt()], 1.0);
    }
    let thermal = gas * temperature;
    let density = pressure / thermal;
    let kinematic_rate = density * area * speed;
    let ordinary = cd * kinematic_rate;
    let (mass, definition) = if thermal >= f64::MIN_POSITIVE
        && thermal.is_finite()
        && density >= f64::MIN_POSITIVE
        && density.is_finite()
        && kinematic_rate >= f64::MIN_POSITIVE
        && kinematic_rate.is_finite()
        && positive(ordinary)
    {
        (ordinary, ordinary)
    } else {
        // NASA's mass-flux relation A*M*p*sqrt(gamma)/(sqrt(R)*sqrt(T)).
        // Keep both denominator factors separate through the final rounding.
        (
            product_ratio(
                &[cd, area, mach, pressure, gamma.sqrt()],
                &[gas.sqrt(), temperature.sqrt()],
            ),
            product_ratio(&[cd, pressure, area, speed], &[gas, temperature]),
        )
    };
    (speed, mass, definition)
}

fn positive(value: f64) -> bool {
    value.is_finite() && value > 0.0
}
