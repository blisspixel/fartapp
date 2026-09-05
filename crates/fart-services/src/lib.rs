//! Bounded native reservoir, restriction, history, and local play adapters.
//!
//! The local experiment session is a narrow service candidate, not a general
//! law capability boundary, ratified case identity, protocol server, or archive.

mod json;
mod parse;
pub mod play;
pub mod presentation;
mod report;
pub mod restriction;

use fart_core::{Transition, withdraw_fraction};
use fart_domain::{Intensity, ModelError};

/// Accepted narrow reservoir request schema, shared with the Go oracle.
pub const REQUEST_SCHEMA: &str = "fart.reservoir-prediction-request/v0alpha1";
/// Versioned report schema; implementation identity remains separate.
pub const REPORT_SCHEMA: &str = "fart.reservoir-prediction/v0alpha1";
/// Narrow model identifier, without a generic law capability claim.
pub const MODEL_ID: &str = "continuum.rigid-calorically-perfect-ideal-mixture";
/// Reviewed equation revision.
pub const MODEL_VERSION: &str = "v0alpha1";
/// Native implementation identity, distinct from the Go oracle.
pub const IMPLEMENTATION_REVISION: &str = "rust-reservoir/v0alpha1";
/// Maximum encoded input bytes, before parsing or allocation of a document tree.
pub const MAX_INPUT_BYTES: usize = 65_536;

/// Structured refusal without an implicit replacement request.
#[derive(Clone, Debug, Eq, PartialEq)]
pub struct Diagnostic {
    /// Stable error-family code.
    pub code: &'static str,
    /// Input, syntax, schema, or model stage.
    pub stage: &'static str,
    /// JSON pointer identifying the refused part of the request.
    pub path: String,
    /// Machine-readable reason without untrusted display text.
    pub reason_code: &'static str,
}

impl Diagnostic {
    pub(crate) fn new(stage: &'static str, path: &str, reason_code: &'static str) -> Self {
        let code = match stage {
            "input" => "FART-E-INPUT-0002",
            "syntax" => "FART-E-JSON-0002",
            "schema" => "FART-E-SCHEMA-0002",
            _ => "FART-E-MODEL-0001",
        };
        Self {
            code,
            stage,
            path: if path.is_empty() {
                "/".to_owned()
            } else {
                path.to_owned()
            },
            reason_code,
        }
    }
}

/// Typed prediction or refusal, with no retained session state.
#[derive(Clone, Debug)]
pub struct PredictionReport {
    outcome: Result<Box<Transition>, Diagnostic>,
    consulted_inputs: &'static [&'static str],
}

impl PredictionReport {
    /// Whether the requested prediction completed with every arithmetic check satisfied.
    pub fn is_predicted(&self) -> bool {
        self.outcome.is_ok()
    }

    /// Inspect the computed account or exact refusal without reparsing JSON.
    pub fn outcome(&self) -> Result<&Transition, &Diagnostic> {
        self.outcome.as_deref()
    }

    /// Serialize one report with bounded finite numbers and no trailing newline.
    pub fn to_json(&self) -> String {
        report::json_value(self).to_string()
    }

    /// Produce a concise scientific report or refusal with a terminating newline.
    pub fn to_text(&self) -> String {
        report::text(self)
    }
}

/// Parse, validate, and predict one bounded document using only its explicit inputs.
pub fn predict_reservoir(data: &[u8]) -> PredictionReport {
    let outcome = if data.len() > MAX_INPUT_BYTES {
        Err(Diagnostic::new("input", "/", "input_too_large"))
    } else {
        parse::request(data).and_then(|request| {
            withdraw_fraction(&request.state, request.withdrawal, request.closure)
                .map(Box::new)
                .map_err(|error| {
                    let mut diagnostic = Diagnostic::new("model", "/", error.reason());
                    diagnostic.code = if error == ModelError::InvariantViolation {
                        "FART-E-NUMERICAL-0001"
                    } else {
                        "FART-E-MODEL-0002"
                    };
                    diagnostic
                })
        })
    };
    PredictionReport {
        outcome,
        consulted_inputs: &["document_bytes"],
    }
}

/// Read failures exposed by the CLI without disclosing a host path.
#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum InputFailure {
    /// The named source does not exist.
    NotFound,
    /// The process cannot read the named source.
    PermissionDenied,
    /// A source could not be read completely.
    Unavailable,
    /// Bounded reading observed more than the maximum input bytes.
    TooLarge,
}

/// Describe an input failure without opening a file or consulting ambient state.
pub fn reservoir_input_failure(failure: InputFailure, reading_stream: bool) -> PredictionReport {
    let reason = match failure {
        InputFailure::NotFound => "input_not_found",
        InputFailure::PermissionDenied => "input_permission_denied",
        InputFailure::Unavailable => "input_unavailable",
        InputFailure::TooLarge => "input_too_large",
    };
    let mut diagnostic = Diagnostic::new("input", "/", reason);
    diagnostic.code = "FART-E-IO-0002";
    PredictionReport {
        outcome: Err(diagnostic),
        consulted_inputs: if reading_stream {
            &["input_stream"]
        } else {
            &["input_source_reference"]
        },
    }
}

/// Preserve the toy table for canonical single-digit intensities only.
pub fn intensity_reply(input: &str) -> Option<&'static str> {
    let bytes = input.as_bytes();
    if bytes.len() != 1 || !bytes[0].is_ascii_digit() {
        return None;
    }
    Intensity::new(bytes[0] - b'0').map(Intensity::reply)
}
