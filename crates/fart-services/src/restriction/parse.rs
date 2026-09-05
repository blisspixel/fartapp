use fart_domain::restriction::{
    Area, AreaCompliance, AreaLaw, DischargeCoefficient, HeatCapacityRatio, HistorySample,
    MAX_HISTORY_SAMPLES, Pressure, Request, Seconds, Stagnation,
};
use fart_domain::{SpecificGasConstant, Temperature};
use serde_json::Value;

use super::{HISTORY_REQUEST_SCHEMA, Kind, MODEL_ID, MODEL_VERSION, REQUEST_SCHEMA, wire};
use crate::Diagnostic;

pub(super) struct HistoryRequest {
    pub stagnation: Stagnation,
    pub back: Pressure,
    pub coefficient: DischargeCoefficient,
    pub samples: Vec<HistorySample>,
}

pub(super) fn prediction(data: &[u8]) -> Result<Request, Diagnostic> {
    let kind = Kind::Prediction;
    let document = wire::document(data, kind)?;
    header(&document, kind)?;
    let back = required(&document, "back_pressure_pa", "", kind)?;
    let coefficient = required(&document, "discharge_coefficient", "", kind)?;
    let stagnation = stagnation(&document["stagnation"], kind)?;
    let area = area(&document["area"])?;
    let back = Pressure::new(back)
        .map_err(|e| wire::issue(kind, "model", "/back_pressure_pa", e.reason()))?;
    let coefficient = DischargeCoefficient::new(coefficient)
        .map_err(|e| wire::issue(kind, "model", "/discharge_coefficient", e.reason()))?;
    Ok(Request::new(stagnation, back, area, coefficient))
}

pub(super) fn history(data: &[u8]) -> Result<HistoryRequest, Diagnostic> {
    let kind = Kind::History;
    let document = wire::document(data, kind)?;
    header(&document, kind)?;
    let back = required(&document, "back_pressure_pa", "", kind)?;
    let coefficient = required(&document, "discharge_coefficient", "", kind)?;
    let samples = document["samples"]
        .as_array()
        .filter(|values| !values.is_empty())
        .ok_or_else(|| wire::issue(kind, "schema", "/samples", "missing_member"))?;
    if samples.len() > MAX_HISTORY_SAMPLES {
        return Err(wire::issue(
            kind,
            "schema",
            "/samples",
            "invalid_sample_count",
        ));
    }
    let stagnation = stagnation(&document["stagnation"], kind)?;
    let back = Pressure::new(back)
        .map_err(|_| wire::issue(kind, "model", "/back_pressure_pa", "nonpositive_quantity"))?;
    let coefficient = DischargeCoefficient::new(coefficient).map_err(|_| {
        wire::issue(
            kind,
            "model",
            "/discharge_coefficient",
            "invalid_discharge_coefficient",
        )
    })?;
    let mut interpreted = Vec::with_capacity(samples.len());
    for (index, sample) in samples.iter().enumerate() {
        let path = format!("/samples/{index}");
        let (Some(time), Some(area)) =
            (sample["time_s"].as_f64(), sample["prescribed_m2"].as_f64())
        else {
            return Err(wire::issue(kind, "schema", &path, "missing_member"));
        };
        let time = Seconds::new(time)
            .map_err(|_| wire::issue(kind, "model", &format!("{path}/time_s"), "invalid_time"))?;
        let area = Area::new(area).map_err(|_| {
            wire::issue(
                kind,
                "model",
                &format!("{path}/prescribed_m2"),
                "negative_area",
            )
        })?;
        interpreted.push(HistorySample::new(time, area));
    }
    Ok(HistoryRequest {
        stagnation,
        back,
        coefficient,
        samples: interpreted,
    })
}

fn header(document: &Value, kind: Kind) -> Result<(), Diagnostic> {
    let (schema, path) = match kind {
        Kind::Prediction => (REQUEST_SCHEMA, "/schema"),
        Kind::History => (HISTORY_REQUEST_SCHEMA, "/"),
    };
    if document["schema"].as_str() != Some(schema) {
        return Err(wire::issue(kind, "schema", path, "unsupported_schema"));
    }
    if document["model"]["id"].as_str() != Some(MODEL_ID)
        || document["model"]["version"].as_str() != Some(MODEL_VERSION)
    {
        return Err(wire::issue(
            kind,
            "schema",
            "/model",
            "unsupported_model_revision",
        ));
    }
    if document["quantity_system"].as_str() != Some("si") {
        return Err(wire::issue(
            kind,
            "schema",
            "/quantity_system",
            "unsupported_quantity_system",
        ));
    }
    Ok(())
}

fn required(document: &Value, key: &str, base: &str, kind: Kind) -> Result<f64, Diagnostic> {
    document[key]
        .as_f64()
        .ok_or_else(|| wire::issue(kind, "schema", &format!("{base}/{key}"), "missing_member"))
}

fn stagnation(document: &Value, kind: Kind) -> Result<Stagnation, Diagnostic> {
    let fields = [
        "pressure_pa",
        "temperature_k",
        "specific_gas_constant_j_per_kg_k",
        "heat_capacity_ratio",
    ];
    let mut values = [0.0; 4];
    for (index, key) in fields.iter().enumerate() {
        values[index] = required(document, key, "/stagnation", kind).map_err(|error| {
            if kind == Kind::History {
                wire::issue(kind, "schema", "/stagnation", "missing_member")
            } else {
                error
            }
        })?;
    }
    let pressure = Pressure::new(values[0])
        .map_err(|e| wire::issue(kind, "model", "/stagnation/pressure_pa", e.reason()))?;
    let temperature = Temperature::new(values[1])
        .map_err(|e| wire::issue(kind, "model", "/stagnation/temperature_k", e.reason()))?;
    let gas_constant = SpecificGasConstant::new(values[2]).map_err(|e| {
        wire::issue(
            kind,
            "model",
            "/stagnation/specific_gas_constant_j_per_kg_k",
            e.reason(),
        )
    })?;
    let gamma = HeatCapacityRatio::new(values[3])
        .map_err(|e| wire::issue(kind, "model", "/stagnation/heat_capacity_ratio", e.reason()))?;
    Stagnation::new(pressure, temperature, gas_constant, gamma)
        .map_err(|e| wire::issue(kind, "model", "/stagnation", e.reason()))
}

fn area(document: &Value) -> Result<AreaLaw, Diagnostic> {
    let kind = Kind::Prediction;
    let prescribed = required(document, "prescribed_m2", "/area", kind)?;
    let prescribed = Area::new(prescribed)
        .map_err(|e| wire::issue(kind, "model", "/area/prescribed_m2", e.reason()))?;
    match document["law"].as_str().unwrap_or("") {
        "prescribed" => {
            for key in ["compliance_m2_per_pa", "maximum_m2"] {
                if document.get(key).is_some() {
                    return Err(wire::issue(
                        kind,
                        "schema",
                        &format!("/area/{key}"),
                        "unexpected_member",
                    ));
                }
            }
            Ok(AreaLaw::prescribed(prescribed))
        }
        "linear-compliance" => {
            let compliance = required(document, "compliance_m2_per_pa", "/area", kind)?;
            let maximum = required(document, "maximum_m2", "/area", kind)?;
            let compliance = AreaCompliance::new(compliance).map_err(|e| {
                wire::issue(kind, "model", "/area/compliance_m2_per_pa", e.reason())
            })?;
            let maximum = Area::new(maximum)
                .map_err(|e| wire::issue(kind, "model", "/area/maximum_m2", e.reason()))?;
            AreaLaw::linear_compliance(prescribed, compliance, maximum)
                .map_err(|e| wire::issue(kind, "model", "/area", e.reason()))
        }
        _ => Err(wire::issue(
            kind,
            "schema",
            "/area/law",
            "unsupported_area_law",
        )),
    }
}
