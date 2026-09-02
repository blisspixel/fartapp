// Package reservoirprediction defines the bounded, read-only request and report
// for the first analytical reservoir predictor. It is an experimental adapter
// over idealmixturereservoir, not a case, realization, or certificate format.
package reservoirprediction

const (
	RequestSchema          = "fart.reservoir-prediction-request/v0alpha1"
	ReportSchema           = "fart.reservoir-prediction/v0alpha1"
	ModelID                = "continuum.rigid-calorically-perfect-ideal-mixture"
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

type ComponentState struct {
	ID                                                   string  `json:"id"`
	MassKilograms                                        float64 `json:"mass_kg"`
	SpecificGasConstantJoulesPerKilogramKelvin           float64 `json:"specific_gas_constant_j_per_kg_k"`
	SpecificIsochoricHeatCapacityJoulesPerKilogramKelvin float64 `json:"specific_isochoric_heat_capacity_j_per_kg_k"`
}

type ReservoirState struct {
	Components                                          []ComponentState `json:"components"`
	TotalMassKilograms                                  float64          `json:"total_mass_kg"`
	VolumeCubicMetres                                   float64          `json:"volume_m3"`
	TemperatureKelvin                                   float64          `json:"temperature_k"`
	MixtureGasConstantJoulesPerKilogramKelvin           float64          `json:"mixture_gas_constant_j_per_kg_k"`
	MixtureSpecificIsochoricHeatJoulesPerKilogramKelvin float64          `json:"mixture_specific_isochoric_heat_capacity_j_per_kg_k"`
	MixtureSpecificIsobaricHeatJoulesPerKilogramKelvin  float64          `json:"mixture_specific_isobaric_heat_capacity_j_per_kg_k"`
	HeatCapacityRatio                                   float64          `json:"heat_capacity_ratio"`
	PressurePascals                                     float64          `json:"pressure_pa"`
	InternalEnergyJoules                                float64          `json:"internal_energy_j"`
}

type ComponentTransfer struct {
	ID               string  `json:"id"`
	MassOutKilograms float64 `json:"mass_out_kg"`
}

type Transfers struct {
	Components                    []ComponentTransfer `json:"components"`
	TotalMassOutKilograms         float64             `json:"total_mass_out_kg"`
	IntegratedEnthalpyOutJoules   float64             `json:"integrated_enthalpy_out_j"`
	HeatIntoReservoirJoules       float64             `json:"heat_into_reservoir_j"`
	BoundaryWorkByReservoirJoules float64             `json:"boundary_work_by_reservoir_j"`
}

type ComponentMassBalance struct {
	ID                string  `json:"id"`
	ResidualKilograms float64 `json:"residual_kg"`
}

type Balances struct {
	Components                 []ComponentMassBalance `json:"components"`
	TotalMassResidualKilograms float64                `json:"total_mass_residual_kg"`
	EnergyResidualJoules       float64                `json:"energy_residual_j"`
	InitialEOSResidualJoules   float64                `json:"initial_eos_residual_j"`
	FinalEOSResidualJoules     float64                `json:"final_eos_residual_j"`
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
	Closure                string                `json:"closure,omitempty"`
	WithdrawalFraction     *float64              `json:"withdrawal_fraction,omitempty"`
	Initial                *ReservoirState       `json:"initial,omitempty"`
	Final                  *ReservoirState       `json:"final,omitempty"`
	Transfers              *Transfers            `json:"transfers,omitempty"`
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
