use fart_domain::ReservoirState;
use serde_json::{Map, Value, json};

use super::{BASELINE_SCHEMA, COMMAND_SCHEMA, MAX_ATTEMPTS, MAX_COMMAND_BYTES, PROFILE, Rejection};

#[derive(Clone, Debug)]
pub(super) struct Command {
    pub value: Value,
    pub operation: Operation,
}

#[derive(Clone, Debug)]
pub(super) enum Operation {
    Start { state: ReservoirState, budget: u32 },
    Predict { fraction: f64, closure: String },
    Observe { research: bool },
    Actions,
    Finish,
}

impl Command {
    pub fn text(&self, key: &str) -> &str {
        self.value[key].as_str().expect("typed command string")
    }
    pub fn revision(&self) -> u32 {
        self.value["expected_revision"]
            .as_u64()
            .expect("typed revision") as u32
    }
    pub fn key(&self) -> Option<&str> {
        self.value.get("idempotency_key").and_then(Value::as_str)
    }
}

pub(super) fn decode(data: &[u8]) -> Result<Command, Rejection> {
    if data.len() > MAX_COMMAND_BYTES {
        return Err(Rejection::new("input_too_large", "/"));
    }
    let value = crate::json::document(data, crate::json::Limits::COMMAND)?;
    command(value)
}

pub(super) fn command(mut value: Value) -> Result<Command, Rejection> {
    let root = object(&value, "/")?;
    exact_string(root, "schema", COMMAND_SCHEMA)?;
    let operation = text(root, "operation")?;
    let common = ["schema", "operation", "actor_id"];
    token(text(root, "actor_id")?, "/actor_id")?;
    let mut fields = common.to_vec();
    match operation {
        "start" => fields.extend([
            "profile",
            "role",
            "session_nonce",
            "idempotency_key",
            "expected_revision",
            "attempt_budget",
            "measurement_interaction",
            "knowledge_policy",
            "termination_policy",
            "baseline",
        ]),
        "predict" => fields.extend([
            "session_ref",
            "idempotency_key",
            "expected_revision",
            "withdrawal_fraction",
            "closure",
        ]),
        "observe" => fields.extend(["session_ref", "view"]),
        "actions" => fields.push("session_ref"),
        "finish" => fields.extend(["session_ref", "idempotency_key", "expected_revision"]),
        _ => return Err(Rejection::new("unsupported_operation", "/operation")),
    }
    closed(root, &fields, "/")?;
    if operation != "start" {
        fingerprint(text(root, "session_ref")?, "/session_ref")?;
    }
    if matches!(operation, "start" | "predict" | "finish") {
        token(text(root, "idempotency_key")?, "/idempotency_key")?;
        bounded_integer(root, "expected_revision", 0, MAX_ATTEMPTS + 1)?;
    }
    let operation = match operation {
        "start" => {
            for (key, expected) in [
                ("profile", PROFILE),
                ("role", "operator"),
                ("measurement_interaction", "none"),
                ("knowledge_policy", "full-reservoir"),
                ("termination_policy", "explicit-finish-or-budget"),
            ] {
                exact_string(root, key, expected)?;
            }
            token(text(root, "session_nonce")?, "/session_nonce")?;
            let budget = bounded_integer(root, "attempt_budget", 1, MAX_ATTEMPTS)?;
            let (baseline, state) = baseline(&root["baseline"])?;
            value["baseline"] = baseline;
            Operation::Start { state, budget }
        }
        "predict" => {
            let fraction = root["withdrawal_fraction"]
                .as_f64()
                .ok_or_else(|| Rejection::new("document_shape_invalid", "/withdrawal_fraction"))?;
            let fraction = if fraction == 0.0 { 0.0 } else { fraction };
            let closure = text(root, "closure")?.to_owned();
            token(&closure, "/closure")?;
            value["withdrawal_fraction"] = json!(fraction);
            Operation::Predict { fraction, closure }
        }
        "observe" => {
            let research = match text(root, "view")? {
                "brief" => false,
                "research" => true,
                _ => return Err(Rejection::new("unsupported_view", "/view")),
            };
            Operation::Observe { research }
        }
        "actions" => Operation::Actions,
        _ => Operation::Finish,
    };
    Ok(Command { value, operation })
}

fn baseline(value: &Value) -> Result<(Value, ReservoirState), Rejection> {
    let root = object(value, "/baseline")?;
    closed(
        root,
        &["schema", "model", "quantity_system", "initial"],
        "/baseline",
    )?;
    exact_string(root, "schema", BASELINE_SCHEMA).map_err(|e| prefix(e, "/baseline"))?;
    exact_string(root, "quantity_system", "si").map_err(|e| prefix(e, "/baseline"))?;
    let model = object(&root["model"], "/baseline/model")?;
    closed(model, &["id", "version"], "/baseline/model")?;
    exact_string(model, "id", crate::MODEL_ID).map_err(|e| prefix(e, "/baseline/model"))?;
    exact_string(model, "version", crate::MODEL_VERSION)
        .map_err(|e| prefix(e, "/baseline/model"))?;
    let initial = object(&root["initial"], "/baseline/initial")?;
    closed(
        initial,
        &["components", "volume_m3", "temperature_k"],
        "/baseline/initial",
    )?;
    let components = initial["components"]
        .as_array()
        .ok_or_else(|| Rejection::new("document_shape_invalid", "/baseline/initial/components"))?;
    for (index, component) in components.iter().enumerate() {
        let path = format!("/baseline/initial/components/{index}");
        let part = object(component, &path)?;
        closed(
            part,
            &[
                "id",
                "mass_kg",
                "specific_gas_constant_j_per_kg_k",
                "isochoric_heat_capacity_j_per_kg_k",
            ],
            &path,
        )?;
        text(part, "id").map_err(|e| prefix(e, &path))?;
        for key in [
            "mass_kg",
            "specific_gas_constant_j_per_kg_k",
            "isochoric_heat_capacity_j_per_kg_k",
        ] {
            if part[key].as_f64().is_none() {
                return Err(Rejection::new(
                    "document_shape_invalid",
                    &format!("{path}/{key}"),
                ));
            }
        }
    }
    for key in ["volume_m3", "temperature_k"] {
        if initial[key].as_f64().is_none() {
            return Err(Rejection::new(
                "document_shape_invalid",
                &format!("/baseline/initial/{key}"),
            ));
        }
    }
    let state =
        crate::parse::initial_state(&root["initial"]).map_err(|e| prefix(e.into(), "/baseline"))?;
    // Canonical typed baseline sorting and binary64 interpretation do not call
    // the solver. Live start separately admits derived initial properties.
    let components: Vec<Value> = state
        .components()
        .iter()
        .map(|part| {
            json!({
                "id":part.id(),"mass_kg":part.mass().get(),
                "specific_gas_constant_j_per_kg_k":part.gas_constant().get(),
                "isochoric_heat_capacity_j_per_kg_k":part.heat_capacity().get(),
            })
        })
        .collect();
    Ok((
        json!({"schema":BASELINE_SCHEMA,"model":{"id":crate::MODEL_ID,"version":crate::MODEL_VERSION},
        "quantity_system":"si","initial":{"components":components,"volume_m3":state.volume().get(),"temperature_k":state.temperature().get()}}),
        state,
    ))
}

fn prefix(mut issue: Rejection, path: &str) -> Rejection {
    issue.path = format!("{path}{}", issue.path);
    issue
}

pub(super) fn object<'a>(
    value: &'a Value,
    path: &str,
) -> Result<&'a Map<String, Value>, Rejection> {
    value
        .as_object()
        .ok_or_else(|| Rejection::new("document_shape_invalid", path))
}
pub(super) fn closed(
    root: &Map<String, Value>,
    fields: &[&str],
    path: &str,
) -> Result<(), Rejection> {
    if root.len() != fields.len() || fields.iter().any(|field| !root.contains_key(*field)) {
        return Err(Rejection::new("document_shape_invalid", path));
    }
    Ok(())
}
pub(super) fn text<'a>(root: &'a Map<String, Value>, key: &str) -> Result<&'a str, Rejection> {
    root.get(key)
        .and_then(Value::as_str)
        .ok_or_else(|| Rejection::new("document_shape_invalid", &format!("/{key}")))
}
fn exact_string(root: &Map<String, Value>, key: &str, expected: &str) -> Result<(), Rejection> {
    if text(root, key)? != expected {
        return Err(Rejection::new(
            "unsupported_profile_value",
            &format!("/{key}"),
        ));
    }
    Ok(())
}
fn token(value: &str, path: &str) -> Result<(), Rejection> {
    if value.is_empty()
        || value.len() > 64
        || !value
            .bytes()
            .all(|b| b.is_ascii_lowercase() || b.is_ascii_digit() || b"._:-".contains(&b))
    {
        return Err(Rejection::new("invalid_token", path));
    }
    Ok(())
}
pub(super) fn fingerprint(value: &str, path: &str) -> Result<(), Rejection> {
    if value.len() != 71
        || !value.starts_with("sha256:")
        || !value[7..]
            .bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
    {
        return Err(Rejection::new("invalid_fingerprint", path));
    }
    Ok(())
}
fn bounded_integer(
    root: &Map<String, Value>,
    key: &str,
    min: u32,
    max: u32,
) -> Result<u32, Rejection> {
    root[key]
        .as_u64()
        .filter(|value| (u64::from(min)..=u64::from(max)).contains(value))
        .map(|v| v as u32)
        .ok_or_else(|| Rejection::new("invalid_bounded_integer", &format!("/{key}")))
}
