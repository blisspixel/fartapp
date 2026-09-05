//! Independent analytical, numerical-limit, and bounded-history evidence.

use fart_core::restriction::{FlowResult, Regime, evaluate, integrate_history};
use fart_domain::restriction::*;
use fart_domain::{SpecificGasConstant, Temperature};

fn state(pressure: f64, temperature: f64, gas: f64, gamma: f64) -> Stagnation {
    Stagnation::new(
        Pressure::new(pressure).unwrap(),
        Temperature::new(temperature).unwrap(),
        SpecificGasConstant::new(gas).unwrap(),
        HeatCapacityRatio::new(gamma).unwrap(),
    )
    .unwrap()
}

fn request(stagnation: Stagnation, back: f64, area: f64, cd: f64) -> Request {
    Request::new(
        stagnation,
        Pressure::new(back).unwrap(),
        AreaLaw::prescribed(Area::new(area).unwrap()),
        DischargeCoefficient::new(cd).unwrap(),
    )
}

fn near(actual: f64, expected: f64, relative: f64) {
    let tolerance = relative * expected.abs() + f64::from_bits(1);
    assert!(
        actual.is_finite() && (actual - expected).abs() <= tolerance,
        "actual {actual:e}, expected {expected:e}, tolerance {tolerance:e}"
    );
}

fn binary(exponent: i32) -> f64 {
    assert!((-1074..=1023).contains(&exponent));
    if exponent < -1022 {
        f64::from_bits(1 << (exponent + 1074))
    } else {
        f64::from_bits(((exponent + 1023) as u64) << 52)
    }
}

fn samples(values: &[(f64, f64)]) -> Vec<HistorySample> {
    values
        .iter()
        .map(|&(time, area)| {
            HistorySample::new(Seconds::new(time).unwrap(), Area::new(area).unwrap())
        })
        .collect()
}

fn check_finite(result: &FlowResult) {
    for value in [
        result.effective_area_m2,
        result.critical_pressure_ratio,
        result.back_pressure_ratio,
        result.throat_mach,
        result.exit_pressure_pa,
        result.exit_temperature_k,
        result.exit_speed_m_per_s,
        result.mass_flow_kg_per_s,
        result.sonic_mass_flow_kg_per_s,
        result.thrust_n,
        result.recoil_n,
        result.mass_flow_residual_kg_per_s,
        result.thrust_residual_n,
        result.recoil_residual_n,
    ] {
        assert!(value.is_finite());
    }
    for claim in result.claims {
        assert!(claim.residual.is_finite() && claim.tolerance.is_finite());
        assert!(claim.residual.abs() <= claim.tolerance);
    }
}

#[test]
fn gamma_three_halves_has_independent_choked_and_subsonic_anchors() {
    // NASA isentropic relations give T*/T0=4/5 and p*/p0=(4/5)^3.
    // Choosing p0=125000,R=200,T0=400 makes sonic density exactly 1 kg/m^3.
    // https://www.grc.nasa.gov/www/k-12/airplane/isentrop.html
    let input = request(state(125000.0, 400.0, 200.0, 1.5), 50000.0, 0.01, 1.0);
    let result = evaluate(&input).unwrap();
    assert_eq!(result.request, input);
    assert_eq!(result.regime, Regime::Choked);
    assert_eq!(result.regime.name(), "choked");
    near(result.critical_pressure_ratio, 64.0 / 125.0, 3e-16);
    near(result.exit_pressure_pa, 64000.0, 3e-16);
    near(result.exit_temperature_k, 320.0, 3e-16);
    near(result.exit_speed_m_per_s, 96000.0_f64.sqrt(), 3e-16);
    near(result.mass_flow_kg_per_s, 0.01 * 96000.0_f64.sqrt(), 3e-16);
    near(
        result.sonic_mass_flow_kg_per_s,
        result.mass_flow_kg_per_s,
        3e-16,
    );
    near(result.thrust_n, 1100.0, 3e-16);
    near(result.recoil_n, -1100.0, 3e-16);
    check_finite(&result);

    // tau=9/10 gives p/p0=729/1000 and M^2=4/9, independently of inversion.
    let result = evaluate(&request(state(1e6, 400.0, 200.0, 1.5), 729000.0, 0.01, 1.0)).unwrap();
    assert_eq!(result.regime, Regime::Subsonic);
    assert_eq!(result.regime.name(), "subsonic");
    near(result.throat_mach, 2.0 / 3.0, 2e-15);
    near(result.exit_temperature_k, 360.0, 3e-16);
    near(
        result.exit_speed_m_per_s,
        (2.0 / 3.0) * 108000.0_f64.sqrt(),
        2e-15,
    );
    near(
        result.mass_flow_kg_per_s,
        0.10125 * (2.0 / 3.0) * 108000.0_f64.sqrt(),
        2e-15,
    );
    near(result.thrust_n, 4860.0, 3e-15);
    assert!(result.mass_flow_kg_per_s < result.sonic_mass_flow_kg_per_s);
    check_finite(&result);
}

#[test]
fn nasa_mass_flow_relation_matches_a_forward_mach_grid() {
    // Eq. 10 supplies a separate forward-Mach oracle; the implementation
    // instead inverts pressure then evaluates the control-section flux.
    // https://www.grc.nasa.gov/www/k-12/airplane/mflchk.html
    for gamma in [1.2, 1.4, 1.5, 5.0 / 3.0, 2.0, 3.0] {
        let stagnation = state(1e6, 300.0, 287.0, gamma);
        for mach in [0.1_f64, 0.25, 2.0 / 3.0, 0.9, 0.999] {
            let factor = 1.0 + (gamma - 1.0) * mach * mach / 2.0;
            let ratio = factor.powf(-gamma / (gamma - 1.0));
            let expected = 0.7 * 0.002 * 1e6 / 300.0_f64.sqrt()
                * (gamma / 287.0).sqrt()
                * mach
                * factor.powf(-(gamma + 1.0) / (2.0 * (gamma - 1.0)));
            let result = evaluate(&request(stagnation, 1e6 * ratio, 0.002, 0.7)).unwrap();
            assert_eq!(result.regime, Regime::Subsonic);
            near(result.throat_mach, mach, 8e-14);
            near(result.mass_flow_kg_per_s, expected, 8e-14);
            near(result.exit_temperature_k, 300.0 / factor, 2e-15);
            check_finite(&result);
        }
    }
}

#[test]
fn choking_plateau_cd_and_area_have_distinct_physical_effects() {
    let stagnation = state(125000.0, 400.0, 200.0, 1.5);
    let reference = evaluate(&request(stagnation, 50000.0, 0.01, 1.0)).unwrap();
    let lower_back = evaluate(&request(stagnation, 25000.0, 0.01, 1.0)).unwrap();
    assert_eq!(reference.mass_flow_kg_per_s, lower_back.mass_flow_kg_per_s);
    assert_eq!(reference.exit_pressure_pa, lower_back.exit_pressure_pa);
    near(lower_back.thrust_n - reference.thrust_n, 250.0, 2e-15);
    let half_cd = evaluate(&request(stagnation, 50000.0, 0.01, 0.5)).unwrap();
    assert_eq!(half_cd.exit_speed_m_per_s, reference.exit_speed_m_per_s);
    near(
        half_cd.mass_flow_kg_per_s,
        reference.mass_flow_kg_per_s / 2.0,
        0.0,
    );
    near(half_cd.thrust_n, 620.0, 3e-16);
    let doubled = evaluate(&request(stagnation, 50000.0, 0.02, 1.0)).unwrap();
    near(
        doubled.mass_flow_kg_per_s,
        2.0 * reference.mass_flow_kg_per_s,
        0.0,
    );
    near(doubled.thrust_n, 2.0 * reference.thrust_n, 0.0);
    // Bracket the analytical transition outside elementary-function roundoff.
    for back in [
        64000.0 * (1.0 - 64.0 * f64::EPSILON),
        64000.0,
        64000.0 * (1.0 + 64.0 * f64::EPSILON),
    ] {
        let result = evaluate(&request(stagnation, back, 0.01, 1.0)).unwrap();
        if back <= 64000.0 {
            assert_eq!(result.regime, Regime::Choked);
        } else {
            assert_eq!(result.regime, Regime::Subsonic);
            assert!(result.throat_mach < 1.0);
        }
        near(
            result.mass_flow_kg_per_s,
            reference.mass_flow_kg_per_s,
            3e-15,
        );
    }
    let compliant = AreaLaw::linear_compliance(
        Area::new(0.001).unwrap(),
        AreaCompliance::new(1e-7).unwrap(),
        Area::new(0.01).unwrap(),
    )
    .unwrap();
    let opened = evaluate(&Request::new(
        stagnation,
        Pressure::new(50000.0).unwrap(),
        compliant,
        DischargeCoefficient::new(1.0).unwrap(),
    ))
    .unwrap();
    near(opened.effective_area_m2, 0.0085, 3e-16);
    near(
        opened.mass_flow_kg_per_s,
        0.85 * reference.mass_flow_kg_per_s,
        3e-16,
    );
}

#[test]
fn sonic_pressure_rounding_does_not_invent_negative_outlet_thrust() {
    for pressure in [125000.0, 140625.0, 1e6, 101325.0, 123.456] {
        for gamma in [1.2, 1.4, 1.5, 5.0 / 3.0, 1.67, 2.0, 10.0] {
            let stagnation = state(pressure, 400.0, 200.0, gamma);
            let sonic_pressure = pressure * stagnation.critical_pressure_ratio();
            let back = sonic_pressure.next_up();
            let result = evaluate(&request(stagnation, back, 1.0, 1e-30)).unwrap();
            assert!(
                result.exit_pressure_pa >= back && result.thrust_n >= 0.0,
                "p0={pressure:?}, gamma={gamma:?}, back={back:?}, critical={:?}, result={result:?}",
                stagnation.critical_pressure_ratio()
            );
        }
    }
}

#[test]
fn exact_closures_and_adverse_pressure_remain_distinct_from_lost_flow() {
    let stagnation = state(125000.0, 400.0, 200.0, 1.5);
    for (back, area) in [(50000.0, 0.0), (250000.0, 0.0), (125000.0, 0.01)] {
        let result = evaluate(&request(stagnation, back, area, 1.0)).unwrap();
        assert_eq!(result.regime, Regime::NoFlow);
        assert_eq!(result.regime.name(), "no-flow");
        assert_eq!(result.mass_flow_kg_per_s, 0.0);
        assert_eq!(result.throat_mach, 0.0);
        assert_eq!(result.thrust_n, 0.0);
        assert_eq!(result.exit_pressure_pa, back);
        if area > 0.0 {
            assert!(result.sonic_mass_flow_kg_per_s > 0.0);
        }
        check_finite(&result);
    }
    assert_eq!(
        evaluate(&request(stagnation, 250000.0, 0.01, 1.0)),
        Err(FlowError::AdversePressure)
    );
    assert_eq!(
        evaluate(&request(
            stagnation,
            50000.0,
            f64::from_bits(1),
            f64::from_bits(1)
        )),
        Err(FlowError::NoRepresentableFlow)
    );
    assert_eq!(
        evaluate(&request(
            state(f64::from_bits(1), 1.0, 1.0, 1.5),
            f64::MAX,
            0.0,
            1.0
        )),
        Err(FlowError::NoRepresentableFlow)
    );
    let equal = evaluate(&request(
        state(f64::from_bits(1), f64::from_bits(1), 1.0, 1.5),
        f64::from_bits(1),
        0.01,
        1.0,
    ))
    .unwrap();
    assert_eq!(equal.regime, Regime::NoFlow);
    check_finite(&equal);
}

#[test]
fn scaling_preserves_finite_states_despite_unrepresentable_intermediates() {
    // R*T0=8 is exact, despite T0*2 overflowing before division by 2.5.
    let result = evaluate(&request(
        state(125000.0, binary(1023), binary(-1020), 1.5),
        50000.0,
        0.01,
        1.0,
    ))
    .unwrap();
    near(result.exit_temperature_k / binary(1023), 0.8, 3e-16);
    near(result.exit_speed_m_per_s, 9.6_f64.sqrt(), 4e-16);
    near(result.mass_flow_kg_per_s, 100.0 * 9.6_f64.sqrt(), 4e-16);
    near(result.thrust_n, 1100.0, 4e-16);
    for gas in [binary(-1020), binary(1023)] {
        let result = evaluate(&request(
            state(125000.0, 400.0, gas, 1.5),
            50000.0,
            0.01,
            1.0,
        ))
        .unwrap();
        let speed = 480.0_f64.sqrt() * gas.sqrt();
        near(result.exit_speed_m_per_s, speed, 1e-15);
        near(result.mass_flow_kg_per_s, 960.0 / speed, 1e-15);
        near(result.thrust_n, 1100.0, 1e-15);
        check_finite(&result);
    }
    let result = evaluate(&request(
        state(125000.0, 1e300, f64::from_bits(1), 1.5),
        50000.0,
        0.01,
        1.0,
    ))
    .unwrap();
    let speed = 1.2_f64.sqrt() * f64::from_bits(1).sqrt() * 1e150;
    near(result.exit_speed_m_per_s, speed, 1e-15);
    near(result.mass_flow_kg_per_s, 960.0 / speed, 1e-15);
}

#[test]
fn adjacent_pressure_and_subnormal_mach_squares_retain_the_linear_limit() {
    let p0 = 125000.0_f64;
    let back = p0.next_down();
    let expected = 0.01 * (2.0 * (p0 / 80000.0) * (p0 - back)).sqrt();
    for gamma in [1.0_f64.next_up(), 1.5, 1e307, 1e308] {
        let result = evaluate(&request(state(p0, 400.0, 200.0, gamma), back, 0.01, 1.0)).unwrap();
        assert_eq!(result.regime, Regime::Subsonic);
        assert!(result.throat_mach > 0.0);
        near(result.mass_flow_kg_per_s, expected, 2e-14);
        if gamma >= 1e307 {
            assert!(result.throat_mach * result.throat_mach < f64::MIN_POSITIVE);
        }
        check_finite(&result);
    }
}

#[test]
fn fixed_and_triangular_area_histories_have_independent_transfer_anchors() {
    let stagnation = state(125000.0, 400.0, 200.0, 1.5);
    for sample_values in [
        vec![(0.0, 0.01), (0.01, 0.01)],
        vec![(0.0, 0.0), (0.01, 0.01), (0.02, 0.0)],
    ] {
        // Both histories have integral A dt = 0.0001 m^2 s. At this exact
        // gamma=3/2 point rho*=1, cp*T*=192000, u*^2/2=48000 J/kg.
        let input = samples(&sample_values);
        let retained = input.clone();
        let result = integrate_history(
            stagnation,
            Pressure::new(50000.0).unwrap(),
            DischargeCoefficient::new(1.0).unwrap(),
            &input,
        )
        .unwrap();
        let mass = 0.0001 * 96000.0_f64.sqrt();
        near(result.mass_out_kg, mass, 1e-15);
        near(result.enthalpy_out_j, mass * 192000.0, 1e-15);
        near(result.kinetic_energy_out_j, mass * 48000.0, 1e-15);
        near(result.total_enthalpy_out_j, mass * 240000.0, 1e-15);
        near(result.impulse_n_s, 11.0, 1e-15);
        near(result.recoil_impulse_n_s, -11.0, 1e-15);
        near(
            result.enthalpy_out_j + result.kinetic_energy_out_j,
            result.total_enthalpy_out_j,
            1e-15,
        );
        assert_eq!(result.recoil_residual_n_s, 0.0);
        assert_eq!(input, retained);
        assert_eq!(result.samples.len(), input.len());
        assert_eq!(result.stagnation, stagnation);
        assert_eq!(result.back_pressure.get(), 50000.0);
        assert_eq!(result.discharge_coefficient.get(), 1.0);
        assert_eq!(result.claims[0].unit, "N s");
    }
}

#[test]
fn history_bounds_strict_time_and_exact_zero_intervals_are_explicit() {
    let ordinary = state(125000.0, 400.0, 200.0, 1.5);
    let back = Pressure::new(50000.0).unwrap();
    let cd = DischargeCoefficient::new(1.0).unwrap();
    assert_eq!(
        integrate_history(ordinary, back, cd, &[]),
        Err(FlowError::InvalidSampleCount)
    );
    let too_many = (0..=MAX_HISTORY_SAMPLES)
        .map(|index| (index as f64, 0.0))
        .collect::<Vec<_>>();
    assert_eq!(
        integrate_history(ordinary, back, cd, &samples(&too_many)),
        Err(FlowError::InvalidSampleCount)
    );
    assert!(
        integrate_history(
            ordinary,
            back,
            cd,
            &samples(&too_many[..MAX_HISTORY_SAMPLES])
        )
        .is_ok()
    );
    for pair in [[(1.0, 0.01), (1.0, 0.01)], [(1.0, 0.01), (0.0, 0.01)]] {
        assert_eq!(
            integrate_history(ordinary, back, cd, &samples(&pair)),
            Err(FlowError::InvalidTime)
        );
    }
    let single = integrate_history(ordinary, back, cd, &samples(&[(10.0, 0.01)])).unwrap();
    assert_eq!(single.mass_out_kg, 0.0);
    assert_eq!(single.impulse_n_s, 0.0);
    let overflowing_cp = state(125000.0, 1.0, 1e308, 1.0_f64.next_up());
    for (back, area) in [(50000.0, 0.0), (125000.0, 0.01)] {
        let history = integrate_history(
            overflowing_cp,
            Pressure::new(back).unwrap(),
            cd,
            &samples(&[(0.0, area), (1.0, area)]),
        )
        .unwrap();
        assert_eq!(history.mass_out_kg, 0.0);
        assert_eq!(history.enthalpy_out_j, 0.0);
        assert_eq!(history.total_enthalpy_out_j, 0.0);
        assert_eq!(history.kinetic_energy_out_j, 0.0);
        assert_eq!(history.impulse_n_s, 0.0);
    }
    assert_eq!(
        integrate_history(
            ordinary,
            Pressure::new(250000.0).unwrap(),
            cd,
            &samples(&[(0.0, 0.01)])
        ),
        Err(FlowError::AdversePressure)
    );
}

#[test]
fn long_intervals_recover_transport_from_a_subnormal_rate_and_zero_point_thrust() {
    let stagnation = state(1e-9, 1.0, 2e-9, 1.5);
    let history = integrate_history(
        stagnation,
        Pressure::new(4e-10).unwrap(),
        DischargeCoefficient::new(1.0).unwrap(),
        &samples(&[(0.0, 3.2e-319), (1e300, 3.2e-319)]),
    )
    .unwrap();
    let point = &history.samples[0].flow;
    assert!(point.mass_flow_kg_per_s > 0.0 && point.mass_flow_kg_per_s < f64::MIN_POSITIVE);
    assert_eq!(point.thrust_n, 0.0);
    // Integration of an already rounded, constant positive rate is exact q*dt;
    // averaging subnormals must not erase q before the long interval rescales it.
    let mass = point.mass_flow_kg_per_s * 1e300;
    near(history.mass_out_kg, mass, 1e-15);
    near(history.total_enthalpy_out_j, mass * 6e-9, 1e-15);
    let impulse = mass * 2.4e-9_f64.sqrt() + (3.2e-319 * 1e300) * 1.12e-10;
    near(history.impulse_n_s, impulse, 2e-15);
}

#[test]
fn history_refuses_unrepresentable_integrals_and_recovers_finite_cp_products() {
    let cd = DischargeCoefficient::new(1.0).unwrap();
    let back = Pressure::new(50000.0).unwrap();
    assert_eq!(
        integrate_history(
            state(125000.0, 400.0, 200.0, 1.5),
            back,
            cd,
            &samples(&[(0.0, 0.01), (f64::MAX, 0.01)])
        ),
        Err(FlowError::NumericalDomain)
    );
    assert_eq!(
        integrate_history(
            state(125000.0, 1.0, 1e-20, 1.5),
            back,
            cd,
            &samples(&[(0.0, 0.01), (f64::from_bits(1), 0.01)])
        ),
        Err(FlowError::NumericalDomain)
    );
    // cp overflows in isolation; m*cp*T remains finite because T and area are tiny.
    let result = integrate_history(
        state(125000.0, 1e-300, 1e308, 1.0_f64.next_up()),
        back,
        cd,
        &samples(&[(0.0, 1e-10), (1.0, 1e-10)]),
    )
    .unwrap();
    assert!(result.total_enthalpy_out_j > 0.0 && result.total_enthalpy_out_j.is_finite());
    near(
        result.enthalpy_out_j + result.kinetic_energy_out_j,
        result.total_enthalpy_out_j,
        2e-15,
    );
}
