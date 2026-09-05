use serde_json::{Value, json};

use super::{PlayService, Rejection, Transcript, fingerprint, wire};

/// Exact comparison of canonical retained values with this implementation's
/// newly computed values. This is not a cross-platform bit-identity promise.
pub const COMPARISON_PROFILE: &str = "fart.play.canonical-current-implementation/v0alpha1";

impl Transcript {
    /// Verify the entire retained chain before doing numerical work, then
    /// freshly admit the baseline and recompute every costed prediction attempt.
    /// The original evidence stays immutable and no live writer is returned.
    pub fn reconstruct(&self) -> Result<ReconstructionSummary, Rejection> {
        self.replay()?;
        let mut service = PlayService::new();
        let mut attempts = 0_u32;
        let mut refusal = None;
        let entries = std::iter::once(&self.value["genesis"])
            .chain(self.value["journal"].as_array().expect("verified journal"));
        for entry in entries {
            let command = wire::command(entry["request"].clone())?;
            if matches!(command.operation, wire::Operation::Predict { .. }) {
                attempts += 1;
            }
            if let Err(issue) = service.process(command) {
                refusal = Some(issue);
                break;
            }
        }
        let fresh = service.end_of_input();
        let difference = fresh
            .transcript()
            .and_then(|transcript| first_difference(&transcript.value, &self.value, ""));
        let matched = refusal.is_none() && difference.is_none();
        let status = if refusal.is_some() {
            "refused"
        } else if matched {
            "matched"
        } else {
            "mismatched"
        };
        let numerical = match status {
            "matched" if attempts == 0 => "no-prediction-attempts",
            "matched" => "matched-current-implementation",
            "mismatched" => "mismatched-current-implementation",
            _ => "reconstruction-refused",
        };
        let mut value = json!({
            "schema":"fart.reservoir-play-reconstruction/v0alpha1",
            "comparison_profile":COMPARISON_PROFILE,
            "implementation_revision":crate::IMPLEMENTATION_REVISION,
            "status":status,
            "prediction_attempts_recomputed":attempts,
            "verification":{"integrity":"verified","control_plane":"verified",
                "baseline_admission":if fresh.transcript().is_some() { "admitted" } else { "refused" },
                "prediction_recomputed":attempts > 0,"numerical_verification":numerical,
                "authentication":"not-established"},
            "retained_summary":self.value["summary"]
        });
        if let Some(transcript) = fresh.transcript() {
            value["reconstructed_transcript"] = transcript.value.clone();
        }
        if let Some(path) = difference {
            value["first_difference"] = json!(path);
        }
        if let Some(issue) = refusal {
            value["refusal"] = issue.value();
        }
        Ok(ReconstructionSummary { value })
    }
}

/// Bounded fresh evidence alongside a retained summary and exact comparison.
#[derive(Clone, Debug)]
pub struct ReconstructionSummary {
    value: Value,
}

impl ReconstructionSummary {
    /// Serialize all fresh reports and fingerprints, preserving both sides of
    /// a mismatch without duplicating the retained transcript.
    pub fn to_json(&self) -> String {
        self.value.to_string()
    }
    /// Whether the complete canonical fresh transcript matched retained values.
    pub fn is_matched(&self) -> bool {
        self.value["status"] == "matched"
    }
    /// Whether the retained journal includes an explicit finish, independently
    /// of whether recomputation matches or the attempt budget was exhausted.
    pub fn is_complete(&self) -> bool {
        self.value["retained_summary"]["complete"] == true
    }
    /// Human projection with comparison scope, completion, and recovery visible.
    pub fn to_text(&self) -> String {
        let mut text = format!(
            "PLAY RECONSTRUCTION {}\n\nRetained integrity: verified\nComparison: {}\nPrediction attempts recomputed: {}\nRetained session complete: {}\n",
            self.value["status"]
                .as_str()
                .expect("owned status")
                .to_uppercase(),
            COMPARISON_PROFILE,
            self.value["prediction_attempts_recomputed"],
            self.is_complete()
        );
        if let Some(path) = self.value.get("first_difference") {
            text.push_str(&format!("First difference: {path}\n"));
        }
        if let Some(refusal) = self.value.get("refusal") {
            text.push_str(&format!(
                "Fresh admission refused: {} at {}\n",
                refusal["reason_code"], refusal["path"]
            ));
        }
        if !self.is_matched() {
            text.push_str("Recovery: preserve the retained transcript and compare the fresh JSON evidence before drawing conclusions.\n");
        }
        text.push_str("\nThis checks current-implementation agreement. It establishes no empirical validation, authentication, or cross-platform bit identity.\n");
        text
    }
}

fn first_difference(fresh: &Value, retained: &Value, path: &str) -> Option<String> {
    if fingerprint::canonical(fresh) == fingerprint::canonical(retained) {
        return None;
    }
    match (fresh, retained) {
        (Value::Object(left), Value::Object(right)) => {
            // Locate the changed report before its derived fingerprints. Every
            // other member is still compared; order only chooses the diagnosis.
            for key in ["genesis", "journal", "request", "report", "receipt"] {
                if let (Some(left), Some(right)) = (left.get(key), right.get(key))
                    && let Some(difference) =
                        first_difference(left, right, &format!("{path}/{key}"))
                {
                    return Some(difference);
                }
            }
            for key in left.keys().chain(right.keys()) {
                let pointer = format!("{path}/{}", key.replace('~', "~0").replace('/', "~1"));
                match (left.get(key), right.get(key)) {
                    (Some(left), Some(right)) => {
                        if let Some(difference) = first_difference(left, right, &pointer) {
                            return Some(difference);
                        }
                    }
                    _ => return Some(pointer),
                }
            }
        }
        (Value::Array(left), Value::Array(right)) => {
            for (index, (left, right)) in left.iter().zip(right).enumerate() {
                if let Some(difference) = first_difference(left, right, &format!("{path}/{index}"))
                {
                    return Some(difference);
                }
            }
            return Some(format!("{path}/{}", left.len().min(right.len())));
        }
        _ => {}
    }
    Some(path.into())
}

#[cfg(test)]
mod tests {
    use super::super::engine::{ADMISSIONS, EVALUATIONS, Session};
    use super::*;

    fn sample() -> Transcript {
        let mut service = PlayService::new();
        for line in include_str!("../../../../testdata/play/reservoir-session.jsonl").lines() {
            assert!(!service.process_json(line.as_bytes()).is_rejected());
        }
        service.end_of_input().transcript().unwrap().clone()
    }

    fn rehash(value: &Value) -> Transcript {
        let command = wire::command(value["genesis"]["request"].clone()).unwrap();
        let wire::Operation::Start { state, budget } = &command.operation else {
            unreachable!()
        };
        let (mut session, _) = Session::start(&command, state.clone(), *budget);
        for entry in value["journal"].as_array().unwrap() {
            let command = wire::command(entry["request"].clone()).unwrap();
            session.accept(&command, entry["receipt"].get("report").cloned());
        }
        session.transcript()
    }

    #[test]
    fn recomputation_matches_complete_and_unfinished_evidence_without_mutation() {
        let sample = sample();
        for count in 0..=sample.value["journal"].as_array().unwrap().len() {
            let mut retained = sample.value.clone();
            retained["journal"].as_array_mut().unwrap().truncate(count);
            let retained = rehash(&retained);
            let before = retained.to_json();
            let predictions = retained.value["journal"]
                .as_array()
                .unwrap()
                .iter()
                .filter(|entry| entry["request"]["operation"] == "predict")
                .count() as u32;
            ADMISSIONS.with(|counter| counter.set(0));
            EVALUATIONS.with(|counter| counter.set(0));
            let result = retained.reconstruct().unwrap();
            ADMISSIONS.with(|counter| assert_eq!(counter.get(), 1));
            EVALUATIONS.with(|counter| assert_eq!(counter.get(), predictions));
            assert!(result.is_matched(), "{}", result.to_json());
            assert_eq!(
                result.is_complete(),
                retained.replay().unwrap().is_complete()
            );
            assert_eq!(result.value["reconstructed_transcript"], retained.value);
            assert_eq!(retained.to_json(), before);
            assert!(
                result
                    .to_text()
                    .contains("current-implementation agreement")
            );
            assert_eq!(
                result.value["verification"]["prediction_recomputed"],
                count > 0
            );
            assert_eq!(
                result.value["verification"]["baseline_admission"],
                "admitted"
            );
            if count == 0 {
                assert_eq!(
                    result.value["verification"]["numerical_verification"],
                    "no-prediction-attempts"
                );
            }
        }
    }

    #[test]
    fn forged_earlier_result_is_detected_even_when_final_account_matches() {
        let original = sample();
        let mut forged = original.value.clone();
        forged["journal"][0]["receipt"]["report"]["final"]["pressure_pa"] = json!(12345.0);
        let forged = rehash(&forged);
        assert!(forged.replay().is_ok());
        let result = forged.reconstruct().unwrap();
        assert!(!result.is_matched());
        assert_eq!(result.value["status"], "mismatched");
        assert_eq!(
            result.value["first_difference"],
            "/journal/0/receipt/report/final/pressure_pa"
        );
        assert_eq!(result.value["reconstructed_transcript"], original.value);
        assert!(result.to_text().contains("Recovery: preserve"));
    }

    #[test]
    fn costed_refusals_are_recomputed_and_compared() {
        let sample = sample();
        let mut service = PlayService::new();
        let start = sample.value["genesis"]["request"].to_string();
        assert!(!service.process_json(start.as_bytes()).is_rejected());
        let mut action = sample.value["journal"][0]["request"].clone();
        action["withdrawal_fraction"] = json!(1.0);
        assert!(
            !service
                .process_json(action.to_string().as_bytes())
                .is_rejected()
        );
        let retained = service.end_of_input().transcript().unwrap().clone();
        ADMISSIONS.with(|counter| counter.set(0));
        EVALUATIONS.with(|counter| counter.set(0));
        let matched = retained.reconstruct().unwrap();
        ADMISSIONS.with(|counter| assert_eq!(counter.get(), 1));
        EVALUATIONS.with(|counter| assert_eq!(counter.get(), 1));
        assert!(matched.is_matched());
        assert_eq!(matched.value["prediction_attempts_recomputed"], 1);
        assert!(!matched.is_complete());
        let mut forged = retained.value;
        forged["journal"][0]["receipt"]["report"]["diagnostics"][0]["reason_code"] =
            json!("unsupported_closure");
        let forged = rehash(&forged);
        assert!(forged.replay().is_ok());
        let result = forged.reconstruct().unwrap();
        assert_eq!(result.value["status"], "mismatched");
        assert!(
            result.value["first_difference"]
                .as_str()
                .unwrap()
                .ends_with("/reason_code")
        );
    }

    #[test]
    fn integrity_only_baseline_can_fail_fresh_numerical_admission_without_predictions() {
        let mut value = sample().value;
        value["journal"] = json!([]);
        value["genesis"]["request"]["baseline"]["initial"]["volume_m3"] = json!(f64::from_bits(1));
        let forged = rehash(&value);
        assert!(forged.replay().is_ok());
        ADMISSIONS.with(|count| count.set(0));
        EVALUATIONS.with(|count| count.set(0));
        let result = forged.reconstruct().unwrap();
        assert_eq!(result.value["status"], "refused");
        assert!(!result.is_matched());
        assert!(!result.is_complete());
        assert_eq!(result.value["prediction_attempts_recomputed"], 0);
        assert!(result.value.get("reconstructed_transcript").is_none());
        assert_eq!(result.value["refusal"]["path"], "/baseline/initial");
        assert_eq!(
            result.value["verification"]["baseline_admission"],
            "refused"
        );
        assert!(result.to_text().contains("Fresh admission refused"));
        ADMISSIONS.with(|count| assert_eq!(count.get(), 1));
        EVALUATIONS.with(|count| assert_eq!(count.get(), 0));
    }

    #[test]
    fn all_retained_profiles_and_integrity_are_checked_before_any_solver_work() {
        let mut retained = sample();
        retained.value["journal"][1]["receipt"]["report"]["implementation_revision"] =
            json!("unknown");
        ADMISSIONS.with(|count| count.set(0));
        EVALUATIONS.with(|count| count.set(0));
        assert!(retained.reconstruct().is_err());
        ADMISSIONS.with(|count| assert_eq!(count.get(), 0));
        EVALUATIONS.with(|count| assert_eq!(count.get(), 0));
    }

    #[test]
    fn first_difference_uses_canonical_numbers_and_escaped_json_pointers() {
        assert_eq!(first_difference(&json!(1), &json!(1.0), ""), None);
        assert_eq!(
            first_difference(&json!({"a":1,"z/":2}), &json!({"a":1.0,"z/":3}), ""),
            Some("/z~1".into())
        );
        assert_eq!(
            first_difference(&json!([]), &json!([1]), "/x"),
            Some("/x/0".into())
        );
        assert_eq!(
            first_difference(&json!({"~":1}), &json!({}), ""),
            Some("/~0".into())
        );
    }
}
