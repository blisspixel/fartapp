// Package restrictionprediction defines the bounded, read-only request and
// report for the first analytical restriction-flow predictor. It is an
// experimental adapter over restrictionflow, not a case, realization, or
// certificate format.
package restrictionprediction

const (
	RequestSchema          = "fart.restriction-prediction-request/v0alpha1"
	ReportSchema           = "fart.restriction-prediction/v0alpha1"
	ModelID                = "continuum.quasi-steady-isentropic-converging-restriction"
	ModelVersion           = "v0alpha1"
	ImplementationRevision = "go-oracle.restriction/v0alpha2"
	QuantitySystem         = "si"
	MaxInputBytes          = 65_536
	maximumJSONDepth       = 32
	maximumMemberNameBytes = 128
)

type Diagnostic struct {
	Code       string `json:"code"`
	Stage      string `json:"stage"`
	Path       string `json:"path"`
	ReasonCode string `json:"reason_code"`
}

type ValidationEnvironment struct {
	ConsultedInputs []string `json:"consulted_inputs"`
	AmbientInputs   []string `json:"ambient_inputs"`
}

type ModelReference struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

type StagnationState struct {
	PressurePascals                            float64 `json:"pressure_pa"`
	TemperatureKelvin                          float64 `json:"temperature_k"`
	SpecificGasConstantJoulesPerKilogramKelvin float64 `json:"specific_gas_constant_j_per_kg_k"`
	HeatCapacityRatio                          float64 `json:"heat_capacity_ratio"`
}

type AreaState struct {
	Law                      string   `json:"law"`
	PrescribedSquareMetres   float64  `json:"prescribed_m2"`
	ComplianceSquareMetresPa *float64 `json:"compliance_m2_per_pa,omitempty"`
	MaximumSquareMetres      *float64 `json:"maximum_m2,omitempty"`
	EffectiveSquareMetres    float64  `json:"effective_m2"`
}

type FlowState struct {
	Regime                     string  `json:"regime"`
	CriticalPressureRatio      float64 `json:"critical_pressure_ratio"`
	BackPressureRatio          float64 `json:"back_pressure_ratio"`
	ThroatMach                 float64 `json:"throat_mach"`
	ExitPressurePascals        float64 `json:"exit_pressure_pa"`
	ExitTemperatureKelvin      float64 `json:"exit_temperature_k"`
	ExitSpeedMetresPerSecond   float64 `json:"exit_speed_m_per_s"`
	MassFlowKilogramsPerSecond float64 `json:"mass_flow_kg_per_s"`
	SonicMassFlowKilogramsPerS float64 `json:"sonic_mass_flow_kg_per_s"`
	ThrustNewtons              float64 `json:"thrust_n"`
	RecoilNewtons              float64 `json:"recoil_n"`
}

type Balances struct {
	MassFlowResidualKilogramsPerSecond float64 `json:"mass_flow_residual_kg_per_s"`
	ThrustResidualNewtons              float64 `json:"thrust_residual_n"`
	RecoilResidualNewtons              float64 `json:"recoil_residual_n"`
}

type Claim struct {
	ID               string  `json:"id"`
	Status           string  `json:"status"`
	Method           string  `json:"method"`
	EquationRevision string  `json:"equation_revision"`
	Residual         float64 `json:"residual"`
	Tolerance        float64 `json:"tolerance"`
	ResidualUnit     string  `json:"residual_unit"`
}

type Nonclaims struct {
	Model     []string `json:"model"`
	Operation []string `json:"operation"`
	Evidence  []string `json:"evidence"`
}

type Report struct {
	Schema                 string                `json:"schema"`
	Status                 string                `json:"status"`
	RequestSchema          string                `json:"request_schema,omitempty"`
	Model                  *ModelReference       `json:"model,omitempty"`
	ImplementationRevision string                `json:"implementation_revision"`
	QuantitySystem         string                `json:"quantity_system,omitempty"`
	Stagnation             *StagnationState      `json:"stagnation,omitempty"`
	BackPressurePascals    *float64              `json:"back_pressure_pa,omitempty"`
	DischargeCoefficient   *float64              `json:"discharge_coefficient,omitempty"`
	Area                   *AreaState            `json:"area,omitempty"`
	Flow                   *FlowState            `json:"flow,omitempty"`
	Balances               *Balances             `json:"balances,omitempty"`
	Assumptions            []string              `json:"assumptions,omitempty"`
	Nonclaims              *Nonclaims            `json:"nonclaims,omitempty"`
	Claims                 []Claim               `json:"claims,omitempty"`
	ValidationEnvironment  ValidationEnvironment `json:"validation_environment"`
	Diagnostics            []Diagnostic          `json:"diagnostics,omitempty"`
}

func (report Report) Predicted() bool {
	return report.Schema == ReportSchema && report.Status == "predicted" &&
		len(report.Diagnostics) == 0
}
