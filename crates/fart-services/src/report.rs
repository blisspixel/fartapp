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
    let transition = match &report.outcome {
        Ok(transition) => transition,
        Err(error) => {
            return format!(
                "reservoir prediction failed: {} {} at {:?}\n",
                error.code, error.reason_code, error.path
            );
        }
    };
    let mut output = format!(
        "RESERVOIR ENDPOINT PREDICTED\n\nModel: {MODEL_ID}@{MODEL_VERSION}\nImplementation: {IMPLEMENTATION_REVISION}\nQuantity system: si (explicit)\nClosure: {}\nWithdrawal fraction: {}\n",
        transition.closure.name(),
        transition.withdrawal.get()
    );
    for (label, state) in [
        ("INITIAL", transition.initial),
        ("FINAL", transition.final_state),
    ] {
        let _ = writeln!(
            output,
            "\n{label}\n  Mass: {} kg\n  Volume: {} m^3\n  Temperature: {} K\n  Pressure: {} Pa\n  Internal energy: {} J",
            state.total_mass_kg,
            state.volume_m3,
            state.temperature_k,
            state.pressure_pa,
            state.internal_energy_j
        );
    }
    let _ = writeln!(
        output,
        "\nTRANSFERS\n  Mass out: {} kg\n  Enthalpy out: {} J\n  Heat into reservoir: {} J\n  Boundary work: 0 J\n\nBALANCE CLAIMS",
        transition.total_mass_out_kg, transition.enthalpy_out_j, transition.heat_in_j
    );
    for claim in transition.claims {
        let _ = writeln!(
            output,
            "  {}: satisfied-within-roundoff; residual {} {}; tolerance {} {}",
            claim.id, claim.residual, claim.unit, claim.tolerance, claim.unit
        );
    }
    output.push_str("\nExperimental analytical endpoint only. No restriction flow, elapsed time, plume,\nphysical audio, case commitment, certificate issuance, or empirical validation.\n");
    output
}
