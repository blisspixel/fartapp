use std::collections::BTreeSet;

use serde_json::{Value, json};

use super::{
    FINGERPRINT_PROFILE, MAX_TRANSCRIPT_BYTES, Rejection,
    engine::Session,
    fingerprint, view,
    wire::{self, Operation},
};

/// Immutable retained evidence. Import does not restore live writer authority.
#[derive(Clone, Debug)]
pub struct Transcript {
    pub(super) value: Value,
}

impl Transcript {
    /// Serialize retained values without inventing missing actions or metadata.
    pub fn to_json(&self) -> String {
        self.value.to_string()
    }
    /// Strictly import bounded evidence and verify its hash and control chain.
    /// Predictions are never recomputed during import or replay.
    pub fn from_json(data: &[u8]) -> Result<Self, Rejection> {
        if data.len() > MAX_TRANSCRIPT_BYTES {
            return Err(Rejection::new("input_too_large", "/"));
        }
        let value = crate::json::document(data, crate::json::Limits::TRANSCRIPT)?;
        let transcript = Self { value };
        transcript.replay()?;
        Ok(transcript)
    }
    /// Verify and project retained evidence without invoking the prediction
    /// core. Rehashing forged numerical results can satisfy integrity; this
    /// operation establishes neither physics correctness nor authorship.
    pub fn replay(&self) -> Result<ReplaySummary, Rejection> {
        let root = wire::object(&self.value, "/")?;
        wire::closed(
            root,
            &[
                "schema",
                "fingerprint_profile",
                "genesis",
                "journal",
                "summary",
            ],
            "/",
        )?;
        if root["schema"] != "fart.reservoir-play-transcript/v0alpha1"
            || root["fingerprint_profile"] != FINGERPRINT_PROFILE
        {
            return Err(Rejection::new("unsupported_transcript_profile", "/"));
        }
        let genesis = entry(&root["genesis"], "/genesis")?;
        let start = wire::command(genesis["request"].clone())?;
        let Operation::Start { state, budget } = &start.operation else {
            return Err(Rejection::new("invalid_genesis", "/genesis/request"));
        };
        if start.revision() != 0 {
            return Err(Rejection::new(
                "invalid_genesis",
                "/genesis/request/expected_revision",
            ));
        }
        // Typed baseline validation has no core call. Session::start only
        // derives fingerprints and a receipt, unlike live admission.
        let (mut session, _) = Session::start(&start, state.clone(), *budget);
        equal(&session.genesis, &root["genesis"], "/genesis")?;
        let mut keys = BTreeSet::from([start.key().expect("start key").to_owned()]);
        let journal = root["journal"]
            .as_array()
            .ok_or_else(|| Rejection::new("document_shape_invalid", "/journal"))?;
        if journal.len() > super::MAX_ATTEMPTS as usize + 1 {
            return Err(Rejection::new("collection_limit_exceeded", "/journal"));
        }
        for (index, retained) in journal.iter().enumerate() {
            let path = format!("/journal/{index}");
            let retained_entry = entry(retained, &path)?;
            let command = wire::command(retained_entry["request"].clone())?;
            if !matches!(
                command.operation,
                Operation::Predict { .. } | Operation::Finish
            ) || command.text("actor_id") != session.actor
                || command.text("session_ref") != session.session_ref
                || command.revision() != session.revision
                || session.finished
                || !keys.insert(command.key().unwrap_or("").to_owned())
            {
                return Err(Rejection::new("invalid_journal_control", &path));
            }
            let report = if let Operation::Predict {
                fraction,
                ref closure,
            } = command.operation
            {
                if session.truncated {
                    return Err(Rejection::new("invalid_journal_control", &path));
                }
                let report = retained_entry["receipt"]
                    .get("report")
                    .ok_or_else(|| Rejection::new("document_shape_invalid", &path))?;
                report_shape(report, &format!("{path}/receipt/report"))?;
                if report["status"] == "predicted" {
                    if !(0.0..1.0).contains(&fraction)
                        || !matches!(closure.as_str(), "rigid-adiabatic" | "rigid-isothermal")
                        || report["withdrawal_fraction"].as_f64() != Some(fraction)
                        || report["closure"] != *closure
                    {
                        return Err(Rejection::new("report_request_mismatch", &path));
                    }
                    baseline_binding(report, &session.genesis["request"]["baseline"], &path)?;
                }
                Some(report.clone())
            } else {
                None
            };
            session.accept(&command, report);
            equal(
                session.journal.last().expect("accepted entry"),
                retained,
                &path,
            )?;
        }
        equal(&session.summary(), &root["summary"], "/summary")?;
        let mut value = session.summary();
        value["schema"] = json!("fart.reservoir-play-replay/v0alpha1");
        value["verification"] = json!({"integrity":"verified","control_plane":"verified",
            "prediction_recomputed":false,"numerical_verification":"not-performed","authentication":"not-established"});
        value["observation"] = view::observation(&session, true);
        Ok(ReplaySummary { value })
    }
}

/// Read-only projection of verified retained control-plane evidence.
#[derive(Clone, Debug)]
pub struct ReplaySummary {
    value: Value,
}

impl ReplaySummary {
    /// Serialize integrity results and retained observation.
    pub fn to_json(&self) -> String {
        self.value.to_string()
    }
    /// True only if the retained journal contains an explicit finish.
    pub fn is_complete(&self) -> bool {
        self.value["complete"] == true
    }
    /// Human projection with the numerical nonclaim visible.
    pub fn to_text(&self) -> String {
        format!(
            "PLAY TRANSCRIPT INTEGRITY VERIFIED\n\n{}\nRetained predictions were not recomputed. Numerical verification and authentication are not established.\n",
            view::summary_text(&self.value)
        )
    }
}

fn entry<'a>(
    value: &'a Value,
    path: &str,
) -> Result<&'a serde_json::Map<String, Value>, Rejection> {
    let entry = wire::object(value, path)?;
    wire::closed(entry, &["request", "receipt"], path)?;
    wire::object(&entry["receipt"], path)?;
    Ok(entry)
}

fn equal(expected: &Value, retained: &Value, path: &str) -> Result<(), Rejection> {
    if fingerprint::canonical(expected) != fingerprint::canonical(retained) {
        return Err(Rejection::new("retained_evidence_mismatch", path));
    }
    Ok(())
}

fn baseline_binding(report: &Value, baseline: &Value, path: &str) -> Result<(), Rejection> {
    let initial = &baseline["initial"];
    for key in ["volume_m3", "temperature_k"] {
        equal(&initial[key], &report["initial"][key], path)?;
    }
    let authored = initial["components"].as_array().expect("typed components");
    let retained = report["initial"]["components"]
        .as_array()
        .expect("checked report components");
    if authored.len() != retained.len() {
        return Err(Rejection::new("report_request_mismatch", path));
    }
    for (input, part) in authored.iter().zip(retained) {
        for (input_key, report_key) in [
            ("id", "id"),
            ("mass_kg", "mass_kg"),
            (
                "specific_gas_constant_j_per_kg_k",
                "specific_gas_constant_j_per_kg_k",
            ),
            (
                "isochoric_heat_capacity_j_per_kg_k",
                "specific_isochoric_heat_capacity_j_per_kg_k",
            ),
        ] {
            equal(&input[input_key], &part[report_key], path)?;
        }
    }
    Ok(())
}

fn report_shape(value: &Value, path: &str) -> Result<(), Rejection> {
    let root = wire::object(value, path)?;
    let mut fields = vec![
        "schema",
        "status",
        "implementation_revision",
        "validation_environment",
    ];
    if root.get("status") == Some(&json!("invalid")) {
        fields.push("diagnostics");
    } else if root.get("status") == Some(&json!("predicted")) {
        fields.extend([
            "request_schema",
            "model",
            "quantity_system",
            "closure",
            "withdrawal_fraction",
            "initial",
            "final",
            "transfers",
            "balances",
            "assumptions",
            "nonclaims",
            "claims",
        ]);
    } else {
        return Err(Rejection::new("document_shape_invalid", path));
    }
    wire::closed(root, &fields, path)?;
    if root["schema"] != crate::REPORT_SCHEMA
        || root["implementation_revision"] != crate::IMPLEMENTATION_REVISION
        || root["validation_environment"]
            != json!({"ambient_inputs":[],"consulted_inputs":["authored_baseline","explicit_action"]})
    {
        return Err(Rejection::new("unsupported_retained_report", path));
    }
    if root["status"] == "invalid" {
        let diagnostics = array(&root["diagnostics"], path)?;
        if diagnostics.len() != 1 {
            return Err(Rejection::new("document_shape_invalid", path));
        }
        let diagnostic = wire::object(&diagnostics[0], path)?;
        wire::closed(diagnostic, &["code", "stage", "path", "reason_code"], path)?;
        for key in ["code", "stage", "path", "reason_code"] {
            wire::text(diagnostic, key)?;
        }
        if diagnostic["stage"] != "model"
            || !matches!(
                wire::text(diagnostic, "code")?,
                "FART-E-MODEL-0001" | "FART-E-MODEL-0002" | "FART-E-NUMERICAL-0001"
            )
            || !matches!(
                wire::text(diagnostic, "reason_code")?,
                "invalid_withdrawal"
                    | "reservoir_depletion"
                    | "unsupported_closure"
                    | "nonfinite_quantity"
                    | "nonpositive_quantity"
                    | "invalid_component_set"
                    | "no_representable_progress"
                    | "numerical_domain_error"
                    | "invariant_violation"
            )
        {
            return Err(Rejection::new("unsupported_retained_report", path));
        }
        return Ok(());
    }
    if root["request_schema"] != crate::REQUEST_SCHEMA
        || root["quantity_system"] != "si"
        || root["model"] != json!({"id":crate::MODEL_ID,"version":crate::MODEL_VERSION})
    {
        return Err(Rejection::new("unsupported_retained_report", path));
    }
    wire::text(root, "closure")?;
    numeric(&root["withdrawal_fraction"], path)?;
    for key in ["initial", "final"] {
        let state = wire::object(&root[key], path)?;
        let quantities = [
            "total_mass_kg",
            "volume_m3",
            "temperature_k",
            "mixture_gas_constant_j_per_kg_k",
            "mixture_specific_isochoric_heat_capacity_j_per_kg_k",
            "mixture_specific_isobaric_heat_capacity_j_per_kg_k",
            "heat_capacity_ratio",
            "pressure_pa",
            "internal_energy_j",
        ];
        let mut fields = quantities.to_vec();
        fields.push("components");
        wire::closed(state, &fields, path)?;
        for key in quantities {
            numeric(&state[key], path)?;
            if state[key].as_f64().expect("checked number") <= 0.0 {
                return Err(Rejection::new(
                    "retained_quantity_invalid",
                    &format!("{path}/{key}"),
                ));
            }
        }
        for part in array(&state["components"], path)? {
            record(
                part,
                &["id"],
                &[
                    "mass_kg",
                    "specific_gas_constant_j_per_kg_k",
                    "specific_isochoric_heat_capacity_j_per_kg_k",
                ],
                path,
            )?;
            for key in [
                "mass_kg",
                "specific_gas_constant_j_per_kg_k",
                "specific_isochoric_heat_capacity_j_per_kg_k",
            ] {
                if part[key].as_f64().expect("checked number") <= 0.0 {
                    return Err(Rejection::new(
                        "retained_quantity_invalid",
                        &format!("{path}/{key}"),
                    ));
                }
            }
        }
    }
    for (key, quantities, component_quantity) in [
        (
            "transfers",
            &[
                "total_mass_out_kg",
                "integrated_enthalpy_out_j",
                "heat_into_reservoir_j",
                "boundary_work_by_reservoir_j",
            ][..],
            "mass_out_kg",
        ),
        (
            "balances",
            &[
                "total_mass_residual_kg",
                "energy_residual_j",
                "initial_eos_residual_j",
                "final_eos_residual_j",
            ][..],
            "residual_kg",
        ),
    ] {
        let object = wire::object(&root[key], path)?;
        let mut fields = quantities.to_vec();
        fields.push("components");
        wire::closed(object, &fields, path)?;
        for key in quantities {
            numeric(&object[*key], path)?;
            if component_quantity == "mass_out_kg"
                && object[*key].as_f64().expect("checked number") < 0.0
            {
                return Err(Rejection::new(
                    "retained_quantity_invalid",
                    &format!("{path}/{key}"),
                ));
            }
        }
        for part in array(&object["components"], path)? {
            record(part, &["id"], &[component_quantity], path)?;
            if component_quantity == "mass_out_kg"
                && part[component_quantity].as_f64().expect("checked number") < 0.0
            {
                return Err(Rejection::new("retained_quantity_invalid", path));
            }
        }
    }
    let expected_ids: Vec<_> = root["initial"]["components"]
        .as_array()
        .expect("checked state array")
        .iter()
        .map(|part| &part["id"])
        .collect();
    for section in ["final", "transfers", "balances"] {
        let ids: Vec<_> = root[section]["components"]
            .as_array()
            .expect("checked component array")
            .iter()
            .map(|part| &part["id"])
            .collect();
        if ids != expected_ids {
            return Err(Rejection::new(
                "retained_component_mismatch",
                &format!("{path}/{section}/components"),
            ));
        }
    }
    let closure_assumption = match wire::text(root, "closure")? {
        "rigid-adiabatic" => "adiabatic-no-heat-transfer",
        "rigid-isothermal" => "prescribed-isothermal-ideal-thermostat",
        _ => return Err(Rejection::new("unsupported_retained_report", path)),
    };
    if root["assumptions"]
        != json!([
            "calorically-perfect-components",
            "homogeneous-equilibrium-state",
            "nonreacting-mixture",
            "single-gas-phase",
            "perfectly-mixed-nonselective-outflow",
            "rigid-volume",
            "sensible-energy-datum-cv-times-temperature",
            closure_assumption
        ])
        || root["nonclaims"]
            != json!({"model":["aperture-and-restriction-flow","elapsed-time-history","exterior-state",
            "momentum-and-recoil","phase-change-and-reaction","plume-and-acoustics"],
            "operation":["case-commitment","certificate-issuance"],"evidence":["empirical-validation"]})
    {
        return Err(Rejection::new("unsupported_retained_report", path));
    }
    let claims = array(&root["claims"], path)?;
    if claims.len() != 4 {
        return Err(Rejection::new(
            "retained_claim_mismatch",
            &format!("{path}/claims"),
        ));
    }
    for (index, (claim, (id, method, unit))) in claims
        .iter()
        .zip([
            ("reservoir.total-mass-balance", "double-entry-balance", "kg"),
            ("reservoir.energy-balance", "double-entry-balance", "J"),
            (
                "reservoir.initial-equation-of-state",
                "derived-state-consistency-residual",
                "J",
            ),
            (
                "reservoir.final-equation-of-state",
                "derived-state-consistency-residual",
                "J",
            ),
        ])
        .enumerate()
    {
        record(
            claim,
            &[
                "id",
                "status",
                "method",
                "equation_revision",
                "residual_unit",
            ],
            &["residual", "tolerance"],
            path,
        )?;
        if claim["id"] != id
            || claim["method"] != method
            || claim["residual_unit"] != unit
            || claim["status"] != "satisfied-within-roundoff"
            || claim["equation_revision"] != format!("{}@{}", crate::MODEL_ID, crate::MODEL_VERSION)
        {
            return Err(Rejection::new(
                "retained_claim_mismatch",
                &format!("{path}/claims/{index}"),
            ));
        }
        let residual = claim["residual"].as_f64().expect("checked residual");
        let tolerance = claim["tolerance"].as_f64().expect("checked tolerance");
        let balance_key = [
            "total_mass_residual_kg",
            "energy_residual_j",
            "initial_eos_residual_j",
            "final_eos_residual_j",
        ][index];
        if tolerance <= 0.0
            || residual.abs() > tolerance
            || root["balances"][balance_key].as_f64() != Some(residual)
        {
            return Err(Rejection::new(
                "retained_claim_inconsistent",
                &format!("{path}/claims/{index}"),
            ));
        }
    }
    Ok(())
}

fn record(value: &Value, texts: &[&str], numbers: &[&str], path: &str) -> Result<(), Rejection> {
    let root = wire::object(value, path)?;
    let mut fields = texts.to_vec();
    fields.extend(numbers);
    wire::closed(root, &fields, path)?;
    for key in texts {
        wire::text(root, key)?;
    }
    for key in numbers {
        numeric(&root[*key], path)?;
    }
    Ok(())
}
fn numeric(value: &Value, path: &str) -> Result<(), Rejection> {
    value
        .as_f64()
        .filter(|v| v.is_finite())
        .map(|_| ())
        .ok_or_else(|| Rejection::new("document_shape_invalid", path))
}
fn array<'a>(value: &'a Value, path: &str) -> Result<&'a Vec<Value>, Rejection> {
    value
        .as_array()
        .filter(|v| !v.is_empty() && v.len() <= 64)
        .ok_or_else(|| Rejection::new("document_shape_invalid", path))
}
