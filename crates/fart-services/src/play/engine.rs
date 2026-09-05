use std::collections::BTreeMap;

use fart_domain::{Closure, ModelError, ReservoirState, WithdrawalFraction};
use serde_json::{Value, json};

use super::{
    MAX_TRANSCRIPT_BYTES, PlayReply, PlayService, Rejection, Transcript, fingerprint, view,
    wire::{Command, Operation},
};

#[derive(Clone, Debug)]
pub(super) struct Session {
    pub baseline: ReservoirState,
    pub actor: String,
    pub session_ref: String,
    pub baseline_fingerprint: String,
    pub account_fingerprint: String,
    pub journal_fingerprint: String,
    pub revision: u32,
    pub attempts: u32,
    pub budget: u32,
    pub finished: bool,
    pub truncated: bool,
    pub last_report: Option<Value>,
    pub genesis: Value,
    pub journal: Vec<Value>,
    receipts: BTreeMap<String, (Vec<u8>, PlayReply)>,
}

impl PlayService {
    pub(super) fn process(&mut self, command: Command) -> Result<PlayReply, Rejection> {
        if let Some(session) = &self.session
            && let Some(key) = command.key()
            && let Some((accepted, reply)) = session.receipts.get(key)
        {
            if *accepted == fingerprint::canonical(&command.value) {
                self.accepted_retries = self.accepted_retries.saturating_add(1);
                return Ok(reply.clone());
            }
            return Err(Rejection::new("idempotency_conflict", "/idempotency_key"));
        }
        if let Operation::Start { state, budget } = &command.operation {
            if self.session.is_some() {
                return Err(Rejection::new("session_already_started", "/operation"));
            }
            if command.revision() != 0 {
                return Err(Rejection::new("stale_revision", "/expected_revision"));
            }
            fart_core::summarize(state)
                .map_err(|error| Rejection::new(error.reason(), "/baseline/initial"))?;
            let (session, reply) = Session::start(&command, state.clone(), *budget);
            self.session = Some(session);
            return Ok(reply);
        }
        let session = self
            .session
            .as_ref()
            .ok_or_else(|| Rejection::new("session_not_started", "/operation"))?;
        if command.text("actor_id") != session.actor {
            return Err(Rejection::new("unauthorized_actor", "/actor_id"));
        }
        if command.text("session_ref") != session.session_ref {
            return Err(Rejection::new("session_reference_mismatch", "/session_ref"));
        }
        match command.operation {
            Operation::Observe { research } => {
                self.observed_commands = self.observed_commands.saturating_add(1);
                return Ok(PlayReply {
                    value: json!({"schema":"fart.reservoir-play-reply/v0alpha1",
                    "status":"observed","observation":view::observation(session,research)}),
                });
            }
            Operation::Actions => {
                self.observed_commands = self.observed_commands.saturating_add(1);
                return Ok(PlayReply {
                    value: json!({"schema":"fart.reservoir-play-reply/v0alpha1",
                    "status":"actions","session_ref":session.session_ref,"current_revision":session.revision,
                    "attempts_remaining":session.budget-session.attempts,"actions":view::actions(session)}),
                });
            }
            _ => {}
        }
        if command.revision() != session.revision {
            return Err(Rejection::new("stale_revision", "/expected_revision"));
        }
        if session.finished {
            return Err(Rejection::new("session_finished", "/operation"));
        }
        if matches!(command.operation, Operation::Predict { .. }) && session.truncated {
            return Err(Rejection::new("attempt_budget_exhausted", "/operation"));
        }
        // A private candidate snapshot keeps receipt retention and admission
        // atomic. It cannot escape as a second public writer.
        let mut candidate = session.clone();
        let report = match &command.operation {
            Operation::Predict { fraction, closure } => {
                Some(evaluate(&candidate.baseline, *fraction, closure))
            }
            _ => None,
        };
        let reply = candidate.accept(&command, report);
        if candidate.transcript().to_json().len() > MAX_TRANSCRIPT_BYTES {
            return Err(Rejection::new("journal_budget_exhausted", "/operation"));
        }
        self.session = Some(candidate);
        Ok(reply)
    }
}

impl Session {
    pub(super) fn start(
        command: &Command,
        baseline: ReservoirState,
        budget: u32,
    ) -> (Self, PlayReply) {
        let baseline_fingerprint = fingerprint::digest("baseline", &command.value["baseline"]);
        let session_ref = fingerprint::digest("session", &command.value);
        let account_fingerprint = initial_account(&baseline_fingerprint);
        let journal_fingerprint = fingerprint::digest(
            "genesis",
            &json!({"session_ref":session_ref,
            "baseline_fingerprint":baseline_fingerprint}),
        );
        let mut session = Self {
            baseline,
            actor: command.text("actor_id").into(),
            session_ref,
            baseline_fingerprint,
            account_fingerprint,
            journal_fingerprint,
            revision: 0,
            attempts: 0,
            budget,
            finished: false,
            truncated: false,
            last_report: None,
            genesis: Value::Null,
            journal: Vec::new(),
            receipts: BTreeMap::new(),
        };
        let reply = session.record(command, "started", 0, None);
        session.genesis = session.journal.pop().expect("start entry");
        (session, reply)
    }

    pub(super) fn accept(&mut self, command: &Command, report: Option<Value>) -> PlayReply {
        self.revision += 1;
        let (outcome, cost) = if let Some(report) = &report {
            self.attempts += 1;
            self.truncated = self.attempts == self.budget;
            if report["status"] == "predicted" {
                self.account_fingerprint = report_account(&self.baseline_fingerprint, report);
                self.last_report = Some(report.clone());
                ("predicted", 1)
            } else {
                ("refused", 1)
            }
        } else {
            self.finished = true;
            ("finished", 0)
        };
        self.record(command, outcome, cost, report)
    }

    fn record(
        &mut self,
        command: &Command,
        outcome: &str,
        cost: u32,
        report: Option<Value>,
    ) -> PlayReply {
        let mut receipt = json!({"operation":command.text("operation"),"outcome":outcome,
            "session_ref":self.session_ref,"revision":self.revision,"attempt_cost":cost,
            "attempts_used":self.attempts,"attempts_remaining":self.budget-self.attempts,
            "finished":self.finished,"truncated":self.truncated,
            "request_fingerprint":fingerprint::digest("request",&command.value),
            "baseline_fingerprint":self.baseline_fingerprint,"account_fingerprint":self.account_fingerprint,
            "previous_journal_fingerprint":self.journal_fingerprint});
        if let Some(report) = report {
            receipt["report"] = report;
        }
        let mut entry = json!({"request":command.value,"receipt":receipt});
        self.journal_fingerprint = fingerprint::digest("journal-entry", &entry);
        receipt["journal_fingerprint"] = json!(self.journal_fingerprint);
        entry["receipt"] = receipt.clone();
        self.journal.push(entry);
        let reply = PlayReply {
            value: json!({"schema":"fart.reservoir-play-reply/v0alpha1","status":"accepted","receipt":receipt}),
        };
        self.receipts.insert(
            command.key().expect("mutating command key").into(),
            (fingerprint::canonical(&command.value), reply.clone()),
        );
        reply
    }

    pub fn summary(&self) -> Value {
        json!({"session_ref":self.session_ref,"revision":self.revision,"complete":self.finished,
            "terminated":self.finished && !self.truncated,"truncated":self.truncated,
            "reason_code":if self.truncated {"attempt_budget_exhausted"} else if self.finished {"explicit_finish"} else {"awaiting_finish"},
            "attempts_used":self.attempts,"attempts_remaining":self.budget-self.attempts,
            "baseline_fingerprint":self.baseline_fingerprint,"account_fingerprint":self.account_fingerprint,
            "journal_fingerprint":self.journal_fingerprint})
    }

    pub fn transcript(&self) -> Transcript {
        Transcript {
            value: json!({"schema":"fart.reservoir-play-transcript/v0alpha1",
            "fingerprint_profile":super::FINGERPRINT_PROFILE,"genesis":self.genesis,
            "journal":self.journal,"summary":self.summary()}),
        }
    }
}

pub(super) fn initial_account(baseline: &str) -> String {
    fingerprint::digest(
        "account",
        &json!({"baseline_fingerprint":baseline,"status":"not-evaluated"}),
    )
}
pub(super) fn report_account(baseline: &str, report: &Value) -> String {
    fingerprint::digest(
        "account",
        &json!({"baseline_fingerprint":baseline,"report":report}),
    )
}

fn evaluate(baseline: &ReservoirState, fraction: f64, closure: &str) -> Value {
    #[cfg(test)]
    EVALUATIONS.with(|count| count.set(count.get() + 1));
    let outcome = WithdrawalFraction::new(fraction)
        .map_err(|error| crate::Diagnostic::new("model", "/withdrawal_fraction", error.reason()))
        .and_then(|withdrawal| {
            let closure = match closure {
                "rigid-adiabatic" => Closure::RigidAdiabatic,
                "rigid-isothermal" => Closure::RigidIsothermal,
                _ => {
                    return Err(crate::Diagnostic::new(
                        "model",
                        "/closure",
                        "unsupported_closure",
                    ));
                }
            };
            fart_core::withdraw_fraction(baseline, withdrawal, closure)
                .map(Box::new)
                .map_err(|error| {
                    let mut issue = crate::Diagnostic::new("model", "/", error.reason());
                    issue.code = if error == ModelError::InvariantViolation {
                        "FART-E-NUMERICAL-0001"
                    } else {
                        "FART-E-MODEL-0002"
                    };
                    issue
                })
        });
    crate::report::json_value(&crate::PredictionReport {
        outcome,
        consulted_inputs: &["authored_baseline", "explicit_action"],
    })
}

#[cfg(test)]
thread_local! { pub(super) static EVALUATIONS: std::cell::Cell<u32> = const { std::cell::Cell::new(0) }; }

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn importing_and_replaying_retained_results_never_enter_prediction_execution() {
        let fixture: Value = serde_json::from_slice(include_bytes!(
            "../../../../testdata/reservoir/synthetic-mixture-adiabatic.json"
        ))
        .unwrap();
        let start = json!({"schema":super::super::COMMAND_SCHEMA,"operation":"start","profile":super::super::PROFILE,
            "actor_id":"operator","role":"operator","session_nonce":"execution-counter","idempotency_key":"start",
            "expected_revision":0,"attempt_budget":1,"measurement_interaction":"none","knowledge_policy":"full-reservoir",
            "termination_policy":"explicit-finish-or-budget","baseline":{"schema":super::super::BASELINE_SCHEMA,
                "model":fixture["model"],"quantity_system":"si","initial":fixture["initial"]}});
        let mut service = PlayService::new();
        let reply = service.process_json(start.to_string().as_bytes());
        let reference = &reply.value["receipt"]["session_ref"];
        let action = json!({"schema":super::super::COMMAND_SCHEMA,"operation":"predict","actor_id":"operator",
            "session_ref":reference,"expected_revision":0,"idempotency_key":"predict","withdrawal_fraction":0.5,"closure":"rigid-isothermal"});
        EVALUATIONS.with(|count| count.set(0));
        assert!(
            !service
                .process_json(action.to_string().as_bytes())
                .is_rejected()
        );
        EVALUATIONS.with(|count| assert_eq!(count.get(), 1));
        let transcript = service.end_of_input().transcript().unwrap().to_json();
        EVALUATIONS.with(|count| count.set(0));
        let replay = Transcript::from_json(transcript.as_bytes())
            .unwrap()
            .replay()
            .unwrap();
        assert!(!replay.is_complete());
        EVALUATIONS.with(|count| assert_eq!(count.get(), 0));
    }
}
