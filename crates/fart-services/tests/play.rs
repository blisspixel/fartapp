//! Local session semantics, retained evidence, and hostile-input boundaries.

use fart_services::play::*;
use serde_json::{Value, json};
use sha2::{Digest, Sha256};

fn start_request(budget: u32) -> Value {
    let fixture: Value = serde_json::from_slice(include_bytes!(
        "../../../testdata/reservoir/synthetic-mixture-adiabatic.json"
    ))
    .unwrap();
    json!({"schema":COMMAND_SCHEMA,"operation":"start","profile":PROFILE,"actor_id":"operator",
        "role":"operator","session_nonce":"test-1","idempotency_key":"start","expected_revision":0,
        "attempt_budget":budget,"measurement_interaction":"none","knowledge_policy":"full-reservoir",
        "termination_policy":"explicit-finish-or-budget","baseline":{"schema":BASELINE_SCHEMA,
            "model":fixture["model"],"quantity_system":"si","initial":fixture["initial"]}})
}
fn send(service: &mut PlayService, value: &Value) -> Value {
    serde_json::from_str(
        &service
            .process_json(&serde_json::to_vec(value).unwrap())
            .to_json(),
    )
    .unwrap()
}
fn started(budget: u32) -> (PlayService, String) {
    let mut service = PlayService::new();
    let reply = send(&mut service, &start_request(budget));
    assert_eq!(reply["status"], "accepted", "{reply}");
    (
        service,
        reply["receipt"]["session_ref"].as_str().unwrap().to_owned(),
    )
}
fn action(reference: &str, revision: u32, key: &str, fraction: f64) -> Value {
    json!({"schema":COMMAND_SCHEMA,"operation":"predict","actor_id":"operator","session_ref":reference,
        "expected_revision":revision,"idempotency_key":key,"withdrawal_fraction":fraction,"closure":"rigid-adiabatic"})
}
fn finish(reference: &str, revision: u32) -> Value {
    json!({"schema":COMMAND_SCHEMA,"operation":"finish","actor_id":"operator","session_ref":reference,
        "expected_revision":revision,"idempotency_key":"finish"})
}
fn observe(reference: &str, view: &str) -> Value {
    json!({"schema":COMMAND_SCHEMA,"operation":"observe","actor_id":"operator","session_ref":reference,"view":view})
}
fn transcript(service: &PlayService) -> String {
    service.end_of_input().transcript().unwrap().to_json()
}

#[test]
fn every_prediction_uses_authored_baseline_and_views_preserve_evidence() {
    let (mut service, reference) = started(4);
    let initial = send(&mut service, &observe(&reference, "brief"));
    assert_eq!(initial["observation"]["account"]["status"], "not-evaluated");
    assert!(initial["observation"]["account"].get("final").is_none());
    let first = send(&mut service, &action(&reference, 0, "half", 0.5));
    let second = send(&mut service, &action(&reference, 1, "quarter", 0.25));
    assert_eq!(first["receipt"]["report"]["final"]["total_mass_kg"], 2.0);
    assert_eq!(second["receipt"]["report"]["final"]["total_mass_kg"], 3.0);
    assert_eq!(
        first["receipt"]["report"]["initial"],
        second["receipt"]["report"]["initial"]
    );
    let before = transcript(&service);
    for view in ["brief", "research"] {
        let observation = send(&mut service, &observe(&reference, view));
        assert_eq!(observation["observation"]["revision"], 2);
        assert_eq!(
            observation["observation"]["account_fingerprint"],
            second["receipt"]["account_fingerprint"]
        );
        assert_eq!(
            observation["observation"]["account"]["final"]["total_mass_kg"],
            3.0
        );
        assert_eq!(
            observation["observation"]["account"]
                .get("report")
                .is_some(),
            view == "research"
        );
        assert_eq!(transcript(&service), before);
    }
    assert!(!service.end_of_input().is_complete());
}

#[test]
fn accepted_retries_precede_revision_and_terminal_checks_and_never_spend() {
    let (mut service, reference) = started(2);
    let request = action(&reference, 0, "withdraw", 0.25);
    let original = service
        .process_json(&serde_json::to_vec(&request).unwrap())
        .to_json();
    let ended = send(&mut service, &finish(&reference, 1));
    assert_eq!(ended["receipt"]["attempt_cost"], 0);
    let before = transcript(&service);
    for command in [request.clone(), start_request(2)] {
        let retry = send(&mut service, &command);
        assert_eq!(retry["status"], "accepted");
    }
    assert_eq!(
        service
            .process_json(&serde_json::to_vec(&request).unwrap())
            .to_json(),
        original
    );
    assert_eq!(transcript(&service), before);
    for (key, value) in [
        ("actor_id", json!("another")),
        ("expected_revision", json!(1)),
        ("withdrawal_fraction", json!(0.5)),
        ("closure", json!("rigid-isothermal")),
    ] {
        let mut conflict = request.clone();
        conflict[key] = value;
        let reply = send(&mut service, &conflict);
        assert_eq!(reply["rejection"]["reason_code"], "idempotency_conflict");
        assert_eq!(reply["rejection"]["current_revision"], 2);
        assert_eq!(transcript(&service), before);
    }
    assert!(service.end_of_input().is_complete());
}

#[test]
fn rejected_commands_leave_budget_revision_journal_and_account_unchanged() {
    let (mut service, reference) = started(3);
    let before = transcript(&service);
    for (key, value, reason) in [
        ("actor_id", json!("another"), "unauthorized_actor"),
        ("expected_revision", json!(1), "stale_revision"),
        (
            "session_ref",
            json!(format!("sha256:{}", "0".repeat(64))),
            "session_reference_mismatch",
        ),
        ("withdrawal_fraction", Value::Null, "document_shape_invalid"),
        ("closure", json!("BAD"), "invalid_token"),
        ("operation", json!("withdraw"), "unsupported_operation"),
    ] {
        let mut request = action(&reference, 0, "new", 0.2);
        request[key] = value;
        let reply = send(&mut service, &request);
        assert_eq!(reply["rejection"]["reason_code"], reason, "{reply}");
        assert_eq!(reply["rejection"]["attempt_cost"], 0);
        assert_eq!(transcript(&service), before);
    }
    let mut new_start = start_request(3);
    new_start["idempotency_key"] = json!("new-start");
    assert_eq!(
        send(&mut service, &new_start)["rejection"]["reason_code"],
        "session_already_started"
    );
    assert_eq!(transcript(&service), before);
}

#[test]
fn model_refusal_costs_an_attempt_and_retains_the_successful_account() {
    let (mut service, reference) = started(3);
    let success = send(&mut service, &action(&reference, 0, "good", 0.25));
    let refused = send(
        &mut service,
        &action(&reference, 1, "invalid-fraction", 1.0),
    );
    assert_eq!(refused["status"], "accepted");
    assert_eq!(refused["receipt"]["outcome"], "refused");
    assert_eq!(refused["receipt"]["attempt_cost"], 1);
    assert_eq!(refused["receipt"]["report"]["status"], "invalid");
    assert_eq!(
        refused["receipt"]["account_fingerprint"],
        success["receipt"]["account_fingerprint"]
    );
    let mut unsupported = action(&reference, 2, "unsupported", 0.2);
    unsupported["closure"] = json!("rigid-radiative");
    let refused = send(&mut service, &unsupported);
    assert_eq!(
        refused["receipt"]["report"]["diagnostics"][0]["reason_code"],
        "unsupported_closure"
    );
    assert_eq!(refused["receipt"]["truncated"], true);
    assert_eq!(
        send(&mut service, &observe(&reference, "research"))["observation"]["account"]["report"],
        success["receipt"]["report"]
    );
    let complete = send(&mut service, &finish(&reference, 3));
    assert_eq!(complete["receipt"]["finished"], true);
    assert_eq!(complete["receipt"]["truncated"], true);
    let replay = Transcript::from_json(transcript(&service).as_bytes())
        .unwrap()
        .replay()
        .unwrap();
    assert!(replay.is_complete());
    let summary: Value = serde_json::from_str(&replay.to_json()).unwrap();
    assert_eq!(summary["terminated"], false);
    assert_eq!(
        summary["observation"]["account"]["report"],
        success["receipt"]["report"]
    );
}

#[test]
fn exhausted_sessions_offer_finish_and_exact_retry_but_never_invent_finish_at_eof() {
    let (mut service, reference) = started(1);
    let prediction = action(&reference, 0, "one", 0.5);
    send(&mut service, &prediction);
    let before = transcript(&service);
    assert!(!service.end_of_input().is_complete());
    assert_eq!(transcript(&service), before);
    let discovery = json!({"schema":COMMAND_SCHEMA,"operation":"actions","actor_id":"operator","session_ref":reference});
    let actions = send(&mut service, &discovery);
    assert_eq!(actions["actions"]["predict"]["available"], false);
    assert_eq!(actions["actions"]["finish"]["available"], true);
    assert_eq!(actions["actions"]["expected_revision"], 1);
    assert_eq!(
        send(&mut service, &action(&reference, 1, "two", 0.5))["rejection"]["reason_code"],
        "attempt_budget_exhausted"
    );
    assert_eq!(send(&mut service, &prediction)["status"], "accepted");
    assert_eq!(transcript(&service), before);
    send(&mut service, &finish(&reference, 1));
    assert_eq!(
        send(&mut service, &action(&reference, 2, "three", 0.5))["rejection"]["reason_code"],
        "session_finished"
    );
    assert_eq!(
        send(&mut service, &discovery)["actions"]["finish"]["available"],
        false
    );
}

#[test]
fn canonical_requests_normalize_negative_zero_and_authored_component_order_only() {
    let (mut service, reference) = started(2);
    let negative = action(&reference, 0, "zero", -0.0);
    let positive = action(&reference, 0, "zero", 0.0);
    let before = send(&mut service, &negative);
    assert_eq!(before, send(&mut service, &positive));
    let mut reordered = start_request(2);
    reordered["baseline"]["initial"]["components"]
        .as_array_mut()
        .unwrap()
        .reverse();
    assert_eq!(send(&mut service, &reordered)["status"], "accepted");
    let (other, _) = started(2);
    let other_summary: Value = serde_json::from_str(&other.end_of_input().to_json()).unwrap();
    assert_eq!(other_summary["session_ref"], reference);
    let mut changed = start_request(2);
    changed["session_nonce"] = json!("test-2");
    let mut other = PlayService::new();
    let changed = send(&mut other, &changed);
    assert_ne!(changed["receipt"]["session_ref"], reference);
    assert_eq!(
        changed["receipt"]["baseline_fingerprint"],
        other_summary["baseline_fingerprint"]
    );
    assert_eq!(
        changed["receipt"]["account_fingerprint"],
        other_summary["account_fingerprint"]
    );
}

#[test]
fn strict_input_refusals_have_exact_paths_and_terminal_safe_presentation() {
    let cases = [
        (
            "/baseline/schema",
            json!("unknown"),
            "unsupported_profile_value",
        ),
        (
            "/baseline/model/id",
            json!("other"),
            "unsupported_profile_value",
        ),
        (
            "/baseline/initial/components/0/mass_kg",
            json!(-1),
            "nonpositive_quantity",
        ),
        (
            "/baseline/initial/components/0/mass_kg",
            Value::Null,
            "document_shape_invalid",
        ),
        (
            "/baseline/initial/volume_m3",
            Value::Null,
            "document_shape_invalid",
        ),
    ];
    for (path, value, reason) in cases {
        let mut request = start_request(1);
        *request.pointer_mut(path).unwrap() = value;
        let result = send(&mut PlayService::new(), &request);
        assert_eq!(result["rejection"]["path"], path, "{result}");
        assert_eq!(result["rejection"]["reason_code"], reason, "{result}");
    }
    for bytes in [
        b"".as_slice(),
        b"{} {}",
        br#"{"x":1,"\u0078":2}"#,
        b"null",
        b"{",
        br#"{"x":1e999}"#,
    ] {
        assert!(PlayService::new().process_json(bytes).is_rejected());
    }
    let reply = PlayService::new().process_json(br#"{"x\u001b[2J\n":1,"x\u001b[2J\n":2}"#);
    let output = reply.to_text();
    assert!(!output.contains('\u{1b}'));
    assert!(output.contains("\\u{1b}[2J\\n"), "{output:?}");
    assert!(!output.contains("[2J\n"));
    assert!(
        PlayService::new()
            .process_json(&vec![b' '; MAX_COMMAND_BYTES + 1])
            .is_rejected()
    );
    let mut unsupported = start_request(1);
    unsupported["extra"] = json!(1);
    assert!(send(&mut PlayService::new(), &unsupported)["rejection"].is_object());
    for (key, value) in [
        ("attempt_budget", json!(0)),
        ("attempt_budget", json!(17)),
        ("attempt_budget", json!(1.0)),
        ("expected_revision", json!(1)),
        ("role", json!("viewer")),
        ("session_nonce", json!("")),
        ("actor_id", json!("x".repeat(65))),
    ] {
        let mut request = start_request(1);
        request[key] = value;
        assert_eq!(
            send(&mut PlayService::new(), &request)["status"],
            "rejected"
        );
    }
}

#[test]
fn transcripts_roundtrip_refuse_tampering_and_bound_arrays_before_decoding_excess_elements() {
    let (mut service, reference) = started(2);
    send(&mut service, &action(&reference, 0, "first", 0.5));
    send(&mut service, &finish(&reference, 1));
    let bytes = transcript(&service);
    let retained = Transcript::from_json(bytes.as_bytes()).unwrap();
    assert_eq!(retained.to_json(), bytes);
    assert!(retained.replay().unwrap().is_complete());
    for path in [
        "/genesis/receipt/revision",
        "/journal/0/receipt/report/final/pressure_pa",
        "/journal/1/request/expected_revision",
        "/summary/attempts_remaining",
    ] {
        let mut value: Value = serde_json::from_str(&bytes).unwrap();
        *value.pointer_mut(path).unwrap() = json!(99);
        assert!(
            Transcript::from_json(value.to_string().as_bytes()).is_err(),
            "{path}"
        );
    }
    for value in [json!({}), json!({"journal":null}), json!({"journal":[{}]})] {
        assert!(Transcript::from_json(value.to_string().as_bytes()).is_err());
    }
    let hostile = format!(
        "{{\"journal\":[{}[THIS_EXCESS_ELEMENT_IS_NEVER_DECODED",
        "null,".repeat(17)
    );
    let error = Transcript::from_json(hostile.as_bytes()).unwrap_err();
    assert_eq!(error.reason(), "collection_limit_exceeded");
    assert_eq!(error.path(), "/journal/17");
    assert_eq!(error.current_revision(), None);
    assert!(error.to_json().contains("collection_limit_exceeded"));
    assert!(Transcript::from_json(&vec![b' '; MAX_TRANSCRIPT_BYTES + 1]).is_err());
}

fn digest(domain: &str, value: &Value) -> String {
    let envelope = json!({"profile":FINGERPRINT_PROFILE,"domain":domain,"value":value});
    let bytes = serde_json_canonicalizer::to_vec(&envelope).unwrap();
    format!(
        "sha256:{}",
        Sha256::digest(bytes)
            .iter()
            .map(|byte| format!("{byte:02x}"))
            .collect::<String>()
    )
}

fn rehash_single_prediction(value: &mut Value) {
    let account = digest(
        "account",
        &json!({"baseline_fingerprint":value["summary"]["baseline_fingerprint"],"report":value["journal"][0]["receipt"]["report"]}),
    );
    let mut previous = value["genesis"]["receipt"]["journal_fingerprint"].clone();
    for entry in value["journal"].as_array_mut().unwrap() {
        entry["receipt"]["account_fingerprint"] = json!(account);
        entry["receipt"]["previous_journal_fingerprint"] = previous;
        entry["receipt"]
            .as_object_mut()
            .unwrap()
            .remove("journal_fingerprint");
        previous = json!(digest("journal-entry", entry));
        entry["receipt"]["journal_fingerprint"] = previous.clone();
    }
    value["summary"]["account_fingerprint"] = json!(account);
    value["summary"]["journal_fingerprint"] = previous;
}

#[test]
fn integrity_replay_never_claims_rehashed_forged_numerics_are_verified() {
    let (mut service, reference) = started(1);
    send(&mut service, &action(&reference, 0, "first", 0.5));
    send(&mut service, &finish(&reference, 1));
    let mut value: Value = serde_json::from_str(&transcript(&service)).unwrap();
    value["journal"][0]["receipt"]["report"]["final"]["pressure_pa"] = json!(12345.0);
    rehash_single_prediction(&mut value);
    let replay = Transcript::from_json(value.to_string().as_bytes())
        .unwrap()
        .replay()
        .unwrap();
    let projected: Value = serde_json::from_str(&replay.to_json()).unwrap();
    assert_eq!(
        projected["observation"]["account"]["final"]["pressure_pa"],
        12345.0
    );
    assert_eq!(projected["verification"]["integrity"], "verified");
    assert_eq!(projected["verification"]["prediction_recomputed"], false);
    assert_eq!(
        projected["verification"]["numerical_verification"],
        "not-performed"
    );
    assert!(replay.to_text().contains("not recomputed"));
}

#[test]
fn rehashing_cannot_remove_or_upgrade_claims_or_substitute_component_identities() {
    let (mut service, reference) = started(1);
    send(&mut service, &action(&reference, 0, "first", 0.5));
    send(&mut service, &finish(&reference, 1));
    let original: Value = serde_json::from_str(&transcript(&service)).unwrap();
    for (path, replacement, reason) in [
        (
            "/claims/0/status",
            json!("empirically-validated"),
            "retained_claim_mismatch",
        ),
        (
            "/claims/0/id",
            json!("law.everything"),
            "retained_claim_mismatch",
        ),
        (
            "/claims/0/method",
            json!("experimental-confirmation"),
            "retained_claim_mismatch",
        ),
        (
            "/claims/0/residual_unit",
            json!("Pa"),
            "retained_claim_mismatch",
        ),
        (
            "/claims/0/equation_revision",
            json!("other@v1"),
            "retained_claim_mismatch",
        ),
        (
            "/final/components/0/id",
            json!("substituted"),
            "retained_component_mismatch",
        ),
        (
            "/transfers/components/0/id",
            json!("substituted"),
            "retained_component_mismatch",
        ),
        (
            "/balances/components/0/id",
            json!("substituted"),
            "retained_component_mismatch",
        ),
        (
            "/assumptions",
            json!(["arbitrary-law"]),
            "unsupported_retained_report",
        ),
        (
            "/nonclaims/evidence",
            json!([]),
            "unsupported_retained_report",
        ),
        (
            "/implementation_revision",
            json!("future/v1"),
            "unsupported_retained_report",
        ),
        (
            "/quantity_system",
            json!("cgs"),
            "unsupported_retained_report",
        ),
        (
            "/final/temperature_k",
            Value::Null,
            "document_shape_invalid",
        ),
        (
            "/final/temperature_k",
            json!(-1),
            "retained_quantity_invalid",
        ),
        (
            "/claims/0/tolerance",
            json!(-1),
            "retained_claim_inconsistent",
        ),
        (
            "/claims/0/residual",
            json!(100),
            "retained_claim_inconsistent",
        ),
    ] {
        let mut forged = original.clone();
        *forged["journal"][0]["receipt"]["report"]
            .pointer_mut(path)
            .unwrap() = replacement;
        rehash_single_prediction(&mut forged);
        let issue = Transcript::from_json(forged.to_string().as_bytes()).unwrap_err();
        assert_eq!(issue.reason(), reason, "{path}: {issue}");
    }
    let mut forged = original.clone();
    forged["journal"][0]["receipt"]["report"]["claims"]
        .as_array_mut()
        .unwrap()
        .pop();
    rehash_single_prediction(&mut forged);
    assert_eq!(
        Transcript::from_json(forged.to_string().as_bytes())
            .unwrap_err()
            .reason(),
        "retained_claim_mismatch"
    );
    let mut forged = original;
    forged["journal"][0]["receipt"]["report"]["final"]["components"]
        .as_array_mut()
        .unwrap()
        .reverse();
    rehash_single_prediction(&mut forged);
    assert_eq!(
        Transcript::from_json(forged.to_string().as_bytes())
            .unwrap_err()
            .reason(),
        "retained_component_mismatch"
    );
}

#[test]
fn human_views_show_cost_state_and_scientific_units_without_changing_evidence() {
    let mut service = PlayService::new();
    assert!(service.end_of_input().to_text().contains("INCOMPLETE"));
    assert!(service.end_of_input().transcript().is_none());
    let start = service.process_json(start_request(3).to_string().as_bytes());
    assert!(start.to_text().contains("PLAY ACTION STARTED"));
    let start: Value = serde_json::from_str(&start.to_json()).unwrap();
    let reference = start["receipt"]["session_ref"].as_str().unwrap();
    let empty = service.process_json(observe(reference, "brief").to_string().as_bytes());
    assert!(
        empty
            .to_text()
            .contains("No prediction has been evaluated successfully.")
    );
    let reply = service.process_json(action(reference, 0, "good", 0.5).to_string().as_bytes());
    assert!(reply.to_text().contains("Attempt cost: 1"));
    assert!(reply.to_text().contains(
        "Receipt values describe the original acceptance; retries spend no additional attempt."
    ));
    let before = transcript(&service);
    let observed = service.process_json(observe(reference, "research").to_string().as_bytes());
    let output = observed.to_text();
    assert!(output.contains("Final mass: 2 kg"));
    assert!(output.contains("Final temperature:"));
    assert!(output.contains(" Pa"));
    assert!(
        output.contains(
            "Human values use six significant digits; JSON retains full numeric precision."
        )
    );
    assert_eq!(transcript(&service), before);
    let discovery = json!({"schema":COMMAND_SCHEMA,"operation":"actions","actor_id":"operator","session_ref":reference});
    assert!(
        service
            .process_json(discovery.to_string().as_bytes())
            .to_text()
            .contains("Predict available: true (1 attempt)")
    );
    let refused =
        service.process_json(action(reference, 1, "refused", -1.0).to_string().as_bytes());
    assert!(
        refused
            .to_text()
            .contains("Latest successful account preserved")
    );
    let finished = service.process_json(finish(reference, 2).to_string().as_bytes());
    assert!(finished.to_text().contains("PLAY ACTION FINISHED"));
    assert!(
        service
            .end_of_input()
            .to_text()
            .contains("PLAY RUN FINISHED")
    );
}

#[test]
fn all_sixteen_attempts_plus_finish_are_retained_without_idempotency_eviction() {
    let (mut service, reference) = started(MAX_ATTEMPTS);
    let first = action(&reference, 0, "attempt-0", 0.0);
    let original = send(&mut service, &first);
    for index in 1..MAX_ATTEMPTS {
        let reply = send(
            &mut service,
            &action(
                &reference,
                index,
                &format!("attempt-{index}"),
                f64::from(index) / 32.0,
            ),
        );
        assert_eq!(reply["receipt"]["revision"], index + 1);
    }
    send(&mut service, &finish(&reference, MAX_ATTEMPTS));
    assert_eq!(send(&mut service, &first), original);
    let value: Value = serde_json::from_str(&transcript(&service)).unwrap();
    assert_eq!(value["journal"].as_array().unwrap().len(), 17);
    assert!(
        Transcript::from_json(value.to_string().as_bytes())
            .unwrap()
            .replay()
            .unwrap()
            .is_complete()
    );
}

#[test]
fn zero_attempt_finish_and_rejected_start_or_oversized_trees_have_no_hidden_work() {
    let mut unstarted = PlayService::new();
    let reference = format!("sha256:{}", "0".repeat(64));
    assert_eq!(
        send(&mut unstarted, &observe(&reference, "brief"))["rejection"]["reason_code"],
        "session_not_started"
    );
    let mut bad_start = start_request(1);
    bad_start["baseline"]["initial"]["volume_m3"] = json!(f64::from_bits(1));
    assert_eq!(send(&mut unstarted, &bad_start)["status"], "rejected");
    assert!(unstarted.end_of_input().transcript().is_none());
    let (mut service, reference) = started(1);
    let before: Value = serde_json::from_str(&transcript(&service)).unwrap();
    let completed = send(&mut service, &finish(&reference, 0));
    assert_eq!(completed["receipt"]["attempts_used"], 0);
    assert_eq!(completed["receipt"]["truncated"], false);
    assert_eq!(
        completed["receipt"]["account_fingerprint"],
        before["summary"]["account_fingerprint"]
    );
    assert!(
        Transcript::from_json(transcript(&service).as_bytes())
            .unwrap()
            .replay()
            .unwrap()
            .is_complete()
    );
    let before = transcript(&service);
    for value in [
        json!({"array":vec![0;65]}),
        json!({"array":vec![vec![0;64];64]}),
    ] {
        let result = send(&mut service, &value);
        assert!(matches!(
            result["rejection"]["reason_code"].as_str().unwrap(),
            "collection_limit_exceeded" | "node_limit_exceeded"
        ));
        assert_eq!(transcript(&service), before);
    }
    for (key, value) in [
        ("view", json!("cinematic")),
        ("session_ref", json!("invalid")),
    ] {
        let mut command = observe(&reference, "brief");
        command[key] = value;
        assert_eq!(send(&mut service, &command)["status"], "rejected");
        assert_eq!(transcript(&service), before);
    }
}

#[test]
fn checked_in_session_recipe_executes_literal_references_and_replays_retained_evidence() {
    let fixture = include_str!("../../../testdata/play/reservoir-session.jsonl");
    let lines: Vec<_> = fixture.lines().collect();
    assert_eq!(lines.len(), 8);
    assert!(fixture.ends_with('\n'));
    let start: Value = serde_json::from_str(lines[0]).unwrap();
    let reservoir: Value = serde_json::from_slice(include_bytes!(
        "../../../testdata/reservoir/synthetic-mixture-adiabatic.json"
    ))
    .unwrap();
    assert_eq!(start["session_nonce"], "example-1");
    assert_eq!(start["actor_id"], "operator");
    assert_eq!(start["attempt_budget"], 4);
    assert_eq!(start["baseline"]["initial"], reservoir["initial"]);
    let mut service = PlayService::new();
    let mut replies = Vec::new();
    for line in lines {
        assert!(line.len() <= MAX_COMMAND_BYTES);
        let reply = service.process_json(line.as_bytes());
        assert!(!reply.is_rejected(), "{}", reply.to_json());
        replies.push(reply.to_json());
    }
    assert_eq!(replies[1], replies[3]);
    assert_eq!(replies[1], replies[7]);
    let prediction: Value = serde_json::from_str(&replies[4]).unwrap();
    assert_eq!(
        prediction["receipt"]["report"]["closure"],
        "rigid-isothermal"
    );
    assert_eq!(prediction["receipt"]["report"]["withdrawal_fraction"], 0.5);
    let summary = service.end_of_input();
    assert!(summary.is_complete());
    let retained = summary.transcript().unwrap();
    let transcript: Value = serde_json::from_str(&retained.to_json()).unwrap();
    assert_eq!(transcript["journal"].as_array().unwrap().len(), 3);
    assert_eq!(transcript["summary"]["revision"], 3);
    assert_eq!(transcript["summary"]["attempts_used"], 2);
    assert_eq!(transcript["summary"]["attempts_remaining"], 2);
    assert_eq!(transcript["summary"]["truncated"], false);
    assert_eq!(
        transcript["summary"]["account_fingerprint"],
        prediction["receipt"]["account_fingerprint"]
    );
    let replay = Transcript::from_json(retained.to_json().as_bytes())
        .unwrap()
        .replay()
        .unwrap();
    assert!(replay.is_complete());
    let replay: Value = serde_json::from_str(&replay.to_json()).unwrap();
    assert_eq!(
        replay["observation"]["account"]["report"],
        prediction["receipt"]["report"]
    );
    assert_eq!(replay["verification"]["prediction_recomputed"], false);
}
