//! Independent analytical anchors and numerical-domain refusals.

use fart_core::{summarize, withdraw_fraction};
use fart_domain::*;

fn component(id: &str, mass: f64, gas: f64, cv: f64) -> Component {
    Component::new(
        id,
        Mass::new(mass).unwrap(),
        SpecificGasConstant::new(gas).unwrap(),
        IsochoricHeatCapacity::new(cv).unwrap(),
    )
    .unwrap()
}

fn state() -> ReservoirState {
    ReservoirState::new(
        vec![
            component("a", 1.0, 200.0, 400.0),
            component("b", 3.0, 400.0, 800.0),
        ],
        Volume::new(1.0).unwrap(),
        Temperature::new(400.0).unwrap(),
    )
    .unwrap()
}

fn close(actual: f64, expected: f64) {
    assert!(
        (actual - expected).abs() <= 2e-13 * expected.abs() + 2e-14,
        "{actual} != {expected}"
    );
}

#[test]
fn closed_form_gamma_three_halves_anchors() {
    let input = state();
    for (closure, temperature, energy, enthalpy, heat) in [
        (Closure::RigidAdiabatic, 200.0, 140_000.0, 980_000.0, 0.0),
        (
            Closure::RigidIsothermal,
            400.0,
            280_000.0,
            1_260_000.0,
            420_000.0,
        ),
    ] {
        let result =
            withdraw_fraction(&input, WithdrawalFraction::new(0.75).unwrap(), closure).unwrap();
        assert_eq!(input, state());
        assert_eq!(result.before, input);
        assert_eq!(result.closure, closure);
        close(result.initial.total_mass_kg, 4.0);
        close(result.initial.gas_constant, 350.0);
        close(result.initial.heat_cv, 700.0);
        close(result.initial.heat_cp, 1050.0);
        close(result.initial.gamma, 1.5);
        close(result.initial.pressure_pa, 560_000.0);
        close(result.initial.internal_energy_j, 1_120_000.0);
        close(result.final_state.total_mass_kg, 1.0);
        close(result.final_state.temperature_k, temperature);
        close(result.final_state.pressure_pa, 350.0 * temperature);
        close(result.final_state.internal_energy_j, energy);
        close(result.enthalpy_out_j, enthalpy);
        close(result.heat_in_j, heat);
        close(result.components[0].mass_out_kg, 0.75);
        close(result.components[1].mass_out_kg, 2.25);
        close(result.total_mass_out_kg, 3.0);
        assert!(
            result
                .claims
                .iter()
                .all(|claim| claim.residual.abs() <= claim.tolerance)
        );
        assert!(
            result
                .components
                .iter()
                .all(|part| part.residual_kg.abs() < 1e-14)
        );
    }
}

#[test]
fn zero_is_exact_and_sequential_withdrawals_compose() {
    for closure in [Closure::RigidAdiabatic, Closure::RigidIsothermal] {
        let input = state();
        let zero =
            withdraw_fraction(&input, WithdrawalFraction::new(0.0).unwrap(), closure).unwrap();
        assert_eq!(zero.before, zero.after);
        assert_eq!(zero.initial, zero.final_state);
        assert_eq!(
            (zero.total_mass_out_kg, zero.enthalpy_out_j, zero.heat_in_j),
            (0.0, 0.0, 0.0)
        );
        let first =
            withdraw_fraction(&input, WithdrawalFraction::new(0.2).unwrap(), closure).unwrap();
        let second =
            withdraw_fraction(&first.after, WithdrawalFraction::new(0.3).unwrap(), closure)
                .unwrap();
        let direct =
            withdraw_fraction(&input, WithdrawalFraction::new(0.44).unwrap(), closure).unwrap();
        close(
            second.final_state.total_mass_kg,
            direct.final_state.total_mass_kg,
        );
        close(
            second.final_state.temperature_k,
            direct.final_state.temperature_k,
        );
        close(
            second.final_state.pressure_pa,
            direct.final_state.pressure_pa,
        );
        close(
            first.enthalpy_out_j + second.enthalpy_out_j,
            direct.enthalpy_out_j,
        );
    }
}

#[test]
fn finite_parameter_grid_preserves_composition_and_balances() {
    for fraction in [1e-10, 0.01, 0.2, 0.8, 0.99] {
        for scale in [1e-9, 1e-3, 1.0, 1e3, 1e9] {
            for closure in [Closure::RigidAdiabatic, Closure::RigidIsothermal] {
                let input = ReservoirState::new(
                    vec![
                        component("b", 3.0 * scale, 400.0, 800.0),
                        component("a", scale, 200.0, 400.0),
                    ],
                    Volume::new(scale).unwrap(),
                    Temperature::new(400.0).unwrap(),
                )
                .unwrap();
                let result =
                    withdraw_fraction(&input, WithdrawalFraction::new(fraction).unwrap(), closure)
                        .unwrap();
                close(
                    result.final_state.total_mass_kg / result.initial.total_mass_kg,
                    1.0 - fraction,
                );
                close(result.final_state.gas_constant, result.initial.gas_constant);
                assert!(result.final_state.temperature_k <= result.initial.temperature_k);
                assert!(
                    result
                        .claims
                        .iter()
                        .all(|claim| claim.residual.abs() <= claim.tolerance)
                );
            }
        }
    }
}

#[test]
fn unrepresentable_progress_and_derived_overflow_are_refused() {
    assert_eq!(
        withdraw_fraction(
            &state(),
            WithdrawalFraction::new(1e-20).unwrap(),
            Closure::RigidAdiabatic
        )
        .unwrap_err(),
        ModelError::NoRepresentableProgress
    );
    let tiny = ReservoirState::new(
        vec![component("a", f64::from_bits(1), 1.0, 1.0)],
        Volume::new(1.0).unwrap(),
        Temperature::new(1.0).unwrap(),
    )
    .unwrap();
    assert_eq!(
        withdraw_fraction(
            &tiny,
            WithdrawalFraction::new(0.5).unwrap(),
            Closure::RigidIsothermal
        )
        .unwrap_err(),
        ModelError::NoRepresentableProgress
    );
    let huge = ReservoirState::new(
        vec![component("a", 1.0, 10.0, 10.0)],
        Volume::new(1.0).unwrap(),
        Temperature::new(1e308).unwrap(),
    )
    .unwrap();
    assert_eq!(summarize(&huge), Err(ModelError::NonFinite));
    let cold = ReservoirState::new(
        vec![component("a", 1e-308, 1e308, 1.0)],
        Volume::new(1.0).unwrap(),
        Temperature::new(1.0).unwrap(),
    )
    .unwrap();
    assert_eq!(
        withdraw_fraction(
            &cold,
            WithdrawalFraction::new(0.75).unwrap(),
            Closure::RigidAdiabatic
        )
        .unwrap_err(),
        ModelError::NumericalDomain
    );
    let underflow = ReservoirState::new(
        vec![component("a", f64::from_bits(1), f64::from_bits(1), 1.0)],
        Volume::new(1.0).unwrap(),
        Temperature::new(1.0).unwrap(),
    )
    .unwrap();
    assert_eq!(summarize(&underflow), Err(ModelError::NonPositive));
}

#[test]
fn scaled_underflow_regression_matches_exact_power_of_two_account() {
    // All expected answers are exact binary64 values. The original unscaled
    // m*R and mass_out*cp products falsely held pressure at 1 and erased heat.
    let mass = 2.0_f64.powi(-500);
    let property = 2.0_f64.powi(-560);
    let fraction = 2.0_f64.powi(-20);
    for count in [1, 64] {
        let input = ReservoirState::new(
            (0..count)
                .map(|index| {
                    component(
                        &format!("c{index:02}"),
                        mass / count as f64,
                        property,
                        property,
                    )
                })
                .collect(),
            Volume::new(2.0_f64.powi(-60)).unwrap(),
            Temperature::new(2.0_f64.powi(1000)).unwrap(),
        )
        .unwrap();
        let result = withdraw_fraction(
            &input,
            WithdrawalFraction::new(fraction).unwrap(),
            Closure::RigidIsothermal,
        )
        .unwrap();
        assert_eq!(result.initial.pressure_pa, 1.0);
        assert_eq!(result.final_state.pressure_pa, 1.0 - fraction);
        assert_eq!(result.initial.gas_constant, property);
        assert_eq!(result.final_state.gas_constant, property);
        assert_eq!(result.initial.heat_cv, property);
        assert_eq!(result.final_state.heat_cv, property);
        assert_eq!(result.enthalpy_out_j, 2.0_f64.powi(-79));
        assert_eq!(result.heat_in_j, 2.0_f64.powi(-80));
        assert_eq!(
            result.final_state.internal_energy_j,
            (1.0 - fraction) * 2.0_f64.powi(-60)
        );
        assert!(result.claims.iter().all(|claim| claim.residual == 0.0));
    }
}

#[test]
fn scaled_mean_preserves_a_tiny_mass_with_a_large_property() {
    let input = ReservoirState::new(
        vec![
            component(
                "a",
                f64::from_bits(1),
                2.0_f64.powi(1023),
                2.0_f64.powi(1023),
            ),
            component("b", 1.0, 2.0_f64.powi(-52), 2.0_f64.powi(-52)),
        ],
        Volume::new(1.0).unwrap(),
        Temperature::new(1.0).unwrap(),
    )
    .unwrap();
    let result = summarize(&input).unwrap();
    assert_eq!(result.gas_constant, 3.0 * 2.0_f64.powi(-52));
    assert_eq!(result.heat_cv, result.gas_constant);
    assert_eq!(result.pressure_pa, result.gas_constant);
    assert_eq!(result.internal_energy_j, result.gas_constant);
}

#[test]
fn adiabatic_decay_preserves_a_finite_temperature_after_intermediate_underflow() {
    let input = ReservoirState::new(
        vec![component(
            "a",
            2.0_f64.powi(-500),
            2.0_f64.powi(523),
            2.0_f64.powi(513),
        )],
        Volume::new(1.0).unwrap(),
        Temperature::new(2.0_f64.powi(1000)).unwrap(),
    )
    .unwrap();
    let result = withdraw_fraction(
        &input,
        WithdrawalFraction::new(0.75).unwrap(),
        Closure::RigidAdiabatic,
    )
    .unwrap();
    assert_eq!(result.initial.gamma, 1025.0);
    assert_eq!(result.initial.pressure_pa, 2.0_f64.powi(1023));
    assert_eq!(result.initial.internal_energy_j, 2.0_f64.powi(1013));
    assert_eq!(result.final_state.temperature_k, 2.0_f64.powi(-1048));
    assert_eq!(result.final_state.pressure_pa, 2.0_f64.powi(-1027));
    assert_eq!(result.final_state.internal_energy_j, 2.0_f64.powi(-1037));
}

#[test]
fn unrepresentable_heat_and_overflowing_eos_evidence_are_explicit_refusals() {
    let tiny_heat = ReservoirState::new(
        vec![component("a", 1.0, f64::from_bits(1), f64::from_bits(1))],
        Volume::new(1.0).unwrap(),
        Temperature::new(3.0).unwrap(),
    )
    .unwrap();
    assert_eq!(
        withdraw_fraction(
            &tiny_heat,
            WithdrawalFraction::new(0.1).unwrap(),
            Closure::RigidIsothermal
        )
        .unwrap_err(),
        ModelError::NoRepresentableProgress
    );
    let overflowing_eos = ReservoirState::new(
        vec![component(
            "a",
            2.0_f64.powi(-500),
            2.0_f64.powi(533),
            2.0_f64.powi(523),
        )],
        Volume::new(2.0_f64.powi(10)).unwrap(),
        Temperature::new(2.0_f64.powi(1000)).unwrap(),
    )
    .unwrap();
    assert!(summarize(&overflowing_eos).is_ok());
    assert_eq!(
        withdraw_fraction(
            &overflowing_eos,
            WithdrawalFraction::new(0.75).unwrap(),
            Closure::RigidAdiabatic
        )
        .unwrap_err(),
        ModelError::InvariantViolation
    );
    let overflowing_mass = ReservoirState::new(
        vec![
            component("a", f64::MAX, 1.0, 1.0),
            component("b", f64::MAX, 1.0, 1.0),
        ],
        Volume::new(1.0).unwrap(),
        Temperature::new(1.0).unwrap(),
    )
    .unwrap();
    assert_eq!(summarize(&overflowing_mass), Err(ModelError::NonFinite));
}
