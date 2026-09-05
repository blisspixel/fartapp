use std::fmt::Write;

use serde_json::{Value, json};

use super::{PROFILE, engine::Session};

pub(super) fn recovery(reason: &str) -> &'static str {
    match reason {
        "stale_revision" => {
            "Observe the current revision, then submit a new action with a new idempotency key."
        }
        "idempotency_conflict" => {
            "Retry the exact accepted envelope, or use a new key for a different action."
        }
        "attempt_budget_exhausted" => {
            "Observe retained results or explicitly finish the truncated session."
        }
        "session_finished" => {
            "Observe retained results or retry an exact previously accepted envelope."
        }
        "session_not_started" => "Start a session with an explicit baseline and policies.",
        "unauthorized_actor" => "Use the actor bound by the local session's start command.",
        "session_reference_mismatch" => "Use the session_ref returned by the start receipt.",
        "session_already_started" => {
            "Use the existing session or start a separate service instance."
        }
        _ => "Correct the indicated input; the rejected command spent no attempt.",
    }
}

pub(super) fn actions(session: &Session) -> Value {
    let predict = !session.finished && !session.truncated;
    json!({"expected_revision":session.revision,"accepted_retries_available":true,
        "predict":{"available":predict,"attempt_cost":1,
            "unavailable_reason":if session.finished {"session_finished"} else if session.truncated {"attempt_budget_exhausted"} else {"none"},
            "arguments":{"withdrawal_fraction":{"type":"number","unit":"1","minimum":0,"exclusive_maximum":1},
                "closure":{"type":"string","supported":["rigid-adiabatic","rigid-isothermal"]}},
            "baseline_policy":"every-prediction-from-authored-baseline","model_refusal_cost":1},
        "finish":{"available":!session.finished,"attempt_cost":0},
        "observe":{"available":true,"attempt_cost":0,"views":["brief","research"]},
        "actions":{"available":true,"attempt_cost":0}})
}

pub(super) fn observation(session: &Session, research: bool) -> Value {
    let mut value = session.summary();
    value["profile"] = json!(PROFILE);
    value["phase"] = json!(if session.finished {
        "finished"
    } else if session.truncated {
        "budget-exhausted"
    } else {
        "active"
    });
    value["view"] = json!(if research { "research" } else { "brief" });
    value["measurement_interaction"] = json!("none");
    value["knowledge_policy"] = json!("full-reservoir");
    value["journal_order"] = json!("control-plane-only-no-source-time");
    value["actions"] = actions(session);
    value["account"] = if let Some(report) = &session.last_report {
        let mut account = json!({"status":"predicted","provenance":"latest-successful-prediction-from-authored-baseline",
            "quantity_system":"si","closure":report["closure"],"withdrawal_fraction":report["withdrawal_fraction"],
            "final":report["final"],"transfers":report["transfers"]});
        if research {
            account["report"] = report.clone();
        }
        account
    } else {
        json!({"status":"not-evaluated","provenance":"authored-baseline","quantity_system":"si"})
    };
    if research {
        value["baseline"] = session.genesis["request"]["baseline"].clone();
    }
    value
}

pub(super) fn reply_text(value: &Value) -> String {
    if value["status"] == "rejected" {
        let issue = &value["rejection"];
        let mut output = format!(
            "PLAY COMMAND REJECTED\n\nReason: {}\nPath: {:?}\nAttempt cost: 0\n",
            text(issue, "reason_code"),
            text(issue, "path")
        );
        if let Some(revision) = issue.get("current_revision") {
            let _ = writeln!(output, "Current revision: {revision}");
        }
        let _ = writeln!(output, "Recovery: {}", text(issue, "recovery"));
        return output;
    }
    if value["status"] == "accepted" {
        let receipt = &value["receipt"];
        let mut output = format!(
            "PLAY ACTION {}\n\nSession: {}\nRevision: {}\nAttempts remaining: {}\nAttempt cost: {}\nAccount: {}\nJournal: {}\n",
            text(receipt, "outcome").to_ascii_uppercase(),
            text(receipt, "session_ref"),
            receipt["revision"],
            receipt["attempts_remaining"],
            receipt["attempt_cost"],
            text(receipt, "account_fingerprint"),
            text(receipt, "journal_fingerprint")
        );
        output.push_str("Receipt values describe the original acceptance; retries spend no additional attempt.\n");
        if receipt["report"]["status"] == "invalid" {
            let diagnostic = &receipt["report"]["diagnostics"][0];
            let _ = writeln!(
                output,
                "Model refusal: {} at {}\nLatest successful account preserved.",
                text(diagnostic, "reason_code"),
                text(diagnostic, "path")
            );
        }
        return output;
    }
    if value["status"] == "observed" {
        let observation = &value["observation"];
        let mut output = format!(
            "PLAY OBSERVATION\n\nSession: {}\nRevision: {}\nPhase: {}\nAttempts remaining: {}\nAccount: {}\n",
            text(observation, "session_ref"),
            observation["revision"],
            text(observation, "phase"),
            observation["attempts_remaining"],
            text(observation, "account_fingerprint")
        );
        if observation["account"]["status"] == "not-evaluated" {
            output.push_str("No prediction has been evaluated successfully.\n");
        } else {
            let final_state = &observation["account"]["final"];
            let _ = writeln!(
                output,
                "Final mass: {} kg\nFinal temperature: {} K\nFinal pressure: {} Pa\n",
                crate::presentation::human_number(
                    final_state["total_mass_kg"]
                        .as_f64()
                        .expect("typed report number")
                ),
                crate::presentation::human_number(
                    final_state["temperature_k"]
                        .as_f64()
                        .expect("typed report number")
                ),
                crate::presentation::human_number(
                    final_state["pressure_pa"]
                        .as_f64()
                        .expect("typed report number")
                )
            );
            output.push_str(
                "Human values use six significant digits; JSON retains full numeric precision.\n",
            );
        }
        output.push_str(
            "Every prediction uses the authored baseline. Journal order is not elapsed time.\n",
        );
        return output;
    }
    format!(
        "PLAY ACTIONS\n\nCurrent revision: {}\nPredict available: {} (1 attempt)\nFinish available: {} (0 attempts)\nObserve and actions are read-only.\n",
        value["current_revision"],
        value["actions"]["predict"]["available"],
        value["actions"]["finish"]["available"]
    )
}

pub(super) fn summary_text(value: &Value) -> String {
    let mut output = format!(
        "PLAY RUN {}\n\nReason: {}\n",
        if value["complete"] == true {
            "FINISHED"
        } else {
            "INCOMPLETE"
        },
        text(value, "reason_code")
    );
    if let Some(session) = value.get("session_ref") {
        let _ = writeln!(
            output,
            "Session: {}\nRevision: {}\nTruncated: {}\nAttempts remaining: {}\nAccount: {}\nJournal: {}",
            session.as_str().unwrap_or(""),
            value["revision"],
            value["truncated"],
            value["attempts_remaining"],
            text(value, "account_fingerprint"),
            text(value, "journal_fingerprint")
        );
    }
    output
}

fn text<'a>(value: &'a Value, key: &str) -> &'a str {
    value[key].as_str().unwrap_or("")
}
