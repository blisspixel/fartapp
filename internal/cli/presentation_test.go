package cli

import (
	"encoding/json"
	"math"
	"os"
	"strconv"
	"testing"
)

func TestHumanNumbersMatchSharedPresentationCases(t *testing.T) {
	data, err := os.ReadFile("testdata/presentation/numbers.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Input   string `json:"input"`
		Display string `json:"display"`
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, item := range cases {
		t.Run(item.Input, func(t *testing.T) {
			value, err := strconv.ParseFloat(item.Input, 64)
			if err != nil {
				t.Fatal(err)
			}
			if got := formatScientificValue(value); got != item.Display {
				t.Fatalf("display = %q, want %q", got, item.Display)
			}
			if value != 0 && formatScientificValue(value) == "0" {
				t.Fatal("nonzero value lost")
			}
		})
	}
	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if got := formatScientificValue(value); got != "unavailable" {
			t.Fatalf("nonfinite display = %q", got)
		}
	}
}
