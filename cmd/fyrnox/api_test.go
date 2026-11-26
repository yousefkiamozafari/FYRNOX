package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestApiSecurity_AllowedAndRejectedOriginsAndHosts(t *testing.T) {
	engine := NewVpnEngine()
	api := NewApiServer(engine, 18080)
	handler, err := api.buildHandler()
	if err != nil {
		t.Fatalf("buildHandler error: %v", err)
	}

	// 1. Allowed Origin (exact WebView2 local origin)
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:18080/api/status", nil)
	req.Host = "127.0.0.1:18080"
	req.Header.Set("Origin", "http://127.0.0.1:18080")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for allowed origin, got %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://127.0.0.1:18080" {
		t.Fatalf("expected exact ACAO header, got %q", got)
	}

	// 2. Rejected external Origin
	reqBadOrigin := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:18080/api/disconnect", nil)
	reqBadOrigin.Host = "127.0.0.1:18080"
	reqBadOrigin.Header.Set("Origin", "https://evil.example")
	recBadOrigin := httptest.NewRecorder()
	handler.ServeHTTP(recBadOrigin, reqBadOrigin)
	if recBadOrigin.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for rejected origin, got %d", recBadOrigin.Code)
	}

	// 3. Rejected cross-site Sec-Fetch-Site on state-mutating POST
	reqCrossSite := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:18080/api/disconnect", nil)
	reqCrossSite.Host = "127.0.0.1:18080"
	reqCrossSite.Header.Set("Sec-Fetch-Site", "cross-site")
	recCrossSite := httptest.NewRecorder()
	handler.ServeHTTP(recCrossSite, reqCrossSite)
	if recCrossSite.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for cross-site POST, got %d", recCrossSite.Code)
	}

	// 4. Invalid Host header (DNS rebinding protection)
	reqBadHost := httptest.NewRequest(http.MethodGet, "http://evil.example:18080/api/status", nil)
	reqBadHost.Host = "evil.example:18080"
	recBadHost := httptest.NewRecorder()
	handler.ServeHTTP(recBadHost, reqBadHost)
	if recBadHost.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for invalid Host header, got %d", recBadHost.Code)
	}

	// 5. Allowed state-mutating POST from exact local origin
	reqValidPost := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:18080/api/select-server", strings.NewReader(`{"hostname":"nonexistent.example"}`))
	reqValidPost.Host = "127.0.0.1:18080"
	reqValidPost.Header.Set("Origin", "http://127.0.0.1:18080")
	reqValidPost.Header.Set("Sec-Fetch-Site", "same-origin")
	recValidPost := httptest.NewRecorder()
	handler.ServeHTTP(recValidPost, reqValidPost)
	if recValidPost.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid local POST, got %d", recValidPost.Code)
	}
}

func TestApiServer_StartAndGracefulShutdown(t *testing.T) {
	engine := NewVpnEngine()
	api := NewApiServer(engine, 0) // ephemeral port
	if err := api.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/status", api.port))
	if err != nil {
		t.Fatalf("GET /api/status failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := api.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	if _, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/status", api.port)); err == nil {
		t.Fatalf("expected connection failure after Shutdown")
	}
}
