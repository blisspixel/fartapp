use std::fmt::Write;

use fart_core::StateSummary;
use fart_domain::{Closure, ReservoirState};
use serde_json::{Value, json};

use crate::{
    IMPLEMENTATION_REVISION, MODEL_ID, MODEL_VERSION, PredictionReport, REPORT_SCHEMA,
    REQUEST_SCHEMA,
};

pub(crate) fn json_value(report: &PredictionReport) -> Value {
    let environment = json!({"consulted_inputs": report.consulted_inputs, "ambient_inputs": []});
    let transition = match &report.outcome {
        Ok(transition) => transition,
        Err(error) => {
            return json!({
                "schema": REPORT_SCHEMA, "status": "invalid", "implementation_revision": IMPLEMENTATION_REVISION,
                "validation_environment": environment,
                "diagnostics": [{"code": error.code, "stage": error.stage, "path": error.path, "reason_code": error.reason_code}],
            });
        }
    };
    let mut assumptions = vec![
        "calorically-perfect-components",
        "homogeneous-equilibrium-state",
        "nonreacting-mixture",
        "single-gas-phase",
        "perfectly-mixed-nonselective-outflow",
        "rigid-volume",
        "sensible-energy-datum-cv-times-temperature",
    ];
    assumptions.push(if transition.closure == Closure::RigidAdiabatic {
        "adiabatic-no-heat-transfer"
    } else {
        "prescribed-isothermal-ideal-thermostat"
    });
    let transfers: Vec<_> = transition
        .components
        .iter()
        .map(|part| json!({"id": part.id, "mass_out_kg": part.mass_out_kg}))
        .collect();
    let balances: Vec<_> = transition
        .components
        .iter()
        .map(|part| json!({"id": part.id, "residual_kg": part.residual_kg}))
        .collect();
    let claims: Vec<_> = transition.claims.iter().map(|claim| json!({
        "id": claim.id, "status": "satisfied-within-roundoff", "method": claim.method,
        "equation_revision": format!("{MODEL_ID}@{MODEL_VERSION}"),
        "residual": claim.residual, "tolerance": claim.tolerance, "residual_unit": claim.unit,
    })).collect();
    json!({
        "schema": REPORT_SCHEMA, "status": "predicted", "request_schema": REQUEST_SCHEMA,
        "model": {"id": MODEL_ID, "version": MODEL_VERSION}, "implementation_revision": IMPLEMENTATION_REVISION,
        "quantity_system": "si", "closure": transition.closure.name(), "withdrawal_fraction": transition.withdrawal.get(),
        "initial": state_value(&transition.before, transition.initial), "final": state_value(&transition.after, transition.final_state),
        "transfers": {"components": transfers, "total_mass_out_kg": transition.total_mass_out_kg,
            "integrated_enthalpy_out_j": transition.enthalpy_out_j, "heat_into_reservoir_j": transition.heat_in_j, "boundary_work_by_reservoir_j": 0.0},
        "balances": {"components": balances, "total_mass_residual_kg": transition.claims[0].residual,
            "energy_residual_j": transition.claims[1].residual, "initial_eos_residual_j": transition.claims[2].residual, "final_eos_residual_j": transition.claims[3].residual},
        "assumptions": assumptions,
        "nonclaims": {
            "model": ["aperture-and-restriction-flow", "elapsed-time-history", "exterior-state", "momentum-and-recoil", "phase-change-and-reaction", "plume-and-acoustics"],
            "operation": ["case-commitment", "certificate-issuance"], "evidence": ["empirical-validation"]},
        "claims": claims, "validation_environment": environment,
    })
}

fn state_value(state: &ReservoirState, summary: StateSummary) -> Value {
    let components: Vec<_> = state.components().iter().map(|part| json!({"id": part.id(), "mass_kg": part.mass().get(),
        "specific_gas_constant_j_per_kg_k": part.gas_constant().get(), "specific_isochoric_heat_capacity_j_per_kg_k": part.heat_capacity().get()})).collect();
    json!({"components": components, "total_mass_kg": summary.total_mass_kg, "volume_m3": summary.volume_m3, "temperature_k": summary.temperature_k,
        "mixture_gas_constant_j_per_kg_k": summary.gas_constant,
        "mixture_specific_isochoric_heat_capacity_j_per_kg_k": summary.heat_cv,
        "mixture_specific_isobaric_heat_capacity_j_per_kg_k": summary.heat_cp,
        "heat_capacity_ratio": summary.gamma, "pressure_pa": summary.pressure_pa, "internal_energy_j": summary.internal_energy_j})
}

pub(crate) fn text(report: &PredictionReport) -> String {
    use crate::presentation::human_number;
    let transition = match &report.outcome {
        Ok(transition) => transition,
        Err(error) => {
            return format!(
                "reservoir prediction failed: {} {} at {:?}\nRecovery: {}\n",
                error.code,
                error.reason_code,
                error.path.chars().take(256).collect::<String>(),
                recovery(error.reason_code)
            );
        }
    };
    let mut output = format!(
        "RESERVOIR ENDPOINT PREDICTED\n\nModel: {MODEL_ID}@{MODEL_VERSION}\nImplementation: {IMPLEMENTATION_REVISION}\nQuantity system: si (explicit)\nClosure: {}\nWithdrawal fraction: {}\n",
        transition.closure.name(),
        human_number(transition.withdrawal.get())
    );
    for (label, state) in [
        ("INITIAL", transition.initial),
        ("FINAL", transition.final_state),
    ] {
        let _ = writeln!(
            output,
            "\n{label}\n  Mass: {} kg\n  Volume: {} m^3\n  Temperature: {} K\n  Pressure: {} Pa\n  Internal energy: {} J",
            human_number(state.total_mass_kg),
            human_number(state.volume_m3),
            human_number(state.temperature_k),
            human_number(state.pressure_pa),
            human_number(state.internal_energy_j)
        );
    }
    let _ = writeln!(
        output,
        "\nTRANSFERS\n  Mass out: {} kg\n  Enthalpy out: {} J\n  Heat into reservoir: {} J\n  Boundary work: 0 J\n\nBALANCE CLAIMS",
        human_number(transition.total_mass_out_kg),
        human_number(transition.enthalpy_out_j),
        human_number(transition.heat_in_j)
    );
    for claim in transition.claims {
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
    output.push_str("\nExperimental analytical endpoint only. No restriction flow, elapsed time, plume,\nphysical audio, case commitment, certificate issuance, or empirical validation.\n");
    output.push_str("Human values: six significant digits; full precision in JSON.\n");
    output
}

fn recovery(reason: &str) -> &'static str {
    match reason {
        "input_not_found" | "input_permission_denied" | "input_unavailable" => {
            "Check the file path and read access, or use - for standard input."
        }
        "input_too_large" => "Provide a request of at most 65,536 bytes.",
        "unsupported_schema" => "Use schema fart.reservoir-prediction-request/v0alpha1.",
        "unsupported_model" => "Use continuum.rigid-calorically-perfect-ideal-mixture@v0alpha1.",
        "unsupported_closure" => "Choose rigid-adiabatic or rigid-isothermal explicitly.",
        _ => {
            "Correct the indicated field; use 'fart help reservoir predict' for the input contract."
        }
    }
}
