package restrictionhistoryprediction

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHistoryRequestUsesExactClosedNonNullShape(t *testing.T) {
	base, err := os.ReadFile(filepath.Join("..", "..", "testdata", "restriction", "gamma15-choked-history.json"))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		edit   func(map[string]any)
		path   string
		reason string
	}{
		{"root alias", func(m map[string]any) { m["Schema"] = m["schema"]; delete(m, "schema") }, "/Schema", "document_shape_invalid"},
		{"alias alongside exact", func(m map[string]any) { m["SCHEMA"] = "shadow" }, "/SCHEMA", "document_shape_invalid"},
		{"nested alias", func(m map[string]any) { m["stagnation"].(map[string]any)["Pressure_pa"] = 125000 }, "/stagnation/Pressure_pa", "document_shape_invalid"},
		{"sample alias", func(m map[string]any) { m["samples"].([]any)[0].(map[string]any)["Time_s"] = 0 }, "/samples/0/Time_s", "document_shape_invalid"},
		{"sample unknown", func(m map[string]any) { m["samples"].([]any)[0].(map[string]any)["future"] = 0 }, "/samples/0/future", "document_shape_invalid"},
		{"null scalar", func(m map[string]any) { m["back_pressure_pa"] = nil }, "/back_pressure_pa", "document_shape_invalid"},
		{"null samples", func(m map[string]any) { m["samples"] = nil }, "/samples", "document_shape_invalid"},
		{"null sample", func(m map[string]any) { m["samples"].([]any)[0] = nil }, "/samples/0", "document_shape_invalid"},
		{"null sample value", func(m map[string]any) { m["samples"].([]any)[0].(map[string]any)["time_s"] = nil }, "/samples/0/time_s", "document_shape_invalid"},
		{"empty object in scalar", func(m map[string]any) { m["discharge_coefficient"] = map[string]any{} }, "/discharge_coefficient", "document_shape_invalid"},
		{"object in array", func(m map[string]any) { m["samples"] = map[string]any{} }, "/samples", "document_shape_invalid"},
		{"scalar in sample", func(m map[string]any) { m["samples"].([]any)[0] = 0 }, "/samples/0", "document_shape_invalid"},
		{"empty sample", func(m map[string]any) { m["samples"].([]any)[0] = map[string]any{} }, "/samples/0", "missing_member"},
		{"empty samples", func(m map[string]any) { m["samples"] = []any{} }, "/samples", "missing_member"},
		{"empty stagnation", func(m map[string]any) { m["stagnation"] = map[string]any{} }, "/stagnation", "missing_member"},
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
			if len(report.Samples) != 0 {
				t.Fatal("invalid shape produced physical history")
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
