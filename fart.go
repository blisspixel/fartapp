package main

// sounds, from gentle to mighty.
var sounds = []string{"pfft", "toot", "braaap", "blorp", "KABLAM"}

// Pick returns the fart sound for an intensity from 1 (gentle) to 5 (mighty).
func Pick(intensity int) string {
	return sounds[intensity-1]
}

// Rate describes an intensity in words.
func Rate(intensity int) string {
	switch {
	case intensity <= 1:
		return "gentle"
	case intensity >= 5:
		return "mighty"
	default:
		return "respectable"
	}
}
