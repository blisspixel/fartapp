//! Native restriction wire, report, refusal, and independent analytical checks.

use fart_services::{
    InputFailure,
    restriction::{self, Kind, Report},
};
use serde_json::{Value, json};

const CHOKED: &[u8] = include_bytes!("../../../testdata/restriction/gamma15-choked.json");
const SUBSONIC: &[u8] = include_bytes!("../../../testdata/restriction/gamma15-subsonic.json");
const COMPLIANT: &[u8] =
    include_bytes!("../../../testdata/restriction/linear-compliance-choked.json");
const ORDINARY: &[u8] =
    include_bytes!("../../../testdata/restriction/ordinary-pressure-subsonic.json");
const HISTORY: &[u8] = include_bytes!("../../../testdata/restriction/gamma15-choked-history.json");

fn document(kind: Kind) -> Value {
    serde_json::from_slice(if kind == Kind::Prediction {
        CHOKED
    } else {
        HISTORY
    })
    .unwrap()
}

fn run(kind: Kind, bytes: &[u8]) -> Report {
    match kind {
        Kind::Prediction => restriction::predict(bytes),
        Kind::History => restriction::history(bytes),
    }
}

fn evaluate(kind: Kind, value: &Value) -> Report {
    run(kind, &serde_json::to_vec(value).unwrap())
}

fn json_report(report: &Report) -> Value {
    serde_json::from_str(&report.to_json()).unwrap()
}

fn refused(report: &Report, code: &str, stage: &str, path: &str, reason: &str) {
    assert!(!report.is_predicted(), "{}", report.to_json());
    assert!(report.prediction_result().is_none());
    assert!(report.history_result().is_none());
    let diagnostic = report.diagnostic().unwrap();
    assert_eq!(
        (
            diagnostic.code,
            diagnostic.stage,
            diagnostic.path.as_str(),
            diagnostic.reason_code
        ),
        (code, stage, path, reason),
        "{}",
        report.to_json()
    );
    let value = json_report(report);
    assert_eq!(value.as_object().unwrap().len(), 5);
    assert_eq!(value["diagnostics"].as_array().unwrap().len(), 1);
    assert!(report.to_text().contains("Recovery:"));
}

fn schema(kind: Kind, value: &Value, path: &str, reason: &str) {
    refused(
        &evaluate(kind, value),
        if kind == Kind::Prediction {
            "FART-E-SCHEMA-0003"
        } else {
            "FART-E-SCHEMA-0004"
        },
        "schema",
        path,
        reason,
    );
}

fn model(kind: Kind, value: &Value, path: &str, reason: &str) {
    refused(
        &evaluate(kind, value),
        if kind == Kind::Prediction {
            "FART-E-MODEL-0003"
        } else {
            "FART-E-MODEL-0005"
        },
        "model",
        path,
        reason,
    );
}

fn close(actual: f64, expected: f64) {
    assert!(
        (actual - expected).abs() <= 64.0 * f64::EPSILON * expected.abs() + f64::from_bits(1),
        "{actual:e} != {expected:e}"
    );
}

#[test]
fn retained_fixtures_expose_complete_current_reports_and_independent_anchors() {
    for bytes in [CHOKED, SUBSONIC, COMPLIANT, ORDINARY] {
        let report = restriction::predict(bytes);
        assert!(report.is_predicted(), "{}", report.to_json());
        assert!(report.diagnostic().is_none());
        assert!(report.history_result().is_none());
        let value = json_report(&report);
        assert_eq!(value["schema"], restriction::REPORT_SCHEMA);
        assert_eq!(value["request_schema"], restriction::REQUEST_SCHEMA);
        assert_eq!(
            value["implementation_revision"],
            restriction::IMPLEMENTATION_REVISION
        );
        assert_eq!(
            value["model"],
            json!({"id":restriction::MODEL_ID,"version":restriction::MODEL_VERSION})
        );
        assert_eq!(
            value["assumptions"],
            json!([
                "calorically-perfect-gas",
                "quasi-steady-flow",
                "isentropic-control-section",
                "converging-restriction-only",
                "discharge-coefficient-scales-mass-flow-only",
                "no-reverse-flow",
                "no-shock-inside-restriction",
                "prescribed-or-linear-quasi-static-area"
            ])
        );
        assert_eq!(
            value["nonclaims"],
            json!({"model":["elapsed-time-history","reservoir-mass-energy-coupling",
            "shock-containing-or-underexpanded-plume","viscous-resolved-vena-contracta","phase-change-and-reaction","acoustics"],
            "operation":["case-commitment","certificate-issuance"],"evidence":["empirical-validation"]})
        );
        check_claims(
            &value,
            &[
                (
                    "restriction.mass-flow-definition",
                    "cd-scaled-exit-mass-flux",
                    "kg/s",
                ),
                (
                    "restriction.thrust-control-surface",
                    "momentum-and-pressure-thrust",
                    "N",
                ),
                (
                    "restriction.recoil-action-reaction",
                    "equal-and-opposite-force",
                    "N",
                ),
            ],
        );
        assert_eq!(
            value["validation_environment"],
            json!({"consulted_inputs":["document_bytes"],"ambient_inputs":[]})
        );
        let text = report.to_text();
        assert!(text.starts_with("RESTRICTION PREDICTED\n"));
        assert!(text.contains("Mass flow:"));
        assert!(text.contains("Human values: six significant digits; full precision in JSON."));
        assert_eq!(report.to_json(), report.clone().to_json());
    }
    let report = restriction::predict(CHOKED);
    let result = report.prediction_result().unwrap();
    close(result.exit_pressure_pa, 64000.0);
    close(result.exit_temperature_k, 320.0);
    close(result.mass_flow_kg_per_s, 0.01 * 96000.0_f64.sqrt());
    close(result.thrust_n, 1100.0);
    close(result.recoil_n, -1100.0);
    close(
        restriction::predict(SUBSONIC)
            .prediction_result()
            .unwrap()
            .throat_mach,
        2.0 / 3.0,
    );
    close(
        restriction::predict(COMPLIANT)
            .prediction_result()
            .unwrap()
            .effective_area_m2,
        0.0085,
    );
}

fn check_claims(report: &Value, expected: &[(&str, &str, &str)]) {
    let claims = report["claims"].as_array().unwrap();
    assert_eq!(claims.len(), expected.len());
    for (claim, (id, method, unit)) in claims.iter().zip(expected) {
        assert_eq!(claim["id"], *id);
        assert_eq!(claim["method"], *method);
        assert_eq!(claim["residual_unit"], *unit);
        assert_eq!(claim["status"], "satisfied-within-roundoff");
        assert_eq!(
            claim["equation_revision"],
            format!("{}@{}", restriction::MODEL_ID, restriction::MODEL_VERSION)
        );
        let residual = claim["residual"].as_f64().unwrap();
        let tolerance = claim["tolerance"].as_f64().unwrap();
        assert!(
            residual.is_finite()
                && tolerance.is_finite()
                && tolerance >= 0.0
                && residual.abs() <= tolerance
        );
    }
}

#[test]
fn area_normalization_and_closed_adverse_boundary_remain_explicit() {
    let mut value: Value = serde_json::from_slice(COMPLIANT).unwrap();
    value["area"]["compliance_m2_per_pa"] = json!(0.0);
    let report = evaluate(Kind::Prediction, &value);
    assert!(report.is_predicted());
    assert_eq!(
        json_report(&report)["area"],
        json!({"law":"prescribed","prescribed_m2":0.001,"effective_m2":0.001})
    );
    value = document(Kind::Prediction);
    value["back_pressure_pa"] = json!(200000.0);
    refused(
        &evaluate(Kind::Prediction, &value),
        "FART-E-MODEL-0004",
        "model",
        "/",
        "adverse_pressure",
    );
    value["area"]["prescribed_m2"] = json!(-0.0);
    let report = evaluate(Kind::Prediction, &value);
    assert!(report.is_predicted());
    let flow = report.prediction_result().unwrap();
    assert_eq!(flow.regime.name(), "no-flow");
    assert_eq!(flow.mass_flow_kg_per_s, 0.0);
    assert_eq!(flow.exit_pressure_pa, 200000.0);
    value["area"]["prescribed_m2"] = json!(0.01);
    value["back_pressure_pa"] = value["stagnation"]["pressure_pa"].clone();
    assert_eq!(
        evaluate(Kind::Prediction, &value)
            .prediction_result()
            .unwrap()
            .regime
            .name(),
        "no-flow"
    );
}

#[test]
fn history_retains_samples_totals_and_frozen_source_nonclaims() {
    let report = restriction::history(HISTORY);
    assert!(report.is_predicted());
    assert!(report.prediction_result().is_none());
    let result = report.history_result().unwrap();
    assert_eq!(result.samples.len(), 2);
    close(result.mass_out_kg, 0.0001 * 96000.0_f64.sqrt());
    close(result.impulse_n_s, 11.0);
    close(
        result.total_enthalpy_out_j,
        result.enthalpy_out_j + result.kinetic_energy_out_j,
    );
    let value = json_report(&report);
    assert_eq!(value["schema"], restriction::HISTORY_REPORT_SCHEMA);
    assert_eq!(value["request_schema"], restriction::HISTORY_REQUEST_SCHEMA);
    assert_eq!(
        value["implementation_revision"],
        restriction::HISTORY_IMPLEMENTATION_REVISION
    );
    assert_eq!(
        value["assumptions"],
        json!([
            "frozen-stagnation-state",
            "prescribed-area-history",
            "trapezoidal-rate-integration",
            "quasi-steady-samples",
            "single-calorically-perfect-gas",
            "enthalpy-out-is-static-exit-enthalpy",
            "total-enthalpy-includes-exit-kinetic-energy"
        ])
    );
    assert_eq!(
        value["nonclaims"]["model"],
        json!([
            "reservoir-coupling-and-blowdown",
            "species-resolved-composition-history",
            "plume-and-acoustics",
            "elapsed-source-depletion"
        ])
    );
    check_claims(
        &value,
        &[(
            "restriction-history.recoil-action-reaction",
            "equal-and-opposite-impulse",
            "N s",
        )],
    );
    assert!(report.to_text().contains("Static exit enthalpy out:"));
    assert!(
        report
            .to_text()
            .contains("Experimental frozen-source history.")
    );
    let mut value = document(Kind::History);
    value["samples"] = json!([{"time_s":7.0,"prescribed_m2":0.01}]);
    let singleton = evaluate(Kind::History, &value);
    assert!(singleton.is_predicted());
    for number in json_report(&singleton)["totals"]
        .as_object()
        .unwrap()
        .values()
    {
        assert_eq!(number.as_f64().unwrap(), 0.0);
    }
    value["samples"] = json!(
        (0..256)
            .map(|i| json!({"time_s":i,"prescribed_m2":0.0}))
            .collect::<Vec<_>>()
    );
    let maximum = evaluate(Kind::History, &value);
    assert_eq!(maximum.history_result().unwrap().samples.len(), 256);
    assert_eq!(maximum.history_result().unwrap().mass_out_kg, 0.0);
    value["samples"]
        .as_array_mut()
        .unwrap()
        .push(json!({"time_s":256,"prescribed_m2":0.0}));
    schema(Kind::History, &value, "/samples", "invalid_sample_count");
}

#[test]
fn all_shape_checks_precede_semantics_with_exact_sorted_pointers() {
    for kind in [Kind::Prediction, Kind::History] {
        for path in [
            "/schema",
            "/model",
            "/model/id",
            "/model/version",
            "/quantity_system",
            "/stagnation",
            "/stagnation/pressure_pa",
            "/stagnation/temperature_k",
            "/stagnation/specific_gas_constant_j_per_kg_k",
            "/stagnation/heat_capacity_ratio",
            "/back_pressure_pa",
            "/discharge_coefficient",
        ] {
            let mut value = document(kind);
            *value.pointer_mut(path).unwrap() = Value::Null;
            schema(kind, &value, path, "document_shape_invalid");
        }
        for invalid_root in [
            Value::Null,
            json!([]),
            json!(true),
            json!("object"),
            json!(1),
        ] {
            schema(kind, &invalid_root, "/", "document_shape_invalid");
        }
        let mut value = document(kind);
        value["schema"] = json!("unsupported");
        value["stagnation"]["pressure_pa"] = json!(-1.0);
        value["stagnation"]["Pressure_pa"] = json!(1.0);
        schema(
            kind,
            &value,
            "/stagnation/Pressure_pa",
            "document_shape_invalid",
        );
        value["A/~"] = Value::Null;
        schema(kind, &value, "/A~1~0", "document_shape_invalid");
        value = document(kind);
        value["model"]["id"] = json!(false);
        schema(kind, &value, "/model/id", "document_shape_invalid");
    }
    for (kind, paths) in [
        (
            Kind::Prediction,
            vec!["/area", "/area/law", "/area/prescribed_m2"],
        ),
        (
            Kind::History,
            vec![
                "/samples",
                "/samples/0",
                "/samples/0/time_s",
                "/samples/0/prescribed_m2",
            ],
        ),
    ] {
        for path in paths {
            let mut value = document(kind);
            *value.pointer_mut(path).unwrap() = json!(true);
            schema(kind, &value, path, "document_shape_invalid");
        }
    }
    let mut value = document(Kind::History);
    value["samples"][0]["time_s"] = json!(-1.0);
    value["samples"][1]["Time_s"] = json!(0);
    schema(
        Kind::History,
        &value,
        "/samples/1/Time_s",
        "document_shape_invalid",
    );
}

#[test]
fn missing_fields_do_not_create_scientific_defaults() {
    for kind in [Kind::Prediction, Kind::History] {
        for (key, path, reason) in [
            (
                "schema",
                if kind == Kind::Prediction {
                    "/schema"
                } else {
                    "/"
                },
                "unsupported_schema",
            ),
            ("model", "/model", "unsupported_model_revision"),
            (
                "quantity_system",
                "/quantity_system",
                "unsupported_quantity_system",
            ),
            ("back_pressure_pa", "/back_pressure_pa", "missing_member"),
            (
                "discharge_coefficient",
                "/discharge_coefficient",
                "missing_member",
            ),
        ] {
            let mut value = document(kind);
            value.as_object_mut().unwrap().remove(key);
            schema(kind, &value, path, reason);
        }
        for field in [
            "pressure_pa",
            "temperature_k",
            "specific_gas_constant_j_per_kg_k",
            "heat_capacity_ratio",
        ] {
            let mut value = document(kind);
            value["stagnation"]["pressure_pa"] = json!(-1.0);
            value["stagnation"].as_object_mut().unwrap().remove(field);
            let field_path = format!("/stagnation/{field}");
            schema(
                kind,
                &value,
                if kind == Kind::History {
                    "/stagnation"
                } else {
                    &field_path
                },
                "missing_member",
            );
        }
    }
    let mut value = document(Kind::Prediction);
    value.as_object_mut().unwrap().remove("area");
    schema(
        Kind::Prediction,
        &value,
        "/area/prescribed_m2",
        "missing_member",
    );
    for field in ["time_s", "prescribed_m2"] {
        let mut value = document(Kind::History);
        value["samples"][1].as_object_mut().unwrap().remove(field);
        schema(Kind::History, &value, "/samples/1", "missing_member");
    }
    for samples in [None, Some(json!([]))] {
        let mut value = document(Kind::History);
        if let Some(samples) = samples {
            value["samples"] = samples;
        } else {
            value.as_object_mut().unwrap().remove("samples");
        }
        schema(Kind::History, &value, "/samples", "missing_member");
    }
}

#[test]
fn semantic_domain_refusals_keep_operation_specific_families() {
    for kind in [Kind::Prediction, Kind::History] {
        for (path, number, reason) in [
            ("/stagnation/pressure_pa", 0.0, "nonpositive_quantity"),
            ("/stagnation/temperature_k", -1.0, "nonpositive_quantity"),
            (
                "/stagnation/specific_gas_constant_j_per_kg_k",
                0.0,
                "nonpositive_quantity",
            ),
            (
                "/stagnation/heat_capacity_ratio",
                1.0,
                "invalid_heat_capacity_ratio",
            ),
            ("/back_pressure_pa", 0.0, "nonpositive_quantity"),
            (
                "/discharge_coefficient",
                1.01,
                "invalid_discharge_coefficient",
            ),
        ] {
            let mut value = document(kind);
            *value.pointer_mut(path).unwrap() = json!(number);
            model(kind, &value, path, reason);
        }
        let mut value = document(kind);
        value["model"]["version"] = json!("next");
        schema(kind, &value, "/model", "unsupported_model_revision");
        value = document(kind);
        value["quantity_system"] = json!("SI");
        schema(
            kind,
            &value,
            "/quantity_system",
            "unsupported_quantity_system",
        );
    }
    let mut value = document(Kind::Prediction);
    value["area"]["law"] = json!("inferred");
    schema(
        Kind::Prediction,
        &value,
        "/area/law",
        "unsupported_area_law",
    );
    value["area"]["prescribed_m2"] = json!(-1.0);
    model(
        Kind::Prediction,
        &value,
        "/area/prescribed_m2",
        "negative_area",
    );
    for key in ["compliance_m2_per_pa", "maximum_m2"] {
        value = document(Kind::Prediction);
        value["area"][key] = json!(0.0);
        schema(
            Kind::Prediction,
            &value,
            &format!("/area/{key}"),
            "unexpected_member",
        );
        value = serde_json::from_slice(COMPLIANT).unwrap();
        value["area"].as_object_mut().unwrap().remove(key);
        schema(
            Kind::Prediction,
            &value,
            &format!("/area/{key}"),
            "missing_member",
        );
    }
    for (key, number, reason) in [
        ("compliance_m2_per_pa", -1.0, "negative_compliance"),
        ("maximum_m2", -1.0, "negative_area"),
    ] {
        value = serde_json::from_slice(COMPLIANT).unwrap();
        value["area"][key] = json!(number);
        model(Kind::Prediction, &value, &format!("/area/{key}"), reason);
    }
    value = serde_json::from_slice(COMPLIANT).unwrap();
    value["area"]["maximum_m2"] = json!(0.0);
    model(Kind::Prediction, &value, "/area", "invalid_area_law");
    for (key, reason) in [
        ("time_s", "invalid_time"),
        ("prescribed_m2", "negative_area"),
    ] {
        value = document(Kind::History);
        value["samples"][1][key] = json!(-1.0);
        model(Kind::History, &value, &format!("/samples/1/{key}"), reason);
    }
    value = document(Kind::History);
    value["samples"][1]["time_s"] = json!(0.0);
    model(Kind::History, &value, "/samples", "invalid_time");
    value = document(Kind::History);
    value["back_pressure_pa"] = json!(200000.0);
    model(Kind::History, &value, "/samples", "adverse_pressure");
}

#[test]
fn syntax_limits_and_unicode_are_independent_of_document_semantics() {
    for kind in [Kind::Prediction, Kind::History] {
        let code = if kind == Kind::Prediction {
            "FART-E-JSON-0003"
        } else {
            "FART-E-JSON-0004"
        };
        for (bytes, path, reason) in [
            (b" \t\n\r".as_slice(), "/", "empty_input"),
            (br#"{"a":1,"\u0061":2}"#, "/a", "duplicate_member"),
            (br#"{"a/~":1,"a/~":2}"#, "/a~1~0", "duplicate_member"),
            (br#"{"a":1,"a":2,"bad":"\ud800"}"#, "/", "malformed_json"),
            (br#"{"a":1,"a":2,"bad":"\udc00"}"#, "/", "malformed_json"),
            (br#"{"bad":"\ud800\u0041"}"#, "/", "malformed_json"),
            (br#"{"bad":"\uQQQQ"}"#, "/", "malformed_json"),
            (br#"{"bad":"\u1"}"#, "/", "malformed_json"),
            (br#"{"bad":"\q"}"#, "/", "malformed_json"),
            (br#"{"a":1,"a":2,"bad":""#, "/a", "duplicate_member"),
            (b"{\"a\":1,\"a\":2,\"bad\":\"\xff\"}", "/", "malformed_json"),
            (b"{}[]", "/", "trailing_json_value"),
            (b"{]", "/", "malformed_json"),
        ] {
            refused(&run(kind, bytes), code, "syntax", path, reason);
        }
        let name = format!("{{\"{}\":0}}", "x".repeat(129));
        refused(
            &run(kind, name.as_bytes()),
            code,
            "syntax",
            "/",
            "member_name_too_long",
        );
        let depth = format!("{}0{}", "[".repeat(33), "]".repeat(33));
        refused(
            &run(kind, depth.as_bytes()),
            code,
            "syntax",
            &"/0".repeat(33),
            "maximum_depth_exceeded",
        );
        let mut boundary = if kind == Kind::Prediction {
            CHOKED.to_vec()
        } else {
            HISTORY.to_vec()
        };
        boundary.resize(restriction::MAX_INPUT_BYTES, b' ');
        assert!(run(kind, &boundary).is_predicted());
        boundary.push(b' ');
        refused(
            &run(kind, &boundary),
            if kind == Kind::Prediction {
                "FART-E-INPUT-0003"
            } else {
                "FART-E-INPUT-0004"
            },
            "input",
            "/",
            "input_too_large",
        );
        for text in [
            r#"{"x":"\uD83D\uDE00"}"#,
            r#"{"x":"\u0061\u00AF"}"#,
            r#"{"x":"\\ud800 \" 1e999"}"#,
        ] {
            refused(
                &run(kind, text.as_bytes()),
                if kind == Kind::Prediction {
                    "FART-E-SCHEMA-0003"
                } else {
                    "FART-E-SCHEMA-0004"
                },
                "schema",
                "/x",
                "document_shape_invalid",
            );
        }
    }
}

#[test]
fn overflow_numbers_are_schema_refusals_and_never_model_substitutes() {
    for kind in [Kind::Prediction, Kind::History] {
        let schema_code = if kind == Kind::Prediction {
            "FART-E-SCHEMA-0003"
        } else {
            "FART-E-SCHEMA-0004"
        };
        let syntax_code = if kind == Kind::Prediction {
            "FART-E-JSON-0003"
        } else {
            "FART-E-JSON-0004"
        };
        for token in [
            "1e999",
            "-1e999",
            "1E+999",
            "1797693134862315900000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
        ] {
            let mut value = document(kind);
            if kind == Kind::Prediction {
                value["area"]["prescribed_m2"] = json!("NUMBER");
            } else {
                value["samples"][0]["time_s"] = json!("NUMBER");
            }
            let bytes = value.to_string().replace("\"NUMBER\"", token);
            refused(
                &run(kind, bytes.as_bytes()),
                schema_code,
                "schema",
                "/",
                "document_shape_invalid",
            );
        }
        for (bytes, code, stage, path, reason) in [
            (
                r#"{"back_pressure_pa":1e999,"back_pressure_pa":1}"#,
                syntax_code,
                "syntax",
                "/back_pressure_pa",
                "duplicate_member",
            ),
            (
                r#"{"back_pressure_pa":1e999,"stagnation":null}"#,
                schema_code,
                "schema",
                "/stagnation",
                "document_shape_invalid",
            ),
            (
                r#"{"back_pressure_pa":1e999,"unknown":0}"#,
                schema_code,
                "schema",
                "/unknown",
                "document_shape_invalid",
            ),
            (
                r#"{"unknown":1e999}"#,
                schema_code,
                "schema",
                "/unknown",
                "document_shape_invalid",
            ),
            (
                r#"{"back_pressure_pa":01e999}"#,
                syntax_code,
                "syntax",
                "/",
                "malformed_json",
            ),
            (
                r#"{"back_pressure_pa":1e999foo}"#,
                syntax_code,
                "syntax",
                "/",
                "malformed_json",
            ),
            (
                r#"{"back_pressure_pa":1e999} []"#,
                syntax_code,
                "syntax",
                "/",
                "trailing_json_value",
            ),
            (
                r#"{"back_pressure_pa":1e999,"discharge_coefficient":-1e999}"#,
                schema_code,
                "schema",
                "/",
                "document_shape_invalid",
            ),
        ] {
            refused(&run(kind, bytes.as_bytes()), code, stage, path, reason);
        }
        let mut value = document(kind);
        value["model"]["id"] = json!("1e999 and -1e999");
        schema(kind, &value, "/model", "unsupported_model_revision");
    }
}

#[test]
fn io_failures_are_operation_specific_and_human_paths_are_escaped() {
    for kind in [Kind::Prediction, Kind::History] {
        for (failure, reason) in [
            (InputFailure::NotFound, "input_not_found"),
            (InputFailure::PermissionDenied, "input_permission_denied"),
            (InputFailure::Unavailable, "input_unavailable"),
            (InputFailure::TooLarge, "input_too_large"),
        ] {
            for reading_stream in [false, true] {
                let report = restriction::input_failure(kind, failure, reading_stream);
                refused(
                    &report,
                    if kind == Kind::Prediction {
                        "FART-E-IO-0003"
                    } else {
                        "FART-E-IO-0004"
                    },
                    "input",
                    "/",
                    reason,
                );
                assert_eq!(
                    json_report(&report)["validation_environment"]["consulted_inputs"],
                    json!([if reading_stream {
                        "input_stream"
                    } else {
                        "input_source_reference"
                    }])
                );
            }
        }
        let report = run(kind, br#"{"x\u001b[2J\n":1,"x\u001b[2J\n":2}"#);
        assert_eq!(report.diagnostic().unwrap().path, "/x\u{1b}[2J\n");
        let text = report.to_text();
        assert!(!text.contains('\u{1b}'));
        assert_eq!(text.lines().count(), 2);
    }
}

#[test]
fn unrepresentable_predictions_refuse_without_partial_accounts() {
    for kind in [Kind::Prediction, Kind::History] {
        let mut value = document(kind);
        value["stagnation"]["temperature_k"] = json!(f64::MAX);
        value["stagnation"]["specific_gas_constant_j_per_kg_k"] = json!(f64::MAX);
        refused(
            &evaluate(kind, &value),
            if kind == Kind::Prediction {
                "FART-E-MODEL-0004"
            } else {
                "FART-E-MODEL-0005"
            },
            "model",
            if kind == Kind::Prediction {
                "/"
            } else {
                "/samples"
            },
            if kind == Kind::Prediction {
                "no_representable_flow"
            } else {
                "numerical_domain_error"
            },
        );
    }
    let mut value = document(Kind::History);
    value["samples"][1]["time_s"] = json!(f64::MAX);
    model(Kind::History, &value, "/samples", "numerical_domain_error");
}
