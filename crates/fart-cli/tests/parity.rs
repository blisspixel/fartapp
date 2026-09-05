//! Opt-in native executable comparison against the independently built Go oracle.
//! Set FARTAPP_GO_ORACLE to an existing executable to make this a required gate.

use std::{
    env, fs,
    io::Write,
    path::Path,
    process::{Command, Output, Stdio},
};

use serde_json::{Value, json};

const RUST_REVISION: &str = "rust-reservoir/v0alpha1";
const GO_REVISION: &str = "go-oracle.reservoir/v0alpha2";
const RELATIVE_ALLOWANCE: f64 = 64.0 * f64::EPSILON;

fn execute(binary: &Path, arguments: &[&str], input: &[u8]) -> Output {
    let mut child = Command::new(binary)
        .args(arguments)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap_or_else(|error| panic!("cannot execute {}: {error}", binary.display()));
    child.stdin.take().unwrap().write_all(input).unwrap();
    child.wait_with_output().unwrap()
}

fn compare_report(go: &Path, label: &str, input: &[u8], expected_success: bool) {
    compare_operation(
        go,
        &["reservoir", "predict", "-", "--format", "json"],
        label,
        input,
        expected_success,
    );
}

fn compare_operation(
    go: &Path,
    arguments: &[&str],
    label: &str,
    input: &[u8],
    expected_success: bool,
) {
    let native = execute(Path::new(env!("CARGO_BIN_EXE_fart")), arguments, input);
    let oracle = execute(go, arguments, input);
    for (name, output) in [("Rust", &native), ("Go", &oracle)] {
        assert_eq!(
            output.status.code(),
            Some(if expected_success { 0 } else { 1 }),
            "{label}: unexpected {name} status: {}",
            String::from_utf8_lossy(&output.stdout)
        );
        assert!(
            output.stderr.is_empty(),
            "{label}: {name} wrote unexpected stderr: {:?}",
            output.stderr
        );
    }
    let native: Value = serde_json::from_slice(&native.stdout).unwrap();
    let oracle: Value = serde_json::from_slice(&oracle.stdout).unwrap();
    compare(&native, &oracle).unwrap_or_else(|error| panic!("{label}: {error}"));
}

fn compare(native: &Value, oracle: &Value) -> Result<(), String> {
    compare_at(native, oracle, native, oracle, "")
}

fn compare_at(
    left: &Value,
    right: &Value,
    native: &Value,
    oracle: &Value,
    path: &str,
) -> Result<(), String> {
    if path == "/implementation_revision" {
        let (rust_revision, go_revision) = match native["schema"].as_str() {
            Some("fart.restriction-prediction/v0alpha1") => (
                "rust-restriction/v0alpha1",
                "go-oracle.restriction/v0alpha2",
            ),
            Some("fart.restriction-history/v0alpha1") => (
                "rust-restriction-history/v0alpha1",
                "go-oracle.restriction-history/v0alpha2",
            ),
            _ => (RUST_REVISION, GO_REVISION),
        };
        return if left == rust_revision && right == go_revision {
            Ok(())
        } else {
            Err(format!(
                "{path}: unexpected implementation revisions: {left}, {right}"
            ))
        };
    }
    match (left, right) {
        (Value::Object(left), Value::Object(right)) => {
            if left.keys().ne(right.keys()) {
                return Err(format!(
                    "{path}: different object members: {:?}, {:?}",
                    left.keys(),
                    right.keys()
                ));
            }
            for (key, value) in left {
                compare_at(
                    value,
                    &right[key],
                    native,
                    oracle,
                    &format!("{path}/{}", key.replace('~', "~0").replace('/', "~1")),
                )?;
            }
            Ok(())
        }
        (Value::Array(left), Value::Array(right)) => {
            if left.len() != right.len() {
                return Err(format!("{path}: different array lengths"));
            }
            for (index, (left, right)) in left.iter().zip(right).enumerate() {
                compare_at(left, right, native, oracle, &format!("{path}/{index}"))?;
            }
            Ok(())
        }
        (Value::Number(left), Value::Number(right)) => {
            let (left, right) = (left.as_f64().unwrap(), right.as_f64().unwrap());
            let allowance = if let Some((left_tolerance, right_tolerance)) =
                residual_bounds(path, native, oracle)
            {
                if left.abs() > left_tolerance || right.abs() > right_tolerance {
                    return Err(format!(
                        "{path}: residual exceeds its own arithmetic allowance"
                    ));
                }
                left_tolerance + right_tolerance
            } else {
                RELATIVE_ALLOWANCE * left.abs().max(right.abs()) + f64::from_bits(1)
            };
            if left.is_finite() && right.is_finite() && (left - right).abs() <= allowance {
                Ok(())
            } else {
                Err(format!(
                    "{path}: {left:e} differs from {right:e}; allowance {allowance:e}"
                ))
            }
        }
        _ if left == right => Ok(()),
        _ => Err(format!("{path}: different values: {left}, {right}")),
    }
}

fn residual_bounds(path: &str, native: &Value, oracle: &Value) -> Option<(f64, f64)> {
    let claim_index = match path {
        "/balances/total_mass_residual_kg" => Some(0),
        "/balances/energy_residual_j" => Some(1),
        "/balances/initial_eos_residual_j" => Some(2),
        "/balances/final_eos_residual_j" => Some(3),
        "/balances/mass_flow_residual_kg_per_s" => Some(0),
        "/balances/thrust_residual_n" => Some(1),
        "/balances/recoil_residual_n" => Some(2),
        "/totals/recoil_residual_n_s" => Some(0),
        _ => path
            .strip_prefix("/claims/")
            .and_then(|rest| rest.strip_suffix("/residual"))
            .and_then(|index| index.parse::<usize>().ok()),
    };
    if let Some(index) = claim_index {
        return Some((
            native["claims"][index]["tolerance"].as_f64()?,
            oracle["claims"][index]["tolerance"].as_f64()?,
        ));
    }
    let index = path
        .strip_prefix("/balances/components/")?
        .strip_suffix("/residual_kg")?
        .parse::<usize>()
        .ok()?;
    let tolerance = |report: &Value| -> Option<f64> {
        let magnitude = report["initial"]["components"][index]["mass_kg"]
            .as_f64()?
            .max(report["final"]["components"][index]["mass_kg"].as_f64()?)
            .max(report["transfers"]["components"][index]["mass_out_kg"].as_f64()?);
        Some(RELATIVE_ALLOWANCE * magnitude + f64::from_bits(1))
    };
    Some((tolerance(native)?, tolerance(oracle)?))
}

#[test]
fn full_report_comparison_does_not_hide_contract_or_small_quantity_drift() {
    let native = json!({"implementation_revision": RUST_REVISION, "quantity": 1e-100,
        "diagnostics": [{"path": "/initial", "reason_code": "invalid_token"}],
        "assumptions": ["one", "two"]});
    let mut oracle = native.clone();
    oracle["implementation_revision"] = json!(GO_REVISION);
    assert!(compare(&native, &oracle).is_ok());
    for (pointer, replacement) in [
        ("/implementation_revision", json!("unreviewed")),
        ("/diagnostics/0/path", json!("/")),
        ("/diagnostics/0/reason_code", json!("other")),
        ("/assumptions", json!(["two", "one"])),
        ("/quantity", json!(0.0)),
        ("/quantity", json!(1.001e-100)),
    ] {
        let mut changed = oracle.clone();
        *changed.pointer_mut(pointer).unwrap() = replacement;
        assert!(compare(&native, &changed).is_err(), "ignored {pointer}");
    }
    oracle["quantity"] = json!(1e-100 * (1.0 + 8.0 * f64::EPSILON));
    assert!(compare(&native, &oracle).is_ok());
    oracle["new_claim"] = json!("unreviewed");
    assert!(compare(&native, &oracle).is_err());
}

#[test]
fn go_oracle_parity_when_explicitly_enabled() {
    let Some(go) = env::var_os("FARTAPP_GO_ORACLE") else {
        eprintln!("Go comparison skipped: FARTAPP_GO_ORACLE is not set");
        return;
    };
    let go = Path::new(&go);
    assert!(
        go.is_file(),
        "explicit FARTAPP_GO_ORACLE is not an existing executable: {}",
        go.display()
    );
    let root = Path::new(env!("CARGO_MANIFEST_DIR")).join("../..");
    let mut fixtures: Vec<_> = fs::read_dir(root.join("testdata/reservoir"))
        .unwrap()
        .map(|entry| entry.unwrap().path())
        .filter(|path| {
            path.extension()
                .is_some_and(|extension| extension == "json")
        })
        .collect();
    fixtures.sort();
    assert!(
        !fixtures.is_empty(),
        "no shared reservoir fixtures discovered"
    );
    for fixture in &fixtures {
        compare_report(
            go,
            &fixture.display().to_string(),
            &fs::read(fixture).unwrap(),
            true,
        );
    }
    let fixture: Value = serde_json::from_slice(include_bytes!(
        "../../../testdata/reservoir/synthetic-mixture-adiabatic.json"
    ))
    .unwrap();
    for closure in ["rigid-adiabatic", "rigid-isothermal"] {
        for fraction in [0.0, 1e-10, 0.01, 0.2, 0.8, 0.99] {
            for scale in [1e-9, 1.0, 1e9] {
                let mut input = fixture.clone();
                input["closure"] = json!(closure);
                input["withdrawal_fraction"] = json!(fraction);
                input["initial"]["volume_m3"] = json!(scale);
                for component in input["initial"]["components"].as_array_mut().unwrap() {
                    component["mass_kg"] = json!(component["mass_kg"].as_f64().unwrap() * scale);
                }
                compare_report(
                    go,
                    &format!("grid {closure}/{fraction}/{scale}"),
                    &serde_json::to_vec(&input).unwrap(),
                    true,
                );
            }
        }
    }
    let mut extreme = fixture.clone();
    extreme["closure"] = json!("rigid-isothermal");
    extreme["withdrawal_fraction"] = json!(2.0_f64.powi(-20));
    extreme["initial"] = json!({"volume_m3": 2.0_f64.powi(-60), "temperature_k": 2.0_f64.powi(1000),
        "components": [{"id": "scaled", "mass_kg": 2.0_f64.powi(-500),
            "specific_gas_constant_j_per_kg_k": 2.0_f64.powi(-560),
            "isochoric_heat_capacity_j_per_kg_k": 2.0_f64.powi(-560)}]});
    compare_report(
        go,
        "exact scaled underflow regression",
        &serde_json::to_vec(&extreme).unwrap(),
        true,
    );
    let hostile = hostile_requests(&fixture);
    for (index, input) in hostile.iter().enumerate() {
        compare_report(go, &format!("hostile case {index}"), input, false);
    }
    for intensity in ["1", "2", "3", "4", "5"] {
        let native = execute(Path::new(env!("CARGO_BIN_EXE_fart")), &[intensity], b"");
        let oracle = execute(go, &[intensity], b"");
        assert_eq!(native.status.code(), Some(0));
        assert_eq!(native.status.code(), oracle.status.code());
        assert_eq!(native.stdout, oracle.stdout);
        assert_eq!(native.stderr, oracle.stderr);
        assert!(native.stderr.is_empty());
    }
    eprintln!(
        "Compared {} shared fixtures, 36 grid requests, one exact underflow request, {} hostile requests, and all five toy intensities",
        fixtures.len(),
        hostile.len()
    );
}

#[test]
fn restriction_and_history_full_go_reports_when_explicitly_enabled() {
    let Some(go) = env::var_os("FARTAPP_GO_ORACLE") else {
        eprintln!("Go comparison skipped: FARTAPP_GO_ORACLE is not set");
        return;
    };
    let go = Path::new(&go);
    assert!(go.is_file(), "explicit Go oracle must exist");
    let root = Path::new(env!("CARGO_MANIFEST_DIR")).join("../../testdata/restriction");
    let mut count = 0;
    let mut check = |operation: &str, label: &str, value: &[u8], success: bool| {
        compare_operation(
            go,
            &["restriction", operation, "-", "--format", "json"],
            label,
            value,
            success,
        );
        count += 1;
    };
    for name in [
        "gamma15-choked",
        "gamma15-subsonic",
        "linear-compliance-choked",
        "ordinary-pressure-subsonic",
        "gamma15-choked-history",
    ] {
        let operation = if name.ends_with("history") {
            "history"
        } else {
            "predict"
        };
        check(
            operation,
            name,
            &fs::read(root.join(format!("{name}.json"))).unwrap(),
            true,
        );
    }
    let prediction: Value =
        serde_json::from_slice(&fs::read(root.join("gamma15-choked.json")).unwrap()).unwrap();
    let history: Value =
        serde_json::from_slice(&fs::read(root.join("gamma15-choked-history.json")).unwrap())
            .unwrap();
    for gamma in [1.01, 1.4, 1.5, 5.0 / 3.0, 3.0] {
        for ratio in [1e-9, 0.3, 0.512, 0.7, 0.99, 1.0] {
            for area in [0.0, 1e-12, 0.01] {
                let mut input = prediction.clone();
                input["stagnation"]["heat_capacity_ratio"] = json!(gamma);
                input["back_pressure_pa"] = json!(125000.0 * ratio);
                input["area"]["prescribed_m2"] = json!(area);
                check(
                    "predict",
                    &format!("flow grid gamma={gamma} ratio={ratio} area={area}"),
                    &serde_json::to_vec(&input).unwrap(),
                    true,
                );
            }
        }
    }
    for size in [1, 2, 3, 64, 65, 256] {
        for ratio in [1e-9, 0.7, 1.0] {
            let mut input = history.clone();
            input["back_pressure_pa"] = json!(125000.0 * ratio);
            input["samples"] = json!((0..size).map(|index| json!({"time_s":f64::from(index) * 0.01,"prescribed_m2":if index % 3 == 0 { 0.0 } else { 0.001 * f64::from(index) }})).collect::<Vec<_>>());
            check(
                "history",
                &format!("history grid samples={size} ratio={ratio}"),
                &serde_json::to_vec(&input).unwrap(),
                true,
            );
        }
    }
    let mut high_temperature = prediction.clone();
    high_temperature["stagnation"]["temperature_k"] = json!(2.0_f64.powi(1023));
    high_temperature["stagnation"]["specific_gas_constant_j_per_kg_k"] = json!(2.0_f64.powi(-1020));
    check(
        "predict",
        "finite sonic temperature despite intermediate overflow",
        &serde_json::to_vec(&high_temperature).unwrap(),
        true,
    );
    let mut tiny_mach = prediction.clone();
    tiny_mach["stagnation"]["heat_capacity_ratio"] = json!(1e308);
    tiny_mach["back_pressure_pa"] = json!(125000.0_f64.next_down());
    check(
        "predict",
        "finite subnormal Mach despite squared underflow",
        &serde_json::to_vec(&tiny_mach).unwrap(),
        true,
    );
    for back in [64000.0_f64.next_down(), 64000.0, 64000.0_f64.next_up()] {
        let mut sonic_boundary = prediction.clone();
        sonic_boundary["back_pressure_pa"] = json!(back);
        sonic_boundary["discharge_coefficient"] = json!(1e-30);
        sonic_boundary["area"]["prescribed_m2"] = json!(1.0);
        check(
            "predict",
            &format!("sonic pressure boundary {back:?}"),
            &serde_json::to_vec(&sonic_boundary).unwrap(),
            true,
        );
    }
    let mut finite_enthalpy = history.clone();
    finite_enthalpy["stagnation"]["temperature_k"] = json!(1e-300);
    finite_enthalpy["stagnation"]["specific_gas_constant_j_per_kg_k"] = json!(1e308);
    finite_enthalpy["stagnation"]["heat_capacity_ratio"] = json!(1.0_f64.next_up());
    finite_enthalpy["samples"] =
        json!([{"time_s":0,"prescribed_m2":1e-10},{"time_s":1,"prescribed_m2":1e-10}]);
    check(
        "history",
        "finite integrated enthalpy despite overflowing cp",
        &serde_json::to_vec(&finite_enthalpy).unwrap(),
        true,
    );
    for (operation, fixture) in [("predict", prediction), ("history", history)] {
        for (index, input) in restriction_hostile_requests(&fixture, operation == "history")
            .iter()
            .enumerate()
        {
            check(
                operation,
                &format!(
                    "{operation} hostile {index}: {}",
                    String::from_utf8_lossy(input)
                        .chars()
                        .take(200)
                        .collect::<String>()
                ),
                input,
                false,
            );
        }
    }
    eprintln!(
        "Compared {count} complete restriction/history reports, including fixtures, grids, boundary regressions, and hostile requests"
    );
}

fn restriction_hostile_requests(fixture: &Value, history: bool) -> Vec<Vec<u8>> {
    let mut requests: Vec<Vec<u8>> = [
        "",
        " ",
        "{",
        "{}",
        "{} {}",
        "null",
        "[]",
        "true",
        r#"{"schema":1,"sch\u0065ma":2}"#,
        r#"{"x":"\ud800"}"#,
        r#"{"x":01e999}"#,
    ]
    .into_iter()
    .map(|text| text.as_bytes().to_vec())
    .collect();
    requests.extend([
        vec![0xff],
        vec![b' '; 65_537],
        format!("{}0{}", "[".repeat(34), "]".repeat(34)).into_bytes(),
        format!("{{\"{}\":0}}", "x".repeat(129)).into_bytes(),
    ]);
    let mut paths = vec![
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
    ];
    paths.extend(if history {
        vec![
            "/samples",
            "/samples/0",
            "/samples/0/time_s",
            "/samples/0/prescribed_m2",
        ]
    } else {
        vec!["/area", "/area/law", "/area/prescribed_m2"]
    });
    for path in paths {
        for value in [Value::Null, json!([])] {
            let mut input = fixture.clone();
            *input.pointer_mut(path).unwrap() = value;
            requests.push(serde_json::to_vec(&input).unwrap());
        }
        let (parent, key) = path.rsplit_once('/').unwrap();
        if let Some(object) = fixture.pointer(parent).and_then(Value::as_object) {
            assert!(object.contains_key(key));
            let mut input = fixture.clone();
            input
                .pointer_mut(parent)
                .unwrap()
                .as_object_mut()
                .unwrap()
                .remove(key);
            requests.push(serde_json::to_vec(&input).unwrap());
        }
    }
    for parent in if history {
        vec!["", "/model", "/stagnation", "/samples/0"]
    } else {
        vec!["", "/model", "/stagnation", "/area"]
    } {
        let mut input = fixture.clone();
        input
            .pointer_mut(parent)
            .unwrap()
            .as_object_mut()
            .unwrap()
            .insert("unknown/~".into(), json!(1));
        requests.push(serde_json::to_vec(&input).unwrap());
    }
    for (path, value) in [
        ("/schema", json!("unknown")),
        ("/model/id", json!("unknown")),
        ("/model/version", json!("v999")),
        ("/quantity_system", json!("cgs")),
        ("/stagnation/pressure_pa", json!(0)),
        ("/stagnation/temperature_k", json!(-1)),
        ("/stagnation/heat_capacity_ratio", json!(1)),
        ("/back_pressure_pa", json!(-1)),
        ("/discharge_coefficient", json!(1.1)),
        ("/discharge_coefficient", json!(0)),
    ] {
        let mut input = fixture.clone();
        *input.pointer_mut(path).unwrap() = value;
        requests.push(serde_json::to_vec(&input).unwrap());
    }
    let mut alias = fixture.clone();
    alias["Schema"] = alias.as_object_mut().unwrap().remove("schema").unwrap();
    requests.push(serde_json::to_vec(&alias).unwrap());
    let input = serde_json::to_string(fixture).unwrap();
    requests.push(input.replace("125000", "1e999").into_bytes());
    let mut invalid = fixture.clone();
    if history {
        invalid["samples"] = json!(vec![fixture["samples"][0].clone(); 257]);
        requests.push(serde_json::to_vec(&invalid).unwrap());
        invalid["samples"] = json!([]);
        requests.push(serde_json::to_vec(&invalid).unwrap());
    } else {
        invalid["area"]["law"] = json!("unknown");
        requests.push(serde_json::to_vec(&invalid).unwrap());
    }
    requests
}

fn hostile_requests(fixture: &Value) -> Vec<Vec<u8>> {
    let mut requests: Vec<Vec<u8>> = [
        "",
        " \r\n\t",
        "{",
        "{}",
        "{} {}",
        "null",
        "null trailing",
        "[]",
        "true",
        r#"{"schema":"one","schema":"two"}"#,
        r#"{"schema":"one","sch\u0065ma":"two"}"#,
        r#"{"x":"\ud800"}"#,
    ]
    .into_iter()
    .map(|text| text.as_bytes().to_vec())
    .collect();
    requests.push(vec![0xff]);
    requests.push(vec![b' '; 65_537]);
    requests.push(format!("{}0{}", "[".repeat(34), "]".repeat(34)).into_bytes());
    requests.push(format!("{{\"{}\":0}}", "x".repeat(129)).into_bytes());
    for (pointer, value) in [
        ("/schema", json!("unknown")),
        ("/model/id", json!("unknown")),
        ("/model/version", json!("v999")),
        ("/quantity_system", json!("cgs")),
        ("/closure", json!("implicit")),
        ("/withdrawal_fraction", json!(-0.1)),
        ("/withdrawal_fraction", json!(1.0)),
        ("/withdrawal_fraction", json!(1e-20)),
        ("/initial/volume_m3", json!(0.0)),
        ("/initial/temperature_k", json!(-1.0)),
        ("/initial/components", json!([])),
        ("/initial/components/0/id", json!("UPPER")),
        (
            "/initial/components/1/id",
            fixture["initial"]["components"][0]["id"].clone(),
        ),
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
        ("/initial", json!([])),
        ("/model", json!(1)),
    ] {
        let mut input = fixture.clone();
        *input.pointer_mut(pointer).unwrap() = value;
        requests.push(serde_json::to_vec(&input).unwrap());
    }
    for pointer in [
        "/schema",
        "/model",
        "/initial",
        "/initial/components",
        "/initial/components/0",
        "/initial/components/0/id",
        "/initial/components/0/mass_kg",
        "/withdrawal_fraction",
    ] {
        let mut input = fixture.clone();
        *input.pointer_mut(pointer).unwrap() = Value::Null;
        requests.push(serde_json::to_vec(&input).unwrap());
    }
    for (parent, key) in [
        ("", "schema"),
        ("", "model"),
        ("", "initial"),
        ("", "withdrawal_fraction"),
        ("/initial", "components"),
        ("/initial", "temperature_k"),
        ("/initial/components/0", "mass_kg"),
    ] {
        let mut input = fixture.clone();
        input
            .pointer_mut(parent)
            .unwrap()
            .as_object_mut()
            .unwrap()
            .remove(key);
        requests.push(serde_json::to_vec(&input).unwrap());
    }
    for parent in ["", "/model", "/initial", "/initial/components/0"] {
        let mut input = fixture.clone();
        input
            .pointer_mut(parent)
            .unwrap()
            .as_object_mut()
            .unwrap()
            .insert("unknown".to_owned(), json!(1));
        requests.push(serde_json::to_vec(&input).unwrap());
    }
    let mut alias = fixture.clone();
    alias["Schema"] = alias.as_object_mut().unwrap().remove("schema").unwrap();
    requests.push(serde_json::to_vec(&alias).unwrap());
    let mut large = fixture.clone();
    large["initial"]["components"] = json!(vec![fixture["initial"]["components"][0].clone(); 65]);
    requests.push(serde_json::to_vec(&large).unwrap());
    for (closure, fraction, mass, gas, cv, volume, temperature) in [
        (
            "rigid-isothermal",
            0.1,
            1.0,
            f64::from_bits(1),
            f64::from_bits(1),
            1.0,
            3.0,
        ),
        (
            "rigid-adiabatic",
            0.75,
            2.0_f64.powi(-500),
            2.0_f64.powi(533),
            2.0_f64.powi(523),
            2.0_f64.powi(10),
            2.0_f64.powi(1000),
        ),
    ] {
        let mut input = fixture.clone();
        input["closure"] = json!(closure);
        input["withdrawal_fraction"] = json!(fraction);
        input["initial"] = json!({"volume_m3": volume, "temperature_k": temperature,
            "components": [{"id": "a", "mass_kg": mass, "specific_gas_constant_j_per_kg_k": gas,
                "isochoric_heat_capacity_j_per_kg_k": cv}]});
        requests.push(serde_json::to_vec(&input).unwrap());
    }
    requests
}
