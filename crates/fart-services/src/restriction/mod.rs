//! Experimental native restriction-flow and prescribed-area history adapters.
//!
//! These read-only services retain their own versioned wire reports. They do
//! not couple reservoir depletion, create cases, or certify empirical physics.

mod parse;
mod report;
mod wire;

use fart_core::restriction::{FlowResult, HistoryResult};

use crate::{Diagnostic, InputFailure};

/// Instantaneous restriction request schema, shared with the Go oracle.
pub const REQUEST_SCHEMA: &str = "fart.restriction-prediction-request/v0alpha1";
/// Instantaneous restriction report schema.
pub const REPORT_SCHEMA: &str = "fart.restriction-prediction/v0alpha1";
/// Prescribed-history request schema.
pub const HISTORY_REQUEST_SCHEMA: &str = "fart.restriction-history-request/v0alpha1";
/// Prescribed-history report schema.
pub const HISTORY_REPORT_SCHEMA: &str = "fart.restriction-history/v0alpha1";
/// The explicit, narrow restriction model shared by these operations.
pub const MODEL_ID: &str = "continuum.quasi-steady-isentropic-converging-restriction";
/// Reviewed equation revision.
pub const MODEL_VERSION: &str = "v0alpha1";
/// Native instantaneous implementation, distinct from the Go oracle.
pub const IMPLEMENTATION_REVISION: &str = "rust-restriction/v0alpha1";
/// Native prescribed-history implementation, distinct from the Go oracle.
pub const HISTORY_IMPLEMENTATION_REVISION: &str = "rust-restriction-history/v0alpha1";
/// Maximum encoded bytes before parsing either request.
pub const MAX_INPUT_BYTES: usize = 65_536;

/// Select the operation's report schema and diagnostic family for I/O failures.
#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum Kind {
    /// One instantaneous restriction prediction.
    Prediction,
    /// Frozen-stagnation integration of prescribed-area samples.
    History,
}

#[derive(Clone, Debug)]
enum Outcome {
    Prediction(Box<FlowResult>),
    History(Box<HistoryResult>),
    Invalid(Diagnostic),
}

/// Immutable typed result or exact refusal. Formatting cannot change evidence.
#[derive(Clone, Debug)]
pub struct Report {
    kind: Kind,
    outcome: Outcome,
    consulted_inputs: &'static [&'static str],
}

impl Report {
    /// Whether the explicit operation succeeded with all report claims satisfied.
    pub fn is_predicted(&self) -> bool {
        !matches!(self.outcome, Outcome::Invalid(_))
    }
    /// Inspect the typed instantaneous result if this was a successful prediction.
    pub fn prediction_result(&self) -> Option<&FlowResult> {
        if let Outcome::Prediction(value) = &self.outcome {
            Some(value)
        } else {
            None
        }
    }
    /// Inspect the typed history if this was a successful integration.
    pub fn history_result(&self) -> Option<&HistoryResult> {
        if let Outcome::History(value) = &self.outcome {
            Some(value)
        } else {
            None
        }
    }
    /// Inspect the exact refusal without parsing human text.
    pub fn diagnostic(&self) -> Option<&Diagnostic> {
        if let Outcome::Invalid(value) = &self.outcome {
            Some(value)
        } else {
            None
        }
    }
    /// Serialize the versioned report without a trailing newline.
    pub fn to_json(&self) -> String {
        report::value(self).to_string()
    }
    /// Present a concise report or escaped refusal with a terminating newline.
    pub fn to_text(&self) -> String {
        report::text(self)
    }
}

/// Predict one instantaneous restriction from explicit bounded JSON inputs.
pub fn predict(data: &[u8]) -> Report {
    let outcome = parse::prediction(data).and_then(|request| {
        fart_core::restriction::evaluate(&request)
            .map(Box::new)
            .map(Outcome::Prediction)
            .map_err(|error| {
                let code = if error == fart_domain::restriction::FlowError::InvariantViolation {
                    "FART-E-NUMERICAL-0002"
                } else {
                    "FART-E-MODEL-0004"
                };
                diagnostic(code, "model", "/", error.reason())
            })
    });
    complete(Kind::Prediction, outcome)
}

/// Integrate a prescribed-area history with an explicitly frozen stagnation state.
pub fn history(data: &[u8]) -> Report {
    let outcome = parse::history(data).and_then(|request| {
        fart_core::restriction::integrate_history(
            request.stagnation,
            request.back,
            request.coefficient,
            &request.samples,
        )
        .map(Box::new)
        .map(Outcome::History)
        .map_err(|error| {
            if error == fart_domain::restriction::FlowError::InvariantViolation {
                return diagnostic("FART-E-NUMERICAL-0003", "model", "/", "invariant_violation");
            }
            let reason = match error.reason() {
                "invalid_time" => "invalid_time",
                "invalid_sample_count" => "invalid_sample_count",
                "adverse_pressure" => "adverse_pressure",
                _ => "numerical_domain_error",
            };
            diagnostic("FART-E-MODEL-0005", "model", "/samples", reason)
        })
    });
    complete(Kind::History, outcome)
}

fn complete(kind: Kind, outcome: Result<Outcome, Diagnostic>) -> Report {
    let report = Report {
        kind,
        outcome: outcome.unwrap_or_else(Outcome::Invalid),
        consulted_inputs: &["document_bytes"],
    };
    // Report evidence is checked independently of the core's result projection.
    if report.is_predicted() && !report::claims_valid(&report) {
        let code = match kind {
            Kind::Prediction => "FART-E-NUMERICAL-0002",
            Kind::History => "FART-E-NUMERICAL-0003",
        };
        return Report {
            kind,
            outcome: Outcome::Invalid(diagnostic(code, "model", "/", "invariant_violation")),
            consulted_inputs: &["document_bytes"],
        };
    }
    report
}

/// Describe a bounded CLI I/O failure without opening a path or reading ambient state.
pub fn input_failure(kind: Kind, failure: InputFailure, reading_stream: bool) -> Report {
    let reason = match failure {
        InputFailure::NotFound => "input_not_found",
        InputFailure::PermissionDenied => "input_permission_denied",
        InputFailure::Unavailable => "input_unavailable",
        InputFailure::TooLarge => "input_too_large",
    };
    let code = match kind {
        Kind::Prediction => "FART-E-IO-0003",
        Kind::History => "FART-E-IO-0004",
    };
    Report {
        kind,
        outcome: Outcome::Invalid(diagnostic(code, "input", "/", reason)),
        consulted_inputs: if reading_stream {
            &["input_stream"]
        } else {
            &["input_source_reference"]
        },
    }
}

fn diagnostic(
    code: &'static str,
    stage: &'static str,
    path: &str,
    reason: &'static str,
) -> Diagnostic {
    Diagnostic {
        code,
        stage,
        path: if path.is_empty() {
            "/".into()
        } else {
            path.into()
        },
        reason_code: reason,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn report_guard_refuses_failed_or_nonfinite_retained_claims() {
        for kind in [Kind::Prediction, Kind::History] {
            let valid = match kind {
                Kind::Prediction => predict(include_bytes!(
                    "../../../../testdata/restriction/gamma15-choked.json"
                )),
                Kind::History => history(include_bytes!(
                    "../../../../testdata/restriction/gamma15-choked-history.json"
                )),
            };
            for (residual, tolerance) in [
                (1.0, 0.0),
                (f64::NAN, 1.0),
                (0.0, f64::INFINITY),
                (0.0, -1.0),
            ] {
                let mut damaged = valid.clone();
                let claim = match &mut damaged.outcome {
                    Outcome::Prediction(result) => &mut result.claims[0],
                    Outcome::History(result) => &mut result.claims[0],
                    Outcome::Invalid(_) => panic!("reference fixture must predict"),
                };
                claim.residual = residual;
                claim.tolerance = tolerance;
                let guarded = complete(kind, Ok(damaged.outcome));
                assert_eq!(
                    guarded.diagnostic().unwrap().code,
                    if kind == Kind::Prediction {
                        "FART-E-NUMERICAL-0002"
                    } else {
                        "FART-E-NUMERICAL-0003"
                    }
                );
                assert_eq!(
                    guarded.diagnostic().unwrap().reason_code,
                    "invariant_violation"
                );
                assert!(!guarded.is_predicted());
                assert!(!guarded.to_json().contains("\"claims\""));
            }
        }
    }
}
