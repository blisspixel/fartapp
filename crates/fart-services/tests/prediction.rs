//! Current Go input compatibility and strict native boundary tests.

use fart_services::*;
use serde_json::{Value, json};

const ADIABATIC: &[u8] =
    include_bytes!("../../../testdata/reservoir/synthetic-mixture-adiabatic.json");
const ISOTHERMAL: &[u8] =
    include_bytes!("../../../testdata/reservoir/synthetic-mixture-isothermal.json");

fn document() -> Value {
    serde_json::from_slice(ADIABATIC).unwrap()
}
fn predict(value: Value) -> PredictionReport {
    predict_reservoir(&serde_json::to_vec(&value).unwrap())
}

#[test]
fn both_existing_fixture_inputs_produce_complete_typed_reports() {
    for (bytes, closure, temperature) in [
        (ADIABATIC, "rigid-adiabatic", 200.0),
        (ISOTHERMAL, "rigid-isothermal", 400.0),
    ] {
        let report = predict_reservoir(bytes);
        assert!(report.is_predicted(), "{}", report.to_json());
        let transition = report.outcome().unwrap();
        assert!((transition.final_state.temperature_k - temperature).abs() < 1e-10);
        let value: Value = serde_json::from_str(&report.to_json()).unwrap();
        assert_eq!(value["schema"], REPORT_SCHEMA);
        assert_eq!(value["implementation_revision"], IMPLEMENTATION_REVISION);
        assert_eq!(value["closure"], closure);
        assert_eq!(value["final"]["total_mass_kg"], 1.0);
        assert_eq!(value["claims"].as_array().unwrap().len(), 4);
        assert_eq!(
            value["nonclaims"]["evidence"],
            json!(["empirical-validation"])
        );
        assert_eq!(value["validation_environment"]["ambient_inputs"], json!([]));
        let text = report.to_text();
        for expected in [
            "RESERVOIR ENDPOINT PREDICTED",
            "INITIAL",
            "FINAL",
            "BALANCE CLAIMS",
            "Experimental analytical endpoint only.",
        ] {
            assert!(text.contains(expected));
        }
    }
}

#[test]
fn zero_and_component_reordering_do_not_invent_changes() {
    let mut input = document();
    input["withdrawal_fraction"] = json!(0.0);
    let report = predict(input.clone());
    let result = report.outcome().unwrap();
    assert_eq!(result.before, result.after);
    let original = report.to_json();
    input["initial"]["components"]
        .as_array_mut()
        .unwrap()
        .reverse();
    assert_eq!(predict(input).to_json(), original);
}

#[test]
fn syntax_limits_nulls_and_duplicates_are_refused_before_map_conversion() {
    let duplicate = String::from_utf8(ADIABATIC.to_vec()).unwrap().replace(
        "\"withdrawal_fraction\": 0.75",
        "\"withdrawal_fraction\": 0.75, \"withdrawal_fraction\": 0.5",
    );
    let escaped_duplicate = duplicate.replace(
        "\"withdrawal_fraction\": 0.5",
        "\"withdrawal_\\u0066raction\": 0.5",
    );
    for bytes in [
        b"{".to_vec(),
        b"{} {}".to_vec(),
        b"null".to_vec(),
        b"{\"x\": null}".to_vec(),
        b"{\"x\": 1e999}".to_vec(),
        b"{\"x\":\"\\ud800\"}".to_vec(),
        vec![0xff],
        duplicate.into_bytes(),
        escaped_duplicate.into_bytes(),
        format!("{}0{}", "[".repeat(34), "]".repeat(34)).into_bytes(),
        format!("{{\"{}\":0}}", "x".repeat(129)).into_bytes(),
        vec![b' '; MAX_INPUT_BYTES + 1],
    ] {
        let result = predict_reservoir(&bytes);
        assert!(!result.is_predicted(), "accepted {:?}", bytes);
        assert!(!result.outcome().unwrap_err().reason_code.is_empty());
        assert!(result.to_text().contains("reservoir prediction failed:"));
        let value: Value = serde_json::from_str(&result.to_json()).unwrap();
        assert_eq!(value["status"], "invalid");
        assert!(value.get("final").is_none());
    }
    for scalar in ["true", "false", "0", "-1", "1.25", "\"text\"", "[]"] {
        assert!(!predict_reservoir(scalar.as_bytes()).is_predicted());
    }
    let mut large = document();
    large["initial"]["components"] = json!(vec![large["initial"]["components"][0].clone(); 65]);
    assert_eq!(
        predict(large).outcome().unwrap_err().reason_code,
        "collection_limit_exceeded"
    );
}

#[test]
fn schema_and_model_refusals_are_explicit_without_defaults() {
    let cases = [
        ("/schema", json!("unknown")),
        ("/model/id", json!("unknown")),
        ("/model/version", json!("v999")),
        ("/quantity_system", json!("cgs")),
        ("/closure", json!("default")),
        ("/withdrawal_fraction", json!(-0.1)),
        ("/withdrawal_fraction", json!(1.0)),
        ("/withdrawal_fraction", json!(1e-20)),
        ("/initial/volume_m3", json!(0.0)),
        ("/initial/temperature_k", json!(-1.0)),
        ("/initial/components", json!([])),
        ("/initial/components/0/id", json!("UPPER")),
        ("/initial/components/0/mass_kg", json!(-1.0)),
        (
            "/initial/components/0/specific_gas_constant_j_per_kg_k",
            json!(0.0),
        ),
        (
            "/initial/components/0/isochoric_heat_capacity_j_per_kg_k",
            json!(0.0),
        ),
        ("/initial/components/0/mass_kg", json!("1")),
        ("/initial/components/0", json!(true)),
        ("/initial", json!([])),
        ("/model", json!(1)),
        ("/closure", json!(false)),
        ("/initial/temperature_k", json!(1e308)),
    ];
    for (pointer, replacement) in cases {
        let mut input = document();
        *input.pointer_mut(pointer).unwrap() = replacement;
        assert!(!predict(input).is_predicted(), "accepted {pointer}");
    }
    for parent in ["", "/model", "/initial", "/initial/components/0"] {
        let mut input = document();
        input
            .pointer_mut(parent)
            .unwrap()
            .as_object_mut()
            .unwrap()
            .insert("Unrecognized".into(), json!(1));
        assert_eq!(
            predict(input).outcome().unwrap_err().reason_code,
            "document_shape_invalid"
        );
    }
    for (parent, key, reason) in [
        ("", "schema", "unsupported_schema"),
        ("", "model", "unsupported_model_revision"),
        ("", "initial", "missing_member"),
        ("/initial", "temperature_k", "missing_member"),
        ("/initial", "components", "missing_component"),
        ("/initial/components/0", "id", "invalid_token"),
        ("/initial/components/0", "mass_kg", "missing_member"),
    ] {
        let mut input = document();
        input
            .pointer_mut(parent)
            .unwrap()
            .as_object_mut()
            .unwrap()
            .remove(key);
        assert_eq!(predict(input).outcome().unwrap_err().reason_code, reason);
    }
    let mut duplicate = document();
    duplicate["initial"]["components"][1]["id"] = json!("component.a");
    assert_eq!(
        predict(duplicate).outcome().unwrap_err().reason_code,
        "duplicate_component_id"
    );
    let mut alias = document();
    alias["Schema"] = alias["schema"].clone();
    assert_eq!(
        predict(alias).outcome().unwrap_err().reason_code,
        "document_shape_invalid"
    );
}

#[test]
fn complete_shape_is_checked_before_semantics_and_null_is_never_omission() {
    for pointer in [
        "",
        "/schema",
        "/model",
        "/model/id",
        "/initial",
        "/initial/components",
        "/initial/components/0",
        "/initial/components/0/id",
        "/initial/components/0/mass_kg",
        "/withdrawal_fraction",
        "/closure",
    ] {
        let mut input = document();
        input["schema"] = json!("unsupported");
        *input.pointer_mut(pointer).unwrap() = Value::Null;
        let report = predict(input);
        let diagnostic = report.outcome().unwrap_err();
        assert_eq!(
            (
                diagnostic.code,
                diagnostic.path.as_str(),
                diagnostic.reason_code
            ),
            ("FART-E-SCHEMA-0002", "/", "document_shape_invalid")
        );
    }
    assert_eq!(
        predict_reservoir(b"null trailing")
            .outcome()
            .unwrap_err()
            .reason_code,
        "trailing_json_value"
    );
    assert_eq!(
        predict_reservoir(b" \r\n\t")
            .outcome()
            .unwrap_err()
            .reason_code,
        "empty_input"
    );
    let mut missing = document();
    missing["initial"]["components"][1]
        .as_object_mut()
        .unwrap()
        .remove("mass_kg");
    assert_eq!(
        predict(missing).outcome().unwrap_err().path,
        "/initial/components/1"
    );
    let mut progress = document();
    progress["withdrawal_fraction"] = json!(1e-20);
    assert_eq!(
        predict(progress).outcome().unwrap_err().code,
        "FART-E-MODEL-0002"
    );
}

#[test]
fn input_failures_and_intensity_facade_preserve_their_boundaries() {
    for (failure, reason) in [
        (InputFailure::NotFound, "input_not_found"),
        (InputFailure::PermissionDenied, "input_permission_denied"),
        (InputFailure::Unavailable, "input_unavailable"),
        (InputFailure::TooLarge, "input_too_large"),
    ] {
        for stream in [false, true] {
            let report = reservoir_input_failure(failure, stream);
            assert_eq!(report.outcome().unwrap_err().reason_code, reason);
            let value: Value = serde_json::from_str(&report.to_json()).unwrap();
            assert_eq!(value["diagnostics"][0]["code"], "FART-E-IO-0002");
            assert_eq!(
                value["validation_environment"]["consulted_inputs"][0],
                if stream {
                    "input_stream"
                } else {
                    "input_source_reference"
                }
            );
        }
    }
    assert_eq!(intensity_reply("3"), Some("braaap (respectable)\n"));
    for invalid in ["", "0", "6", "01", "+1", "1.0", " 1", "\u{ff11}"] {
        assert_eq!(intensity_reply(invalid), None);
    }
}
