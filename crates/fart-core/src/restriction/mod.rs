//! Pure quasi-steady converging-restriction flow and prescribed-area histories.
//!
//! The gas is calorically perfect and isentropic to the control section. There
//! is no reverse flow, reservoir coupling, diverging nozzle, shock, or acoustic
//! model. A frozen-source history integrates prescribed samples, not depletion.

mod flow;
mod history;

pub use flow::{FlowResult, evaluate};
pub use history::{HistoryInstant, HistoryResult, integrate_history};

/// Flow regime of the converging control section, without a plume inference.
#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum Regime {
    /// Authored zero area or equal source and back pressure.
    NoFlow,
    /// An open, forward-flowing control section with Mach below one.
    Subsonic,
    /// A sonic control section with mass flow independent of lower back pressure.
    Choked,
}

impl Regime {
    /// Stable report spelling shared by instantaneous and history accounts.
    pub const fn name(self) -> &'static str {
        match self {
            Self::NoFlow => "no-flow",
            Self::Subsonic => "subsonic",
            Self::Choked => "choked",
        }
    }
}
