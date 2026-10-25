package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWindowApiEndpoints(t *testing.T) {
	engine := &VpnEngine{}
	server := NewApiServer(engine, 0)

	endpoints := []string{
		"/api/window/drag",
		"/api/window/minimize",
		"/api/window/maximize",
		"/api/window/close",
	}

	for _, ep := range endpoints {
		req := httptest.NewRequest(http.MethodPost, ep, nil)
		rec := httptest.NewRecorder()

		switch ep {
		case "/api/window/drag":
			server.handleWindowDrag(rec, req)
		case "/api/window/minimize":
			server.handleWindowMinimize(rec, req)
		case "/api/window/maximize":
			server.handleWindowMaximize(rec, req)
		case "/api/window/close":
			server.handleWindowClose(rec, req)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("Endpoint %s returned status %d, expected 200", ep, rec.Code)
		}
	}
}
