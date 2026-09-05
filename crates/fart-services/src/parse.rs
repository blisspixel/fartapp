use std::{collections::BTreeSet, fmt};

use fart_domain::{
    Closure, Component, IsochoricHeatCapacity, MAX_COMPONENTS, Mass, ModelError, ReservoirState,
    SpecificGasConstant, Temperature, Volume, WithdrawalFraction, validate_component_id,
};
use serde::de::{self, DeserializeSeed, MapAccess, SeqAccess, Visitor};
use serde_json::{Map, Number, Value};

use crate::{Diagnostic, MODEL_ID, MODEL_VERSION, REQUEST_SCHEMA};

pub(crate) struct Request {
    pub state: ReservoirState,
    pub withdrawal: WithdrawalFraction,
    pub closure: Closure,
}

pub(crate) fn request(data: &[u8]) -> Result<Request, Diagnostic> {
    if data.iter().all(|byte| b" \t\r\n".contains(byte)) {
        return Err(Diagnostic::new("syntax", "/", "empty_input"));
    }
    let mut issue = None;
    let mut decoder = serde_json::Deserializer::from_slice(data);
    let parsed = StrictValue {
        depth: 0,
        path: String::new(),
        issue: &mut issue,
    }
    .deserialize(&mut decoder);
    let value = parsed
        .map_err(|_| issue.unwrap_or_else(|| Diagnostic::new("syntax", "/", "malformed_json")))?;
    decoder
        .end()
        .map_err(|_| Diagnostic::new("syntax", "/", "trailing_json_value"))?;
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
    let initial = root
        .get("initial")
        .and_then(Value::as_object)
        .unwrap_or(&empty);
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
    fart_core::summarize(&state).map_err(|error| model_error("/initial", error))?;
    Ok(Request {
        state,
        withdrawal,
        closure,
    })
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

// Serde owns JSON grammar and Unicode. The byte-limited tree preserves evidence
// long enough to check syntax and duplicates before closed, nonnull shape policy.
struct StrictValue<'a> {
    depth: usize,
    path: String,
    issue: &'a mut Option<Diagnostic>,
}

impl StrictValue<'_> {
    fn refuse<E: de::Error>(&mut self, reason: &'static str) -> E {
        *self.issue = Some(Diagnostic::new("syntax", &self.path, reason));
        E::custom(reason)
    }
}

impl<'de> DeserializeSeed<'de> for StrictValue<'_> {
    type Value = Value;
    fn deserialize<D: de::Deserializer<'de>>(mut self, deserializer: D) -> Result<Value, D::Error> {
        if self.depth > 32 {
            return Err(self.refuse("maximum_depth_exceeded"));
        }
        deserializer.deserialize_any(self)
    }
}

impl<'de> Visitor<'de> for StrictValue<'_> {
    type Value = Value;
    fn expecting(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        formatter.write_str("bounded JSON without duplicate members")
    }
    fn visit_bool<E: de::Error>(self, value: bool) -> Result<Value, E> {
        Ok(Value::Bool(value))
    }
    fn visit_i64<E: de::Error>(self, value: i64) -> Result<Value, E> {
        Ok(Value::Number(value.into()))
    }
    fn visit_u64<E: de::Error>(self, value: u64) -> Result<Value, E> {
        Ok(Value::Number(value.into()))
    }
    fn visit_f64<E: de::Error>(mut self, value: f64) -> Result<Value, E> {
        Number::from_f64(value)
            .map(Value::Number)
            .ok_or_else(|| self.refuse("malformed_json"))
    }
    fn visit_str<E: de::Error>(self, value: &str) -> Result<Value, E> {
        Ok(Value::String(value.to_owned()))
    }
    fn visit_string<E: de::Error>(self, value: String) -> Result<Value, E> {
        Ok(Value::String(value))
    }
    fn visit_unit<E: de::Error>(self) -> Result<Value, E> {
        Ok(Value::Null)
    }
    fn visit_seq<A: SeqAccess<'de>>(self, mut sequence: A) -> Result<Value, A::Error> {
        let mut values = Vec::new();
        while let Some(value) = sequence.next_element_seed(StrictValue {
            depth: self.depth + 1,
            path: format!("{}/{}", self.path, values.len()),
            issue: &mut *self.issue,
        })? {
            values.push(value);
        }
        Ok(Value::Array(values))
    }
    fn visit_map<A: MapAccess<'de>>(mut self, mut access: A) -> Result<Value, A::Error> {
        let mut values = Map::new();
        while let Some(key) = access.next_key::<String>()? {
            if key.len() > 128 {
                return Err(self.refuse("member_name_too_long"));
            }
            if values.contains_key(&key) {
                *self.issue = Some(Diagnostic::new(
                    "syntax",
                    &join(&self.path, &key),
                    "duplicate_member",
                ));
                return Err(de::Error::custom("duplicate_member"));
            }
            let value = access.next_value_seed(StrictValue {
                depth: self.depth + 1,
                path: join(&self.path, &key),
                issue: &mut *self.issue,
            })?;
            values.insert(key, value);
        }
        Ok(Value::Object(values))
    }
}
