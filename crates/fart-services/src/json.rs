//! Shared grammar checks; each protocol selects its own allocation limits.

use std::fmt;

use serde::de::{self, DeserializeSeed, MapAccess, SeqAccess, Visitor};
use serde_json::{Map, Number, Value};

use crate::Diagnostic;

#[derive(Clone, Copy)]
pub(crate) struct Limits {
    pub nodes: usize,
    pub array: usize,
    pub journal: usize,
}

impl Limits {
    pub const RESERVOIR: Self = Self {
        nodes: usize::MAX,
        array: usize::MAX,
        journal: usize::MAX,
    };
    pub const COMMAND: Self = Self {
        nodes: 4096,
        array: 64,
        journal: 17,
    };
    pub const TRANSCRIPT: Self = Self {
        nodes: 65_536,
        array: 64,
        journal: 17,
    };
}

pub(crate) fn document(data: &[u8], limits: Limits) -> Result<Value, Diagnostic> {
    if data.iter().all(|byte| b" \t\r\n".contains(byte)) {
        return Err(Diagnostic::new("syntax", "/", "empty_input"));
    }
    let mut issue = None;
    let mut remaining = limits.nodes;
    let mut decoder = serde_json::Deserializer::from_slice(data);
    let parsed = StrictValue {
        depth: 0,
        path: String::new(),
        issue: &mut issue,
        remaining: &mut remaining,
        limits,
        allowed: true,
    }
    .deserialize(&mut decoder);
    let value = parsed
        .map_err(|_| issue.unwrap_or_else(|| Diagnostic::new("syntax", "/", "malformed_json")))?;
    decoder
        .end()
        .map_err(|_| Diagnostic::new("syntax", "/", "trailing_json_value"))?;
    Ok(value)
}

struct StrictValue<'a> {
    depth: usize,
    path: String,
    issue: &'a mut Option<Diagnostic>,
    remaining: &'a mut usize,
    limits: Limits,
    allowed: bool,
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
        if !self.allowed {
            return Err(self.refuse("collection_limit_exceeded"));
        }
        if self.depth > 32 {
            return Err(self.refuse("maximum_depth_exceeded"));
        }
        if *self.remaining == 0 {
            return Err(self.refuse("node_limit_exceeded"));
        }
        *self.remaining -= 1;
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
        let limit = if self.path == "/journal" {
            self.limits.journal
        } else {
            self.limits.array
        };
        // The seed refuses the first excess element before its value is decoded.
        while let Some(value) = sequence.next_element_seed(StrictValue {
            depth: self.depth + 1,
            path: format!("{}/{}", self.path, values.len()),
            issue: &mut *self.issue,
            remaining: &mut *self.remaining,
            limits: self.limits,
            allowed: values.len() < limit,
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
            let path = join(&self.path, &key);
            if values.contains_key(&key) {
                *self.issue = Some(Diagnostic::new("syntax", &path, "duplicate_member"));
                return Err(de::Error::custom("duplicate_member"));
            }
            let value = access.next_value_seed(StrictValue {
                depth: self.depth + 1,
                path,
                issue: &mut *self.issue,
                remaining: &mut *self.remaining,
                limits: self.limits,
                allowed: true,
            })?;
            values.insert(key, value);
        }
        Ok(Value::Object(values))
    }
}

fn join(path: &str, key: &str) -> String {
    format!("{path}/{}", key.replace('~', "~0").replace('/', "~1"))
}
