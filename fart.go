package main

import "fmt"

const (
	minimumIntensity = 1
	maximumIntensity = 5
)

type emission struct {
	sound  string
	rating string
}

var emissions = [...]emission{
	{sound: "pfft", rating: "gentle"},
	{sound: "toot", rating: "respectable"},
	{sound: "braaap", rating: "respectable"},
	{sound: "blorp", rating: "respectable"},
	{sound: "KABLAM", rating: "mighty"},
}

// intensity stores a validated, zero-based index into emissions. Keeping the
// type private forces command input through newIntensity before it reaches the
// permanent oracle table.
type intensity uint8

func newIntensity(value int) (intensity, error) {
	if value < minimumIntensity || value > maximumIntensity {
		return 0, fmt.Errorf(
			"intensity %d must be from %d to %d",
			value,
			minimumIntensity,
			maximumIntensity,
		)
	}
	return intensity(value - minimumIntensity), nil
}

func (i intensity) emission() emission {
	return emissions[i]
}
