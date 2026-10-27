package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServiceMatrixProbeDirect(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer ts.Close()

	tgt := ServiceProbeTarget{
		ID:       "test-service",
		Name:     "Test Service",
		Category: "Test",
		URL:      ts.URL,
	}

	res := probeTarget(tgt, nil)
	if res.Status != "unrestricted" {
		t.Fatalf("Expected unrestricted, got %s", res.Status)
	}
	if res.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", res.StatusCode)
	}
}

func TestServiceMatrixGeoblocked(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("country, region or territory not supported"))
	}))
	defer ts.Close()

	tgt := ServiceProbeTarget{
		ID:       "openai",
		Name:     "OpenAI / ChatGPT",
		Category: "AI",
		URL:      ts.URL,
	}

	res := probeTarget(tgt, nil)
	if res.Status != "geoblocked" {
		t.Fatalf("Expected geoblocked, got %s", res.Status)
	}
}
