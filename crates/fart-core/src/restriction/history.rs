use fart_domain::restriction::{
    AreaLaw, DischargeCoefficient, FlowError, HistorySample, MAX_HISTORY_SAMPLES, Pressure,
    Request, Stagnation,
};

use crate::{
    BalanceClaim,
    numeric::{product_over, product_ratio},
};

use super::{FlowResult, evaluate};

/// One prescribed time and its finite quasi-steady control-section account.
#[derive(Clone, Debug, PartialEq)]
pub struct HistoryInstant {
    /// Prescribed sample time in seconds.
    pub time_s: f64,
    /// Quasi-steady flow at the sample's prescribed area.
    pub flow: FlowResult,
}

/// Finite trapezoidal integrals over a bounded prescribed-area history.
#[derive(Clone, Debug, PartialEq)]
pub struct HistoryResult {
    /// Frozen upstream total state.
    pub stagnation: Stagnation,
    /// Frozen exterior static pressure.
    pub back_pressure: Pressure,
    /// Constant mass-flow discharge multiplier.
    pub discharge_coefficient: DischargeCoefficient,
    /// At most 256 evaluated samples in strictly increasing time order.
    pub samples: Vec<HistoryInstant>,
    /// Integrated mass leaving the frozen source in kilograms.
    pub mass_out_kg: f64,
    /// Integrated static exit enthalpy in joules, with datum cp*T_exit.
    pub enthalpy_out_j: f64,
    /// Integrated exit kinetic energy in joules.
    pub kinetic_energy_out_j: f64,
    /// Integrated source stagnation enthalpy in joules, with datum cp*T0.
    pub total_enthalpy_out_j: f64,
    /// Integrated momentum and pressure thrust in newton seconds.
    pub impulse_n_s: f64,
    /// Source recoil impulse in newton seconds.
    pub recoil_impulse_n_s: f64,
    /// Recoil action-reaction residual in newton seconds.
    pub recoil_residual_n_s: f64,
    /// Declared recoil-impulse balance evidence.
    pub claims: [BalanceClaim; 1],
}

/// Integrate explicitly sampled areas over fixed stagnation, back pressure and Cd.
/// One sample has zero duration. No component composition or depletion is inferred.
pub fn integrate_history(
    stagnation: Stagnation,
    back: Pressure,
    cd: DischargeCoefficient,
    samples: &[HistorySample],
) -> Result<HistoryResult, FlowError> {
    if samples.is_empty() || samples.len() > MAX_HISTORY_SAMPLES {
        return Err(FlowError::InvalidSampleCount);
    }
    let mut instants = Vec::with_capacity(samples.len());
    for (index, sample) in samples.iter().enumerate() {
        if index > 0 && sample.time().get() <= samples[index - 1].time().get() {
            return Err(FlowError::InvalidTime);
        }
        let request = Request::new(stagnation, back, AreaLaw::prescribed(sample.area()), cd);
        instants.push(HistoryInstant {
            time_s: sample.time().get(),
            flow: evaluate(&request)?,
        });
    }
    let (mut mass_out, mut static_energy, mut kinetic_energy, mut total_energy, mut impulse) =
        (0.0, 0.0, 0.0, 0.0, 0.0);
    for pair in instants.windows(2) {
        let dt = pair[1].time_s - pair[0].time_s;
        let (left, right) = (&pair[0].flow, &pair[1].flow);
        let mass = trapezoid(left.mass_flow_kg_per_s, right.mass_flow_kg_per_s, dt, 1.0);
        if left.mass_flow_kg_per_s == 0.0 && right.mass_flow_kg_per_s == 0.0 {
            // Exact closure/equal pressure needs no cp or energy calculation.
            continue;
        }
        if !positive(mass) {
            return Err(FlowError::NumericalDomain);
        }
        let open = if left.mass_flow_kg_per_s == 0.0 {
            right
        } else {
            left
        };
        let gas = stagnation.gas_constant().get();
        let gamma = stagnation.gamma().get();
        let cp = gas + gas / (gamma - 1.0);
        let static_part = enthalpy(mass, cp, gas, gamma, open.exit_temperature_k);
        let kinetic_part = product_over(
            &[mass, 0.5, open.exit_speed_m_per_s, open.exit_speed_m_per_s],
            1.0,
        );
        let total_part = enthalpy(mass, cp, gas, gamma, stagnation.temperature().get());
        let pressure_impulse = trapezoid(
            left.effective_area_m2,
            right.effective_area_m2,
            dt,
            open.exit_pressure_pa - back.get(),
        );
        let momentum_impulse = product_over(&[mass, open.exit_speed_m_per_s], 1.0);
        let impulse_part = momentum_impulse + pressure_impulse;
        if ![static_part, kinetic_part, total_part, impulse_part]
            .into_iter()
            .all(positive)
        {
            return Err(FlowError::NumericalDomain);
        }
        mass_out += mass;
        static_energy += static_part;
        kinetic_energy += kinetic_part;
        total_energy += total_part;
        impulse += impulse_part;
        if ![
            mass_out,
            static_energy,
            kinetic_energy,
            total_energy,
            impulse,
        ]
        .into_iter()
        .all(f64::is_finite)
        {
            return Err(FlowError::NumericalDomain);
        }
    }
    let recoil = -impulse;
    let residual = crate::signed_sum(vec![recoil, impulse]);
    let claim = crate::claim(
        "restriction-history.recoil-action-reaction",
        "equal-and-opposite-impulse",
        "N s",
        residual,
        &[impulse, recoil],
    )
    .map_err(|_| FlowError::InvariantViolation)?;
    Ok(HistoryResult {
        stagnation,
        back_pressure: back,
        discharge_coefficient: cd,
        samples: instants,
        mass_out_kg: mass_out,
        enthalpy_out_j: static_energy,
        kinetic_energy_out_j: kinetic_energy,
        total_enthalpy_out_j: total_energy,
        impulse_n_s: impulse,
        recoil_impulse_n_s: recoil,
        recoil_residual_n_s: residual,
        claims: [claim],
    })
}

fn enthalpy(mass: f64, cp: f64, gas: f64, gamma: f64, temperature: f64) -> f64 {
    if cp.is_finite() {
        product_over(&[mass, cp, temperature], 1.0)
    } else {
        product_ratio(&[mass, gas, gamma, temperature], &[gamma - 1.0])
    }
}

fn trapezoid(left: f64, right: f64, dt: f64, factor: f64) -> f64 {
    let scale = left.max(right);
    if scale == 0.0 {
        return 0.0;
    }
    product_over(
        &[scale, (left / scale + right / scale) / 2.0, dt, factor],
        1.0,
    )
}

fn positive(value: f64) -> bool {
    value.is_finite() && value > 0.0
}
