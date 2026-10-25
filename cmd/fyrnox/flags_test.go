package main

import (
	"testing"
)

func TestFlagCoverageForVerifiedCountries(t *testing.T) {
	engine := NewVpnEngine()
	servers := engine.GetServers()
	if len(servers) < 30 {
		t.Fatalf("expected at least 30 servers, got %d", len(servers))
	}

	codes := make(map[string]bool)
	for _, s := range servers {
		if s.CountryCode == "" {
			t.Fatalf("server %s missing country code", s.Hostname)
		}
		codes[s.CountryCode] = true
	}

	t.Logf("Found %d distinct countries in server catalog: %+v", len(codes), codes)
	if len(codes) < 25 {
		t.Fatalf("expected at least 25 unique country codes, got %d", len(codes))
	}
}
