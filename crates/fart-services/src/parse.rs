use std::collections::BTreeSet;

use fart_domain::{
    Closure, Component, IsochoricHeatCapacity, MAX_COMPONENTS, Mass, ModelError, ReservoirState,
    SpecificGasConstant, Temperature, Volume, WithdrawalFraction, validate_component_id,
};
use serde_json::{Map, Value};

use crate::{Diagnostic, MODEL_ID, MODEL_VERSION, REQUEST_SCHEMA};

pub(crate) struct Request {
    pub state: ReservoirState,
    pub withdrawal: WithdrawalFraction,
    pub closure: Closure,
}

pub(crate) fn request(data: &[u8]) -> Result<Request, Diagnostic> {
    let value = crate::json::document(data, crate::json::Limits::RESERVOIR)?;
    shape(&value, Shape::Request)?;
    let root = object(&value)?;
    let empty = Map::new();
    if string(root, "schema") != REQUEST_SCHEMA {
        return Err(schema("/schema", "unsupported_schema"));
    }
    let model = root
        .get("model")
        .and_then(Value::as_object)
        .unwrap_or(&empty);
    if string(model, "id") != MODEL_ID || string(model, "version") != MODEL_VERSION {
        return Err(schema("/model", "unsupported_model_revision"));
    }
    if string(root, "quantity_system") != "si" {
        return Err(schema("/quantity_system", "unsupported_quantity_system"));
    }
    let withdrawal = WithdrawalFraction::new(number(root, "withdrawal_fraction", "")?)
        .map_err(|error| model_error("/withdrawal_fraction", error))?;
    let closure = match string(root, "closure") {
        "rigid-adiabatic" => Closure::RigidAdiabatic,
        "rigid-isothermal" => Closure::RigidIsothermal,
        _ => return Err(schema("/closure", "unsupported_closure")),
    };
    let state = initial_state(root.get("initial").unwrap_or(&Value::Null))?;
    fart_core::summarize(&state).map_err(|error| model_error("/initial", error))?;
    Ok(Request {
        state,
        withdrawal,
        closure,
    })
}

pub(crate) fn initial_state(value: &Value) -> Result<ReservoirState, Diagnostic> {
    let empty = Map::new();
    let initial = value.as_object().unwrap_or(&empty);
    // Omission remains a semantic refusal. Explicit null and every wrong type
    // were rejected as document shape errors before inspecting any values.
    let volume_value = number(initial, "volume_m3", "/initial")?;
    let temperature_value = number(initial, "temperature_k", "/initial")?;
    let volume =
        Volume::new(volume_value).map_err(|error| model_error("/initial/volume_m3", error))?;
    let temperature = Temperature::new(temperature_value)
        .map_err(|error| model_error("/initial/temperature_k", error))?;
    let components = initial
        .get("components")
        .and_then(Value::as_array)
        .map(Vec::as_slice)
        .unwrap_or(&[]);
    if components.is_empty() {
        return Err(schema("/initial/components", "missing_component"));
    }
    if components.len() > MAX_COMPONENTS {
        return Err(schema("/initial/components", "collection_limit_exceeded"));
    }
    let mut entries = Vec::with_capacity(components.len());
    let mut ids = BTreeSet::new();
    for (index, value) in components.iter().enumerate() {
        let part = object(value)?;
        let id = string(part, "id");
        let path = format!("/initial/components/{index}");
        validate_component_id(id).map_err(|error| schema(&format!("{path}/id"), error.reason()))?;
        if !ids.insert(id) {
            return Err(schema(&format!("{path}/id"), "duplicate_component_id"));
        }
        entries.push((id, path, part));
    }
    entries.sort_by_key(|entry| entry.0);
    let mut parts = Vec::with_capacity(entries.len());
    for (id, path, part) in entries {
        if [
            "mass_kg",
            "specific_gas_constant_j_per_kg_k",
            "isochoric_heat_capacity_j_per_kg_k",
        ]
        .iter()
        .any(|key| !part.contains_key(*key))
        {
            return Err(schema(&path, "missing_member"));
        }
        let mass = Mass::new(number(part, "mass_kg", &path)?)
            .map_err(|error| model_error(&format!("{path}/mass_kg"), error))?;
        let gas =
            SpecificGasConstant::new(number(part, "specific_gas_constant_j_per_kg_k", &path)?)
                .map_err(|error| {
                    model_error(&format!("{path}/specific_gas_constant_j_per_kg_k"), error)
                })?;
        let cv =
            IsochoricHeatCapacity::new(number(part, "isochoric_heat_capacity_j_per_kg_k", &path)?)
                .map_err(|error| {
                    model_error(&format!("{path}/isochoric_heat_capacity_j_per_kg_k"), error)
                })?;
        parts.push(Component::new(id, mass, gas, cv).map_err(|error| model_error(&path, error))?);
    }
    let state = ReservoirState::new(parts, volume, temperature)
        .map_err(|error| model_error("/initial", error))?;
    Ok(state)
}

fn schema(path: &str, reason: &'static str) -> Diagnostic {
    Diagnostic::new("schema", path, reason)
}
fn shape_error() -> Diagnostic {
    schema("/", "document_shape_invalid")
}
fn model_error(path: &str, error: ModelError) -> Diagnostic {
    Diagnostic::new("model", path, error.reason())
}

#[derive(Clone, Copy)]
enum Shape {
    Request,
    Model,
    Initial,
    Component,
    Components,
    String,
    Number,
}

fn shape(value: &Value, expected: Shape) -> Result<(), Diagnostic> {
    let members = match expected {
        Shape::String => return value.as_str().map(|_| ()).ok_or_else(shape_error),
        Shape::Number => return value.as_f64().map(|_| ()).ok_or_else(shape_error),
        Shape::Components => {
            for component in value.as_array().ok_or_else(shape_error)? {
                shape(component, Shape::Component)?;
            }
            return Ok(());
        }
        Shape::Request => &[
            ("schema", Shape::String),
            ("model", Shape::Model),
            ("quantity_system", Shape::String),
            ("closure", Shape::String),
            ("withdrawal_fraction", Shape::Number),
            ("initial", Shape::Initial),
        ][..],
        Shape::Model => &[("id", Shape::String), ("version", Shape::String)],
        Shape::Initial => &[
            ("components", Shape::Components),
            ("volume_m3", Shape::Number),
            ("temperature_k", Shape::Number),
        ],
        Shape::Component => &[
            ("id", Shape::String),
            ("mass_kg", Shape::Number),
            ("specific_gas_constant_j_per_kg_k", Shape::Number),
            ("isochoric_heat_capacity_j_per_kg_k", Shape::Number),
        ],
    };
    for (key, child) in object(value)? {
        let (_, expected) = members
            .iter()
            .find(|(name, _)| *name == key)
            .ok_or_else(shape_error)?;
        shape(child, *expected)?;
    }
    Ok(())
}

fn object(value: &Value) -> Result<&Map<String, Value>, Diagnostic> {
    value.as_object().ok_or_else(shape_error)
}

fn string<'a>(object: &'a Map<String, Value>, key: &str) -> &'a str {
    object.get(key).and_then(Value::as_str).unwrap_or("")
}

fn number(object: &Map<String, Value>, key: &str, path: &str) -> Result<f64, Diagnostic> {
    object
        .get(key)
        .ok_or_else(|| schema(&join(path, key), "missing_member"))?
        .as_f64()
        .ok_or_else(shape_error)
}

fn join(path: &str, key: &str) -> String {
    format!("{path}/{}", key.replace('~', "~0").replace('/', "~1"))
}
