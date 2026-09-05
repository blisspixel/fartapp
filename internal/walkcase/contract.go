// Package walkcase adapts an experimental explicit SI coupled-blowdown model.
// Its witnesses compare versioned software accounts. They are not case
// identities, empirical evidence, signatures, or archive certificates.
package walkcase

const (
	RequestSchema          = "fart.walk-case/v0alpha1"
	ReportSchema           = "fart.walk-report/v0alpha1"
	ModelID                = "continuum.quasi-steady-coupled-blowdown"
	ModelVersion           = "v0alpha1"
	ImplementationRevision = "go-oracle.walk/v0alpha2"
	WitnessSchema          = "fart.walk-witness/v0alpha1"
	InputDigestSchema      = "fart.walk-normalized-input/v0alpha1"
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

type DimensionDiagnostic struct {
	Quantity  string `json:"quantity"`
	Unit      string `json:"unit"`
	Dimension string `json:"dimension"`
	Status    string `json:"status"`
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

type ModelReference struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

type NumericalPolicy struct {
	Method           string `json:"method"`
	Precision        string `json:"precision"`
	GoVersion        string `json:"go_version"`
	OperatingSystem  string `json:"operating_system"`
	Architecture     string `json:"architecture"`
	MaximumSamples   int    `json:"maximum_samples"`
	HistorySemantics string `json:"history_semantics"`
}

type ComponentMass struct {
	ID               string  `json:"id"`
	MassKilograms    float64 `json:"mass_kg"`
	MassOutKilograms float64 `json:"mass_out_kg"`
}

type HistorySample struct {
	TimeSeconds                float64         `json:"time_s"`
	MassKilograms              float64         `json:"reservoir_mass_kg"`
	PressurePascals            float64         `json:"reservoir_pressure_pa"`
	TemperatureKelvin          float64         `json:"reservoir_temperature_k"`
	MassFlowKilogramsPerSecond float64         `json:"mass_flow_kg_per_s"`
	SourceTotalEnthalpyWatts   float64         `json:"source_total_enthalpy_flow_w"`
	ExitSpeedMetresPerSecond   float64         `json:"exit_speed_m_per_s"`
	ExitPressurePascals        float64         `json:"exit_pressure_pa"`
	ExitTemperatureKelvin      float64         `json:"exit_temperature_k"`
	EffectiveAreaSquareMetres  float64         `json:"effective_area_m2"`
	ThrustNewtons              float64         `json:"thrust_n"`
	RecoilNewtons              float64         `json:"recoil_n"`
	Regime                     string          `json:"regime"`
	Components                 []ComponentMass `json:"components"`
}

type Nonclaims struct {
	Model     []string `json:"model"`
	Operation []string `json:"operation"`
	Evidence  []string `json:"evidence"`
}

type RestrictionSnapshot struct {
	Regime                string  `json:"regime"`
	MassFlowKilogramsPerS float64 `json:"mass_flow_kg_per_s"`
	CriticalPressureRatio float64 `json:"critical_pressure_ratio"`
	BackPressureRatio     float64 `json:"back_pressure_ratio"`
	ThrustNewtons         float64 `json:"thrust_n"`
}

type Endpoint struct {
	MassKilograms        float64 `json:"mass_kg"`
	PressurePascals      float64 `json:"pressure_pa"`
	TemperatureKelvin    float64 `json:"temperature_k"`
	InternalEnergyJoules float64 `json:"internal_energy_j"`
}

type Signature struct {
	EquivalentDiameterMetres float64  `json:"equivalent_diameter_m"`
	StrokeLengthMetres       float64  `json:"stroke_length_m"`
	FormationNumber          *float64 `json:"formation_number,omitempty"`
	ChokedOccurred           bool     `json:"choked_occurred"`
}

type BranchComparison struct {
	PrescribedAreaSquareMetres float64 `json:"variant_prescribed_area_m2"`
	BaselineStop               string  `json:"baseline_stop"`
	VariantStop                string  `json:"variant_stop"`
	BothEqualized              bool    `json:"both_equalized"`
	BaselineElapsedSeconds     float64 `json:"baseline_elapsed_s"`
	VariantElapsedSeconds      float64 `json:"variant_elapsed_s"`
	BaselineMassOutKg          float64 `json:"baseline_mass_out_kg"`
	VariantMassOutKg           float64 `json:"variant_mass_out_kg"`
	SameMassEndpoint           bool    `json:"same_mass_endpoint"`
	MassComparisonToleranceKg  float64 `json:"mass_comparison_tolerance_kg"`
	Variant                    *Report `json:"variant"`
}

type Report struct {
	Schema                               string                `json:"schema"`
	Status                               string                `json:"status"`
	Operation                            string                `json:"operation,omitempty"`
	RequestSchema                        string                `json:"request_schema,omitempty"`
	ImplementationRevision               string                `json:"implementation_revision"`
	Model                                *ModelReference       `json:"model,omitempty"`
	NumericalPolicy                      *NumericalPolicy      `json:"numerical_policy,omitempty"`
	Inputs                               *requestDocument      `json:"inputs,omitempty"`
	QuantitySystem                       string                `json:"quantity_system,omitempty"`
	LawContext                           string                `json:"law_context,omitempty"`
	Closure                              string                `json:"closure,omitempty"`
	Stop                                 string                `json:"stop,omitempty"`
	ElapsedSeconds                       *float64              `json:"elapsed_s,omitempty"`
	Steps                                *int                  `json:"steps,omitempty"`
	EqualizationFraction                 *float64              `json:"equalization_fraction,omitempty"`
	EndpointReachability                 string                `json:"endpoint_reachability,omitempty"`
	EqualizationPressureTolerancePascals *float64              `json:"equalization_pressure_tolerance_pa,omitempty"`
	InitialRestriction                   *RestrictionSnapshot  `json:"initial_restriction,omitempty"`
	Initial                              *Endpoint             `json:"initial,omitempty"`
	Final                                *Endpoint             `json:"final,omitempty"`
	MassOutKilograms                     *float64              `json:"mass_out_kg,omitempty"`
	EnthalpyOutJoules                    *float64              `json:"enthalpy_out_j,omitempty"`
	HeatInJoules                         *float64              `json:"heat_in_j,omitempty"`
	ImpulseNewtonSeconds                 *float64              `json:"impulse_n_s,omitempty"`
	RecoilImpulseNewtonSeconds           *float64              `json:"recoil_impulse_n_s,omitempty"`
	History                              []HistorySample       `json:"history,omitempty"`
	Signature                            *Signature            `json:"signature,omitempty"`
	Dimensions                           []DimensionDiagnostic `json:"dimensions,omitempty"`
	Explanation                          []string              `json:"explanation,omitempty"`
	Branch                               *BranchComparison     `json:"branch,omitempty"`
	Witness                              string                `json:"witness,omitempty"`
	WitnessSchema                        string                `json:"witness_schema,omitempty"`
	WitnessAlgorithm                     string                `json:"witness_algorithm,omitempty"`
	InputDigest                          string                `json:"input_digest,omitempty"`
	InputDigestSchema                    string                `json:"input_digest_schema,omitempty"`
	ExpectedWitness                      string                `json:"expected_witness,omitempty"`
	ReconstructedWitness                 string                `json:"reconstructed_witness,omitempty"`
	WitnessMatch                         *bool                 `json:"witness_match,omitempty"`
	Assumptions                          []string              `json:"assumptions,omitempty"`
	Nonclaims                            *Nonclaims            `json:"nonclaims,omitempty"`
	Claims                               []Claim               `json:"claims,omitempty"`
	ValidationEnvironment                ValidationEnvironment `json:"validation_environment"`
	Diagnostics                          []Diagnostic          `json:"diagnostics,omitempty"`
}

func (report Report) Predicted() bool {
	return report.Schema == ReportSchema && report.Status == "predicted" &&
		len(report.Diagnostics) == 0
}
