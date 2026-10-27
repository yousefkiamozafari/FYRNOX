package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type ServiceProbeTarget struct {
	ID       string
	Name     string
	Category string
	URL      string
}

var defaultServiceTargets = []ServiceProbeTarget{
	{ID: "openai", Name: "OpenAI / ChatGPT", Category: "AI", URL: "https://api.openai.com/v1/models"},
	{ID: "steam", Name: "Steam Store", Category: "Gaming", URL: "https://store.steampowered.com"},
	{ID: "spotify", Name: "Spotify Web API", Category: "Media", URL: "https://api.spotify.com/v1"},
	{ID: "docker", Name: "Docker Hub", Category: "Developer", URL: "https://hub.docker.com"},
	{ID: "youtube", Name: "YouTube Service", Category: "Video", URL: "https://www.youtube.com/generate_204"},
	{ID: "discord", Name: "Discord Gateway", Category: "Community", URL: "https://discord.com/api/v10/gateway"},
}

type ServiceStatus struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Status      string `json:"status"` // "unrestricted", "geoblocked", "blocked"
	StatusCode  int    `json:"status_code"`
	LatencyMs   int    `json:"latency_ms"`
	Message     string `json:"message"`
	LastChecked string `json:"last_checked"`
}

type ServiceMatrixResult struct {
	Services    []ServiceStatus `json:"services"`
	VpnActive   bool            `json:"vpn_active"`
	ExitCountry string          `json:"exit_country"`
	LastUpdated string          `json:"last_updated"`
}

type ServiceMatrixManager struct {
	mu          sync.Mutex
	lastResult  ServiceMatrixResult
	lastChecked time.Time
}

var globalServiceMatrix = &ServiceMatrixManager{}

func (m *ServiceMatrixManager) GetMatrix(engine *VpnEngine) ServiceMatrixResult {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Return cached result if fresh (< 30 seconds old) and not empty
	if len(m.lastResult.Services) > 0 && time.Since(m.lastChecked) < 30*time.Second {
		return m.lastResult
	}

	// Otherwise execute probes
	return m.runProbesLocked(engine)
}

func (m *ServiceMatrixManager) Refresh(engine *VpnEngine) ServiceMatrixResult {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.runProbesLocked(engine)
}

func (m *ServiceMatrixManager) runProbesLocked(engine *VpnEngine) ServiceMatrixResult {
	vpnActive := false
	exitCountry := ""
	var proxyURL *url.URL

	if engine != nil {
		status := engine.GetStatus()
		vpnActive = (status.State == StateConnected)
		exitCountry = status.ExitCountry
		if vpnActive {
			// Probe via the local HTTP proxy listener
			proxyURL, _ = url.Parse("http://127.0.0.1:20809")
		}
	}

	var wg sync.WaitGroup
	results := make([]ServiceStatus, len(defaultServiceTargets))

	for i, target := range defaultServiceTargets {
		wg.Add(1)
		go func(idx int, tgt ServiceProbeTarget) {
			defer wg.Done()
			results[idx] = probeTarget(tgt, proxyURL)
		}(i, target)
	}

	wg.Wait()

	m.lastChecked = time.Now()
	m.lastResult = ServiceMatrixResult{
		Services:    results,
		VpnActive:   vpnActive,
		ExitCountry: exitCountry,
		LastUpdated: m.lastChecked.Format("15:04:05"),
	}
	return m.lastResult
}

func probeTarget(tgt ServiceProbeTarget, proxyURL *url.URL) ServiceStatus {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		DisableKeepAlives: true,
	}
	if proxyURL != nil {
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   2500 * time.Millisecond,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, tgt.URL, nil)
	if err != nil {
		return ServiceStatus{
			ID:          tgt.ID,
			Name:        tgt.Name,
			Category:    tgt.Category,
			Status:      "blocked",
			StatusCode:  0,
			LatencyMs:   int(time.Since(start).Milliseconds()),
			Message:     "Request build error",
			LastChecked: time.Now().Format("15:04:05"),
		}
	}

	// Set modern browser User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "*/*")

	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		return ServiceStatus{
			ID:          tgt.ID,
			Name:        tgt.Name,
			Category:    tgt.Category,
			Status:      "blocked",
			StatusCode:  0,
			LatencyMs:   latency,
			Message:     "Filtered / Connection Timed Out",
			LastChecked: time.Now().Format("15:04:05"),
		}
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	bodyStr := strings.ToLower(string(bodyBytes))

	status := "unrestricted"
	msg := fmt.Sprintf("HTTP %d OK", resp.StatusCode)

	switch tgt.ID {
	case "openai":
		// 401 Unauthorized means OpenAI reached and requires auth token
		if resp.StatusCode == 401 || resp.StatusCode == 200 {
			status = "unrestricted"
			msg = "Unrestricted (API reachable)"
		} else if resp.StatusCode == 403 || strings.Contains(bodyStr, "unsupported_country") || strings.Contains(bodyStr, "country, region") || strings.Contains(bodyStr, "cloudflare") {
			status = "geoblocked"
			msg = "Geoblocked / Region Restricted"
		} else {
			status = "unrestricted"
		}

	case "steam":
		if resp.StatusCode == 200 || resp.StatusCode == 302 {
			status = "unrestricted"
			msg = "Unrestricted (Store accessible)"
		} else if resp.StatusCode == 451 || resp.StatusCode == 403 {
			status = "geoblocked"
			msg = "Geoblocked"
		} else {
			status = "unrestricted"
		}

	case "spotify":
		if resp.StatusCode == 200 || resp.StatusCode == 401 {
			status = "unrestricted"
			msg = "Unrestricted (API reachable)"
		} else if resp.StatusCode == 403 || strings.Contains(bodyStr, "restricted") {
			status = "geoblocked"
			msg = "Geoblocked / Region Restricted"
		} else {
			status = "unrestricted"
		}

	case "docker":
		if resp.StatusCode == 200 {
			status = "unrestricted"
			msg = "Unrestricted (Hub accessible)"
		} else if resp.StatusCode == 403 || strings.Contains(bodyStr, "not available in your region") {
			status = "geoblocked"
			msg = "Sanctioned / 403 Forbidden"
		} else {
			status = "unrestricted"
		}

	case "youtube":
		if resp.StatusCode == 204 || resp.StatusCode == 200 {
			status = "unrestricted"
			msg = "Unrestricted (No Filtering)"
		} else if resp.StatusCode == 403 || resp.StatusCode == 451 {
			status = "geoblocked"
			msg = "Filtered / Restricted"
		} else {
			status = "unrestricted"
		}

	case "discord":
		if resp.StatusCode == 200 {
			status = "unrestricted"
			msg = "Unrestricted (Gateway live)"
		} else if resp.StatusCode == 403 || resp.StatusCode == 451 {
			status = "geoblocked"
			msg = "Geoblocked / Access Denied"
		} else {
			status = "unrestricted"
		}
	}

	return ServiceStatus{
		ID:          tgt.ID,
		Name:        tgt.Name,
		Category:    tgt.Category,
		Status:      status,
		StatusCode:  resp.StatusCode,
		LatencyMs:   latency,
		Message:     msg,
		LastChecked: time.Now().Format("15:04:05"),
	}
}
