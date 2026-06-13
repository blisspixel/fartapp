package main

import "testing"

func TestRate(t *testing.T) {
	if Rate(1) != "gentle" {
		t.Errorf("Rate(1) = %q, want gentle", Rate(1))
	}
	if Rate(5) != "mighty" {
		t.Errorf("Rate(5) = %q, want mighty", Rate(5))
	}
}
