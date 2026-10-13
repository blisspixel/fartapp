package cli

import (
	"embed"
	"fmt"
	"io"
)

//go:embed presets/gentle.json presets/ordinary.json presets/choked.json
var presetFiles embed.FS

func presetForIntensity(level int) string {
	switch {
	case level <= 2:
		return "gentle"
	case level >= 5:
		return "choked"
	default:
		return "ordinary"
	}
}

func presetBytes(name string) ([]byte, error) {
	switch name {
	case "gentle", "ordinary", "choked":
	default:
		return nil, fmt.Errorf("unknown preset %s; expected gentle, ordinary, or choked", quoteInput(name))
	}
	data, err := presetFiles.ReadFile("presets/" + name + ".json")
	if err != nil {
		return nil, err
	}
	return data, nil
}

const walkNewHelp = `Write one synthetic laboratory walk case.

Usage:
  fartapp walk new --preset gentle|ordinary|choked

Presets:
  gentle    Smaller prescribed area. Quick Play 1 and 2. Subsonic.
  ordinary  testdata/walk/ordinary-low-pressure.json. Quick Play 3 and 4.
  choked    testdata/walk/isothermal-choked.json. Quick Play 5.

Stdout is strict fart.walk-case/v0alpha1 JSON. Pipe it to walk simulate.
These presets are synthetic laboratory cases, not a Reference Pfft or a
biological claim.

Example:
  fartapp walk new --preset ordinary
`

func runWalkNew(args []string, stdout, stderr io.Writer) int {
	options, values, err := parseValuedOptions(args, "--preset")
	if err != nil {
		writeDiagnostic(stderr, "invalid walk new: %v\n", err)
		return 1
	}
	if options.help {
		return writeText(stdout, stderr, walkNewHelp)
	}
	if len(options.positional) != 0 {
		writeDiagnostic(stderr, "usage: fartapp walk new --preset gentle|ordinary|choked\n")
		return 1
	}
	name, ok := values["--preset"]
	if !ok {
		writeDiagnostic(stderr, "usage: fartapp walk new --preset gentle|ordinary|choked\n")
		return 1
	}
	data, err := presetBytes(name)
	if err != nil {
		writeDiagnostic(stderr, "invalid walk new: %v\n", err)
		return 1
	}
	if len(data) == 0 || data[len(data)-1] != '\n' {
		data = append(append([]byte{}, data...), '\n')
	}
	return writeText(stdout, stderr, string(data))
}
