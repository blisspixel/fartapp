//! Validated inputs for an experimental rigid ideal-mixture reservoir.
//!
//! These types declare SI quantities, not a general law or case contract.
//! Construction rejects nonfinite values and invalid domains. No public mutable
//! fields or unchecked constructors can bypass those requirements.

use std::fmt;

/// Maximum number of components in one bounded reservoir state.
pub const MAX_COMPONENTS: usize = 64;

/// Validate the bounded lowercase ASCII identifier shared by all model inputs.
pub fn validate_component_id(id: &str) -> Result<(), ModelError> {
    if id.is_empty()
        || id.len() > 128
        || !id.bytes().all(|byte| {
            byte.is_ascii_lowercase() || byte.is_ascii_digit() || b"._:-".contains(&byte)
        })
    {
        return Err(ModelError::InvalidComponentId);
    }
    Ok(())
}

/// A refusal from quantity validation or the analytical model.
#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum ModelError {
    /// A value is NaN or infinite.
    NonFinite,
    /// A quantity that must be positive is zero or negative.
    NonPositive,
    /// The component collection is empty or exceeds the model budget.
    InvalidComponents,
    /// A component identifier is outside the bounded ASCII token grammar.
    InvalidComponentId,
    /// Two components have the same identifier.
    DuplicateComponentId,
    /// Withdrawal is negative or nonfinite.
    InvalidWithdrawal,
    /// Withdrawal would leave no positive retained fraction.
    Exhausted,
    /// Positive withdrawal cannot produce representable positive progress.
    NoRepresentableProgress,
    /// A derived state or transfer is outside the numerical model domain.
    NumericalDomain,
    /// An arithmetic invariant exceeds the declared roundoff allowance.
    InvariantViolation,
}

impl ModelError {
    /// Stable machine-readable reason, independent of localized presentation.
    pub const fn reason(self) -> &'static str {
        match self {
            Self::NonFinite => "nonfinite_quantity",
            Self::NonPositive => "nonpositive_quantity",
            Self::InvalidComponents => "invalid_component_set",
            Self::InvalidComponentId => "invalid_token",
            Self::DuplicateComponentId => "duplicate_component_id",
            Self::InvalidWithdrawal => "invalid_withdrawal",
            Self::Exhausted => "reservoir_depletion",
            Self::NoRepresentableProgress => "no_representable_progress",
            Self::NumericalDomain => "numerical_domain_error",
            Self::InvariantViolation => "invariant_violation",
        }
    }
}

impl fmt::Display for ModelError {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        formatter.write_str(self.reason())
    }
}

impl std::error::Error for ModelError {}

macro_rules! positive_quantity {
    ($name:ident, $description:literal) => {
        #[doc = $description]
        #[derive(Clone, Copy, Debug, PartialEq)]
        pub struct $name(f64);

        impl $name {
            /// Validate a finite, strictly positive value in the documented SI unit.
            pub fn new(value: f64) -> Result<Self, ModelError> {
                if !value.is_finite() {
                    return Err(ModelError::NonFinite);
                }
                if value <= 0.0 {
                    return Err(ModelError::NonPositive);
                }
                Ok(Self(value))
            }

            /// Return the magnitude in the documented SI unit.
            pub const fn get(self) -> f64 {
                self.0
            }
        }
    };
}

positive_quantity!(Mass, "Strictly positive mass in kilograms.");
positive_quantity!(Volume, "Strictly positive volume in cubic metres.");
positive_quantity!(
    Temperature,
    "Strictly positive absolute temperature in kelvin."
);
positive_quantity!(
    SpecificGasConstant,
    "Specific gas constant in joules per kilogram kelvin."
);
positive_quantity!(
    IsochoricHeatCapacity,
    "Specific isochoric heat capacity in joules per kilogram kelvin."
);

/// One homogeneous, calorically perfect ideal-gas component.
#[derive(Clone, Debug, PartialEq)]
pub struct Component {
    id: String,
    mass: Mass,
    gas_constant: SpecificGasConstant,
    heat_capacity: IsochoricHeatCapacity,
}

impl Component {
    /// Construct a component with a 1..=128-byte lowercase ASCII identifier.
    pub fn new(
        id: impl Into<String>,
        mass: Mass,
        gas_constant: SpecificGasConstant,
        heat_capacity: IsochoricHeatCapacity,
    ) -> Result<Self, ModelError> {
        let id = id.into();
        validate_component_id(&id)?;
        Ok(Self {
            id,
            mass,
            gas_constant,
            heat_capacity,
        })
    }

    /// The declared component identifier.
    pub fn id(&self) -> &str {
        &self.id
    }
    /// Retained component mass.
    pub const fn mass(&self) -> Mass {
        self.mass
    }
    /// Component-specific ideal gas constant.
    pub const fn gas_constant(&self) -> SpecificGasConstant {
        self.gas_constant
    }
    /// Constant specific isochoric heat capacity.
    pub const fn heat_capacity(&self) -> IsochoricHeatCapacity {
        self.heat_capacity
    }
}

/// Immutable reservoir inputs, with components ordered by identifier.
#[derive(Clone, Debug, PartialEq)]
pub struct ReservoirState {
    components: Vec<Component>,
    volume: Volume,
    temperature: Temperature,
}

impl ReservoirState {
    /// Validate collection size and unique identifiers, then normalize ordering.
    /// Derived quantities are checked by the core before a prediction is returned.
    pub fn new(
        mut components: Vec<Component>,
        volume: Volume,
        temperature: Temperature,
    ) -> Result<Self, ModelError> {
        if components.is_empty() || components.len() > MAX_COMPONENTS {
            return Err(ModelError::InvalidComponents);
        }
        components.sort_by(|left, right| left.id.cmp(&right.id));
        if components.windows(2).any(|pair| pair[0].id == pair[1].id) {
            return Err(ModelError::DuplicateComponentId);
        }
        Ok(Self {
            components,
            volume,
            temperature,
        })
    }

    /// Components in normalized identifier order, without mutable access.
    pub fn components(&self) -> &[Component] {
        &self.components
    }
    /// The rigid reservoir volume.
    pub const fn volume(&self) -> Volume {
        self.volume
    }
    /// Homogeneous absolute temperature.
    pub const fn temperature(&self) -> Temperature {
        self.temperature
    }
}

/// One of the two explicitly supported rigid-reservoir closures.
#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum Closure {
    /// No heat transfer across the rigid reservoir boundary.
    RigidAdiabatic,
    /// An ideal thermostat supplies heat to hold temperature constant.
    RigidIsothermal,
}

impl Closure {
    /// Stable request/report spelling.
    pub const fn name(self) -> &'static str {
        match self {
            Self::RigidAdiabatic => "rigid-adiabatic",
            Self::RigidIsothermal => "rigid-isothermal",
        }
    }
}

/// A finite nonnegative component-wise withdrawal fraction strictly below one.
#[derive(Clone, Copy, Debug, PartialEq)]
pub struct WithdrawalFraction(f64);

impl WithdrawalFraction {
    /// Validate that positive retained mass is possible in real arithmetic.
    pub fn new(value: f64) -> Result<Self, ModelError> {
        if !value.is_finite() || value < 0.0 {
            return Err(ModelError::InvalidWithdrawal);
        }
        if value >= 1.0 {
            return Err(ModelError::Exhausted);
        }
        Ok(Self(value))
    }
    /// Return the dimensionless fraction.
    pub const fn get(self) -> f64 {
        self.0
    }
}

/// Validated index into the permanent toy intensity table, unrelated to physics.
#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub struct Intensity(u8);

impl Intensity {
    /// Accept only the permanent integer domain 1..=5.
    pub fn new(value: u8) -> Option<Self> {
        (1..=5).contains(&value).then_some(Self(value))
    }
    /// Exact historical stdout, including its terminating newline.
    pub const fn reply(self) -> &'static str {
        const REPLIES: [&str; 5] = [
            "pfft (gentle)\n",
            "toot (respectable)\n",
            "braaap (respectable)\n",
            "blorp (respectable)\n",
            "KABLAM (mighty)\n",
        ];
        REPLIES[(self.0 - 1) as usize]
    }
}
