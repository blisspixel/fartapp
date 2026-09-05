//! Construction boundaries and explicit quantity semantics for restriction inputs.

use fart_domain::restriction::*;
use fart_domain::{SpecificGasConstant, Temperature};

#[test]
fn quantities_refuse_invalid_domains_without_unchecked_construction() {
    for value in [f64::NAN, f64::INFINITY, f64::NEG_INFINITY] {
        assert_eq!(Pressure::new(value), Err(FlowError::NonFinite));
        assert_eq!(Area::new(value), Err(FlowError::NonFinite));
        assert_eq!(AreaCompliance::new(value), Err(FlowError::NonFinite));
        assert_eq!(HeatCapacityRatio::new(value), Err(FlowError::NonFinite));
        assert_eq!(DischargeCoefficient::new(value), Err(FlowError::NonFinite));
        assert_eq!(Seconds::new(value), Err(FlowError::InvalidTime));
    }
    for value in [-1.0, 0.0] {
        assert_eq!(Pressure::new(value), Err(FlowError::NonPositive));
        assert_eq!(
            HeatCapacityRatio::new(value),
            Err(FlowError::InvalidHeatCapacityRatio)
        );
        assert_eq!(
            DischargeCoefficient::new(value),
            Err(FlowError::InvalidDischargeCoefficient)
        );
    }
    assert_eq!(
        HeatCapacityRatio::new(1.0),
        Err(FlowError::InvalidHeatCapacityRatio)
    );
    assert_eq!(
        DischargeCoefficient::new(1.1),
        Err(FlowError::InvalidDischargeCoefficient)
    );
    assert_eq!(Area::new(-1.0), Err(FlowError::NegativeArea));
    assert_eq!(
        AreaCompliance::new(-1.0),
        Err(FlowError::NegativeCompliance)
    );
    assert_eq!(Seconds::new(-1.0), Err(FlowError::InvalidTime));
    assert_eq!(
        Pressure::new(f64::from_bits(1)).unwrap().get(),
        f64::from_bits(1)
    );
    assert_eq!(
        DischargeCoefficient::new(f64::from_bits(1)).unwrap().get(),
        f64::from_bits(1)
    );
}

#[test]
fn compliance_is_bounded_and_zero_compliance_retains_legacy_projection() {
    let rest = Area::new(0.001).unwrap();
    let cap = Area::new(0.01).unwrap();
    let compliance = AreaCompliance::new(1e-7).unwrap();
    let law = AreaLaw::linear_compliance(rest, compliance, cap).unwrap();
    assert_eq!(law.name(), "linear-compliance");
    assert_eq!(law.prescribed_area(), rest);
    assert_eq!(law.compliance(), compliance);
    assert_eq!(law.maximum(), cap);
    assert_eq!(law.effective(-75000.0).unwrap(), rest);
    assert!((law.effective(75000.0).unwrap().get() - 0.0085).abs() < 1e-17);
    assert_eq!(law.effective(f64::MAX).unwrap(), cap);
    assert_eq!(law.effective(f64::NAN), Err(FlowError::NonFinite));
    assert_eq!(
        AreaLaw::linear_compliance(cap, compliance, rest),
        Err(FlowError::InvalidAreaLaw)
    );
    let zero = Area::new(0.0).unwrap();
    let zero_compliance =
        AreaLaw::linear_compliance(rest, AreaCompliance::new(0.0).unwrap(), cap).unwrap();
    assert_eq!(zero_compliance.name(), "prescribed");
    assert_eq!(zero_compliance.effective(75000.0).unwrap(), rest);
    assert_eq!(AreaLaw::prescribed(zero).effective(1.0).unwrap(), zero);
    let underflow =
        AreaLaw::linear_compliance(zero, AreaCompliance::new(f64::from_bits(1)).unwrap(), cap)
            .unwrap();
    assert_eq!(
        underflow.effective(0.25),
        Err(FlowError::NoRepresentableFlow)
    );
    assert_eq!(underflow.effective(0.0).unwrap(), zero);
    let closed = AreaLaw::linear_compliance(zero, compliance, zero).unwrap();
    assert_eq!(closed.effective(1.0).unwrap(), zero);
}

#[test]
fn typed_request_keeps_every_authored_quantity_and_sonic_limit() {
    let pressure = Pressure::new(125000.0).unwrap();
    let temperature = Temperature::new(400.0).unwrap();
    let gas = SpecificGasConstant::new(200.0).unwrap();
    let gamma = HeatCapacityRatio::new(1.0_f64.next_up()).unwrap();
    let stagnation = Stagnation::new(pressure, temperature, gas, gamma).unwrap();
    assert!((stagnation.critical_pressure_ratio() - (-0.5_f64).exp()).abs() < 2e-16);
    assert_eq!(stagnation.pressure(), pressure);
    assert_eq!(stagnation.temperature(), temperature);
    assert_eq!(stagnation.gas_constant(), gas);
    assert_eq!(stagnation.gamma(), gamma);
    let area = AreaLaw::prescribed(Area::new(0.01).unwrap());
    let cd = DischargeCoefficient::new(0.5).unwrap();
    let request = Request::new(stagnation, pressure, area, cd);
    assert_eq!(request.stagnation(), stagnation);
    assert_eq!(request.back_pressure(), pressure);
    assert_eq!(request.area(), area);
    assert_eq!(request.discharge_coefficient(), cd);
    let sample = HistorySample::new(Seconds::new(0.0).unwrap(), area.prescribed_area());
    assert_eq!(sample.time().get(), 0.0);
    assert_eq!(sample.area(), area.prescribed_area());
}

#[test]
fn every_refusal_has_a_stable_machine_reason() {
    let errors = [
        (FlowError::NonFinite, "nonfinite_quantity"),
        (FlowError::NonPositive, "nonpositive_quantity"),
        (FlowError::NegativeArea, "negative_area"),
        (FlowError::NegativeCompliance, "negative_compliance"),
        (
            FlowError::InvalidDischargeCoefficient,
            "invalid_discharge_coefficient",
        ),
        (
            FlowError::InvalidHeatCapacityRatio,
            "invalid_heat_capacity_ratio",
        ),
        (FlowError::InvalidAreaLaw, "invalid_area_law"),
        (FlowError::InvalidStagnation, "invalid_stagnation"),
        (FlowError::AdversePressure, "adverse_pressure"),
        (FlowError::NoRepresentableFlow, "no_representable_flow"),
        (FlowError::InvalidSampleCount, "invalid_sample_count"),
        (FlowError::InvalidTime, "invalid_time"),
        (FlowError::NumericalDomain, "numerical_domain_error"),
        (FlowError::InvariantViolation, "invariant_violation"),
    ];
    for (error, reason) in errors {
        assert_eq!(error.reason(), reason);
        assert_eq!(error.to_string(), reason);
        assert!(std::error::Error::source(&error).is_none());
    }
}
