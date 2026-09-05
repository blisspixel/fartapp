//! A bounded, local, single-operator reservoir experiment session.
//!
//! Every prediction uses the immutable authored baseline. Revision and journal
//! order express control-plane order, never elapsed physical time. Fingerprints
//! identify this provisional protocol's retained values, not ratified cases,
//! authenticated actors, fresh nonces, or independently verified physics.

mod engine;
mod fingerprint;
mod transcript;
mod view;
mod wire;

use serde_json::{Value, json};

pub use transcript::{ReplaySummary, Transcript};

/// Strict command schema for this experimental local protocol.
pub const COMMAND_SCHEMA: &str = "fart.reservoir-play-command/v0alpha1";
/// Explicit baseline schema, with no withdrawal or closure defaults.
pub const BASELINE_SCHEMA: &str = "fart.reservoir-play-baseline/v0alpha1";
/// The only supported session profile.
pub const PROFILE: &str = "reservoir-experiment/v0alpha1";
/// Canonical byte and digest profile used for provisional fingerprints.
pub const FINGERPRINT_PROFILE: &str = "fart.play.rfc8785-sha256/v0alpha1";
/// Maximum bytes of one command, before allocating a JSON tree.
pub const MAX_COMMAND_BYTES: usize = 65_536;
/// Maximum bytes of a retained transcript.
pub const MAX_TRANSCRIPT_BYTES: usize = 8 * 1024 * 1024;
/// Maximum admitted prediction attempts; accepted retries spend none.
pub const MAX_ATTEMPTS: u32 = 16;

/// A structural or control-plane refusal that never changes session evidence.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Rejection {
    reason: String,
    path: String,
    current_revision: Option<u32>,
}

impl Rejection {
    pub(super) fn new(reason: &str, path: &str) -> Self {
        Self {
            reason: reason.into(),
            path: path.into(),
            current_revision: None,
        }
    }
    /// Stable machine-readable refusal reason.
    pub fn reason(&self) -> &str {
        &self.reason
    }
    /// JSON pointer of the refused input or control field.
    pub fn path(&self) -> &str {
        &self.path
    }
    /// Revision to observe before submitting a new action after a conflict.
    pub fn current_revision(&self) -> Option<u32> {
        self.current_revision
    }
    /// Serialize the refusal without untrusted prose or filesystem details.
    pub fn to_json(&self) -> String {
        self.value().to_string()
    }
    fn value(&self) -> Value {
        let mut value = json!({"code":"FART-E-PLAY-0001", "reason_code":self.reason, "path":self.path,
            "attempt_cost":0, "recovery":view::recovery(&self.reason)});
        if let Some(revision) = self.current_revision {
            value["current_revision"] = json!(revision);
        }
        value
    }
}

impl From<crate::Diagnostic> for Rejection {
    fn from(issue: crate::Diagnostic) -> Self {
        Self::new(issue.reason_code, &issue.path)
    }
}

impl std::fmt::Display for Rejection {
    fn fmt(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(formatter, "{} at {:?}", self.reason, self.path)
    }
}

impl std::error::Error for Rejection {}

/// One immutable reply. Retrying an accepted envelope returns identical bytes.
#[derive(Clone, Debug)]
pub struct PlayReply {
    value: Value,
}

impl PlayReply {
    /// Serialize one reply without a trailing newline.
    pub fn to_json(&self) -> String {
        self.value.to_string()
    }
    /// Whether the command was rejected before any evidence or budget change.
    pub fn is_rejected(&self) -> bool {
        self.value["status"] == "rejected"
    }
    /// Present the reply for a person without changing its retained account.
    pub fn to_text(&self) -> String {
        view::reply_text(&self.value)
    }
    fn rejected(issue: Rejection) -> Self {
        Self {
            value: json!({"schema":"fart.reservoir-play-reply/v0alpha1", "status":"rejected", "rejection":issue.value()}),
        }
    }
}

/// Pure end-of-input projection, including full retained evidence when started.
#[derive(Clone, Debug)]
pub struct RunSummary {
    value: Value,
    transcript: Option<Transcript>,
}

impl RunSummary {
    /// True only when an explicit finish was accepted, including truncated runs.
    pub fn is_complete(&self) -> bool {
        self.value["complete"] == true
    }
    /// Inspect retained evidence without extending the journal.
    pub fn transcript(&self) -> Option<&Transcript> {
        self.transcript.as_ref()
    }
    /// Serialize the summary and complete retained transcript if one exists.
    pub fn to_json(&self) -> String {
        let mut value = self.value.clone();
        if let Some(transcript) = &self.transcript {
            value["transcript"] = transcript.value.clone();
        }
        value.to_string()
    }
    /// Present completion, remaining budget, and fingerprints for a person.
    pub fn to_text(&self) -> String {
        view::summary_text(&self.value)
    }
}

/// One local writer. This live authority cannot be cloned or restored from a
/// transcript; callers must explicitly start a separate session for new work.
#[derive(Debug, Default)]
pub struct PlayService {
    session: Option<engine::Session>,
    rejected_commands: u32,
    observed_commands: u32,
    accepted_retries: u32,
}

impl PlayService {
    /// Create an unstarted service without ambient state, randomness, or I/O.
    pub fn new() -> Self {
        Self::default()
    }
    /// Parse and process one strict command. Model refusals are accepted,
    /// costed attempts; malformed or unauthorized commands are free refusals.
    pub fn process_json(&mut self, data: &[u8]) -> PlayReply {
        let result = wire::decode(data).and_then(|command| self.process(command));
        match result {
            Ok(reply) => reply,
            Err(mut issue) => {
                self.rejected_commands = self.rejected_commands.saturating_add(1);
                issue.current_revision = self.session.as_ref().map(|session| session.revision);
                PlayReply::rejected(issue)
            }
        }
    }
    /// Produce an immutable EOF summary. EOF never invents a finish action.
    pub fn end_of_input(&self) -> RunSummary {
        let mut value = self.session.as_ref().map_or_else(
            || json!({"complete":false,"reason_code":"session_not_started"}),
            engine::Session::summary,
        );
        value["schema"] = json!("fart.reservoir-play-run-summary/v0alpha1");
        value["review_counters"] = json!({"rejected_commands":self.rejected_commands,
            "observed_commands":self.observed_commands,"accepted_retries":self.accepted_retries});
        RunSummary {
            value,
            transcript: self.session.as_ref().map(engine::Session::transcript),
        }
    }
}
