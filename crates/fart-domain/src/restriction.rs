//! Validated SI inputs for a converging restriction and a frozen-source history.
//!
//! No constructor selects an atmosphere, gas, area, or history implicitly.

use std::fmt;

use crate::{SpecificGasConstant, Temperature};

/// Maximum number of prescribed-area samples in one frozen-source history.
pub const MAX_HISTORY_SAMPLES: usize = 256;

/// Quantity, model-domain, or arithmetic refusal for the restriction model.
#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum FlowError {
    /// A declared quantity is NaN or infinite.
    NonFinite,
    /// A strictly positive quantity is zero or negative.
    NonPositive,
    /// A declared area is negative.
    NegativeArea,
    /// A pressure-area compliance is negative.
    NegativeCompliance,
    /// The discharge coefficient is outside (0, 1].
    InvalidDischargeCoefficient,
    /// The calorically perfect heat-capacity ratio is not greater than one.
    InvalidHeatCapacityRatio,
    /// The maximum opening is smaller than its resting area.
    InvalidAreaLaw,
    /// The declared stagnation state has no representable critical ratio.
    InvalidStagnation,
    /// An open restriction would require unsupported reverse flow.
    AdversePressure,
    /// Positive flow or its reported state cannot be represented finitely.
    NoRepresentableFlow,
    /// A history has no samples or exceeds its sample budget.
    InvalidSampleCount,
    /// Sample times are nonfinite, negative, or not strictly increasing.
    InvalidTime,
    /// A requested integral is outside the finite representable domain.
    NumericalDomain,
    /// A computed arithmetic balance exceeds its roundoff allowance.
    InvariantViolation,
}

impl FlowError {
    /// Stable reason token for adapters; no untrusted request prose is included.
    pub const fn reason(self) -> &'static str {
        match self {
            Self::NonFinite => "nonfinite_quantity",
            Self::NonPositive => "nonpositive_quantity",
            Self::NegativeArea => "negative_area",
            Self::NegativeCompliance => "negative_compliance",
            Self::InvalidDischargeCoefficient => "invalid_discharge_coefficient",
            Self::InvalidHeatCapacityRatio => "invalid_heat_capacity_ratio",
            Self::InvalidAreaLaw => "invalid_area_law",
            Self::InvalidStagnation => "invalid_stagnation",
            Self::AdversePressure => "adverse_pressure",
            Self::NoRepresentableFlow => "no_representable_flow",
            Self::InvalidSampleCount => "invalid_sample_count",
            Self::InvalidTime => "invalid_time",
            Self::NumericalDomain => "numerical_domain_error",
            Self::InvariantViolation => "invariant_violation",
        }
    }
}

impl fmt::Display for FlowError {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        formatter.write_str(self.reason())
    }
}

impl std::error::Error for FlowError {}

macro_rules! quantity {
    ($name:ident, $description:literal, $valid:expr, $reason:ident) => {
        #[doc = $description]
        #[derive(Clone, Copy, Debug, PartialEq)]
        pub struct $name(f64);

        impl $name {
            /// Validate the finite magnitude and the quantity's documented domain.
            pub fn new(value: f64) -> Result<Self, FlowError> {
                if !value.is_finite() {
                    return Err(FlowError::NonFinite);
                }
                if !($valid)(value) {
                    return Err(FlowError::$reason);
                }
                Ok(Self(value))
            }

            /// Magnitude in the documented SI unit, or a dimensionless ratio.
            pub const fn get(self) -> f64 {
                self.0
            }
        }
    };
}

quantity!(
    Pressure,
    "Strictly positive absolute pressure in pascals.",
    |v: f64| v > 0.0,
    NonPositive
);
quantity!(
    Area,
    "Nonnegative restriction area in square metres.",
    |v: f64| v >= 0.0,
    NegativeArea
);
quantity!(
    AreaCompliance,
    "Nonnegative opening compliance in square metres per pascal.",
    |v: f64| v >= 0.0,
    NegativeCompliance
);
quantity!(
    HeatCapacityRatio,
    "Calorically perfect heat-capacity ratio greater than one.",
    |v: f64| v > 1.0,
    InvalidHeatCapacityRatio
);
quantity!(
    DischargeCoefficient,
    "Mass-flow discharge coefficient in (0, 1].",
    |v: f64| v > 0.0 && v <= 1.0,
    InvalidDischargeCoefficient
);

/// Nonnegative prescribed sampling time in seconds; not a reservoir clock.
#[derive(Clone, Copy, Debug, PartialEq)]
pub struct Seconds(f64);

impl Seconds {
    /// Validate a finite nonnegative sample time.
    pub fn new(value: f64) -> Result<Self, FlowError> {
        if !value.is_finite() || value < 0.0 {
            return Err(FlowError::InvalidTime);
        }
        Ok(Self(value))
    }
    /// The sampling time in seconds.
    pub const fn get(self) -> f64 {
        self.0
    }
}

/// Immutable total conditions of one calorically perfect gas.
#[derive(Clone, Copy, Debug, PartialEq)]
pub struct Stagnation {
    pressure: Pressure,
    temperature: Temperature,
    gas_constant: SpecificGasConstant,
    gamma: HeatCapacityRatio,
}

impl Stagnation {
    /// Construct explicit total conditions and validate the critical pressure ratio.
    pub fn new(
        pressure: Pressure,
        temperature: Temperature,
        gas_constant: SpecificGasConstant,
        gamma: HeatCapacityRatio,
    ) -> Result<Self, FlowError> {
        let state = Self {
            pressure,
            temperature,
            gas_constant,
            gamma,
        };
        let critical = state.critical_pressure_ratio();
        if !critical.is_finite() || critical <= 0.0 || critical >= 1.0 {
            return Err(FlowError::InvalidStagnation);
        }
        Ok(state)
    }
    /// Total pressure in the upstream stagnation state.
    pub const fn pressure(self) -> Pressure {
        self.pressure
    }
    /// Total temperature in the upstream stagnation state.
    pub const fn temperature(self) -> Temperature {
        self.temperature
    }
    /// Declared gas constant in J/(kg K).
    pub const fn gas_constant(self) -> SpecificGasConstant {
        self.gas_constant
    }
    /// Declared ratio of isobaric to isochoric heat capacity.
    pub const fn gamma(self) -> HeatCapacityRatio {
        self.gamma
    }
    /// Sonic static-to-total pressure ratio; log1p preserves the gamma -> 1 limit.
    pub fn critical_pressure_ratio(self) -> f64 {
        let gamma = self.gamma.get();
        (-(gamma / (gamma - 1.0)) * ((gamma - 1.0) / 2.0).ln_1p()).exp()
    }
}

/// Prescribed area or capped linear quasi-static opening law.
#[derive(Clone, Copy, Debug, PartialEq)]
pub struct AreaLaw {
    prescribed: Area,
    compliance: AreaCompliance,
    maximum: Area,
}

impl AreaLaw {
    /// A fixed nonnegative area, including an explicitly closed restriction.
    pub const fn prescribed(area: Area) -> Self {
        Self {
            prescribed: area,
            compliance: AreaCompliance(0.0),
            maximum: area,
        }
    }
    /// A = min(maximum, resting + compliance * max(overpressure, 0)).
    pub fn linear_compliance(
        resting: Area,
        compliance: AreaCompliance,
        maximum: Area,
    ) -> Result<Self, FlowError> {
        if maximum.get() < resting.get() {
            return Err(FlowError::InvalidAreaLaw);
        }
        Ok(Self {
            prescribed: resting,
            compliance,
            maximum,
        })
    }
    /// The existing model's normalized spelling; zero compliance is prescribed.
    pub const fn name(self) -> &'static str {
        if self.compliance.get() == 0.0 {
            "prescribed"
        } else {
            "linear-compliance"
        }
    }
    /// Fixed area or resting area of the compliant opening.
    pub const fn prescribed_area(self) -> Area {
        self.prescribed
    }
    /// Pressure-area compliance, zero for a prescribed area.
    pub const fn compliance(self) -> AreaCompliance {
        self.compliance
    }
    /// Maximum effective area.
    pub const fn maximum(self) -> Area {
        self.maximum
    }
    /// Evaluate signed overpressure in pascals without hiding positive opening underflow.
    pub fn effective(self, overpressure: f64) -> Result<Area, FlowError> {
        if !overpressure.is_finite() {
            return Err(FlowError::NonFinite);
        }
        let opening = self.compliance.get() * overpressure.max(0.0);
        if opening == 0.0
            && overpressure > 0.0
            && self.compliance.get() > 0.0
            && self.prescribed.get() == 0.0
            && self.maximum.get() > 0.0
        {
            return Err(FlowError::NoRepresentableFlow);
        }
        Area::new((self.prescribed.get() + opening).min(self.maximum.get()))
    }
}

/// Complete, explicitly authored restriction input. No numerical work is implicit.
#[derive(Clone, Copy, Debug, PartialEq)]
pub struct Request {
    stagnation: Stagnation,
    back: Pressure,
    area: AreaLaw,
    cd: DischargeCoefficient,
}

impl Request {
    /// Combine validated inputs; open adverse pressure is refused by evaluation.
    pub const fn new(
        stagnation: Stagnation,
        back: Pressure,
        area: AreaLaw,
        cd: DischargeCoefficient,
    ) -> Self {
        Self {
            stagnation,
            back,
            area,
            cd,
        }
    }
    /// Explicit upstream total conditions.
    pub const fn stagnation(self) -> Stagnation {
        self.stagnation
    }
    /// Explicit exterior static pressure.
    pub const fn back_pressure(self) -> Pressure {
        self.back
    }
    /// Authored opening law.
    pub const fn area(self) -> AreaLaw {
        self.area
    }
    /// Mass-flow multiplier, with no change to exit kinematics or pressure thrust.
    pub const fn discharge_coefficient(self) -> DischargeCoefficient {
        self.cd
    }
}

/// One immutable time and prescribed area in a frozen-source history.
#[derive(Clone, Copy, Debug, PartialEq)]
pub struct HistorySample {
    time: Seconds,
    area: Area,
}

impl HistorySample {
    /// Combine validated quantities; history integration checks strict time ordering.
    pub const fn new(time: Seconds, area: Area) -> Self {
        Self { time, area }
    }
    /// Authored sample time.
    pub const fn time(self) -> Seconds {
        self.time
    }
    /// Authored area at that time.
    pub const fn area(self) -> Area {
        self.area
    }
}
