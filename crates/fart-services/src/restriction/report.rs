use std::fmt::Write;

use fart_core::{
    BalanceClaim,
    restriction::{FlowResult, HistoryResult},
};
use serde_json::{Value, json};

use super::{
    HISTORY_IMPLEMENTATION_REVISION, HISTORY_REPORT_SCHEMA, HISTORY_REQUEST_SCHEMA,
    IMPLEMENTATION_REVISION, Kind, MODEL_ID, MODEL_VERSION, Outcome, REPORT_SCHEMA, REQUEST_SCHEMA,
    Report,
};
use crate::presentation::human_number;

pub(super) fn claims_valid(report: &Report) -> bool {
    claims(report).iter().all(|claim| {
        claim.residual.is_finite()
            && claim.tolerance.is_finite()
            && claim.tolerance >= 0.0
            && claim.residual.abs() <= claim.tolerance
    })
}

fn claims(report: &Report) -> &[BalanceClaim] {
    match &report.outcome {
        Outcome::Prediction(result) => &result.claims,
        Outcome::History(result) => &result.claims,
        Outcome::Invalid(_) => &[],
    }
}

fn identity(kind: Kind) -> (&'static str, &'static str, &'static str) {
    match kind {
        Kind::Prediction => (REPORT_SCHEMA, REQUEST_SCHEMA, IMPLEMENTATION_REVISION),
        Kind::History => (
            HISTORY_REPORT_SCHEMA,
            HISTORY_REQUEST_SCHEMA,
            HISTORY_IMPLEMENTATION_REVISION,
        ),
    }
}

pub(super) fn value(report: &Report) -> Value {
    let (schema, request_schema, implementation) = identity(report.kind);
    let environment = json!({"consulted_inputs":report.consulted_inputs,"ambient_inputs":[]});
    if let Outcome::Invalid(error) = &report.outcome {
        return json!({"schema":schema,"status":"invalid","implementation_revision":implementation,
            "validation_environment":environment,
            "diagnostics":[{"code":error.code,"stage":error.stage,"path":error.path,"reason_code":error.reason_code}]});
    }
    let mut value = match &report.outcome {
        Outcome::Prediction(result) => prediction(result),
        Outcome::History(result) => history(result),
        Outcome::Invalid(_) => unreachable!("invalid reports returned above"),
    };
    value["schema"] = json!(schema);
    value["status"] = json!("predicted");
    value["request_schema"] = json!(request_schema);
    value["model"] = json!({"id":MODEL_ID,"version":MODEL_VERSION});
    value["implementation_revision"] = json!(implementation);
    value["quantity_system"] = json!("si");
    value["validation_environment"] = environment;
    value["claims"] =
        json!(claims(report).iter().map(|claim| json!({
        "id":claim.id,"status":"satisfied-within-roundoff","method":claim.method,
        "equation_revision":format!("{MODEL_ID}@{MODEL_VERSION}"),
        "residual":claim.residual,"tolerance":claim.tolerance,"residual_unit":claim.unit,
    })).collect::<Vec<_>>());
    value
}

fn prediction(result: &FlowResult) -> Value {
    let request = result.request;
    let stagnation = request.stagnation();
    let law = request.area();
    let mut area = json!({"law":law.name(),"prescribed_m2":law.prescribed_area().get(),"effective_m2":result.effective_area_m2});
    if law.name() == "linear-compliance" {
        area["compliance_m2_per_pa"] = json!(law.compliance().get());
        area["maximum_m2"] = json!(law.maximum().get());
    }
    json!({
        "stagnation":{
            "pressure_pa":stagnation.pressure().get(),"temperature_k":stagnation.temperature().get(),
            "specific_gas_constant_j_per_kg_k":stagnation.gas_constant().get(),"heat_capacity_ratio":stagnation.gamma().get()
        },
        "back_pressure_pa":request.back_pressure().get(),"discharge_coefficient":request.discharge_coefficient().get(),
        "area":area,
        "flow":{
            "regime":result.regime.name(),"critical_pressure_ratio":result.critical_pressure_ratio,
            "back_pressure_ratio":result.back_pressure_ratio,"throat_mach":result.throat_mach,
            "exit_pressure_pa":result.exit_pressure_pa,"exit_temperature_k":result.exit_temperature_k,
            "exit_speed_m_per_s":result.exit_speed_m_per_s,"mass_flow_kg_per_s":result.mass_flow_kg_per_s,
            "sonic_mass_flow_kg_per_s":result.sonic_mass_flow_kg_per_s,"thrust_n":result.thrust_n,"recoil_n":result.recoil_n
        },
        "balances":{
            "mass_flow_residual_kg_per_s":result.mass_flow_residual_kg_per_s,
            "thrust_residual_n":result.thrust_residual_n,"recoil_residual_n":result.recoil_residual_n
        },
        "assumptions":["calorically-perfect-gas","quasi-steady-flow","isentropic-control-section",
            "converging-restriction-only","discharge-coefficient-scales-mass-flow-only","no-reverse-flow",
            "no-shock-inside-restriction","prescribed-or-linear-quasi-static-area"],
        "nonclaims":{
            "model":["elapsed-time-history","reservoir-mass-energy-coupling","shock-containing-or-underexpanded-plume",
                "viscous-resolved-vena-contracta","phase-change-and-reaction","acoustics"],
            "operation":["case-commitment","certificate-issuance"],"evidence":["empirical-validation"]
        }
    })
}

fn history(result: &HistoryResult) -> Value {
    let samples: Vec<_> = result.samples.iter().map(|sample| {
        let flow = &sample.flow;
        json!({"time_s":sample.time_s,"prescribed_m2":flow.request.area().prescribed_area().get(),
            "effective_m2":flow.effective_area_m2,"regime":flow.regime.name(),"exit_pressure_pa":flow.exit_pressure_pa,
            "mass_flow_kg_per_s":flow.mass_flow_kg_per_s,"thrust_n":flow.thrust_n,"recoil_n":flow.recoil_n})
    }).collect();
    json!({"samples":samples,
        "totals":{"mass_out_kg":result.mass_out_kg,"enthalpy_out_j":result.enthalpy_out_j,
            "kinetic_energy_out_j":result.kinetic_energy_out_j,"total_enthalpy_out_j":result.total_enthalpy_out_j,
            "impulse_n_s":result.impulse_n_s,"recoil_impulse_n_s":result.recoil_impulse_n_s,
            "recoil_residual_n_s":result.recoil_residual_n_s},
        "assumptions":["frozen-stagnation-state","prescribed-area-history","trapezoidal-rate-integration",
            "quasi-steady-samples","single-calorically-perfect-gas","enthalpy-out-is-static-exit-enthalpy",
            "total-enthalpy-includes-exit-kinetic-energy"],
        "nonclaims":{
            "model":["reservoir-coupling-and-blowdown","species-resolved-composition-history","plume-and-acoustics","elapsed-source-depletion"],
            "operation":["case-commitment","certificate-issuance"],"evidence":["empirical-validation"]
        }
    })
}

pub(super) fn text(report: &Report) -> String {
    let (_, _, implementation) = identity(report.kind);
    let (name, title) = match report.kind {
        Kind::Prediction => ("restriction prediction", "RESTRICTION PREDICTED"),
        Kind::History => (
            "restriction history",
            "PRESCRIBED RESTRICTION HISTORY PREDICTED",
        ),
    };
    if let Outcome::Invalid(error) = &report.outcome {
        return format!(
            "{name} failed: {} {} at {:?}\nRecovery: {}\n",
            error.code,
            error.reason_code,
            error.path.chars().take(256).collect::<String>(),
            recovery(report.kind, error.reason_code)
        );
    }
    let mut output = format!(
        "{title}\n\nModel: {MODEL_ID}@{MODEL_VERSION}\nImplementation: {implementation}\nQuantity system: si (explicit)\n"
    );
    match &report.outcome {
        Outcome::Prediction(result) => {
            let _ = writeln!(
                output,
                "\nCONTROL SECTION\n  Regime: {}\n  Area law: {}\n  Effective area: {} m^2\n  Mach: {}\n  Exit pressure: {} Pa\n  Exit temperature: {} K\n  Exit speed: {} m/s\n  Mass flow: {} kg/s\n  Thrust: {} N\n  Source recoil: {} N",
                result.regime.name(),
                result.request.area().name(),
                human_number(result.effective_area_m2),
                human_number(result.throat_mach),
                human_number(result.exit_pressure_pa),
                human_number(result.exit_temperature_k),
                human_number(result.exit_speed_m_per_s),
                human_number(result.mass_flow_kg_per_s),
                human_number(result.thrust_n),
                human_number(result.recoil_n)
            );
        }
        Outcome::History(result) => {
            let _ = writeln!(
                output,
                "\nFROZEN SOURCE HISTORY\n  Samples: {}\n  Mass out: {} kg\n  Static exit enthalpy out: {} J\n  Exit kinetic energy out: {} J\n  Total stagnation enthalpy out: {} J\n  Thrust impulse: {} N s\n  Source recoil impulse: {} N s",
                result.samples.len(),
                human_number(result.mass_out_kg),
                human_number(result.enthalpy_out_j),
                human_number(result.kinetic_energy_out_j),
                human_number(result.total_enthalpy_out_j),
                human_number(result.impulse_n_s),
                human_number(result.recoil_impulse_n_s)
            );
        }
        Outcome::Invalid(_) => unreachable!("invalid reports returned above"),
    }
    output.push_str("\nBALANCE CLAIMS\n");
    for claim in claims(report) {
        let _ = writeln!(
            output,
            "  {}: satisfied-within-roundoff; residual {} {}; tolerance {} {}",
            claim.id,
            human_number(claim.residual),
            claim.unit,
            human_number(claim.tolerance),
            claim.unit
        );
    }
    if report.kind == Kind::Prediction {
        output.push_str("\nExperimental instantaneous control-section prediction. No elapsed-time history,\nreservoir coupling, plume, acoustics, case commitment, certificate issuance,\nor empirical validation.\n");
    } else {
        output.push_str("\nExperimental frozen-source history. No reservoir depletion, composition history,\nplume, acoustics, case commitment, certificate issuance, or empirical validation.\n");
    }
    output.push_str("Human values: six significant digits; full precision in JSON.\n");
    output
}

fn recovery(kind: Kind, reason: &str) -> &'static str {
    match reason {
        "input_not_found" | "input_permission_denied" | "input_unavailable" => {
            "Check the file path and read access, or use - for standard input."
        }
        "input_too_large" => "Provide a request of at most 65,536 bytes.",
        "unsupported_schema" if kind == Kind::Prediction => {
            "Use schema fart.restriction-prediction-request/v0alpha1."
        }
        "unsupported_schema" => "Use schema fart.restriction-history-request/v0alpha1.",
        "unsupported_model_revision" => {
            "Use continuum.quasi-steady-isentropic-converging-restriction@v0alpha1."
        }
        "unsupported_quantity_system" => {
            "Declare quantity_system as si with explicit SI input values."
        }
        "unsupported_area_law" => {
            "Choose prescribed or linear-compliance and supply that law's explicit parameters."
        }
        "invalid_time" => "Use finite nonnegative sample times in strictly increasing order.",
        "invalid_sample_count" => "Provide between 1 and 256 prescribed-area samples.",
        "adverse_pressure" => {
            "Open restrictions require back pressure no greater than the declared source pressure; reverse flow is unsupported."
        }
        _ if kind == Kind::Prediction => {
            "Correct the indicated field; use 'fart help restriction predict' for the input contract."
        }
        _ => {
            "Correct the indicated field; use 'fart help restriction history' for the input contract."
        }
    }
}
