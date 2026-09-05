package reservoirprediction

import (
	"math"
	"os"
	"testing"
)

func TestExtremeFiniteReportsRetainIndependentTransfersAndEndpoints(t *testing.T) {
	for _, test := range []struct {
		file                         string
		finalP, finalT, finalU, heat float64
	}{
		{"extreme-finite-isothermal.json", 1 - math.Ldexp(1, -20), math.Ldexp(1, 1000),
			math.Ldexp(1, -60) * (1 - math.Ldexp(1, -20)), math.Ldexp(1, -80)},
		{"extreme-finite-adiabatic.json", math.Ldexp(1, -1027), math.Ldexp(1, -1048), math.Ldexp(1, -1037), 0},
	} {
		data, err := os.ReadFile("../../testdata/reservoir/" + test.file)
		if err != nil {
			t.Fatal(err)
		}
		report := Predict(data)
		if !report.Predicted() {
			t.Fatalf("%s: %#v", test.file, report.Diagnostics)
		}
		if report.Final.PressurePascals != test.finalP || report.Final.TemperatureKelvin != test.finalT ||
			report.Final.InternalEnergyJoules != test.finalU || report.Transfers.HeatIntoReservoirJoules != test.heat {
			t.Fatalf("%s lost independently known endpoint or heat", test.file)
		}
		for _, claim := range report.Claims {
			if math.Abs(claim.Residual) > claim.Tolerance || math.IsInf(claim.Tolerance, 0) {
				t.Fatalf("%s: invalid finite balance evidence", test.file)
			}
		}
	}
}
