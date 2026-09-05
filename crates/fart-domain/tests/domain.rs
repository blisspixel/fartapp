//! Quantity, collection, and permanent toy behavior contracts.

use fart_domain::*;

fn component(id: &str) -> Component {
    Component::new(
        id,
        Mass::new(1.0).unwrap(),
        SpecificGasConstant::new(200.0).unwrap(),
        IsochoricHeatCapacity::new(400.0).unwrap(),
    )
    .unwrap()
}

#[test]
fn quantities_refuse_invalid_values_and_retain_units() {
    for value in [0.0, -0.0, -1.0, f64::NAN, f64::INFINITY, f64::NEG_INFINITY] {
        assert!(Mass::new(value).is_err());
        assert!(Volume::new(value).is_err());
        assert!(Temperature::new(value).is_err());
        assert!(SpecificGasConstant::new(value).is_err());
        assert!(IsochoricHeatCapacity::new(value).is_err());
    }
    for value in [f64::from_bits(1), 1.0, f64::MAX] {
        assert_eq!(Mass::new(value).unwrap().get(), value);
        assert_eq!(Volume::new(value).unwrap().get(), value);
        assert_eq!(Temperature::new(value).unwrap().get(), value);
        assert_eq!(SpecificGasConstant::new(value).unwrap().get(), value);
        assert_eq!(IsochoricHeatCapacity::new(value).unwrap().get(), value);
    }
}

#[test]
fn state_normalizes_components_without_exposing_mutation() {
    let state = ReservoirState::new(
        vec![component("b"), component("a")],
        Volume::new(2.0).unwrap(),
        Temperature::new(300.0).unwrap(),
    )
    .unwrap();
    assert_eq!(state.components()[0].id(), "a");
    assert_eq!(state.components()[0].mass().get(), 1.0);
    assert_eq!(state.components()[0].gas_constant().get(), 200.0);
    assert_eq!(state.components()[0].heat_capacity().get(), 400.0);
    assert_eq!(state.volume().get(), 2.0);
    assert_eq!(state.temperature().get(), 300.0);
    let volume = state.volume();
    let temperature = state.temperature();
    assert_eq!(
        ReservoirState::new(vec![], volume, temperature),
        Err(ModelError::InvalidComponents)
    );
    assert_eq!(
        ReservoirState::new(vec![component("a"), component("a")], volume, temperature),
        Err(ModelError::DuplicateComponentId)
    );
    assert_eq!(
        ReservoirState::new(
            vec![component("a"); MAX_COMPONENTS + 1],
            volume,
            temperature
        ),
        Err(ModelError::InvalidComponents)
    );
    let many = (0..MAX_COMPONENTS)
        .map(|index| component(&format!("component.{index}")))
        .collect();
    assert_eq!(
        ReservoirState::new(many, volume, temperature)
            .unwrap()
            .components()
            .len(),
        MAX_COMPONENTS
    );
}

#[test]
fn component_identifiers_use_the_existing_narrow_ascii_grammar() {
    for id in ["", "UPPER", "a b", "a/b", "\u{00e9}", &"a".repeat(129)] {
        assert_eq!(
            Component::new(
                id,
                Mass::new(1.0).unwrap(),
                SpecificGasConstant::new(1.0).unwrap(),
                IsochoricHeatCapacity::new(1.0).unwrap()
            ),
            Err(ModelError::InvalidComponentId)
        );
    }
    assert_eq!(component("._:-09az").id(), "._:-09az");
    assert_eq!(component(&"a".repeat(128)).id().len(), 128);
}

#[test]
fn withdrawal_preserves_zero_and_refuses_exhaustion() {
    for value in [0.0, 0.5, f64::from_bits(1.0_f64.to_bits() - 1)] {
        assert_eq!(WithdrawalFraction::new(value).unwrap().get(), value);
    }
    for value in [-1.0, f64::NAN, f64::INFINITY] {
        assert_eq!(
            WithdrawalFraction::new(value),
            Err(ModelError::InvalidWithdrawal)
        );
    }
    for value in [1.0, 2.0] {
        assert_eq!(WithdrawalFraction::new(value), Err(ModelError::Exhausted));
    }
    assert_eq!(Closure::RigidAdiabatic.name(), "rigid-adiabatic");
    assert_eq!(Closure::RigidIsothermal.name(), "rigid-isothermal");
}

#[test]
fn permanent_intensity_table_is_exact() {
    for (value, expected) in [
        (1, "pfft (gentle)\n"),
        (2, "toot (respectable)\n"),
        (3, "braaap (respectable)\n"),
        (4, "blorp (respectable)\n"),
        (5, "KABLAM (mighty)\n"),
    ] {
        assert_eq!(Intensity::new(value).unwrap().reply(), expected);
    }
    assert!(Intensity::new(0).is_none());
    assert!(Intensity::new(6).is_none());
}

#[test]
fn errors_have_stable_nonempty_machine_reasons() {
    for error in [
        ModelError::NonFinite,
        ModelError::NonPositive,
        ModelError::InvalidComponents,
        ModelError::InvalidComponentId,
        ModelError::DuplicateComponentId,
        ModelError::InvalidWithdrawal,
        ModelError::Exhausted,
        ModelError::NoRepresentableProgress,
        ModelError::NumericalDomain,
        ModelError::InvariantViolation,
    ] {
        assert!(!error.reason().is_empty());
        assert_eq!(error.to_string(), error.reason());
    }
}
