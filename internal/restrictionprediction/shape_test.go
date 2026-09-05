package restrictionprediction

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRestrictionRequestUsesExactClosedNonNullShape(t *testing.T) {
	base := readFixture(t, "gamma15-choked.json")
	tests := []struct {
		name   string
		edit   func(map[string]any)
		path   string
		reason string
	}{
		{"root alias", func(m map[string]any) { m["Schema"] = m["schema"]; delete(m, "schema") }, "/Schema", "document_shape_invalid"},
		{"alias alongside exact", func(m map[string]any) { m["SCHEMA"] = "shadow" }, "/SCHEMA", "document_shape_invalid"},
		{"nested alias", func(m map[string]any) { m["stagnation"].(map[string]any)["Pressure_pa"] = 125000 }, "/stagnation/Pressure_pa", "document_shape_invalid"},
		{"nested unknown", func(m map[string]any) { m["area"].(map[string]any)["future"] = 0 }, "/area/future", "document_shape_invalid"},
		{"null root scalar", func(m map[string]any) { m["schema"] = nil }, "/schema", "document_shape_invalid"},
		{"null required scalar", func(m map[string]any) { m["stagnation"].(map[string]any)["pressure_pa"] = nil }, "/stagnation/pressure_pa", "document_shape_invalid"},
		{"null optional closure field", func(m map[string]any) { m["area"].(map[string]any)["compliance_m2_per_pa"] = nil }, "/area/compliance_m2_per_pa", "document_shape_invalid"},
		{"empty object in scalar", func(m map[string]any) { m["back_pressure_pa"] = map[string]any{} }, "/back_pressure_pa", "document_shape_invalid"},
		{"scalar in object", func(m map[string]any) { m["area"] = 1 }, "/area", "document_shape_invalid"},
		{"empty model", func(m map[string]any) { m["model"] = map[string]any{} }, "/model", "unsupported_model_revision"},
		{"empty stagnation", func(m map[string]any) { m["stagnation"] = map[string]any{} }, "/stagnation/pressure_pa", "missing_member"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var document map[string]any
			if err := json.Unmarshal(base, &document); err != nil {
				t.Fatal(err)
			}
			test.edit(document)
			input, err := json.Marshal(document)
			if err != nil {
				t.Fatal(err)
			}
			report := Predict(input)
			if report.Predicted() || len(report.Diagnostics) != 1 || report.Diagnostics[0].Stage != "schema" || report.Diagnostics[0].Path != test.path || report.Diagnostics[0].ReasonCode != test.reason {
				t.Fatalf("report=%#v", report.Diagnostics)
			}
			if report.Flow != nil || len(report.Claims) != 0 {
				t.Fatal("invalid shape produced a physical result")
			}
		})
	}
	for _, input := range []string{`null`, `0`, `[]`} {
		report := Predict([]byte(input))
		if report.Predicted() || report.Diagnostics[0].ReasonCode != "document_shape_invalid" || report.Diagnostics[0].Path != "/" {
			t.Fatalf("root shape=%#v", report.Diagnostics)
		}
	}
	report := Predict([]byte(strings.Repeat("[", 34) + strings.Repeat("]", 34)))
	if report.Diagnostics[0].Stage != "syntax" || report.Diagnostics[0].ReasonCode != "maximum_depth_exceeded" {
		t.Fatal("shape inspection bypassed the existing syntax limit")
	}
}
