package main

import (
	"testing"
)

func TestIsDomesticIranTarget(t *testing.T) {
	tests := []struct {
		target   string
		expected bool
	}{
		{"shaparak.ir", true},
		{"pep.shaparak.ir:443", true},
		{"bmi.ir:80", true},
		{"bankmellat.ir", true},
		{"tejaratbank.ir", true},
		{"zarinpal.com", true},
		{"snapp.ir", true},
		{"digikala.com", true},
		{"varzesh3.com", true},
		{"myket.ir", true},
		{"google.com", false},
		{"youtube.com:443", false},
		{"twitter.com", false},
		{"1.1.1.1:53", false},
		{"8.8.8.8", false},
		// Iranian CIDR test: 5.160.0.1 (in 5.160.0.0/12)
		{"5.160.0.1:443", true},
	}

	for _, tt := range tests {
		got := IsDomesticIranTarget(tt.target)
		if got != tt.expected {
			t.Errorf("IsDomesticIranTarget(%q) = %v; want %v", tt.target, got, tt.expected)
		}
	}
}
