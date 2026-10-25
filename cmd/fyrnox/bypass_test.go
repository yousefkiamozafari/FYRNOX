package main

import (
	"strings"
	"testing"
)

func TestBypassTargetRules(t *testing.T) {
	customRules := "*.ir; mycompany.local; 172.20.*"

	cases := []struct {
		target   string
		expected bool
	}{
		// Local / loopback
		{target: "localhost:8080", expected: true},
		{target: "127.0.0.1:443", expected: true},
		{target: "[::1]:80", expected: true},
		{target: "router.lan", expected: true},

		// Private IPs
		{target: "192.168.1.1:80", expected: true},
		{target: "10.0.0.5:443", expected: true},
		{target: "172.16.0.1:8080", expected: true},

		// Custom bypass rules
		{target: "varzesh3.ir:443", expected: true},
		{target: "sub.varzesh3.ir:443", expected: true},
		{target: "mycompany.local:80", expected: true},
		{target: "172.20.10.4:80", expected: true},

		// Standard internet targets (must route through tunnel)
		{target: "google.com:443", expected: false},
		{target: "api.ipify.org:443", expected: false},
		{target: "cloudflare.com:443", expected: false},
		{target: "1.1.1.1:443", expected: false},
		{target: "8.8.8.8:53", expected: false},
	}

	for _, c := range cases {
		got := isBypassTarget(c.target, customRules)
		if got != c.expected {
			t.Errorf("isBypassTarget(%q) = %v, want %v", c.target, got, c.expected)
		}
	}
}

func TestFormatBypassRules(t *testing.T) {
	rules := FormatBypassRules("*.ir; mycompany.local; 192.168.*")
	if !strings.Contains(rules, "<local>") {
		t.Errorf("expected <local> in bypass rules: %s", rules)
	}
	if !strings.Contains(rules, "*.ir") {
		t.Errorf("expected *.ir in bypass rules: %s", rules)
	}
	if !strings.Contains(rules, "mycompany.local") {
		t.Errorf("expected mycompany.local in bypass rules: %s", rules)
	}
}
