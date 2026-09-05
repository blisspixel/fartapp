// Package restrictionhistoryprediction is the experimental adapter for a
// prescribed-area restriction history. Stagnation is frozen. The report is not
// a blowdown, case, or certificate.
package restrictionhistoryprediction

const (
	RequestSchema          = "fart.restriction-history-request/v0alpha1"
	ReportSchema           = "fart.restriction-history/v0alpha1"
	ModelID                = "continuum.quasi-steady-isentropic-converging-restriction"
	ModelVersion           = "v0alpha1"
	ImplementationRevision = "go-oracle/v0alpha1"
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

type Sample struct {
	TimeSeconds            float64 `json:"time_s"`
	PrescribedSquareMetres float64 `json:"prescribed_m2"`
	EffectiveSquareMetres  float64 `json:"effective_m2"`
	Regime                 string  `json:"regime"`
	ExitPressurePascals    float64 `json:"exit_pressure_pa"`
	MassFlowKilogramsPerS  float64 `json:"mass_flow_kg_per_s"`
	ThrustNewtons          float64 `json:"thrust_n"`
	RecoilNewtons          float64 `json:"recoil_n"`
}

type Totals struct {
	MassOutKilograms            float64 `json:"mass_out_kg"`
	EnthalpyOutJoules           float64 `json:"enthalpy_out_j"`
	KineticEnergyOutJoules      float64 `json:"kinetic_energy_out_j"`
	TotalEnthalpyOutJoules      float64 `json:"total_enthalpy_out_j"`
	ImpulseNewtonSeconds        float64 `json:"impulse_n_s"`
	RecoilImpulseNewtonSeconds  float64 `json:"recoil_impulse_n_s"`
	RecoilResidualNewtonSeconds float64 `json:"recoil_residual_n_s"`
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
	Samples                []Sample              `json:"samples,omitempty"`
	Totals                 *Totals               `json:"totals,omitempty"`
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
