package main

import (
	"testing"
)

func TestLatencyClassificationThresholds(t *testing.T) {
	cases := []struct {
		ping     int
		expected string
	}{
		{ping: 45, expected: "fast"},
		{ping: 100, expected: "fast"},
		{ping: 200, expected: "fast"},
		{ping: 201, expected: "medium"},
		{ping: 500, expected: "medium"},
		{ping: 501, expected: "slow"},
		{ping: 850, expected: "slow"},
		{ping: -1, expected: "unavailable"},
		{ping: 0, expected: "unavailable"},
	}
	for _, c := range cases {
		var got string
		if c.ping > 0 && c.ping <= 200 {
			got = "fast"
		} else if c.ping > 200 && c.ping <= 500 {
			got = "medium"
		} else if c.ping > 500 {
			got = "slow"
		} else {
			got = "unavailable"
		}
		if got != c.expected {
			t.Errorf("ping %d: got %s, want %s", c.ping, got, c.expected)
		}
	}
}

func TestComputeMedianLatency(t *testing.T) {
	tests := []struct {
		name     string
		samples  []int
		expected int
	}{
		{
			name:     "empty samples",
			samples:  []int{},
			expected: -1,
		},
		{
			name:     "all failed (-1)",
			samples:  []int{-1, -1, -1},
			expected: -1,
		},
		{
			name:     "single successful sample",
			samples:  []int{-1, 120, -1},
			expected: 120,
		},
		{
			name:     "three valid samples odd count",
			samples:  []int{140, 95, 110},
			expected: 110,
		},
		{
			name:     "even valid samples",
			samples:  []int{100, 200},
			expected: 150,
		},
		{
			name:     "four samples with one failure",
			samples:  []int{-1, 90, 80, 100},
			expected: 90,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := computeMedianLatency(tc.samples)
			if actual != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, actual)
			}
		})
	}
}
