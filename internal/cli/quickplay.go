package cli

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/blisspixel/fartapp/internal/walkcase"
)

const quickplaySchema = "fart.quick-play/v0alpha1"

const quickplayHelp = `Run one synthetic walk and print a short account.

Usage:
  fartapp quickplay [1-5] [--seed <integer>] [--format text|json]

Intensity selects a versioned laboratory preset. It does not calibrate a body.
  1 and 2  gentle, a smaller-area subsonic sibling of ordinary-low-pressure
  3 and 4  ordinary, testdata/walk/ordinary-low-pressure.json
  5        choked, testdata/walk/isothermal-choked.json

The seed is recorded and does not change the calculated account. The same
intensity and seed produce the same JSON. Default intensity is 3 and the
default seed is 1.

Synthetic laboratory preset, not a Reference Pfft or a biological claim.

Example:
  fartapp quickplay 3
`

var quickplayNext = []string{
	"fartapp walk branch testdata/walk/ordinary-low-pressure.json",
	"fartapp walk explain testdata/walk/ordinary-low-pressure.json",
	"fartapp walk certify testdata/walk/ordinary-low-pressure.json",
}

type quickplayReport struct {
	Schema       string   `json:"schema"`
	Intensity    int      `json:"intensity"`
	Seed         int64    `json:"seed"`
	Preset       string   `json:"preset"`
	Sound        string   `json:"sound"`
	Rating       string   `json:"rating"`
	ElapsedS     float64  `json:"elapsed_s"`
	PeakMassFlow float64  `json:"peak_mass_flow_kg_per_s"`
	Choked       bool     `json:"choked"`
	ChokedCheek  string   `json:"choked_cheek_criterion"`
	MassOutKg    float64  `json:"mass_out_kg"`
	Explanation  string   `json:"explanation"`
	Nonclaims    []string `json:"nonclaims"`
	TryNext      []string `json:"try_next"`
}

func runQuickplay(args []string, stdout, stderr io.Writer) int {
	options, values, err := parseValuedOptions(args, "--seed")
	if err != nil {
		writeDiagnostic(stderr, "invalid quickplay: %v\n", err)
		return 1
	}
	if options.help {
		return writeText(stdout, stderr, quickplayHelp)
	}
	if len(options.positional) > 1 {
		writeDiagnostic(stderr, "usage: fartapp quickplay [1-5] [--seed <integer>] [--format text|json]\n")
		return 1
	}
	level := 3
	if len(options.positional) == 1 {
		raw := options.positional[0]
		value, convErr := strconv.Atoi(raw)
		if convErr != nil || strconv.Itoa(value) != raw {
			writeDiagnostic(stderr, "invalid intensity %s: must be a canonical integer from %d to %d\n", quoteInput(raw), minimumIntensity, maximumIntensity)
			return 1
		}
		if _, err = newIntensity(value); err != nil {
			writeDiagnostic(stderr, "invalid intensity %s: must be a canonical integer from %d to %d\n", quoteInput(raw), minimumIntensity, maximumIntensity)
			return 1
		}
		level = value
	}
	seed := int64(1)
	if raw, ok := values["--seed"]; ok {
		parsed, convErr := strconv.ParseInt(raw, 10, 64)
		if convErr != nil || strconv.FormatInt(parsed, 10) != raw {
			writeDiagnostic(stderr, "invalid quickplay seed %s: must be a canonical integer\n", quoteInput(raw))
			return 1
		}
		seed = parsed
	}
	name := presetForIntensity(level)
	data, err := presetBytes(name)
	if err != nil {
		writeDiagnostic(stderr, "quickplay failed: %s\n", quoteInput(err.Error()))
		return 1
	}
	account := walkcase.Run(data, "simulate")
	if !account.Predicted() {
		return writeWalkReport(account, "quickplay", options.format, stdout, stderr)
	}
	event := emissions[level-minimumIntensity]
	elapsed := 0.0
	if account.ElapsedSeconds != nil {
		elapsed = *account.ElapsedSeconds
	}
	massOut := 0.0
	if account.MassOutKilograms != nil {
		massOut = *account.MassOutKilograms
	}
	peak := 0.0
	for _, sample := range account.History {
		if sample.MassFlowKilogramsPerSecond > peak {
			peak = sample.MassFlowKilogramsPerSecond
		}
	}
	choked := account.Signature != nil && account.Signature.ChokedOccurred
	cheek := "not met"
	if choked {
		cheek = "met"
	}
	explanation := "The walk account did not include an explanation."
	if len(account.Explanation) > 0 && account.Explanation[0] != "" {
		explanation = account.Explanation[0]
	}
	report := quickplayReport{
		Schema:       quickplaySchema,
		Intensity:    level,
		Seed:         seed,
		Preset:       name,
		Sound:        event.sound,
		Rating:       event.rating,
		ElapsedS:     elapsed,
		PeakMassFlow: peak,
		Choked:       choked,
		ChokedCheek:  cheek,
		MassOutKg:    massOut,
		Explanation:  explanation,
		Nonclaims:    []string{"Synthetic laboratory preset, not a Reference Pfft or a biological claim."},
		TryNext:      append([]string(nil), quickplayNext...),
	}
	return writeValue(stdout, stderr, options.format, report, formatQuickplay)
}

func formatQuickplay(report quickplayReport) string {
	var output strings.Builder
	fmt.Fprintf(&output, "Quick Play %d: %s (%s)\n", report.Intensity, report.Sound, report.Rating)
	fmt.Fprintf(&output, "Preset: %s\n%s\nSeed: %d\n", report.Preset, report.Nonclaims[0], report.Seed)
	fmt.Fprintf(&output, "Duration: %s s\nPeak mass flow: %s kg/s\n", formatScientificValue(report.ElapsedS), formatScientificValue(report.PeakMassFlow))
	fmt.Fprintf(&output, "Choked Cheek Criterion: %s (model regime flag, not a biological claim)\n", report.ChokedCheek)
	fmt.Fprintf(&output, "Mass out: %s kg\nExplanation: %s\nTry next:\n", formatScientificValue(report.MassOutKg), report.Explanation)
	for _, command := range report.TryNext {
		fmt.Fprintf(&output, "  %s\n", command)
	}
	return output.String()
}
