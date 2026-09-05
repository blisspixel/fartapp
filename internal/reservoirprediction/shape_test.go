package reservoirprediction

import (
	"encoding/json"
	"os"
	"testing"
)

func TestExactShapeRejectsAliasesAndExplicitNull(t *testing.T) {
	data, err := os.ReadFile("../../testdata/reservoir/synthetic-mixture-adiabatic.json")
	if err != nil {
		t.Fatal(err)
	}
	invalidInputs := [][]byte{[]byte("null"), []byte("[]")}
	for _, mutate := range []func(map[string]any){
		func(v map[string]any) { v["Schema"] = v["schema"]; delete(v, "schema") },
		func(v map[string]any) {
			c := v["initial"].(map[string]any)["components"].([]any)[0].(map[string]any)
			c["Mass_Kg"] = c["mass_kg"]
			delete(c, "mass_kg")
		},
		func(v map[string]any) { v["initial"] = nil },
		func(v map[string]any) { v["withdrawal_fraction"] = nil },
	} {
		var value map[string]any
		if err := json.Unmarshal(data, &value); err != nil {
			t.Fatal(err)
		}
		mutate(value)
		invalid, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		invalidInputs = append(invalidInputs, invalid)
	}
	for _, invalid := range invalidInputs {
		report := Predict(invalid)
		if report.Predicted() || len(report.Diagnostics) != 1 || report.Diagnostics[0].ReasonCode != "document_shape_invalid" {
			t.Fatalf("shape refusal = %#v", report.Diagnostics)
		}
	}
}
