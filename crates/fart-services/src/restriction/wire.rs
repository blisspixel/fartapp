use std::borrow::Cow;

use serde::de::IgnoredAny;
use serde_json::Value;

use super::{Kind, MAX_INPUT_BYTES, diagnostic};
use crate::{Diagnostic, json};

#[derive(Clone, Copy)]
enum Shape {
    Object(&'static [(&'static str, Shape)]),
    Array(&'static Shape),
    String,
    Number,
}

const MODEL: Shape = Shape::Object(&[("id", Shape::String), ("version", Shape::String)]);
const STAGNATION: Shape = Shape::Object(&[
    ("pressure_pa", Shape::Number),
    ("temperature_k", Shape::Number),
    ("specific_gas_constant_j_per_kg_k", Shape::Number),
    ("heat_capacity_ratio", Shape::Number),
]);
const AREA: Shape = Shape::Object(&[
    ("law", Shape::String),
    ("prescribed_m2", Shape::Number),
    ("compliance_m2_per_pa", Shape::Number),
    ("maximum_m2", Shape::Number),
]);
const SAMPLE: Shape = Shape::Object(&[("time_s", Shape::Number), ("prescribed_m2", Shape::Number)]);

pub(super) fn document(data: &[u8], kind: Kind) -> Result<Value, Diagnostic> {
    if data.len() > MAX_INPUT_BYTES {
        return Err(issue(kind, "input", "/", "input_too_large"));
    }
    // Match the oracle's global Unicode preflight, including its precedence
    // over a duplicate that appears earlier in the document. The JSON grammar
    // and actual string decoding remain serde_json's responsibility.
    if std::str::from_utf8(data).is_err() || !paired_unicode_escapes(data) {
        return Err(issue(kind, "syntax", "/", "malformed_json"));
    }
    // The Go syntax pass accepts arbitrary JSON number magnitudes. Its typed
    // binary64 decode happens only after the complete shape pass. serde_json's
    // Value rejects overflow earlier, so only a shape-validation copy replaces
    // syntactically valid overflow tokens. The flag unconditionally refuses the
    // request before this function can return any replacement to the model.
    let (shape_bytes, overflow) = shape_bytes(data);
    let value = json::document(&shape_bytes, json::Limits::RESERVOIR)
        .map_err(|error| issue(kind, "syntax", &error.path, error.reason_code))?;
    let operation = match kind {
        Kind::Prediction => ("area", AREA),
        Kind::History => ("samples", Shape::Array(&SAMPLE)),
    };
    let fields = [
        ("schema", Shape::String),
        ("model", MODEL),
        ("quantity_system", Shape::String),
        ("stagnation", STAGNATION),
        ("back_pressure_pa", Shape::Number),
        ("discharge_coefficient", Shape::Number),
        operation,
    ];
    object_shape(&value, &fields, "")
        .map_err(|path| issue(kind, "schema", &path, "document_shape_invalid"))?;
    if overflow {
        return Err(issue(kind, "schema", "/", "document_shape_invalid"));
    }
    Ok(value)
}

fn object_shape(value: &Value, fields: &[(&str, Shape)], path: &str) -> Result<(), String> {
    let object = value.as_object().ok_or_else(|| path.to_owned())?;
    // serde_json's default Map is a BTreeMap. Sorting explicitly also keeps the
    // diagnostic order independent of a future feature change in that crate.
    let mut keys: Vec<_> = object.keys().collect();
    keys.sort_unstable();
    for key in keys {
        let member_path = format!("{path}/{}", key.replace('~', "~0").replace('/', "~1"));
        let expected = fields
            .iter()
            .find(|(name, _)| *name == key)
            .map(|(_, shape)| *shape)
            .ok_or_else(|| member_path.clone())?;
        shape(&object[key], expected, &member_path)?;
    }
    Ok(())
}

fn shape(value: &Value, expected: Shape, path: &str) -> Result<(), String> {
    match expected {
        Shape::Object(fields) => object_shape(value, fields, path),
        Shape::Array(item) => {
            let values = value.as_array().ok_or_else(|| path.to_owned())?;
            for (index, value) in values.iter().enumerate() {
                shape(value, *item, &format!("{path}/{index}"))?;
            }
            Ok(())
        }
        Shape::String if value.is_string() => Ok(()),
        Shape::Number if value.is_number() => Ok(()),
        _ => Err(path.to_owned()),
    }
}

fn shape_bytes(data: &[u8]) -> (Cow<'_, [u8]>, bool) {
    let mut copy = Vec::new();
    let mut copied = 0;
    let mut index = 0;
    while index < data.len() {
        if data[index] == b'"' {
            index += 1;
            while index < data.len() {
                match data[index] {
                    b'\\' => index = (index + 2).min(data.len()),
                    b'"' => {
                        index += 1;
                        break;
                    }
                    _ => index += 1,
                }
            }
        } else if data[index] == b'-' || data[index].is_ascii_digit() {
            let start = index;
            while index < data.len() && !b" \t\r\n{}[],:\"".contains(&data[index]) {
                index += 1;
            }
            let token = &data[start..index];
            let overflows = std::str::from_utf8(token)
                .ok()
                .and_then(|text| text.parse::<f64>().ok())
                .is_some_and(|number| !number.is_finite());
            // IgnoredAny validates JSON's actual number grammar without trying
            // to store a binary64. For example, 01e999 must remain malformed.
            if overflows && serde_json::from_slice::<IgnoredAny>(token).is_ok() {
                copy.extend_from_slice(&data[copied..start]);
                copy.push(b'0');
                copied = index;
            }
        } else {
            index += 1;
        }
    }
    if copied == 0 {
        (Cow::Borrowed(data), false)
    } else {
        copy.extend_from_slice(&data[copied..]);
        (Cow::Owned(copy), true)
    }
}

fn paired_unicode_escapes(data: &[u8]) -> bool {
    let mut in_string = false;
    let mut index = 0;
    while index < data.len() {
        if data[index] == b'"' {
            in_string = !in_string;
        } else if data[index] == b'\\' && in_string && index + 1 < data.len() {
            if data[index + 1] != b'u' {
                index += 1;
            } else if let Some(value) = hex_quad(data, index + 2) {
                if (0xdc00..=0xdfff).contains(&value) {
                    return false;
                }
                if (0xd800..=0xdbff).contains(&value) {
                    if data.get(index + 6..index + 8) != Some(b"\\u")
                        || !hex_quad(data, index + 8)
                            .is_some_and(|low| (0xdc00..=0xdfff).contains(&low))
                    {
                        return false;
                    }
                    index += 11;
                } else {
                    index += 5;
                }
            }
        }
        index += 1;
    }
    true
}

fn hex_quad(data: &[u8], start: usize) -> Option<u16> {
    data.get(start..start + 4)?
        .iter()
        .try_fold(0_u16, |value, byte| {
            char::from(*byte)
                .to_digit(16)
                .map(|digit| (value << 4) | digit as u16)
        })
}

pub(super) fn issue(
    kind: Kind,
    stage: &'static str,
    path: &str,
    reason: &'static str,
) -> Diagnostic {
    let code = match (kind, stage) {
        (Kind::Prediction, "input") => "FART-E-INPUT-0003",
        (Kind::History, "input") => "FART-E-INPUT-0004",
        (Kind::Prediction, "syntax") => "FART-E-JSON-0003",
        (Kind::History, "syntax") => "FART-E-JSON-0004",
        (Kind::Prediction, "schema") => "FART-E-SCHEMA-0003",
        (Kind::History, "schema") => "FART-E-SCHEMA-0004",
        (Kind::Prediction, _) => "FART-E-MODEL-0003",
        (Kind::History, _) => "FART-E-MODEL-0005",
    };
    diagnostic(code, stage, path, reason)
}
